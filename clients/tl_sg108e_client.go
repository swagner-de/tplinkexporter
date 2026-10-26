package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var allInfoRe = regexp.MustCompile(`all_info = ({[^;]*});`)

type TPLINKSwitchClient interface {
	GetPortStats() ([]portStats, error)
	GetHost() string
}

type TPLINKSwitch struct {
	host     string
	username string
	password string
	client   *http.Client
}

type portStats struct {
	State      int
	LinkStatus int
	PktCount   map[string]int
}

func (client *TPLINKSwitch) GetHost() string {
	return client.host
}

// login authenticates against the switch web UI. The session cookie set by the
// switch is stored in the client's cookie jar and sent on subsequent requests.
func (client *TPLINKSwitch) login() error {
	resp, err := client.client.PostForm(
		fmt.Sprintf("http://%s/logon.cgi", client.host),
		url.Values{"username": {client.username}, "password": {client.password}, "logon": {"Login"}},
	)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	return nil
}

func (client *TPLINKSwitch) GetPortStats() ([]portStats, error) {
	if err := client.login(); err != nil {
		return nil, err
	}

	resp, err := client.client.Get(fmt.Sprintf("http://%s/PortStatisticsRpm.htm", client.host))
	if err != nil {
		return nil, fmt.Errorf("port statistics request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading port statistics response: %w", err)
	}

	return parsePortStats(body)
}

// parsePortStats extracts per-port statistics from the inline JavaScript
// `all_info` blob returned by PortStatisticsRpm.htm. The switch returns a
// JavaScript object (not valid JSON), so we rewrite the known keys and quote
// them before unmarshalling.
func parsePortStats(body []byte) ([]portStats, error) {
	jbody := strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(
				string(body), "link_status", `"linkStatus"`),
			"state", `"State"`),
		"pkts", `"Pkts"`)

	res := allInfoRe.FindStringSubmatch(jbody)
	if res == nil {
		return nil, errors.New("unexpected response for port statistics http call: " + jbody)
	}

	type allInfo struct {
		State      []int
		LinkStatus []int
		Pkts       []int
	}
	var jparsed allInfo
	if err := json.Unmarshal([]byte(res[1]), &jparsed); err != nil {
		return nil, fmt.Errorf("parsing port statistics payload: %w", err)
	}

	portcount := len(jparsed.State)
	portsInfos := make([]portStats, 0, portcount)
	for i := range portcount {
		var portInfo portStats
		portInfo.State = jparsed.State[i]
		portInfo.LinkStatus = jparsed.LinkStatus[i]
		// Counters are only meaningful for enabled ports (State == 1).
		if portInfo.State == 1 {
			portInfo.PktCount = map[string]int{
				"TxGoodPkt": jparsed.Pkts[4*i],
				"TxBadPkt":  jparsed.Pkts[4*i+1],
				"RxGoodPkt": jparsed.Pkts[4*i+2],
				"RxBadPkt":  jparsed.Pkts[4*i+3],
			}
		}
		portsInfos = append(portsInfos, portInfo)
	}
	return portsInfos, nil
}

/*
sample output of PortStatisticsRpm.htm call:
<script>
var max_port_num = 8;
var port_middle_num  = 16;
var all_info = {
state:[1,1,1,1,1,1,1,1,0,0],
link_status:[6,6,0,6,0,0,0,5,0,0],
pkts:[1901830310,0,1338131260,33254, ... 4 counters per port: TxGood,TxBad,RxGood,RxBad]
};
var tip = "";
</script>
*/

func NewTPLinkSwitch(host, username, password string, timeout time.Duration) *TPLINKSwitch {
	jar, _ := cookiejar.New(nil)
	return &TPLINKSwitch{
		host:     host,
		username: username,
		password: password,
		client: &http.Client{
			Jar:     jar,
			Timeout: timeout,
		},
	}
}

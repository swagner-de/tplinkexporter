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

var (
	allInfoRe    = regexp.MustCompile(`all_info = ({[^;]*});`)
	infoDsRe     = regexp.MustCompile(`info_ds = ({[^;]*});`)
	bareKeyRe    = regexp.MustCompile(`([{,]\s*)([A-Za-z_]\w*)(\s*:)`)
	trailingComa = regexp.MustCompile(`,(\s*[}\]])`)
)

type TPLINKSwitchClient interface {
	GetPortStats() ([]portStats, error)
	GetPortSettings() ([]portSetting, error)
	GetSystemInfo() (*systemInfo, error)
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

type portSetting struct {
	State             int
	SpeedConfig       int
	SpeedActual       int
	FlowControlConfig int
	FlowControlActual int
}

type systemInfo struct {
	Name     string
	MAC      string
	IP       string
	Firmware string
	Hardware string
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

// fetchPage logs in and GETs an Easy Smart web UI page. The switch has a short
// session table and authorises by source IP (CVE-2017-17746), so re-logging in
// before each page is the most reliable way to avoid mid-scrape session drops.
func (client *TPLINKSwitch) fetchPage(page string) ([]byte, error) {
	if err := client.login(); err != nil {
		return nil, err
	}
	resp, err := client.client.Get(fmt.Sprintf("http://%s/%s", client.host, page))
	if err != nil {
		return nil, fmt.Errorf("request for %s failed: %w", page, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading %s response: %w", page, err)
	}
	return body, nil
}

func (client *TPLINKSwitch) GetPortStats() ([]portStats, error) {
	body, err := client.fetchPage("PortStatisticsRpm.htm")
	if err != nil {
		return nil, err
	}
	return parsePortStats(body)
}

func (client *TPLINKSwitch) GetPortSettings() ([]portSetting, error) {
	body, err := client.fetchPage("PortSettingRpm.htm")
	if err != nil {
		return nil, err
	}
	return parsePortSettings(body)
}

func (client *TPLINKSwitch) GetSystemInfo() (*systemInfo, error) {
	body, err := client.fetchPage("SystemInfoRpm.htm")
	if err != nil {
		return nil, err
	}
	return parseSystemInfo(body)
}

// jsObjectToJSON turns an inline JS object literal (bare keys, trailing commas)
// into valid JSON.
func jsObjectToJSON(blob string) string {
	blob = bareKeyRe.ReplaceAllString(blob, `$1"$2"$3`)
	blob = trailingComa.ReplaceAllString(blob, `$1`)
	return blob
}

func parsePortSettings(body []byte) ([]portSetting, error) {
	res := allInfoRe.FindStringSubmatch(string(body))
	if res == nil {
		return nil, errors.New("unexpected response for port settings http call: " + string(body))
	}
	type allInfo struct {
		State  []int `json:"state"`
		SpdCfg []int `json:"spd_cfg"`
		SpdAct []int `json:"spd_act"`
		FcCfg  []int `json:"fc_cfg"`
		FcAct  []int `json:"fc_act"`
	}
	var parsed allInfo
	if err := json.Unmarshal([]byte(jsObjectToJSON(res[1])), &parsed); err != nil {
		return nil, fmt.Errorf("parsing port settings payload: %w", err)
	}
	settings := make([]portSetting, 0, len(parsed.State))
	for i := range parsed.State {
		settings = append(settings, portSetting{
			State:             parsed.State[i],
			SpeedConfig:       parsed.SpdCfg[i],
			SpeedActual:       parsed.SpdAct[i],
			FlowControlConfig: parsed.FcCfg[i],
			FlowControlActual: parsed.FcAct[i],
		})
	}
	return settings, nil
}

// parseSystemInfo reads the info_ds object; each field is a single-element array.
func parseSystemInfo(body []byte) (*systemInfo, error) {
	res := infoDsRe.FindStringSubmatch(string(body))
	if res == nil {
		return nil, errors.New("unexpected response for system info http call: " + string(body))
	}
	type infoDs struct {
		Descri   []string `json:"descriStr"`
		MAC      []string `json:"macStr"`
		IP       []string `json:"ipStr"`
		Firmware []string `json:"firmwareStr"`
		Hardware []string `json:"hardwareStr"`
	}
	var parsed infoDs
	if err := json.Unmarshal([]byte(jsObjectToJSON(res[1])), &parsed); err != nil {
		return nil, fmt.Errorf("parsing system info payload: %w", err)
	}
	first := func(s []string) string {
		if len(s) > 0 {
			return s[0]
		}
		return ""
	}
	return &systemInfo{
		Name:     first(parsed.Descri),
		MAC:      first(parsed.MAC),
		IP:       first(parsed.IP),
		Firmware: first(parsed.Firmware),
		Hardware: first(parsed.Hardware),
	}, nil
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

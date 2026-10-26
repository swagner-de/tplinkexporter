package clients

import "testing"

const samplePktsLine = "pkts:[1901830310,0,1338131260,33254,4291149014,0,2311488878,564," +
	"0,0,0,0,1814018004,0,33552310,0,0,0,0,0,0,0,0,0,0,0,0,0,1678459124,0,1866169392,0,0,0]"

var samplePayload = `<script>
var max_port_num = 8;
var port_middle_num  = 16;
var all_info = {
state:[1,1,1,1,1,1,1,1,0,0],
link_status:[6,6,0,6,0,0,0,5,0,0],
` + samplePktsLine + `
};
var tip = "";
</script>`

func TestParsePortStats(t *testing.T) {
	stats, err := parsePortStats([]byte(samplePayload))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(stats) != 10 {
		t.Fatalf("len(stats) = %d, want 10", len(stats))
	}

	if stats[0].State != 1 {
		t.Errorf("port 1 State = %d, want 1", stats[0].State)
	}
	if stats[0].LinkStatus != 6 {
		t.Errorf("port 1 LinkStatus = %d, want 6", stats[0].LinkStatus)
	}
	if got := stats[0].PktCount["TxGoodPkt"]; got != 1901830310 {
		t.Errorf("port 1 TxGoodPkt = %d, want 1901830310", got)
	}
	if got := stats[0].PktCount["RxGoodPkt"]; got != 1338131260 {
		t.Errorf("port 1 RxGoodPkt = %d, want 1338131260", got)
	}
	if got := stats[0].PktCount["RxBadPkt"]; got != 33254 {
		t.Errorf("port 1 RxBadPkt = %d, want 33254", got)
	}

	// Port 9 is disabled (state 0), so it carries no packet counters.
	if stats[8].State != 0 {
		t.Errorf("port 9 State = %d, want 0", stats[8].State)
	}
	if stats[8].PktCount != nil {
		t.Errorf("port 9 PktCount = %v, want nil", stats[8].PktCount)
	}
}

func TestParsePortStatsUnexpectedBody(t *testing.T) {
	if _, err := parsePortStats([]byte("<html>login page</html>")); err == nil {
		t.Fatal("expected error for unexpected body, got nil")
	}
}

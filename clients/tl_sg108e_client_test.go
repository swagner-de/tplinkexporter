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

// Captured from a TL-SG108PE (firmware 1.0.1): ports 1,4,8 up at 1000M (spd 6),
// ports 2-3,5-7 enabled but no link (spd 0), ports 9-10 disabled.
var samplePortSettings = `<script>
var max_port_num = 8;
var all_info = {
state:[1,1,1,1,1,1,1,1,0,0],
trunk_info:[0,0,0,0,0,0,0,0,0,0],
spd_cfg:[1,1,1,1,1,1,1,1,0,0],
spd_act:[6,0,0,6,0,0,0,6,0,0],
fc_cfg:[0,0,0,0,0,0,0,0,0,0],
fc_act:[0,0,0,0,0,0,0,0,0,0]
};
var tip = "";
</script>`

func TestParsePortSettings(t *testing.T) {
	settings, err := parsePortSettings([]byte(samplePortSettings))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(settings) != 10 {
		t.Fatalf("len(settings) = %d, want 10", len(settings))
	}
	if settings[0].SpeedConfig != 1 || settings[0].SpeedActual != 6 {
		t.Errorf("port 1 speed cfg/act = %d/%d, want 1/6", settings[0].SpeedConfig, settings[0].SpeedActual)
	}
	if settings[1].SpeedActual != 0 {
		t.Errorf("port 2 SpeedActual = %d, want 0 (no link)", settings[1].SpeedActual)
	}
	if settings[0].FlowControlConfig != 0 || settings[0].FlowControlActual != 0 {
		t.Errorf("port 1 flow control = %d/%d, want 0/0", settings[0].FlowControlConfig, settings[0].FlowControlActual)
	}
}

// Structure matches SystemInfoRpm; values are placeholders, not a real device.
var sampleSystemInfo = `<script>
var info_ds = {
descriStr:["switch1"],
macStr:["00:00:5E:00:53:01"],
ipStr:["192.0.2.10"],
netmaskStr:["255.255.255.0"],
gatewayStr:["192.0.2.1"],
firmwareStr:["1.0.1 Build 20191204 Rel.71847"],
hardwareStr:["TL-SG108PE 2.0"]
};
var tip = "";
</script>`

func TestParseSystemInfo(t *testing.T) {
	info, err := parseSystemInfo([]byte(sampleSystemInfo))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.Name != "switch1" {
		t.Errorf("Name = %q, want switch1", info.Name)
	}
	if info.Firmware != "1.0.1 Build 20191204 Rel.71847" {
		t.Errorf("Firmware = %q", info.Firmware)
	}
	if info.Hardware != "TL-SG108PE 2.0" {
		t.Errorf("Hardware = %q", info.Hardware)
	}
}

// A login redirect (no data blob) must be an error, not silent empty data, for
// every parser.
func TestParsersRejectLoginRedirect(t *testing.T) {
	body := []byte("<html>login page</html>")
	if _, err := parsePortStats(body); err == nil {
		t.Error("parsePortStats: expected error, got nil")
	}
	if _, err := parsePortSettings(body); err == nil {
		t.Error("parsePortSettings: expected error, got nil")
	}
	if _, err := parseSystemInfo(body); err == nil {
		t.Error("parseSystemInfo: expected error, got nil")
	}
}

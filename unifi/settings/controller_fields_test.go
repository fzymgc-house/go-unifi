package settings

import (
	"encoding/json"
	"strings"
	"testing"
)

// Network 10.6 stores three keys on global_switch that the generated struct lacked. The two
// numbers arrive as JSON numbers today and as strings on some controllers, and a zero must
// survive the round trip, because set/setting/global_switch replaces the whole section.
func TestGlobalSwitchRoundTripsControllerFields(t *testing.T) {
	for name, raw := range map[string]string{
		"numbers": `{"key":"global_switch","dhcp_snoop":true,"auto_stp_edge_detection_enabled":false,"link_debounce":0,"poe_staging_delay_msec":200}`,
		"strings": `{"key":"global_switch","dhcp_snoop":true,"auto_stp_edge_detection_enabled":false,"link_debounce":"0","poe_staging_delay_msec":"200"}`,
	} {
		t.Run(name, func(t *testing.T) {
			var s GlobalSwitch
			if err := json.Unmarshal([]byte(raw), &s); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if !s.DHCPSnoop {
				t.Error("DHCPSnoop = false")
			}
			if s.AutoStpEdgeDetectionEnabled == nil || *s.AutoStpEdgeDetectionEnabled {
				t.Errorf("AutoStpEdgeDetectionEnabled = %v, want false", s.AutoStpEdgeDetectionEnabled)
			}
			if s.LinkDebounce == nil || *s.LinkDebounce != 0 || s.PoeStagingDelayMsec == nil || *s.PoeStagingDelayMsec != 200 {
				t.Errorf("LinkDebounce=%v PoeStagingDelayMsec=%v", s.LinkDebounce, s.PoeStagingDelayMsec)
			}
			out, err := json.Marshal(&s)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			for _, want := range []string{`"auto_stp_edge_detection_enabled":false`, `"link_debounce":0`, `"poe_staging_delay_msec":200`} {
				if !strings.Contains(string(out), want) {
					t.Errorf("marshal dropped %s: %s", want, out)
				}
			}
		})
	}
}

func TestIgmpSnoopingRoundTripsAutoUnknownTrafficHandling(t *testing.T) {
	var s IgmpSnooping
	if err := json.Unmarshal([]byte(`{"key":"igmp_snooping","enabled":true,"auto_unknown_traffic_handling":true}`), &s); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if s.AutoUnknownTrafficHandling == nil || !*s.AutoUnknownTrafficHandling {
		t.Fatalf("AutoUnknownTrafficHandling = %v, want true", s.AutoUnknownTrafficHandling)
	}
	out, err := json.Marshal(&s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `"auto_unknown_traffic_handling":true`) {
		t.Errorf("marshal dropped auto_unknown_traffic_handling: %s", out)
	}
}

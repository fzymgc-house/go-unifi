package unifi_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ubiquiti-community/go-unifi/unifi"
)

// Network 10.6 returns external_id on every network and ipv6_enabled on most.
// A typed read-modify-write must carry both, or the update deletes them.
func TestNetworkMarshalKeepsControllerFields(t *testing.T) {
	for _, purpose := range []string{unifi.PurposeCorporate, unifi.PurposeGuest, unifi.PurposeVLANOnly} {
		t.Run(purpose, func(t *testing.T) {
			raw := `{"purpose":"` + purpose + `","name":"n","external_id":"e460f890-5af6-4a4c-a68b-7de25dfc526a","ipv6_enabled":true,"ip_subnet":"192.168.40.1/22"}`
			var n unifi.Network
			if err := json.Unmarshal([]byte(raw), &n); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if n.ExternalID != "e460f890-5af6-4a4c-a68b-7de25dfc526a" {
				t.Errorf("ExternalID = %q", n.ExternalID)
			}
			if n.IPV6Enabled == nil || !*n.IPV6Enabled {
				t.Errorf("IPV6Enabled = %v, want true", n.IPV6Enabled)
			}
			out, err := json.Marshal(&n)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			for _, want := range []string{`"external_id":"e460f890-5af6-4a4c-a68b-7de25dfc526a"`, `"ipv6_enabled":true`} {
				if !strings.Contains(string(out), want) {
					t.Errorf("marshal dropped %s: %s", want, out)
				}
			}
		})
	}
}

// A network the controller stores without ipv6_enabled must not gain the key on update.
func TestNetworkMarshalOmitsAbsentIPv6Toggle(t *testing.T) {
	var n unifi.Network
	if err := json.Unmarshal([]byte(`{"purpose":"vlan-only","name":"IoT","vlan":3020}`), &n); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := json.Marshal(&n)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(out), "ipv6_enabled") {
		t.Errorf("marshal invented ipv6_enabled: %s", out)
	}
}

// The nine portconf keys Network 10.6 stores on a port profile that the generated struct lacked.
func TestPortProfileRoundTripsControllerFields(t *testing.T) {
	raw := `{
		"_id": "5f1f9a4e9c9b7d0a4c3a1b2c",
		"name": "AllMainDefault",
		"eee_enabled": false,
		"flow_control_enabled": true,
		"link_debounce_auto": true,
		"multicast_router_mode": "NONE",
		"precision_time_protocol_enabled": true,
		"stp_bpdu_guard_enabled": false,
		"stp_edge_state": "disabled",
		"stp_uplink": false,
		"tagged_networkconf_ids": ["604d59e39c9b7d1005aabaad"]
	}`
	var p unifi.PortProfile
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.FlowControlEnabled == nil || !*p.FlowControlEnabled || p.EeeEnabled == nil || *p.EeeEnabled {
		t.Errorf("flow_control_enabled=%v eee_enabled=%v", p.FlowControlEnabled, p.EeeEnabled)
	}
	if p.MulticastRouterMode != "NONE" || p.StpEdgeState != "disabled" {
		t.Errorf("multicast_router_mode=%q stp_edge_state=%q", p.MulticastRouterMode, p.StpEdgeState)
	}
	out, err := json.Marshal(&p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{
		`"eee_enabled":false`, `"flow_control_enabled":true`, `"link_debounce_auto":true`, `"multicast_router_mode":"NONE"`,
		`"precision_time_protocol_enabled":true`, `"stp_bpdu_guard_enabled":false`, `"stp_edge_state":"disabled"`,
		`"stp_uplink":false`, `"tagged_networkconf_ids":["604d59e39c9b7d1005aabaad"]`,
	} {
		if !strings.Contains(string(out), want) {
			t.Errorf("marshal dropped %s", want)
		}
	}
}

// The v2 apgroups endpoint returns for_wlanconf on every group. A false value must survive a PUT.
func TestAPGroupRoundTripsForWLANConf(t *testing.T) {
	var g unifi.APGroup
	if err := json.Unmarshal([]byte(`{"_id":"604d4e579c9b7d1005aaba9c","name":"All APs","device_macs":[],"for_wlanconf":false}`), &g); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if g.ForWLANConf == nil || *g.ForWLANConf {
		t.Fatalf("ForWLANConf = %v, want false", g.ForWLANConf)
	}
	out, err := json.Marshal(&g)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `"for_wlanconf":false`) {
		t.Errorf("marshal dropped for_wlanconf: %s", out)
	}
}

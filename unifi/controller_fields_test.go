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

// A network's zone is firewall_zone_id. The controller replaces the whole object on update, so an
// encoder that drops the key moves the network back to its default zone.
func TestNetworkMarshalKeepsFirewallZone(t *testing.T) {
	for _, purpose := range []string{unifi.PurposeCorporate, unifi.PurposeGuest, unifi.PurposeVLANOnly, unifi.PurposeWAN} {
		t.Run(purpose, func(t *testing.T) {
			raw := `{"purpose":"` + purpose + `","name":"n","firewall_zone_id":"6abd09ef1f816fd2e03301ab"}`
			var n unifi.Network
			if err := json.Unmarshal([]byte(raw), &n); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			out, err := json.Marshal(&n)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if !strings.Contains(string(out), `"firewall_zone_id":"6abd09ef1f816fd2e03301ab"`) {
				t.Errorf("marshal dropped firewall_zone_id: %s", out)
			}
		})
	}
}

// ipv6_aliases holds extra IPv6 addresses on a LAN network, such as a ULA gateway beside a
// delegated prefix. The LAN encoders always send it, as they send ip_aliases.
func TestNetworkMarshalKeepsIPv6Aliases(t *testing.T) {
	for _, purpose := range []string{unifi.PurposeCorporate, unifi.PurposeGuest} {
		t.Run(purpose, func(t *testing.T) {
			raw := `{"purpose":"` + purpose + `","name":"n","ipv6_interface_type":"pd","ipv6_aliases":["fd00:601::1/64"]}`
			var n unifi.Network
			if err := json.Unmarshal([]byte(raw), &n); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if len(n.IPV6Aliases) != 1 || n.IPV6Aliases[0] != "fd00:601::1/64" {
				t.Errorf("IPV6Aliases = %v", n.IPV6Aliases)
			}
			out, err := json.Marshal(&n)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if !strings.Contains(string(out), `"ipv6_aliases":["fd00:601::1/64"]`) {
				t.Errorf("marshal dropped ipv6_aliases: %s", out)
			}

			var empty unifi.Network
			if err := json.Unmarshal([]byte(`{"purpose":"`+purpose+`","name":"n"}`), &empty); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			out, err = json.Marshal(&empty)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if !strings.Contains(string(out), `"ipv6_aliases":[]`) {
				t.Errorf("marshal of a network without aliases must send an empty list: %s", out)
			}
		})
	}
}

// Network 10.6 stores routing_table_id on each WAN network.
func TestNetworkMarshalKeepsWANRoutingTable(t *testing.T) {
	var n unifi.Network
	if err := json.Unmarshal([]byte(`{"purpose":"wan","name":"Internet 1","routing_table_id":201}`), &n); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if n.RoutingTableID == nil || *n.RoutingTableID != 201 {
		t.Errorf("RoutingTableID = %v, want 201", n.RoutingTableID)
	}
	out, err := json.Marshal(&n)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `"routing_table_id":201`) {
		t.Errorf("marshal dropped routing_table_id: %s", out)
	}
}

// Network 10.6 stores external_id and a null cloud_template on every firewall zone.
func TestFirewallZoneRoundTripsControllerFields(t *testing.T) {
	raw := `{"name":"Internal","zone_key":"internal","network_ids":[],"cloud_template":null,"external_id":"53b02959-ebf9-445d-8b3f-6f1c2e9d7a10"}`
	var z unifi.FirewallZone
	if err := json.Unmarshal([]byte(raw), &z); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := json.Marshal(&z)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{`"external_id":"53b02959-ebf9-445d-8b3f-6f1c2e9d7a10"`, `"cloud_template":null`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("marshal dropped %s: %s", want, out)
		}
	}
}

// Network 10.6 stores origin_id on every firewall policy.
func TestFirewallPolicyRoundTripsOriginID(t *testing.T) {
	raw := `{"name":"Allow DNS","action":"ALLOW","origin_id":"6abd09ef1f816fd2e03301c4"}`
	var p unifi.FirewallPolicy
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := json.Marshal(&p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `"origin_id":"6abd09ef1f816fd2e03301c4"`) {
		t.Errorf("marshal dropped origin_id: %s", out)
	}
}

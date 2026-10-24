package unifi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// storedACLRule is the shape Network 11.0.97 returns for a MAC ACL rule that an
// object policy compiled, with documentation MACs.
const storedACLRule = `{
	"_id": "obj_00000000-0000-0000-0000-000000000000_1",
	"acl_index": 1,
	"action": "ALLOW",
	"description": "Automatically created from the \"tv\" Object.",
	"enabled": true,
	"external_id": "00000000-0000-0000-0000-000000000001",
	"ip_acl_protocol": "ALL",
	"name": "tv",
	"origin_id": "00000000-0000-0000-0000-000000000000",
	"origin_type": "object_acl_rule",
	"predefined": true,
	"specific_enforcers": [],
	"traffic_destination": {
		"ips_or_subnets": [],
		"network_ids": [],
		"ports": [],
		"specific_mac_addresses": ["02:00:00:00:00:02", "ff:ff:ff:ff:ff:ff"],
		"type": "CLIENT_MAC"
	},
	"traffic_source": {
		"ips_or_subnets": [],
		"network_ids": [],
		"ports": [],
		"specific_mac_addresses": ["02:00:00:00:00:01"],
		"type": "CLIENT_MAC"
	},
	"type": "MAC"
}`

// TestAclRuleRoundTripsStoredRule decodes a stored rule and encodes it again. Every key
// with a value comes back unchanged. The empty IP, network and port lists stay out, as
// in the body the ACL rule editor of the UI sends for a MAC rule.
func TestAclRuleRoundTripsStoredRule(t *testing.T) {
	var rule AclRule
	if err := json.Unmarshal([]byte(storedACLRule), &rule); err != nil {
		t.Fatal(err)
	}

	got, err := json.Marshal(&rule)
	if err != nil {
		t.Fatal(err)
	}
	assert.JSONEq(t, `{
		"_id": "obj_00000000-0000-0000-0000-000000000000_1",
		"acl_index": 1,
		"action": "ALLOW",
		"description": "Automatically created from the \"tv\" Object.",
		"enabled": true,
		"external_id": "00000000-0000-0000-0000-000000000001",
		"ip_acl_protocol": "ALL",
		"name": "tv",
		"origin_id": "00000000-0000-0000-0000-000000000000",
		"origin_type": "object_acl_rule",
		"predefined": true,
		"specific_enforcers": [],
		"traffic_destination": {
			"specific_mac_addresses": ["02:00:00:00:00:02", "ff:ff:ff:ff:ff:ff"],
			"type": "CLIENT_MAC"
		},
		"traffic_source": {
			"specific_mac_addresses": ["02:00:00:00:00:01"],
			"type": "CLIENT_MAC"
		},
		"type": "MAC"
	}`, string(got))
}

// TestAclRuleMarshalSendsEmptyLists guards the two lists where an empty value means
// something: no enforcers puts the rule on every switch, and no destination MACs
// matches every device.
func TestAclRuleMarshalSendsEmptyLists(t *testing.T) {
	index := int64(2)
	got, err := json.Marshal(&AclRule{
		AclIndex:          &index,
		Action:            "BLOCK",
		Enabled:           true,
		MACAclNetworkID:   "net1",
		Name:              "block",
		SpecificEnforcers: []string{},
		TrafficSource: &AclRuleTrafficSource{
			Type:                 "CLIENT_MAC",
			SpecificMACAddresses: []string{"02:00:00:00:00:01"},
		},
		TrafficDestination: &AclRuleTrafficDestination{
			Type:                 "CLIENT_MAC",
			SpecificMACAddresses: []string{},
		},
		Type: "MAC",
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.JSONEq(t, `{
		"acl_index": 2,
		"action": "BLOCK",
		"enabled": true,
		"mac_acl_network_id": "net1",
		"name": "block",
		"predefined": false,
		"specific_enforcers": [],
		"traffic_source": {"specific_mac_addresses": ["02:00:00:00:00:01"], "type": "CLIENT_MAC"},
		"traffic_destination": {"specific_mac_addresses": [], "type": "CLIENT_MAC"},
		"type": "MAC"
	}`, string(got))
}

// TestDeleteAclRulePostsBatchDelete locks the delete to the call the UI makes: a POST of
// the id list to acl-rules/batch-delete.
func TestDeleteAclRulePostsBatchDelete(t *testing.T) {
	var method, path, body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handleNewStyleSetup(w, r) {
			return
		}
		if r.Method == http.MethodPost && r.URL.Path == loginPathNew {
			w.Header().Set("X-Csrf-Token", "tok")
			w.WriteHeader(http.StatusOK)
			return
		}
		b, _ := io.ReadAll(r.Body)
		method, path, body = r.Method, r.URL.Path, string(b)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	c, err := New(context.Background(), &Config{BaseURL: srv.URL, Username: "admin", Password: "admin"})
	if err != nil {
		t.Fatalf("client init: %v", err)
	}
	if err := c.DeleteAclRule(context.Background(), "default", "rule1"); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.MethodPost, method)
	assert.Equal(t, "/proxy/network/v2/api/site/default/acl-rules/batch-delete", path)
	assert.JSONEq(t, `["rule1"]`, body)
}

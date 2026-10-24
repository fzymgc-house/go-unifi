// Code generated from ace.jar fields *.json files
// DO NOT EDIT.

package unifi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/ubiquiti-community/go-unifi/unifi/types"
)

// just to fix compile issues with the import.
var (
	_ context.Context
	_ fmt.Formatter
	_ json.Marshaler
	_ types.Number
	_ strconv.NumError
	_ strings.Builder
)

type AclRule struct {
	ID     string `json:"_id,omitempty"`
	SiteID string `json:"site_id,omitempty"`

	Hidden   bool   `json:"attr_hidden,omitempty"`
	HiddenID string `json:"attr_hidden_id,omitempty"`
	NoDelete bool   `json:"attr_no_delete,omitempty"`
	NoEdit   bool   `json:"attr_no_edit,omitempty"`

	AclIndex           *int64                     `json:"acl_index,omitempty"`
	Action             string                     `json:"action,omitempty"` // ALLOW|BLOCK
	Description        string                     `json:"description,omitempty"`
	Enabled            bool                       `json:"enabled"`
	ExternalID         string                     `json:"external_id,omitempty"`
	IPAclProtocol      string                     `json:"ip_acl_protocol,omitempty"` // ALL|TCP|UDP|TCP_AND_UDP
	MACAclNetworkID    string                     `json:"mac_acl_network_id,omitempty"`
	Name               string                     `json:"name,omitempty"`
	OriginID           string                     `json:"origin_id,omitempty"`
	OriginType         string                     `json:"origin_type,omitempty"`
	Predefined         bool                       `json:"predefined"`
	SpecificEnforcers  []string                   `json:"specific_enforcers"` // ^([0-9A-Fa-f]{2}[:]){5}([0-9A-Fa-f]{2})$
	TrafficDestination *AclRuleTrafficDestination `json:"traffic_destination,omitempty"`
	TrafficSource      *AclRuleTrafficSource      `json:"traffic_source,omitempty"`
	Type               string                     `json:"type,omitempty"` // MAC|IPV4|IPV6
}

func (dst *AclRule) UnmarshalJSON(b []byte) error {
	type Alias AclRule
	aux := &struct {
		AclIndex *types.Number `json:"acl_index"`

		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}
	if aux.AclIndex != nil {
		if val, err := aux.AclIndex.Int64(); err == nil {
			dst.AclIndex = &val
		} else if string(*aux.AclIndex) == "" {
			var zero int64
			dst.AclIndex = &zero
		}
	}

	return nil
}

type AclRuleTrafficDestination struct {
	IPsOrSubnets         []string `json:"ips_or_subnets,omitempty"`
	MACMask              string   `json:"mac_mask,omitempty"` // ^([0-9A-Fa-f]{2}[:]){5}([0-9A-Fa-f]{2})$
	NetworkIDs           []string `json:"network_ids,omitempty"`
	Ports                []int64  `json:"ports,omitempty"`
	SpecificMACAddresses []string `json:"specific_mac_addresses"` // ^([0-9A-Fa-f]{2}[:]){5}([0-9A-Fa-f]{2})$
	Type                 string   `json:"type,omitempty"`         // CLIENT_MAC|IP_OR_SUBNET|NETWORK
}

func (dst *AclRuleTrafficDestination) UnmarshalJSON(b []byte) error {
	type Alias AclRuleTrafficDestination
	aux := &struct {
		Ports []types.Number `json:"ports"`

		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}
	dst.Ports = make([]int64, len(aux.Ports))
	for i, v := range aux.Ports {
		if val, err := v.Int64(); err == nil {
			dst.Ports[i] = val
		}
	}

	return nil
}

type AclRuleTrafficSource struct {
	IPsOrSubnets         []string `json:"ips_or_subnets,omitempty"`
	MACMask              string   `json:"mac_mask,omitempty"` // ^([0-9A-Fa-f]{2}[:]){5}([0-9A-Fa-f]{2})$
	NetworkIDs           []string `json:"network_ids,omitempty"`
	Ports                []int64  `json:"ports,omitempty"`
	SpecificMACAddresses []string `json:"specific_mac_addresses"` // ^([0-9A-Fa-f]{2}[:]){5}([0-9A-Fa-f]{2})$
	Type                 string   `json:"type,omitempty"`         // CLIENT_MAC|IP_OR_SUBNET|NETWORK
}

func (dst *AclRuleTrafficSource) UnmarshalJSON(b []byte) error {
	type Alias AclRuleTrafficSource
	aux := &struct {
		Ports []types.Number `json:"ports"`

		*Alias
	}{
		Alias: (*Alias)(dst),
	}

	err := json.Unmarshal(b, &aux)
	if err != nil {
		return fmt.Errorf("unable to unmarshal alias: %w", err)
	}
	dst.Ports = make([]int64, len(aux.Ports))
	for i, v := range aux.Ports {
		if val, err := v.Int64(); err == nil {
			dst.Ports[i] = val
		}
	}

	return nil
}

func (c *ApiClient) listAclRule(
	ctx context.Context,
	site string,
	query ...map[string]string,
) ([]AclRule, error) {
	var respBody []AclRule

	err := c.do(
		ctx,
		http.MethodGet,
		fmt.Sprintf("v2/api/site/%s/acl-rules", site),
		nil,
		&respBody,
		query...,
	)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

func (c *ApiClient) getAclRule(
	ctx context.Context,
	site string,
	id string,
) (*AclRule, error) {
	respBody, err := c.listAclRule(ctx, site)
	if err != nil {
		return nil, err
	}

	if len(respBody) == 0 {
		return nil, &NotFoundError{}
	}

	for _, val := range respBody {
		if val.ID == id {
			return &val, nil
		}
	}

	return nil, &NotFoundError{}
}

func (c *ApiClient) deleteAclRule(
	ctx context.Context,
	site string,
	id string,
) error {
	err := c.do(
		ctx,
		http.MethodPost,
		fmt.Sprintf("v2/api/site/%s/acl-rules/batch-delete", site),
		[]string{id},
		nil,
	)
	if err != nil {
		return err
	}
	return nil
}

func (c *ApiClient) createAclRule(
	ctx context.Context,
	site string,
	d *AclRule,
) (*AclRule, error) {
	var respBody AclRule

	err := c.do(
		ctx,
		http.MethodPost,
		fmt.Sprintf("v2/api/site/%s/acl-rules", site),
		d,
		&respBody,
	)
	if err != nil {
		return nil, err
	}

	return &respBody, nil
}

func (c *ApiClient) updateAclRule(
	ctx context.Context,
	site string,
	d *AclRule,
) (*AclRule, error) {
	var respBody AclRule
	err := c.do(
		ctx,
		http.MethodPut,
		fmt.Sprintf("v2/api/site/%s/acl-rules/%s", site, d.ID),
		d,
		&respBody,
	)
	if err != nil {
		return nil, err
	}

	return &respBody, nil
}

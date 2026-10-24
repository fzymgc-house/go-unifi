// Code generated from ace.jar fields *.json files
// DO NOT EDIT.

package unifi

import (
	"context"
)

func (c *ApiClient) ListAclRule(ctx context.Context, site string) ([]AclRule, error) {
	return c.listAclRule(ctx, site)
}

func (c *ApiClient) GetAclRule(
	ctx context.Context,
	site,
	id string,
) (*AclRule, error) {
	return c.getAclRule(ctx, site, id)
}

func (c *ApiClient) DeleteAclRule(ctx context.Context, site, id string) error {
	return c.deleteAclRule(ctx, site, id)
}

func (c *ApiClient) CreateAclRule(ctx context.Context, site string, d *AclRule) (*AclRule, error) {
	return c.createAclRule(ctx, site, d)
}

func (c *ApiClient) UpdateAclRule(ctx context.Context, site string, d *AclRule) (*AclRule, error) {
	return c.updateAclRule(ctx, site, d)
}

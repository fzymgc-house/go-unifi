package unifi

import "context"

// Do performs an authenticated request against the controller API and decodes
// the JSON response into respBody.
//
// relativeURL is resolved like every generated method's path: a path without a
// leading slash (e.g. "api/s/default/stat/device/<mac>") is joined to the API
// prefix, and routing through the Cloud Connector applies when configured.
//
// Pass *json.RawMessage (or map[string]any) as respBody, and json.RawMessage or
// a map as reqBody, to round-trip fields this package does not model: typed
// structs drop unknown keys on decode, so a read-modify-write through them
// deletes those keys from the controller when the endpoint replaces the whole
// object or array on PUT.
func (c *ApiClient) Do(
	ctx context.Context,
	method, relativeURL string,
	reqBody any,
	respBody any,
	query ...map[string]string,
) error {
	return c.do(ctx, method, relativeURL, reqBody, respBody, query...)
}

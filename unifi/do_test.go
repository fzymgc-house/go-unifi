package unifi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// A port override carrying keys DevicePortOverrides does not model. Do must
// round-trip them byte-for-byte in both directions.
const doTestOverrides = `[{"port_idx":6,"op_mode":"aggregate","aggregate_members":[6,7],"lag_idx":1},` +
	`{"port_idx":12,"stp_edge_state":"disabled","x_future_key":{"nested":[1,2]}}]`

func TestDo_RoundTripsUnmodelledFields(t *testing.T) {
	devicePath := "/proxy/network/api/s/default/rest/device/abc123"
	var putBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handleNewStyleSetup(w, r) {
			return
		}
		if r.Method == http.MethodPost && r.URL.Path == loginPathNew {
			w.Header().Set("X-Csrf-Token", "tok")
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path != devicePath {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[{"_id":"abc123","port_overrides":` +
				doTestOverrides + `}]}`))
		case http.MethodPut:
			putBody, _ = io.ReadAll(r.Body)
			_, _ = w.Write([]byte(`{"meta":{"rc":"ok"},"data":[]}`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(srv.Close)

	c, err := New(context.Background(), &Config{BaseURL: srv.URL, Username: "admin", Password: "admin"})
	if err != nil {
		t.Fatalf("client init: %v", err)
	}

	var got struct {
		Data []struct {
			PortOverrides json.RawMessage `json:"port_overrides"`
		} `json:"data"`
	}
	if err := c.Do(context.Background(), http.MethodGet, "api/s/default/rest/device/abc123", nil, &got); err != nil {
		t.Fatalf("GET: %v", err)
	}
	if len(got.Data) != 1 || string(got.Data[0].PortOverrides) != doTestOverrides {
		t.Fatalf("GET port_overrides = %s, want %s", got.Data[0].PortOverrides, doTestOverrides)
	}

	req := map[string]json.RawMessage{"port_overrides": got.Data[0].PortOverrides}
	if err := c.Do(context.Background(), http.MethodPut, "api/s/default/rest/device/abc123", req, nil); err != nil {
		t.Fatalf("PUT: %v", err)
	}
	if want := `{"port_overrides":` + doTestOverrides + `}`; string(putBody) != want {
		t.Fatalf("PUT body = %s, want %s", putBody, want)
	}
}

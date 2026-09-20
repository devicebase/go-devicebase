package devicebase

import (
	"net/http"
	"testing"
)

// listResponse is a real GET /v1/devices body: {code, message, data:[...]},
// with no trace_id and no pagination metadata.
const listResponse = `{
	"code": 200,
	"message": "success",
	"data": [
		{
			"id": 10,
			"serialno": "db-mtthisv311f1",
			"device_sn": "f3ad1396-4fb9-4037-81e7-7496f261f3f4",
			"state": "free",
			"name": "Richie-Macbook-Air-7.local",
			"alias_name": "Richie-Macbook-Air-7.local",
			"udid": "b2:da:29:43:81:03",
			"type": "computer",
			"brand": "Apple",
			"model": "MacBook Air",
			"os_type": "macOS",
			"os_version": "26.0",
			"display": "1470x956",
			"location": "Beijing",
			"operator": "CMCC",
			"network": "wifi",
			"updated_at": "2026-09-20T15:06:19Z"
		},
		{
			"id": 11,
			"serialno": "db-mtsi49bf0mqb",
			"device_sn": "8b1c0e2a-0000-4000-8000-000000000001",
			"state": "free",
			"name": "Browser-9222",
			"type": "browser",
			"os_type": "Chrome"
		}
	]
}`

func TestListDevicesDecodesRows(t *testing.T) {
	r := newRecorder(t)
	defer r.server.Close()
	r.response = listResponse

	devices, err := r.client().ListDevices(ListDevicesRequest{})
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}
	if len(devices) != 2 {
		t.Fatalf("len(devices) = %d, want 2", len(devices))
	}

	first := devices[0]
	if first.Serialno != "db-mtthisv311f1" {
		t.Errorf("Serialno = %q", first.Serialno)
	}
	if first.Type != "computer" {
		t.Errorf("Type = %q", first.Type)
	}
	if first.OSType != "macOS" {
		t.Errorf("OSType = %q", first.OSType)
	}
	if first.State != "free" {
		t.Errorf("State = %q", first.State)
	}
	if first.ID != 10 {
		t.Errorf("ID = %d", first.ID)
	}
}

func TestListDevicesSendsItsFilters(t *testing.T) {
	tests := []struct {
		name      string
		request   ListDevicesRequest
		wantQuery map[string]string
		wantEmpty []string
	}{
		{
			name:      "no filters sends an unadorned path",
			request:   ListDevicesRequest{},
			wantEmpty: []string{"keyword", "state", "type", "limit"},
		},
		{
			name:      "category bucket",
			request:   ListDevicesRequest{Type: "browser"},
			wantQuery: map[string]string{"type": "browser"},
			wantEmpty: []string{"keyword", "state", "limit"},
		},
		{
			name: "all filters",
			request: ListDevicesRequest{
				Keyword: "Samsung",
				State:   "free",
				Type:    "mobile",
				Limit:   50,
			},
			wantQuery: map[string]string{
				"keyword": "Samsung",
				"state":   "free",
				"type":    "mobile",
				"limit":   "50",
			},
		},
		{
			name:      "a zero limit is omitted",
			request:   ListDevicesRequest{Limit: 0},
			wantEmpty: []string{"limit"},
		},
		{
			name:      "a negative limit is omitted",
			request:   ListDevicesRequest{Limit: -1},
			wantEmpty: []string{"limit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRecorder(t)
			defer r.server.Close()
			r.response = `{"code":200,"message":"success","data":[]}`

			if _, err := r.client().ListDevices(tt.request); err != nil {
				t.Fatalf("ListDevices: %v", err)
			}

			got := r.last(t)
			if got.Method != http.MethodGet {
				t.Errorf("Method = %q, want GET", got.Method)
			}
			if got.Path != "/v1/devices" {
				t.Errorf("Path = %q, want /v1/devices", got.Path)
			}

			for key, want := range tt.wantQuery {
				if got.Query.Get(key) != want {
					t.Errorf("query %s = %q, want %q", key, got.Query.Get(key), want)
				}
			}
			for _, key := range tt.wantEmpty {
				if _, present := got.Query[key]; present {
					t.Errorf("query %s should have been omitted, got %q", key, got.Query.Get(key))
				}
			}
		})
	}
}

func TestListDevicesSurfacesAnEnvelopeError(t *testing.T) {
	r := newRecorder(t)
	defer r.server.Close()
	r.response = `{"code":401,"message":"API Key 不存在或已停用"}`

	_, err := r.client().ListDevices(ListDevicesRequest{})
	if err == nil {
		t.Fatal("expected the envelope's non-2xx code to raise an error")
	}
	if _, ok := err.(*BusinessError); !ok {
		t.Fatalf("error type = %T, want *BusinessError", err)
	}
}

func TestListDevicesOnAnEmptyList(t *testing.T) {
	r := newRecorder(t)
	defer r.server.Close()
	r.response = `{"code":200,"message":"success","data":[]}`

	devices, err := r.client().ListDevices(ListDevicesRequest{Type: "other"})
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}
	if len(devices) != 0 {
		t.Errorf("len(devices) = %d, want 0", len(devices))
	}
}

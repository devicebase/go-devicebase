package devicebase

import (
	"encoding/json"
	"testing"
)

func TestPointMarshals(t *testing.T) {
	data, err := json.Marshal(Point{X: 100, Y: 200})
	if err != nil {
		t.Fatalf("marshal Point: %v", err)
	}
	want := `{"x":100,"y":200}`
	if string(data) != want {
		t.Errorf("marshal = %s, want %s", data, want)
	}
}

func TestBoundsMarshals(t *testing.T) {
	data, err := json.Marshal(Bounds{X1: 0, Y1: 100, X2: 200, Y2: 300})
	if err != nil {
		t.Fatalf("marshal Bounds: %v", err)
	}
	want := `{"x1":0,"y1":100,"x2":200,"y2":300}`
	if string(data) != want {
		t.Errorf("marshal = %s, want %s", data, want)
	}
}

func TestLaunchAppRequestMarshals(t *testing.T) {
	data, err := json.Marshal(LaunchAppRequest{AppName: "com.tencent.mm"})
	if err != nil {
		t.Fatalf("marshal LaunchAppRequest: %v", err)
	}
	want := `{"app_name":"com.tencent.mm"}`
	if string(data) != want {
		t.Errorf("marshal = %s, want %s", data, want)
	}
}

func TestInputTextRequestMarshals(t *testing.T) {
	data, err := json.Marshal(InputTextRequest{Text: "hello world"})
	if err != nil {
		t.Fatalf("marshal InputTextRequest: %v", err)
	}
	want := `{"text":"hello world"}`
	if string(data) != want {
		t.Errorf("marshal = %s, want %s", data, want)
	}
}

// The optional fields are omitted when zero so the server applies its own
// defaults rather than receiving an explicit zero.
func TestOptionalFieldsOmittedWhenZero(t *testing.T) {
	tests := []struct {
		name string
		body any
		want string
	}{
		{
			name: "computer click without a button",
			body: ComputerClickRequest{X: 10, Y: 20},
			want: `{"x":10,"y":20}`,
		},
		{
			name: "computer click with a button",
			body: ComputerClickRequest{X: 10, Y: 20, Button: ButtonRight},
			want: `{"x":10,"y":20,"button":"right"}`,
		},
		{
			name: "long click without a duration",
			body: ComputerLongClickRequest{X: 1, Y: 2},
			want: `{"x":1,"y":2}`,
		},
		{
			name: "long click with a duration",
			body: ComputerLongClickRequest{X: 1, Y: 2, Duration: 3},
			want: `{"x":1,"y":2,"duration":3}`,
		},
		{
			name: "scroll without an amount",
			body: ScrollRequest{Direction: ScrollDown},
			want: `{"direction":"down"}`,
		},
		{
			name: "scroll with an amount",
			body: ScrollRequest{Direction: ScrollDown, Amount: 5},
			want: `{"direction":"down","amount":5}`,
		},
		{
			// A zero timeout must omit the field: sending 0 would ask the
			// server for a zero-second budget instead of its 120s default.
			name: "bash without a timeout",
			body: ComputerBashRequest{Command: "ls"},
			want: `{"command":"ls"}`,
		},
		{
			name: "bash with a timeout",
			body: ComputerBashRequest{Command: "ls", Timeout: 30},
			want: `{"command":"ls","timeout":30}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(data) != tt.want {
				t.Errorf("marshal = %s, want %s", data, tt.want)
			}
		})
	}
}

func TestWaitRequestSendsSeconds(t *testing.T) {
	data, err := json.Marshal(WaitRequest{Seconds: 2.5})
	if err != nil {
		t.Fatalf("marshal WaitRequest: %v", err)
	}
	want := `{"seconds":2.5}`
	if string(data) != want {
		t.Errorf("marshal = %s, want %s", data, want)
	}
}

func TestDeviceDecodesAListRow(t *testing.T) {
	// A real row from GET /v1/devices, trimmed to the fields the SDK models.
	row := `{
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
	}`

	var device Device
	if err := json.Unmarshal([]byte(row), &device); err != nil {
		t.Fatalf("unmarshal Device: %v", err)
	}

	if device.Serialno != "db-mtthisv311f1" {
		t.Errorf("Serialno = %q", device.Serialno)
	}
	if device.DeviceSN != "f3ad1396-4fb9-4037-81e7-7496f261f3f4" {
		t.Errorf("DeviceSN = %q", device.DeviceSN)
	}
	if device.Type != "computer" {
		t.Errorf("Type = %q", device.Type)
	}
	if device.OSType != "macOS" {
		t.Errorf("OSType = %q", device.OSType)
	}
	if device.ID != 10 {
		t.Errorf("ID = %d", device.ID)
	}
}

func TestOperationResultSuccessField(t *testing.T) {
	result := &OperationResult{Success: true, Data: map[string]any{"success": true}}
	if !result.Success {
		t.Error("Success should be true")
	}
}

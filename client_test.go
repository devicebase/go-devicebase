package devicebase

import (
	"net/http"
	"testing"
	"time"
)

func TestNewClientDefaults(t *testing.T) {
	t.Setenv("DEVICEBASE_API_KEY", "env-key")
	t.Setenv(envBaseURL, "")

	client := NewClient(WithSerial("device123"))
	if client.Serial() != "device123" {
		t.Errorf("serial = %q, want %q", client.Serial(), "device123")
	}
	if client.http.apiKey != "env-key" {
		t.Errorf("apiKey = %q, want %q", client.http.apiKey, "env-key")
	}
	if client.http.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", client.http.baseURL, defaultBaseURL)
	}
	if client.http.hc.Timeout != defaultTimeout {
		t.Errorf("timeout = %v, want %v", client.http.hc.Timeout, defaultTimeout)
	}
}

func TestNewClientReadsBaseURLFromEnv(t *testing.T) {
	t.Setenv(envAPIKey, "env-key")
	t.Setenv(envBaseURL, "http://127.0.0.1:8000")

	client := NewClient()
	if client.http.baseURL != "http://127.0.0.1:8000" {
		t.Errorf("baseURL = %q", client.http.baseURL)
	}
}

func TestNewClientOverrides(t *testing.T) {
	client := NewClient(
		WithAPIKey("my-key"),
		WithSerial("serial-1"),
		WithBaseURL("http://localhost:8080/"),
		WithTimeout(5*time.Second),
	)
	if client.http.apiKey != "my-key" {
		t.Errorf("apiKey = %q", client.http.apiKey)
	}
	if client.Serial() != "serial-1" {
		t.Errorf("serial = %q", client.Serial())
	}
	// A trailing slash would turn every appended path into "//v1/…".
	if client.http.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL = %q, want the trailing slash trimmed", client.http.baseURL)
	}
	if client.http.hc.Timeout != 5*time.Second {
		t.Errorf("timeout = %v", client.http.hc.Timeout)
	}
}

// A client can be built before the environment is ready; the missing key
// surfaces from the first request instead of panicking at construction.
func TestNewClientWithoutAKeyFailsOnFirstUse(t *testing.T) {
	t.Setenv(envAPIKey, "")
	t.Setenv(envBaseURL, "http://127.0.0.1:1")

	client := NewClient(WithSerial("s"))
	if _, err := client.Back(); err == nil {
		t.Fatal("expected an AuthenticationError")
	}
}

// --- Mobile actions -------------------------------------------------------

const mobileSerial = "db-mttul4i41di8"

func mobileCases() []callCase {
	return []callCase{
		{
			name:       "get device info",
			call:       func(c *Client) error { _, err := c.GetDeviceInfo(); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/deviceinfo/" + mobileSerial,
		},
		{
			name:       "tap",
			call:       func(c *Client) error { _, err := c.Tap(100, 200); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/tap/" + mobileSerial,
			wantBody:   `{"x":100,"y":200}`,
		},
		{
			name:       "double tap",
			call:       func(c *Client) error { _, err := c.DoubleTap(10, 20); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/double_tap/" + mobileSerial,
			wantBody:   `{"x":10,"y":20}`,
		},
		{
			name:       "long press",
			call:       func(c *Client) error { _, err := c.LongPress(30, 40); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/long_press/" + mobileSerial,
			wantBody:   `{"x":30,"y":40}`,
		},
		{
			name:       "swipe",
			call:       func(c *Client) error { _, err := c.Swipe(0, 100, 300, 100); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/swipe/" + mobileSerial,
			wantBody:   `{"x1":0,"y1":100,"x2":300,"y2":100}`,
		},
		{
			name:       "back",
			call:       func(c *Client) error { _, err := c.Back(); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/back/" + mobileSerial,
		},
		{
			name:       "home",
			call:       func(c *Client) error { _, err := c.Home(); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/home/" + mobileSerial,
		},
		{
			name:       "launch app",
			call:       func(c *Client) error { _, err := c.LaunchApp("com.tencent.mm"); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/launch_app/" + mobileSerial,
			wantBody:   `{"app_name":"com.tencent.mm"}`,
		},
		{
			name:       "stop app",
			call:       func(c *Client) error { _, err := c.StopApp("com.tencent.mm"); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/stop_app/" + mobileSerial,
			wantBody:   `{"app_name":"com.tencent.mm"}`,
		},
		{
			name:       "stop current app",
			call:       func(c *Client) error { _, err := c.StopCurrentApp(); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/stop_current_app/" + mobileSerial,
		},
		{
			name:       "current app",
			call:       func(c *Client) error { _, err := c.GetCurrentApp(); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/current_app/" + mobileSerial,
		},
		{
			name:       "input text",
			call:       func(c *Client) error { _, err := c.InputText("hello world"); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/input/" + mobileSerial,
			wantBody:   `{"text":"hello world"}`,
		},
		{
			name:       "clear text",
			call:       func(c *Client) error { _, err := c.ClearText(); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/clear_text/" + mobileSerial,
		},
		{
			name:       "bash",
			call:       func(c *Client) error { _, err := c.Bash("ls -la"); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/bash/" + mobileSerial,
			wantBody:   `{"command":"ls -la"}`,
		},
		{
			name:       "dump hierarchy",
			call:       func(c *Client) error { _, err := c.DumpHierarchy(); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/dump_hierarchy/" + mobileSerial,
		},
		{
			name:       "install app",
			call:       func(c *Client) error { _, err := c.InstallApp("/tmp/app.apk"); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/install_app/" + mobileSerial,
			wantBody:   `{"app_path":"/tmp/app.apk"}`,
		},
		{
			name:       "install status",
			call:       func(c *Client) error { _, err := c.InstallStatus("install-42"); return err },
			wantMethod: http.MethodGet,
			wantPath:   "/v1/install_status/" + mobileSerial,
			wantQuery:  map[string]string{"install_id": "install-42"},
		},
		{
			// The CLI posts to /v1/screen; GET is not part of the contract.
			name:       "screenshot is a POST",
			call:       func(c *Client) error { _, err := c.GetScreenshot(); return err },
			wantMethod: http.MethodPost,
			wantPath:   "/v1/screen/" + mobileSerial,
		},
	}
}

func TestMobileActions(t *testing.T) {
	runCalls(t, []Option{WithSerial(mobileSerial)}, mobileCases())
}

func TestGetDeviceInfoReturnsTheBoundSerial(t *testing.T) {
	r := newRecorder(t)
	defer r.server.Close()
	r.response = `{"device":{"serialno":"` + mobileSerial + `"}}`

	info, err := r.client(WithSerial(mobileSerial)).GetDeviceInfo()
	if err != nil {
		t.Fatalf("GetDeviceInfo: %v", err)
	}
	if info.Serial != mobileSerial {
		t.Errorf("Serial = %q, want %q", info.Serial, mobileSerial)
	}
	if info.Data["device"] == nil {
		t.Error("Data should carry the response payload")
	}
}

func TestGetScreenshotReturnsBytes(t *testing.T) {
	jpeg := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	r := newRecorder(t)
	defer r.server.Close()
	r.response = string(jpeg)

	data, err := r.client(WithSerial(mobileSerial)).GetScreenshot()
	if err != nil {
		t.Fatalf("GetScreenshot: %v", err)
	}
	if string(data) != string(jpeg) {
		t.Errorf("data = %v, want %v", data, jpeg)
	}
}

// Success mirrors a top-level "success" field when the server sends one. The
// control API's envelope has no such field — it reports failures through a
// non-2xx "code", which the transport raises as a BusinessError — so the flag
// defaults to true and only an explicit false flips it.
func TestOperationResultReadsTheSuccessFlag(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		success bool
	}{
		{"explicit false", `{"success":false}`, false},
		{"explicit true", `{"success":true}`, true},
		{"absent defaults to true", `{"code":200,"message":"success","data":{}}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRecorder(t)
			defer r.server.Close()
			r.response = tt.body

			result, err := r.client(WithSerial(mobileSerial)).Back()
			if err != nil {
				t.Fatalf("Back: %v", err)
			}
			if result.Success != tt.success {
				t.Errorf("Success = %v, want %v", result.Success, tt.success)
			}
		})
	}
}

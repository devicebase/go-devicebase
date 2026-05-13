package devicebase

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockServer struct {
	server *httptest.Server
	method string
	path   string
	body   map[string]any
}

func newMockServer(handler http.HandlerFunc) *mockServer {
	return &mockServer{
		server: httptest.NewServer(handler),
	}
}

func TestNewClientDefaults(t *testing.T) {
	t.Setenv("DEVICEBASE_API_KEY", "env-key")
	client := NewClient(WithSerial("device123"))
	if client.serial != "device123" {
		t.Errorf("serial = %q, want %q", client.serial, "device123")
	}
	if client.http.apiKey != "env-key" {
		t.Errorf("apiKey = %q, want %q", client.http.apiKey, "env-key")
	}
	if client.http.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", client.http.baseURL, defaultBaseURL)
	}
}

func TestNewClientOverrides(t *testing.T) {
	client := NewClient(
		WithAPIKey("my-key"),
		WithSerial("serial-1"),
		WithBaseURL("http://localhost:8080"),
	)
	if client.http.apiKey != "my-key" {
		t.Errorf("apiKey = %q", client.http.apiKey)
	}
	if client.serial != "serial-1" {
		t.Errorf("serial = %q", client.serial)
	}
	if client.http.baseURL != "http://localhost:8080" {
		t.Errorf("baseURL = %q", client.http.baseURL)
	}
}

func TestGetDeviceInfo(t *testing.T) {
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		if r.URL.Path != "/v1/deviceinfo/abc123" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"serial":  "abc123",
			"status":  "online",
			"brand":   "Samsung",
		})
	})
	defer ms.server.Close()

	client := NewClient(
		WithAPIKey("key"),
		WithSerial("abc123"),
		WithBaseURL(ms.server.URL),
	)

	info, err := client.GetDeviceInfo()
	if err != nil {
		t.Fatalf("GetDeviceInfo: %v", err)
	}
	if info.Serial != "abc123" {
		t.Errorf("Serial = %q", info.Serial)
	}
	if info.Data["status"] != "online" {
		t.Errorf("Data[status] = %v", info.Data["status"])
	}
}

func TestTap(t *testing.T) {
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/tap/dev1" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		json.Unmarshal(body, &got)
		if got["x"] != float64(100) || got["y"] != float64(200) {
			t.Errorf("body = %v", got)
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	defer ms.server.Close()

	client := NewClient(WithAPIKey("key"), WithSerial("dev1"), WithBaseURL(ms.server.URL))
	result, err := client.Tap(100, 200)
	if err != nil {
		t.Fatalf("Tap: %v", err)
	}
	if !result.Success {
		t.Error("Success should be true")
	}
}

func TestDoubleTap(t *testing.T) {
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/double_tap/dev1" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	defer ms.server.Close()

	client := NewClient(WithAPIKey("key"), WithSerial("dev1"), WithBaseURL(ms.server.URL))
	result, err := client.DoubleTap(50, 75)
	if err != nil {
		t.Fatalf("DoubleTap: %v", err)
	}
	if !result.Success {
		t.Error("Success should be true")
	}
}

func TestLongPress(t *testing.T) {
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/long_press/dev1" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	defer ms.server.Close()

	client := NewClient(WithAPIKey("key"), WithSerial("dev1"), WithBaseURL(ms.server.URL))
	result, err := client.LongPress(100, 200)
	if err != nil {
		t.Fatalf("LongPress: %v", err)
	}
	if !result.Success {
		t.Error("Success should be true")
	}
}

func TestSwipe(t *testing.T) {
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/swipe/dev1" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		json.Unmarshal(body, &got)
		if got["x1"] != float64(0) || got["y1"] != float64(500) || got["x2"] != float64(500) || got["y2"] != float64(500) {
			t.Errorf("body = %v", got)
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	defer ms.server.Close()

	client := NewClient(WithAPIKey("key"), WithSerial("dev1"), WithBaseURL(ms.server.URL))
	result, err := client.Swipe(0, 500, 500, 500)
	if err != nil {
		t.Fatalf("Swipe: %v", err)
	}
	if !result.Success {
		t.Error("Success should be true")
	}
}

func TestBack(t *testing.T) {
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/back/dev1" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	defer ms.server.Close()

	client := NewClient(WithAPIKey("key"), WithSerial("dev1"), WithBaseURL(ms.server.URL))
	result, err := client.Back()
	if err != nil {
		t.Fatalf("Back: %v", err)
	}
	if !result.Success {
		t.Error("Success should be true")
	}
}

func TestHome(t *testing.T) {
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/home/dev1" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	defer ms.server.Close()

	client := NewClient(WithAPIKey("key"), WithSerial("dev1"), WithBaseURL(ms.server.URL))
	result, err := client.Home()
	if err != nil {
		t.Fatalf("Home: %v", err)
	}
	if !result.Success {
		t.Error("Success should be true")
	}
}

func TestLaunchApp(t *testing.T) {
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/launch_app/dev1" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		json.Unmarshal(body, &got)
		if got["app_name"] != "com.tencent.mm" {
			t.Errorf("body = %v", got)
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	defer ms.server.Close()

	client := NewClient(WithAPIKey("key"), WithSerial("dev1"), WithBaseURL(ms.server.URL))
	result, err := client.LaunchApp("com.tencent.mm")
	if err != nil {
		t.Fatalf("LaunchApp: %v", err)
	}
	if !result.Success {
		t.Error("Success should be true")
	}
}

func TestGetCurrentApp(t *testing.T) {
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/current_app/dev1" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"package": "com.tencent.mm",
		})
	})
	defer ms.server.Close()

	client := NewClient(WithAPIKey("key"), WithSerial("dev1"), WithBaseURL(ms.server.URL))
	app, err := client.GetCurrentApp()
	if err != nil {
		t.Fatalf("GetCurrentApp: %v", err)
	}
	if app.Data["package"] != "com.tencent.mm" {
		t.Errorf("package = %v", app.Data["package"])
	}
}

func TestInputText(t *testing.T) {
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/input/dev1" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]any
		json.Unmarshal(body, &got)
		if got["text"] != "hello" {
			t.Errorf("body = %v", got)
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	defer ms.server.Close()

	client := NewClient(WithAPIKey("key"), WithSerial("dev1"), WithBaseURL(ms.server.URL))
	result, err := client.InputText("hello")
	if err != nil {
		t.Fatalf("InputText: %v", err)
	}
	if !result.Success {
		t.Error("Success should be true")
	}
}

func TestClearText(t *testing.T) {
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/clear_text/dev1" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	})
	defer ms.server.Close()

	client := NewClient(WithAPIKey("key"), WithSerial("dev1"), WithBaseURL(ms.server.URL))
	result, err := client.ClearText()
	if err != nil {
		t.Fatalf("ClearText: %v", err)
	}
	if !result.Success {
		t.Error("Success should be true")
	}
}

func TestDumpHierarchy(t *testing.T) {
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/dump_hierarchy/dev1" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"hierarchy": "<node/>",
		})
	})
	defer ms.server.Close()

	client := NewClient(WithAPIKey("key"), WithSerial("dev1"), WithBaseURL(ms.server.URL))
	info, err := client.DumpHierarchy()
	if err != nil {
		t.Fatalf("DumpHierarchy: %v", err)
	}
	if info.Data["hierarchy"] != "<node/>" {
		t.Errorf("hierarchy = %v", info.Data["hierarchy"])
	}
}

func TestGetScreenshot(t *testing.T) {
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00}
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/v1/screen/dev1" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(jpegData)
	})
	defer ms.server.Close()

	client := NewClient(WithAPIKey("key"), WithSerial("dev1"), WithBaseURL(ms.server.URL))
	data, err := client.GetScreenshot()
	if err != nil {
		t.Fatalf("GetScreenshot: %v", err)
	}
	if string(data) != string(jpegData) {
		t.Errorf("screenshot = %v, want %v", data, jpegData)
	}
}

func TestDownloadScreenshot(t *testing.T) {
	jpegData := []byte{0xFF, 0xD8, 0xFF}
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/v1/screenshot/dev1" {
			t.Errorf("Path = %q", r.URL.Path)
		}
		w.Write(jpegData)
	})
	defer ms.server.Close()

	client := NewClient(WithAPIKey("key"), WithSerial("dev1"), WithBaseURL(ms.server.URL))
	data, err := client.DownloadScreenshot()
	if err != nil {
		t.Fatalf("DownloadScreenshot: %v", err)
	}
	if string(data) != string(jpegData) {
		t.Errorf("data = %v, want %v", data, jpegData)
	}
}

func TestClientErrorPropagation(t *testing.T) {
	ms := newMockServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	})
	defer ms.server.Close()

	client := NewClient(WithAPIKey("key"), WithSerial("dev1"), WithBaseURL(ms.server.URL))
	_, err := client.GetDeviceInfo()
	if err == nil {
		t.Fatal("expected error")
	}
	var dnfe *DeviceNotFoundError
	if !errors.As(err, &dnfe) {
		t.Errorf("error type = %T, want DeviceNotFoundError", err)
	}
}

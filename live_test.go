package devicebase

import (
	"os"
	"testing"
)

// Live smoke tests against a real Devicebase server.
//
// They are skipped unless DEVICEBASE_LIVE_TEST=1, so the default `go test ./...`
// stays hermetic. To run them:
//
//	DEVICEBASE_LIVE_TEST=1 \
//	DEVICEBASE_API_KEY=<key> \
//	DEVICEBASE_BASE_URL=http://127.0.0.1:8000 \
//	go test -run TestLive -v ./...
//
// Only read-only actions are exercised: these tests must not drive a real
// device, because the serials they pick are whatever the server happens to
// have registered.
func requireLive(t *testing.T) *Client {
	t.Helper()

	if os.Getenv("DEVICEBASE_LIVE_TEST") != "1" {
		t.Skip("set DEVICEBASE_LIVE_TEST=1 to run live tests")
	}
	if os.Getenv(envAPIKey) == "" {
		t.Skip(envAPIKey + " is not set")
	}
	return NewClient()
}

// firstDevice returns the first device of the given type, so a live test can
// discover a serialno instead of hard-coding one.
func firstDevice(t *testing.T, client *Client, deviceType string) Device {
	t.Helper()

	devices, err := client.ListDevices(ListDevicesRequest{Type: deviceType})
	if err != nil {
		t.Fatalf("ListDevices(%s): %v", deviceType, err)
	}
	if len(devices) == 0 {
		t.Skipf("no %s device is registered", deviceType)
	}
	return devices[0]
}

func TestLiveListDevices(t *testing.T) {
	client := requireLive(t)

	devices, err := client.ListDevices(ListDevicesRequest{})
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}
	t.Logf("found %d devices", len(devices))

	for _, d := range devices {
		t.Logf("  %-9s %-18s %-8s os_type=%s", d.Type, d.Serialno, d.State, d.OSType)
		if d.Serialno == "" {
			t.Errorf("device %d has no serialno", d.ID)
		}
	}
}

func TestLiveListDevicesFilters(t *testing.T) {
	client := requireLive(t)

	for _, deviceType := range []string{"mobile", "browser", "computer"} {
		devices, err := client.ListDevices(ListDevicesRequest{Type: deviceType})
		if err != nil {
			t.Errorf("ListDevices(type=%s): %v", deviceType, err)
			continue
		}
		t.Logf("%-9s -> %d device(s)", deviceType, len(devices))
	}

	// A limit has to reach the server, not just be dropped.
	limited, err := client.ListDevices(ListDevicesRequest{Limit: 1})
	if err != nil {
		t.Fatalf("ListDevices(limit=1): %v", err)
	}
	if len(limited) > 1 {
		t.Errorf("limit=1 returned %d devices", len(limited))
	}
}

func TestLiveMobileReadOnly(t *testing.T) {
	client := requireLive(t)
	device := firstDevice(t, client, "mobile")
	client = NewClient(WithSerial(device.Serialno))

	if info, err := client.GetDeviceInfo(); err != nil {
		t.Errorf("GetDeviceInfo: %v", err)
	} else {
		t.Logf("device info keys: %d", len(info.Data))
	}

	if _, err := client.GetCurrentApp(); err != nil {
		t.Errorf("GetCurrentApp: %v", err)
	}
	if _, err := client.DumpHierarchy(); err != nil {
		t.Errorf("DumpHierarchy: %v", err)
	}

	data, err := client.GetScreenshot()
	if err != nil {
		t.Fatalf("GetScreenshot: %v", err)
	}
	t.Logf("screenshot: %d bytes", len(data))
	if !looksLikeAnImage(data) {
		t.Errorf("screenshot does not start with a known image signature: % x", data[:min(8, len(data))])
	}
}

func TestLiveComputerReadOnly(t *testing.T) {
	client := requireLive(t)
	device := firstDevice(t, client, "computer")

	if _, err := client.ComputerPosition(device.Serialno); err != nil {
		t.Errorf("ComputerPosition: %v", err)
	}
	if _, err := client.ComputerScreenSize(device.Serialno); err != nil {
		t.Errorf("ComputerScreenSize: %v", err)
	}
	if _, err := client.ComputerPermissions(device.Serialno); err != nil {
		t.Errorf("ComputerPermissions: %v", err)
	}

	// A wait is a no-op on the device; it proves the widened deadline works.
	if _, err := client.ComputerWait(device.Serialno, 500); err != nil {
		t.Errorf("ComputerWait: %v", err)
	}

	// The command's own exit status rides in the payload, not in the error.
	result, err := client.ComputerBash(device.Serialno, "echo live-ok", 10)
	if err != nil {
		t.Fatalf("ComputerBash: %v", err)
	}
	t.Logf("bash payload: %v", result.Data)
}

func TestLiveBrowserReadOnly(t *testing.T) {
	client := requireLive(t)
	device := firstDevice(t, client, "browser")

	if _, err := client.BrowserState(device.Serialno); err != nil {
		t.Errorf("BrowserState: %v", err)
	}
	if _, err := client.BrowserTabs(device.Serialno); err != nil {
		t.Errorf("BrowserTabs: %v", err)
	}

	// A selector that cannot match must still be a well-formed request: the
	// server answers it rather than erroring on the path.
	if _, err := client.BrowserExists(device.Serialno, "#definitely-not-present"); err != nil {
		t.Errorf("BrowserExists: %v", err)
	}
}

// A business failure arrives as HTTP 200 with a non-2xx code; the SDK has to
// raise it rather than hand back a success.
func TestLiveBusinessErrorIsRaised(t *testing.T) {
	client := requireLive(t)

	// An unknown serialno: the gateway answers with the envelope's own error.
	_, err := client.ComputerScreenSize("db-does-not-exist-0000")
	if err == nil {
		t.Fatal("expected an error for an unknown serialno")
	}
	t.Logf("error type %T: %v", err, err)
}

func looksLikeAnImage(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	switch {
	case data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF:
		return true // JPEG
	case data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G':
		return true // PNG
	case data[0] == 'G' && data[1] == 'I' && data[2] == 'F':
		return true // GIF
	}
	return false
}

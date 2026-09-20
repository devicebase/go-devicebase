// Package devicebase is a Go SDK for the Devicebase device-control API.
//
// Three device platforms are covered, each with its own path family:
//
//	mobile   (Android / HarmonyOS / iOS)  /v1/{action}/{serialno}
//	browser  (Chrome / Chromium / Edge)    /api/browser/{serialno}/{action...}
//	computer (macOS / Windows / Linux)     /api/computer/{serialno}/{action}
//
// The mobile methods are bound to the serialno given by WithSerialno; browser,
// computer and list methods take the serialno per call, because a serialno is
// only meaningful within one platform family.
package devicebase

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const envBaseURL = "DEVICEBASE_BASE_URL"
const envAPIKey = "DEVICEBASE_API_KEY"

// Client is the main client for interacting with the DeviceBase API.
//
// It is safe to reuse across devices: the browser, computer and list methods
// take the device's serialno explicitly, so one client can drive many devices.
type Client struct {
	serialno string
	http     *httpClient
}

// Option configures a Client.
type Option func(*Client)

// WithAPIKey sets the API key for authentication.
// Falls back to the DEVICEBASE_API_KEY environment variable if not set.
func WithAPIKey(key string) Option {
	return func(c *Client) { c.http.apiKey = key }
}

// WithSerialno sets the device serialno used by the mobile methods.
//
// The server keys devices by "serialno" (e.g. "db-mttul4i41di8"); the device's
// device_sn UUID resolves too. Devices without a bound serialno can still use the
// browser, computer and list methods, which take a serialno per call.
func WithSerialno(serialno string) Option {
	return func(c *Client) { c.serialno = serialno }
}

// WithSerial is an alias for [WithSerialno].
//
// Deprecated: use WithSerialno. Removed in the next major release.
func WithSerial(serial string) Option { return WithSerialno(serial) }

// WithBaseURL sets the API base URL.
// Falls back to the DEVICEBASE_BASE_URL environment variable, or https://api.devicebase.cn.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		// Trim here as well as in newHTTPClient, or a trailing slash would turn
		// every appended path into "//v1/…".
		c.http.baseURL = strings.TrimRight(url, "/")
	}
}

// WithTimeout sets the default HTTP request timeout (30s when unset).
//
// Blocking actions are not bound by it: the computer bash and wait methods
// raise the deadline per call to cover however long they were asked to run.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.http.hc.Timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client. Its own Timeout wins for the
// default deadline; blocking actions still raise it per call, since an
// unbounded client would otherwise hang forever.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.http.hc = hc
		}
	}
}

// NewClient creates a new DeviceBase client with the given options.
//
// It does not read the API key eagerly — a missing key surfaces from the first
// request as an AuthenticationError, so a client can be constructed before the
// environment is fully set up.
func NewClient(opts ...Option) *Client {
	apiKey := os.Getenv(envAPIKey)
	baseURL := os.Getenv(envBaseURL)
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	c := &Client{
		http: newHTTPClient(baseURL, apiKey, nil),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Serialno returns the serialno bound by [WithSerialno].
func (c *Client) Serialno() string { return c.serialno }

// Serial is an alias for [Client.Serialno].
//
// Deprecated: use Serialno. Removed in the next major release.
func (c *Client) Serial() string { return c.serialno }

// --- Mobile ---------------------------------------------------------------

// GetDeviceInfo returns detailed information about the device.
func (c *Client) GetDeviceInfo() (*DeviceInfo, error) {
	data, err := c.http.doJSON(http.MethodPost, c.mobilePath("deviceinfo"), nil)
	if err != nil {
		return nil, err
	}
	return &DeviceInfo{Serialno: c.serialno, Data: data}, nil
}

// Tap performs a single tap at the specified coordinates.
func (c *Client) Tap(x, y int) (*OperationResult, error) {
	return c.doOperation(c.mobilePath("tap"), Point{X: x, Y: y})
}

// DoubleTap performs a double tap at the specified coordinates.
func (c *Client) DoubleTap(x, y int) (*OperationResult, error) {
	return c.doOperation(c.mobilePath("double_tap"), Point{X: x, Y: y})
}

// LongPress performs a long press at the specified coordinates.
func (c *Client) LongPress(x, y int) (*OperationResult, error) {
	return c.doOperation(c.mobilePath("long_press"), Point{X: x, Y: y})
}

// Swipe performs a swipe gesture from (x1,y1) to (x2,y2).
func (c *Client) Swipe(x1, y1, x2, y2 int) (*OperationResult, error) {
	return c.doOperation(c.mobilePath("swipe"), Bounds{X1: x1, Y1: y1, X2: x2, Y2: y2})
}

// Back simulates the device back button press.
func (c *Client) Back() (*OperationResult, error) {
	return c.doOperation(c.mobilePath("back"), nil)
}

// Home simulates the device home button press.
func (c *Client) Home() (*OperationResult, error) {
	return c.doOperation(c.mobilePath("home"), nil)
}

// LaunchApp launches an application on the device.
func (c *Client) LaunchApp(appName string) (*OperationResult, error) {
	return c.doOperation(c.mobilePath("launch_app"), LaunchAppRequest{AppName: appName})
}

// StopApp stops an application on the device.
func (c *Client) StopApp(appName string) (*OperationResult, error) {
	return c.doOperation(c.mobilePath("stop_app"), LaunchAppRequest{AppName: appName})
}

// StopCurrentApp stops the app currently in the foreground.
func (c *Client) StopCurrentApp() (*OperationResult, error) {
	return c.doOperation(c.mobilePath("stop_current_app"), nil)
}

// GetCurrentApp returns information about the currently running foreground app.
func (c *Client) GetCurrentApp() (*AppInfo, error) {
	data, err := c.http.doJSON(http.MethodPost, c.mobilePath("current_app"), nil)
	if err != nil {
		return nil, err
	}
	return &AppInfo{Data: data}, nil
}

// InputText inputs text into the currently focused field.
func (c *Client) InputText(text string) (*OperationResult, error) {
	return c.doOperation(c.mobilePath("input"), InputTextRequest{Text: text})
}

// ClearText clears text in the currently focused field.
func (c *Client) ClearText() (*OperationResult, error) {
	return c.doOperation(c.mobilePath("clear_text"), nil)
}

// Bash runs a shell command on the device (adb/hdc platforms only).
//
// The command's own exit status comes back in the payload as data.exitCode — a
// non-zero value is not an API error, so this returns no error for it.
func (c *Client) Bash(command string) (*OperationResult, error) {
	return c.doOperation(c.mobilePath("bash"), CommandRequest{Command: command})
}

// DumpHierarchy returns the current UI hierarchy structure.
func (c *Client) DumpHierarchy() (*HierarchyInfo, error) {
	data, err := c.http.doJSON(http.MethodPost, c.mobilePath("dump_hierarchy"), nil)
	if err != nil {
		return nil, err
	}
	return &HierarchyInfo{Data: data}, nil
}

// InstallApp installs a package from a path on the agent host (not a local
// file) and returns an install id for the background task.
func (c *Client) InstallApp(appPath string) (*OperationResult, error) {
	return c.doOperation(c.mobilePath("install_app"), InstallAppRequest{AppPath: appPath})
}

// InstallStatus queries the background install task started by InstallApp.
func (c *Client) InstallStatus(installID string) (*OperationResult, error) {
	query := buildQuery(map[string]string{"install_id": installID})
	return c.doGetOperation(c.mobilePath("install_status") + query)
}

// GetScreenshot returns a screenshot of the device screen as image bytes.
//
// The server decides the format (JPEG for every platform), so the bytes are
// returned without a decoded wrapper. Works for browser and computer serials
// too — the server dispatches /v1/screen by device type.
func (c *Client) GetScreenshot() ([]byte, error) {
	return c.http.doRaw(http.MethodPost, c.mobilePath("screen"))
}

// DownloadScreenshot downloads a screenshot as a file attachment.
//
// This is an SDK-only extra with no equivalent in the Devicebase CLI: it hits
// GET /v1/screenshot/{serialno} rather than the cross-family /v1/screen route.
func (c *Client) DownloadScreenshot() ([]byte, error) {
	return c.http.doRaw(http.MethodGet, fmt.Sprintf("/v1/screenshot/%s", c.serialno))
}

// mobilePath builds /v1/{action}/{serialno} — the mobile route family. The
// control server registers these paths and redirects them to the /api/*
// handlers, so the method and body survive the hop.
func (c *Client) mobilePath(action string) string {
	return fmt.Sprintf("/v1/%s/%s", action, c.serialno)
}

func (c *Client) doOperation(path string, body any) (*OperationResult, error) {
	return c.doOperationWith(c.http, http.MethodPost, path, body)
}

// doGetOperation runs a read-only action. The browser family carries selectors
// as query parameters on GET, so those actions have no body at all.
func (c *Client) doGetOperation(path string) (*OperationResult, error) {
	return c.doOperationWith(c.http, http.MethodGet, path, nil)
}

func (c *Client) doOperationWith(
	transport *httpClient,
	method, path string,
	body any,
) (*OperationResult, error) {
	data, err := transport.doJSON(method, path, body)
	if err != nil {
		return nil, err
	}
	success := true
	if s, ok := data["success"]; ok {
		if b, ok := s.(bool); ok {
			success = b
		}
	}
	return &OperationResult{Success: success, Data: data}, nil
}

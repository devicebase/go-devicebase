package devicebase

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

const envBaseURL = "DEVICEBASE_BASE_URL"
const envAPIKey = "DEVICEBASE_API_KEY"

// Client is the main client for interacting with the DeviceBase API.
type Client struct {
	serial string
	http   *httpClient
}

// Option configures a Client.
type Option func(*Client)

// WithAPIKey sets the API key for authentication.
// Falls back to the DEVICEBASE_API_KEY environment variable if not set.
func WithAPIKey(key string) Option {
	return func(c *Client) { c.http.apiKey = key }
}

// WithSerial sets the device serial number.
func WithSerial(serial string) Option {
	return func(c *Client) { c.serial = serial }
}

// WithBaseURL sets the API base URL.
// Falls back to the DEVICEBASE_BASE_URL environment variable, or https://api.devicebase.cn.
func WithBaseURL(url string) Option {
	return func(c *Client) { c.http.baseURL = url }
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		c.http.httpClient.Timeout = timeout
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) {
		c.http.httpClient = hc
	}
}

// NewClient creates a new DeviceBase client with the given options.
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

// GetDeviceInfo returns detailed information about the device.
func (c *Client) GetDeviceInfo() (*DeviceInfo, error) {
	data, err := c.http.doJSON("POST", fmt.Sprintf("/v1/deviceinfo/%s", c.serial), nil)
	if err != nil {
		return nil, err
	}
	return &DeviceInfo{Serial: c.serial, Data: data}, nil
}

// Tap performs a single tap at the specified coordinates.
func (c *Client) Tap(x, y int) (*OperationResult, error) {
	return c.doOperation(fmt.Sprintf("/v1/tap/%s", c.serial), pointRequest{X: x, Y: y})
}

// DoubleTap performs a double tap at the specified coordinates.
func (c *Client) DoubleTap(x, y int) (*OperationResult, error) {
	return c.doOperation(fmt.Sprintf("/v1/double_tap/%s", c.serial), pointRequest{X: x, Y: y})
}

// LongPress performs a long press at the specified coordinates.
func (c *Client) LongPress(x, y int) (*OperationResult, error) {
	return c.doOperation(fmt.Sprintf("/v1/long_press/%s", c.serial), pointRequest{X: x, Y: y})
}

// Swipe performs a swipe gesture from (x1,y1) to (x2,y2).
func (c *Client) Swipe(x1, y1, x2, y2 int) (*OperationResult, error) {
	return c.doOperation(fmt.Sprintf("/v1/swipe/%s", c.serial), boundsRequest{X1: x1, Y1: y1, X2: x2, Y2: y2})
}

// Back simulates the device back button press.
func (c *Client) Back() (*OperationResult, error) {
	return c.doOperation(fmt.Sprintf("/v1/back/%s", c.serial), nil)
}

// Home simulates the device home button press.
func (c *Client) Home() (*OperationResult, error) {
	return c.doOperation(fmt.Sprintf("/v1/home/%s", c.serial), nil)
}

// LaunchApp launches an application on the device.
func (c *Client) LaunchApp(appName string) (*OperationResult, error) {
	return c.doOperation(fmt.Sprintf("/v1/launch_app/%s", c.serial), launchAppRequest{AppName: appName})
}

// GetCurrentApp returns information about the currently running foreground app.
func (c *Client) GetCurrentApp() (*AppInfo, error) {
	data, err := c.http.doJSON("POST", fmt.Sprintf("/v1/current_app/%s", c.serial), nil)
	if err != nil {
		return nil, err
	}
	return &AppInfo{Data: data}, nil
}

// InputText inputs text into the currently focused field.
func (c *Client) InputText(text string) (*OperationResult, error) {
	return c.doOperation(fmt.Sprintf("/v1/input/%s", c.serial), inputTextRequest{Text: text})
}

// ClearText clears text in the currently focused field.
func (c *Client) ClearText() (*OperationResult, error) {
	return c.doOperation(fmt.Sprintf("/v1/clear_text/%s", c.serial), nil)
}

// DumpHierarchy returns the current UI hierarchy structure.
func (c *Client) DumpHierarchy() (*HierarchyInfo, error) {
	data, err := c.http.doJSON("POST", fmt.Sprintf("/v1/dump_hierarchy/%s", c.serial), nil)
	if err != nil {
		return nil, err
	}
	return &HierarchyInfo{Data: data}, nil
}

// GetScreenshot returns a screenshot of the device screen as JPEG bytes.
func (c *Client) GetScreenshot() ([]byte, error) {
	return c.http.doRaw("GET", fmt.Sprintf("/v1/screen/%s", c.serial))
}

// DownloadScreenshot downloads the screenshot as a file attachment.
func (c *Client) DownloadScreenshot() ([]byte, error) {
	return c.http.doRaw("GET", fmt.Sprintf("/v1/screenshot/%s", c.serial))
}

func (c *Client) doOperation(path string, body any) (*OperationResult, error) {
	data, err := c.http.doJSON("POST", path, body)
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

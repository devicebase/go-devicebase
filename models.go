package devicebase

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Timestamp is a timestamp parsed from an API response.
//
// The server is not consistent about the format: device rows carry a naive
// local time with no zone offset ("2026-09-20T15:11:31"), while other fields
// are RFC 3339. A plain time.Time rejects the naive form and would fail the
// whole response over a display field, so both are accepted here.
type Timestamp struct {
	time.Time
}

// timestampLayouts are tried in order; the first that parses wins.
var timestampLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02",
}

func (ts *Timestamp) UnmarshalJSON(data []byte) error {
	text := strings.Trim(string(data), `"`)
	if text == "" || text == "null" {
		return nil
	}
	for _, layout := range timestampLayouts {
		if parsed, err := time.Parse(layout, text); err == nil {
			ts.Time = parsed
			return nil
		}
	}
	return fmt.Errorf("unrecognized timestamp %q", text)
}

func (ts Timestamp) MarshalJSON() ([]byte, error) {
	if ts.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(ts.Time)
}

// DeviceInfo contains device status, hardware info, and connection state.
type DeviceInfo struct {
	Serialno string
	Data     map[string]any
}

// AppInfo contains information about the currently running application.
type AppInfo struct {
	Data map[string]any
}

// HierarchyInfo contains the UI element tree.
type HierarchyInfo struct {
	Data map[string]any
}

// OperationResult is the result of a device control operation.
type OperationResult struct {
	Success bool
	Data    map[string]any
}

// Device is one row from the device list.
//
// Serialno is the platform-issued identifier (e.g. "db-mttul4i41di8") and is
// what every control method takes. DeviceSN is the physical serial number —
// the gateway resolves either, but Serialno is the primary key.
type Device struct {
	ID        int       `json:"id"`
	Serialno  string    `json:"serialno"`
	DeviceSN  string    `json:"device_sn"`
	State     string    `json:"state"`
	Name      string    `json:"name"`
	AliasName string    `json:"alias_name"`
	UDID      string    `json:"udid"`
	Type      string    `json:"type"`
	Brand     string    `json:"brand"`
	Model     string    `json:"model"`
	OSType    string    `json:"os_type"`
	OSVersion string    `json:"os_version"`
	Display   string    `json:"display"`
	Location  string    `json:"location"`
	Operator  string    `json:"operator"`
	Network   string    `json:"network"`
	UpdatedAt Timestamp `json:"updated_at"`
}

// UnmarshalJSON reads the identifier from either spelling of the key.
//
// The current service sends "serialno"; the older Python service sends
// "serial" for the same column. Without the fallback, Serialno decodes to ""
// against the older service and every list-driven lookup silently breaks.
func (d *Device) UnmarshalJSON(data []byte) error {
	// An alias type, so decoding does not recurse back into this method.
	type deviceAlias Device
	aux := struct {
		*deviceAlias
		LegacySerial string `json:"serial"`
	}{deviceAlias: (*deviceAlias)(d)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if d.Serialno == "" {
		d.Serialno = aux.LegacySerial
	}
	return nil
}

// --- Geometry -------------------------------------------------------------

// Point is a single screen coordinate: mobile tap / double-tap / long-press,
// and computer click / move.
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Bounds is a screen rectangle: mobile swipe path (x1,y1)→(x2,y2) and computer
// drag.
type Bounds struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

// Mouse buttons accepted by ComputerClick, and the value the server assumes
// when Button is left empty.
const (
	ButtonLeft   = "left"
	ButtonRight  = "right"
	ButtonMiddle = "middle"
)

// Scroll directions accepted by ComputerScroll.
const (
	ScrollUp    = "up"
	ScrollDown  = "down"
	ScrollLeft  = "left"
	ScrollRight = "right"
)

// --- Request payloads -----------------------------------------------------

// LaunchAppRequest is the body for the actions that take an app name: mobile
// launch_app / stop_app and computer launch_app.
type LaunchAppRequest struct {
	AppName string `json:"app_name"`
}

// InputTextRequest is the body for the actions that insert text: mobile input,
// browser input (CDP Input.insertText) and computer type_text.
type InputTextRequest struct {
	Text string `json:"text"`
}

// CommandRequest is the body for the mobile bash action (adb/hdc only).
type CommandRequest struct {
	Command string `json:"command"`
}

// InstallAppRequest is the body for the mobile install_app action. AppPath is a
// path on the agent host, not a local file.
type InstallAppRequest struct {
	AppPath string `json:"app_path"`
}

// URLRequest is the body for browser navigate and tab/open.
type URLRequest struct {
	URL string `json:"url"`
}

// SelectorRequest is the body for the browser click action.
type SelectorRequest struct {
	Selector string `json:"selector"`
}

// SelectorValueRequest is the body for the browser fill and select actions.
type SelectorValueRequest struct {
	Selector string `json:"selector"`
	Value    string `json:"value"`
}

// ScriptRequest is the body for the browser execute action.
type ScriptRequest struct {
	Script string `json:"script"`
}

// KeysRequest is the body for the browser and computer hotkey actions.
type KeysRequest struct {
	Keys []string `json:"keys"`
}

// TabIDRequest is the body for the browser tab close/switch actions.
type TabIDRequest struct {
	TabID string `json:"tab_id"`
}

// ComputerClickRequest is the body for the computer click action. An empty
// Button is omitted and the server defaults it to "left".
type ComputerClickRequest struct {
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Button string `json:"button,omitempty"`
}

// ComputerLongClickRequest is the body for the computer long_click action.
// Duration is in seconds and is omitted when zero, leaving the driver default.
type ComputerLongClickRequest struct {
	X        int `json:"x"`
	Y        int `json:"y"`
	Duration int `json:"duration,omitempty"`
}

// ScrollRequest is the body for the computer scroll action.
type ScrollRequest struct {
	Direction string `json:"direction"`
	Amount    int    `json:"amount,omitempty"`
}

// PressRequest is the body for the computer press action.
type PressRequest struct {
	Key string `json:"key"`
}

// WaitRequest is the body for the computer wait action. Seconds is what the
// server sleeps for; the CLI's `wait` argument is milliseconds.
type WaitRequest struct {
	Seconds float64 `json:"seconds"`
}

// ComputerBashRequest is the body for the computer bash action. Timeout is in
// seconds and is omitted when zero, so the server applies its 120s default.
type ComputerBashRequest struct {
	Command string `json:"command"`
	Timeout int    `json:"timeout,omitempty"`
}

// ListDevicesRequest filters the device list. Every field is optional; an
// empty value is left out of the query string.
//
// Type accepts either a category bucket (mobile|browser|computer) or a system
// type (android|harmonyos|ios|macos|windows|linux|chrome|chromium|edge|other);
// buckets are resolved server-side against the device's os_type.
type ListDevicesRequest struct {
	Keyword string
	State   string
	Type    string

	// Limit caps how many devices come back. The server honours it (default 10,
	// clamped to 1-100); zero omits it and lets the server decide.
	Limit int
}

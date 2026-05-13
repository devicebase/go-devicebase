package devicebase

// DeviceInfo contains device status, hardware info, and connection state.
type DeviceInfo struct {
	Serial string
	Data   map[string]any
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

// Internal request types for JSON serialization.

type pointRequest struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type boundsRequest struct {
	X1 int `json:"x1"`
	Y1 int `json:"y1"`
	X2 int `json:"x2"`
	Y2 int `json:"y2"`
}

type launchAppRequest struct {
	AppName string `json:"app_name"`
}

type inputTextRequest struct {
	Text string `json:"text"`
}

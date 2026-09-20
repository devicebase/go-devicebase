package devicebase

import (
	"fmt"
	"time"
)

// Computer (desktop) platform client.
//
// Path family: POST/GET /api/computer/{serialno}/{action}.
//
// The serialno is the platform serialno of a registered computer device — the
// serialno field from ListDevices with Type "computer". Coordinates are
// absolute screen pixels. The read-only actions (position, screen_size,
// permissions) carry no body.

// defaultBashTimeoutSeconds mirrors the server's own default (and the agent
// Bash tool's): used only to size the client deadline when the caller omits it.
const defaultBashTimeoutSeconds = 120

// actionTimeoutMargin covers connect and body-read overhead on top of whatever
// a blocking action was asked to wait for.
const actionTimeoutMargin = 15 * time.Second

// bashTimeout bounds the HTTP deadline for a bash call. It must exceed the
// requested command timeout, or the client would abort a command the server is
// still running and report a transport error.
func bashTimeout(timeoutSeconds int) time.Duration {
	if timeoutSeconds <= 0 {
		timeoutSeconds = defaultBashTimeoutSeconds
	}
	return time.Duration(timeoutSeconds)*time.Second + actionTimeoutMargin
}

// waitTimeout is the same idea for the wait action, whose budget arrives as
// milliseconds. The shared 30s default would abort every wait past 30s.
func waitTimeout(ms int) time.Duration {
	if ms < 0 {
		ms = 0
	}
	return time.Duration(ms)*time.Millisecond + actionTimeoutMargin
}

// computerPath builds /api/computer/{serialno}/{action}.
func computerPath(action, serialno string) string {
	return fmt.Sprintf("/api/computer/%s/%s", serialno, action)
}

// --- Mouse ----------------------------------------------------------------

// ComputerClick clicks at absolute screen coordinates. An empty Button is
// omitted and the server defaults it to ButtonLeft.
func (c *Client) ComputerClick(serialno string, req ComputerClickRequest) (*OperationResult, error) {
	return c.doOperation(computerPath("click", serialno), req)
}

// ComputerDoubleClick double clicks at absolute screen coordinates (left button).
func (c *Client) ComputerDoubleClick(serialno string, p Point) (*OperationResult, error) {
	return c.doOperation(computerPath("double_click", serialno), p)
}

// ComputerLongClick presses and holds the left button at the coordinates.
// A zero Duration is omitted, leaving the driver default.
func (c *Client) ComputerLongClick(serialno string, req ComputerLongClickRequest) (*OperationResult, error) {
	return c.doOperation(computerPath("long_click", serialno), req)
}

// ComputerMove moves the mouse to absolute screen coordinates without clicking.
func (c *Client) ComputerMove(serialno string, p Point) (*OperationResult, error) {
	return c.doOperation(computerPath("move", serialno), p)
}

// ComputerDrag presses the left button at (x1,y1), moves to (x2,y2) and releases.
func (c *Client) ComputerDrag(serialno string, b Bounds) (*OperationResult, error) {
	return c.doOperation(computerPath("drag", serialno), b)
}

// ComputerScroll scrolls the mouse wheel. Direction is one of ScrollUp,
// ScrollDown, ScrollLeft, ScrollRight; a zero Amount is omitted.
func (c *Client) ComputerScroll(serialno string, req ScrollRequest) (*OperationResult, error) {
	return c.doOperation(computerPath("scroll", serialno), req)
}

// --- Keyboard -------------------------------------------------------------

// ComputerTypeText types text at the current caret of the focused app.
func (c *Client) ComputerTypeText(serialno, text string) (*OperationResult, error) {
	return c.doOperation(computerPath("type_text", serialno), InputTextRequest{Text: text})
}

// ComputerPress presses a single key, e.g. "Enter" or "F5".
func (c *Client) ComputerPress(serialno, key string) (*OperationResult, error) {
	return c.doOperation(computerPath("press", serialno), PressRequest{Key: key})
}

// ComputerHotkey presses the given keys together.
func (c *Client) ComputerHotkey(serialno string, keys []string) (*OperationResult, error) {
	return c.doOperation(computerPath("hotkey", serialno), KeysRequest{Keys: keys})
}

// --- System ---------------------------------------------------------------

// ComputerPosition returns the current mouse position.
func (c *Client) ComputerPosition(serialno string) (*OperationResult, error) {
	return c.doGetOperation(computerPath("position", serialno))
}

// ComputerScreenSize returns the primary screen size.
func (c *Client) ComputerScreenSize(serialno string) (*OperationResult, error) {
	return c.doGetOperation(computerPath("screen_size", serialno))
}

// ComputerPermissions returns the desktop-control permission status.
func (c *Client) ComputerPermissions(serialno string) (*OperationResult, error) {
	return c.doGetOperation(computerPath("permissions", serialno))
}

// ComputerLaunchApp launches a desktop application.
func (c *Client) ComputerLaunchApp(serialno, appName string) (*OperationResult, error) {
	return c.doOperation(computerPath("launch_app", serialno), LaunchAppRequest{AppName: appName})
}

// ComputerWait blocks for the given duration in milliseconds.
//
// The transport deadline is widened to cover the wait itself — the 30s default
// would abort any wait longer than that while the server is still sleeping.
func (c *Client) ComputerWait(serialno string, ms int) (*OperationResult, error) {
	return c.doOperationWith(
		c.http.withTimeout(waitTimeout(ms)),
		"POST",
		computerPath("wait", serialno),
		WaitRequest{Seconds: float64(ms) / 1000},
	)
}

// ComputerBash runs a shell command on the host machine that owns this device,
// under the platform default shell and as the desktop user — it is not
// sandboxed, so treat it as shell access.
//
// timeoutSeconds is the server-side command budget; zero omits the field and
// lets the server apply its 120s default. The transport deadline is widened to
// cover it. The command's own exit status comes back in the payload as
// data.exitCode — a non-zero value is not an API error, so this returns no
// error for it.
func (c *Client) ComputerBash(serialno, command string, timeoutSeconds int) (*OperationResult, error) {
	return c.doOperationWith(
		c.http.withTimeout(bashTimeout(timeoutSeconds)),
		"POST",
		computerPath("bash", serialno),
		ComputerBashRequest{Command: command, Timeout: timeoutSeconds},
	)
}

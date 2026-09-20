package devicebase

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// The live computer-device serialno, used only for readable paths in assertions.
const computerSerial = "db-mtthisv311f1"

func computerCases() []callCase {
	base := "/api/computer/" + computerSerial + "/"

	return []callCase{
		{
			name: "click",
			call: func(c *Client) error {
				_, err := c.ComputerClick(computerSerial, ComputerClickRequest{X: 640, Y: 360})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "click",
			wantBody:   `{"x":640,"y":360}`,
		},
		{
			name: "click with a button",
			call: func(c *Client) error {
				_, err := c.ComputerClick(computerSerial, ComputerClickRequest{
					X: 640, Y: 360, Button: ButtonRight,
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "click",
			wantBody:   `{"x":640,"y":360,"button":"right"}`,
		},
		{
			name: "double click",
			call: func(c *Client) error {
				_, err := c.ComputerDoubleClick(computerSerial, Point{X: 10, Y: 20})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "double_click",
			wantBody:   `{"x":10,"y":20}`,
		},
		{
			name: "long click",
			call: func(c *Client) error {
				_, err := c.ComputerLongClick(computerSerial, ComputerLongClickRequest{
					X: 10, Y: 20, Duration: 2,
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "long_click",
			wantBody:   `{"x":10,"y":20,"duration":2}`,
		},
		{
			name: "move",
			call: func(c *Client) error {
				_, err := c.ComputerMove(computerSerial, Point{X: 100, Y: 100})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "move",
			wantBody:   `{"x":100,"y":100}`,
		},
		{
			name: "drag",
			call: func(c *Client) error {
				_, err := c.ComputerDrag(computerSerial, Bounds{X1: 100, Y1: 100, X2: 800, Y2: 600})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "drag",
			wantBody:   `{"x1":100,"y1":100,"x2":800,"y2":600}`,
		},
		{
			name: "scroll",
			call: func(c *Client) error {
				_, err := c.ComputerScroll(computerSerial, ScrollRequest{
					Direction: ScrollDown, Amount: 5,
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "scroll",
			wantBody:   `{"direction":"down","amount":5}`,
		},
		{
			name: "type text",
			call: func(c *Client) error {
				_, err := c.ComputerTypeText(computerSerial, "hello")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "type_text",
			wantBody:   `{"text":"hello"}`,
		},
		{
			name: "press",
			call: func(c *Client) error {
				_, err := c.ComputerPress(computerSerial, "Enter")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "press",
			wantBody:   `{"key":"Enter"}`,
		},
		{
			name: "hotkey",
			call: func(c *Client) error {
				_, err := c.ComputerHotkey(computerSerial, []string{"Control", "Shift", "Escape"})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "hotkey",
			wantBody:   `{"keys":["Control","Shift","Escape"]}`,
		},
		{
			name: "position",
			call: func(c *Client) error {
				_, err := c.ComputerPosition(computerSerial)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   base + "position",
		},
		{
			name: "screen size",
			call: func(c *Client) error {
				_, err := c.ComputerScreenSize(computerSerial)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   base + "screen_size",
		},
		{
			name: "permissions",
			call: func(c *Client) error {
				_, err := c.ComputerPermissions(computerSerial)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   base + "permissions",
		},
		{
			name: "launch app",
			call: func(c *Client) error {
				_, err := c.ComputerLaunchApp(computerSerial, "Visual Studio Code")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "launch_app",
			wantBody:   `{"app_name":"Visual Studio Code"}`,
		},
		{
			// The SDK takes milliseconds like the CLI; the wire field is seconds.
			name: "wait converts milliseconds to seconds",
			call: func(c *Client) error {
				_, err := c.ComputerWait(computerSerial, 2500)
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "wait",
			wantBody:   `{"seconds":2.5}`,
		},
		{
			name: "bash",
			call: func(c *Client) error {
				_, err := c.ComputerBash(computerSerial, "ls -la", 0)
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "bash",
			wantBody:   `{"command":"ls -la"}`,
		},
		{
			name: "bash with a timeout",
			call: func(c *Client) error {
				_, err := c.ComputerBash(computerSerial, "sleep 5", 30)
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "bash",
			wantBody:   `{"command":"sleep 5","timeout":30}`,
		},
	}
}

func TestComputerActions(t *testing.T) {
	runCalls(t, nil, computerCases())
}

func TestComputerCoversEveryContractAction(t *testing.T) {
	// The contract exposes 15 computer actions; the table above has 17 cases
	// because click and bash are each covered twice.
	if got := len(computerCases()); got != 17 {
		t.Errorf("computer cases = %d, want 17", got)
	}
}

// --- Blocking actions widen their deadline ---------------------------------

func TestWaitTimeoutExceedsTheRequestedWait(t *testing.T) {
	tests := []struct {
		ms   int
		want time.Duration
	}{
		{0, actionTimeoutMargin},
		{-5, actionTimeoutMargin},
		{2500, 2500*time.Millisecond + actionTimeoutMargin},
		{300_000, 300*time.Second + actionTimeoutMargin},
	}

	for _, tt := range tests {
		if got := waitTimeout(tt.ms); got != tt.want {
			t.Errorf("waitTimeout(%d) = %v, want %v", tt.ms, got, tt.want)
		}
	}
}

func TestBashTimeoutExceedsTheRequestedTimeout(t *testing.T) {
	tests := []struct {
		seconds int
		want    time.Duration
	}{
		// An omitted timeout must still outlast the server's 120s default.
		{0, 120*time.Second + actionTimeoutMargin},
		{-1, 120*time.Second + actionTimeoutMargin},
		{30, 30*time.Second + actionTimeoutMargin},
		{600, 600*time.Second + actionTimeoutMargin},
	}

	for _, tt := range tests {
		if got := bashTimeout(tt.seconds); got != tt.want {
			t.Errorf("bashTimeout(%d) = %v, want %v", tt.seconds, got, tt.want)
		}
	}
}

// Blocking actions raise their own deadline, so a slow call must survive a
// client-wide timeout that would abort an ordinary request.
//
// The server here sleeps well past the shared default; Wait and Bash have to
// outlast it, while a plain action is expected to time out — which proves the
// deadline really is per call rather than a global bump.
func TestBlockingActionsOutlastTheDefaultTimeout(t *testing.T) {
	slowServer := func(sleep time.Duration) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(sleep)
			w.Write([]byte(`{"code":200,"message":"success","data":{}}`))
		}))
	}

	const sharedDefault = 40 * time.Millisecond
	const serverDelay = 200 * time.Millisecond

	t.Run("wait outlasts the default", func(t *testing.T) {
		server := slowServer(serverDelay)
		defer server.Close()

		client := NewClient(
			WithAPIKey("k"),
			WithBaseURL(server.URL),
			WithTimeout(sharedDefault),
		)
		if _, err := client.ComputerWait(computerSerial, 50); err != nil {
			t.Fatalf("ComputerWait should widen its own deadline: %v", err)
		}
	})

	t.Run("bash outlasts the default", func(t *testing.T) {
		server := slowServer(serverDelay)
		defer server.Close()

		client := NewClient(
			WithAPIKey("k"),
			WithBaseURL(server.URL),
			WithTimeout(sharedDefault),
		)
		if _, err := client.ComputerBash(computerSerial, "ls", 1); err != nil {
			t.Fatalf("ComputerBash should widen its own deadline: %v", err)
		}
	})

	t.Run("an ordinary action does not", func(t *testing.T) {
		server := slowServer(serverDelay)
		defer server.Close()

		client := NewClient(
			WithAPIKey("k"),
			WithBaseURL(server.URL),
			WithTimeout(sharedDefault),
		)
		if _, err := client.ComputerPosition(computerSerial); err == nil {
			t.Fatal("a non-blocking action should still be bound by the shared default")
		}
	})
}

// The per-call deadline is a copy: a blocking action must not leave the shared
// default raised behind it.
func TestBlockingActionsLeaveTheSharedDeadlineAlone(t *testing.T) {
	r := newRecorder(t)
	defer r.server.Close()

	client := r.client()
	if _, err := client.ComputerWait(computerSerial, 60_000); err != nil {
		t.Fatalf("ComputerWait: %v", err)
	}
	if client.http.hc.Timeout != defaultTimeout {
		t.Errorf("shared default = %v, want %v", client.http.hc.Timeout, defaultTimeout)
	}
}

// A bash timeout that exceeds the server's own cap must still be sent, so the
// server can reject it rather than the SDK silently truncating it.
func TestComputerBashSendsTheTimeoutVerbatim(t *testing.T) {
	r := newRecorder(t)
	defer r.server.Close()

	if _, err := r.client().ComputerBash(computerSerial, "ls", 600); err != nil {
		t.Fatalf("ComputerBash: %v", err)
	}

	got := r.last(t).Body
	if got["timeout"] != float64(600) {
		t.Errorf("timeout = %v, want 600", got["timeout"])
	}
}

package devicebase

import (
	"net/http"
	"testing"
)

// The live browser-device serialno, used only for readable paths in assertions.
const browserSerial = "db-mtsi49bf0mqb"

func browserCases() []callCase {
	base := "/api/browser/" + browserSerial + "/"

	return []callCase{
		{
			name: "navigate",
			call: func(c *Client) error {
				_, err := c.BrowserNavigate(browserSerial, "https://example.com")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "navigate",
			wantBody:   `{"url":"https://example.com"}`,
		},
		{
			name: "refresh",
			call: func(c *Client) error {
				_, err := c.BrowserRefresh(browserSerial)
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "refresh",
		},
		{
			name: "go back",
			call: func(c *Client) error {
				_, err := c.BrowserGoBack(browserSerial)
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "go_back",
		},
		{
			name: "go forward",
			call: func(c *Client) error {
				_, err := c.BrowserGoForward(browserSerial)
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "go_forward",
		},
		{
			name: "input",
			call: func(c *Client) error {
				_, err := c.BrowserInput(browserSerial, "hello 世界")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "input",
			wantBody:   `{"text":"hello 世界"}`,
		},
		{
			name: "click",
			call: func(c *Client) error {
				_, err := c.BrowserClick(browserSerial, "button#submit")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "click",
			wantBody:   `{"selector":"button#submit"}`,
		},
		{
			name: "fill",
			call: func(c *Client) error {
				_, err := c.BrowserFill(browserSerial, "#search", "devicebase")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "fill",
			wantBody:   `{"selector":"#search","value":"devicebase"}`,
		},
		{
			name: "select",
			call: func(c *Client) error {
				_, err := c.BrowserSelect(browserSerial, "#country", "CN")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "select",
			wantBody:   `{"selector":"#country","value":"CN"}`,
		},
		{
			// The read-only DOM actions carry the selector as a query parameter.
			name: "text",
			call: func(c *Client) error {
				_, err := c.BrowserText(browserSerial, "#search")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   base + "text",
			wantQuery:  map[string]string{"selector": "#search"},
		},
		{
			name: "attribute",
			call: func(c *Client) error {
				_, err := c.BrowserAttribute(browserSerial, "a.logo", "href")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   base + "attribute",
			wantQuery:  map[string]string{"selector": "a.logo", "attribute": "href"},
		},
		{
			name: "exists",
			call: func(c *Client) error {
				_, err := c.BrowserExists(browserSerial, ".modal")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   base + "exists",
			wantQuery:  map[string]string{"selector": ".modal"},
		},
		{
			name: "execute",
			call: func(c *Client) error {
				_, err := c.BrowserExecute(browserSerial, "document.title")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "execute",
			wantBody:   `{"script":"document.title"}`,
		},
		{
			name: "hotkey",
			call: func(c *Client) error {
				_, err := c.BrowserHotkey(browserSerial, []string{"Meta", "a"})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "hotkey",
			wantBody:   `{"keys":["Meta","a"]}`,
		},
		{
			name: "state",
			call: func(c *Client) error {
				_, err := c.BrowserState(browserSerial)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   base + "state",
		},
		{
			name: "tabs",
			call: func(c *Client) error {
				_, err := c.BrowserTabs(browserSerial)
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   base + "tabs",
		},
		{
			name: "tab open",
			call: func(c *Client) error {
				_, err := c.BrowserTabOpen(browserSerial, "https://example.com")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "tab/open",
			wantBody:   `{"url":"https://example.com"}`,
		},
		{
			name: "tab close",
			call: func(c *Client) error {
				_, err := c.BrowserTabClose(browserSerial, "tab-7")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "tab/close",
			wantBody:   `{"tab_id":"tab-7"}`,
		},
		{
			name: "tab close all",
			call: func(c *Client) error {
				_, err := c.BrowserTabCloseAll(browserSerial)
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "tab/close_all",
		},
		{
			name: "tab switch",
			call: func(c *Client) error {
				_, err := c.BrowserTabSwitch(browserSerial, "tab-7")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "tab/switch",
			wantBody:   `{"tab_id":"tab-7"}`,
		},
		{
			name: "launch",
			call: func(c *Client) error {
				_, err := c.BrowserLaunch(browserSerial)
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "launch",
		},
		{
			name: "close",
			call: func(c *Client) error {
				_, err := c.BrowserClose(browserSerial)
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   base + "close",
		},
	}
}

func TestBrowserActions(t *testing.T) {
	runCalls(t, nil, browserCases())
}

func TestBrowserCoversEveryContractAction(t *testing.T) {
	// The contract exposes 21 browser actions; guard against one being dropped
	// when the method list is edited.
	if got := len(browserCases()); got != 21 {
		t.Errorf("browser cases = %d, want 21", got)
	}
}

package devicebase

import "fmt"

// Browser (Chrome/CDP) platform client.
//
// Path family: POST/GET /api/browser/{serial}/{action...}.
//
// The serial is the platform serialno of a registered browser device — the
// serialno field from ListDevices with Type "browser". Selectors are CSS
// selectors. The read-only actions (state, tabs, text, attribute, exists) carry
// no body; selectors travel as query parameters.

// browserPath builds /api/browser/{serial}/{action}.
func browserPath(action, serial string) string {
	return fmt.Sprintf("/api/browser/%s/%s", serial, action)
}

// --- Navigation -----------------------------------------------------------

// BrowserNavigate navigates the current tab to a URL.
func (c *Client) BrowserNavigate(serial, url string) (*OperationResult, error) {
	return c.doOperation(browserPath("navigate", serial), URLRequest{URL: url})
}

// BrowserRefresh reloads the current page.
func (c *Client) BrowserRefresh(serial string) (*OperationResult, error) {
	return c.doOperation(browserPath("refresh", serial), nil)
}

// BrowserGoBack navigates back in the browser history.
func (c *Client) BrowserGoBack(serial string) (*OperationResult, error) {
	return c.doOperation(browserPath("go_back", serial), nil)
}

// BrowserGoForward navigates forward in the browser history.
func (c *Client) BrowserGoForward(serial string) (*OperationResult, error) {
	return c.doOperation(browserPath("go_forward", serial), nil)
}

// BrowserInput inserts text into the focused page element via CDP
// Input.insertText, which is reliable for CJK unlike synthesised key events.
func (c *Client) BrowserInput(serial, text string) (*OperationResult, error) {
	return c.doOperation(browserPath("input", serial), InputTextRequest{Text: text})
}

// --- DOM ------------------------------------------------------------------

// BrowserClick clicks the element matching a CSS selector.
func (c *Client) BrowserClick(serial, selector string) (*OperationResult, error) {
	return c.doOperation(browserPath("click", serial), SelectorRequest{Selector: selector})
}

// BrowserFill clears an input and types a value into it.
func (c *Client) BrowserFill(serial, selector, value string) (*OperationResult, error) {
	return c.doOperation(browserPath("fill", serial), SelectorValueRequest{
		Selector: selector,
		Value:    value,
	})
}

// BrowserSelect picks an option in a dropdown.
func (c *Client) BrowserSelect(serial, selector, value string) (*OperationResult, error) {
	return c.doOperation(browserPath("select", serial), SelectorValueRequest{
		Selector: selector,
		Value:    value,
	})
}

// BrowserText returns an element's text content.
func (c *Client) BrowserText(serial, selector string) (*OperationResult, error) {
	query := buildQuery(map[string]string{"selector": selector})
	return c.doGetOperation(browserPath("text", serial) + query)
}

// BrowserAttribute returns one attribute of an element.
func (c *Client) BrowserAttribute(serial, selector, attribute string) (*OperationResult, error) {
	query := buildQuery(map[string]string{"selector": selector, "attribute": attribute})
	return c.doGetOperation(browserPath("attribute", serial) + query)
}

// BrowserExists reports whether an element is present.
func (c *Client) BrowserExists(serial, selector string) (*OperationResult, error) {
	query := buildQuery(map[string]string{"selector": selector})
	return c.doGetOperation(browserPath("exists", serial) + query)
}

// BrowserExecute evaluates JavaScript in the page.
//
// Danger tier: the script runs with the page's own privileges, the same reach
// as shell access to the browser profile.
func (c *Client) BrowserExecute(serial, script string) (*OperationResult, error) {
	return c.doOperation(browserPath("execute", serial), ScriptRequest{Script: script})
}

// BrowserHotkey presses the given keys together, e.g. ["Meta", "a"].
//
// Editing shortcuts (select-all, cut, copy, undo, redo) act on the page.
// Browser-chrome shortcuts such as Control+t are not reachable — CDP drives the
// page, not the browser UI.
func (c *Client) BrowserHotkey(serial string, keys []string) (*OperationResult, error) {
	return c.doOperation(browserPath("hotkey", serial), KeysRequest{Keys: keys})
}

// --- Tabs and state -------------------------------------------------------

// BrowserState returns the current URL, title, viewport and tab count.
func (c *Client) BrowserState(serial string) (*OperationResult, error) {
	return c.doGetOperation(browserPath("state", serial))
}

// BrowserTabs lists the open tabs.
func (c *Client) BrowserTabs(serial string) (*OperationResult, error) {
	return c.doGetOperation(browserPath("tabs", serial))
}

// BrowserTabOpen opens a new tab at a URL.
func (c *Client) BrowserTabOpen(serial, url string) (*OperationResult, error) {
	return c.doOperation(browserPath("tab/open", serial), URLRequest{URL: url})
}

// BrowserTabClose closes one tab.
func (c *Client) BrowserTabClose(serial, tabID string) (*OperationResult, error) {
	return c.doOperation(browserPath("tab/close", serial), TabIDRequest{TabID: tabID})
}

// BrowserTabCloseAll closes every tab.
func (c *Client) BrowserTabCloseAll(serial string) (*OperationResult, error) {
	return c.doOperation(browserPath("tab/close_all", serial), nil)
}

// BrowserTabSwitch focuses one tab.
func (c *Client) BrowserTabSwitch(serial, tabID string) (*OperationResult, error) {
	return c.doOperation(browserPath("tab/switch", serial), TabIDRequest{TabID: tabID})
}

// --- Lifecycle ------------------------------------------------------------

// BrowserLaunch starts the browser / CDP endpoint.
func (c *Client) BrowserLaunch(serial string) (*OperationResult, error) {
	return c.doOperation(browserPath("launch", serial), nil)
}

// BrowserClose stops the browser / CDP endpoint.
func (c *Client) BrowserClose(serial string) (*OperationResult, error) {
	return c.doOperation(browserPath("close", serial), nil)
}

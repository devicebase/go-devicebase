package devicebase

import "fmt"

// Browser (Chrome/CDP) platform client.
//
// Path family: POST/GET /api/browser/{serialno}/{action...}.
//
// The serialno is the platform serialno of a registered browser device — the
// serialno field from ListDevices with Type "browser". Selectors are CSS
// selectors. The read-only actions (state, tabs, text, attribute, exists) carry
// no body; selectors travel as query parameters.

// browserPath builds /api/browser/{serialno}/{action}.
func browserPath(action, serialno string) string {
	return fmt.Sprintf("/api/browser/%s/%s", serialno, action)
}

// --- Navigation -----------------------------------------------------------

// BrowserNavigate navigates the current tab to a URL.
func (c *Client) BrowserNavigate(serialno, url string) (*OperationResult, error) {
	return c.doOperation(browserPath("navigate", serialno), URLRequest{URL: url})
}

// BrowserRefresh reloads the current page.
func (c *Client) BrowserRefresh(serialno string) (*OperationResult, error) {
	return c.doOperation(browserPath("refresh", serialno), nil)
}

// BrowserGoBack navigates back in the browser history.
func (c *Client) BrowserGoBack(serialno string) (*OperationResult, error) {
	return c.doOperation(browserPath("go_back", serialno), nil)
}

// BrowserGoForward navigates forward in the browser history.
func (c *Client) BrowserGoForward(serialno string) (*OperationResult, error) {
	return c.doOperation(browserPath("go_forward", serialno), nil)
}

// BrowserInput inserts text into the focused page element via CDP
// Input.insertText, which is reliable for CJK unlike synthesised key events.
func (c *Client) BrowserInput(serialno, text string) (*OperationResult, error) {
	return c.doOperation(browserPath("input", serialno), InputTextRequest{Text: text})
}

// --- DOM ------------------------------------------------------------------

// BrowserClick clicks the element matching a CSS selector.
func (c *Client) BrowserClick(serialno, selector string) (*OperationResult, error) {
	return c.doOperation(browserPath("click", serialno), SelectorRequest{Selector: selector})
}

// BrowserFill clears an input and types a value into it.
func (c *Client) BrowserFill(serialno, selector, value string) (*OperationResult, error) {
	return c.doOperation(browserPath("fill", serialno), SelectorValueRequest{
		Selector: selector,
		Value:    value,
	})
}

// BrowserSelect picks an option in a dropdown.
func (c *Client) BrowserSelect(serialno, selector, value string) (*OperationResult, error) {
	return c.doOperation(browserPath("select", serialno), SelectorValueRequest{
		Selector: selector,
		Value:    value,
	})
}

// BrowserText returns an element's text content.
func (c *Client) BrowserText(serialno, selector string) (*OperationResult, error) {
	query := buildQuery(map[string]string{"selector": selector})
	return c.doGetOperation(browserPath("text", serialno) + query)
}

// BrowserAttribute returns one attribute of an element.
func (c *Client) BrowserAttribute(serialno, selector, attribute string) (*OperationResult, error) {
	query := buildQuery(map[string]string{"selector": selector, "attribute": attribute})
	return c.doGetOperation(browserPath("attribute", serialno) + query)
}

// BrowserExists reports whether an element is present.
func (c *Client) BrowserExists(serialno, selector string) (*OperationResult, error) {
	query := buildQuery(map[string]string{"selector": selector})
	return c.doGetOperation(browserPath("exists", serialno) + query)
}

// BrowserExecute evaluates JavaScript in the page.
//
// Danger tier: the script runs with the page's own privileges, the same reach
// as shell access to the browser profile.
func (c *Client) BrowserExecute(serialno, script string) (*OperationResult, error) {
	return c.doOperation(browserPath("execute", serialno), ScriptRequest{Script: script})
}

// BrowserHotkey presses the given keys together, e.g. ["Meta", "a"].
//
// Editing shortcuts (select-all, cut, copy, undo, redo) act on the page.
// Browser-chrome shortcuts such as Control+t are not reachable — CDP drives the
// page, not the browser UI.
func (c *Client) BrowserHotkey(serialno string, keys []string) (*OperationResult, error) {
	return c.doOperation(browserPath("hotkey", serialno), KeysRequest{Keys: keys})
}

// --- Tabs and state -------------------------------------------------------

// BrowserState returns the current URL, title, viewport and tab count.
func (c *Client) BrowserState(serialno string) (*OperationResult, error) {
	return c.doGetOperation(browserPath("state", serialno))
}

// BrowserTabs lists the open tabs.
func (c *Client) BrowserTabs(serialno string) (*OperationResult, error) {
	return c.doGetOperation(browserPath("tabs", serialno))
}

// BrowserTabOpen opens a new tab at a URL.
func (c *Client) BrowserTabOpen(serialno, url string) (*OperationResult, error) {
	return c.doOperation(browserPath("tab/open", serialno), URLRequest{URL: url})
}

// BrowserTabClose closes one tab.
func (c *Client) BrowserTabClose(serialno, tabID string) (*OperationResult, error) {
	return c.doOperation(browserPath("tab/close", serialno), TabIDRequest{TabID: tabID})
}

// BrowserTabCloseAll closes every tab.
func (c *Client) BrowserTabCloseAll(serialno string) (*OperationResult, error) {
	return c.doOperation(browserPath("tab/close_all", serialno), nil)
}

// BrowserTabSwitch focuses one tab.
func (c *Client) BrowserTabSwitch(serialno, tabID string) (*OperationResult, error) {
	return c.doOperation(browserPath("tab/switch", serialno), TabIDRequest{TabID: tabID})
}

// --- Lifecycle ------------------------------------------------------------

// BrowserLaunch starts the browser / CDP endpoint.
func (c *Client) BrowserLaunch(serialno string) (*OperationResult, error) {
	return c.doOperation(browserPath("launch", serialno), nil)
}

// BrowserClose stops the browser / CDP endpoint.
func (c *Client) BrowserClose(serialno string) (*OperationResult, error) {
	return c.doOperation(browserPath("close", serialno), nil)
}

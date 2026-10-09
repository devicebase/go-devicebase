package devicebase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Cloud browser lifecycle — POST /v1/browser/create, DELETE /v1/browser/{serialno},
// GET /v1/browser/{serialno}/status and GET /v1/browser/quota.
//
// Not to be confused with browser.go: those are device actions against an
// already registered browser (/api/browser/{serialno}/{action}), while these
// four are the account-level resource lifecycle that creates and destroys one.
// A cloud browser is a browser the platform runs for you on its own cluster; a
// Chrome you attached yourself is not one of them — the device list marks the
// difference with IsCloud.
//
// The contract lives in devicebase-ts/openapi/src/api/routes/browser.ts.
//
// Status codes carry meaning here and decide whether a retry is worth it: 409
// for a conflict that retrying will not fix (quota exhausted, not a cloud
// browser), 503 for the temporary kind (no capacity, node unreachable), 502
// for a platform↔node problem somebody has to repair. They surface as
// *HTTPError with the matching StatusCode.

// CloudBrowserCreateRequest carries the creation parameters. Every field is
// optional: the platform picks the machine, so nothing here chooses a node.
//
// There is no Headless field. A cloud browser always runs headless — it lives
// on a machine nobody is looking at, with no display to put a window on — and
// the platform pins that itself rather than leaving it to a flag.
type CloudBrowserCreateRequest struct {
	// Name is the label shown in the device list (max 100 characters). The
	// server applies its own naming rule when it is empty.
	Name string `json:"name,omitempty"`
	// WindowSize is the "1366x768" shape the node parses.
	WindowSize string `json:"window_size,omitempty"`
	// WaitSeconds is how long the platform waits for the browser to register
	// before answering (0-60; 0 = answer immediately). Nil lets the server
	// apply its own default of 15.
	WaitSeconds *int `json:"wait_seconds,omitempty"`
}

// CloudBrowserCreateResult is the data of a successful creation.
//
// Serialno is the platform key every other browser method takes. Creation is
// asynchronous, so CreateCloudBrowser waits for the browser to register (15
// seconds by default) in order to hand one back — nobody should have to write a
// poll loop to use what they just created.
//
// If it has not come up within the wait, Registered is false and Serialno is
// empty while the call still succeeds: DeviceSN remains a valid handle for
// CloudBrowserStatus and DeleteCloudBrowser. That is not a failure — a slow
// start is normal — and reporting it as one would invite a retry that creates a
// second browser.
type CloudBrowserCreateResult struct {
	// Serialno is the platform identifier (db-…), empty until it registers.
	Serialno string `json:"serialno"`
	// DeviceSN is the node-side identifier the browser registers under.
	DeviceSN string `json:"device_sn"`
	// Name is the platform's own identity for this browser
	// ("Browser-" + the first 8 characters of the serial). It is not the name
	// you asked for — that is AliasName.
	Name string `json:"name"`
	// AliasName is the name you passed to CreateCloudBrowser, verbatim, or
	// empty when you passed none. The platform pins it at creation and the
	// node's later re-registrations cannot change it, so it is stable for the
	// life of the browser — nodes name their instances "<requested>-<port>",
	// which is not what a caller should be shown.
	AliasName string `json:"alias_name"`
	// Registered says whether the device row exists yet.
	Registered bool `json:"registered"`
}

// CloudBrowserQuota is the account's cloud browser allowance. Counted from the
// same source as the check creation performs, so the two cannot disagree.
type CloudBrowserQuota struct {
	Limit     int `json:"limit"`
	Used      int `json:"used"`
	Remaining int `json:"remaining"`
}

// CloudBrowserStatus answers whether a created browser has registered itself.
//
// Registered is false until the browser starts and reports in — that is a
// normal answer, not an error, so polling is quiet. The remaining fields are
// only meaningful once it is true.
type CloudBrowserStatus struct {
	Registered bool   `json:"registered"`
	DeviceID   int    `json:"device_id,omitempty"`
	Serialno   string `json:"serialno,omitempty"`
	// Name is the platform's own identity name; AliasName is the caller's.
	Name      string `json:"name,omitempty"`
	AliasName string `json:"alias_name,omitempty"`
	State     string `json:"state,omitempty"`
	ServerURL string `json:"server_url,omitempty"`
	IsCloud   bool   `json:"is_cloud,omitempty"`
}

// createDispatchBudget and createTimeoutSlack bound this call's own deadline.
//
// Creating is the one call that blocks: the platform hands the instruction to a
// node (up to 20s) and then waits for the browser to register (up to
// WaitSeconds) before answering. The shared 30s default would abort a request
// the server is still working on.
const (
	createDispatchBudget = 20 * time.Second
	createTimeoutSlack   = 25 * time.Second
)

// CreateCloudBrowser asks the platform for a new cloud browser.
//
// Waits for it to come up (15 seconds by default; see
// CloudBrowserCreateRequest.WaitSeconds) and returns the Serialno to drive it
// with. If it has not registered within the wait the call still succeeds — see
// CloudBrowserCreateResult.
func (c *Client) CreateCloudBrowser(req CloudBrowserCreateRequest) (*CloudBrowserCreateResult, error) {
	wait := 15
	if req.WaitSeconds != nil {
		wait = *req.WaitSeconds
	}
	if wait < 0 {
		// Negative input is the server's to reject; treating it as "no wait"
		// keeps the deadline arithmetic sane on the way there.
		wait = 0
	}
	transport := c.http.withTimeout(createDispatchBudget + createTimeoutSlack + time.Duration(wait)*time.Second)

	data, err := transport.doRequest(http.MethodPost, "/v1/browser/create", req)
	if err != nil {
		return nil, err
	}

	var envelope struct {
		Data CloudBrowserCreateResult `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("decode create result: %w", err)
	}
	return &envelope.Data, nil
}

// DeleteCloudBrowser destroys a cloud browser and its profile. Irreversible,
// and not the same as BrowserClose, which only stops the CDP engine.
//
// The identifier is the platform serialno ("db-…") or the Serial returned by
// CreateCloudBrowser. The call does not wait for the machine: the platform
// queues the reap and the node collects it on its next heartbeat, so this
// succeeds even while the node is offline.
func (c *Client) DeleteCloudBrowser(identifier string) error {
	_, err := c.http.doRequest(http.MethodDelete, browserIdentifierPath(identifier), nil)
	return err
}

// CloudBrowserStatus reports whether the browser has registered itself yet,
// and under which serialno.
func (c *Client) CloudBrowserStatus(identifier string) (*CloudBrowserStatus, error) {
	data, err := c.http.doRequest(http.MethodGet, browserIdentifierPath(identifier)+"/status", nil)
	if err != nil {
		return nil, err
	}

	var envelope struct {
		Data CloudBrowserStatus `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("decode browser status: %w", err)
	}
	return &envelope.Data, nil
}

// CloudBrowserQuota returns how many cloud browsers the account may still
// create — asked before creating beats being told no after placement ran.
func (c *Client) CloudBrowserQuota() (*CloudBrowserQuota, error) {
	data, err := c.http.doRequest(http.MethodGet, "/v1/browser/quota", nil)
	if err != nil {
		return nil, err
	}

	var envelope struct {
		Data CloudBrowserQuota `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("decode browser quota: %w", err)
	}
	return &envelope.Data, nil
}

// browserIdentifierPath builds /v1/browser/{identifier}, escaping the segment:
// the identifier is caller-supplied and, unlike a serialno, not guaranteed to
// be URL-safe.
func browserIdentifierPath(identifier string) string {
	return "/v1/browser/" + url.PathEscape(identifier)
}

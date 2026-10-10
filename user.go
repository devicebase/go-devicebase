package devicebase

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// The account itself — GET /v1/user/info and POST /v1/user/checkin.
//
// The third account-level group, alongside ListDevices and the cloud browser
// lifecycle: none of them addresses a device, so none takes a serialno. This
// one answers "whose key is this, and what does the account have".
//
// Only ever your own account — the API key decides whose record this is, and
// there is no parameter for pointing it at somebody else.
//
// The contract lives in devicebase-ts/openapi/src/api/routes/user.ts; the
// console carries the same two capabilities behind session auth.

// UserInfo is the account behind the API key.
type UserInfo struct {
	ID int `json:"id"`
	// Username is the name shown in the console.
	Username string `json:"username"`
	Mobile   string `json:"mobile"`
	// Credits is the points balance (积分).
	Credits int `json:"credits"`
	// RegisteredAt is when the account was created, naive ISO format.
	RegisteredAt string `json:"registered_at"`
	// CanCheckin says whether today's reward is still unclaimed.
	CanCheckin bool `json:"can_checkin"`
}

// UserCheckin is the result of claiming the daily points.
//
// AlreadyChecked is a normal outcome rather than an error: claiming twice in
// one day answers with AlreadyChecked true, CreditsEarned 0 and the platform's
// own message ("今日已签到"). That is what makes UserCheckin safe to call from
// a daily scheduled task.
type UserCheckin struct {
	// Success is whether this call actually granted points.
	Success bool `json:"success"`
	// CreditsEarned is 0 when the day was already claimed.
	CreditsEarned int `json:"credits_earned"`
	// ConsecutiveDays counts the streak this claim belongs to (1 when the
	// previous day was not claimed).
	ConsecutiveDays int  `json:"consecutive_days"`
	AlreadyChecked  bool `json:"already_checked"`
	// Message is the platform's own wording, for display as-is.
	Message string `json:"message"`
	// Credits is the balance after this call.
	Credits int `json:"credits"`
}

// UserInfo returns the account behind the API key: name, phone, points balance,
// registration date, and whether today's check-in is still unclaimed.
func (c *Client) UserInfo() (*UserInfo, error) {
	data, err := c.http.doRequest(http.MethodGet, "/v1/user/info", nil)
	if err != nil {
		return nil, err
	}

	var envelope struct {
		Data UserInfo `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("decode user info: %w", err)
	}
	return &envelope.Data, nil
}

// UserCheckin claims the daily points: 25 on the first day, +10 per consecutive
// day, up to 95 a day.
//
// Claiming twice in one day is not an error — the second call returns
// AlreadyChecked true and grants nothing — so a daily scheduled task can run
// this unconditionally, without asking CanCheckin first.
//
// An empty object is sent rather than no body at all: the call has no
// parameters either way, and {} is what a deployment predating the server's
// body-less-POST tolerance also accepts (the client always sends
// content-type: application/json).
func (c *Client) UserCheckin() (*UserCheckin, error) {
	data, err := c.http.doRequest(http.MethodPost, "/v1/user/checkin", struct{}{})
	if err != nil {
		return nil, err
	}

	var envelope struct {
		Data UserCheckin `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("decode check-in result: %w", err)
	}
	return &envelope.Data, nil
}

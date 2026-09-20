package devicebase

import (
	"fmt"
	"net/http"
)

// maxErrorBodyChars caps the body text kept in an error message. A failed
// action's envelope is small; this only guards against a misrouted binary
// response being pasted into a log.
const maxErrorBodyChars = 4096

// Error is the base error type for all DeviceBase SDK errors.
//
// Errors raised from an HTTP status carry StatusCode; errors raised from the
// response envelope carry no status (the transport succeeded) — see
// BusinessError.
type Error struct {
	Message    string
	StatusCode int
}

func (e *Error) Error() string { return e.Message }

// AuthenticationError is returned when the API key is missing, or when the
// server rejects it (HTTP 401).
type AuthenticationError struct {
	Message    string
	StatusCode int
}

func (e *AuthenticationError) Error() string { return e.Message }

// DeviceNotFoundError is returned when the device is not found or not connected
// (HTTP 404).
type DeviceNotFoundError struct {
	Message    string
	StatusCode int
}

func (e *DeviceNotFoundError) Error() string { return e.Message }

// ValidationError is returned when the server rejects the request parameters.
type ValidationError struct {
	Message    string
	StatusCode int
}

func (e *ValidationError) Error() string { return e.Message }

// BusinessError is returned when the API answers with a success status but
// reports a failure inside the response envelope.
//
// The control API returns HTTP 200 with a non-2xx "code" for action failures —
// a browser selector that matches nothing, a computer command that cannot run.
// Guarding on the HTTP status alone would report those as success, so the
// envelope is inspected too.
type BusinessError struct {
	// Code is the envelope's code, e.g. 502.
	Code int
	// Body is the raw response body, which carries the server's own message.
	Body string
	// Message is the formatted error text.
	Message string
}

func (e *BusinessError) Error() string { return e.Message }

// newBusinessError builds a BusinessError from an envelope code and the raw
// body it was read from.
func newBusinessError(code int, body []byte) *BusinessError {
	return &BusinessError{
		Code:    code,
		Body:    string(body),
		Message: fmt.Sprintf("API error (code %d): %s", code, truncateBody(body)),
	}
}

// newHTTPError maps a non-2xx HTTP status onto the matching error type,
// carrying the server's own body so its message reaches the caller.
func newHTTPError(statusCode int, body []byte) error {
	message := fmt.Sprintf("API error (HTTP %d): %s", statusCode, truncateBody(body))

	switch statusCode {
	case http.StatusUnauthorized:
		return &AuthenticationError{Message: message, StatusCode: statusCode}
	case http.StatusNotFound:
		return &DeviceNotFoundError{Message: message, StatusCode: statusCode}
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		// The gateway reports validation failures as 400; 422 is kept for
		// compatibility with older deployments.
		return &ValidationError{Message: message, StatusCode: statusCode}
	default:
		return &Error{Message: message, StatusCode: statusCode}
	}
}

// truncateBody renders a response body for an error message, bounded so a
// binary or oversized payload cannot flood a log.
func truncateBody(body []byte) string {
	if len(body) <= maxErrorBodyChars {
		return string(body)
	}
	return string(body[:maxErrorBodyChars]) + "… (truncated)"
}

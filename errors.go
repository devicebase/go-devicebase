package devicebase

import "fmt"

// Error is the base error type for all DeviceBase SDK errors.
type Error struct {
	Message    string
	StatusCode int
}

func (e *Error) Error() string { return e.Message }

// AuthenticationError is returned when the API key is missing or invalid.
type AuthenticationError struct {
	Message    string
	StatusCode int
}

func (e *AuthenticationError) Error() string { return e.Message }

// DeviceNotFoundError is returned when the device is not found or not connected.
type DeviceNotFoundError struct {
	Message    string
	StatusCode int
}

func (e *DeviceNotFoundError) Error() string { return e.Message }

// ValidationError is returned when request validation fails.
type ValidationError struct {
	Message    string
	StatusCode int
}

func (e *ValidationError) Error() string { return e.Message }

// newError creates a base Error from an HTTP status and body.
func newError(statusCode int, body string) error {
	return &Error{
		Message:    fmt.Sprintf("API error: %d - %s", statusCode, body),
		StatusCode: statusCode,
	}
}

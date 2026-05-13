package devicebase

import (
	"testing"
)

func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		message string
		code    int
	}{
		{
			name:    "base error",
			err:     newError(500, "internal"),
			message: "API error: 500 - internal",
			code:    500,
		},
		{
			name:    "authentication error",
			err:     &AuthenticationError{Message: "auth failed", StatusCode: 401},
			message: "auth failed",
			code:    401,
		},
		{
			name:    "device not found error",
			err:     &DeviceNotFoundError{Message: "not found", StatusCode: 404},
			message: "not found",
			code:    404,
		},
		{
			name:    "validation error",
			err:     &ValidationError{Message: "invalid input", StatusCode: 422},
			message: "invalid input",
			code:    422,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.message {
				t.Errorf("Error() = %q, want %q", tt.err.Error(), tt.message)
			}
		})
	}
}

func TestErrorStatusCodes(t *testing.T) {
	err := &AuthenticationError{Message: "auth", StatusCode: 401}
	if err.StatusCode != 401 {
		t.Errorf("StatusCode = %d, want 401", err.StatusCode)
	}

	dnf := &DeviceNotFoundError{Message: "not found", StatusCode: 404}
	if dnf.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", dnf.StatusCode)
	}

	ve := &ValidationError{Message: "invalid", StatusCode: 422}
	if ve.StatusCode != 422 {
		t.Errorf("StatusCode = %d, want 422", ve.StatusCode)
	}

	be := newError(500, "internal")
	if e, ok := be.(*Error); !ok {
		t.Error("newError should return *Error")
	} else if e.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", e.StatusCode)
	}
}

func TestErrorTypeDistinction(t *testing.T) {
	var err error = &AuthenticationError{Message: "auth", StatusCode: 401}

	if _, ok := err.(*DeviceNotFoundError); ok {
		t.Error("AuthenticationError should not match DeviceNotFoundError via type assertion")
	}

	if _, ok := err.(*ValidationError); ok {
		t.Error("AuthenticationError should not match ValidationError via type assertion")
	}

	if _, ok := err.(*AuthenticationError); !ok {
		t.Error("AuthenticationError should match itself")
	}
}

package devicebase

import (
	"strings"
	"testing"
)

func TestHTTPErrorMessages(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		want    string
		wantTyp any
	}{
		{
			name:    "unauthorized",
			status:  401,
			body:    `{"code":401,"message":"API Key 不存在或已停用"}`,
			want:    `API error (HTTP 401): {"code":401,"message":"API Key 不存在或已停用"}`,
			wantTyp: &AuthenticationError{},
		},
		{
			name:    "not found",
			status:  404,
			body:    `{"code":404,"message":"设备不存在"}`,
			want:    `API error (HTTP 404): {"code":404,"message":"设备不存在"}`,
			wantTyp: &DeviceNotFoundError{},
		},
		{
			// The gateway reports validation failures as 400.
			name:    "bad request is a validation error",
			status:  400,
			body:    `{"detail":"body.username: Field required"}`,
			want:    `API error (HTTP 400): {"detail":"body.username: Field required"}`,
			wantTyp: &ValidationError{},
		},
		{
			name:    "unprocessable entity is also a validation error",
			status:  422,
			body:    "invalid input",
			want:    `API error (HTTP 422): invalid input`,
			wantTyp: &ValidationError{},
		},
		{
			name:    "server error",
			status:  500,
			body:    "internal",
			want:    `API error (HTTP 500): internal`,
			wantTyp: &Error{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := newHTTPError(tt.status, []byte(tt.body))
			if err.Error() != tt.want {
				t.Errorf("Error() = %q, want %q", err.Error(), tt.want)
			}
			if !sameType(err, tt.wantTyp) {
				t.Errorf("error type = %T, want %T", err, tt.wantTyp)
			}
		})
	}
}

func TestHTTPErrorStatusCode(t *testing.T) {
	err := newHTTPError(404, []byte("nope"))
	notFound, ok := err.(*DeviceNotFoundError)
	if !ok {
		t.Fatalf("error type = %T, want *DeviceNotFoundError", err)
	}
	if notFound.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", notFound.StatusCode)
	}
}

// A business failure carries the envelope's code, not an HTTP status — the
// transport itself succeeded.
func TestBusinessErrorMessage(t *testing.T) {
	body := `{"code":502,"message":"-32602: Invalid parameters"}`
	err := newBusinessError(502, []byte(body))

	if err.Code != 502 {
		t.Errorf("Code = %d, want 502", err.Code)
	}
	if err.Body != body {
		t.Errorf("Body = %q, want %q", err.Body, body)
	}
	want := `API error (code 502): {"code":502,"message":"-32602: Invalid parameters"}`
	if err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestBusinessErrorTruncatesALongBody(t *testing.T) {
	long := strings.Repeat("x", maxErrorBodyChars+100)
	err := newBusinessError(500, []byte(long))

	if !strings.HasSuffix(err.Error(), "… (truncated)") {
		t.Error("a body over the cap should be marked as truncated")
	}
	if err.Body != long {
		t.Error("Body should keep the untruncated payload")
	}
}

func TestErrorTypeDistinction(t *testing.T) {
	err := error(&AuthenticationError{Message: "auth", StatusCode: 401})

	if _, ok := err.(*DeviceNotFoundError); ok {
		t.Error("AuthenticationError should not match DeviceNotFoundError")
	}
	if _, ok := err.(*ValidationError); ok {
		t.Error("AuthenticationError should not match ValidationError")
	}
	if _, ok := err.(*BusinessError); ok {
		t.Error("AuthenticationError should not match BusinessError")
	}
	if _, ok := err.(*AuthenticationError); !ok {
		t.Error("AuthenticationError should match itself")
	}
}

// sameType reports whether err has the same concrete type as want, without
// relying on errors.As pointer-target plumbing in a table test.
func sameType(err error, want any) bool {
	switch want.(type) {
	case *AuthenticationError:
		_, ok := err.(*AuthenticationError)
		return ok
	case *DeviceNotFoundError:
		_, ok := err.(*DeviceNotFoundError)
		return ok
	case *ValidationError:
		_, ok := err.(*ValidationError)
		return ok
	case *BusinessError:
		_, ok := err.(*BusinessError)
		return ok
	case *Error:
		_, ok := err.(*Error)
		return ok
	}
	return false
}

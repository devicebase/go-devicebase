package devicebase

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPClientDoJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Authorization = %q, want %q", r.Header.Get("Authorization"), "Bearer test-key")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}

		if r.Method != "POST" {
			t.Errorf("Method = %q, want POST", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			var got map[string]any
			json.Unmarshal(body, &got)
			want := map[string]any{"x": float64(100), "y": float64(200)}
			if got["x"] != want["x"] || got["y"] != want["y"] {
				t.Errorf("body = %v, want %v", got, want)
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	}))
	defer server.Close()

	client := newHTTPClient(server.URL, "test-key", nil)
	data, err := client.doJSON("POST", "/v1/tap/test-device", pointRequest{X: 100, Y: 200})
	if err != nil {
		t.Fatalf("doJSON: %v", err)
	}
	if data["success"] != true {
		t.Errorf("response = %v, want success=true", data)
	}
}

func TestHTTPClientDoJSONEmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := newHTTPClient(server.URL, "key", nil)
	data, err := client.doJSON("POST", "/v1/back/serial", nil)
	if err != nil {
		t.Fatalf("doJSON: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty map, got %v", data)
	}
}

func TestHTTPClientDoRaw(t *testing.T) {
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Method = %q, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(jpegData)
	}))
	defer server.Close()

	client := newHTTPClient(server.URL, "key", nil)
	data, err := client.doRaw("GET", "/v1/screen/serial")
	if err != nil {
		t.Fatalf("doRaw: %v", err)
	}
	if string(data) != string(jpegData) {
		t.Errorf("data = %v, want %v", data, jpegData)
	}
}

func TestHTTPClientAuthenticationError(t *testing.T) {
	server := newStatusServer(401)
	defer server.Close()

	client := newHTTPClient(server.URL, "key", nil)
	_, err := client.doJSON("POST", "/v1/test/serial", nil)
	if err == nil {
		t.Fatal("expected error")
	}

	var authErr *AuthenticationError
	if !isErrorType(err, &authErr) {
		t.Errorf("error type = %T, want *AuthenticationError", err)
	}
}

func TestHTTPClientDeviceNotFoundError(t *testing.T) {
	server := newStatusServer(404)
	defer server.Close()

	client := newHTTPClient(server.URL, "key", nil)
	_, err := client.doJSON("POST", "/v1/test/serial", nil)
	if err == nil {
		t.Fatal("expected error")
	}

	var notFoundErr *DeviceNotFoundError
	if !isErrorType(err, &notFoundErr) {
		t.Errorf("error type = %T, want *DeviceNotFoundError", err)
	}
}

func TestHTTPClientValidationError(t *testing.T) {
	server := newStatusServer(422)
	defer server.Close()

	client := newHTTPClient(server.URL, "key", nil)
	_, err := client.doJSON("POST", "/v1/test/serial", nil)
	if err == nil {
		t.Fatal("expected error")
	}

	var valErr *ValidationError
	if !isErrorType(err, &valErr) {
		t.Errorf("error type = %T, want *ValidationError", err)
	}
}

func TestHTTPClientGenericError(t *testing.T) {
	server := newStatusServer(500)
	defer server.Close()

	client := newHTTPClient(server.URL, "key", nil)
	_, err := client.doJSON("POST", "/v1/test/serial", nil)
	if err == nil {
		t.Fatal("expected error")
	}

	var baseErr *Error
	if !isErrorType(err, &baseErr) {
		t.Errorf("error type = %T, want *Error", err)
	}
}

func newStatusServer(statusCode int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte("error body"))
	}))
}

// isErrorType uses a type assertion to check if err matches the target type.
// This avoids the errors.As nil-pointer issues in table-driven tests.
func isErrorType(err error, target any) bool {
	switch t := target.(type) {
	case **AuthenticationError:
		_, ok := err.(*AuthenticationError)
		return ok
	case **DeviceNotFoundError:
		_, ok := err.(*DeviceNotFoundError)
		return ok
	case **ValidationError:
		_, ok := err.(*ValidationError)
		return ok
	case **Error:
		// Error should NOT match the subtypes
		_, ok := err.(*Error)
		if !ok {
			return false
		}
		// Ensure it's not a subtype
		_, isAuth := err.(*AuthenticationError)
		_, isDev := err.(*DeviceNotFoundError)
		_, isVal := err.(*ValidationError)
		_ = t
		return !isAuth && !isDev && !isVal
	}
	return false
}

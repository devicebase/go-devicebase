package devicebase

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPClientDoJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("Authorization = %q, want %q", r.Header.Get("Authorization"), "Bearer test-key")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
		}
		if r.Method != http.MethodPost {
			t.Errorf("Method = %q, want POST", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			var got map[string]any
			json.Unmarshal(body, &got)
			if got["x"] != float64(100) || got["y"] != float64(200) {
				t.Errorf("body = %v, want x=100 y=200", got)
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{"success": true})
	}))
	defer server.Close()

	client := newHTTPClient(server.URL, "test-key", nil)
	data, err := client.doJSON(http.MethodPost, "/v1/tap/test-device", Point{X: 100, Y: 200})
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
	data, err := client.doJSON(http.MethodPost, "/v1/back/serial", nil)
	if err != nil {
		t.Fatalf("doJSON: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty map, got %v", data)
	}
}

// A bare JSON array is not an object envelope; it is wrapped rather than
// rejected so an endpoint that answers with a list still decodes.
func TestHTTPClientDecodesATopLevelArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"id":1},{"id":2}]`))
	}))
	defer server.Close()

	client := newHTTPClient(server.URL, "key", nil)
	data, err := client.doJSON(http.MethodGet, "/v1/devices", nil)
	if err != nil {
		t.Fatalf("doJSON: %v", err)
	}
	rows, ok := data["data"].([]any)
	if !ok {
		t.Fatalf("data = %T, want []any", data["data"])
	}
	if len(rows) != 2 {
		t.Errorf("len(rows) = %d, want 2", len(rows))
	}
}

func TestHTTPClientRejectsANonJSONBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json at all"))
	}))
	defer server.Close()

	client := newHTTPClient(server.URL, "key", nil)
	if _, err := client.doJSON(http.MethodGet, "/v1/devices", nil); err == nil {
		t.Fatal("expected an error for a non-JSON body")
	}
}

func TestHTTPClientDoRaw(t *testing.T) {
	jpegData := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %q, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(jpegData)
	}))
	defer server.Close()

	client := newHTTPClient(server.URL, "key", nil)
	data, err := client.doRaw(http.MethodPost, "/v1/screen/serial")
	if err != nil {
		t.Fatalf("doRaw: %v", err)
	}
	if string(data) != string(jpegData) {
		t.Errorf("data = %v, want %v", data, jpegData)
	}
}

// --- Business errors carried inside a successful response -----------------

// The control API answers HTTP 200 with a non-2xx code when an action fails in
// the driver. Trusting the status line alone would report that as success.
func TestHTTPClientSurfacesAnEnvelopeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"code":502,"message":"-32602: Invalid parameters"}`))
	}))
	defer server.Close()

	client := newHTTPClient(server.URL, "key", nil)
	_, err := client.doJSON(http.MethodPost, "/api/browser/x/click", nil)
	if err == nil {
		t.Fatal("expected the envelope's non-2xx code to raise an error")
	}

	bizErr, ok := err.(*BusinessError)
	if !ok {
		t.Fatalf("error type = %T, want *BusinessError", err)
	}
	if bizErr.Code != 502 {
		t.Errorf("Code = %d, want 502", bizErr.Code)
	}
}

func TestHTTPClientEnvelopeCodesInThe2xxRangePass(t *testing.T) {
	for _, body := range []string{
		`{"code":200,"message":"success"}`,
		`{"code":201,"message":"created"}`,
		`{"success":true}`,      // no envelope at all
		`{"code":"200"}`,        // a string code is not an envelope code
		`[{"code":502}]`,        // an array is left alone
		`{"data":{"code":502}}`, // a nested code is not the envelope's
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(body))
		}))
		client := newHTTPClient(server.URL, "key", nil)
		if _, err := client.doJSON(http.MethodPost, "/v1/x", nil); err != nil {
			t.Errorf("body %s: unexpected error %v", body, err)
		}
		server.Close()
	}
}

// A screenshot that fails server-side comes back as an envelope on an HTTP 200,
// which must not be handed to the caller as image bytes.
func TestHTTPClientDoRawSurfacesAnEnvelopeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code":503,"message":"device offline"}`))
	}))
	defer server.Close()

	client := newHTTPClient(server.URL, "key", nil)
	_, err := client.doRaw(http.MethodPost, "/v1/screen/serial")
	if err == nil {
		t.Fatal("expected the envelope's non-2xx code to raise an error")
	}
	if _, ok := err.(*BusinessError); !ok {
		t.Fatalf("error type = %T, want *BusinessError", err)
	}
}

// --- Credentials and timeouts ---------------------------------------------

func TestHTTPClientRejectsAMissingAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("no request should be sent without an API key")
	}))
	defer server.Close()

	client := newHTTPClient(server.URL, "", nil)
	_, err := client.doJSON(http.MethodPost, "/v1/back/serial", nil)
	if err == nil {
		t.Fatal("expected an error")
	}
	if _, ok := err.(*AuthenticationError); !ok {
		t.Fatalf("error type = %T, want *AuthenticationError", err)
	}
}

func TestDefaultTimeoutIsApplied(t *testing.T) {
	client := newHTTPClient("http://example.invalid", "key", nil)
	if client.hc.Timeout != defaultTimeout {
		t.Errorf("Timeout = %v, want %v", client.hc.Timeout, defaultTimeout)
	}
}

// withTimeout copies the transport so a blocking action can raise its deadline
// without disturbing the shared default.
func TestWithTimeoutLeavesTheOriginalUntouched(t *testing.T) {
	client := newHTTPClient("http://example.invalid", "key", nil)
	slow := client.withTimeout(2 * time.Minute)

	if slow.hc.Timeout != 2*time.Minute {
		t.Errorf("copy Timeout = %v, want 2m", slow.hc.Timeout)
	}
	if client.hc.Timeout != defaultTimeout {
		t.Errorf("original Timeout = %v, want %v", client.hc.Timeout, defaultTimeout)
	}
	if slow == client {
		t.Error("withTimeout should return a distinct transport")
	}
}

func TestHTTPClientErrorsCarryTheServerBody(t *testing.T) {
	server := newStatusServer(401)
	defer server.Close()

	client := newHTTPClient(server.URL, "key", nil)
	_, err := client.doJSON(http.MethodPost, "/v1/test/serial", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "error body") {
		t.Errorf("error %q should carry the server's body", err)
	}
	if !strings.HasPrefix(err.Error(), "API error (HTTP 401): ") {
		t.Errorf("error %q should use the CLI's wording", err)
	}
}

func TestHTTPClientDeviceNotFoundError(t *testing.T) {
	server := newStatusServer(404)
	defer server.Close()

	client := newHTTPClient(server.URL, "key", nil)
	_, err := client.doJSON(http.MethodPost, "/v1/test/serial", nil)

	var notFoundErr *DeviceNotFoundError
	if !isErrorType(err, &notFoundErr) {
		t.Errorf("error type = %T, want *DeviceNotFoundError", err)
	}
}

func TestHTTPClientValidationError(t *testing.T) {
	for _, status := range []int{400, 422} {
		server := newStatusServer(status)
		client := newHTTPClient(server.URL, "key", nil)
		_, err := client.doJSON(http.MethodPost, "/v1/test/serial", nil)

		var valErr *ValidationError
		if !isErrorType(err, &valErr) {
			t.Errorf("status %d: error type = %T, want *ValidationError", status, err)
		}
		server.Close()
	}
}

func TestHTTPClientGenericError(t *testing.T) {
	server := newStatusServer(500)
	defer server.Close()

	client := newHTTPClient(server.URL, "key", nil)
	_, err := client.doJSON(http.MethodPost, "/v1/test/serial", nil)

	var baseErr *Error
	if !isErrorType(err, &baseErr) {
		t.Errorf("error type = %T, want *Error", err)
	}
}

func TestBuildQuery(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
		want   string
	}{
		{
			name:   "all empty produces nothing",
			params: map[string]string{"keyword": "", "state": ""},
			want:   "",
		},
		{
			name:   "empty values are dropped",
			params: map[string]string{"keyword": "pixel", "state": ""},
			want:   "?keyword=pixel",
		},
		{
			// Selectors travel as query parameters on the browser GET actions,
			// so they have to be escaped.
			name:   "values are URL-escaped",
			params: map[string]string{"selector": "a.link[name=x]"},
			want:   "?selector=a.link%5Bname%3Dx%5D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildQuery(tt.params); got != tt.want {
				t.Errorf("buildQuery = %q, want %q", got, tt.want)
			}
		})
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
	switch target.(type) {
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
		// *Error must not match the subtypes.
		_, ok := err.(*Error)
		if !ok {
			return false
		}
		_, isAuth := err.(*AuthenticationError)
		_, isDev := err.(*DeviceNotFoundError)
		_, isVal := err.(*ValidationError)
		return !isAuth && !isDev && !isVal
	}
	return false
}

package devicebase

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// capturedRequest is one request a recording server received, with the body
// already decoded so a test can assert on it as a map.
type capturedRequest struct {
	Method string
	Path   string
	Query  url.Values
	Body   map[string]any
}

// recorder is an httptest server that remembers what it was sent, so a test can
// assert on the method, path, query and body of each call.
type recorder struct {
	server   *httptest.Server
	requests []capturedRequest

	// response is the body sent back; status is the HTTP status.
	response string
	status   int
}

func newRecorder(t *testing.T) *recorder {
	t.Helper()

	r := &recorder{
		response: `{"code":200,"message":"success","data":{}}`,
		status:   http.StatusOK,
	}
	r.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		captured := capturedRequest{
			Method: req.Method,
			Path:   req.URL.Path,
			Query:  req.URL.Query(),
		}
		if raw, err := io.ReadAll(req.Body); err == nil && len(raw) > 0 {
			var body map[string]any
			if err := json.Unmarshal(raw, &body); err == nil {
				captured.Body = body
			}
		}
		r.requests = append(r.requests, captured)

		w.WriteHeader(r.status)
		w.Write([]byte(r.response))
	}))
	return r
}

// client returns a Client pointed at the recorder.
func (r *recorder) client(opts ...Option) *Client {
	base := []Option{
		WithAPIKey("test-key"),
		WithBaseURL(r.server.URL),
	}
	return NewClient(append(base, opts...)...)
}

// last returns the most recent request, failing the test when there was none.
func (r *recorder) last(t *testing.T) capturedRequest {
	t.Helper()
	if len(r.requests) == 0 {
		t.Fatal("no request was sent")
	}
	return r.requests[len(r.requests)-1]
}

// callCase describes one SDK method call and the request it should produce.
type callCase struct {
	name string
	// call invokes the method under test.
	call func(*Client) error
	// wantMethod and wantPath are the expected HTTP method and path.
	wantMethod string
	wantPath   string
	// wantQuery lists the query parameters the call must send, if any.
	wantQuery map[string]string
	// wantBody is the expected JSON body; empty means no body is expected.
	wantBody string
}

// runCalls exercises each case against a fresh recording server. opts are
// applied to the client under test — the mobile actions need WithSerial, while
// the browser and computer actions take their serial per call.
func runCalls(t *testing.T, opts []Option, cases []callCase) {
	t.Helper()

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			r := newRecorder(t)
			defer r.server.Close()

			if err := tt.call(r.client(opts...)); err != nil {
				t.Fatalf("call: %v", err)
			}

			got := r.last(t)
			if got.Method != tt.wantMethod {
				t.Errorf("Method = %q, want %q", got.Method, tt.wantMethod)
			}
			if got.Path != tt.wantPath {
				t.Errorf("Path = %q, want %q", got.Path, tt.wantPath)
			}

			for key, want := range tt.wantQuery {
				if got.Query.Get(key) != want {
					t.Errorf("query %s = %q, want %q", key, got.Query.Get(key), want)
				}
			}

			if tt.wantBody == "" {
				if len(got.Body) > 0 {
					t.Errorf("body = %v, want none", got.Body)
				}
				return
			}
			encoded, err := json.Marshal(got.Body)
			if err != nil {
				t.Fatalf("marshal recorded body: %v", err)
			}
			var want, have any
			if err := json.Unmarshal([]byte(tt.wantBody), &want); err != nil {
				t.Fatalf("bad wantBody %q: %v", tt.wantBody, err)
			}
			if err := json.Unmarshal(encoded, &have); err != nil {
				t.Fatalf("bad recorded body: %v", err)
			}
			if !jsonEqual(want, have) {
				t.Errorf("body = %s, want %s", encoded, tt.wantBody)
			}
		})
	}
}

// jsonEqual compares decoded JSON values by re-encoding, so key order does not
// matter.
func jsonEqual(a, b any) bool {
	left, errA := json.Marshal(a)
	right, errB := json.Marshal(b)
	if errA != nil || errB != nil {
		return false
	}
	return string(left) == string(right)
}

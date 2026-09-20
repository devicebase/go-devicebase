package devicebase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.devicebase.cn"

// defaultTimeout bounds a request end to end, including the body read.
//
// Blocking actions have to raise it per call — http.Client.Timeout wins over
// any longer server-side allowance, so `computer wait` and `computer bash`
// would otherwise be aborted client-side while the server is still working.
const defaultTimeout = 30 * time.Second

type httpClient struct {
	baseURL string
	apiKey  string
	hc      *http.Client
}

func newHTTPClient(baseURL, apiKey string, hc *http.Client) *httpClient {
	if hc == nil {
		hc = &http.Client{Timeout: defaultTimeout}
	}
	return &httpClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		hc:      hc,
	}
}

// withTimeout returns a copy of the transport whose deadline is d, leaving the
// shared default untouched for every other endpoint.
func (c *httpClient) withTimeout(d time.Duration) *httpClient {
	cp := *c
	hc := *c.hc
	hc.Timeout = d
	cp.hc = &hc
	return &cp
}

// doRequest sends a request and returns the raw response body.
//
// Two failure layers are mapped to errors: a non-2xx HTTP status, and a
// business failure reported inside an otherwise successful response (see
// envelopeError).
func (c *httpClient) doRequest(method, path string, body any) ([]byte, error) {
	if c.apiKey == "" {
		return nil, &AuthenticationError{
			Message: "API key is required: pass it with WithAPIKey or set the " +
				envAPIKey + " environment variable",
			StatusCode: http.StatusUnauthorized,
		}
	}

	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return data, newHTTPError(resp.StatusCode, data)
	}
	if err := envelopeError(data); err != nil {
		return data, err
	}
	return data, nil
}

// doJSON sends a request and decodes an object response.
func (c *httpClient) doJSON(method, path string, body any) (map[string]any, error) {
	data, err := c.doRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	return decodeObject(data)
}

// doRaw sends a request and returns the body untouched, for endpoints that
// answer with bytes rather than JSON (screenshots).
func (c *httpClient) doRaw(method, path string) ([]byte, error) {
	return c.doRequest(method, path, nil)
}

// decodeObject tolerates a top-level JSON array by wrapping it, so an endpoint
// that answers with a bare list still decodes.
func decodeObject(data []byte) (map[string]any, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return map[string]any{}, nil
	}

	var obj map[string]any
	if err := json.Unmarshal(trimmed, &obj); err == nil {
		return obj, nil
	}

	var arr []any
	if err := json.Unmarshal(trimmed, &arr); err == nil {
		return map[string]any{"data": arr}, nil
	}
	return nil, fmt.Errorf("decode response: %s", truncateBody(trimmed))
}

// envelopeError surfaces a business error carried inside an otherwise
// successful response.
//
// The control API reports action failures in the response envelope — HTTP 200
// with a non-2xx "code" field, e.g. {"code":502,"message":"-32602: ..."} when a
// browser action fails in the driver. Trusting the status line alone reports
// those as success.
//
// Bodies that are not an envelope (arrays, empty, non-JSON) are left alone, as
// are codes inside the 2xx range.
func envelopeError(data []byte) error {
	var envelope struct {
		Code *int `json:"code"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil
	}
	if envelope.Code == nil || (*envelope.Code >= 200 && *envelope.Code < 300) {
		return nil
	}
	return newBusinessError(*envelope.Code, data)
}

// buildQuery renders query parameters, skipping empty values.
//
// url.Values escapes every key and value, which matters for the browser
// endpoints whose selectors travel as query parameters.
func buildQuery(params map[string]string) string {
	q := url.Values{}
	for key, value := range params {
		if value != "" {
			q.Set(key, value)
		}
	}
	if len(q) == 0 {
		return ""
	}
	return "?" + q.Encode()
}

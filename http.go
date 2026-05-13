package devicebase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const defaultBaseURL = "https://api.devicebase.cn"

type httpClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func newHTTPClient(baseURL, apiKey string, hc *http.Client) *httpClient {
	if hc == nil {
		hc = &http.Client{}
	}
	return &httpClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: hc,
	}
}

func (c *httpClient) doJSON(method, path string, body any) (map[string]any, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if err := handleError(resp); err != nil {
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if len(bodyBytes) == 0 {
		return map[string]any{}, nil
	}

	var result map[string]any
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}

func (c *httpClient) doRaw(method, path string) ([]byte, error) {
	req, err := http.NewRequest(method, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if err := handleError(resp); err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}

func (c *httpClient) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
}

func handleError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return &AuthenticationError{
			Message:    "authentication failed - invalid API key",
			StatusCode: resp.StatusCode,
		}
	case http.StatusNotFound:
		return &DeviceNotFoundError{
			Message:    "device not found or not connected",
			StatusCode: resp.StatusCode,
		}
	case http.StatusUnprocessableEntity:
		return &ValidationError{
			Message:    fmt.Sprintf("validation error: %s", bodyStr),
			StatusCode: resp.StatusCode,
		}
	default:
		return newError(resp.StatusCode, bodyStr)
	}
}

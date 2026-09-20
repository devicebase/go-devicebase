package devicebase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// ListDevices returns the devices accessible to the current API key.
//
// Filtering is resolved server-side: Type accepts either a category bucket
// (mobile|browser|computer) or a system type
// (android|harmonyos|ios|macos|windows|linux|chrome|chromium|edge|other), and
// buckets match against the device's os_type because a device row only carries
// the coarse type. This is how a serialno is discovered before driving a
// device — ListDevicesRequest{Type: "browser"} for a browser endpoint.
//
// The response carries no pagination metadata, so Limit is the only bound on
// the result size.
func (c *Client) ListDevices(req ListDevicesRequest) ([]Device, error) {
	query := buildQuery(map[string]string{
		"keyword": req.Keyword,
		"state":   req.State,
		"type":    req.Type,
		"limit":   formatLimit(req.Limit),
	})

	data, err := c.http.doRequest(http.MethodGet, "/v1/devices"+query, nil)
	if err != nil {
		return nil, err
	}

	var envelope struct {
		Data []Device `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, fmt.Errorf("decode device list: %w", err)
	}
	return envelope.Data, nil
}

// formatLimit renders an optional limit, leaving it out when unset so the
// server applies its own default.
func formatLimit(limit int) string {
	if limit <= 0 {
		return ""
	}
	return strconv.Itoa(limit)
}

package devicebase

import (
	"net/http"
	"testing"
)

// The cloud browser lifecycle lives on /v1/browser/* — a different surface from
// the device actions in browser_test.go, which is why it gets its own count
// guard and its own file. Reference: devicebase-ts/openapi/src/api/routes/browser.ts.

func cloudBrowserCases() []callCase {
	return []callCase{
		{
			name: "create with no fields sends an empty object",
			call: func(c *Client) error {
				_, err := c.CreateCloudBrowser(CloudBrowserCreateRequest{})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/browser/create",
			wantBody:   `{}`,
		},
		{
			name: "create with every field",
			call: func(c *Client) error {
				wait := 30
				_, err := c.CreateCloudBrowser(CloudBrowserCreateRequest{
					Name:        "my-browser",
					WindowSize:  "1366x768",
					WaitSeconds: &wait,
				})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/browser/create",
			wantBody:   `{"name":"my-browser","window_size":"1366x768","wait_seconds":30}`,
		},
		{
			// 云浏览器恒为 headless：这个字段根本不该出现（平台自己钉死模式）
			name: "create never sends a headless mode",
			call: func(c *Client) error {
				_, err := c.CreateCloudBrowser(CloudBrowserCreateRequest{Name: "b"})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/browser/create",
			wantBody:   `{"name":"b"}`,
		},
		{
			name: "delete by serialno",
			call: func(c *Client) error {
				return c.DeleteCloudBrowser("db-mtabc123")
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/v1/browser/db-mtabc123",
		},
		{
			name: "status by the serial create returned",
			call: func(c *Client) error {
				_, err := c.CloudBrowserStatus("3f2a1b4c-0000-1111-2222-333344445555")
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/browser/3f2a1b4c-0000-1111-2222-333344445555/status",
		},
		{
			name: "quota",
			call: func(c *Client) error {
				_, err := c.CloudBrowserQuota()
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/browser/quota",
		},
	}
}

func TestCloudBrowserLifecycle(t *testing.T) {
	runCalls(t, nil, cloudBrowserCases())
}

func TestCloudBrowserCoversEveryLifecycleCall(t *testing.T) {
	// Six calls: three shapes of create, plus delete, status and quota. Guard
	// against one being dropped when the list is edited.
	if got := len(cloudBrowserCases()); got != 6 {
		t.Errorf("cloud browser cases = %d, want 6", got)
	}
}

func TestCloudBrowserCreateDecodesTheResult(t *testing.T) {
	r := newRecorder(t)
	defer r.server.Close()
	// The registration already happened by the time create answers — which is
	// the point of the wait: the caller gets the key it actually uses.
	r.response = `{"code":200,"message":"success","data":{
		"device_sn":"3f2a-uuid","serialno":"db-mtabc123","name":"Browser-3f2a1b4c",
		"alias_name":"my-browser","registered":true}}`

	result, err := r.client().CreateCloudBrowser(CloudBrowserCreateRequest{Name: "b"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if result.Serialno != "db-mtabc123" {
		t.Errorf("Serialno = %q, want db-mtabc123", result.Serialno)
	}
	if result.DeviceSN != "3f2a-uuid" {
		t.Errorf("DeviceSN = %q, want 3f2a-uuid", result.DeviceSN)
	}
	if result.Name != "Browser-3f2a1b4c" {
		t.Errorf("Name = %q, want the platform identity name", result.Name)
	}
	// 调用方要的那个名字在 AliasName 里，原样、不带节点补的端口后缀
	if result.AliasName != "my-browser" || !result.Registered {
		t.Errorf("result = %+v, want a registered browser aliased my-browser", result)
	}
}

func TestCloudBrowserStatusDistinguishesNotUpYet(t *testing.T) {
	// "Not registered yet" is a normal answer, not an error: the SDK must not
	// turn it into a failure, or polling would be useless.
	r := newRecorder(t)
	defer r.server.Close()
	r.response = `{"code":200,"message":"success","data":{"registered":false}}`

	status, err := r.client().CloudBrowserStatus("3f2a-uuid")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if status.Registered {
		t.Error("registered = true, want false")
	}

	r.response = `{"code":200,"message":"success","data":{"registered":true,"serialno":"db-mtabc123","state":"free","is_cloud":true}}`
	status, err = r.client().CloudBrowserStatus("3f2a-uuid")
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !status.Registered || status.Serialno != "db-mtabc123" || status.State != "free" || !status.IsCloud {
		t.Errorf("status = %+v, want the registered device", status)
	}
}

func TestCloudBrowserQuotaDecodesTheNumbers(t *testing.T) {
	r := newRecorder(t)
	defer r.server.Close()
	r.response = `{"code":200,"message":"success","data":{"limit":10,"used":3,"remaining":7}}`

	quota, err := r.client().CloudBrowserQuota()
	if err != nil {
		t.Fatalf("quota: %v", err)
	}
	if quota.Limit != 10 || quota.Used != 3 || quota.Remaining != 7 {
		t.Errorf("quota = %+v, want {10 3 7}", quota)
	}
}

func TestCloudBrowserConflictSurfacesAsHTTPError(t *testing.T) {
	// 409 is the "retrying will not help" class — quota exhausted, or the
	// identifier is not a cloud browser. Callers branch on the status code.
	r := newRecorder(t)
	defer r.server.Close()
	r.status = http.StatusConflict
	r.response = `{"code":409,"message":"你的浏览器数量已达上限（10 台）。删除不再使用的浏览器，或联系管理员调整上限。","trace_id":"t"}`

	_, err := r.client().CreateCloudBrowser(CloudBrowserCreateRequest{})
	if err == nil {
		t.Fatal("expected an error for a 409 response")
	}
	// 409 has no dedicated type — it lands on the base *Error, which still
	// carries the status code the caller branches on.
	conflict, ok := err.(*Error)
	if !ok {
		t.Fatalf("error type = %T (%v), want *Error", err, err)
	}
	if conflict.StatusCode != http.StatusConflict {
		t.Errorf("StatusCode = %d, want 409", conflict.StatusCode)
	}
}

package devicebase

import (
	"net/http"
	"testing"
)

// The account surface lives on /v1/user/* — like the cloud browser lifecycle,
// it addresses no device, which is why it gets its own file and count guard.
// Reference: devicebase-ts/openapi/src/api/routes/user.ts.

func userCases() []callCase {
	return []callCase{
		{
			name: "info takes no arguments",
			call: func(c *Client) error {
				_, err := c.UserInfo()
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/v1/user/info",
		},
		{
			// 空对象而不是没有请求体：签到没有参数，而 {} 在还没有「空体容忍」
			// 的部署上也照样被接受（客户端总会带上 content-type: application/json）
			name: "checkin posts an empty object",
			call: func(c *Client) error {
				_, err := c.UserCheckin()
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/v1/user/checkin",
			wantBody:   `{}`,
		},
	}
}

func TestUserAccountCalls(t *testing.T) {
	runCalls(t, nil, userCases())
}

func TestUserCoversEveryAccountCall(t *testing.T) {
	// Two calls, one per endpoint. Guard against one being dropped when the
	// list is edited.
	if got := len(userCases()); got != 2 {
		t.Errorf("user cases = %d, want 2", got)
	}
}

func TestUserInfoDecodesTheRecord(t *testing.T) {
	r := newRecorder(t)
	defer r.server.Close()
	r.response = `{"code":200,"message":"success","data":{"id":42,"username":"richie",
		"mobile":"13800138000","credits":1259,"registered_at":"2026-01-02T03:04:05",
		"can_checkin":true}}`

	info, err := r.client().UserInfo()
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	want := &UserInfo{
		ID:           42,
		Username:     "richie",
		Mobile:       "13800138000",
		Credits:      1259,
		RegisteredAt: "2026-01-02T03:04:05",
		CanCheckin:   true,
	}
	if *info != *want {
		t.Errorf("info = %+v, want %+v", info, want)
	}
}

func TestUserCheckinDecodesTheReward(t *testing.T) {
	r := newRecorder(t)
	defer r.server.Close()
	r.response = `{"code":200,"message":"success","data":{"success":true,
		"credits_earned":35,"consecutive_days":2,"already_checked":false,
		"message":"签到成功！获得35积分","credits":1294}}`

	result, err := r.client().UserCheckin()
	if err != nil {
		t.Fatalf("checkin: %v", err)
	}
	want := &UserCheckin{
		Success:         true,
		CreditsEarned:   35,
		ConsecutiveDays: 2,
		AlreadyChecked:  false,
		Message:         "签到成功！获得35积分",
		Credits:         1294,
	}
	if *result != *want {
		t.Errorf("checkin = %+v, want %+v", result, want)
	}
}

func TestUserCheckinAlreadyCheckedIsNotAnError(t *testing.T) {
	// 「已领过」是正常回答而不是错误 —— 每日定时任务可以无脑重复执行它
	r := newRecorder(t)
	defer r.server.Close()
	r.response = `{"code":200,"message":"success","data":{"success":false,
		"credits_earned":0,"consecutive_days":2,"already_checked":true,
		"message":"今日已签到","credits":1294}}`

	result, err := r.client().UserCheckin()
	if err != nil {
		t.Fatalf("checkin: %v", err)
	}
	if !result.AlreadyChecked || result.Success || result.CreditsEarned != 0 {
		t.Errorf("checkin = %+v, want an already-checked answer with nothing granted", result)
	}
	if result.Message != "今日已签到" {
		t.Errorf("message = %q, want the platform's own wording", result.Message)
	}
}

package devicebase

import (
	"encoding/json"
	"testing"
)

func TestPointRequest(t *testing.T) {
	p := pointRequest{X: 100, Y: 200}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal pointRequest: %v", err)
	}
	want := `{"x":100,"y":200}`
	if string(data) != want {
		t.Errorf("marshal = %s, want %s", data, want)
	}

	var got pointRequest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal pointRequest: %v", err)
	}
	if got != p {
		t.Errorf("round-trip = %+v, want %+v", got, p)
	}
}

func TestBoundsRequest(t *testing.T) {
	b := boundsRequest{X1: 0, Y1: 100, X2: 200, Y2: 300}
	data, err := json.Marshal(b)
	if err != nil {
		t.Fatalf("marshal boundsRequest: %v", err)
	}
	want := `{"x1":0,"y1":100,"x2":200,"y2":300}`
	if string(data) != want {
		t.Errorf("marshal = %s, want %s", data, want)
	}
}

func TestLaunchAppRequest(t *testing.T) {
	r := launchAppRequest{AppName: "com.tencent.mm"}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal launchAppRequest: %v", err)
	}
	want := `{"app_name":"com.tencent.mm"}`
	if string(data) != want {
		t.Errorf("marshal = %s, want %s", data, want)
	}
}

func TestInputTextRequest(t *testing.T) {
	r := inputTextRequest{Text: "hello world"}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal inputTextRequest: %v", err)
	}
	want := `{"text":"hello world"}`
	if string(data) != want {
		t.Errorf("marshal = %s, want %s", data, want)
	}
}

func TestOperationResultSuccessField(t *testing.T) {
	result := &OperationResult{Success: true, Data: map[string]any{"success": true}}
	if !result.Success {
		t.Error("Success should be true")
	}
}

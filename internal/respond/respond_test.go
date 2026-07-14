package respond

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorResponseShape(t *testing.T) {
	rr := httptest.NewRecorder()
	Error(rr, http.StatusBadRequest, "bad request")

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if body["ok"] != false || body["error"] != "bad request" {
		t.Fatalf("body = %#v", body)
	}
}

func TestWebhookOKResponseShape(t *testing.T) {
	rr := httptest.NewRecorder()
	WebhookOK(rr, 123)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d", rr.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if body["ok"] != true || body["log_id"].(float64) != 123 {
		t.Fatalf("body = %#v", body)
	}
}

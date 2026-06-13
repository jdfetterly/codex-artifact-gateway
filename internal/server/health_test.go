package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpointReturnsJSON(t *testing.T) {
	root, feedbackDir, _ := fixture(t)
	handler := mustHandler(t, root, feedbackDir)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	var payload map[string]any
	if err := json.Unmarshal(res.Body.Bytes(), &payload); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if payload["ok"] != true {
		t.Fatalf("ok = %#v", payload["ok"])
	}
	if _, ok := payload["feedback_dir"]; ok {
		t.Fatalf("feedback_dir should not be exposed: %#v", payload["feedback_dir"])
	}
	if payload["root_count"].(float64) != 1 {
		t.Fatalf("root_count = %#v", payload["root_count"])
	}
	if _, ok := payload["uptime_seconds"]; !ok {
		t.Fatal("uptime_seconds missing")
	}
}

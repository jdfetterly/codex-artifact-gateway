package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdfetterly/codex-artifact-gateway/internal/gateway"
)

func TestOpenRedirectsToViewAndViewServesInjectedHTML(t *testing.T) {
	root, feedbackDir, page := fixture(t)
	handler := mustHandler(t, root, feedbackDir)

	openReq := httptest.NewRequest(http.MethodGet, "/open?path="+page, nil)
	openRes := httptest.NewRecorder()
	handler.ServeHTTP(openRes, openReq)

	if openRes.Code != http.StatusFound {
		t.Fatalf("/open status = %d, want %d", openRes.Code, http.StatusFound)
	}
	location := openRes.Header().Get("Location")
	if !strings.HasPrefix(location, "/view/"+filepath.Base(root)+"/runs/review.html") {
		t.Fatalf("Location = %q", location)
	}

	viewReq := httptest.NewRequest(http.MethodGet, location, nil)
	viewRes := httptest.NewRecorder()
	handler.ServeHTTP(viewRes, viewReq)

	if viewRes.Code != http.StatusOK {
		t.Fatalf("/view status = %d, want %d", viewRes.Code, http.StatusOK)
	}
	body := viewRes.Body.String()
	if !strings.Contains(body, "<button>Approve</button>") {
		t.Fatal("served HTML did not preserve page content")
	}
	if !strings.Contains(body, "codex-gateway-feedback") {
		t.Fatal("served HTML did not include feedback drawer")
	}
}

func TestViewServesRelativeAssets(t *testing.T) {
	root, feedbackDir, _ := fixture(t)
	handler := mustHandler(t, root, feedbackDir)

	req := httptest.NewRequest(http.MethodGet, "/view/"+filepath.Base(root)+"/runs/style.css", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if got := strings.Split(res.Header().Get("Content-Type"), ";")[0]; got != "text/css" {
		t.Fatalf("Content-Type = %q, want text/css", got)
	}
	if !strings.Contains(res.Body.String(), "color: black") {
		t.Fatal("asset body missing")
	}
}

func TestRecentListsHTMLWithMobileOverflowGuards(t *testing.T) {
	root, feedbackDir, _ := fixture(t)
	longPage := filepath.Join(root, "runs", "agent-capability-model-working-definition-2026-05-25.html")
	if err := os.WriteFile(longPage, []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	privatePage := filepath.Join(root, ".codex", "private.html")
	if err := os.MkdirAll(filepath.Dir(privatePage), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(privatePage, []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	handler := mustHandler(t, root, feedbackDir)

	req := httptest.NewRequest(http.MethodGet, "/recent", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	body := res.Body.String()
	if !strings.Contains(body, "review.html") {
		t.Fatal("recent page missing HTML file")
	}
	if !strings.Contains(body, "agent-<wbr>capability-<wbr>model") {
		t.Fatal("recent page missing rendered soft breaks for long file names")
	}
	if strings.Contains(body, "private.html") {
		t.Fatal("recent page listed private hidden artifact")
	}
	for _, guard := range []string{"overflow-wrap: anywhere", "grid-template-columns: 1fr", "word-break: break-word"} {
		if !strings.Contains(body, guard) {
			t.Fatalf("recent page missing mobile guard %q", guard)
		}
	}
}

func TestResolvePageAcceptsPastedPath(t *testing.T) {
	root, feedbackDir, page := fixture(t)
	handler := mustHandler(t, root, feedbackDir)

	req := httptest.NewRequest(http.MethodPost, "/resolve", strings.NewReader("path="+page))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusFound)
	}
	if !strings.Contains(res.Header().Get("Location"), "/view/"+filepath.Base(root)+"/runs/review.html") {
		t.Fatalf("unexpected Location %q", res.Header().Get("Location"))
	}
}

func TestResolvePageRejectsLargeRequestBody(t *testing.T) {
	root, feedbackDir, _ := fixture(t)
	handler := mustHandler(t, root, feedbackDir)

	req := httptest.NewRequest(http.MethodPost, "/resolve", strings.NewReader("path="+strings.Repeat("a", maxResolveBodyBytes+1)))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
}

func TestFeedbackEndpointWritesEntry(t *testing.T) {
	root, feedbackDir, _ := fixture(t)
	handler := mustHandler(t, root, feedbackDir)

	payload := gateway.FeedbackEntry{
		ArtifactPath: "/view/" + filepath.Base(root) + "/runs/review.html",
		Kind:         "looks_good",
		Comment:      "Ship it.",
		Href:         "http://127.0.0.1:8767/view/" + filepath.Base(root) + "/runs/review.html",
		Title:        "Review",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "UnitTest")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	files, err := filepath.Glob(filepath.Join(feedbackDir, "*-feedback.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatalf("feedback files = %d, want 1", len(files))
	}
	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Ship it.") || !strings.Contains(string(content), "UnitTest") {
		t.Fatalf("feedback log missing expected data: %s", string(content))
	}
}

func TestFeedbackEndpointRejectsLargeRequestBody(t *testing.T) {
	root, feedbackDir, _ := fixture(t)
	handler := mustHandler(t, root, feedbackDir)

	req := httptest.NewRequest(http.MethodPost, "/api/feedback", strings.NewReader(strings.Repeat("a", maxFeedbackBodyBytes+1)))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
}

func TestValidateListenAddrAllowsOnlyLoopback(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:8767", "localhost:8767", "[::1]:8767"} {
		if err := ValidateListenAddr(addr); err != nil {
			t.Fatalf("ValidateListenAddr(%q) returned error: %v", addr, err)
		}
	}
	for _, addr := range []string{"0.0.0.0:8767", ":8767", "192.168.1.5:8767", "example.com:8767"} {
		if err := ValidateListenAddr(addr); err == nil {
			t.Fatalf("ValidateListenAddr(%q) returned nil error", addr)
		}
	}
}

func fixture(t *testing.T) (root string, feedbackDir string, page string) {
	t.Helper()
	root = t.TempDir()
	feedbackDir = filepath.Join(t.TempDir(), "feedback")
	page = filepath.Join(root, "runs", "review.html")
	if err := os.MkdirAll(filepath.Dir(page), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(page, []byte("<html><head><title>Review</title></head><body><button>Approve</button></body></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "runs", "style.css"), []byte("body { color: black; }"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root, feedbackDir, page
}

func mustHandler(t *testing.T, root string, feedbackDir string) http.Handler {
	t.Helper()
	policy, err := gateway.NewPolicy([]string{root})
	if err != nil {
		t.Fatal(err)
	}
	return NewHandler(Config{
		Policy:      policy,
		FeedbackDir: feedbackDir,
	})
}

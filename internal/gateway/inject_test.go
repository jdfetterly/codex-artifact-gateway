package gateway

import (
	"strings"
	"testing"
)

func TestInjectFeedbackDrawerBeforeBodyClose(t *testing.T) {
	html := "<html><head><title>Review</title></head><body><main>Report</main></body></html>"

	result := InjectFeedbackDrawer([]byte(html), "/view/root/report.html")

	output := string(result)
	if !strings.Contains(output, "codex-gateway-feedback") {
		t.Fatal("feedback drawer marker missing")
	}
	if !strings.Contains(output, `data-artifact-path="/view/root/report.html"`) {
		t.Fatal("artifact path marker missing")
	}
	if strings.Index(output, "codex-gateway-feedback") > strings.Index(output, "</body>") {
		t.Fatal("drawer was not injected before body close")
	}
}

func TestInjectFeedbackDrawerAppendsWhenBodyCloseMissing(t *testing.T) {
	html := "<main>Report</main>"

	result := string(InjectFeedbackDrawer([]byte(html), "/view/root/report.html"))

	if !strings.HasPrefix(result, html) {
		t.Fatal("original HTML prefix was not preserved")
	}
	if !strings.Contains(result, "codex-gateway-feedback") {
		t.Fatal("feedback drawer marker missing")
	}
}

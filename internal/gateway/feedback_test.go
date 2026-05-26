package gateway

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeedbackStoreWritesJSONLEntry(t *testing.T) {
	dir := t.TempDir()
	store := FeedbackStore{Dir: dir}

	written, err := store.Append(FeedbackEntry{
		ArtifactPath: "/view/root/report.html",
		Kind:         "needs_changes",
		Comment:      "Tighten the conclusion.",
		Href:         "http://127.0.0.1:8767/view/root/report.html",
		Title:        "Review",
		UserAgent:    "Mobile Safari",
	})
	if err != nil {
		t.Fatalf("Append returned error: %v", err)
	}

	if filepath.Dir(written) != dir {
		t.Fatalf("feedback written to %q, want dir %q", written, dir)
	}
	content, err := os.ReadFile(written)
	if err != nil {
		t.Fatal(err)
	}
	lines := splitLines(string(content))
	if len(lines) != 1 {
		t.Fatalf("got %d JSONL lines, want 1", len(lines))
	}
	var entry FeedbackEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatal(err)
	}
	if entry.Kind != "needs_changes" || entry.ArtifactPath != "/view/root/report.html" {
		t.Fatalf("unexpected entry: %+v", entry)
	}
	if entry.CreatedAt.IsZero() {
		t.Fatal("CreatedAt was not set")
	}
}

func splitLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

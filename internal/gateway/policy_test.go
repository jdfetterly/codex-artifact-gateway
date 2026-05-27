package gateway

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveInputAcceptsFileURLInsideAllowedRoot(t *testing.T) {
	root := t.TempDir()
	page := filepath.Join(root, "reports", "daily.html")
	if err := os.MkdirAll(filepath.Dir(page), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(page, []byte("<html><body>Daily</body></html>"), 0o644); err != nil {
		t.Fatal(err)
	}

	policy, err := NewPolicy([]string{root})
	if err != nil {
		t.Fatal(err)
	}

	resolved, err := policy.ResolveInput("file://" + page)
	if err != nil {
		t.Fatalf("ResolveInput returned error: %v", err)
	}

	if resolved.RootName != filepath.Base(root) {
		t.Fatalf("RootName = %q, want %q", resolved.RootName, filepath.Base(root))
	}
	if resolved.RelativePath != "reports/daily.html" {
		t.Fatalf("RelativePath = %q, want reports/daily.html", resolved.RelativePath)
	}
	if resolved.ViewPath != "/view/"+filepath.Base(root)+"/reports/daily.html" {
		t.Fatalf("ViewPath = %q", resolved.ViewPath)
	}
}

func TestResolveInputRejectsOutsideRootAndTraversal(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "secret.html")
	if err := os.WriteFile(outside, []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	policy, err := NewPolicy([]string{root})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := policy.ResolveInput(outside); err == nil {
		t.Fatal("expected outside-root path to be rejected")
	}

	traversal := filepath.Join(root, "..", filepath.Base(outside), "secret.html")
	if _, err := policy.ResolveInput(traversal); err == nil {
		t.Fatal("expected traversal path to be rejected")
	}
}

func TestResolveInputRejectsUnsupportedFileTypes(t *testing.T) {
	root := t.TempDir()
	script := filepath.Join(root, "run.py")
	if err := os.WriteFile(script, []byte("print('no')"), 0o644); err != nil {
		t.Fatal(err)
	}
	policy, err := NewPolicy([]string{root})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := policy.ResolveInput(script); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported file type error, got %v", err)
	}
}

func TestResolveInputRejectsPrivateHomePathsWhenUsingBroadRoot(t *testing.T) {
	root := t.TempDir()
	privateHTML := filepath.Join(root, ".codex", "session.html")
	libraryHTML := filepath.Join(root, "Library", "Logs", "gateway.html")
	for _, path := range []string{privateHTML, libraryHTML} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("<html></html>"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	policy, err := NewPolicy([]string{root})
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{privateHTML, libraryHTML} {
		if _, err := policy.ResolveInput(path); err == nil || !strings.Contains(err.Error(), "private path component") {
			t.Fatalf("expected private path rejection for %s, got %v", path, err)
		}
	}
}

func TestResolveViewPathRejectsSymlinkComponents(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "report.html"), []byte("<html></html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Skipf("symlink unsupported on this system: %v", err)
	}
	policy, err := NewPolicy([]string{root})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := policy.ResolveViewPath(filepath.Base(root), "linked/report.html"); err == nil {
		t.Fatal("expected symlink component to be rejected")
	}
}

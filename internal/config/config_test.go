package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigRoundTripWithDefaultsAndExplicitRoot(t *testing.T) {
	home := t.TempDir()
	binaryPath := filepath.Join(home, "bin", "codex-artifact-gateway")
	root := filepath.Join(home, "Documents", "Codex")

	cfg := Default(home, binaryPath)
	cfg.Roots = []string{root}
	cfg.TailscaleCLIPath = "/Applications/Tailscale.app/Contents/MacOS/Tailscale"

	path := filepath.Join(home, "Library", "Application Support", "codex-artifact-gateway", "config.json")
	if err := Write(path, cfg); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	read, err := Read(path)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}

	if read.Addr != "127.0.0.1:8767" {
		t.Fatalf("Addr = %q", read.Addr)
	}
	if read.FeedbackDir != filepath.Join(home, "Documents", "Codex", "codex-artifact-gateway-feedback") {
		t.Fatalf("FeedbackDir = %q", read.FeedbackDir)
	}
	if read.BinaryPath != binaryPath {
		t.Fatalf("BinaryPath = %q", read.BinaryPath)
	}
	if read.LaunchAgentLabel != "com.jdfetterly.codex-artifact-gateway" {
		t.Fatalf("LaunchAgentLabel = %q", read.LaunchAgentLabel)
	}
	if !read.ManageTailscale {
		t.Fatal("ManageTailscale = false, want true")
	}
	if len(read.Roots) != 1 || read.Roots[0] != root {
		t.Fatalf("Roots = %#v", read.Roots)
	}
}

func TestPathHelpersUseMacOSUserLocations(t *testing.T) {
	home := t.TempDir()

	if got := ConfigPath(home); got != filepath.Join(home, "Library", "Application Support", "codex-artifact-gateway", "config.json") {
		t.Fatalf("ConfigPath = %q", got)
	}
	if got := LaunchAgentPath(home); got != filepath.Join(home, "Library", "LaunchAgents", "com.jdfetterly.codex-artifact-gateway.plist") {
		t.Fatalf("LaunchAgentPath = %q", got)
	}
	if got := LogPath(home, "out"); got != filepath.Join(home, "Library", "Logs", "codex-artifact-gateway.out.log") {
		t.Fatalf("LogPath = %q", got)
	}
}

func TestWriteCreatesParentDirectory(t *testing.T) {
	home := t.TempDir()
	path := ConfigPath(home)
	cfg := Default(home, "/tmp/codex-artifact-gateway")

	if err := Write(path, cfg); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Dir(path)); err != nil {
		t.Fatalf("expected parent directory to exist: %v", err)
	}
}

func TestWriteUsesPrivateConfigPermissions(t *testing.T) {
	home := t.TempDir()
	path := ConfigPath(home)
	cfg := Default(home, "/tmp/codex-artifact-gateway")

	if err := Write(path, cfg); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("config dir mode = %#o, want 0700", got)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("config file mode = %#o, want 0600", got)
	}
}

func TestWriteRepairsPermissiveConfigPermissions(t *testing.T) {
	home := t.TempDir()
	path := ConfigPath(home)
	cfg := Default(home, "/tmp/codex-artifact-gateway")

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Write(path, cfg); err != nil {
		t.Fatalf("Write returned error: %v", err)
	}
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("config dir mode = %#o, want 0700", got)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := fileInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("config file mode = %#o, want 0600", got)
	}
}

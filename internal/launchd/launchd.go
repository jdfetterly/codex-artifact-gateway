package launchd

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Label      string
	Program    string
	ConfigPath string
	StdoutPath string
	StderrPath string
}

type RunFunc func(name string, args ...string) (string, error)

type Manager struct {
	UID int
	Run RunFunc
}

func Plist(cfg Config) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>` + xml(cfg.Label) + `</string>
  <key>ProgramArguments</key>
  <array>
    <string>` + xml(cfg.Program) + `</string>
    <string>serve</string>
    <string>--config</string>
    <string>` + xml(cfg.ConfigPath) + `</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>` + xml(cfg.StdoutPath) + `</string>
  <key>StandardErrorPath</key>
  <string>` + xml(cfg.StderrPath) + `</string>
</dict>
</plist>
`
}

func WritePlist(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	for _, logPath := range []string{cfg.StdoutPath, cfg.StderrPath} {
		if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
			return err
		}
	}
	return os.WriteFile(path, []byte(Plist(cfg)), 0o644)
}

func (m Manager) Load(plistPath string) error {
	_, err := m.run("launchctl", "bootstrap", m.domain(), plistPath)
	if err != nil && !strings.Contains(err.Error(), "already") && !strings.Contains(err.Error(), "Bootstrap failed: 5") {
		return err
	}
	return nil
}

func (m Manager) Start(label string) error {
	_, err := m.run("launchctl", "kickstart", "-k", m.domain()+"/"+label)
	return err
}

func (m Manager) Stop(label string) error {
	_, err := m.run("launchctl", "bootout", m.domain()+"/"+label)
	if err != nil && !strings.Contains(err.Error(), "No such process") && !strings.Contains(err.Error(), "Could not find service") {
		return err
	}
	return nil
}

func (m Manager) Status(label string) (string, error) {
	out, err := m.run("launchctl", "list", label)
	if err != nil {
		return "not loaded", nil
	}
	if strings.TrimSpace(out) == "" {
		return "loaded", nil
	}
	return strings.TrimSpace(out), nil
}

func (m Manager) domain() string {
	return fmt.Sprintf("gui/%d", m.UID)
}

func (m Manager) run(name string, args ...string) (string, error) {
	if m.Run == nil {
		return "", fmt.Errorf("launchd runner not configured")
	}
	return m.Run(name, args...)
}

func xml(s string) string {
	return html.EscapeString(s)
}

package launchd

import (
	"strings"
	"testing"
)

func TestPlistIncludesStableBinaryConfigLogsAndLabel(t *testing.T) {
	plist := Plist(Config{
		Label:      "com.jdfetterly.codex-artifact-gateway",
		Program:    "/usr/local/bin/codex-artifact-gateway",
		ConfigPath: "/Users/example/Library/Application Support/codex-artifact-gateway/config.json",
		StdoutPath: "/Users/example/Library/Logs/codex-artifact-gateway.out.log",
		StderrPath: "/Users/example/Library/Logs/codex-artifact-gateway.err.log",
	})

	for _, want := range []string{
		"<string>com.jdfetterly.codex-artifact-gateway</string>",
		"<string>/usr/local/bin/codex-artifact-gateway</string>",
		"<string>serve</string>",
		"<string>--config</string>",
		"<string>/Users/example/Library/Application Support/codex-artifact-gateway/config.json</string>",
		"<key>StandardOutPath</key>",
		"<string>/Users/example/Library/Logs/codex-artifact-gateway.out.log</string>",
		"<key>StandardErrorPath</key>",
		"<string>/Users/example/Library/Logs/codex-artifact-gateway.err.log</string>",
	} {
		if !strings.Contains(plist, want) {
			t.Fatalf("plist missing %q\n%s", want, plist)
		}
	}
}

func TestManagerUsesBootstrapKickstartAndBootout(t *testing.T) {
	var calls []string
	manager := Manager{
		UID: 501,
		Run: func(name string, args ...string) (string, error) {
			calls = append(calls, name+" "+strings.Join(args, " "))
			return "", nil
		},
	}

	if err := manager.Load("/tmp/agent.plist"); err != nil {
		t.Fatal(err)
	}
	if err := manager.Start("com.jdfetterly.codex-artifact-gateway"); err != nil {
		t.Fatal(err)
	}
	if err := manager.Stop("com.jdfetterly.codex-artifact-gateway"); err != nil {
		t.Fatal(err)
	}

	joined := strings.Join(calls, "\n")
	for _, want := range []string{
		"launchctl bootstrap gui/501 /tmp/agent.plist",
		"launchctl kickstart -k gui/501/com.jdfetterly.codex-artifact-gateway",
		"launchctl bootout gui/501/com.jdfetterly.codex-artifact-gateway",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing command %q in:\n%s", want, joined)
		}
	}
}

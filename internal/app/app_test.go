package app

import (
	"strings"
	"testing"

	"github.com/jdfetterly/codex-artifact-gateway/internal/config"
)

func TestSetupWritesConfigStartsLaunchAgentTailscaleAndChecksHealth(t *testing.T) {
	var calls []string
	var saved config.Config
	runner := FakeRunner{
		WriteConfigFunc: func(path string, cfg config.Config) error {
			calls = append(calls, "write-config")
			saved = cfg
			return nil
		},
		InstallLaunchAgentFunc: func(cfg config.Config, configPath string) error {
			calls = append(calls, "install-launch-agent")
			return nil
		},
		StartLaunchAgentFunc: func(label string) error {
			calls = append(calls, "start-launch-agent")
			return nil
		},
		CheckHealthFunc: func(addr string) error {
			calls = append(calls, "check-health "+addr)
			return nil
		},
		StartTailscaleServeFunc: func(path string, addr string) (string, error) {
			calls = append(calls, "tailscale-serve "+path+" "+addr)
			return "https://example.tail.ts.net/", nil
		},
	}

	out, err := Setup(SetupOptions{
		Home:         "/Users/jd",
		Roots:        []string{"/Users/jd/Documents/Codex", "/Users/jd/Reference"},
		BinaryPath:   "/Users/jd/bin/codex-artifact-gateway",
		TailscaleCLI: "/Applications/Tailscale.app/Contents/MacOS/Tailscale",
		Runner:       runner,
	})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Join(saved.Roots, ",") != "/Users/jd/Documents/Codex,/Users/jd/Reference" {
		t.Fatalf("saved roots = %#v", saved.Roots)
	}
	if !strings.Contains(out, "https://example.tail.ts.net/recent") {
		t.Fatalf("setup output missing URL: %s", out)
	}
	want := "write-config,install-launch-agent,start-launch-agent,check-health 127.0.0.1:8767,tailscale-serve /Applications/Tailscale.app/Contents/MacOS/Tailscale 127.0.0.1:8767"
	if strings.Join(calls, ",") != want {
		t.Fatalf("calls = %q, want %q", strings.Join(calls, ","), want)
	}
}

func TestStopStopsLaunchAgentAndManagedTailscaleServe(t *testing.T) {
	var calls []string
	runner := FakeRunner{
		ReadConfigFunc: func(path string) (config.Config, error) {
			return config.Config{
				Addr:             "127.0.0.1:8767",
				TailscaleCLIPath: "/Applications/Tailscale.app/Contents/MacOS/Tailscale",
				ManageTailscale:  true,
				LaunchAgentLabel: config.DefaultLaunchAgentLabel,
			}, nil
		},
		StopLaunchAgentFunc: func(label string) error {
			calls = append(calls, "stop-launch-agent "+label)
			return nil
		},
		StopTailscaleServeFunc: func(path string) error {
			calls = append(calls, "tailscale-off "+path)
			return nil
		},
	}

	out, err := Stop(StopOptions{Home: "/Users/jd", Runner: runner})
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "stopped") {
		t.Fatalf("unexpected output: %s", out)
	}
	want := "stop-launch-agent com.jdfetterly.codex-artifact-gateway,tailscale-off /Applications/Tailscale.app/Contents/MacOS/Tailscale"
	if strings.Join(calls, ",") != want {
		t.Fatalf("calls = %q, want %q", strings.Join(calls, ","), want)
	}
}

func TestStatusReportsDegradedState(t *testing.T) {
	runner := FakeRunner{
		ReadConfigFunc: func(path string) (config.Config, error) {
			return config.Default("/Users/jd", "/Users/jd/bin/codex-artifact-gateway"), nil
		},
		LaunchAgentStatusFunc: func(label string) (string, error) {
			return "not loaded", nil
		},
		CheckHealthFunc: func(addr string) error {
			return errFake("connection refused")
		},
		TailscaleStatusFunc: func(path string) (string, error) {
			return "", errFake("tailscale missing")
		},
		TailscaleServeStatusFunc: func(path string) (string, error) {
			return "", errFake("serve missing")
		},
	}

	out, err := Status(StatusOptions{Home: "/Users/jd", Runner: runner})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"roots: /Users/jd/Documents/Codex", "service: not loaded", "health: connection refused", "tailscale: tailscale missing", "serve: serve missing"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status missing %q:\n%s", want, out)
		}
	}
}

func TestStatusReportsTailscaleServeURLAndIPhone(t *testing.T) {
	runner := FakeRunner{
		ReadConfigFunc: func(path string) (config.Config, error) {
			cfg := config.Default("/Users/jd", "/Users/jd/bin/codex-artifact-gateway")
			cfg.TailscaleCLIPath = "/Applications/Tailscale.app/Contents/MacOS/Tailscale"
			return cfg, nil
		},
		LaunchAgentStatusFunc: func(label string) (string, error) {
			return "loaded", nil
		},
		CheckHealthFunc: func(addr string) error {
			return nil
		},
		TailscaleStatusFunc: func(path string) (string, error) {
			return "100.84.76.13 iphone182 jdfetterly@ iOS idle", nil
		},
		TailscaleServeStatusFunc: func(path string) (string, error) {
			return "https://jds-macbook-pro.tail13d577.ts.net (tailnet only)\n|-- / proxy http://127.0.0.1:8767", nil
		},
	}

	out, err := Status(StatusOptions{Home: "/Users/jd", Runner: runner})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"health: ok", "iphone: visible on tailnet", "serve: https://jds-macbook-pro.tail13d577.ts.net/recent"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status missing %q:\n%s", want, out)
		}
	}
}

type errFake string

func (e errFake) Error() string { return string(e) }

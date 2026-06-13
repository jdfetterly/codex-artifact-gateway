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
			return "https://gateway.example.com/", nil
		},
	}

	out, err := Setup(SetupOptions{
		Home:         "/Users/example",
		Roots:        []string{"/Users/example/Documents/Codex", "/Users/example/Reference"},
		BinaryPath:   "/Users/example/bin/codex-artifact-gateway",
		TailscaleCLI: "/Applications/Tailscale.app/Contents/MacOS/Tailscale",
		Runner:       runner,
	})
	if err != nil {
		t.Fatal(err)
	}

	if strings.Join(saved.Roots, ",") != "/Users/example/Documents/Codex,/Users/example/Reference" {
		t.Fatalf("saved roots = %#v", saved.Roots)
	}
	for _, want := range []string{
		"Open this on your iPhone:\nhttps://gateway.example.com/",
		"Gateway can open HTML files from these folders and their subfolders:",
		"- /Users/example/Documents/Codex",
		"- /Users/example/Reference",
		"Gateway was installed from:\n/Users/example/bin/codex-artifact-gateway",
		"./codex-artifact-gateway doctor",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("setup output missing %q:\n%s", want, out)
		}
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

	out, err := Stop(StopOptions{Home: "/Users/example", Runner: runner})
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
			return config.Default("/Users/example", "/Users/example/bin/codex-artifact-gateway"), nil
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

	out, err := Status(StatusOptions{Home: "/Users/example", Runner: runner})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Gateway status",
		"Not available yet. Run setup or check Tailscale Serve.",
		"Gateway can open HTML files from these folders and their subfolders:",
		"- /Users/example/Documents/Codex",
		"Gateway is not running at http://127.0.0.1:8767.",
		"Tailscale check failed: tailscale missing",
		"Install location:\n/Users/example/bin/codex-artifact-gateway",
		"Feedback is saved here:\n/Users/example/Documents/Codex/codex-artifact-gateway-feedback",
		"Service:\nnot loaded",
		"Tailscale Serve:\nserve missing",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("status missing %q:\n%s", want, out)
		}
	}
}

func TestStatusReportsTailscaleServeURLAndIPhone(t *testing.T) {
	runner := FakeRunner{
		ReadConfigFunc: func(path string) (config.Config, error) {
			cfg := config.Default("/Users/example", "/Users/example/bin/codex-artifact-gateway")
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
			return "100.84.76.13 iphone user@example iOS idle", nil
		},
		TailscaleServeStatusFunc: func(path string) (string, error) {
			return "https://macbook.example.com (tailnet only)\n|-- / proxy http://127.0.0.1:8767", nil
		},
	}

	out, err := Status(StatusOptions{Home: "/Users/example", Runner: runner})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Phone URL:\nhttps://macbook.example.com/",
		"Local check:\nhttp://127.0.0.1:8767 is working on this Mac.",
		"Your private phone URL is connected through Tailscale Serve.",
		"Tailscale Serve:\nhttps://macbook.example.com/",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("status missing %q:\n%s", want, out)
		}
	}
}

func TestSetupWarnsAboutUnstableInstallLocation(t *testing.T) {
	runner := FakeRunner{
		WriteConfigFunc: func(path string, cfg config.Config) error {
			return nil
		},
		InstallLaunchAgentFunc: func(cfg config.Config, configPath string) error {
			return nil
		},
		StartLaunchAgentFunc: func(label string) error {
			return nil
		},
		CheckHealthFunc: func(addr string) error {
			return nil
		},
		StartTailscaleServeFunc: func(path string, addr string) (string, error) {
			return "https://gateway.example.com/", nil
		},
	}

	out, err := Setup(SetupOptions{
		Home:         "/Users/example",
		Roots:        []string{"/Users/example/Documents/Codex"},
		BinaryPath:   "/Users/example/Downloads/codex-artifact-gateway",
		TailscaleCLI: "/Applications/Tailscale.app/Contents/MacOS/Tailscale",
		Runner:       runner,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Warning: Gateway is installed from a folder that may be moved or deleted.") {
		t.Fatalf("setup output missing unstable install warning:\n%s", out)
	}
}

func TestDoctorIncludesBeginnerTroubleshooting(t *testing.T) {
	runner := FakeRunner{
		ReadConfigFunc: func(path string) (config.Config, error) {
			cfg := config.Default("/Users/example", "/Users/example/bin/codex-artifact-gateway")
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
			return "100.84.76.13 iphone user@example iOS idle", nil
		},
		TailscaleServeStatusFunc: func(path string) (string, error) {
			return "https://macbook.example.com (tailnet only)\n|-- / proxy http://127.0.0.1:8767", nil
		},
	}

	out, err := Doctor(StatusOptions{Home: "/Users/example", Runner: runner})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"If you do not see your file on the phone:",
		"Make sure it is an .html file.",
		"Open the phone URL and choose \"Paste a file path\"",
		"./codex-artifact-gateway status",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("doctor output missing %q:\n%s", want, out)
		}
	}
}

type errFake string

func (e errFake) Error() string { return string(e) }

package tailscale

import (
	"errors"
	"strings"
	"testing"
)

func TestDetectCLIUsesPathThenAppBundle(t *testing.T) {
	pathCLI, err := DetectCLI(func(name string) (string, error) {
		if name == "tailscale" {
			return "/opt/homebrew/bin/tailscale", nil
		}
		return "", errors.New("not found")
	}, "/Applications/Tailscale.app/Contents/MacOS/Tailscale", func(path string) bool {
		return true
	})
	if err != nil {
		t.Fatal(err)
	}
	if pathCLI != "/opt/homebrew/bin/tailscale" {
		t.Fatalf("DetectCLI = %q", pathCLI)
	}

	appCLI, err := DetectCLI(func(name string) (string, error) {
		return "", errors.New("not found")
	}, "/Applications/Tailscale.app/Contents/MacOS/Tailscale", func(path string) bool {
		return path == "/Applications/Tailscale.app/Contents/MacOS/Tailscale"
	})
	if err != nil {
		t.Fatal(err)
	}
	if appCLI != "/Applications/Tailscale.app/Contents/MacOS/Tailscale" {
		t.Fatalf("DetectCLI fallback = %q", appCLI)
	}
}

func TestServeURLParsesTailnetURL(t *testing.T) {
	output := `Available within your tailnet:

https://jds-macbook-pro.tail13d577.ts.net/
|-- proxy http://127.0.0.1:8767
`
	got := ParseServeURL(output)
	if got != "https://jds-macbook-pro.tail13d577.ts.net/" {
		t.Fatalf("ParseServeURL = %q", got)
	}
}

func TestClientStartsAndStopsServe(t *testing.T) {
	var calls []string
	client := Client{
		Path: "/Applications/Tailscale.app/Contents/MacOS/Tailscale",
		Run: func(name string, args ...string) (string, error) {
			calls = append(calls, name+" "+strings.Join(args, " "))
			if args[0] == "serve" && args[1] == "--bg" {
				return "https://example.tail.ts.net/\n|-- proxy http://127.0.0.1:8767", nil
			}
			return "", nil
		},
	}

	url, err := client.StartServe("127.0.0.1:8767")
	if err != nil {
		t.Fatal(err)
	}
	if url != "https://example.tail.ts.net/" {
		t.Fatalf("StartServe url = %q", url)
	}
	if err := client.StopServe(); err != nil {
		t.Fatal(err)
	}

	joined := strings.Join(calls, "\n")
	for _, want := range []string{
		"/Applications/Tailscale.app/Contents/MacOS/Tailscale serve --bg http://127.0.0.1:8767",
		"/Applications/Tailscale.app/Contents/MacOS/Tailscale serve --https=443 off",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing call %q in:\n%s", want, joined)
		}
	}
}

func TestStatusDetectsIPhonePresence(t *testing.T) {
	status := `100.111.16.45   jds-macbook-pro  jdfetterly@  macOS  -
100.84.76.13    iphone182        jdfetterly@  iOS    idle, tx 48612 rx 16052`

	if !HasIPhone(status) {
		t.Fatal("expected iPhone to be detected")
	}
}

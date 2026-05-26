package tailscale

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const AppBundleCLI = "/Applications/Tailscale.app/Contents/MacOS/Tailscale"

type RunFunc func(name string, args ...string) (string, error)

type Client struct {
	Path string
	Run  RunFunc
}

func DetectCLI(lookPath func(string) (string, error), appBundlePath string, exists func(string) bool) (string, error) {
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	if exists == nil {
		exists = fileExists
	}
	if path, err := lookPath("tailscale"); err == nil && path != "" {
		return path, nil
	}
	if exists(appBundlePath) {
		return appBundlePath, nil
	}
	return "", fmt.Errorf("tailscale CLI not found")
}

func ParseServeURL(output string) string {
	re := regexp.MustCompile(`https://[^\s]+`)
	match := re.FindString(output)
	return match
}

func HasIPhone(status string) bool {
	lower := strings.ToLower(status)
	return strings.Contains(lower, "iphone") && strings.Contains(lower, "ios")
}

func (c Client) StartServe(addr string) (string, error) {
	out, err := c.run(c.Path, "serve", "--bg", "http://"+addr)
	if err != nil {
		return "", err
	}
	url := ParseServeURL(out)
	if url == "" {
		return "", fmt.Errorf("could not find Tailscale Serve URL in output: %s", out)
	}
	return url, nil
}

func (c Client) StopServe() error {
	_, err := c.run(c.Path, "serve", "--https=443", "off")
	return err
}

func (c Client) ServeStatus() (string, error) {
	out, err := c.run(c.Path, "serve", "status")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (c Client) Status() (string, error) {
	out, err := c.run(c.Path, "status", "--self")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (c Client) run(name string, args ...string) (string, error) {
	if c.Run == nil {
		return "", fmt.Errorf("tailscale runner not configured")
	}
	return c.Run(name, args...)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

package app

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jdfetterly/codex-artifact-gateway/internal/config"
	"github.com/jdfetterly/codex-artifact-gateway/internal/launchd"
	"github.com/jdfetterly/codex-artifact-gateway/internal/tailscale"
)

type Runner interface {
	WriteConfig(path string, cfg config.Config) error
	ReadConfig(path string) (config.Config, error)
	InstallLaunchAgent(cfg config.Config, configPath string) error
	StartLaunchAgent(label string) error
	StopLaunchAgent(label string) error
	LaunchAgentStatus(label string) (string, error)
	CheckHealth(addr string) error
	StartTailscaleServe(path string, addr string) (string, error)
	StopTailscaleServe(path string) error
	TailscaleStatus(path string) (string, error)
	TailscaleServeStatus(path string) (string, error)
}

type SetupOptions struct {
	Home         string
	Root         string
	Roots        []string
	BinaryPath   string
	TailscaleCLI string
	Runner       Runner
}

type StartOptions struct {
	Home   string
	Runner Runner
}

type StopOptions struct {
	Home   string
	Runner Runner
}

type StatusOptions struct {
	Home   string
	Runner Runner
}

func Setup(opts SetupOptions) (string, error) {
	runner := opts.Runner
	if runner == nil {
		runner = SystemRunner{Home: opts.Home}
	}
	roots := opts.Roots
	if len(roots) == 0 && opts.Root != "" {
		roots = []string{opts.Root}
	}
	if len(roots) == 0 {
		roots = []string{config.DefaultRoot(opts.Home)}
	}
	cfg := config.Default(opts.Home, opts.BinaryPath)
	cfg.Roots = roots
	cfg.TailscaleCLIPath = opts.TailscaleCLI
	if cfg.TailscaleCLIPath == "" {
		detected, err := tailscale.DetectCLI(exec.LookPath, tailscale.AppBundleCLI, nil)
		if err != nil {
			return "", err
		}
		cfg.TailscaleCLIPath = detected
	}
	configPath := config.ConfigPath(opts.Home)
	if err := runner.WriteConfig(configPath, cfg); err != nil {
		return "", err
	}
	if err := runner.InstallLaunchAgent(cfg, configPath); err != nil {
		return "", err
	}
	if err := runner.StartLaunchAgent(cfg.LaunchAgentLabel); err != nil {
		return "", err
	}
	if err := runner.CheckHealth(cfg.Addr); err != nil {
		return "", err
	}
	url := ""
	if cfg.ManageTailscale {
		serveURL, err := runner.StartTailscaleServe(cfg.TailscaleCLIPath, cfg.Addr)
		if err != nil {
			return "", err
		}
		url = strings.TrimRight(serveURL, "/") + "/recent"
	}
	return fmt.Sprintf("setup complete\nrecent: %s\nconfig: %s\nfeedback: %s\n", url, configPath, cfg.FeedbackDir), nil
}

func Start(opts StartOptions) (string, error) {
	runner := opts.Runner
	if runner == nil {
		runner = SystemRunner{Home: opts.Home}
	}
	cfg, err := runner.ReadConfig(config.ConfigPath(opts.Home))
	if err != nil {
		return "", err
	}
	if err := runner.StartLaunchAgent(cfg.LaunchAgentLabel); err != nil {
		return "", err
	}
	if err := runner.CheckHealth(cfg.Addr); err != nil {
		return "", err
	}
	url := ""
	if cfg.ManageTailscale {
		serveURL, err := runner.StartTailscaleServe(cfg.TailscaleCLIPath, cfg.Addr)
		if err != nil {
			return "", err
		}
		url = strings.TrimRight(serveURL, "/") + "/recent"
	}
	return fmt.Sprintf("started\nrecent: %s\n", url), nil
}

func Stop(opts StopOptions) (string, error) {
	runner := opts.Runner
	if runner == nil {
		runner = SystemRunner{Home: opts.Home}
	}
	cfg, err := runner.ReadConfig(config.ConfigPath(opts.Home))
	if err != nil {
		return "", err
	}
	if err := runner.StopLaunchAgent(cfg.LaunchAgentLabel); err != nil {
		return "", err
	}
	if cfg.ManageTailscale {
		if err := runner.StopTailscaleServe(cfg.TailscaleCLIPath); err != nil {
			return "", err
		}
	}
	return "stopped\n", nil
}

func Status(opts StatusOptions) (string, error) {
	runner := opts.Runner
	if runner == nil {
		runner = SystemRunner{Home: opts.Home}
	}
	cfg, err := runner.ReadConfig(config.ConfigPath(opts.Home))
	if err != nil {
		return fmt.Sprintf("config: %v\nrepair: codex-artifact-gateway setup --root %s\n", err, config.DefaultRoot(opts.Home)), nil
	}
	var builder strings.Builder
	builder.WriteString("config: " + config.ConfigPath(opts.Home) + "\n")
	builder.WriteString("addr: " + cfg.Addr + "\n")
	builder.WriteString("roots: " + strings.Join(cfg.Roots, ", ") + "\n")
	builder.WriteString("feedback: " + cfg.FeedbackDir + "\n")
	service, err := runner.LaunchAgentStatus(cfg.LaunchAgentLabel)
	if err != nil {
		service = err.Error()
	}
	builder.WriteString("service: " + service + "\n")
	if err := runner.CheckHealth(cfg.Addr); err != nil {
		builder.WriteString("health: " + err.Error() + "\n")
	} else {
		builder.WriteString("health: ok\n")
	}
	tailscalePath := cfg.TailscaleCLIPath
	if tailscalePath == "" {
		tailscalePath = tailscale.AppBundleCLI
	}
	status, err := runner.TailscaleStatus(tailscalePath)
	if err != nil {
		builder.WriteString("tailscale: " + err.Error() + "\n")
	} else {
		builder.WriteString("tailscale: ok\n")
		if tailscale.HasIPhone(status) {
			builder.WriteString("iphone: visible on tailnet\n")
		}
	}
	serveStatus, err := runner.TailscaleServeStatus(tailscalePath)
	if err != nil {
		builder.WriteString("serve: " + err.Error() + "\n")
	} else if serveURL := tailscale.ParseServeURL(serveStatus); serveURL != "" {
		builder.WriteString("serve: " + strings.TrimRight(serveURL, "/") + "/recent\n")
	} else {
		builder.WriteString("serve: not configured\n")
	}
	return builder.String(), nil
}

func Doctor(opts StatusOptions) (string, error) {
	status, err := Status(opts)
	if err != nil {
		return "", err
	}
	return status + "repair commands:\n  codex-artifact-gateway setup --root " + config.DefaultRoot(opts.Home) + "\n  codex-artifact-gateway status\n", nil
}

type SystemRunner struct {
	Home string
}

func (r SystemRunner) WriteConfig(path string, cfg config.Config) error {
	return config.Write(path, cfg)
}

func (r SystemRunner) ReadConfig(path string) (config.Config, error) {
	return config.Read(path)
}

func (r SystemRunner) InstallLaunchAgent(cfg config.Config, configPath string) error {
	plistPath := config.LaunchAgentPath(r.Home)
	return launchd.WritePlist(plistPath, launchd.Config{
		Label:      cfg.LaunchAgentLabel,
		Program:    cfg.BinaryPath,
		ConfigPath: configPath,
		StdoutPath: config.LogPath(r.Home, "out"),
		StderrPath: config.LogPath(r.Home, "err"),
	})
}

func (r SystemRunner) StartLaunchAgent(label string) error {
	manager := launchd.Manager{UID: os.Getuid(), Run: runCommand}
	plistPath := config.LaunchAgentPath(r.Home)
	if err := manager.Load(plistPath); err != nil {
		return err
	}
	return manager.Start(label)
}

func (r SystemRunner) StopLaunchAgent(label string) error {
	manager := launchd.Manager{UID: os.Getuid(), Run: runCommand}
	return manager.Stop(label)
}

func (r SystemRunner) LaunchAgentStatus(label string) (string, error) {
	manager := launchd.Manager{UID: os.Getuid(), Run: runCommand}
	return manager.Status(label)
}

func (r SystemRunner) CheckHealth(addr string) error {
	client := http.Client{Timeout: 3 * time.Second}
	var lastErr error
	for i := 0; i < 20; i++ {
		res, err := client.Get("http://" + addr + "/health")
		if err != nil {
			lastErr = err
			time.Sleep(250 * time.Millisecond)
			continue
		}
		if res.StatusCode == http.StatusOK {
			_ = res.Body.Close()
			return nil
		}
		lastErr = fmt.Errorf("health returned %s", res.Status)
		_ = res.Body.Close()
		time.Sleep(250 * time.Millisecond)
	}
	return lastErr
}

func (r SystemRunner) StartTailscaleServe(path string, addr string) (string, error) {
	return tailscale.Client{Path: path, Run: runCommand}.StartServe(addr)
}

func (r SystemRunner) StopTailscaleServe(path string) error {
	return tailscale.Client{Path: path, Run: runCommand}.StopServe()
}

func (r SystemRunner) TailscaleStatus(path string) (string, error) {
	return tailscale.Client{Path: path, Run: runCommand}.Status()
}

func (r SystemRunner) TailscaleServeStatus(path string) (string, error) {
	return tailscale.Client{Path: path, Run: runCommand}.ServeStatus()
}

func StableBinaryPath(home string) (string, error) {
	cwd, err := os.Getwd()
	if err == nil {
		local := filepath.Join(cwd, "codex-artifact-gateway")
		if info, statErr := os.Stat(local); statErr == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
			return local, nil
		}
	}
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	if strings.Contains(exe, "go-build") || strings.HasPrefix(exe, os.TempDir()) {
		return "", fmt.Errorf("setup needs a stable binary; run: go build ./cmd/codex-artifact-gateway")
	}
	return exe, nil
}

func runCommand(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

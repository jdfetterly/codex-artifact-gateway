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
		url = baseURL(serveURL)
	}
	return setupMessage(url, cfg, configPath), nil
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
		url = baseURL(serveURL)
	}
	return fmt.Sprintf("Gateway started.\n\nOpen this on your iPhone:\n%s\n", url), nil
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
		return fmt.Sprintf("Gateway is not set up yet.\n\nRun:\n  ./codex-artifact-gateway setup --root %s\n\nDetails:\n  %v\n", config.DefaultRoot(opts.Home), err), nil
	}
	var builder strings.Builder
	service, err := runner.LaunchAgentStatus(cfg.LaunchAgentLabel)
	if err != nil {
		service = err.Error()
	}
	builder.WriteString("Gateway status\n\n")
	builder.WriteString("Phone URL:\n")
	serveURL := ""
	tailscalePath := cfg.TailscaleCLIPath
	if tailscalePath == "" {
		tailscalePath = tailscale.AppBundleCLI
	}
	serveStatus, serveErr := runner.TailscaleServeStatus(tailscalePath)
	if serveErr == nil {
		serveURL = baseURL(tailscale.ParseServeURL(serveStatus))
	}
	if serveURL == "" {
		builder.WriteString("Not available yet. Run setup or check Tailscale Serve.\n\n")
	} else {
		builder.WriteString(serveURL + "\n\n")
	}
	builder.WriteString("Gateway can open HTML files from these folders and their subfolders:\n")
	builder.WriteString(formatRoots(cfg.Roots))
	builder.WriteString("\n")
	builder.WriteString("Local check:\n")
	if err := runner.CheckHealth(cfg.Addr); err != nil {
		builder.WriteString("Gateway is not running at http://" + cfg.Addr + ".\n")
		builder.WriteString("Run:\n  ./codex-artifact-gateway setup --root " + firstRootOrDefault(cfg.Roots, opts.Home) + "\n\n")
	} else {
		builder.WriteString("http://" + cfg.Addr + " is working on this Mac.\n\n")
	}
	builder.WriteString("Tailscale check:\n")
	_, err = runner.TailscaleStatus(tailscalePath)
	if err != nil {
		builder.WriteString("Tailscale check failed: " + err.Error() + "\n\n")
	} else if serveURL != "" {
		builder.WriteString("Your private phone URL is connected through Tailscale Serve.\n\n")
	} else {
		builder.WriteString("Tailscale is running on this Mac, but Gateway does not have a phone URL yet.\n\n")
	}
	builder.WriteString("Install location:\n")
	builder.WriteString(cfg.BinaryPath + "\n")
	if warning := unstableInstallWarning(cfg.BinaryPath); warning != "" {
		builder.WriteString("\n" + warning + "\n")
	}
	builder.WriteString("\nFeedback is saved here:\n")
	builder.WriteString(cfg.FeedbackDir + "\n\n")
	builder.WriteString("Config file:\n")
	builder.WriteString(config.ConfigPath(opts.Home) + "\n\n")
	builder.WriteString("Service:\n")
	builder.WriteString(service + "\n\n")
	builder.WriteString("Tailscale Serve:\n")
	if serveErr != nil {
		builder.WriteString(serveErr.Error() + "\n")
	} else if serveURL != "" {
		builder.WriteString(serveURL + "\n")
	} else {
		builder.WriteString("not configured\n")
	}
	return builder.String(), nil
}

func Doctor(opts StatusOptions) (string, error) {
	status, err := Status(opts)
	if err != nil {
		return "", err
	}
	return status + "\nIf you do not see your file on the phone:\n- Make sure it is an .html file.\n- Make sure it is inside one of the folders listed above.\n- Open the phone URL and choose \"Paste a file path\" if it does not appear in recent files.\n\nUseful commands:\n  ./codex-artifact-gateway setup --root " + config.DefaultRoot(opts.Home) + "\n  ./codex-artifact-gateway status\n", nil
}

func setupMessage(phoneURL string, cfg config.Config, configPath string) string {
	var builder strings.Builder
	builder.WriteString("Setup complete.\n\n")
	if phoneURL != "" {
		builder.WriteString("Open this on your iPhone:\n")
		builder.WriteString(phoneURL + "\n\n")
	}
	builder.WriteString("Gateway can open HTML files from these folders and their subfolders:\n")
	builder.WriteString(formatRoots(cfg.Roots))
	builder.WriteString("\n")
	builder.WriteString("Gateway was installed from:\n")
	builder.WriteString(cfg.BinaryPath + "\n")
	if warning := unstableInstallWarning(cfg.BinaryPath); warning != "" {
		builder.WriteString("\n" + warning + "\n")
	}
	builder.WriteString("\nDo not move or delete the install folder unless you run setup again.\n\n")
	builder.WriteString("Useful commands:\n")
	builder.WriteString("  ./codex-artifact-gateway status\n")
	builder.WriteString("  ./codex-artifact-gateway doctor\n")
	builder.WriteString("  ./codex-artifact-gateway stop\n\n")
	builder.WriteString("Config file:\n")
	builder.WriteString(configPath + "\n\n")
	builder.WriteString("Feedback is saved here:\n")
	builder.WriteString(cfg.FeedbackDir + "\n")
	return builder.String()
}

func baseURL(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	return strings.TrimRight(strings.TrimSpace(raw), "/") + "/"
}

func formatRoots(roots []string) string {
	if len(roots) == 0 {
		return "- none configured\n"
	}
	var builder strings.Builder
	for _, root := range roots {
		builder.WriteString("- " + root + "\n")
	}
	return builder.String()
}

func firstRootOrDefault(roots []string, home string) string {
	if len(roots) > 0 {
		return roots[0]
	}
	return config.DefaultRoot(home)
}

func unstableInstallWarning(binaryPath string) string {
	if binaryPath == "" {
		return ""
	}
	clean := filepath.Clean(binaryPath)
	parts := strings.Split(clean, string(os.PathSeparator))
	underNamedDir := false
	for _, part := range parts {
		if part == "Downloads" || part == ".Trash" || part == "Trash" {
			underNamedDir = true
			break
		}
	}
	if !strings.HasPrefix(clean, "/tmp/") && !strings.HasPrefix(clean, "/private/tmp/") && !underNamedDir {
		return ""
	}
	return "Warning: Gateway is installed from a folder that may be moved or deleted.\nIf this file moves, the background service will stop working.\nInstall from a stable folder such as ~/Developer/codex-artifact-gateway."
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

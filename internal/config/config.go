package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	AppName                 = "codex-artifact-gateway"
	DefaultAddr             = "127.0.0.1:8767"
	DefaultLaunchAgentLabel = "com.jdfetterly.codex-artifact-gateway"
)

type Config struct {
	Roots            []string `json:"roots"`
	FeedbackDir      string   `json:"feedback_dir"`
	Addr             string   `json:"addr"`
	BinaryPath       string   `json:"binary_path"`
	TailscaleCLIPath string   `json:"tailscale_cli_path"`
	ManageTailscale  bool     `json:"manage_tailscale"`
	LaunchAgentLabel string   `json:"launch_agent_label"`
}

func Default(home string, binaryPath string) Config {
	return Config{
		Roots:            []string{DefaultRoot(home)},
		FeedbackDir:      filepath.Join(home, "Documents", "Codex", "codex-artifact-gateway-feedback"),
		Addr:             DefaultAddr,
		BinaryPath:       binaryPath,
		ManageTailscale:  true,
		LaunchAgentLabel: DefaultLaunchAgentLabel,
	}
}

func DefaultRoot(home string) string {
	return filepath.Join(home, "Documents", "Codex")
}

func ConfigPath(home string) string {
	return filepath.Join(home, "Library", "Application Support", AppName, "config.json")
}

func AppSupportDir(home string) string {
	return filepath.Join(home, "Library", "Application Support", AppName)
}

func LaunchAgentPath(home string) string {
	return filepath.Join(home, "Library", "LaunchAgents", DefaultLaunchAgentLabel+".plist")
}

func LogPath(home string, stream string) string {
	return filepath.Join(home, "Library", "Logs", AppName+"."+stream+".log")
}

func Read(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func Write(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	return os.WriteFile(path, encoded, 0o644)
}

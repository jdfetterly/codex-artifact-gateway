package app

import "github.com/jdfetterly/codex-artifact-gateway/internal/config"

type FakeRunner struct {
	WriteConfigFunc          func(path string, cfg config.Config) error
	ReadConfigFunc           func(path string) (config.Config, error)
	InstallLaunchAgentFunc   func(cfg config.Config, configPath string) error
	StartLaunchAgentFunc     func(label string) error
	StopLaunchAgentFunc      func(label string) error
	LaunchAgentStatusFunc    func(label string) (string, error)
	CheckHealthFunc          func(addr string) error
	StartTailscaleServeFunc  func(path string, addr string) (string, error)
	StopTailscaleServeFunc   func(path string) error
	TailscaleStatusFunc      func(path string) (string, error)
	TailscaleServeStatusFunc func(path string) (string, error)
}

func (f FakeRunner) WriteConfig(path string, cfg config.Config) error {
	return f.WriteConfigFunc(path, cfg)
}

func (f FakeRunner) ReadConfig(path string) (config.Config, error) {
	return f.ReadConfigFunc(path)
}

func (f FakeRunner) InstallLaunchAgent(cfg config.Config, configPath string) error {
	return f.InstallLaunchAgentFunc(cfg, configPath)
}

func (f FakeRunner) StartLaunchAgent(label string) error {
	return f.StartLaunchAgentFunc(label)
}

func (f FakeRunner) StopLaunchAgent(label string) error {
	return f.StopLaunchAgentFunc(label)
}

func (f FakeRunner) LaunchAgentStatus(label string) (string, error) {
	return f.LaunchAgentStatusFunc(label)
}

func (f FakeRunner) CheckHealth(addr string) error {
	return f.CheckHealthFunc(addr)
}

func (f FakeRunner) StartTailscaleServe(path string, addr string) (string, error) {
	return f.StartTailscaleServeFunc(path, addr)
}

func (f FakeRunner) StopTailscaleServe(path string) error {
	return f.StopTailscaleServeFunc(path)
}

func (f FakeRunner) TailscaleStatus(path string) (string, error) {
	return f.TailscaleStatusFunc(path)
}

func (f FakeRunner) TailscaleServeStatus(path string) (string, error) {
	return f.TailscaleServeStatusFunc(path)
}

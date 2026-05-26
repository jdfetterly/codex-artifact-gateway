package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/jdfetterly/codex-artifact-gateway/internal/app"
	"github.com/jdfetterly/codex-artifact-gateway/internal/config"
	"github.com/jdfetterly/codex-artifact-gateway/internal/gateway"
	"github.com/jdfetterly/codex-artifact-gateway/internal/server"
	"github.com/jdfetterly/codex-artifact-gateway/internal/tailscale"
)

type rootFlags []string

func (r *rootFlags) String() string {
	return fmt.Sprint([]string(*r))
}

func (r *rootFlags) Set(value string) error {
	*r = append(*r, value)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	var err error
	var output string
	switch os.Args[1] {
	case "serve":
		err = serve(os.Args[2:])
	case "setup":
		output, err = setup(os.Args[2:])
	case "start":
		output, err = start()
	case "stop":
		output, err = stop()
	case "status":
		output, err = status()
	case "doctor":
		output, err = doctor()
	default:
		usage()
		os.Exit(2)
	}
	if output != "" {
		fmt.Print(output)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func serve(args []string) error {
	var roots rootFlags
	defaultFeedbackDir, err := gateway.DefaultFeedbackDir()
	if err != nil {
		return err
	}
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	flags.Var(&roots, "root", "allowlisted artifact root; repeat for multiple roots")
	addr := flags.String("addr", "127.0.0.1:8767", "listen address")
	feedbackDir := flags.String("feedback-dir", defaultFeedbackDir, "feedback log directory")
	configPath := flags.String("config", "", "path to saved gateway config")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *configPath != "" {
		cfg, err := config.Read(*configPath)
		if err != nil {
			return err
		}
		roots = cfg.Roots
		*addr = cfg.Addr
		*feedbackDir = cfg.FeedbackDir
	}
	if len(roots) == 0 {
		return fmt.Errorf("no roots configured\n\nRun with at least one explicit root:\n  codex-artifact-gateway serve --root /path/to/codex-artifacts")
	}
	policy, err := gateway.NewPolicy(roots)
	if err != nil {
		return err
	}
	fmt.Print(server.StartupMessage(*addr, roots, *feedbackDir))
	return server.Serve(server.Config{Policy: policy, FeedbackDir: *feedbackDir}, *addr)
}

func setup(args []string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	var roots rootFlags
	flags := flag.NewFlagSet("setup", flag.ContinueOnError)
	flags.Var(&roots, "root", "allowlisted artifact root; repeat for multiple roots")
	tailscaleCLI := flags.String("tailscale-cli", "", "path to Tailscale CLI")
	if err := flags.Parse(args); err != nil {
		return "", err
	}
	root := config.DefaultRoot(home)
	if len(roots) > 0 {
		root = roots[0]
	}
	binaryPath, err := app.StableBinaryPath(home)
	if err != nil {
		return "", err
	}
	cliPath := *tailscaleCLI
	if cliPath == "" {
		cliPath, err = tailscale.DetectCLI(exec.LookPath, tailscale.AppBundleCLI, nil)
		if err != nil {
			return "", err
		}
	}
	return app.Setup(app.SetupOptions{
		Home:         home,
		Root:         root,
		BinaryPath:   binaryPath,
		TailscaleCLI: cliPath,
	})
}

func start() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return app.Start(app.StartOptions{Home: home})
}

func stop() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return app.Stop(app.StopOptions{Home: home})
}

func status() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return app.Status(app.StatusOptions{Home: home})
}

func doctor() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return app.Doctor(app.StatusOptions{Home: home})
}

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  codex-artifact-gateway setup [--root /path/to/artifacts]")
	fmt.Fprintln(os.Stderr, "  codex-artifact-gateway start|stop|status|doctor")
	fmt.Fprintln(os.Stderr, "  codex-artifact-gateway serve --root /path/to/artifacts [--root /another/root] [--addr 127.0.0.1:8767]")
}

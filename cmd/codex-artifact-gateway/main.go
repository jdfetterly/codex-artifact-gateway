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

type serveOptions struct {
	Roots       []string
	Addr        string
	FeedbackDir string
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
	opts, err := parseServeOptions(args, gateway.DefaultFeedbackDir, config.Read)
	if err != nil {
		return err
	}
	if len(opts.Roots) == 0 {
		return fmt.Errorf("no roots configured\n\nRun with at least one explicit root:\n  codex-artifact-gateway serve --root /path/to/codex-artifacts")
	}
	if err := server.ValidateListenAddr(opts.Addr); err != nil {
		return err
	}
	policy, err := gateway.NewPolicy(opts.Roots)
	if err != nil {
		return err
	}
	fmt.Print(server.StartupMessage(opts.Addr, opts.Roots, opts.FeedbackDir))
	return server.Serve(server.Config{Policy: policy, FeedbackDir: opts.FeedbackDir}, opts.Addr)
}

func parseServeOptions(args []string, defaultFeedbackDir func() (string, error), readConfig func(string) (config.Config, error)) (serveOptions, error) {
	var roots rootFlags
	flags := flag.NewFlagSet("serve", flag.ContinueOnError)
	flags.Var(&roots, "root", "allowlisted artifact root; repeat for multiple roots")
	addr := flags.String("addr", config.DefaultAddr, "listen address")
	feedbackDir := flags.String("feedback-dir", "", "feedback log directory")
	configPath := flags.String("config", "", "path to saved gateway config")
	if err := flags.Parse(args); err != nil {
		return serveOptions{}, err
	}
	if *configPath != "" {
		cfg, err := readConfig(*configPath)
		if err != nil {
			return serveOptions{}, err
		}
		return serveOptions{
			Roots:       cfg.Roots,
			Addr:        cfg.Addr,
			FeedbackDir: cfg.FeedbackDir,
		}, nil
	}
	dir := *feedbackDir
	if dir == "" {
		var err error
		dir, err = defaultFeedbackDir()
		if err != nil {
			return serveOptions{}, err
		}
	}
	return serveOptions{
		Roots:       []string(roots),
		Addr:        *addr,
		FeedbackDir: dir,
	}, nil
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
	setupRoots := []string{config.DefaultRoot(home)}
	if len(roots) > 0 {
		setupRoots = roots
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
		Roots:        setupRoots,
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
	fmt.Fprintln(os.Stderr, "  codex-artifact-gateway setup [--root /path/to/artifacts] [--root /another/root]")
	fmt.Fprintln(os.Stderr, "  codex-artifact-gateway start|stop|status|doctor")
	fmt.Fprintln(os.Stderr, "  codex-artifact-gateway serve --root /path/to/artifacts [--root /another/root] [--addr 127.0.0.1:8767]")
}

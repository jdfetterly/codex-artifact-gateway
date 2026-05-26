package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jdfetterly/codex-artifact-gateway/internal/gateway"
	"github.com/jdfetterly/codex-artifact-gateway/internal/server"
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
	if len(os.Args) < 2 || os.Args[1] != "serve" {
		usage()
		os.Exit(2)
	}
	if err := serve(os.Args[2:]); err != nil {
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
	if err := flags.Parse(args); err != nil {
		return err
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

func usage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  codex-artifact-gateway serve --root /path/to/artifacts [--root /another/root] [--addr 127.0.0.1:8767]")
}

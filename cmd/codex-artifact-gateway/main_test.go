package main

import (
	"errors"
	"testing"

	"github.com/jdfetterly/codex-artifact-gateway/internal/config"
)

func TestServeOptionsFromConfigDoesNotNeedDefaultFeedbackDir(t *testing.T) {
	defaultFeedbackDir := func() (string, error) {
		return "", errors.New("HOME is not defined")
	}
	readConfig := func(path string) (config.Config, error) {
		if path != "/tmp/config.json" {
			t.Fatalf("config path = %q", path)
		}
		return config.Config{
			Roots:       []string{"/Users/example"},
			Addr:        "127.0.0.1:8767",
			FeedbackDir: "/Users/example/feedback",
		}, nil
	}

	opts, err := parseServeOptions([]string{"--config", "/tmp/config.json"}, defaultFeedbackDir, readConfig)
	if err != nil {
		t.Fatal(err)
	}

	if opts.Addr != "127.0.0.1:8767" {
		t.Fatalf("addr = %q", opts.Addr)
	}
	if opts.FeedbackDir != "/Users/example/feedback" {
		t.Fatalf("feedback dir = %q", opts.FeedbackDir)
	}
	if len(opts.Roots) != 1 || opts.Roots[0] != "/Users/example" {
		t.Fatalf("roots = %#v", opts.Roots)
	}
}

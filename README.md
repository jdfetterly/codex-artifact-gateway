# Codex Artifact Gateway

Codex Artifact Gateway is an unofficial local-first gateway for opening interactive Codex-generated HTML artifacts from an iPhone while the files stay on a Mac.

The first release is intentionally narrow: macOS host, iPhone client, Tailscale-only access, Codex HTML artifacts, preserved page interactivity, local feedback capture, and a Go single-binary implementation.

## Why This Exists

Codex workflows often produce local HTML review pages. Those pages can include buttons, filters, forms, and feedback controls, but a `file:///Users/...` link is not usable from an iPhone. This project provides a private resolver URL that can be opened from mobile over Tailscale.

## Current Status

V0.1 local implementation is available. The gateway can run as a persistent macOS LaunchAgent and expose a private Tailscale URL for iPhone review.

## Quick Start

Build the binary once from this repository:

```bash
go build ./cmd/codex-artifact-gateway
```

Install and start the gateway:

```bash
./codex-artifact-gateway setup \
  --root /Users/jdfetterly/Documents/Codex
```

The setup command:

- writes `~/Library/Application Support/codex-artifact-gateway/config.json`
- installs a user LaunchAgent
- starts the local gateway on `127.0.0.1:8767`
- configures Tailscale Serve
- prints the iPhone `/recent` URL

Check or stop it:

```bash
./codex-artifact-gateway status
./codex-artifact-gateway doctor
./codex-artifact-gateway stop
```

The gateway binds to `127.0.0.1:8767` by default.

Open locally:

```text
http://127.0.0.1:8767/recent
```

Open a specific file:

```text
http://127.0.0.1:8767/open?path=file:///Users/example/report.html
```

Use the paste resolver for existing local file links:

```text
http://127.0.0.1:8767/resolve
```

## Tailscale Serve

`setup` and `start` manage Tailscale Serve automatically. The CLI looks for `tailscale` on `PATH`, then falls back to `/Applications/Tailscale.app/Contents/MacOS/Tailscale`.

`stop` disables the managed Tailscale Serve proxy so the tailnet URL does not point at a stale local port.

## Feedback Logs

By default, feedback is appended under the current user's app configuration directory:

```text
~/Library/Application Support/codex-artifact-gateway/feedback/YYYY-MM-DD-feedback.jsonl
```

Override this with:

```bash
codex-artifact-gateway serve \
  --root /path/to/codex-artifacts \
  --feedback-dir /path/to/feedback
```

## Development

Manual development server:

```bash
go run ./cmd/codex-artifact-gateway serve \
  --root /path/to/codex-artifacts
```

Serve can also read the saved setup config:

```bash
./codex-artifact-gateway serve \
  --config "$HOME/Library/Application Support/codex-artifact-gateway/config.json"
```

Run the test suite:

```bash
go test ./...
```

## Core Documents

- [SCOPE_AND_OUTCOMES.md](SCOPE_AND_OUTCOMES.md): v0.1 boundary, success criteria, and deferred ideas.
- [AGENTS.md](AGENTS.md): working rules for future agents and contributors.
- [SECURITY.md](SECURITY.md): supported security boundary.

## Initial Non-Goals

- Public internet hosting.
- Generic file server behavior.
- Windows, Linux, or Android support.
- Non-Codex workflows as first-class use cases.
- Treating feedback as approval to mutate files or trigger external actions.

## License

MIT.

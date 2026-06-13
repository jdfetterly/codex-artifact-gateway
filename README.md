# Codex Artifact Gateway

Open local Codex-generated HTML artifacts from an iPhone or iPad over your private Tailscale network without publishing them to the public internet.

Codex Artifact Gateway is an unofficial, local-first macOS utility for reviewing local HTML artifacts on trusted mobile devices. It runs on your Mac, serves only explicitly allowlisted artifact roots, and is designed for localhost plus Tailscale Serve.

This is not an OpenAI product and does not imply OpenAI ownership or endorsement.

## What It Does

- Opens local Codex-generated HTML files from trusted mobile browsers.
- Resolves local paths and `file:///Users/...` URLs into private review links.
- Preserves client-side HTML behavior such as JavaScript, filters, tabs, charts, forms, and relative assets.
- Adds an optional generic feedback drawer at response time without modifying the original HTML file.
- Stores feedback locally as append-only JSONL on the Mac.
- Installs as a user-level macOS LaunchAgent and configures Tailscale Serve.
- Rejects requests outside configured roots and blocks private path components such as `.codex`, `.ssh`, and top-level `Library`.

## Why This Exists

Codex workflows often produce useful local HTML review pages, but a `file:///Users/...` link from a Mac cannot be opened directly from an iPhone. Uploading those artifacts to a public host just to review them is usually the wrong tradeoff.

Gateway gives you a private resolver URL for trusted devices on your tailnet while keeping the files on your Mac.

## How It Works

1. Build and run Gateway on your Mac.
2. Allowlist the local directories that contain Codex-generated HTML artifacts.
3. Gateway serves those artifacts on `127.0.0.1:8767`.
4. Tailscale Serve exposes that local port only to trusted devices on your tailnet.
5. Open `/recent`, `/resolve`, or a generated `/open?...` link from your iPhone or iPad.

Backend-dependent pages should stay with the app that owns the backend. Gateway preserves client-side page behavior and captures local feedback; it does not proxy unrelated application APIs or make app-specific action buttons work from a static artifact URL.

## Support Matrix

| Area | Supported in v0.1 |
| --- | --- |
| Host | macOS user session |
| Client | iPhone/iPad browser on the same tailnet |
| Network exposure | Localhost plus Tailscale Serve |
| Artifact type | Codex-generated local HTML and relative assets |
| Interactivity | Client-side JavaScript already present in the page |
| Feedback | Local append-only JSONL |
| Public internet | Not supported |
| Backend API proxying | Not supported |

## Security Model

Gateway is private by default:

- The local server binds to `127.0.0.1:8767`.
- Mobile access is intended to go through Tailscale Serve to trusted tailnet devices.
- Artifact roots must be explicitly allowlisted.
- Private path components are rejected even when a broad root is configured.
- Feedback is treated as untrusted user input.
- The gateway should run as your logged-in macOS user, not with `sudo`.

Do not expose Gateway with Tailscale Funnel, public tunnels, reverse proxies, public interfaces such as `0.0.0.0`, or generic file-hosting infrastructure.

Anyone who can access the tailnet URL should be treated as able to view supported files under the configured allowlisted roots. Tailnet access is not per-file authorization.

## What It Is Not

Gateway is intentionally narrow. It is not:

- A public file host.
- A generic whole-home-directory file server.
- A multi-user SaaS product.
- A reverse proxy for arbitrary application APIs.
- A privileged daemon.
- An approval system for publishing, sending, merging, or mutating external systems.

If an HTML page depends on a separate backend API to save state or perform app-specific actions, serve that page through the app that owns the backend. Gateway is for private artifact viewing and local feedback capture, not for making unrelated application backends available.

## Quick Start

Prerequisites:

- macOS.
- Go 1.22 or newer.
- Tailscale installed and signed in on the Mac.
- An iPhone, iPad, or other trusted device on the same tailnet.

Build the binary from this repository:

```bash
go build ./cmd/codex-artifact-gateway
```

Install and start the gateway as your logged-in macOS user:

```bash
./codex-artifact-gateway setup \
  --root "$HOME/Documents/Codex"
```

Prefer narrow artifact roots. Repeat `--root` for every local artifact tree the phone should be able to open:

```bash
./codex-artifact-gateway setup \
  --root "$HOME/Documents/Codex" \
  --root "$HOME/Reference"
```

Avoid broad roots such as `$HOME`. Anyone with access to the tailnet URL should be treated as able to view supported files under the configured roots, so each root should map to an artifact tree you are comfortable reviewing from trusted mobile devices.

The setup command:

- writes `~/Library/Application Support/codex-artifact-gateway/config.json`
- saves the configured allowlisted roots
- installs a user LaunchAgent
- starts the local gateway on `127.0.0.1:8767`
- configures Tailscale Serve
- prints the mobile `/recent` URL

Check or stop the gateway:

```bash
./codex-artifact-gateway status
./codex-artifact-gateway doctor
./codex-artifact-gateway stop
```

`stop` also disables the managed Tailscale Serve proxy so the tailnet URL does not point at a stale local port.

First-run check:

1. Run `./codex-artifact-gateway doctor`.
2. Open `http://127.0.0.1:8767/recent` on the Mac.
3. Open the printed Tailscale `/recent` URL from the phone.

## Common Workflows

Open recent artifacts locally:

```text
http://127.0.0.1:8767/recent
```

Open a specific local HTML file:

```text
http://127.0.0.1:8767/open?path=file:///Users/example/report.html
```

Paste an existing local path or `file:///` URL:

```text
http://127.0.0.1:8767/resolve
```

Serve manually for development:

```bash
go run ./cmd/codex-artifact-gateway serve \
  --root /path/to/codex-artifacts
```

Serve from a saved setup config:

```bash
./codex-artifact-gateway serve \
  --config "$HOME/Library/Application Support/codex-artifact-gateway/config.json"
```

## Feedback Logs

By default, feedback is appended under the configured feedback directory:

```text
~/Documents/Codex/codex-artifact-gateway-feedback/YYYY-MM-DD-feedback.jsonl
```

Override this with:

```bash
./codex-artifact-gateway serve \
  --root /path/to/codex-artifacts \
  --feedback-dir /path/to/feedback
```

Feedback may contain user-controlled text, URLs, browser metadata, and artifact references. Do not put secrets, credentials, private paths, or instructions for external actions in feedback, and escape or sanitize feedback before displaying it in another tool.

## Project Status

The current implementation supports the first public milestone: a macOS host, iPhone/iPad browser review over Tailscale, Codex-generated HTML artifacts, local feedback capture, explicit allowlisted roots, and a Go single-binary build from source.

Release packaging is intentionally minimal for the initial launch. Future packaging may include Homebrew after the source-build path has been validated from a clean checkout.

## Development

Run tests:

```bash
go test ./...
```

Check static issues:

```bash
go vet ./...
```

Core files:

- `cmd/codex-artifact-gateway/main.go`: CLI entrypoint.
- `internal/server`: HTTP routes and artifact serving.
- `internal/gateway`: path policy, feedback storage, and HTML injection.
- `internal/app`: setup, status, LaunchAgent, and Tailscale orchestration.
- `internal/config`: local config paths and defaults.
- `internal/launchd`: user LaunchAgent plist generation.
- `internal/tailscale`: Tailscale CLI integration.

## Documentation

- [SCOPE_AND_OUTCOMES.md](SCOPE_AND_OUTCOMES.md): v0.1 boundary, success criteria, and deferred ideas.
- [SECURITY.md](SECURITY.md): supported security boundary and reporting guidance.
- [AGENTS.md](AGENTS.md): working rules for future agents and contributors.

## License

Copyright 2026 JD Fetterly.

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE).

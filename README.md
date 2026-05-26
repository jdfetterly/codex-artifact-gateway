# Codex Artifact Gateway

Codex Artifact Gateway is an unofficial local-first gateway for opening interactive Codex-generated HTML artifacts from an iPhone while the files stay on a Mac.

The first release is intentionally narrow: macOS host, iPhone client, Tailscale-only access, Codex HTML artifacts, preserved page interactivity, local feedback capture, and a Go single-binary implementation.

## Why This Exists

Codex workflows often produce local HTML review pages. Those pages can include buttons, filters, forms, and feedback controls, but a `file:///Users/...` link is not usable from an iPhone. This project provides a private resolver URL that can be opened from mobile over Tailscale.

## Current Status

Planning and project setup for a Go implementation. No production implementation should be assumed yet.

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

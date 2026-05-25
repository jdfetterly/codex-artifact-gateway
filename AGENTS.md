# AGENTS.md

## Project Contract

This repository is for Codex Mobile HTML Gateway: an unofficial macOS-to-iPhone, Tailscale-only gateway for opening and interacting with Codex-generated local HTML artifacts.

Use `SCOPE_AND_OUTCOMES.md` as the source of truth for product scope. Do not expand this project into generic file hosting, public tunnel infrastructure, multi-user SaaS, non-Codex workflows, or external action dispatch without explicit JD approval.

## Current Priority

The immediate goal is a narrow v0.1:

- macOS host.
- iPhone client.
- Tailscale-only access.
- Codex-generated HTML artifacts.
- Interactive HTML viewing.
- Local feedback capture.
- Allowlisted roots.

Assume nothing is already set up unless it exists in this repository.

## Safety Rules

- Keep the gateway private by default.
- Treat Tailscale Serve as the intended exposure layer unless the scope doc changes.
- Never add public internet exposure by default.
- Never serve arbitrary home-directory files.
- Never execute shell commands from HTTP requests.
- Never treat feedback submissions as approvals to mutate files, publish content, send messages, merge code, or perform external actions.
- Do not weaken path allowlist checks.
- Do not follow symlinks or path traversal unless a reviewed design explicitly permits it.

## Implementation Guidance

- Prefer small, testable units for path resolution, file serving, HTML injection, link resolving, and feedback logging.
- Preserve original HTML files; any feedback UI injection should happen at response time.
- Store feedback in append-only structured files.
- Keep dependencies minimal.
- If the public technology direction changes, update `SCOPE_AND_OUTCOMES.md` in the same change.
- If CLI behavior changes, update `README.md` in the same change.

## Verification

Before claiming a change is complete:

- Run the relevant tests.
- For broad changes, run the full test suite.
- For UI or injected-feedback changes, verify the page in a mobile viewport when feasible.
- For path policy changes, include tests for allowed roots and rejected outside-root paths.

## Documentation Style

- Keep docs direct and operator-facing.
- Prefer Markdown for plans, scope, implementation notes, and agent handoffs.
- Create HTML only when visual/browser review adds real value.
- Keep examples generic enough for public open-source use.
- Avoid embedding private local paths except where explicitly marked as local examples.

## Git Hygiene

- Do not remove or rewrite user-created files unless JD asks.
- Ignore generated caches, screenshots, local feedback logs, and local runtime state unless they are intentional fixtures.
- Keep commits focused around one product or implementation decision.

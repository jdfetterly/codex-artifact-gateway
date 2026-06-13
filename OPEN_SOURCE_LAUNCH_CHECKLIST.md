# Open Source Launch Checklist

Use this checklist to move Codex Artifact Gateway from local/private state toward a public-ready open source project without expanding the product scope.

## Launch Gates

- [x] Repository owner confirmed Apache-2.0 is the intended license.
- [x] `go test ./...` passes.
- [x] `go vet ./...` passes.
- [x] `git diff --check` passes.
- [ ] `git status --short --branch` is clean except for intentional launch docs.
- [ ] `git ls-remote --heads origin main` shows the expected remote branch before assuming GitHub has the current code.
- [x] No private feedback logs, local configs, screenshots, credentials, tokens, or private runtime state are tracked.
- [x] Clean-clone serve-only RC smoke passes without running `setup`, LaunchAgent, Tailscale Serve, `git push`, `git tag`, or visibility changes.
- [ ] Repository owner explicitly approves each externally visible step: commit, push, visibility change, release tag, and public announcement.

## License

- [x] Confirm `LICENSE` matches Apache-2.0.
- [x] Confirm copyright holder and year are correct: `Copyright 2026 JD Fetterly`.
- [x] Keep the README license label consistent with `LICENSE`.
- [ ] Consider SPDX headers later if the project needs stricter packaging hygiene; do not block v0.1 on this.

## Security

- [x] Run `go test ./...`.
- [x] Run `go vet ./...`.
- [x] Re-check path traversal, outside-root, symlink, and private-path protections.
- [x] Confirm broad roots still reject hidden/private paths such as `.codex`, `.ssh`, top-level `Library`, credentials, tokens, and private runtime state.
- [x] Review `SECURITY.md` for a public reporting path and clear local/tailnet-only boundary.
- [x] Confirm README tells users to run as the logged-in macOS user, not with `sudo` or a privileged service.
- [x] Confirm runtime listen address is loopback-only; reject `0.0.0.0`, LAN addresses, and missing-host binds.
- [x] Confirm HTTP handlers enforce request body limits and server timeouts.
- [x] Confirm `/health` does not expose private local paths.
- [x] Run `gosec ./...` if available; if not installed, record that rather than adding tooling just for launch. `gosec` was not installed for this pass.
- [x] Confirm dependency surface with `go list -m all`; v0.1 should stay standard-library-only unless a dependency removes real complexity.
- [x] Do not add Tailscale Funnel, public tunnel, reverse proxy, `0.0.0.0`, arbitrary file serving, command execution, or feedback-as-approval behavior.

## RC Validation Evidence

- [x] Commit tested: `a33641f`.
- [x] Clean clone path: `/private/tmp/codex-artifact-gateway-rc-20260613100237`.
- [x] Smoke mode: `./codex-artifact-gateway serve --addr 127.0.0.1:8768 --root /private/tmp/codex-artifact-gateway-fixture-20260613100237/artifacts --feedback-dir /private/tmp/codex-artifact-gateway-fixture-20260613100237/feedback`.
- [x] Verified local endpoints: `/health`, `/recent`, `/open`, `/view`, `/resolve`, `/api/feedback`.
- [x] Verified rejection cases: outside-root path, `.codex` private path, unsupported `.py`, feedback `.jsonl`, oversized feedback body, and `0.0.0.0` bind.
- [x] Verified listener: `127.0.0.1:8768` only.
- [x] Smoke server stopped after validation.

## Repository Hygiene

- [x] Confirm the built `codex-artifact-gateway` binary is ignored and not tracked.
- [x] Confirm `.gitignore` covers local feedback JSONL, `codex-artifact-gateway-feedback/`, runtime state, screenshots, coverage, editor files, logs, and build output.
- [x] Run `git ls-files` and scan for private paths or generated artifacts.
- [x] Run a lightweight tracked-content scan for obvious secrets, tokens, private keys, Tailscale URLs, tailnet hostnames, and private local paths.
- [x] Keep examples generic enough for public use; mark any local path examples clearly.

## Documentation

- [x] README accurately describes the narrow v0.1 scope: macOS host, iPhone client, Tailscale-only access, Codex HTML artifacts, local feedback capture, allowlisted roots, and Go binary.
- [ ] Quick start works from a clean clone.
- [x] Source-build plus serve-only validation works from a clean clone.
- [x] Tailscale-only positioning is clear; do not add public tunnel or generic file-hosting instructions.
- [x] Feedback storage behavior is clear: local append-only JSONL, user input only, not approval for external actions.
- [x] Public examples use generic paths or clearly marked examples, not private local defaults.
- [x] Scope boundaries, limitations, troubleshooting, and security model are easy to find.
- [x] README and scope doc both state that the project is unofficial and not endorsed by OpenAI.
- [x] Public docs do not mention private workflow context, project-specific backend actions, or private Tailscale hostnames.

## GitHub Settings

- [ ] Set a concise repo description.
- [ ] Add topics such as `codex`, `tailscale`, `macos`, `iphone`, `local-first`, `html-artifacts`, and `go`.
- [ ] Decide repository visibility only after launch gates pass.
- [ ] Confirm default branch is `main`.
- [ ] Decide whether Issues are open for v0.1; Discussions can wait.
- [ ] Enable security advisories, Dependabot alerts, or code scanning if available.
- [ ] Add branch protection later if the project starts accepting external contributions.
- [ ] Keep the repository private until the tracked-content scan and final docs review pass.

## Release Roadmap

- [x] Keep `go build ./cmd/codex-artifact-gateway` as the initial install path.
- [ ] Add Homebrew installation after the first public-ready release, either through a tap or formula.
- [ ] Plan a simple version tag such as `v0.1.0`.
- [ ] Draft short release notes covering purpose, supported setup, limitations, and security boundary.
- [x] Avoid promising Homebrew until packaging exists and has been tested.

## Publish Steps

Run these only after the launch gates pass and the repository owner explicitly approves publishing:

```bash
git push -u origin main
git ls-remote --heads origin main
```

After the remote contents are reviewed and the repository owner approves the tag, optionally create the first release tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Do not make the repository public until the repository owner separately approves the visibility change after reviewing the pushed private remote.

# Security Policy

Codex Artifact Gateway is intended for private local and tailnet-only use.

## Supported Boundary

- Run the gateway on macOS.
- Expose it only through Tailscale or localhost during development.
- Configure only artifact roots that the user is comfortable browsing from trusted devices.
- Run it as the logged-in macOS user, not with `sudo` or as a privileged service.
- Treat feedback as untrusted user input, not as an approval or command.

## Exposure Model

The gateway is intended to bind locally and be reached from trusted devices through Tailscale Serve. Do not expose it through Tailscale Funnel, public tunnels, reverse proxies, public interfaces, or generic file-hosting infrastructure.

Anyone with access to the tailnet URL should be treated as able to view supported files under the configured allowlisted roots. Tailnet access is not per-file authorization.

## Not Supported

- Public internet exposure.
- Serving arbitrary home-directory files.
- Running as a privileged user.
- Executing commands from HTTP requests.
- Proxying arbitrary application APIs or app-specific backend actions.
- Treating feedback submissions as authenticated approvals.

## Feedback Logs

Feedback logs may contain user-controlled text, URLs, browser metadata, and artifact references. Do not store secrets, credentials, private keys, tokens, private local paths, or external-action instructions in feedback. Escape or sanitize feedback before displaying it in another system.

## Reporting Issues

Use GitHub Security Advisories for vulnerabilities once this project is public and private vulnerability reporting is enabled. If private reporting is not available yet, open a public issue with only a non-sensitive summary and ask for a private reporting path. Public GitHub issues are fine for non-sensitive bugs only.

Do not paste secrets, tokens, private keys, config files, feedback JSONL, Tailscale URLs, tailnet hostnames, `.ssh` paths, `.codex` paths, or screenshots containing private local paths into public issues. Redact local absolute paths and private artifact content before sharing diagnostics.

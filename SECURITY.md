# Security Policy

Codex Mobile HTML Gateway is intended for private local and tailnet-only use.

## Supported Boundary

- Run the gateway on macOS.
- Expose it only through Tailscale or localhost during development.
- Configure only artifact roots that the user is comfortable browsing from trusted devices.
- Treat feedback as user input, not as an approval or command.

## Not Supported

- Public internet exposure.
- Serving arbitrary home-directory files.
- Running as a privileged user.
- Executing commands from HTTP requests.
- Treating feedback submissions as authenticated approvals.

## Reporting Issues

Open a GitHub issue once this project has a public repository. Until then, report issues directly to the repository owner.

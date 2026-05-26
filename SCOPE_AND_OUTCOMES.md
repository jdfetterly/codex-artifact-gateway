# Scope And Outcomes

## Project Purpose

Codex Artifact Gateway lets a macOS-hosted Codex workflow open local HTML artifacts on an iPhone through Tailscale while preserving page interactivity and capturing feedback back on the Mac.

The project should become an open-source tool, but the first version should solve one concrete problem well: reviewing and responding to Codex-generated HTML pages from iPhone without publishing those files to the public internet.

This is an unofficial project and should not imply ownership or endorsement by OpenAI.

## Initial Audience

- Primary user: a Codex desktop user on macOS who reviews outputs from an iPhone.
- Primary client: iPhone browser or Codex-related mobile browsing context over Tailscale.
- Primary artifacts: Codex-generated local HTML files and their local browser assets.

## V0.1 Scope

V0.1 supports:

- macOS host.
- iPhone client.
- Tailscale-only access pattern.
- Codex-generated HTML artifacts.
- Universal resolver links for local paths and `file:///Users/...` URLs.
- Opening links from Codex project chat.
- Opening links from emails that use the gateway URL.
- A paste/resolve path for existing local file links.
- Preserved page interactivity, including JavaScript, buttons, filters, forms, and relative assets.
- A generic feedback surface for pages that do not already have a backend.
- Local feedback capture as structured append-only files.
- Explicit allowlisted artifact roots.

## Out Of Scope For V0.1

V0.1 does not support:

- Public internet exposure.
- Windows or Linux hosts.
- Android clients.
- Generic whole-filesystem serving.
- Non-Codex artifact workflows as first-class use cases.
- Cloud sync, hosted accounts, teams, or multi-tenant access control.
- Browser extensions.
- Automatic rewriting of every email or chat link.
- Treating mobile feedback as approval to mutate source files, send messages, merge code, publish content, or trigger external actions.
- Replacing artifact-specific feedback endpoints when those already exist.

## Important Design Options To Decide

### Implementation Language

Decision: Go. The public project should target a single-binary CLI/server, Homebrew-ready distribution, and low-friction open-source install.

### Exposure Model

Options:

- Tailscale Serve in front of a localhost-bound gateway.
- Gateway binds directly to the Tailscale interface.
- LAN-only browser access.
- Public tunnel.

Recommendation for v0.1: localhost-bound gateway plus Tailscale Serve. Do not support public tunnels.

### Link Model

Options:

- Universal resolver link that works from laptop or iPhone.
- Separate mobile-only links.
- Manual paste box only.

Recommendation for v0.1: universal resolver links, plus a paste/resolve page for existing `file:///` links.

### Feedback Model

Options:

- Injected generic feedback drawer.
- Artifact-specific JavaScript SDK.
- Native app callbacks.
- No generic feedback; only preserve existing page behavior.

Recommendation for v0.1: injected generic feedback drawer plus preserved existing behavior. Defer SDKs and adapters.

## Design Principles

- Mobile review should not require publishing local artifacts.
- The gateway should not need to know whether the user is on laptop or iPhone before a link is generated.
- Resolver URLs should be stable enough to use in Codex chat and email.
- File access must be explicit and allowlisted.
- Feedback writes should be local, inspectable, append-only, and easy for Codex or a human to review later.
- The original HTML artifact should not be modified just to inject feedback UI.
- The open-source project should be easy to explain in one sentence and difficult to misuse as a general-purpose public file server.

## Success Outcomes

The first useful release is successful when:

1. A Codex-generated local HTML file can be opened from iPhone over Tailscale.
2. The same resolver link can be used from Codex chat or email.
3. The rendered page preserves normal HTML behavior and relative assets.
4. Existing buttons and page controls remain usable from iPhone.
5. A user can leave feedback from iPhone.
6. Feedback is stored locally with enough context to identify the artifact, URL, timestamp, browser, and comment.
7. Requests outside configured roots are rejected.
8. The repository can be understood, run, and tested by another developer.

## Public Project Readiness Checklist

Before publishing publicly:

- Use `codex-artifact-gateway` as the project and CLI name unless renamed before release.
- Keep the implementation in Go.
- Confirm install story.
- Add working implementation.
- Add tests for path allowlists, feedback writes, and HTML serving.
- Add clear screenshots or demo artifacts without private data.
- Review README for open-source clarity.
- Confirm license and security policy.
- Remove private local paths except marked examples.

## Deferred Ideas

- iOS Shortcut for rewriting copied `file:///Users/...` links into gateway URLs.
- Project-specific feedback adapters.
- Browser extension or Codex-app integration for automatic link rewriting.
- QR-code launch page.
- Desktop redirect behavior for laptop users.
- Tailscale identity header awareness.
- Optional local passcode in addition to tailnet access.
- Support for non-Codex artifact workflows.

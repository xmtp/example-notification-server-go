# Design Spec: Add Zed Editor Configuration

**Issue:** https://github.com/xmtp/example-notification-server-go/issues/71

## Requirements

- **REQ-1:** When the project is opened in Zed, the Go language server (gopls) shall be enabled with standard Go formatting (tabs, format-on-save).
- **REQ-2:** When the project is opened in Zed, golangci-lint diagnostics shall be provided using the project's existing config at `dev/.golangci.yaml`.
- **REQ-3:** The configuration shall be minimal — only settings required for good Go development results.

## System Design

Add a single file: `.zed/settings.json`

This mirrors the existing `.vscode/settings.json` pattern already in the repo. The config will:
1. Configure Go language settings (hard tabs, format on save via language server)
2. Register both `gopls` and `golangci-lint` as language servers for Go
3. Pass `--config=./dev/.golangci.yaml` to golangci-lint since the config is not in the project root
4. Use golangci-lint v2 CLI flags (matching the project's `version: "2"` config)

## Testing & Validation

- Verify `.zed/settings.json` is valid JSON
- Verify golangci-lint flags match v2 syntax
- Verify config path matches existing `dev/.golangci.yaml`
- Verify consistency with existing `.vscode/settings.json` approach

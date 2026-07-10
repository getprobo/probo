# Changelog

All notable changes to the `@probo/skills` package will be documented in
this file.

## Unreleased

### Added

- `@probo/skills` npm package with multi-agent manifests (`.claude-plugin/`,
  `.codex-plugin/`, `.agents/`) and Probo MCP wiring via `.mcp.json`
  (`PROBO_BASE_URL`, OAuth 2.0 sign-in)
- Skills: `open-source-compliance`, `access-review`, `missing-signatures`
- Commands: `access-review`, `missing-signatures` (Claude Code slash commands
  delegating to shared skills)
- Portable relative `references/` paths in skill bodies (no
  `${CLAUDE_PLUGIN_ROOT}`)
- Resumable session notes for `access-review` and `missing-signatures`
- Repo-root marketplace catalogs for Git installs (`getprobo/probo`):
  `.claude-plugin/marketplace.json` (Claude Code) and
  `.agents/plugins/marketplace.json` (Codex)
- Package-local marketplace catalogs under `packages/skills/`
- `COMPATIBILITY.md`, validation script, and release workflow
  (`release-npm-skills.yaml`, tag `@probo/skills/v*`)

### Changed

- Package published as `@probo/skills` in `packages/skills/` (multi-agent
  scope, not Claude-specific)
- Access-review skill records entry notes only after successful API writes
- Release checksums derived from `npm pack` tarball contents
- Validation enforces `.claude-plugin/marketplace.json` structure and Codex
  marketplace shape; `package.json` files list trimmed to paths that exist

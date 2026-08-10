# Changelog

All notable changes to the `@probo/skills` package will be documented in
this file.

## Unreleased

## [0.3.0] - 2026-08-10

### Added

- Agent Plugins 1.0.0 portable package: `plugin.json` and `mcp.json` at the
  package root, both declaring the canonical
  `https://agent-plugins.org/schemas/1.0.0/…` schema identifiers, so any
  conformant client can load the skills and Probo MCP servers without
  client-specific wiring
- Both hosted Probo instances ship as Streamable HTTP servers: `probo-us`
  (`https://us.probo.com/api/mcp/v1`) and `probo-eu`
  (`https://eu.probo.com/api/mcp/v1`)
- Validation of the Agent Plugins manifest and MCP schemas (closed field sets,
  plugin name constraints, transport variants, literal HTTPS URLs, no
  credential headers), of Agent Skills frontmatter (`name` matching the skill
  directory, `description` and `compatibility` length limits), and of
  `.mcp.json` parity with `mcp.json`

### Fixed

- MCP endpoint path: the Probo MCP API is served at `<instance-root>/api/mcp/v1`.
  The previous `/mcp/v1` path returned the console HTML page instead of the MCP
  endpoint, so no bundled server could connect

### Changed

- **Breaking:** MCP configuration no longer reads `PROBO_BASE_URL`. Agent
  Plugins performs no placeholder expansion in remote MCP URLs, so the hosted
  regional endpoints are declared literally. Self-hosted users add their
  instance in the agent, e.g.
  `claude mcp add --transport http probo https://probo.example.com/api/mcp/v1`
- Skills select a region instead of assuming a single `probo` server, and
  discover it by calling a list tool on each connected server when the region
  is unknown
- Client manifest versions (`.claude-plugin/`, `.codex-plugin/`) track
  `package.json`, which validation now enforces

## [0.2.2] - 2026-08-04

### Fixed

- `access-review` skill: reference doc now points agents to the new `source_name` and `connector_id` entry fields instead of inferring the source tool from account data

## [0.2.1] - 2026-08-03

### Fixed

- `compliance-portal-commitments` skill: reference doc corrected to use the `compliance_portal_id` argument name (previously documented the stale `trust_center_id` name, which no longer matches the current API)

## [0.2.0] - 2026-07-25

### Added

- Skill: `compliance-portal-commitments` — create or update the public
  commitments shown on a Probo compliance portal, grounded strictly in the
  organization's own published Probo policies and written in a factual,
  understated engineering voice (references: `voice.md`, `portal-mechanics.md`)

## [0.1.0] - 2026-07-10

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
  marketplace shape; rejects non-object JSON roots in catalog files

# @probo/skills

Multi-agent compliance skills for open-source GRC workflows. Ships Agent
Skills–compatible instructions and wires agents to the [Probo MCP
API](https://github.com/getprobo/probo/tree/main/pkg/server/api/mcp/v1) via
OAuth 2.0.

**Supported agents:** Claude Code, Codex, OpenCode, Cursor (MCP + skills).

Marketplace catalogs: `.claude-plugin/marketplace.json` at the repo root or
under `packages/skills/` (Claude Code), `.agents/plugins/marketplace.json` at
the repo root or under `packages/skills/` (Codex). See
[COMPATIBILITY.md](./COMPATIBILITY.md).

## Install

### From GitHub or npm

**From GitHub** (repo-root catalog at `.claude-plugin/marketplace.json`):

```bash
claude plugin marketplace add getprobo/probo
# or, from a local clone:
claude plugin marketplace add .
claude plugin install probo@probo
```

**From the package directory** (catalog at
`packages/skills/.claude-plugin/marketplace.json`, resolves `@probo/skills`
from npm):

```bash
claude plugin marketplace add ./packages/skills/.claude-plugin
claude plugin install probo@probo
```

When consuming the published package, the package-level marketplace entry
resolves `@probo/skills` from npm (see
`packages/skills/.claude-plugin/marketplace.json`).

### Configure Probo MCP

Set your Probo instance URL before starting Claude Code:

```bash
export PROBO_BASE_URL="https://your-probo-instance.example.com"
```

The plugin `.mcp.json` connects to `${PROBO_BASE_URL}/mcp/v1`. Probo MCP
authenticates with **OAuth 2.0** — no API token or bearer header is required in
the plugin config. On first use, sign in from Claude Code:

```text
/mcp
```

Or from your shell:

```bash
claude mcp login probo
```

Claude Code discovers Probo's authorization server via
`/.well-known/oauth-protected-resource` and stores tokens securely.

### Local development

```bash
claude --plugin-dir ./packages/skills
```

## What's included

| Component | Location | Purpose |
| --- | --- | --- |
| MCP | `.mcp.json` | Probo API connection |
| Skills | `skills/` | Compliance workflows |
| Commands | `commands/` | `access-review`, `missing-signatures` — semi-auto workflows |

Skills: `/probo:<skill-name>` (e.g. `/probo:open-source-compliance`, `/probo:missing-signatures`).

Commands: `/probo:<command-name>` (e.g. `/probo:access-review`, `/probo:missing-signatures`).

## Adding content

See [`contrib/claude/skills.md`](../../contrib/claude/skills.md).

```bash
npm --workspace @probo/skills run validate
claude --plugin-dir ./packages/skills
```

## Release

Published to npm as `@probo/skills`. See
[`contrib/claude/release/skills.md`](../../contrib/claude/release/skills.md).

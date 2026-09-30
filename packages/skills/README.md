# @probo/skills

Multi-agent compliance skills for open-source GRC workflows. Ships an
[Agent Plugins 1.0.0](https://agent-plugins.org/) package: Agent
Skills–compatible instructions plus portable MCP wiring to the [Probo MCP
API](https://github.com/getprobo/probo/tree/main/pkg/server/api/mcp/v1) over
OAuth 2.0.

**Supported agents:** any Agent Plugins client, plus Claude Code, Codex,
OpenCode, and Cursor through their own plugin or MCP configuration.

The portable package is `plugin.json`, `skills/`, and `mcp.json` at the package
root. Client-specific manifests (`.claude-plugin/`, `.codex-plugin/`,
`.mcp.json`) and marketplace catalogs sit alongside it. See
[COMPATIBILITY.md](./COMPATIBILITY.md).

## Probo MCP servers

Both hosted Probo instances ship in `mcp.json`; connect the one for your
region and sign in with OAuth 2.0.

| Server | Endpoint |
| --- | --- |
| `probo-us` | `https://us.probo.com/api/mcp/v1` |
| `probo-eu` | `https://eu.probo.com/api/mcp/v1` |

Agent Plugins 1.0.0 defines no placeholder expansion for remote MCP URLs, so a
self-hosted instance is added in the agent instead of in the package:

```bash
claude mcp add --transport http probo https://probo.example.com/api/mcp/v1
```

No API token or bearer header belongs in any of these configs. Probo MCP
authenticates with OAuth 2.0, discovered from
`/.well-known/oauth-protected-resource` on the instance root.

## Install

### Agent Plugins client

Point the client at the package root (`packages/skills` in a clone, or the
installed `@probo/skills` directory). It reads `plugin.json`, discovers the
skills under `skills/`, and loads both servers from `mcp.json`.

### Claude Code

**From GitHub** (repo-root catalog at `.claude-plugin/marketplace.json`):

```bash
claude plugin marketplace add getprobo/probo
# or, from a local clone:
claude plugin marketplace add .
claude plugin install probo@probo
claude mcp login probo-us   # or /mcp in session
```

**From the package directory** (catalog at
`packages/skills/.claude-plugin/marketplace.json`, resolves `@probo/skills`
from npm):

```bash
claude plugin marketplace add ./packages/skills/.claude-plugin
claude plugin install probo@probo
```

### Local development

```bash
claude --plugin-dir ./packages/skills
```

## What's included

| Component | Location | Purpose |
| --- | --- | --- |
| Manifest | `plugin.json` | Agent Plugins 1.0.0 plugin identity |
| MCP | `mcp.json` | Portable hosted Probo servers |
| Skills | `skills/` | Compliance workflows |
| Commands | `commands/` | `access-review`, `missing-signatures` — semi-auto workflows (Claude Code) |

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

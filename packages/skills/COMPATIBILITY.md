<!--
Copyright (c) 2026 Probo Inc <hello@probo.com>.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
-->

# Multi-agent compatibility

`@probo/skills` is an [Agent Plugins 1.0.0](https://agent-plugins.org/)
package. The portable core is `plugin.json`, Agent Skills under `skills/`, and
Probo MCP servers in `mcp.json`. Client-specific manifests carry the same
components to **Claude Code**, **Codex**, **OpenCode**, and **Cursor** until
those clients load the portable package directly.

## Portable package (Agent Plugins 1.0.0)

```
plugin.json    # $schema + plugin identity (spec §5)
skills/        # one directory per skill, each with SKILL.md (spec §7.1)
mcp.json       # $schema + mcpServers (spec §7.2)
```

Both documents declare
`https://agent-plugins.org/schemas/1.0.0/{plugin,mcp}.schema.json`, and the
spec requires the two versions to match.

Commands, marketplace catalogs, and the client manifests are outside the v1
format — Agent Plugins v1 standardizes only skills and MCP servers. Clients
ignore the extra files.

## Probo MCP servers

`mcp.json` ships both hosted instances as Streamable HTTP servers:

| Server | Endpoint |
| --- | --- |
| `probo-us` | `https://us.probo.com/api/mcp/v1` |
| `probo-eu` | `https://eu.probo.com/api/mcp/v1` |

Connect the server for your region. Skills that need to discover the region
call `listOrganizations` on each connected server and keep the one that returns
the target organization.

Authentication is **OAuth 2.0** only, discovered from
`/.well-known/oauth-protected-resource` on the instance root. Never configure a
static bearer token: Agent Plugins treats `headers` as visible package data,
and a pre-set `Authorization` header also stops Claude Code from starting the
OAuth flow.

### Self-hosted instances

Agent Plugins 1.0.0 performs no placeholder or environment-variable expansion
in remote MCP URLs (spec §7.2.1), so a self-hosted endpoint cannot ship in the
package. Add it in the agent:

```bash
claude mcp add --transport http probo https://probo.example.com/api/mcp/v1
codex mcp add --transport http probo https://probo.example.com/api/mcp/v1
```

The MCP path is always `<instance-root>/api/mcp/v1`.

## What works where

| Component | Agent Plugins client | Claude Code | Codex | OpenCode | Cursor |
| --- | --- | --- | --- | --- | --- |
| Probo MCP (OAuth) | ✅ `mcp.json` | ✅ `.mcp.json` | ✅ `.codex-plugin` + `.mcp.json` | ✅ Manual MCP config | ✅ IDE MCP settings |
| Skills (`SKILL.md`) | ✅ `skills/` | ✅ `skills/` | ✅ `skills/` via `.codex-plugin` | ✅ `.opencode/skills/` or `.claude/skills/` | ✅ Copy/symlink to `.cursor/skills/` |
| Commands | ❌ Outside v1 | ✅ `commands/` → `/probo:…` | ⚠️ Use skills instead | ⚠️ Native `skill` tool | ❌ Use skill or rules |
| Plugin manifest | `plugin.json` | `.claude-plugin/` | `.codex-plugin/` | Discovery paths (no manifest) | No native manifest |
| Marketplace catalog | — | `.claude-plugin/marketplace.json` (repo root or package) | `.agents/plugins/marketplace.json` (repo root or package) | — | — |

### Claude Code

**From the monorepo or GitHub** (repo-root catalog at
`.claude-plugin/marketplace.json`):

```bash
claude plugin marketplace add getprobo/probo
# or, from a local clone:
claude plugin marketplace add .
claude plugin install probo@probo
claude mcp login probo-us   # or /mcp in session
/probo:access-review Q3 GitHub review
```

**From the package directory** (catalog resolves `@probo/skills` from npm):

```bash
claude plugin marketplace add ./packages/skills/.claude-plugin
claude plugin install probo@probo
claude mcp login probo-us
```

Or install the plugin directory directly:

```bash
claude --plugin-dir ./packages/skills
```

### Codex

**From the monorepo or GitHub** (repo-root catalog at
`.agents/plugins/marketplace.json`):

```bash
codex plugin marketplace add getprobo/probo
# or, from a local clone:
codex plugin marketplace add .
codex plugin install probo@probo
codex mcp login probo-us
```

**From the package directory** (catalog at
`packages/skills/.agents/plugins/marketplace.json`):

```bash
codex plugin marketplace add ./packages/skills
codex plugin install probo@probo
codex mcp login probo-us
```

Or install the plugin directory directly:

```bash
codex plugin install ./packages/skills
codex mcp login probo-us
```

Skills load from `./skills/` via `.codex-plugin/plugin.json`. The repo-root
marketplace `source.path` is `./packages/skills`; the package-level
catalog uses `./` (plugin package root).

### OpenCode

OpenCode discovers skills at `.opencode/skills/`, `.claude/skills/`, and
`~/.config/opencode/skills/`. Options:

**Option A — symlink from this package:**

```bash
mkdir -p .opencode/skills
ln -s ../../packages/skills/skills/access-review .opencode/skills/access-review
ln -s ../../packages/skills/skills/open-source-compliance .opencode/skills/open-source-compliance
```

**Option B — Claude Code bridge:** install
[`opencode-claude-code-bridge`](https://www.npmjs.com/package/opencode-claude-code-bridge)
to import Claude plugins and MCP configs into OpenCode.

Configure a Probo MCP server in `opencode.json` or global OpenCode MCP
settings, then authenticate. Invoke via the native `skill` tool
(`access-review`).

### Cursor

1. Add a Probo MCP server in Cursor settings (HTTP URL:
   `https://us.probo.com/api/mcp/v1` or `https://eu.probo.com/api/mcp/v1`,
   OAuth).
2. Copy or symlink skills into `.cursor/skills/`:

```bash
mkdir -p .cursor/skills
cp -r packages/skills/skills/access-review .cursor/skills/
```

Reference the skill in chat or add a Cursor rule pointing at the skill.

## Portable vs agent-specific paths

| Path | Portable? |
| --- | --- |
| `plugin.json` | ✅ Agent Plugins manifest (spec §5) |
| `mcp.json` | ✅ Agent Plugins MCP configuration (spec §7.2) |
| `skills/<name>/SKILL.md` | ✅ Agent Skills standard |
| `skills/<name>/references/*.md` | ✅ Relative to skill directory |
| `.mcp.json` | Claude Code and Codex mirror of `mcp.json` (`type: http`) |
| `${CLAUDE_PLUGIN_ROOT}` | ❌ Claude Code only — avoid in skill bodies |
| `commands/*.md` | Claude Code slash commands only |

Skill bodies use **relative** `references/` paths so they work once the skill
directory is discovered, regardless of which agent loads it.

## npm package layout

```
@probo/skills/
  plugin.json                        # Agent Plugins manifest
  mcp.json                           # Agent Plugins MCP configuration
  skills/                            # Shared skills (all agents)
  commands/                          # Claude Code commands only
  .mcp.json                          # Claude Code / Codex MCP wiring
  .claude-plugin/plugin.json         # Claude Code manifest
  .claude-plugin/marketplace.json    # Claude marketplace (npm)
  .codex-plugin/plugin.json          # Codex manifest
  .agents/plugins/marketplace.json   # Codex marketplace (package-local)
```

Repo root (monorepo / `getprobo/probo` Git installs):

```
.claude-plugin/marketplace.json      # Claude marketplace → packages/skills
.agents/plugins/marketplace.json     # Codex marketplace → packages/skills
```

## Validation

```bash
npm --workspace @probo/skills run validate
```

`scripts/validate.mjs` enforces the Agent Plugins closed manifest and MCP
schemas (canonical `$schema` values, plugin name constraints, transport
variants, literal HTTPS URLs, no credential headers), Agent Skills frontmatter
rules, and that `.mcp.json` stays in lockstep with `mcp.json`.

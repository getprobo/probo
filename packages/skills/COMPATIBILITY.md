<!--
Copyright (c) 2026 Probo Inc <hello@probo.com>.

Permission to use, copy, modify, and/or distribute this software for any
purpose with or without fee is hereby granted, provided that the above
copyright notice and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
PERFORMANCE OF THIS SOFTWARE.
-->

# Multi-agent compatibility

`@probo/skills` targets **Claude Code**, **Codex**, **OpenCode**, and
other MCP-capable agents (including **Cursor**). The portable core is Probo
MCP plus Agent Skills–compatible `SKILL.md` files.

## What works where

| Component | Claude Code | Codex | OpenCode | Cursor |
| --- | --- | --- | --- | --- |
| Probo MCP (OAuth) | ✅ Plugin `.mcp.json` | ✅ `.codex-plugin` + `.mcp.json` | ✅ Manual MCP config | ✅ IDE MCP settings |
| Skills (`SKILL.md`) | ✅ `skills/` | ✅ `skills/` via `.codex-plugin` | ✅ `.opencode/skills/` or `.claude/skills/` | ✅ Copy/symlink to `.cursor/skills/` |
| Commands | ✅ `commands/` → `/probo:…` | ⚠️ Use skills instead | ⚠️ Native `skill` tool | ❌ Use skill or rules |
| Plugin manifest | `.claude-plugin/` | `.codex-plugin/` | Discovery paths (no manifest) | No native manifest |
| Marketplace catalog | `.claude-plugin/marketplace.json` (repo root or package) | `.agents/plugins/marketplace.json` (repo root or package) | — | — |

## Probo MCP (all agents)

Set the instance URL:

```bash
export PROBO_BASE_URL="https://your-probo-instance.example.com"
```

Endpoint: `${PROBO_BASE_URL}/mcp/v1` (HTTP, OAuth 2.0). Do not configure a
static bearer token — OAuth discovery uses
`/.well-known/oauth-protected-resource`.

### Claude Code

**From the monorepo or GitHub** (repo-root catalog at
`.claude-plugin/marketplace.json`):

```bash
claude plugin marketplace add getprobo/probo
# or, from a local clone:
claude plugin marketplace add .
claude plugin install probo@probo
claude mcp login probo   # or /mcp in session
/probo:access-review Q3 GitHub review
```

**From the package directory** (catalog resolves `@probo/skills` from npm):

```bash
claude plugin marketplace add ./packages/skills/.claude-plugin
claude plugin install probo@probo
claude mcp login probo
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
codex mcp login probo
```

**From the package directory** (catalog at
`packages/skills/.agents/plugins/marketplace.json`):

```bash
codex plugin marketplace add ./packages/skills
codex plugin install probo@probo
codex mcp login probo
```

Or install the plugin directory directly:

```bash
codex plugin install ./packages/skills
codex mcp login probo
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

Configure Probo MCP in `opencode.json` or global OpenCode MCP settings, then
authenticate. Invoke via the native `skill` tool (`access-review`).

### Cursor

1. Add Probo MCP in Cursor settings (HTTP URL: `${PROBO_BASE_URL}/mcp/v1`,
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
| `skills/<name>/SKILL.md` | ✅ Agent Skills standard |
| `skills/<name>/references/*.md` | ✅ Relative to skill directory |
| `.mcp.json` with `${PROBO_BASE_URL}` | ✅ Standard env var |
| `${CLAUDE_PLUGIN_ROOT}` | ❌ Claude Code only — avoid in skill bodies |
| `commands/*.md` | Claude Code slash commands only |

Skill bodies use **relative** `references/` paths so they work once the skill
directory is discovered, regardless of which agent loads it.

## npm package layout

```
@probo/skills/
  .claude-plugin/plugin.json         # Claude Code manifest
  .claude-plugin/marketplace.json    # Claude marketplace (npm)
  .codex-plugin/plugin.json          # Codex manifest
  .agents/plugins/marketplace.json   # Codex marketplace (package-local)
  .mcp.json                          # Shared MCP wiring
  skills/                            # Shared skills (all agents)
  commands/                          # Claude Code commands only
```

Repo root (monorepo / `getprobo/probo` Git installs):

```
.claude-plugin/marketplace.json      # Claude marketplace → packages/skills
.agents/plugins/marketplace.json     # Codex marketplace → packages/skills
```

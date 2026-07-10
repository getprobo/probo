# Agent skills (`packages/skills`)

npm package [`@probo/skills`](../../packages/skills) ships multi-agent
compliance skills and agent plugin wiring powered by the Probo MCP API.
Compatible with **Claude Code**, **Codex**, **OpenCode**, and **Cursor** (via
MCP + skills). See [`COMPATIBILITY.md`](../../packages/skills/COMPATIBILITY.md).

## What this package ships

A **skills package** bundling:

| Component | Role |
| --- | --- |
| `skills/` | Agent Skills–compatible workflow instructions |
| `.mcp.json` | Connects agents to Probo (`/mcp/v1`, OAuth 2.0) |
| `commands/` | Explicit slash commands (e.g. `access-review`, Claude Code only) |
| `agents/` | Optional specialized subagents |
| `hooks/` | Optional event automation |

Individual capabilities are namespaced under `probo`:

- Skills: `/probo:<skill-name>` (e.g. `/probo:open-source-compliance`,
  `/probo:missing-signatures`, `/probo:access-review`)
- Commands: `/probo:<command-name>` (e.g. `/probo:access-review`,
  `/probo:missing-signatures`)

Published to npm as `@probo/skills`. Agent-specific manifests (`.claude-plugin/`,
`.codex-plugin/`) ship inside the same package.

## Directory structure

```
.claude-plugin/marketplace.json   # repo root — Claude catalog for getprobo/probo
.agents/plugins/marketplace.json   # repo root — Codex catalog for getprobo/probo

packages/skills/
  .claude-plugin/
    plugin.json           # Claude Code manifest (required)
    marketplace.json      # Claude marketplace catalog (npm)
  .agents/plugins/
    marketplace.json      # Codex catalog when marketplace root is the package
  .codex-plugin/
    plugin.json           # Codex manifest
  .mcp.json               # Probo MCP server wiring
  skills/
    <skill-name>/
      SKILL.md
      references/
  commands/
  agents/
  hooks/
  scripts/validate.mjs
  package.json
  CHANGELOG.md
```

Only `plugin.json` belongs inside `.claude-plugin/`. All other directories
must sit at the package root.

## plugin.json rules

Claude Code validates the manifest strictly. Common pitfalls:

| Field | Expected type | Notes |
| --- | --- | --- |
| `name` | string | Skill namespace (`probo` → `/probo:open-source-compliance`) |
| `repository` | string URL | **Not** the npm-style `{ type, url }` object |
| `bugs` | string URL | **Not** the npm-style `{ url }` object |
| `version` | string | Bump on every release when using explicit versioning |

Run `npm --workspace @probo/skills run validate` before publishing.

## Probo MCP configuration

The package `.mcp.json` expects one environment variable:

- `PROBO_BASE_URL` — instance root URL

Authentication is OAuth 2.0 only. Users complete sign-in via `/mcp` or
`claude mcp login probo`. Do not document API keys or bearer tokens in the
plugin config — a pre-set `Authorization` header prevents Claude Code from
starting the OAuth flow.

## Adding a skill

1. Create `skills/<name>/SKILL.md` with YAML frontmatter (`name`, `description`).
2. Add `references/` for detailed workflow docs loaded on demand.
3. Validate and test:

```bash
npm --workspace @probo/skills run validate
claude --plugin-dir ./packages/skills
/probo:<name>
```

4. Update `packages/skills/CHANGELOG.md` under `## Unreleased`.

Skills must be self-contained — npm installs do not include `contrib/claude/`
from the monorepo.

## Adding a command

Use commands for explicit, user-invoked workflows on Claude Code only. Pair a
thin `commands/<name>.md` with a shared `skills/<name>/SKILL.md` so Codex and
OpenCode load the same workflow. Reference docs live under
`skills/<name>/references/` using paths relative to the skill directory (not
`${CLAUDE_PLUGIN_ROOT}`).

1. Create `commands/<name>.md` with frontmatter (`description`,
   `argument-hint`, `disable-model-invocation: true` when writes are involved).
2. Add reference docs under `skills/<name>/references/`.
3. Register paths in `scripts/validate.mjs`.
4. Test: `/probo:<name> <args>` after `claude --plugin-dir ./packages/skills`.

## Distribution

Published to npm as `@probo/skills`. Claude marketplace entry:

```json
{
  "source": {
    "source": "npm",
    "package": "@probo/skills"
  }
}
```

Release process: [`contrib/claude/release/skills.md`](release/skills.md).

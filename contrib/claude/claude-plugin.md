# Claude Code Plugin (`packages/claude-plugin`)

npm package [`@probo/claude-plugin`](../../packages/claude-plugin) that ships a
[Claude Code plugin](https://code.claude.com/docs/en/plugins) for open-source
compliance workflows powered by the Probo MCP API.

## What this plugin is

A **Claude plugin** (the installable unit) bundling:

| Component | Role |
| --- | --- |
| `.mcp.json` | Connects Claude to Probo (`/mcp/v1`, bearer auth) |
| `skills/` | Compliance workflow instructions Claude can invoke |
| `commands/` | Optional legacy slash commands |
| `agents/` | Optional specialized subagents |
| `hooks/` | Optional event automation |

Individual capabilities inside the plugin are **skills** (for example
`/probo:open-source-compliance`). The npm package name stays
`@probo/claude-plugin` because that matches Claude Code's distribution model.

## Directory structure

```
packages/claude-plugin/
  .claude-plugin/
    plugin.json           # Plugin manifest (required)
    marketplace.json      # Marketplace catalog for npm distribution
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
must sit at the plugin root.

## plugin.json rules

Claude Code validates the manifest strictly. Common pitfalls:

| Field | Expected type | Notes |
| --- | --- | --- |
| `name` | string | Skill namespace (`probo` → `/probo:open-source-compliance`) |
| `repository` | string URL | **Not** the npm-style `{ type, url }` object |
| `bugs` | string URL | **Not** the npm-style `{ url }` object |
| `version` | string | Bump on every release when using explicit versioning |

Run `npm --workspace @probo/claude-plugin run validate` before publishing.

## Probo MCP configuration

The plugin `.mcp.json` reads:

- `PROBO_BASE_URL` — instance root URL
- `PROBO_API_TOKEN` — personal API key or OAuth token

Users must export both before launching Claude Code. Document new required
variables in the plugin README when they change.

## Adding a skill

1. Create `skills/<name>/SKILL.md` with YAML frontmatter (`name`, `description`).
2. Add `references/` for detailed workflow docs loaded on demand.
3. Validate and test:

```bash
npm --workspace @probo/claude-plugin run validate
claude --plugin-dir ./packages/claude-plugin
/probo:<name>
```

4. Update `packages/claude-plugin/CHANGELOG.md` under `## Unreleased`.

Skills must be self-contained — npm installs do not include `contrib/claude/`
from the monorepo.

## Distribution

Published to npm as `@probo/claude-plugin`. Marketplace entry:

```json
{
  "source": {
    "source": "npm",
    "package": "@probo/claude-plugin"
  }
}
```

Release process: [`contrib/claude/release/claude-plugin.md`](release/claude-plugin.md).

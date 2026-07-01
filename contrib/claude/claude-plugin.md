# Claude Code Plugin (`packages/claude-plugin`)

npm package `@probo/claude-plugin` that ships a [Claude Code
plugin](https://code.claude.com/docs/en/plugins) with Probo development skills,
commands, agents, and hooks.

## Directory structure

```
packages/claude-plugin/
  .claude-plugin/
    plugin.json           # Plugin manifest (required)
    marketplace.json      # Marketplace catalog for npm distribution
  skills/
    <skill-name>/
      SKILL.md            # Skill definition with YAML frontmatter
      references/         # Optional deep-dive docs loaded on demand
      scripts/            # Optional helper scripts
  commands/               # Legacy flat Markdown commands (prefer skills/)
  agents/                 # Custom agent definitions
  hooks/                  # hooks.json and hook scripts
  bin/                    # Executables added to Bash tool PATH
  scripts/
    validate.mjs          # Pre-publish structural checks
  package.json
  CHANGELOG.md
```

Only `plugin.json` belongs inside `.claude-plugin/`. All component directories
(`skills/`, `commands/`, `agents/`, `hooks/`) must sit at the plugin root.

## plugin.json rules

Claude Code validates the manifest strictly. Common pitfalls:

| Field | Expected type | Notes |
| --- | --- | --- |
| `name` | string | Becomes the skill namespace (`probo` → `/probo:commit`) |
| `repository` | string URL | **Not** the npm-style `{ type, url }` object |
| `bugs` | string URL | **Not** the npm-style `{ url }` object |
| `version` | string | Bump on every release when using explicit versioning |

Run `npm --workspace @probo/claude-plugin run validate` before publishing.

## Adding a skill

1. Create `skills/<name>/SKILL.md`:

```markdown
---
name: my-skill
description: When to invoke this skill — be specific for auto-discovery
---

# My skill

Instructions for Claude…
```

2. Add supporting material under `references/`, `scripts/`, or `assets/` as
   needed. Skills must be self-contained — npm installs do not include
   `contrib/claude/` from the monorepo.

3. Validate and test locally:

```bash
npm --workspace @probo/claude-plugin run validate
claude --plugin-dir ./packages/claude-plugin
/probo:my-skill
```

4. Update `packages/claude-plugin/CHANGELOG.md` under `## Unreleased`.

## Adding commands, agents, or hooks

| Component | Location | Format |
| --- | --- | --- |
| Commands | `commands/<name>.md` | Markdown with frontmatter (legacy; prefer skills) |
| Agents | `agents/<name>.md` | Markdown agent definition |
| Hooks | `hooks/hooks.json` | Event handlers; scripts use `${CLAUDE_PLUGIN_ROOT}` |

See the [Claude Code plugins
reference](https://code.claude.com/docs/en/plugins-reference) for schemas.

## Syncing with contrib guides

Many Probo conventions live in `contrib/claude/*.md` and `.cursor/rules/`. When
adding a skill:

- Copy or summarize the relevant guide into the skill's `references/` directory
- Keep the `SKILL.md` body concise; load references only when needed
- Update both the skill and the contrib guide when conventions change

## Distribution

The package publishes to npm as `@probo/claude-plugin`. Consumers install via
the marketplace entry in `.claude-plugin/marketplace.json`:

```json
{
  "source": {
    "source": "npm",
    "package": "@probo/claude-plugin"
  }
}
```

For monorepo-local testing, use a relative marketplace path or
`claude --plugin-dir ./packages/claude-plugin`.

Release process: [`contrib/claude/release/claude-plugin.md`](release/claude-plugin.md).

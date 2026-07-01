# @probo/claude-plugin

Claude Code plugin for Probo development. Ships skills, commands, agents, and
hooks that encode Probo engineering conventions.

## Install

### From npm (recommended)

Add the bundled marketplace, then install the plugin:

```bash
claude plugin marketplace add ./packages/claude-plugin/.claude-plugin
claude plugin install probo@probo
```

When consuming the published package from npm, point the marketplace entry at
`@probo/claude-plugin` (see `.claude-plugin/marketplace.json`).

### Local development

Test changes without publishing:

```bash
claude --plugin-dir ./packages/claude-plugin
```

Or scaffold-style auto-load from the skills directory:

```bash
claude plugin init probo-dev   # one-time, outside this repo
# copy or symlink packages/claude-plugin into ~/.claude/skills/
```

## What's included

| Component | Location | Status |
| --- | --- | --- |
| Skills | `skills/` | `commit` skill (more coming) |
| Commands | `commands/` | placeholder |
| Agents | `agents/` | placeholder |
| Hooks | `hooks/` | placeholder |

Skills are namespaced as `/probo:<skill-name>` (for example `/probo:commit`).

## Adding content

See [`contrib/claude/claude-plugin.md`](../../contrib/claude/claude-plugin.md)
for the full authoring guide.

Quick checklist for a new skill:

1. Create `skills/<name>/SKILL.md` with YAML frontmatter (`name`, `description`)
2. Add optional `references/`, `scripts/`, or `assets/` subdirectories
3. Run `npm --workspace @probo/claude-plugin run validate`
4. Test with `claude --plugin-dir ./packages/claude-plugin`

## Release

Published to npm as `@probo/claude-plugin`. See
[`contrib/claude/release/claude-plugin.md`](../../contrib/claude/release/claude-plugin.md).

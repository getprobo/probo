# @probo/claude-plugin

Claude Code plugin for open-source compliance workflows. Bundles skills that
guide compliance work and wires Claude to the [Probo MCP
API](https://github.com/getprobo/probo/tree/main/pkg/server/api/mcp/v1) (270+
tools for GRC entities and operations).

## Install

### From npm

Add the marketplace catalog, then install the plugin:

```bash
claude plugin marketplace add ./packages/claude-plugin/.claude-plugin
claude plugin install probo@probo
```

When consuming the published package, the marketplace entry resolves
`@probo/claude-plugin` from npm (see `.claude-plugin/marketplace.json`).

### Configure Probo MCP

Set environment variables before starting Claude Code:

```bash
export PROBO_BASE_URL="https://your-probo-instance.example.com"
export PROBO_API_TOKEN="your-personal-api-key-or-oauth-token"
```

The plugin ships a `.mcp.json` that connects to `${PROBO_BASE_URL}/mcp/v1`
with bearer authentication. Create an API key in the Probo console under
Settings → API keys.

### Local development

```bash
claude --plugin-dir ./packages/claude-plugin
```

## What's included

| Component | Location | Purpose |
| --- | --- | --- |
| MCP | `.mcp.json` | Probo API connection |
| Skills | `skills/` | Compliance workflows |
| Commands | `commands/` | Reserved |
| Agents | `agents/` | Reserved |
| Hooks | `hooks/` | Reserved |

Skills are namespaced as `/probo:<skill-name>` (for example
`/probo:open-source-compliance`).

## Adding content

See [`contrib/claude/claude-plugin.md`](../../contrib/claude/claude-plugin.md).

```bash
npm --workspace @probo/claude-plugin run validate
claude --plugin-dir ./packages/claude-plugin
```

## Release

Published to npm as `@probo/claude-plugin`. See
[`contrib/claude/release/claude-plugin.md`](../../contrib/claude/release/claude-plugin.md).

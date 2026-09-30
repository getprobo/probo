# Agent skills (`packages/skills`)

npm package [`@probo/skills`](../../packages/skills) ships multi-agent
compliance skills and agent plugin wiring powered by the Probo MCP API. It is an
[Agent Plugins 1.0.0](https://agent-plugins.org/) package, and also carries
client-specific manifests for **Claude Code**, **Codex**, **OpenCode**, and
**Cursor**. See [`COMPATIBILITY.md`](../../packages/skills/COMPATIBILITY.md).

## What this package ships

A **skills package** bundling:

| Component | Role |
| --- | --- |
| `plugin.json` | Agent Plugins manifest (portable, spec §5) |
| `mcp.json` | Agent Plugins MCP configuration (portable, spec §7.2) |
| `skills/` | Agent Skills–compatible workflow instructions |
| `.mcp.json` | Claude Code / Codex mirror of `mcp.json` (`type: http`) |
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
  plugin.json             # Agent Plugins manifest (required, spec §5)
  mcp.json                # Agent Plugins MCP configuration (spec §7.2)
  skills/
    <skill-name>/
      SKILL.md
      references/
  .claude-plugin/
    plugin.json           # Claude Code manifest (required)
    marketplace.json      # Claude marketplace catalog (npm)
  .agents/plugins/
    marketplace.json      # Codex catalog when marketplace root is the package
  .codex-plugin/
    plugin.json           # Codex manifest
  .mcp.json               # Claude Code / Codex MCP wiring
  commands/
  agents/
  hooks/
  scripts/validate.mjs
  package.json
  CHANGELOG.md
```

`plugin.json`, `mcp.json`, and `skills/` are the fixed portable locations — a
conformant client reads nothing else. Only the Claude manifest belongs inside
`.claude-plugin/`; all other directories must sit at the package root.

## Agent Plugins rules

The portable manifest and MCP configuration use **closed** schemas: any extra
top-level field is a violation, and client-specific data belongs under
`extensions` keyed by reverse-domain namespace.

| Rule | Detail |
| --- | --- |
| `$schema` | Required in both files; must be the canonical `…/schemas/1.0.0/{plugin,mcp}.schema.json` identifier, and the versions must match |
| `name` | 1–64 chars, lowercase alphanumeric plus `-` and `.`, alphanumeric at both ends, no `--` or `..` |
| Transports | `stdio`, `streamable-http`, `sse` — Claude's `http` keyword is not portable |
| Remote URLs | Absolute, literal HTTPS, no userinfo or fragment. **No placeholder or env-var expansion**, so a self-hosted Probo URL cannot ship in the package |
| `headers` | Visible package data — never put credentials there. Probo MCP is OAuth 2.0 only |
| Skills | Immediate children of `skills/` containing `SKILL.md`; frontmatter `name` must match the directory name |

## Claude plugin.json rules

Claude Code validates its own manifest strictly. Common pitfalls:

| Field | Expected type | Notes |
| --- | --- | --- |
| `name` | string | Skill namespace (`probo` → `/probo:open-source-compliance`) |
| `repository` | string URL | **Not** the npm-style `{ type, url }` object |
| `bugs` | string URL | **Not** the npm-style `{ url }` object |
| `version` | string | Must match `package.json`; validation enforces it |

Run `npm --workspace @probo/skills run validate` before publishing.

## Probo MCP configuration

Both hosted instances ship as literal endpoints, in `mcp.json`
(`streamable-http`) and mirrored in `.mcp.json` (`http`):

| Server | Endpoint |
| --- | --- |
| `probo-us` | `https://us.probo.com/api/mcp/v1` |
| `probo-eu` | `https://eu.probo.com/api/mcp/v1` |

The MCP API is mounted at `<instance-root>/api/mcp/v1` (`/api` prefix included
— `/mcp/v1` hits the console SPA). Self-hosted instances are added in the agent
because the spec forbids expansion in remote URLs.

Authentication is OAuth 2.0 only, discovered from
`/.well-known/oauth-protected-resource` on the instance root. Users complete
sign-in via `/mcp` or `claude mcp login probo-us`. Do not document API keys or
bearer tokens in the plugin config — a pre-set `Authorization` header prevents
Claude Code from starting the OAuth flow.

## Adding a skill

1. Create `skills/<name>/SKILL.md` with YAML frontmatter (`name` matching the
   directory, `description` under 1024 chars).
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

# API Surface Rules

Every feature must be exposed through **all four interfaces**: GraphQL, MCP, CLI, and n8n. When adding a new endpoint or editing an existing type, keep all four in sync:

- **GraphQL** — `pkg/server/api/console/v1/graphql/*.graphql` (+ codegen) — see [`contrib/claude/graphql.md`](graphql.md)
- **MCP** — `pkg/server/api/mcp/v1/` (+ codegen) — see [`contrib/claude/mcp.md`](mcp.md)
- **CLI** — `pkg/cmd/` — see [`contrib/claude/cli.md`](cli.md)
- **n8n** — `packages/n8n-node/` — see [`contrib/claude/n8n.md`](n8n.md)

If you add a mutation in GraphQL, add the corresponding MCP tool, CLI command, and n8n node. If you rename or change a type, update it everywhere.

Every new Go API endpoint must have end-to-end tests in `e2e/`.

# AGENTS.md

Detailed guides for specific subsystems live in `contrib/claude/`:

- [`contrib/claude/make.md`](contrib/claude/make.md) — GNUmakefile targets, codegen, overridable variables
- [`contrib/claude/api-surface.md`](contrib/claude/api-surface.md) — GraphQL / MCP / CLI / n8n sync rules
- [`contrib/claude/go-style.md`](contrib/claude/go-style.md) — Go project deps, style (declarations, calls, imports, errors, naming, logging)
- [`contrib/claude/go-testing.md`](contrib/claude/go-testing.md) — Go test conventions (parallel, require vs assert, naming)
- [`contrib/claude/go-service.md`](contrib/claude/go-service.md) — Go service orchestration (Run, graceful shutdown, crash propagation)
- [`contrib/claude/go-worker.md`](contrib/claude/go-worker.md) — Go worker pattern (poll-based, bounded concurrency, FOR UPDATE SKIP LOCKED)
- [`contrib/claude/gid.md`](contrib/claude/gid.md) — Global identifiers (GID layout, TenantID, entity type registry)
- [`contrib/claude/coredata.md`](contrib/claude/coredata.md) — Data access layer (Scoper, SQL patterns, filters, order fields, migrations)
- [`contrib/claude/logging.md`](contrib/claude/logging.md) — Structured logging (PII-free rules, field helpers, logger wiring)
- [`contrib/claude/graphql.md`](contrib/claude/graphql.md) — Go GraphQL backend (gqlgen, @goModel, connection types, cursor pagination)
- [`contrib/claude/mcp.md`](contrib/claude/mcp.md) — MCP API patterns (specification.yaml, mcpgen, resolvers, type helpers)
- [`contrib/claude/cli.md`](contrib/claude/cli.md) — CLI command patterns (cobra, huh prompts, pagination, output formatting)
- [`contrib/claude/authorization.md`](contrib/claude/authorization.md) — IAM policy-based authorization (policies, conditions, roles, AuthorizationAttributer)
- [`contrib/claude/validation.md`](contrib/claude/validation.md) — Validation framework (fluent API, validators, error codes, propagation)
- [`contrib/claude/e2e.md`](contrib/claude/e2e.md) — End-to-end testing (factory builders, RBAC tests, tenant isolation, assertions)
- [`contrib/claude/agent.md`](contrib/claude/agent.md) — Agent orchestration framework (tools, handoffs, execution)
- [`contrib/claude/app-arborescence.md`](contrib/claude/app-arborescence.md) — Frontend app folder layout (pages, routes, loaders, skeletons, _components)
- [`contrib/claude/relay.md`](contrib/claude/relay.md) — Frontend Relay client (queries, fragments, mutations, pagination)
- [`contrib/claude/react-components.md`](contrib/claude/react-components.md) — React component shape (file/export, props, configure vs data via hooks)
- [`contrib/claude/ui.md`](contrib/claude/ui.md) — @probo/ui, Tailwind, tailwind-variants, folders, skeletons, compound components
- [`contrib/claude/commit.md`](contrib/claude/commit.md) — Commit message conventions
- [`contrib/claude/license.md`](contrib/claude/license.md) — ISC license header (all file types)
- [`contrib/claude/release.md`](contrib/claude/release.md) — Release process (version bump, changelog, tag, push)
- [`contrib/claude/sandbox.md`](contrib/claude/sandbox.md) — Lima sandbox environments (create, manage, access services)
- [`contrib/claude/n8n.md`](contrib/claude/n8n.md) — n8n community node (resources, operations, GraphQL helpers)

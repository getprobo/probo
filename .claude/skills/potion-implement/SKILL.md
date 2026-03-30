---
name: potion-implement
description: >
  Master implementation orchestrator for Probo. Analyzes tasks, determines
  which language stacks are involved (Go backend, TypeScript frontend, or
  both), and delegates to stack-specific implementer agents. For cross-stack
  tasks, orchestrates sequentially -- upstream first, then downstream with
  actual changes as context. Use when someone asks to "add", "create",
  "build", "implement", "write", or "code" anything. Also triggers for
  tickets, specs, feature descriptions, or any request to make code changes.
allowed-tools: Read, Glob, Grep, Agent
model: opus
effort: high
---

# Probo -- Master Implementation Orchestrator

This skill does NOT implement code itself. It analyzes incoming tasks,
determines which stack(s) are involved, and delegates to the right
stack-specific implementer agent(s).

## Load guidelines

Before analyzing any task, read the shared guidelines and every stack's index:

- **Shared conventions:** `.claude/guidelines/shared.md`
- **Go Backend:** `.claude/guidelines/go-backend/index.md`
- **TypeScript Frontend:** `.claude/guidelines/typescript-frontend/index.md`

## Stack routing table

Use this table to map modules and file paths to their owning stack.

### Go Backend (Go 1.26)
- **Frameworks:** chi router, gqlgen, mcpgen, pgx, testify
- **Modules:** cmd, pkg/server, pkg/probo, pkg/iam, pkg/trust, pkg/coredata, pkg/validator, pkg/gid, pkg/agent, pkg/llm, pkg/agents, pkg/cmd, pkg/cli, pkg/certmanager, pkg/webhook, pkg/mailer, pkg/slack, pkg/connector, pkg/bootstrap, e2e
- **Implementer agent:** `potion-go-backend-implementer`
- **Guidelines:** `.claude/guidelines/go-backend/`

### TypeScript Frontend (React 19 + Relay 19)
- **Frameworks:** React, Relay, React Router v7, Tailwind CSS v4, Vite, Storybook 10
- **Modules:** apps/console, apps/trust, packages/ui, packages/relay, packages/helpers, packages/hooks, packages/emails, packages/n8n-node, packages/routes, packages/coredata, packages/i18n
- **Implementer agent:** `potion-typescript-frontend-implementer`
- **Guidelines:** `.claude/guidelines/typescript-frontend/`

## Task analysis

For every incoming task, run through these steps before spawning any agent:

1. **Read the task description.** Understand what is being asked -- feature,
   bugfix, refactor, migration, etc.
2. **Identify affected modules.** Look for file paths, feature names, module
   names, or domain concepts that map to known modules.
3. **Map modules to stacks** using the routing table above. Each module belongs
   to exactly one stack.
4. **Classify the task:**
   - **Single-stack** -- all affected modules belong to one stack.
   - **Cross-stack** -- affected modules span both Go backend and TypeScript frontend.

## Critical rule: three-interface sync

Every new feature must be exposed through all three interfaces and kept in sync:

1. **GraphQL** -- `pkg/server/api/console/v1/schema.graphql` (+ codegen)
2. **MCP** -- `pkg/server/api/mcp/v1/specification.yaml` (+ codegen)
3. **CLI** -- `pkg/cmd/`

If the task adds a new domain entity or mutation, ensure all three are planned.
Every new Go API endpoint must also have end-to-end tests in `e2e/`.

## Single-stack delegation

When only one stack is involved:

1. Spawn the appropriate implementer agent with the full task description.
2. Let it handle the implementation end-to-end.
3. No further orchestration needed.

## Cross-stack orchestration

When both stacks are involved, order matters. Implement upstream before
downstream so that downstream agents can reference the actual changes.

### Step-by-step

1. **Determine dependency order** using the direction rules below.
2. **Spawn the upstream implementer first.** Pass it the full task description
   scoped to its stack. Wait for it to complete.
3. **Read upstream changes.** After the upstream agent finishes, read the files
   it created or modified. Extract the contract -- GraphQL schema changes,
   response types, function signatures.
4. **Spawn the downstream implementer** with upstream context:
   ```
   "The Go backend implementer created {summary of changes}.
   Here are the relevant details: {schema changes, new types, etc.}
   Now implement the TypeScript frontend part that integrates with these changes."
   ```
5. **Verify coherence.** After both agents finish, check that the downstream
   implementation actually uses the upstream contract correctly -- matching
   GraphQL operations, field names, type shapes, etc.

### Dependency direction rules

| Task type | Order | Reasoning |
|-----------|-------|-----------|
| New API + frontend page | Go Backend then TypeScript Frontend | Frontend consumes the GraphQL API |
| Frontend form + backend validation | Go Backend then TypeScript Frontend | Validation defines constraints |
| Schema change + UI update | Go Backend then TypeScript Frontend | Schema generates types |
| Independent changes | Parallel | No dependency |
| Bug fix in one stack | Single stack only | No cross-stack coordination |

### Cross-stack contract: the GraphQL schema

The primary contract between stacks is the `.graphql` schema file:

1. Go backend edits `pkg/server/api/console/v1/schema.graphql`
2. `go generate` regenerates Go resolver stubs and types (gqlgen)
3. Go developer implements resolver bodies
4. `npm run relay` regenerates TypeScript types from the same `.graphql` file
5. Frontend components consume generated types via Relay fragments

After the Go implementer modifies the schema, tell the TypeScript implementer
to run `npm run relay` and use the generated types.

## When the stack is unclear

If the task description does not clearly map to any stack in the routing table,
do NOT guess. Ask the user:

> "This task could touch the Go backend or the TypeScript frontend. Which
> stack(s) should I target?"

## Post-orchestration checklist

After all implementer agents have finished:

- [ ] Every affected stack had its implementer agent spawned
- [ ] Cross-stack contracts are coherent (GraphQL types match, endpoints align)
- [ ] No orphaned references (e.g., frontend calling an API that was not created)
- [ ] Shared conventions from guidelines were respected across all stacks
- [ ] If a new entity was added: GraphQL + MCP + CLI all present
- [ ] If Go API changed: e2e tests in `e2e/` were created or updated
- [ ] ISC license headers present on all new files

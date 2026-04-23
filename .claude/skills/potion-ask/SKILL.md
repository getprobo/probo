---
name: potion-ask
description: >
  Answers questions about the Probo codebase across both Go backend and
  TypeScript frontend stacks. Use when someone asks "where is...", "how does
  X work", "why is Y done this way", "what pattern does Z use", "explain the
  architecture", "find the code that handles...", "what does this module do",
  or any question about understanding this project. Also triggers for
  onboarding questions like "how do I get started", "what should I know",
  "walk me through the codebase", or "how are the stacks connected".
allowed-tools: Read, Glob, Grep
model: sonnet
effort: medium
---

# Probo -- Codebase Q&A

Before answering, read the guidelines at `.claude/guidelines/` -- start with
`shared.md` for cross-stack conventions, then check the relevant stack's
`index.md` for architecture overview and topic files (`patterns.md`,
`conventions.md`, `pitfalls.md`, `testing.md`).

## Stack routing

Determine which stack the question targets before exploring:

| Signal | Stack | Guidelines path |
|--------|-------|----------------|
| Go, `pkg/`, `cmd/`, `e2e/`, coredata, GraphQL resolvers, MCP, CLI, service layer, IAM, SCIM, SAML | Go Backend | `.claude/guidelines/go-backend/` |
| TypeScript, React, Relay, `apps/`, `packages/`, frontend, UI, components, hooks, Vite | TypeScript Frontend | `.claude/guidelines/typescript-frontend/` |
| GraphQL schema, GID, cross-stack, API contract, deployment, CI/CD, license | Shared | `.claude/guidelines/shared.md` |

## Answering strategy

1. **Check guidelines first.** Most architecture and pattern questions are
   already answered there. Do not explore what is already documented.
2. **Locate the module.** Use the module map below to narrow scope.
3. **Explore with precision.** Grep and Glob to find specific code. Read
   files to confirm. Never say "it is probably in..." -- find it.
4. **Cite your sources.** Reference specific files and line numbers.

## Module map -- Go Backend

| Module | Path | Purpose |
|--------|------|---------|
| cmd | `cmd/` | Binary entrypoints: `probod`, `prb`, `probod-bootstrap`, `acme-keygen` |
| pkg/server | `pkg/server/` | HTTP server, chi router, middleware stack, all API surface handlers |
| pkg/server/api/console/v1 | `pkg/server/api/console/v1/` | Console GraphQL API (gqlgen) -- 80+ type mapping files |
| pkg/server/api/trust/v1 | `pkg/server/api/trust/v1/` | Trust center GraphQL API -- NDA directive, read-only |
| pkg/server/api/connect/v1 | `pkg/server/api/connect/v1/` | IAM GraphQL API + SAML/OIDC/SCIM handlers |
| pkg/server/api/mcp/v1 | `pkg/server/api/mcp/v1/` | MCP API (mcpgen) -- AI-agent access to domain objects |
| pkg/probo | `pkg/probo/` | Core business logic -- 40+ domain sub-services |
| pkg/iam | `pkg/iam/` | Identity and access management, ABAC policy evaluation |
| pkg/iam/policy | `pkg/iam/policy/` | Pure in-memory IAM policy evaluator |
| pkg/iam/scim | `pkg/iam/scim/` | SCIM 2.0 provisioning bridge |
| pkg/trust | `pkg/trust/` | Public trust center service layer |
| pkg/coredata | `pkg/coredata/` | All raw SQL, entity types, filters, migrations |
| pkg/validator | `pkg/validator/` | Fluent validation framework |
| pkg/gid | `pkg/gid/` | 192-bit tenant-scoped entity identifiers |
| pkg/agent | `pkg/agent/` | LLM agent orchestration framework |
| pkg/llm | `pkg/llm/` | Provider-agnostic LLM abstraction |
| pkg/cmd | `pkg/cmd/` | CLI commands for `prb` tool (cobra) |
| e2e | `e2e/` | End-to-end integration tests |

## Module map -- TypeScript Frontend

| Module | Path | Purpose |
|--------|------|---------|
| apps/console | `apps/console/` | Admin dashboard SPA (React + Relay, port 5173) |
| apps/trust | `apps/trust/` | Public trust center SPA (React + Relay, port 5174) |
| packages/ui | `packages/ui/` | Shared design system (Tailwind Variants, Radix primitives) |
| packages/relay | `packages/relay/` | Relay FetchFunction factory + typed error classes |
| packages/helpers | `packages/helpers/` | Domain formatters, enum labels/variants, utility functions |
| packages/hooks | `packages/hooks/` | Shared React hooks |
| packages/emails | `packages/emails/` | React Email templates for transactional emails |
| packages/n8n-node | `packages/n8n-node/` | n8n community node for Probo API |

## Canonical examples

These files represent "the right way" in this project:

| File | Demonstrates |
|------|-------------|
| `pkg/coredata/asset.go` | Complete coredata entity: LoadByID, Insert, Update, Delete, CursorKey, AuthorizationAttributes, Snapshot |
| `pkg/probo/vendor_service.go` | Service layer pattern: Request struct, Validate(), pg.WithTx, webhook in same tx |
| `pkg/server/api/console/v1/v1_resolver.go` | GraphQL resolver: authorize, ProboService, service call, type mapping |
| `pkg/iam/policy/example_test.go` | Policy DSL: Allow/Deny helpers, ABAC conditions, Go example tests |
| `e2e/console/vendor_test.go` | E2E test: factory builders, RBAC, tenant isolation, timestamp assertions |
| `apps/console/src/routes/documentsRoutes.ts` | New-style route definitions using Loader components (no withQueryRef) |
| `apps/console/src/pages/organizations/documents/DocumentsPageLoader.tsx` | Canonical Loader component: useQueryLoader + useEffect + Suspense |
| `apps/trust/src/pages/DocumentPage.tsx` | Full page with colocated queries/mutations, typed error handling |
| `packages/ui/src/Atoms/Badge/Badge.tsx` | Canonical UI atom: tv() variant factory, asChild/Slot, typed props |
| `packages/helpers/src/audits.ts` | Domain helper: as const enum array, getXLabel with Translator, getXVariant |

## How to handle different question types

**"Where is X?"** -- Grep, check module map, return exact file + lines.

**"How does X work?"** -- Find entry point, trace data flow through layers,
explain each step with file references.

**"Why is X done this way?"** -- Check guidelines for rationale, then
`contrib/claude/` reference docs, then code comments. If no rationale exists,
say so -- do not invent.

**"What pattern for X?"** -- Reference the relevant stack's patterns in
guidelines. Point to the canonical example that best matches.

**"How do I get started?"** -- Walk through: structure (shared.md) ->
architecture (stack index.md) -> patterns -> canonical examples ->
`make build` / `make test`.

**"How are the stacks connected?"** -- Explain the GraphQL schema contract
from shared.md. The Go backend authors `.graphql` schema files, gqlgen
generates Go resolvers, and the Relay compiler generates TypeScript types
from the same `.graphql` files.

## Key patterns quick reference -- Go Backend

- **Two-level service tree**: `Service` (global) -> `WithTenant(tenantID)` -> `TenantService` with sub-services
- **Request struct + Validate()**: every mutating method takes a `*Request` with fluent validation
- **All SQL in pkg/coredata only**: no other package may contain SQL queries
- **pgx.StrictNamedArgs**: never NamedArgs (approval blocker)
- **Error wrapping**: `fmt.Errorf("cannot <action>: %w", err)` (never "failed to")
- **Scoper for tenant isolation**: entity structs have no TenantID field
- **Three interfaces**: every feature must have GraphQL + MCP + CLI endpoints

## Key patterns quick reference -- TypeScript Frontend

- **Relay colocated operations**: all GraphQL in component files, not `hooks/graph/`
- **Loader component pattern**: `useQueryLoader` + `useEffect` (not deprecated `withQueryRef`)
- **tv() for variants**: tailwind-variants, not cn()
- **useMutation + useToast**: not deprecated `useMutationWithToasts`
- **Permission fragments**: `canCreate: permission(action: "core:asset:create")`
- **Dual Relay environments** (console): `coreEnvironment` + `iamEnvironment`

## Rules

- Never guess. If you cannot find it, say so and suggest where to look.
- Prefer code snippets from the actual codebase over abstract descriptions.
- Note migrations or inconsistencies when relevant (e.g., "module X uses the
  old pattern, the rest of the codebase uses the new one").
- Keep answers focused. Answer the question, then offer to go deeper.
- For complex exploration, delegate to the explorer agent.

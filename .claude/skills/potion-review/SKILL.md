---
name: potion-review
description: >
  Reviews code for Probo across both Go backend and TypeScript frontend
  stacks. Determines which stacks are in the diff, applies stack-specific
  review checklists, and can delegate to specialized reviewer sub-agents
  for large changes. Use when someone asks to "review", "check", "audit",
  "look at", or "verify" code changes, a PR, or specific files.
allowed-tools: Read, Glob, Grep, Agent
model: opus
effort: high
---

# Probo -- Multi-Stack Code Review

## Load guidelines

Before reviewing, read the shared conventions and each stack's overview:

- **Shared conventions:** `.claude/guidelines/shared.md`
- **Go Backend:** `.claude/guidelines/go-backend/index.md`
- **TypeScript Frontend:** `.claude/guidelines/typescript-frontend/index.md`

## Stack routing

Map every file in the diff to a stack using paths and module ownership.

### Go Backend
- **Paths:** `pkg/`, `cmd/`, `e2e/`, `*.go`
- **Guidelines:** `.claude/guidelines/go-backend/`

### TypeScript Frontend
- **Paths:** `apps/`, `packages/`, `*.ts`, `*.tsx`
- **Guidelines:** `.claude/guidelines/typescript-frontend/`

## Review strategy

Choose the approach based on the size and stack spread of the change.

### Small changes (1-3 files, single stack)
Run the review checklist below directly using that stack's guidelines -- no
need for sub-agents.

### Medium changes (4-10 files, 1-2 stacks)
Spawn 2-3 relevant topic reviewers based on what the changes touch. Tell each
reviewer which stack's guidelines to load:
- Backend route/service changes -> `potion-pattern-reviewer` + `potion-architecture-reviewer`
- Frontend component changes -> `potion-style-reviewer` + `potion-test-reviewer`
- Database migrations -> `potion-security-reviewer` + `potion-architecture-reviewer`
- New feature across modules -> `potion-architecture-reviewer` + `potion-pattern-reviewer` + `potion-test-reviewer`

### Large changes (10+ files, multiple stacks)
Spawn all available topic reviewers in parallel. For each reviewer, pass the
stack context so it knows which guidelines to load:
```
Review these files using {stack_name} guidelines:
- Architecture: .claude/guidelines/{stack_name}/index.md
- Patterns: .claude/guidelines/{stack_name}/patterns.md
- Conventions: .claude/guidelines/{stack_name}/conventions.md
- Testing: .claude/guidelines/{stack_name}/testing.md
```

## Topic reviewer dispatch with stack context

The master reviewer PASSES stack context to each topic reviewer -- reviewers do
not detect it themselves.

- **potion-architecture-reviewer** -- module placement, layer boundaries, dependencies:
  - For Go Backend files -> load `.claude/guidelines/go-backend/index.md`
  - For TypeScript Frontend files -> load `.claude/guidelines/typescript-frontend/index.md`

- **potion-pattern-reviewer** -- error handling, data access, DI, type usage:
  - For Go Backend files -> load `.claude/guidelines/go-backend/patterns.md`
  - For TypeScript Frontend files -> load `.claude/guidelines/typescript-frontend/patterns.md`

- **potion-test-reviewer** -- test quality, coverage, conventions:
  - For Go Backend files -> load `.claude/guidelines/go-backend/testing.md`
  - For TypeScript Frontend files -> load `.claude/guidelines/typescript-frontend/testing.md`

- **potion-security-reviewer** -- auth, data exposure, injection, SQL safety:
  - For Go Backend files -> load `.claude/guidelines/go-backend/pitfalls.md`
  - For TypeScript Frontend files -> load `.claude/guidelines/typescript-frontend/pitfalls.md`

- **potion-style-reviewer** -- naming, formatting, exports, conventions:
  - For Go Backend files -> load `.claude/guidelines/go-backend/conventions.md`
  - For TypeScript Frontend files -> load `.claude/guidelines/typescript-frontend/conventions.md`

- **potion-duplication-reviewer** -- code duplication, missed reuse:
  - For Go Backend files -> load `.claude/guidelines/go-backend/patterns.md`
  - For TypeScript Frontend files -> load `.claude/guidelines/typescript-frontend/patterns.md`

Each sub-agent returns findings in JSON format. After all complete:
1. Collect all findings
2. Deduplicate (same file:line from multiple agents -> keep the most specific)
3. Sort by severity (blockers first, then suggestions)
4. Group findings by stack
5. Present unified report using the format below

## Cross-stack review

For changes spanning both stacks, additionally check:

- [ ] **API contract alignment** -- does the frontend consume what the backend provides? Are GraphQL field names and types consistent?
- [ ] **Schema consistency** -- `.graphql` schema changes match both Go resolver implementations and Relay fragment expectations
- [ ] **Cross-stack imports** -- flag if a stack imports directly from another stack's internals (should go through the GraphQL contract)
- [ ] **Migration ordering** -- database/schema changes are applied before code that depends on them
- [ ] **Three-interface rule** -- if a new mutation was added in GraphQL, check for corresponding MCP tool and CLI command

## Review checklist -- Go Backend

### Architecture and Design
- [ ] Change is in the correct module (see Go Backend module map in guidelines)
- [ ] Respects layer boundaries: resolver -> service -> coredata (no SQL outside coredata)
- [ ] No circular dependencies introduced
- [ ] Public API surface is intentional

### Pattern compliance
- [ ] Two-level service tree followed (Service -> TenantService -> sub-services)
- [ ] Request struct with `Validate()` for mutating methods
- [ ] `pgx.StrictNamedArgs` used (never `pgx.NamedArgs` -- approval blocker)
- [ ] `SQLFragment()` returns static SQL (no conditional string building -- approval blocker)
- [ ] Error wrapping uses `fmt.Errorf("cannot <action>: %w", err)` (never "failed to" -- approval blocker)
- [ ] Scoper pattern used for tenant isolation (no TenantID on entity structs)
- [ ] `maps.Copy` for argument merging
- [ ] Cursor-based pagination (not OFFSET)

### Error handling
- [ ] Sentinel errors from coredata mapped to domain errors in service layer
- [ ] GraphQL resolvers use `gqlutils.Internal(ctx)` for unexpected errors (log first)
- [ ] MCP resolvers use `MustAuthorize()` with panic recovery middleware
- [ ] No bare `return err` without wrapping

### Testing
- [ ] `t.Parallel()` at top-level AND every subtest (approval blocker)
- [ ] `require` for preconditions, `assert` for value checks
- [ ] E2E tests cover RBAC (owner/admin/viewer) and tenant isolation
- [ ] Factory builders used for test data
- [ ] Inline GraphQL queries as package-level `const` strings

### Naming and style
- [ ] `type ()`, `const ()`, `var ()` grouped blocks (not individual declarations)
- [ ] String-based enums, not iota
- [ ] One argument per line or all on one line (never mixed -- approval blocker)
- [ ] Short receiver names matching type initial
- [ ] ISC license header with current year

### Observability
- [ ] Structured logging with `go.gearno.de/kit/log` typed fields
- [ ] No PII, PHI, or sensitive data in log messages
- [ ] Correlation IDs propagated via context

## Review checklist -- TypeScript Frontend

### Architecture and Design
- [ ] Change is in the correct module (apps/ vs packages/)
- [ ] Feature-slice architecture respected (pages organized by domain)
- [ ] No circular dependencies between packages

### Pattern compliance
- [ ] Relay operations colocated in component files (not in `hooks/graph/`)
- [ ] Loader component pattern used (not deprecated `withQueryRef` -- approval blocker)
- [ ] `useMutation` + `useToast` (not deprecated `useMutationWithToasts`)
- [ ] `tv()` from tailwind-variants for variant logic
- [ ] Permission fragments used for access control UI gating
- [ ] Correct Relay environment for page area (core vs IAM)

### Error handling
- [ ] Mutation `onCompleted`/`onError` callbacks handle errors
- [ ] Error boundaries in place for route groups
- [ ] `formatError()` from `@probo/helpers` for user-facing messages

### Testing
- [ ] Storybook stories for new UI atoms/molecules in packages/ui
- [ ] Vitest tests for new utility functions in packages/helpers
- [ ] No unit tests required in apps/ (covered by Go e2e tests)

### Types and safety
- [ ] No hand-written TypeScript interfaces for GraphQL data (use Relay generated types)
- [ ] `z.infer<typeof schema>` for form data types (zod as single source of truth)
- [ ] Named exports everywhere (default exports only for lazy-loaded pages)

### Naming and style
- [ ] PascalCase for components, camelCase for hooks, files named correctly
- [ ] Import ordering: external, aliased (#/), relative
- [ ] All user-visible strings through `useTranslate()` hook
- [ ] ISC license header with current year

## Severity classification

**Blockers** (must fix before merge):
- Security issues (auth bypass, data exposure, SQL injection)
- Approval blockers from CLAUDE.md (pgx.NamedArgs, "failed to" errors, missing t.Parallel, withQueryRef, mixed multiline style)
- Missing error handling in critical paths
- Pattern violations that set a bad precedent
- Missing tests for new functionality
- Cross-stack contract mismatches
- Missing three-interface sync (GraphQL without MCP/CLI)

**Suggestions** (nice to have):
- Minor naming improvements
- Extra test cases for edge cases
- Documentation improvements
- Performance optimizations

## How to report each finding

For each finding, use this format:

```
**[BLOCKER/SUGGESTION]** {file}:{line} -- {what is wrong}
  Stack: {Go Backend / TypeScript Frontend}
  Why: {reference to stack-specific guideline or pattern that this violates}
  Fix: {specific suggestion, ideally with code or a reference to the canonical example}
```

## Aggregation

After all topic reviewers have returned their findings:

1. **Collect** all findings from every reviewer
2. **Deduplicate** -- same file:line reported by multiple reviewers -> keep the most specific finding
3. **Sort** by severity (blockers first, then suggestions)
4. **Group by stack** -- present findings under their stack heading so the developer knows which context to look at
5. **Cross-stack summary** -- if the change spans both stacks, add a summary section highlighting any cross-stack issues (contract mismatches, type inconsistencies, import violations)

## Common pitfalls to watch for

These are real issues found during codebase analysis:

**Go Backend -- approval blockers:**
- `pgx.NamedArgs` instead of `pgx.StrictNamedArgs`
- Conditional string building in `SQLFragment()`
- Error messages starting with "failed to" instead of "cannot"
- Missing `t.Parallel()` in e2e subtests
- `panic` in GraphQL resolvers
- Mixed inline/expanded multiline function call style

**TypeScript Frontend -- approval blockers:**
- `withQueryRef` in route definitions
- `useMutationWithToasts` hook

**Cross-cutting:**
- Three-interface sync: GraphQL mutation without matching MCP tool and CLI command
- ISC license header with outdated year
- Access control in UI conditionals instead of ABAC policies

## Reference files

### Go Backend
- Canonical implementation: `pkg/probo/vendor_service.go`
- Canonical test: `e2e/console/vendor_test.go`
- Guidelines: `.claude/guidelines/go-backend/`

### TypeScript Frontend
- Canonical implementation: `apps/console/src/pages/organizations/documents/DocumentsPageLoader.tsx`
- Canonical test: `packages/helpers/src/file.test.ts`
- Guidelines: `.claude/guidelines/typescript-frontend/`

- Shared guidelines: `.claude/guidelines/shared.md`

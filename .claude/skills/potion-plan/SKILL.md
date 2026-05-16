---
name: potion-plan
description: >
  Plans feature implementations, refactors, and architectural changes in
  Probo across its Go backend and TypeScript frontend stacks. Identifies
  which stacks are involved, determines execution order based on data
  flow, and produces a stack-labeled, step-by-step plan with exact file
  paths, canonical pattern references, codegen commands, and the
  four-surface API checklist. Use when someone asks to "plan", "design",
  "break down", "spec out", "architect", or "how should I implement"
  something. Triggers on tickets, user stories, feature requests, or
  questions like "what files would I need to change for X" or "what's the
  best approach for X". Always runs BEFORE `/potion-implement` for
  non-trivial work.
allowed-tools: Read, Write, Glob, Grep, AskUserQuestion, Agent, TodoWrite
model: opus
effort: high
---

# Probo — Multi-Stack Implementation Planning

Before planning, load:
- `.claude/guidelines/shared.md` — cross-cutting rules
- `.claude/guidelines/go-backend/index.md` — Go architecture
- `.claude/guidelines/typescript-frontend/index.md` — TS architecture

## When to use this skill

- Planning a new feature that may span the Go backend and TS frontend
- Designing an architectural change (e.g. new domain entity, new IAM action)
- Breaking down a four-surface change (GraphQL ↔ MCP ↔ CLI ↔ n8n)
- Planning a config field addition (which touches 11 files — `shared.md` § 4)
- Planning a database migration that will be consumed by code

Use this BEFORE the implement skill. Planning catches architectural mistakes
when they are cheapest to fix — before any code is written.

---

## Phase 0 — Pre-planning gate

### 1. Classify the task type

| Type | Signals | Planning focus |
| --- | --- | --- |
| **New feature** | "add", "create", "build", "new" | Entry point per stack, data flow, four-surface coverage |
| **Refactor** | "refactor", "extract", "move", "rename", "split" | Migration path, contract compat across stacks |
| **Bug fix** | "fix", "broken", "doesn't work", "regression" | Which stack owns the root cause; minimal fix |
| **Migration** | "upgrade", "migrate", "replace", "switch to" | Rollback strategy, incremental steps, parity |

### 2. Explore before designing

Do your homework:

- **Grep the affected modules** in each stack. Read the canonical example
  for each module before proposing changes.
- **Check cross-stack contracts.** Read the relevant `*.graphql` schema
  and the Relay fragment that consumes it.
- **Check `pkg/server/api/mcp/v1/specification.yaml`** if the task touches
  a backend operation — MCP is easy to forget.
- **Check `packages/n8n-node/nodes/Probo/actions/`** for the n8n surface.
- **Check recent commits** in the affected paths. The PR-mining sample in
  `shared.md` § 13 lists 19 review-enforced rules.

### 3. Ask targeted clarifying questions

Use `AskUserQuestion` only for ambiguity you can't resolve from code:

- **Scope of surfaces** — does this change need all four surfaces or only
  some? (e.g. a connect API change may not need n8n)
- **Acceptance criteria** — what does "done" look like?
- **Migration strategy** — for refactors, can we change all surfaces in
  one PR, or do we need a deprecated phase?
- **Cross-stack ownership** — should both Console and Trust apps consume
  this, or only Console?

Skip questions when the requirement is already specific.

---

## Phase 1 — Design the plan

### 1. Restate the requirement

Write a clear summary with explicit acceptance criteria.

### 2. Identify stacks and modules

| Stack | Modules likely affected |
| --- | --- |
| Go backend | which of `pkg-coredata`, `pkg-probo`, `pkg-iam`, `pkg-server/api/{console,trust,connect,mcp,cookiebanner}/v1`, `pkg-cmd`, `pkg-validator`, `pkg-webhook`, `pkg-mailer`, `e2e` |
| TS frontend | which of `apps-console`, `apps-trust`, `packages-ui`, `packages-relay`, `packages-helpers`, `packages-hooks`, `packages-n8n-node`, `packages-emails` |

### 3. Determine execution order

| Task type | Order | Reasoning |
| --- | --- | --- |
| New backend operation + console page | Go backend → TS frontend | Frontend consumes the API |
| Form + backend validator | Go backend → TS frontend | Validators define constraints |
| New email | TS (`packages/emails`) → Go (`pkg/mailer` consumes via `go:embed`) | Templates must be built first |
| Schema migration + code change | Go backend (migration step before code step) | DB invariant before consumer |
| Independent | Parallel | No dependency |
| Trust portal feature | Go (`pkg/trust` + `pkg/server/api/trust/v1`) → `apps/trust` | Same direction |

### 4. Reference stack-specific patterns

For **Go backend** work:
- `.claude/guidelines/go-backend/patterns.md` — Service/TenantService, Request+Validate, worker pattern, authorization, SQL composition, GraphQL/MCP/CLI shapes, outbox
- Canonical: `pkg/probo/vendor_service.go`, `pkg/server/api/console/v1/vendor_resolvers.go`, `pkg/coredata/cookie_banner.go`

For **TS frontend** work:
- `.claude/guidelines/typescript-frontend/patterns.md` — `*PageLoader` shape, Relay data flow, `@probo/ui` compound components, forms, n8n feature slices
- Canonical: `apps/console/src/pages/organizations/findings/FindingsPage.tsx`, `FindingsPageLoader.tsx`, `apps/console/src/environments.ts`

### 5. Identify cross-stack integration points

For every cross-stack change, document:

- **GraphQL contract** — operation name, variables, response shape, error
  codes. Remember: fields whose resolvers can fail **must NOT be `!`** —
  use Relay `@required` on the consumer side (PR #720).
- **GID flow** — entity types crossing the boundary use base64url; new
  entity types require `pkg/coredata/entity_type_reg.go` updates.
- **n8n action shape** — exported action name MUST equal the operation
  value string.

### 6. Design the approach (by task type)

#### New feature
1. Identify the entry point in each stack
2. Trace the data flow: GraphQL/MCP/CLI/n8n entry → resolver/handler →
   `pkg/probo` service → `pkg/coredata` → Postgres; for the frontend,
   `*PageLoader` → query → fragment → mutation
3. Define the four-surface API contract (GraphQL + MCP + CLI + n8n)
4. Plan e2e tests in `e2e/console/` and `e2e/mcp/`
5. Plan Storybook stories for new `@probo/ui` components

#### Refactor
1. Grep all usages across both stacks
2. Plan the migration path — can the contract coexist with the old one?
3. For GraphQL renames: plan a deprecated alias before removal
4. Update plan: contract → consumers → remove old contract

#### Bug fix
1. Determine which stack owns the root cause (not just where the symptom
   appears — a frontend symptom may have a backend root cause)
2. Plan the minimal fix in the owning stack
3. Plan a regression test (e2e for cross-stack issues, unit for in-stack)

#### Migration
1. Define feature parity across stacks
2. Plan rollback for each stack
3. Plan SQL migration with random-time portion (Probo migrations use
   date + random 6-digit time, not wall clock — see project memory)
4. Plan coexistence period (old + new schema column or GraphQL alias)

### 7. Check pitfalls per stack

Cross-stack pitfalls (always check):
- **All SQL in `pkg/coredata`** (`shared.md` § 13 #1) — never inline raw SQL in `pkg/probo`, workers, or handlers
- **Wrap errors with `cannot <verb> <noun>: %w`** (`shared.md` § 13 #2)
- **GraphQL fields whose resolvers can fail must NOT be `!`** (`shared.md` § 13 #4)
- **Frontend uses Relay-generated types** — never declare local TS types (`shared.md` § 13 #6)
- **Use `pkg/baseurl` for URL construction in Go**; use `new URL(...)` in TS (`shared.md` § 12)
- **`http.DefaultClient` is forbidden** — always `kit/httpclient.WithSSRFProtection()`
- **Never log PII** — entity GIDs only
- **OAuth/PKCE codes must be cleaned up on failure** (PR #957)
- **API keys / signing keys configured as arrays** to support rotation (PR #957)
- **Mutations should update the Relay store**, not refetch (PR #1000)

Go-specific pitfalls — see `.claude/guidelines/go-backend/pitfalls.md`:
- `pkg/coredata/agent_run.go:472` hardcoded SQL `'PENDING'` (drift)
- `pkg/iam/policy` — `In()`/`NotIn()` builders documented but missing
- `pkg/iam/oidc` — provider `error_description` logged verbatim (drift)
- `pkg/agent/tools/search` — bare `http.Client` (SSRF gap)

TS-specific pitfalls — see `.claude/guidelines/typescript-frontend/pitfalls.md`:
- Forgetting `*PageLoader` provider (`CoreRelayProvider` / `IAMRelayProvider`)
- Crossing the core/iam Relay environment boundary silently fails codegen
- Inline SVGs forbidden — extract as React component or use Phosphor icons
- `commit*` is not a good name for a mutation handler — use the action verb

---

## Phase 2 — Produce the plan

### File structure mapping

Before defining steps, map every file that will be created or modified.
This locks in decomposition decisions before writing steps.

For each file:
- **Path** — verified with Glob (never guessed)
- **Action** — create, modify, or delete
- **Responsibility** — one clear purpose
- **Based on** — canonical example it follows

Follow codebase conventions for file organization. Files that change
together should live together. Split by responsibility, not by layer.

### Step granularity

Each step must be a **single, concrete action** completable in 2-5 minutes.

**Bad step:** "Implement the service layer"
**Good step:** "Create `pkg/probo/finding_service.go` with `Create` method
following the Request+Validate pattern at
`pkg/probo/vendor_service.go:83-145`. Include `pg.WithTx` block that calls
`webhook.InsertData` for the `finding.created` event."

Each step must include:
- **Exact file path** (verified with Glob/Grep)
- **What to do** (create, modify specific lines, delete, wire up)
- **Code** — actual code or detailed pseudo-code. Show file contents for
  new files, before/after for modifications. Never write "follow pattern X"
  without showing the resulting code.
- **Verification** (exact command and expected output, e.g.
  `go build ./...`, `make lint`, `make test MODULE=./pkg/probo`,
  `npx relay-compiler`, `npx n8n-node lint`)

### Plan output format

```
# Plan: {feature name}

> Implement with `/potion-implement`. Track progress with TodoWrite.

**Goal:** {one sentence: what this achieves}
**Type:** {Feature | Refactor | Bug fix | Migration}
**Tech:** {key libraries/frameworks: gqlgen, Relay, pgx, etc.}

### Summary
{2-3 sentences: what and why}

### Acceptance criteria
- [ ] {Criterion 1 — specific, testable}
- [ ] {Criterion 2}

### Stacks involved
| Stack | Role | Why needed |
|-------|------|-----------|
| Go backend | upstream / downstream / sole | … |
| TS frontend | upstream / downstream / sole | … |

### Four-surface coverage (for backend operations)
- [ ] GraphQL — schema + resolver + `go generate ./pkg/server/api/<api>/v1`
- [ ] MCP — `specification.yaml` + `go generate ./pkg/server/api/mcp/v1` + resolver body + `pkg/server/api/mcp/v1/types/<entity>.go`
- [ ] CLI — `pkg/cmd/<resource>/<verb>.go`
- [ ] n8n — `packages/n8n-node/nodes/Probo/actions/<resource>/<op>.ts` + register in `actions/index.ts` + `Probo.node.ts`

### Execution order
{Stack A first because data flow direction: …}

## Go backend (Go 1.26)

### File structure
| File | Action | Responsibility | Based on |
|------|--------|---------------|----------|

### Delivery stages

#### Foundation
{Minimum viable slice for this stack — usually: migration + coredata entity + service skeleton}

1. **{Step name}**
   - File: `{exact path}`
   - Action: {create | modify lines N-M}
   - Code:
     ```go
     {actual code or detailed pseudo-code}
     ```
   - Verify: `go build ./pkg/...` → expect no errors

#### Core
{Complete happy path — resolvers, MCP tool, CLI command, validation, IAM action}

#### Hardening
{Edge cases, error switch defaults, e2e tests, RBAC matrix tests}

### Testing
- Go unit tests: `make test MODULE=./pkg/probo` (testify + parallel + black-box `*_test` package)
- E2E: `e2e/console/<resource>_test.go`, `e2e/mcp/<resource>_test.go` — factory builders + RBAC matrix + tenant isolation
- Run: `make test` (or `make test-e2e` for full e2e)

## TypeScript frontend (TS, Node 24+, npm 11+)

### File structure
| File | Action | Responsibility | Based on |
|------|--------|---------------|----------|

### Delivery stages

#### Foundation
{Minimum viable slice — usually: GraphQL fragment + page skeleton + page loader}

1. **{Step name}**
   - File: `{exact path}`
   - Action: {create | modify}
   - Code:
     ```tsx
     {actual code or detailed pseudo-code}
     ```
   - Verify: `make relay && npm run -w apps/console lint`

#### Core
{Page implementation, mutations, forms, list filtering}

#### Hardening
{Loading skeletons, error boundaries, Storybook stories for new `@probo/ui`}

### Testing
- Vitest: `npm run -w apps/console test` (or `-w packages/ui`, etc.)
- Storybook stories for new `@probo/ui` components
- Manual: `npm run -w apps/console dev`

### Cross-stack integration points
| Contract | Upstream | Downstream | Shape |
|----------|----------|------------|-------|
| {GraphQL operation name} | Go backend | apps/console | {variables + response} |
| {MCP tool name} | Go backend | (n8n + AI agents) | {input + output schema} |

### Dependency graph
- Go Step 1 → Go Step 2
- Go (all) → TS Step 1
- TS Step 2 ∥ TS Step 3 (parallel-safe — no shared file)

### Risks and mitigations
| Risk | Stack | Impact | Mitigation |
|------|-------|--------|------------|
| {Specific risk, not generic} | {Stack} | {What goes wrong} | {How to prevent} |
```

---

## Phase 3 — Verify the plan

### 1. Save as draft

Save to `docs/plans/{YYYY-MM-DD}-{feature-name}.md` (referred to as
`{plan-file}` below).

### 2. Mechanical checks

**Placeholder scan** — Grep the plan for banned phrases:
```
Grep({
  pattern: "TBD|TODO|fill in later|add appropriate|add validation|write tests|similar to step|see docs|handle edge cases|as needed|if applicable",
  path: "{plan-file}",
  "-i": true,
  output_mode: "content"
})
```
Any matches are plan failures. Replace each with concrete content:

| Banned phrase | What to write instead |
|--------------|----------------------|
| "TBD", "TODO", "fill in later" | Actual content, or move to Risks as open question |
| "Add appropriate error handling" | Which error type, how to catch it, what to return — for Go: `cannot <verb>: %w` wrapping; for resolvers: switch with mandatory `default:` → `gqlutils.Internal(ctx)` |
| "Add validation" | Which fields, what `validator.*` calls, what error messages |
| "Write tests for the above" | Exact test file (`e2e/console/<x>_test.go`), test names, key assertions |
| "Similar to step N" | Repeat full details — steps may be read out of order |
| "Handle edge cases" | List each edge case + expected behavior |

**File path verification** — for every file path mentioned in the plan,
verify it exists with Glob. Remove or correct any unresolved path.

**Criteria coverage** — every acceptance criterion must map to ≥ 1
implementation step. Every backend operation in the plan must have all
four surfaces covered (GraphQL + MCP + CLI + n8n).

### 3. Cognitive review

- [ ] **Type consistency** — function names, type names, signatures in
      later steps match earlier definitions. Import paths reference files
      created in prior steps. GraphQL operation names in the TS section
      match the schema names in the Go section.
- [ ] **Dependencies** — steps ordered so inputs exist when needed.
      Codegen run between schema edits and consumer code. Migrations
      committed before code that depends on them.
- [ ] **Scope** — plan solves the requirement, no more, no less. No
      "while we're at it" additions.
- [ ] **Step completeness** — every step has file path, action, code
      block, verification command. File structure table accounts for
      every file mentioned in steps.
- [ ] **Cross-stack coherence** — frontend operation names match GraphQL
      schema names, MCP tool input shapes match resolver expectations,
      n8n action exports match operation strings.
- [ ] **Four-surface check** — for any backend operation change, all four
      surfaces have steps.
- [ ] **No drift introduced** — plan does not add to known active drift
      (e.g. don't add new hardcoded SQL outside `pkg/coredata`).

### 4. Parallel review agents (non-trivial plans only)

Dispatch parallel review agents if the plan meets ANY of these:
- Touches 3+ modules
- Has 10+ implementation steps
- Crosses both stacks
- Adds a new entity type (registry update + 4-surface impact)
- Touches IAM (security-critical)

Launch 3 agents in parallel, each reading the saved plan file. Each rates
findings 0-100 confidence and reports only ≥ 80.

**Agent 1 — Completeness:**
> Review the plan at `{plan-file}` for gaps.
> Check: every acceptance criterion has matching steps; four-surface
> coverage is complete for any backend operation; e2e tests exist for
> new GraphQL/MCP endpoints; new `@probo/ui` components have Storybook
> stories. Read the project guidelines at `.claude/guidelines/shared.md`
> and the relevant stack guidelines for context.
> Rate each finding 0-100 confidence. Report only ≥ 80.

**Agent 2 — Consistency:**
> Review the plan at `{plan-file}` for internal consistency.
> Check: GraphQL operation names match between Go and TS sections; MCP
> tool input shapes match resolver code; n8n exported action names
> match operation strings; codegen commands run between schema edits and
> consumer steps; migration steps come before code steps that depend on
> them.
> Rate each finding 0-100 confidence. Report only ≥ 80.

**Agent 3 — Feasibility:**
> Review the plan at `{plan-file}` against the actual codebase.
> Check: referenced canonical examples exist and support the plan
> (`pkg/probo/vendor_service.go`, `apps/console/src/pages/.../FindingsPage.tsx`,
> etc.). Verification commands are realistic. Step granularity is
> appropriate (no 30-minute mega-steps). The plan respects known active
> drift listed in `shared.md` § 14.
> Rate each finding 0-100 confidence. Report only ≥ 80.

Skip this step for trivial plans (< 3 modules, < 10 steps, single stack,
no IAM, no new entity type).

### 5. Fix and re-save

Fix all issues found. Re-save the plan.

---

## Phase 4 — Present and hand off

1. **Track** — call TodoWrite with one entry per step:
   ```json
   {
     "todos": [
       { "id": "{feature-name}-1", "task": "Foundation — Step 1: {description}", "status": "pending" },
       { "id": "{feature-name}-2", "task": "Foundation — Step 2: {description}", "status": "pending" }
     ]
   }
   ```
2. **Present** summary highlighting key design decisions and any open
   questions from the Risks section.
3. **Hand off** — offer implementation:
   > Plan saved to `{plan-file}` with {N} steps tracked.
   >
   > Ready to implement? Use `/potion-implement` to start execution.

---

## Key patterns quick reference

### Go backend
- **Service / TenantService** — `Service.WithTenant(tenantID) → *TenantService → *FooService`. Sub-services hold `svc *TenantService` only — never construct a Scoper. Service methods are authorization-free; IAM checks happen in the resolver.
- **Request + Validate** — every mutating method takes a `Request` struct with a `Validate() error` method. `Validate()` is the first line, uses `validator.New() + v.Check(...) + v.Error()`. Update requests use double pointers (`**string`) to distinguish "no change" from "set NULL".
- **Authorization** — resolver first line: `if err := r.authorize(ctx, id, action); err != nil { return nil, err }`. MCP uses `MustAuthorize` (panics on internal error).
- **Worker** — `Claim` (FOR UPDATE SKIP LOCKED, returns `worker.ErrNoTask`), `Process`, `RecoverStale` (5-min default).
- **SQL composition** — `fmt.Sprintf` template + `pgx.StrictNamedArgs` + `maps.Copy` to merge args.
- **Outbox** — `webhook.InsertData(ctx, tx, ...)` inside the same `pg.WithTx` as the entity write.
- **Error switch** — every resolver error path has a mandatory `default:` returning `gqlutils.Internal(ctx)`.
- **Composition root** — `pkg/probod/probod.go` is the only place dependencies are wired.

### TypeScript frontend
- **`*PageLoader` shape** — `CoreRelayProvider` (or `IAMRelayProvider`) wraps; `useQueryLoader` in `useEffect`; show `*PageSkeleton` while `queryRef` is null; `Suspense` wraps `*Page`.
- **Relay data flow** — preloaded query → `usePreloadedQuery` → `useFragment` per row → `useMutation` with `@deleteEdge`/`@appendEdge`/`@prependEdge`. Update the store; do NOT refetch (PR #1000).
- **Two-environment split** — `apps/console/src/pages/iam/**` compiles against `__generated__/iam/`; everything else against `__generated__/core/`. Crossing this boundary silently fails Relay codegen.
- **`@probo/ui` compound components** — flat exports (`*Root`, `*Shell`, `*Skeleton`), `tailwind-variants` in `variants.ts`, skeleton co-located, custom `Slot` for `asChild`.
- **Forms** — `react-hook-form` + Zod resolver; translator-injected helpers from `@probo/helpers`.
- **n8n action** — exported action name MUST equal operation value string; IAM ops use `proboConnectApiRequest`.

## Stack reference

### Go backend
- Modules: `pkg-coredata`, `pkg-gid`, `pkg-iam`, `pkg-probo`, `pkg-server`, `pkg-agent`, `pkg-llm`, `pkg-validator`, `pkg-{accessreview,connector,esign,docgen,cookiebanner,trust,filemanager,filevalidation,bootstrap,probod,probodconfig,cmd,cli,page,certmanager,crypto}`, `cmd`, `e2e`, `internal`, `pkg-net-infra`, `pkg-{mail,mailer,mailman,slack,webhook}`
- Patterns: `.claude/guidelines/go-backend/patterns.md`
- Testing: `.claude/guidelines/go-backend/testing.md`
- Canonical example: `pkg/probo/vendor_service.go`

### TS frontend
- Modules: `apps-console`, `apps-trust`, `packages-ui`, `packages-relay`, `packages-routes`, `packages-helpers`, `packages-hooks`, `packages-i18n`, `packages-emails`, `packages-n8n-node`, `packages-cookie-banner`, `packages-prosemirror`, `packages-coredata`, `packages-vendors`, `packages-react-lazy`
- Patterns: `.claude/guidelines/typescript-frontend/patterns.md`
- Testing: `.claude/guidelines/typescript-frontend/testing.md`
- Canonical example: `apps/console/src/pages/organizations/findings/FindingsPage.tsx`

## Canonical examples (cite these in plans)

- `pkg/coredata/cookie_banner.go` — full coredata entity pattern
- `pkg/probo/vendor_service.go` — Request+Validate + tx + outbox
- `pkg/probo/evidence_description_worker.go` — worker pattern
- `pkg/server/api/console/v1/vendor_resolvers.go` — resolver shape
- `pkg/server/api/mcp/v1/specification.yaml` — MCP source of truth
- `pkg/connector/oauth2.go` — OAuth2 with HMAC stateless state token
- `pkg/probod/probod.go` — composition root
- `apps/console/src/pages/organizations/findings/FindingsPage.tsx` — current-pattern page
- `apps/console/src/pages/organizations/findings/FindingsPageLoader.tsx` — `*PageLoader` shape
- `apps/console/src/environments.ts` — Relay environments
- `packages/ui/src/atoms/Button/` — `@probo/ui` shape
- `e2e/console/<entity>_test.go` — e2e pattern with factory builders + RBAC matrix

## Rules

- Never guess file paths. Glob/Grep to verify they exist.
- Each stack section references that stack's actual patterns from the
  guidelines.
- Cross-stack integration points must be explicit (GraphQL operation
  shape, MCP tool schema, n8n action export name).
- Execution order must be justified by data flow direction.
- For complex plans (3+ modules in a single stack OR cross-stack), consider
  delegating to the `potion-planner` agent for a focused planning session.
- Plans should be implementable by someone who only reads the plan and
  the guidelines — no assumed tribal knowledge.
- Every risk must have a mitigation. "Query may be slow" is not a risk —
  "Query against `findings` may exceed 500ms when org has > 100k rows;
  mitigate by adding the index in the migration step" is.
- For backend operations, the four-surface checklist is mandatory.
- For config field changes, the 11-file checklist is mandatory.

---
name: potion-plan
description: >
  Plans feature implementations, refactors, and architectural changes in
  Probo across both Go backend and TypeScript frontend stacks. Identifies
  which stacks are involved, determines execution order based on data flow,
  and creates stack-labeled implementation sections. Use when someone asks
  to "plan", "design", "break down", "spec out", "architect", or "how
  should I implement" something. Also triggers for tickets, user stories,
  feature requests, or specs that need an implementation approach. Even
  "what files would I need to change for X" or "what is the best approach
  for X" should activate this skill.
allowed-tools: Read, Write, Glob, Grep, AskUserQuestion, Agent, TodoWrite
model: opus
effort: high
---

# Probo -- Multi-Stack Implementation Planning

Before planning, load shared conventions and each stack's architecture:
- `.claude/guidelines/shared.md` for cross-cutting conventions
- `.claude/guidelines/go-backend/index.md` for Go Backend architecture
- `.claude/guidelines/typescript-frontend/index.md` for TypeScript Frontend architecture

## When to use this skill

- Planning a new feature that may span stacks
- Designing an architecture change
- Breaking down a large task into stack-aware steps

Use this BEFORE the implement skill. Planning catches architectural mistakes
when they are cheapest to fix -- before any code is written.

---

## Phase 0: Pre-planning -- Gather context

Before designing anything, understand the requirement deeply. Skipping this
phase is the number one cause of plans that miss the mark.

### 1. Classify the task type

Determine which kind of change this is -- it shapes the entire planning approach:

| Type | Signals | Planning focus |
|------|---------|---------------|
| **New feature** | "add", "create", "build", "new" | Entry point, data flow, stacks involved, API contracts |
| **Refactor** | "refactor", "extract", "move", "rename", "split" | Migration path, backward compat, cross-stack contracts |
| **Bug fix** | "fix", "broken", "does not work", "regression" | Root cause vs. symptoms, which stack owns the bug |
| **Migration** | "upgrade", "migrate", "replace", "switch to" | Rollback strategy, incremental steps, feature parity |

### 2. Explore the codebase

Before asking questions, do your homework:

- **Read relevant code** in each potentially affected stack. Grep for related
  functionality. Understand what exists before proposing what to build.
- **Check cross-stack contracts.** The GraphQL schema files are the primary
  contract between Go backend and TypeScript frontend. Read the relevant
  `.graphql` file in `pkg/server/api/*/v1/`.
- **Check recent changes.** Look at recent commits in the affected areas.
- **Identify unknowns.** Note what you could not determine from the code alone.

### 3. Ask targeted clarifying questions

Use `AskUserQuestion` to surface ambiguity. Only ask questions whose answers
you could NOT determine from the code. Focus on:

- **Acceptance criteria** -- What does "done" look like? What behaviors are expected?
- **Scope boundaries** -- What is explicitly out of scope?
- **Constraints** -- Performance requirements? Backward compatibility? Deadlines?
- **Edge cases** -- How should the system behave in non-happy-path scenarios?
- **Prior decisions** -- Has this been attempted before? Any rejected approaches?
- **Stack preferences** -- Should both stacks change, or should one adapt to the other?

**Skip this step** if the requirement is already clear and specific (e.g., a
well-written ticket with acceptance criteria, or a trivial change).

---

## Phase 1: Design the plan

### 1. Restate the requirement

Write a clear, specific summary. This is your contract:
- What is being built or changed?
- What is the expected user-facing behavior?
- What are the acceptance criteria (explicit or gathered in Phase 0)?

### 2. Identify stacks involved

Which stacks are affected by this change? Read each stack's module map and
guidelines to determine whether it participates:

- **Go Backend** (Go 1.26) -- modules: cmd, pkg/server, pkg/probo, pkg/iam, pkg/trust, pkg/coredata, pkg/validator, pkg/gid, pkg/agent, pkg/llm, pkg/cmd, e2e
- **TypeScript Frontend** (React 19 + Relay 19) -- modules: apps/console, apps/trust, packages/ui, packages/relay, packages/helpers, packages/hooks

### 3. Determine execution order

Which stack is upstream (provides data/API) vs downstream (consumes)?
Order implementation so dependencies are built before consumers.

| Task type | Order | Reasoning |
|-----------|-------|-----------|
| New API + frontend page | Go Backend -> TypeScript Frontend | Frontend consumes the GraphQL API |
| Frontend form + backend validation | Go Backend -> TypeScript Frontend | Validation defines constraints |
| Independent changes | Parallel | No dependency |
| Shared type change | Schema -> Go Backend -> TypeScript Frontend | Types flow downstream |
| Database migration + API update | Go Backend -> TypeScript Frontend | Schema change flows up |

### 4. Reference stack-specific patterns

For each affected stack, consult its patterns and conventions:

For Go Backend work: see `.claude/guidelines/go-backend/patterns.md`
For TypeScript Frontend work: see `.claude/guidelines/typescript-frontend/patterns.md`

### 5. Identify cross-stack integration points

- The GraphQL schema file is the contract (e.g., `pkg/server/api/console/v1/schema.graphql`)
- Custom scalars must agree: GID (string), Datetime (string), CursorKey (string)
- GraphQL error codes map to typed exceptions in `@probo/relay`
- Relay compiler reads Go-side `.graphql` files directly via relative path

### 6. Design the approach (by task type)

#### For new features
1. Identify the entry point in each stack
2. Trace the data flow across stack boundaries via the GraphQL schema
3. Define the API contract (new types, mutations, queries in `.graphql`)
4. For Go: plan coredata entity + service + resolver + MCP tool + CLI command + e2e test
5. For TypeScript: plan Relay queries/fragments, page component, Loader, route
6. Plan integration tests that verify cross-stack behavior

#### For refactors
1. Identify all files affected across stacks (Grep for usage)
2. Design the migration path -- can stacks be migrated independently?
3. Define backward compatibility strategy for cross-stack contracts
4. Plan: update schema first, then Go resolvers, then regenerate Relay types, then update TS

#### For bug fixes
1. Determine which stack owns the root cause (not just where the symptom appears)
2. Plan the minimal fix in the owning stack
3. If the fix changes a contract, plan downstream stack updates
4. Plan regression tests (e2e for Go, Storybook for UI components)

#### For migrations
1. Define feature parity across all affected stacks
2. Plan rollback strategy for each stack independently
3. Design incremental migration: one stack at a time when possible
4. Plan for contract coexistence (old and new API versions)

### 7. Check pitfalls per stack

These are real issues found in this codebase -- check each one against your plan:

**Go Backend pitfalls:**
- Using `pgx.NamedArgs` instead of `pgx.StrictNamedArgs` (approval blocker)
- Conditional string building in `SQLFragment()` (approval blocker)
- Error messages starting with "failed to" instead of "cannot" (approval blocker)
- Missing `t.Parallel()` in e2e subtests (approval blocker)
- Using `panic` in GraphQL resolvers (approval blocker -- MCP is the exception)
- Missing node resolver for types implementing Node
- Forgetting to register new entity type in `NewEntityFromID` switch

**TypeScript Frontend pitfalls:**
- Using `withQueryRef` in route definitions (approval blocker -- use Loader component)
- Using `useMutationWithToasts` (deprecated -- use `useMutation` + `useToast`)
- Wrong Relay environment for page area (IAM pages use `iamEnvironment`)
- Forgetting `@appendEdge`/`@deleteEdge` on mutations
- Hardcoding paths without `getPathPrefix()` in apps/trust
- Hand-writing TypeScript interfaces that mirror GraphQL types

---

## Phase 2: Produce the plan

### File structure mapping

Before defining steps, map every file that will be created or modified.
This locks in decomposition decisions before writing steps.

For each file:
- **Path** -- verified with Glob (never guessed)
- **Action** -- create, modify, or delete
- **Responsibility** -- one clear purpose
- **Based on** -- canonical example it follows

Follow codebase conventions for file organization. Files that change
together should live together. Split by responsibility, not by layer.

### Step granularity

Each step must be a **single, concrete action** completable in 2-5 minutes.

**Bad step:** "Implement the service layer"
**Good step:** "Create `pkg/probo/widget_service.go` with the
`CreateWidget` method following the pattern in
`pkg/probo/vendor_service.go:45-80`"

Each step must include:
- **Exact file path** (verified with Glob/Grep -- never guessed)
- **What to do** (create, modify specific lines, delete, wire up)
- **Code** -- actual code or detailed pseudo-code for the change. If the
  step creates a file, show the file contents. If it modifies a file, show
  the before/after or the new code to insert. Never write "follow pattern X"
  without also showing what the resulting code looks like.
- **Verification** (exact command to run and expected output)

### Plan output format

```
# Plan: {feature name}

> Implement with `/potion-implement`. Track progress with TodoWrite.

**Goal:** {one sentence: what this achieves}
**Type:** {Feature | Refactor | Bug fix | Migration}
**Tech:** {key technologies, libraries, or frameworks involved}

### Summary
{2-3 sentences: what this plan achieves and why this approach}

### Acceptance criteria
- [ ] {Criterion 1 -- specific, testable}
- [ ] {Criterion 2}

### Stacks involved
| Stack | Role | Why needed |
|-------|------|-----------|

### Execution order
{Which stack goes first and why -- justified by data flow direction}

## Go Backend (Go 1.26)

### File structure
| File | Action | Responsibility | Based on |
|------|--------|---------------|----------|

### Delivery stages

Group steps into stages. Each stage delivers working, testable software.
Small changes within this stack may use a single stage.

#### Foundation
{Minimum viable slice for this stack.}

1. **{Step name}**
   - File: `{exact path}`
   - Action: {create | modify lines N-M | wire up in X}
   - Code:
     ```go
     {actual code or detailed pseudo-code}
     ```
   - Verify: `{command}` -> expect `{output}`

#### Core
{Complete happy path for this stack.}

#### Hardening (if needed)
{Edge cases, error handling, validation.}

### Testing
- {Exact test file and test names}
- Run: `make test MODULE=./pkg/foo`

## TypeScript Frontend (React 19 + Relay 19)

### File structure
| File | Action | Responsibility | Based on |
|------|--------|---------------|----------|

### Delivery stages

#### Foundation
{Minimum viable slice for this stack.}

1. **{Step name}**
   - File: `{exact path}`
   - Action: {create | modify | wire up}
   - Code:
     ```tsx
     {actual code or detailed pseudo-code}
     ```
   - Verify: `{command}` -> expect `{output}`

#### Core
{Complete happy path for this stack.}

### Testing
- {Storybook story file and name, or vitest file}
- Run: `cd packages/ui && npm run storybook` or `cd packages/helpers && npx vitest run`

### Cross-stack integration points
| Contract | Upstream | Downstream | Shape |
|----------|----------|------------|-------|
| {Endpoint/type} | {Stack} | {Stack} | {Request/response/type definition} |

### Dependency graph
- {Go Backend} Step 1 -> {Go Backend} Step 2
- {Go Backend} completes -> {TypeScript Frontend} begins (needs GraphQL schema from Go)
- {TypeScript Frontend} Step 2 || {TypeScript Frontend} Step 3 (parallel-safe)

### Risks and mitigations
| Risk | Stack | Impact | Mitigation |
|------|-------|--------|------------|
| {Specific risk} | {Which stack} | {What goes wrong} | {How to prevent/recover} |
```

---

## Phase 3: Verify the plan

Save the plan as a draft, then verify it -- tools first for mechanical
checks, then judgment for what tools cannot catch. Non-trivial plans get
parallel review agents for fresh eyes.

### 1. Save as draft

Save to `docs/plans/{YYYY-MM-DD}-{feature-name}.md` (referred to as
`{plan-file}` below). This makes the plan available for tool-assisted
verification in the next steps.

### 2. Mechanical checks

Run these tool-assisted checks on the saved draft. Fix any failures
before proceeding to cognitive review.

**Placeholder scan** -- Grep the plan for banned phrases:
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
| "TBD", "TODO", "fill in later" | The actual content, or add to Risks as an open question |
| "Add appropriate error handling" | Which error type, how to catch it, what to return |
| "Add validation" | Which fields, what constraints, what error messages |
| "Write tests for the above" | Exact test file, test names, and key assertions |
| "Similar to step N" | Repeat the full details -- steps may be read out of order |
| "Handle edge cases" | List each edge case and its expected behavior |

**File path verification** -- for every file path mentioned in the plan,
verify it exists:
```
Glob({ pattern: "{exact_path}" })
```
Remove or correct any path that does not resolve.

**Criteria coverage** -- read the acceptance criteria and verify each one
maps to at least one implementation step. List any uncovered criteria and
add steps for them.

### 3. Cognitive review

These checks require judgment -- re-read the plan and verify:

- [ ] **Type consistency** -- function names, type names, and method
      signatures used in later steps match earlier definitions (e.g.,
      `createWidget` in step 3 is not called `buildWidget` in step 7).
      Import paths reference files actually created in prior steps.
- [ ] **Dependencies** -- steps are ordered so each step's inputs exist
      when it runs. Parallel-safe steps are explicitly identified.
      No circular dependencies.
- [ ] **Scope** -- plan solves the stated requirement, no more, no less.
      No speculative features or "while we are at it" additions.
      If > 5 modules touched, splitting has been considered and justified.
- [ ] **Step completeness** -- every step has: file path, action, code
      block, verification command. File structure table accounts for every
      file mentioned in steps. Testing plan covers all new behavior.

### Cross-stack coherence
- [ ] API contracts match between upstream and downstream steps
- [ ] GraphQL types are defined before any stack references them
- [ ] Execution order is justified by data flow direction
- [ ] No orphaned references (e.g., frontend calling an API not in the plan)
- [ ] Three-interface rule: if adding a feature, GraphQL + MCP + CLI all planned

### 4. Parallel review agents (non-trivial plans only)

Dispatch parallel review agents if the plan meets **any** of these:
- Touches 3+ modules
- Has 10+ implementation steps
- Involves cross-cutting architectural changes

Launch 3 agents in parallel, each reading the saved plan file. Each agent
rates findings on a 0-100 confidence scale and reports only issues >= 80.

**Agent 1 -- Completeness:**
> Review the plan at `{plan-file}` for gaps.
> Check: does every acceptance criterion have matching steps? Are there
> untested behaviors? Missing error handling paths? Edge cases not
> addressed? Read the project guidelines at `.claude/guidelines/` for
> context on what patterns are expected.
> Rate each finding 0-100 confidence. Report only >= 80.

**Agent 2 -- Consistency:**
> Review the plan at `{plan-file}` for internal consistency.
> Check: do names, types, and signatures match across steps? Are
> dependencies ordered correctly? Do import paths reference files created
> in prior steps? Does the dependency graph have gaps?
> Rate each finding 0-100 confidence. Report only >= 80.

**Agent 3 -- Feasibility:**
> Review the plan at `{plan-file}` against the actual codebase.
> Check: do the referenced canonical examples exist and support the plan?
> Are verification commands realistic? Is step granularity appropriate?
> Read the cited files and verify the patterns match.
> Rate each finding 0-100 confidence. Report only >= 80.

Skip this step for trivial plans (< 3 modules, < 10 steps, no
architectural changes).

### 5. Fix and re-save

Fix all issues found in steps 2-4. Re-save the plan to
`{plan-file}`.

---

## Phase 4: Present and hand off

1. **Track** -- call the TodoWrite tool with one entry per implementation
   step so progress is visible in Claude Code's native task list:
   ```json
   {
     "todos": [
       { "id": "{feature-name}-1", "task": "Foundation -- Step 1: {description}", "status": "pending" },
       { "id": "{feature-name}-2", "task": "Foundation -- Step 2: {description}", "status": "pending" }
     ]
   }
   ```
2. **Present** a summary highlighting key design decisions and any
   remaining open questions from the Risks section.
3. **Hand off** -- offer to start implementation:

   > Plan saved to `{plan-file}` with {N} steps tracked.
   >
   > Ready to implement? Use `/potion-implement` to start execution.

## Key patterns quick reference

**Go Backend:**
- Two-level service tree: `Service` -> `TenantService` with sub-services
- Request struct + `Validate()` with fluent validator
- All SQL in `pkg/coredata` only (no SQL in service or resolver packages)
- `pgx.StrictNamedArgs` always (never `NamedArgs`)
- Error wrapping: `fmt.Errorf("cannot <action>: %w", err)`
- ABAC policies in `pkg/probo/policies.go` and `pkg/iam/iam_policies.go`

**TypeScript Frontend:**
- Relay colocated operations (queries/fragments in component files)
- Loader component pattern (`useQueryLoader` + `useEffect`)
- `tv()` from tailwind-variants for component variants
- `useMutation` + `useToast` for mutations
- Permission fragments: `canX: permission(action: "core:entity:verb")`

## Stack reference

### Go Backend (Go 1.26)
- Modules: cmd, pkg/server, pkg/probo, pkg/iam, pkg/trust, pkg/coredata, pkg/validator, pkg/gid, pkg/cmd, e2e
- Patterns: `.claude/guidelines/go-backend/patterns.md`
- Testing: `.claude/guidelines/go-backend/testing.md`
- Canonical example: `pkg/probo/vendor_service.go`

### TypeScript Frontend (React 19 + Relay 19)
- Modules: apps/console, apps/trust, packages/ui, packages/relay, packages/helpers, packages/hooks
- Patterns: `.claude/guidelines/typescript-frontend/patterns.md`
- Testing: `.claude/guidelines/typescript-frontend/testing.md`
- Canonical example: `apps/console/src/pages/organizations/documents/DocumentsPageLoader.tsx`

## Canonical examples

When suggesting patterns, point to these real files:

- `pkg/coredata/asset.go` -- Complete coredata entity with all standard methods
- `pkg/probo/vendor_service.go` -- Service layer: Request, Validate, WithTx, webhook
- `pkg/server/api/console/v1/v1_resolver.go` -- GraphQL resolver pattern
- `e2e/console/vendor_test.go` -- E2E test with RBAC and tenant isolation
- `apps/console/src/pages/organizations/documents/DocumentsPageLoader.tsx` -- Canonical Loader component
- `packages/ui/src/Atoms/Badge/Badge.tsx` -- UI atom with tv() variants

## Rules

- Never guess file paths. Glob/Grep to verify they exist.
- Each stack section references that stack's actual patterns from the guidelines.
- Cross-stack integration points must be explicit (GraphQL schema shape).
- Execution order must be justified by data flow direction.
- If the requirement is ambiguous after Phase 0, list what still needs clarification.
- For complex plans touching 3+ modules within a single stack, consider
  delegating to the planner agent for a focused planning session.
- Plans should be implementable by someone who only reads the plan
  and the guidelines -- no assumed tribal knowledge.
- Every risk must have a mitigation. "API might be slow" is not a risk --
  "Query may exceed 500ms for tables > 1M rows; mitigate with index on
  `tenant_id`" is.

---
name: potion-review
description: >
  Reviews code for Probo across its Go backend and TypeScript frontend
  stacks. Determines which stack(s) are in the diff, dispatches
  specialized reviewer sub-agents (architecture, pattern, security,
  style, test, duplication) with the right stack guidelines, and
  aggregates findings. Use when someone asks to "review", "check",
  "audit", "look over", or "give feedback on" code, a diff, a PR, or a
  set of files. Supports filtering by domain — "review architecture
  only", "security review", "duplication check". Also triggers for
  "second opinion", "adversarial review", "cross-model review", "have
  Codex check this", or "GPT review" — these activate an additional
  pass through OpenAI Codex via the local Codex MCP server.
allowed-tools: Read, Glob, Grep, Agent
model: opus
effort: high
---

# Probo — Multi-Stack Code Review

## Load guidelines

Before reviewing:
- **Shared conventions:** `.claude/guidelines/shared.md`
- **Go backend overview:** `.claude/guidelines/go-backend/index.md`
- **TS frontend overview:** `.claude/guidelines/typescript-frontend/index.md`

## Stack routing

Map every file in the diff to a stack using paths and module ownership.

### Go backend
- **Modules:** `pkg-coredata`, `pkg-gid`, `pkg-iam`, `pkg-probo`, `pkg-server` (all `api/*` subpackages), `pkg-agent`, `pkg-llm`, `pkg-validator`, `pkg-accessreview`, `pkg-connector`, `pkg-esign`, `pkg-docgen`, `pkg-cookiebanner`, `pkg-trust`, `pkg-{mail,mailer,mailman}`, `pkg-slack`, `pkg-webhook`, `pkg-filemanager`, `pkg-filevalidation`, `pkg-bootstrap`, `pkg-probod`, `pkg-probodconfig`, `pkg-cmd`, `pkg-cli`, `pkg-page`, `pkg-certmanager`, `pkg-crypto`, `pkg-net-infra`, `cmd`, `e2e`, `internal`
- **Paths:** `pkg/`, `cmd/`, `e2e/`, `internal/`
- **Guidelines:** `.claude/guidelines/go-backend/`

### TypeScript frontend
- **Modules:** `apps-console`, `apps-trust`, `packages-ui`, `packages-relay`, `packages-routes`, `packages-helpers`, `packages-hooks`, `packages-i18n`, `packages-emails`, `packages-n8n-node`, `packages-cookie-banner`, `packages-prosemirror`, `packages-coredata`, `packages-vendors`, `packages-react-lazy`, `packages-eslint-config`, `packages-tsconfig`
- **Paths:** `apps/`, `packages/` (TS workspaces)
- **Guidelines:** `.claude/guidelines/typescript-frontend/`

## Available sub-agents

- `potion-architecture-reviewer` — module placement, layer boundaries, dependency direction, public API surface
- `potion-pattern-reviewer` — error handling, data access (Scoper, SQL composition, `pg.WithTx`), DI, Service/TenantService, Request+Validate, Relay mutations, type usage
- `potion-security-reviewer` — authentication, IAM authorize calls, tenant isolation (Scoper, GID), data exposure, injection, secrets handling, type safety in security paths
- `potion-style-reviewer` — naming, formatting, import ordering, grouped declarations, license headers, i18n wrapping, export patterns
- `potion-test-reviewer` — new Go API endpoints have e2e tests, tests follow Probo conventions (parallel, require vs assert, factory builders, RBAC matrix, tenant isolation), TS UI has Storybook stories
- `potion-duplication-reviewer` — near-identical service methods, copy-pasted SQL, duplicated Relay fragments, missed `pkg/validator`, `pkg/baseurl`, `packages/helpers` reuse
- `potion-adversarial-reviewer` — opt-in cross-model second opinion via OpenAI Codex (MCP). Dispatched only on explicit user request.

## Review strategy

Choose the approach based on the size and stack spread of the change.

### Domain filter (user-requested)

If the user asks for a specific domain ("architecture review only",
"security review", "duplication check"), dispatch only the matching
sub-agent(s) and skip the others. Aggregate as usual.

### Small changes (1-3 files, single stack)

Run the review checklist below directly using that stack's guidelines —
no need for sub-agents.

### Medium changes (4-10 files, 1-2 stacks)

Spawn 2-3 relevant topic reviewers based on what the changes touch:

- Backend route/service changes → `potion-pattern-reviewer` + `potion-architecture-reviewer`
- IAM, OAuth2, OIDC, SAML, PKCE, secrets, signing keys → `potion-security-reviewer` + `potion-pattern-reviewer` (these areas are reviewer hot zones — `pkg/iam/oauth2server/` line-by-line, per `shared.md` § 13)
- Frontend page/component changes → `potion-style-reviewer` + `potion-test-reviewer`
- Database migrations → `potion-security-reviewer` + `potion-architecture-reviewer`
- New feature across modules → `potion-architecture-reviewer` + `potion-pattern-reviewer` + `potion-test-reviewer`

### Large changes (10+ files, multiple stacks)

Spawn all six topic reviewers in parallel. For each, pass the stack
context so it knows which guidelines to load:

```
Review these files using the {stack_name} guidelines:
- Architecture: .claude/guidelines/{stack}/index.md
- Patterns: .claude/guidelines/{stack}/patterns.md
- Conventions: .claude/guidelines/{stack}/conventions.md
- Testing: .claude/guidelines/{stack}/testing.md
- Pitfalls: .claude/guidelines/{stack}/pitfalls.md
- Plus shared: .claude/guidelines/shared.md
```

### Adversarial second-opinion (opt-in, any change size)

When the user uses any of the adversarial trigger phrases listed in this
skill's description ("second opinion", "adversarial review",
"cross-model review", "have Codex check this", "GPT review"), additionally
dispatch `potion-adversarial-reviewer` **once with the full diff and the
shared guidelines path** (`.claude/guidelines/shared.md`) — not once per
stack. The goal is fresh cross-cutting critique from a different model,
not stack-specific depth (the stack-aware specialists already provide
that).

The adversarial reviewer forwards the diff to OpenAI Codex via the local
Codex MCP server and returns Codex's findings tagged with
`category: "adversarial"` so they remain attributable in the merged report.

**Default behavior without those phrases: do NOT dispatch the adversarial
reviewer**, even though it is available. Codex calls are slow and
dual-billed — they should only fire when the user explicitly opts in.

If the adversarial reviewer reports a `(setup)` finding, surface it
prominently — it means the end user has not installed the Codex MCP
server locally, and the standard reviewers' findings are the only output
for this run. The fix message includes:
`claude mcp add --scope user --transport stdio codex -- codex mcp-server`
plus authentication via `codex login` (ChatGPT Plus/Pro) or
`OPENAI_API_KEY`.

## Topic reviewer dispatch with stack context

The master reviewer PASSES stack context to each topic reviewer —
reviewers do not detect it themselves.

| Sub-agent | Focus | Per-stack topic file |
| --- | --- | --- |
| `potion-architecture-reviewer` | Module placement, layer boundaries, dependency direction, public API surface | `{stack}/index.md` (Architecture Overview) |
| `potion-pattern-reviewer` | Service/TenantService, Request+Validate, Scoper, SQL composition, `pg.WithTx`, Relay mutations, type usage | `{stack}/patterns.md` |
| `potion-security-reviewer` | Auth, IAM `authorize`, tenant isolation, secrets, SSRF, OAuth/PKCE cleanup, signing-key arrays | `{stack}/pitfalls.md` + `shared.md` § 12 |
| `potion-style-reviewer` | Naming (`New*`), license headers, import ordering, struct-tag minimalism, i18n wrapping, icon source | `{stack}/conventions.md` |
| `potion-test-reviewer` | E2E for new GraphQL/MCP endpoints, parallel + `require`/`assert`, factory builders, RBAC matrix, tenant isolation, Storybook stories | `{stack}/testing.md` |
| `potion-duplication-reviewer` | Re-implementations of `pkg/validator`, `pkg/baseurl`, `pkg/page`, `packages/helpers`, `@probo/ui` primitives | `{stack}/patterns.md` (shared utilities reference) |
| `potion-adversarial-reviewer` | Cross-model second opinion via Codex MCP — failure classes (auth, data loss, rollback, concurrency, dependency, version skew, observability) | `shared.md` (passed once, not per stack) |

Each sub-agent returns findings in JSON format. After all complete:
1. Collect all findings
2. Deduplicate (same `file:line` from multiple agents → keep most specific)
3. Sort by severity (blockers first)
4. Group by stack
5. Present unified report (see "How to report" below)

## Cross-stack review

For changes spanning both stacks, additionally check:

- [ ] **API contract alignment** — does the frontend consume what the
      backend provides? GraphQL operation names, variable types, response
      shapes match.
- [ ] **Relay-generated types** — frontend uses `__generated__/<env>/*`
      types, never declares local types that duplicate GraphQL output
      (`shared.md` § 13 #6, PR #800).
- [ ] **GraphQL nullability** — fields whose resolvers can fail are NOT
      `!`-non-null; consumer uses Relay `@required` (`shared.md` § 13 #4,
      PR #720).
- [ ] **Cross-stack imports** — TS code does not import directly from
      `pkg/` (impossible by build), and Go code does not embed TS source
      (only built artifacts via `go:embed dist/...` for emails).
- [ ] **Migration ordering** — SQL migrations applied before code that
      depends on them; new entity types registered in
      `pkg/coredata/entity_type_reg.go` before service code uses them.
- [ ] **Four-surface coverage** — for any backend operation change, all
      four surfaces (GraphQL + MCP + CLI + n8n) are present in the diff.
      PR #1132 was explicitly blocked for surfaces lagging behind:
      *"Add e2e, mcp, prb surfaces to cookiebanner"*.

## Review checklist (single stack — direct review)

### Architecture & Design
- [ ] New code is in the correct module (`pkg/coredata` for SQL,
      `pkg/probo` for services, `pkg/server/api/...` for resolvers,
      `pkg/iam/policy` for IAM policy code, etc.)
- [ ] Layer boundaries respected — services have no SQL, resolvers have
      no business logic, services have no auth checks (resolvers do)
- [ ] No circular dependencies introduced
- [ ] Public API surface intentional — exports from `packages/ui` go
      through the barrel `index.ts`

### Pattern compliance — Go
- [ ] Service / TenantService shape — sub-services hold `svc *TenantService`,
      methods read `s.svc.scope`/`s.svc.pg`/`s.svc.logger`, never construct
      a Scoper
- [ ] Mutating methods follow Request + Validate — `Validate()` is the
      first line, uses `validator.New()` per call
- [ ] Update requests use `**string` for "no change vs set NULL"
- [ ] All SQL is in `pkg/coredata` — none in `pkg/probo`, workers, or
      handlers (`shared.md` § 13 #1)
- [ ] SQL composition uses `fmt.Sprintf` template + `pgx.StrictNamedArgs`
      + `maps.Copy`, not string concatenation
- [ ] Tenant isolation: every read/write goes through a `Scoper`;
      `coredata.NewNoScope()` is justified in a comment
- [ ] `pg.WithTx` wraps multi-statement writes; `webhook.InsertData` is
      inside the same tx as the entity write
- [ ] Resolvers: first line `r.authorize(ctx, id, action)`; error switch
      has mandatory `default:` returning `gqlutils.Internal(ctx)`
- [ ] MCP resolvers use `MustAuthorize` (panicking variant)
- [ ] Workers: `Claim` uses `FOR UPDATE SKIP LOCKED`, returns
      `worker.ErrNoTask`; `RecoverStale` exists
- [ ] Constructors named `New*`, never `Build*` / `Make*` (`shared.md` § 13 #8)
- [ ] Errors wrapped: `fmt.Errorf("cannot <verb> <noun>: %w", err)`
      (`shared.md` § 13 #2)
- [ ] No `errors.As(err, &ptr)` — use `errors.AsType[T](err)` from kit
- [ ] No raw `http.Client` — use `go.gearno.de/kit/httpclient` with
      `WithSSRFProtection()`
- [ ] No `fmt.Sprintf` for URLs — use `pkg/baseurl` or `net/url`
      (`shared.md` § 13 #7, PR #800)
- [ ] No `http.StatusXxx` numeric literals — use the `http.StatusXxx`
      constants (`shared.md` § 13 #18)
- [ ] No `json` struct tags on internal-only structs (`shared.md` § 13 #9)
- [ ] Webhook payloads use `pkg/webhook/types`, never `coredata` structs
      directly (`shared.md` § 13 #13, PR #720)
- [ ] Switch / case blocks > ~10 cases extracted into private functions
      (`shared.md` § 13 #17)

### Pattern compliance — TS
- [ ] `*PageLoader` mounts the right Relay provider (`CoreRelayProvider`
      or `IAMRelayProvider`); `useQueryLoader` in `useEffect`; skeleton
      shown until `queryRef`; `Suspense` wrapping
- [ ] No crossing of core/iam Relay environment boundary —
      `apps/console/src/pages/iam/**` only consumes `__generated__/iam/`
- [ ] Frontend types come from Relay-generated artifacts; no local types
      duplicate GraphQL output (`shared.md` § 13 #6, PR #800)
- [ ] Mutations use `@deleteEdge` / `@appendEdge` / `@prependEdge` to
      update the Relay store; do NOT refetch when the response carries
      the data (`shared.md` § 13 #10, PR #1000)
- [ ] `usePaginationFragment` uses `@connection(filters: [])` so filter
      changes do not invalidate the connection
- [ ] `@probo/ui` compound components — flat exports (`*Root`, `*Shell`,
      `*Skeleton`), `tailwind-variants` in `variants.ts`, skeleton
      co-located, no import of `*Root` from skeleton
- [ ] No inline SVGs — use a React component or Phosphor icon
      (`shared.md` § 13 #5)
- [ ] Mutation handler names use the action verb, not `commit*`
      (`shared.md` § 13 #15, PR #1073)
- [ ] Reuse `@probo/ui` primitives instead of duplicating in app pages
      (`shared.md` § 13 #16)
- [ ] All user-visible strings wrapped via `useTranslate`
- [ ] No `template literal + URL`; use `new URL(...)` and
      `URLSearchParams`

### Error handling
- [ ] Project error types used (Go: typed sentinel + wrapped; TS: typed
      classes from `@probo/relay`)
- [ ] Errors propagated correctly through layers
- [ ] Boundary errors: GraphQL `gqlutils.Internal(ctx)` catch-all; MCP /
      HTTP `jsonutil.RenderInternalServerError(w)`; never expose
      stack traces, SQL errors, file paths, or provider
      `error_description`
- [ ] OAuth / PKCE: code is cleaned up on failure (PR #957)

### Testing
- [ ] New Go API endpoints have e2e tests (`e2e/console/<x>_test.go` and
      `e2e/mcp/<x>_test.go`) (`shared.md` § 13 #12)
- [ ] Go tests are in `*_test` (black-box) packages, not the package
      under test (`shared.md` § 13 #14, PR #1023)
- [ ] All Go tests call `t.Parallel()`
- [ ] `require` for halting failures, `assert` for accumulating
- [ ] Factory builders + RBAC matrix + tenant isolation assertions in
      e2e
- [ ] Security-sensitive code (`pkg/iam/oauth2server`, OIDC, PKCE,
      ID-token) has 100% unit test coverage (`shared.md` § 13 #11)
- [ ] New `@probo/ui` components have Storybook stories
- [ ] Vitest tests assert behavior, not implementation details

### Types & safety
- [ ] No `any`, `unknown` without narrowing in TS; no untyped Go escape
      hatches
- [ ] Shared types used where they exist (Relay-generated for TS;
      `@probo/coredata` for the one shared enum)
- [ ] New types placed in correct location

### Naming & style
- [ ] Files follow project naming convention (snake_case for Go;
      kebab-case folders + PascalCase components for TS)
- [ ] License header (ISC) on every new source file (`shared.md` § 6) —
      year ranges expanded when editing
- [ ] Free-form commit messages, not Conventional Commits; signed with
      `-s -S`; no `Co-Authored-By` for AI (`shared.md` § 5)

### Observability
- [ ] Go uses `go.gearno.de/kit/log` exclusively, `*Ctx` variants,
      typed field helpers (`log.String`, `log.Int`, `log.Error`,
      `log.Duration`); no `fmt.Sprintf` into messages
- [ ] No PII in logs (no emails, names, IPs, raw bodies, full query
      strings, OAuth `error_description`) — entity GIDs only
      (`shared.md` § 8)

## Severity classification

**Blockers** (must fix before merge):
- Security issues (SSRF gap, missing IAM `authorize`, secrets in code,
  signing keys not rotatable, PII in logs, OAuth code not cleaned up on
  failure)
- Missing error handling (no `default:` in resolver switch, bare `return err`)
- Pattern violations that set a bad precedent (SQL outside `pkg/coredata`,
  local TS types duplicating GraphQL)
- Missing tests for new functionality (especially `pkg/iam/oauth2server`)
- Cross-stack contract mismatches
- Missing surfaces (GraphQL operation added without MCP/CLI/n8n
  counterpart)

**Suggestions** (nice to have):
- Minor naming improvements
- Extra test cases for edge cases
- Documentation improvements
- Performance optimizations
- Storybook story additions

## How to report each finding

```
**[BLOCKER/SUGGESTION]** {file}:{line} — {what's wrong}
  Stack: {go-backend | typescript-frontend | shared}
  Why: {reference to guideline section or PR-mining rule, e.g.
        "shared.md § 13 #1 — All SQL in pkg/coredata (PR #800)"}
  Fix: {specific suggestion, ideally with code, or reference to canonical
        example, e.g. "Move query into a coredata method following
        pkg/coredata/cookie_banner.go:LoadByCategory"}
```

## Aggregation

After all topic reviewers return their findings:

1. **Collect** findings from every reviewer
2. **Deduplicate** — same `file:line` reported by multiple reviewers →
   keep the most specific finding (or merge them under one entry citing
   all reviewers that flagged it)
3. **Sort** by severity (blockers first)
4. **Group by stack** — present findings under their stack heading so
   the developer knows which context to load
5. **Adversarial findings** — when present, present them in a separate
   "Adversarial (Codex)" subsection after the standard findings, so the
   user can clearly attribute disagreements
6. **Cross-stack summary** — if the change spans both stacks, add a
   summary section highlighting any cross-stack issues (contract
   mismatches, type inconsistencies, surfaces lagging)

## Common pitfalls to watch for

These are real issues found during codebase analysis (see
`shared.md` § 14 "Known Drift / Active Violations"):

- **`pkg/probo/agent_run.go:472`** — hardcoded SQL literal in service code (drift). When touching this file, migrate the query into a coredata method.
- **`pkg/server/api/csp.go`** — outbound HTTP path lacks `WithSSRFProtection()` (drift).
- **OIDC `error_description` PII leak** — at least one path surfaces raw provider `error_description` to clients/logs.
- **`apps/console/src/routes/`** — legacy `loaderFromQueryLoader` / `withQueryRef` from `@probo/routes` are deprecated; new code uses `*PageLoader`.
- **`contrib/claude/react-components.md`** — older components don't yet match the "props for configuration, data from hooks" rule. Refactor opportunistically when touching a file.

## Reference files

### Go backend
- Canonical implementation: `pkg/probo/vendor_service.go`, `pkg/server/api/console/v1/vendor_resolvers.go`, `pkg/coredata/cookie_banner.go`
- Canonical test: `e2e/console/vendor_test.go`
- Guidelines: `.claude/guidelines/go-backend/`

### TS frontend
- Canonical implementation: `apps/console/src/pages/organizations/findings/FindingsPage.tsx`, `FindingsPageLoader.tsx`
- Canonical environment wiring: `apps/console/src/environments.ts`
- Canonical UI primitive: `packages/ui/src/atoms/Button/`
- Guidelines: `.claude/guidelines/typescript-frontend/`

### Shared
- Shared guidelines: `.claude/guidelines/shared.md`
- Authoritative subsystem docs: `contrib/claude/*.md`

---
name: potion-ask
description: >
  Answers questions about the Probo codebase — a polyglot Go + TypeScript
  monorepo for an open-source compliance platform. Use this skill when
  someone asks "where is...", "how does X work", "why is Y done this way",
  "what pattern does Probo use for Z", "explain the architecture", "find
  the code that handles...", "what does this module do", or any onboarding
  question ("how do I get started", "walk me through the codebase"). Loads
  guidelines lazily — shared first, then the relevant stack — and delegates
  deep searches to the `potion-explorer` agent. Triggers for both Go
  backend questions (gqlgen, pgx, chi, Service/TenantService, IAM, MCP,
  CLI, workers) and TypeScript frontend questions (React, Relay, Vite,
  tailwind-variants, *PageLoader, n8n node).
allowed-tools: Read, Glob, Grep, Agent
model: sonnet
effort: high
---

# Probo — Codebase Q&A

Probo is an **open-source compliance platform** living in a polyglot
monorepo: a single Go daemon (`probod`) serves four GraphQL APIs (console,
trust, connect) plus an MCP API to two React 19 + Relay 19 SPAs
(`apps/console`, `apps/trust`), a CLI (`prb`), and an n8n community node.

## Load guidelines lazily

Don't read everything up front. Pick the right files for the question:

1. **Always start with** `.claude/guidelines/shared.md` — cross-cutting rules
   (four-surface API rule, GIDs, tenant isolation, error handling, security,
   logging, review-enforced standards).
2. **For Go backend questions** (anything in `pkg/`, `cmd/`, `e2e/`):
   read `.claude/guidelines/go-backend/index.md` first, then drill into
   `patterns.md`, `conventions.md`, `pitfalls.md`, or
   `module-notes/{module}.md` for the specific topic.
3. **For TypeScript frontend questions** (anything in `apps/` or
   `packages/`): read `.claude/guidelines/typescript-frontend/index.md`,
   then drill into `patterns.md`, `conventions.md`, `pitfalls.md`, or
   `module-notes/{module}.md`.
4. **For cross-stack questions** (e.g. "how does GIDs work end-to-end",
   "where is the API contract for vendors"): read `shared.md` plus both
   stack `index.md` files.

## Stack routing — figure out which stack first

Use this table to decide which stack(s) own the question. If the question
is ambiguous, load `shared.md` first and ask the user to clarify.

| Signal | Route to |
| --- | --- |
| `pkg/`, `cmd/`, `e2e/`, `internal/` paths | Go backend |
| `apps/console`, `apps/trust`, `packages/*` (TS) | TypeScript frontend |
| Keywords: gqlgen, pgx, chi, cobra, huh, kit/worker, IAM policy, MCP, mcpgen, Scoper, GID, validator, llm, agent | Go backend |
| Keywords: React, Relay, Vite, Vitest, tailwind-variants, Radix, Ariakit, Tiptap, Storybook, *PageLoader, useFragment, useQueryLoader, n8n | TypeScript frontend |
| Module names: `pkg-coredata`, `pkg-iam`, `pkg-probo`, `pkg-server`, `pkg-agent`, `pkg-llm`, etc. | Go backend |
| Module names: `apps-console`, `apps-trust`, `packages-ui`, `packages-relay`, `packages-routes`, `packages-helpers`, `packages-n8n-node`, etc. | TypeScript frontend |
| Four-surface API rule (GraphQL ↔ MCP ↔ CLI ↔ n8n) | Both — backend defines, n8n consumes |
| Config propagation (11 files), GIDs, tenant isolation, license headers, commits | shared.md |

## Module map — Go backend

| Module | Path | Purpose |
| --- | --- | --- |
| pkg-coredata | `pkg/coredata` | Sole owner of SQL — entity files, `Scoper`, filters, 364 migrations |
| pkg-gid | `pkg/gid` | 24-byte tenant-scoped global IDs; `entity_type_reg.go` registry |
| pkg-iam | `pkg/iam` | Policy-as-Go-code authorization; OIDC, SAML, SCIM subdirs |
| pkg-probo | `pkg/probo` | Domain services (`Service` → `TenantService` → `*FooService`); workers |
| pkg-server | `pkg/server` | chi router; `api/{console,trust,connect}/v1` (gqlgen) + `api/mcp/v1` (mcpgen) + `api/cookiebanner` |
| pkg-agent | `pkg/agent` | LLM agent orchestration framework (FunctionTool, Handoff, Checkpointer) |
| pkg-llm | `pkg/llm` | Provider-agnostic LLM client (Anthropic, OpenAI, Bedrock); OTel tracing |
| pkg-validator | `pkg/validator` | Fluent validation (`v.Check(...)`, `v.Error()`) |
| pkg-accessreview | `pkg/accessreview` | Access-review campaigns; pluggable drivers |
| pkg-connector | `pkg/connector` | OAuth2/API-key 3rd-party connector framework |
| pkg-esign | `pkg/esign` | E-signature certificate workers |
| pkg-docgen | `pkg/docgen` | HTML→PDF rendering (chromedp + Mermaid) |
| pkg-cookiebanner | `pkg/cookiebanner` | Cookie banner domain logic |
| pkg-trust | `pkg/trust` | Trust center service layer (public, unauthenticated) |
| pkg-{mail,mailer,mailman} | `pkg/mail*` | Outbound email outbox |
| pkg-slack, pkg-webhook | `pkg/slack`, `pkg/webhook` | Outbound Slack + webhook delivery |
| pkg-filemanager, pkg-filevalidation | `pkg/filemanager`, `pkg/filevalidation` | S3/SeaweedFS storage |
| pkg-cli, pkg-cmd | `pkg/cli`, `pkg/cmd` | `prb` CLI (cobra + huh prompts) |
| pkg-probod | `pkg/probod` | Composition root; `Run()` + graceful shutdown |
| pkg-probodconfig | `pkg/probodconfig` | Daemon config struct (one file per subsystem) |
| pkg-bootstrap | `pkg/bootstrap` | Env-vars → `probodconfig.FullConfig` YAML generator |
| pkg-certmanager | `pkg/certmanager` | ACME custom-domain TLS |
| pkg-crypto | `pkg/crypto` | AES-256-GCM, PBKDF2, SHA-256 primitives |
| pkg-page | `pkg/page` | Cursor pagination types |
| e2e | `e2e/` | E2E suite — `e2e/console/` (43 files), `e2e/mcp/` (22 files), factory builders |

## Module map — TypeScript frontend

| Module | Path | Purpose |
| --- | --- | --- |
| apps-console | `apps/console` | Compliance SPA (437 TS/TSX) — two Relay envs (core/iam), `*PageLoader` pattern |
| apps-trust | `apps/trust` | Public trust portal (50 files) — magic-link / OIDC / NDA |
| packages-ui (`@probo/ui`) | `packages/ui` | Component library (285 files); `tailwind-variants`, compound exports |
| packages-relay (`@probo/relay`) | `packages/relay` | `makeFetchQuery` + 6 typed error classes |
| packages-routes (`@probo/routes`) | `packages/routes` | `AppRoute` type; legacy `loaderFromQueryLoader` (deprecated) |
| packages-helpers (`@probo/helpers`) | `packages/helpers` | `formatDate`, `formatError`, `sprintf`, `faviconUrl` — translator-injected |
| packages-hooks (`@probo/hooks`) | `packages/hooks` | 9 hooks: `usePageTitle`, `useFavicon`, `useToggle`, … |
| packages-i18n (`@probo/i18n`) | `packages/i18n` | Custom zero-dep i18n — currently dormant (`"en"` hardcoded) |
| packages-emails (`@probo/emails`) | `packages/emails` | 14 React Email templates → `dist/*.html.tmpl` (consumed by `pkg/mailer` via `go:embed`) |
| packages-n8n-node (`@probo/n8n-nodes-probo`) | `packages/n8n-node` | n8n community node (236 files); resource × operation feature slices |
| packages-cookie-banner | `packages/cookie-banner` | Vanilla web component; dual ESM + IIFE |
| packages-prosemirror | `packages/prosemirror` | Markdown ↔ ProseMirror node trees |
| packages-coredata | `packages/coredata` | Shared `TrustCenterDocumentAccessStatus` enum |
| packages-vendors, packages-react-lazy | `packages/{vendors,react-lazy}` | Static vendor data; lazy-load helper |

## Canonical examples (cite these)

| File | Demonstrates |
| --- | --- |
| `pkg/coredata/cookie_banner.go` | Full entity pattern: struct, `CursorKey`, `AuthorizationAttributes`, scope+filter+cursor query, `Insert` with `scope.GetTenantID()`, FOR UPDATE SKIP LOCKED |
| `pkg/probo/vendor_service.go` | Request + Validate, `pg.WithTx` with `webhook.InsertData` inside the same tx, double-pointer optional fields |
| `pkg/probo/evidence_description_worker.go` | Worker pattern: `Claim` (FOR UPDATE SKIP LOCKED), `Process`, `RecoverStale` |
| `pkg/server/api/console/v1/vendor_resolvers.go` | Resolver shape: `r.authorize(ctx, id, action)` first, error switch with mandatory `default:` → `gqlutils.Internal(ctx)` |
| `pkg/server/api/mcp/v1/specification.yaml` | Single source of truth for MCP tools (mcpgen) |
| `pkg/probod/probod.go` | Composition root: `wg.Go`, `cancel(fmt.Errorf(...))`, ordered `stop*()`, `pgClient.Close()` last |
| `pkg/connector/oauth2.go` | OAuth2 with HMAC-signed stateless `state`, three `TokenEndpointAuth` modes, SSRF-protected transport |
| `apps/console/src/pages/organizations/findings/FindingsPage.tsx` | The full current-pattern page: preloaded query, `usePaginationFragment` with `@connection(filters: [])`, `useFragment` rows, `useMutation` with `@deleteEdge`, `useTransition`, `useToast`, `useTranslate` |
| `apps/console/src/pages/organizations/findings/FindingsPageLoader.tsx` | `*PageLoader` shape: `CoreRelayProvider` → `useQueryLoader` in `useEffect` → `*PageSkeleton` → `Suspense` |
| `apps/console/src/environments.ts` | Two Relay envs (coreEnvironment + iamEnvironment), 1-min query cache, `makeFetchQuery` |
| `packages/ui/src/atoms/Button/` | Modern `@probo/ui` shape: flat exports, `tailwind-variants` in `variants.ts`, co-located skeleton |
| `e2e/console/vendor_test.go` | E2E factory builders + RBAC matrix tests + tenant isolation assertions |

## Answering strategy by question type

**"Where is X?"** → Grep, narrow with the module map, return exact
file:line. Don't say "probably" — find it.

**"How does X work?"** → Identify the entry point (CLI command? GraphQL
query? React route?), trace the data flow through layers, cite each step.
For GraphQL → resolver → service → coredata → Postgres. For Relay →
`*PageLoader` → query → useFragment → server.

**"Why is X done this way?"** → Check `shared.md` § 13 (review-enforced
standards) and the relevant stack's `pitfalls.md` for rationale. Also
look at the doc under `contrib/claude/` (the authoritative source). If no
rationale documented, say so; don't invent one.

**"What pattern for X?"** → Reference the canonical example table above
plus the `patterns.md` section in the right stack. Always cite a file.

**"How do I add a new feature on the backend?"** → Walk through the
four-surface rule (`shared.md` § 3): GraphQL → MCP → CLI → n8n must all
be updated. Reference `pkg/probo/vendor_service.go` for the service shape.

**"How do I get started?"** → Walk: top-level layout → stack summary →
canonical examples → `make stack-up` + `make dev-config`.

## Key cross-cutting rules to surface

When the question even tangentially touches one of these, mention them:

- **Four-surface API rule** (`shared.md` § 3): every backend operation lives on GraphQL + MCP + CLI + n8n. PR #1132 explicitly blocked surfaces lagging behind GraphQL.
- **All SQL in `pkg/coredata`** (`shared.md` § 13 #1): inline SQL outside `pkg/coredata` is a review blocker.
- **Wrap errors with `cannot ...: %w`** (`shared.md` § 13 #2; `pkg/...` Go style).
- **GraphQL fields whose resolvers can fail must NOT be `!` non-null** — use Relay `@required` (`shared.md` § 13 #4).
- **Frontend uses Relay-generated types** — never declare local TS types that duplicate GraphQL output (`shared.md` § 13 #6).
- **Tenant isolation** is enforced at the data layer via `coredata.Scoper`, never on the frontend (`shared.md` § 9).
- **License headers** (ISC) on every source file (`shared.md` § 6).
- **Commits**: free-form, NOT Conventional Commits, signed with `-s -S`, no `Co-Authored-By` for AI (`shared.md` § 5).

## Delegating deep exploration

For complex tasks (tracing data through 5+ files, mapping all uses of a
pattern across both stacks, full module audits), spawn the
`potion-explorer` agent rather than doing it yourself:

```
Agent: potion-explorer
Prompt: |
  Trace how a vendor record flows from creation in the Console UI
  through to webhook delivery. Cite every file and line.
```

The explorer is read-only and returns a structured report with file
references.

## Rules

- Never guess. If you can't find something, say so and suggest where to
  look (the right `contrib/claude/` doc or stack guideline).
- Cite specific files and line numbers in every answer.
- Prefer code snippets from the actual codebase over abstract descriptions.
- Note migrations or active drift when relevant — e.g.
  `apps/console/src/routes/` is mid-migration to colocated `routes.ts`;
  legacy `loaderFromQueryLoader` is deprecated; `pkg/coredata/agent_run.go:472`
  has hardcoded SQL pending refactor.
- Keep answers focused. Answer the question, then offer to go deeper.
- For complex exploration, delegate to `potion-explorer`.

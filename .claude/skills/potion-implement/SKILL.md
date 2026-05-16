---
name: potion-implement
description: >
  Master implementation orchestrator for Probo. Analyzes incoming tasks,
  determines which language stack(s) are involved (Go backend or
  TypeScript frontend), and delegates to the right stack-specific
  implementer agent. For cross-stack work (e.g. new GraphQL endpoint +
  console page + CLI command + n8n action), orchestrates upstream first
  (Go) then downstream (TS) so the frontend can reference the actual API
  shape. Use when someone asks to "add", "create", "build", "implement",
  "write", or "code" anything in Probo. Triggers on tickets, specs, user
  stories, and feature descriptions even without explicit verbs. Enforces
  the four-surface API rule (GraphQL ↔ MCP ↔ CLI ↔ n8n) for every backend
  change.
allowed-tools: Read, Glob, Grep, Agent
model: opus
effort: high
---

# Probo — Master Implementation Orchestrator

This skill does NOT implement code itself. It analyzes the incoming task,
determines which stack(s) are involved, and delegates to the right
stack-specific implementer agent.

## Load guidelines first

Before analyzing any task:

- **Shared conventions:** `.claude/guidelines/shared.md` (always)
- **Go backend overview:** `.claude/guidelines/go-backend/index.md`
- **TypeScript frontend overview:** `.claude/guidelines/typescript-frontend/index.md`

## Stack routing table

Map modules and file paths to their owning stack.

### Go backend (Go 1.26)
- **Frameworks:** chi/v5, gqlgen, pgx/v5, go.gearno.de/kit, cobra, huh, anthropic-sdk-go, openai-go, aws-sdk-go-v2, OpenTelemetry, testify
- **Modules:** `pkg-coredata`, `pkg-gid`, `pkg-iam`, `pkg-probo`, `pkg-server` (`api/{console,trust,connect}/v1`, `api/mcp/v1`, `api/cookiebanner`), `pkg-agent`, `pkg-llm`, `pkg-validator`, `pkg-accessreview`, `pkg-connector`, `pkg-esign`, `pkg-docgen`, `pkg-cookiebanner`, `pkg-trust`, `pkg-{mail,mailer,mailman}`, `pkg-slack`, `pkg-webhook`, `pkg-filemanager`, `pkg-filevalidation`, `pkg-bootstrap`, `pkg-probod`, `pkg-probodconfig`, `pkg-cmd`, `pkg-cli`, `pkg-page`, `pkg-certmanager`, `pkg-crypto`, `pkg-net-infra`, `cmd`, `e2e`, `internal`
- **Paths:** `pkg/`, `cmd/`, `e2e/`, `internal/`
- **Implementer agent:** `potion-go-backend-implementer`
- **Guidelines:** `.claude/guidelines/go-backend/`

### TypeScript frontend (TS, Node 24+, npm 11+)
- **Frameworks:** React 19, Relay 19, Vite, Vitest, react-router v7, react-hook-form, Zod, tailwind-variants, Radix, Ariakit, Tiptap, Storybook, React Email, turborepo
- **Modules:** `apps-console`, `apps-trust`, `packages-ui`, `packages-relay`, `packages-routes`, `packages-helpers`, `packages-hooks`, `packages-i18n`, `packages-emails`, `packages-n8n-node`, `packages-cookie-banner`, `packages-prosemirror`, `packages-coredata`, `packages-vendors`, `packages-react-lazy`, `packages-eslint-config`, `packages-tsconfig`
- **Paths:** `apps/`, `packages/` (TS workspaces)
- **Implementer agent:** `potion-typescript-frontend-implementer`
- **Guidelines:** `.claude/guidelines/typescript-frontend/`

## Task analysis (before spawning anything)

1. **Read the task carefully.** Feature, bugfix, refactor, migration?
2. **Identify affected modules** by scanning paths, feature names, domain
   concepts.
3. **Map modules to stacks** using the table above.
4. **Apply the four-surface rule.** If the task adds or changes a
   *backend operation* (vendor, control, finding, document, risk, etc.),
   it almost always implies updates on **all four** surfaces:
   GraphQL ↔ MCP ↔ CLI (`prb`) ↔ n8n. Surface this to the user as part
   of the orchestration plan; surfaces lagging behind is a documented
   reviewer block (PR #1132 *"Add e2e, mcp, prb surfaces to cookiebanner"*).
5. **Classify the task:**
   - **Single-stack** — all affected modules belong to one stack.
   - **Cross-stack** — affected modules span two stacks (e.g. backend +
     console page).
   - **Three- or four-surface backend** — Go-only but touches multiple
     API surfaces (GraphQL + MCP + CLI). Still single-stack, but the
     Go implementer must update all four surfaces.

## Single-stack delegation

When only one stack is involved (and only one frontend or only Go):

1. Spawn the `potion-{stack}-implementer` agent with the full task
   description.
2. For Go-only work that touches a backend operation: explicitly remind
   the agent to apply the **four-surface rule** (GraphQL ↔ MCP ↔ CLI ↔
   n8n). Even though n8n lives in the TS workspace, n8n changes for an
   existing resource are still in scope for the Go implementer because
   they're a thin GraphQL shim — only spawn the TS implementer if the
   n8n change requires new TS types or new actions/index.ts wiring.
3. Let the agent run end-to-end.

## Cross-stack orchestration

When both stacks are involved, order matters. Implement upstream first.

### Step-by-step

1. **Determine dependency order** using the rules below.
2. **Spawn the upstream implementer first** with the task scoped to its
   stack. Wait for it to finish.
3. **Read the changes** the upstream agent made (the Bash tool output
   it returned, plus any new/modified files). Extract the contract:
   GraphQL operation names + variables, response shape, error codes,
   field names, GIDs.
4. **Spawn the downstream implementer** with explicit upstream context:
   ```
   The Go implementer added/changed:
   - GraphQL: query/mutation `<name>` in `pkg/server/api/console/v1/graphql/<file>.graphql`
     - Variables: <list>
     - Response: <shape>
   - MCP tool: `<tool>` in `pkg/server/api/mcp/v1/specification.yaml`
   - CLI: `prb <resource> <verb>` in `pkg/cmd/<resource>/<verb>.go`
   Now implement the TypeScript side:
   - Add the page/component/action that consumes this contract.
   - Use Relay-generated types (do NOT declare local types — `shared.md` § 13 #6).
   - For mutations, update the Relay store directly when the response carries the data — do NOT refetch.
   ```
5. **Verify coherence.** After both finish:
   - Frontend operations match the GraphQL schema names + variables.
   - n8n action is registered in BOTH `actions/index.ts` AND
     `Probo.node.ts` (per `shared.md` § 3 step 4).
   - Migration ordering: SQL migration committed before code that depends
     on it.
   - License headers on every new file.

### Dependency direction rules

| Task type | Order | Reasoning |
| --- | --- | --- |
| New GraphQL operation + console page | Go backend → TS frontend | Frontend consumes the API |
| New form + backend validation | Go backend → TS frontend | Validators define constraints |
| Schema migration + code that uses new column | Go backend (migration first, then code) | DB invariant must exist before code |
| Independent change in only one stack | Single agent | No dependency |
| Refactor that crosses stacks | Go backend → TS frontend | TS adapts to new contract |
| Email template + Go consumer | TS frontend (`packages/emails`) → Go backend (`pkg/mailer` + `go:embed`) | The Go side embeds the rendered HTML — TS must build first; remember `npm run build -w @probo/emails` |
| Trust portal feature (public, unauthenticated) | Go (`pkg/trust` + `pkg/server/api/trust/v1`) → `apps/trust` | Same backend-first rule |

### When the Go side spans 3+ surfaces

If the change adds a new resource or operation, the Go implementer must
do **all four**:

1. **GraphQL** — schema in `pkg/server/api/{console,connect,trust}/v1/graphql/*.graphql`,
   resolvers in the same package, then `go generate ./pkg/server/api/<api>/v1`.
2. **MCP** — declare the tool in `pkg/server/api/mcp/v1/specification.yaml`,
   `go generate ./pkg/server/api/mcp/v1`, write the resolver body, add type
   helpers in `pkg/server/api/mcp/v1/types/*.go` (one file per entity).
   Use `MustAuthorize` (panicking variant — see `contrib/claude/mcp.md`).
3. **CLI (`prb`)** — leaf command file under `pkg/cmd/<resource>/<verb>.go`,
   one GraphQL `const` per leaf, unexported `*Response` struct, `NewCmdVerb(f)`
   constructor.
4. **n8n** — register in `packages/n8n-node/nodes/Probo/actions/index.ts`
   AND in `Probo.node.ts` (properties array). Add per-resource files under
   `actions/<resource>/`. `npx n8n-node lint` must pass. Export name MUST
   equal the operation value string. **This is the only step that crosses
   into the TS implementer's territory** — small additions can stay with
   the Go implementer if you brief it on the n8n shape; bigger ones (new
   resource folder) should hand off to the TS implementer.

## Codegen reminders to pass to implementers

| Trigger | Command | Run after touching |
| --- | --- | --- |
| `pkg/server/api/console/v1/graphql/*.graphql` | `go generate ./pkg/server/api/console/v1` | Console GraphQL schema |
| `pkg/server/api/connect/v1/graphql/*.graphql` | `go generate ./pkg/server/api/connect/v1` | Connect GraphQL schema |
| `pkg/server/api/trust/v1/graphql/*.graphql` | `go generate ./pkg/server/api/trust/v1` | Trust GraphQL schema |
| `pkg/server/api/mcp/v1/specification.yaml` | `go generate ./pkg/server/api/mcp/v1` | MCP spec |
| Any console GraphQL fragment/op edit | `make relay` | Relay compiler |
| n8n GraphQL ops in `packages/n8n-node` | turbo build (auto) | n8n actions |

When delegating, pass the right codegen command to the implementer.

## Configuration changes — the 11-file rule

If the task touches configuration (new field, new section), remind the
Go implementer to update **all 11** files per `shared.md` § 4:

1. `pkg/probodconfig/<section>.go`
2. `pkg/probodconfig/config.go`
3. `pkg/probod/builder.go`
4. `GNUmakefile` (`make dev-config` args + `cmd/probod-bootstrap` flags)
5. `e2e/internal/testutil/testutil.go`
6. `contrib/lima/provision.sh`
7. `contrib/helm/charts/probo/values.yaml`
8. `contrib/helm/charts/probo/values-production.yaml.example`
9. `contrib/helm/charts/probo/templates/deployment.yaml`
10. `contrib/helm/charts/probo/templates/secret.yaml` (for secrets)
11. `contrib/helm/charts/probo/templates/configmap.yaml` (non-secret)

## When the stack is unclear

If the task description does not clearly map to any stack, do NOT guess.
Ask the user:

> "This task could touch the Go backend (`pkg/...`), the TypeScript
> console (`apps/console`), or both. Where should I make the changes?"

## Post-orchestration checklist

After all implementers finish:

- [ ] Every affected stack had its implementer agent spawned
- [ ] Four-surface rule respected for any backend operation change
- [ ] GraphQL ↔ Relay type names align (frontend uses generated types, no local duplicates)
- [ ] Migration files come before code that depends on them
- [ ] Codegen has been run for every modified schema/spec
- [ ] License headers on every new file (ISC)
- [ ] No `Co-Authored-By` in commits (Probo rule)
- [ ] No PII in logs (entity GIDs only, never emails, names, IPs)
- [ ] No raw `http.Client` in Go (use `go.gearno.de/kit/httpclient` with `WithSSRFProtection()`)
- [ ] Errors wrapped with `cannot <verb> <noun>: %w`, never bare returns

## Reference files

- Shared guidelines: `.claude/guidelines/shared.md`
- Go backend guidelines: `.claude/guidelines/go-backend/`
- TS frontend guidelines: `.claude/guidelines/typescript-frontend/`
- Authoritative subsystem docs: `contrib/claude/*.md` (28 guides — when this skill or the guidelines disagree with a doc, the doc wins)

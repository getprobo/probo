# Probo — Go Backend — Conventions

> Cross-cutting workflow rules (commits, branching, releases, license
> headers, CI gates, PR-mining standards) live in [shared.md](../shared.md).
> This file documents Go-language and Go-package conventions specific to
> the backend.

---

## 1. Naming

### Packages

- Package name = directory name (lower case, single word). `pkg/probod`,
  `pkg/coredata`, `pkg/iam`, `pkg/llm`. Multi-word package directory
  names use underscores **only** when unavoidable (`pkg/server/api/console/v1`).
- Consumer-facing public API lives at the package root; private
  sub-packages stay internal (`pkg/iam/policy` is exported and consumed
  by `pkg/probo`; `pkg/iam/oidc` is exported and consumed only by
  `pkg/probod`).

### Constructors

- **`New*` is the project constructor convention.** Frequency-3 reviewer
  rule (PR #957: *"s/BuildMetadata/NewMetadata/g"*) — `Build*`, `Make*`,
  `Create*` (when returning a domain object, not a DB row) are
  rejected in review. Examples: `NewService`, `NewVendor`,
  `NewCookieBanner`, `NewClient`, `NewMux`, `NewAuthorizer`.
- The matching method on a CLI command file is `NewCmd<Verb>` (e.g.
  `NewCmdList`, `NewCmdCreate`). See
  [`contrib/claude/cli.md`](../../../contrib/claude/cli.md).
- Factory builder methods on entities use `New*().With*().Create()`
  (e2e factory). `Create` is the verb for the terminal step *only*
  inside the e2e factory pattern.

### Types

- **Receiver name** = single letter matching the type, lower case.
  `s` for service types (`*VendorService`, `*Service`), `r` for resolver
  types, `cb` for `*CookieBanner`, `v` for `*Vendor`. Stay consistent
  across all methods on the same type.
- **Interfaces** end in `-er` when they describe behaviour (`Scoper`,
  `Querier`, `Connector`, `Provider`, `Configurable`,
  `AuthorizationAttributer`, `StaleRecoverer`).
- **Error variables**: `Err<Subject><Reason>` (e.g. `ErrResourceNotFound`,
  `ErrSignatureNotCancellable`, `ErrContextLength`). Defined as
  package-level `errors.New`; matched with `errors.Is`.
- **Error types** (when fields are needed): struct named
  `<Subject><Reason>Error` with a `New<Name>Error(...)` constructor and a
  `Error()` method. Matched with `errors.As` or `errors.AsType[T]`.
- **Action constants** (`pkg/probo/actions.go`, `pkg/iam/iam_actions.go`):
  `Action<Resource><Verb>` Go identifier mapped to a
  `service:resource:verb` string (e.g. `ActionVendorRead = "core:vendor:read"`).

### Files

- **`snake_case.go`** for every Go file. One file per entity in the
  flat-package convention: `vendor.go`, `vendor_filter.go`,
  `vendor_order_field.go`, `vendor_service.go`, `vendor_resolvers.go`.
- **Special files** at the package root: `errors.go` (sentinel errors +
  error types), `actions.go` (IAM action constants), `policies.go`
  (policy declarations), `service.go` (root constructor), `scope.go`,
  `entity_type_reg.go`, `migrations.go`.
- **Test files**: `<source>_test.go` co-located with the source for
  internal tests, **`<source>_test.go` in package `<name>_test` for
  black-box tests** (the preferred form — see [testing.md](./testing.md)).
- **Migrations**: `pkg/coredata/migrations/YYYYMMDDTHHMMSSZ.sql`. The time
  portion is **random 6 digits**, not the wall clock — prevents collisions
  when two developers branch off `main` at similar times. See
  [shared.md § Memory note](../shared.md#5-git--workflow).

---

## 2. Imports

> Source: [`contrib/claude/go-style.md`](../../../contrib/claude/go-style.md).

Two import groups, separated by a blank line:

```go
import (
    "context"
    "fmt"
    "time"

    "github.com/jackc/pgx/v5"
    "go.gearno.de/kit/pg"

    "github.com/getprobo/probo/pkg/coredata"
    "github.com/getprobo/probo/pkg/gid"
    "github.com/getprobo/probo/pkg/validator"
)
```

- **Group 1**: stdlib only.
- **Group 2**: everything else (third-party + internal `getprobo` modules)
  sorted alphabetically. The codebase does **not** separate third-party
  from internal — that is intentional per `pkg/probo/vendor_service.go`
  and the wider corpus.
- `goimports` formatting is enforced by `make lint` (`go-fmt` target).
  Save with `goimports`, not `gofmt`.

---

## 3. Error wrapping

> Cross-stack rule in [shared.md § 11](../shared.md#11-error-handling-principles).
> PR-mining frequency 5 (*"error wrap"*).

- **Always wrap with context**: `fmt.Errorf("cannot <verb> <noun>: %w", err)`.
  Lower-case message, no trailing period, `%w` not `%v`. **Bare
  `return err` blocks PR approval.**
- Wrap at the package boundary that adds context (e.g. `pkg/probo`
  service methods wrap `coredata` errors with their own verb).
- For type assertions on error chains, **use `errors.AsType[T](err)`**
  from the kit, **not** `errors.As(err, &ptr)` — codified in
  `contrib/claude/go-style.md` and PR #1038 review.
- Sentinel errors (`coredata.ErrResourceNotFound`,
  `validator.ValidationErrors`, `worker.ErrNoTask`) propagate unchanged
  through service methods; wrappers are added when crossing a layer.

---

## 4. HTTP / status codes

- **Use `http.StatusXxx` constants**, never bare integer literals.
  PR #720 reviewer comment (frequency 1 in review sample, but documented):
  *"Use http.StatusX const please."*
- For internal errors at HTTP/MCP boundaries, call
  `jsonutil.RenderInternalServerError(w)` — never
  `http.Error(w, "...", 500)`.
- Outbound HTTP **must** go through `go.gearno.de/kit/httpclient` with
  `WithSSRFProtection()` enabled by default — see
  [shared.md § 12](../shared.md#12-security-baseline-cross-stack).

---

## 5. URL construction

- **Never `fmt.Sprintf` or `+` to build URLs.** Use `net/url` (`url.URL`,
  `url.Values`) for parsing and building, plus `url.PathEscape` /
  `url.QueryEscape` for segments.
- For application URLs (`/api/...`, `/console/...`, `/trust/...`),
  route through **`pkg/baseurl`** — frequency-3 reviewer rule (PR #800).
  The package centralizes scheme + host + path-prefix, so changes to the
  deployment shape do not ripple.

---

## 6. Struct tags

- **`db:"..."` tags only on coredata entity structs** that map to
  PostgreSQL rows (`pkg/coredata/*.go`). Service request/response structs
  do not get `db` tags.
- **`json:"..."` tags only on structs that are serialized to external
  output** — wire-protocol DTOs (`pkg/webhook/types/*.go`), GraphQL types
  in `pkg/server/api/*/v1/types/`, MCP types, agent message types
  (`pkg/llm/message.go`).
  Frequency-3 reviewer rule (PR #1023): *"i would avoid json tag at this
  level."* Adding `json` tags to internal-only structs (e.g. service
  requests, agent internal state) is a review blocker.
- **`gqlgen` tags**: `@goModel`, `@goField(omittable: true)` declared in
  `.graphql` schema files, not on Go struct tags directly.

---

## 7. Switch / control flow

- **Extract complex `switch`/`case` blocks into private dedicated
  functions.** Frequency-2 reviewer rule (PR #957). If a switch arm has
  more than ~5 lines or makes side-effecting calls, pull it into a
  `func handleXxx(ctx, ...) error` private helper.
- **Resolver error switches must include a `default:`** that logs
  server-side and returns `gqlutils.Internal(ctx)` — see
  [patterns.md § 7](./patterns.md#7-graphql-resolver-shape).
- **Nil checks on `*string` and `**string`** before dereferencing — the
  validator framework auto-derefs in its own checks but service code
  must guard manually.

---

## 8. Pointer-literal style (Go 1.26)

- Use `new(<expr>)` to create a pointer-to-value:
  `new(time.Now())`, `new("hello")`, `new(1)`. Documented in
  `contrib/claude/go-style.md` and visible across `pkg/probo`.
- `&Foo{...}` remains the form for struct literals; `new(Foo{...})` is
  not used.

---

## 9. Concurrency primitives

- **Top-level orchestration**: `sync.WaitGroup` + `context.WithCancelCause`.
  `pkg/probod/probod.go` is the reference.
- **Bounded fan-out**: `errgroup.Group` (only inside subsystems that must
  fail together — e.g. `runTrustCenterServer`'s cert provisioner +
  HTTP/HTTPS quartet).
- **Caches**: `sync.Map` for read-heavy, key-stable caches (e.g.
  `pkg/webhook/sender.go` signing-secret cache). Don't reach for it for
  general state — a `map[K]V` + `sync.RWMutex` is clearer when keys
  change.
- **Once-only finalisation**: `sync.Once` (e.g. `pkg/llm/trace.go`
  `tracedStream` ends the span exactly once whether the stream is
  exhausted or closed early).

---

## 10. Project layout

- **Flat packages by default.** `pkg/probo`, `pkg/coredata`, `pkg/probod`,
  `pkg/iam`, `pkg/llm`, `pkg/agent` all keep every file at the package
  root. Sub-packages exist only when a clear sub-domain emerges
  (`pkg/iam/{policy,oidc,saml,scim}`, `pkg/llm/{anthropic,openai,bedrock}`,
  `pkg/agent/tools/{browser,search,security}`,
  `pkg/server/api/{console,trust,connect,mcp}/v1`).
- **Versioned API roots**: `pkg/server/api/<surface>/v1/`. The `v1`
  segment is mandatory; new versions branch under `v2`, `v3`, etc.
  (none today).
- **`internal/`** is reserved for codegen tooling
  (`internal/cmd/genmodels`). Domain code does not go in `internal/`.

---

## 11. Test packages

- **Black-box tests preferred**: `package <name>_test`, importing the
  package under test. Frequency-2 reviewer rule (PR #1023): *"this test
  must no be in probo package."*
- **`*_test` package** keeps tests honest about the public API and
  prevents reaching into unexported internals.
- Internal helpers that need to be exercised cross-file inside tests
  should be lifted to a public surface or accessed via an exported
  test-only seam (e.g. constructor variant) — never via a same-package
  test.

---

## 12. Test execution

> Source: [`contrib/claude/go-testing.md`](../../../contrib/claude/go-testing.md).

- **`t.Parallel()` is mandatory** at every level: top-level test, every
  `t.Run(...)` subtest, every nested subtest. The e2e profile makes this
  explicit; CI runs e2e with `-race`. Missing `t.Parallel()` serialises
  the suite and is called out in review.
- **`require` for fail-fast preconditions, `assert` for value assertions
  that should accumulate.** A `require.NoError(t, err)` aborts the test
  at the failing line; an `assert.Equal` keeps going. Mixing them
  intentionally is the project style.
- Use **factory builders** (`factory.NewVendor(c).WithName(...).Create()`)
  for non-trivial test setup; flat `factory.CreateXxx(c, factory.Attrs{...})`
  for one-liners. See [testing.md](./testing.md).
- **Auth-sensitive code requires 100% unit-test coverage** — frequency-3
  reviewer rule (PR #957): *"this file must have unit test 100%"*. Applies
  to `pkg/iam/oauth2server/`, `pkg/iam/oidc/`, `pkg/iam/saml/`, PKCE,
  ID-token verification, and any new file that touches token issuance.

---

## 13. Logging

> Cross-stack PII rules in [shared.md § 8](../shared.md#8-logging-principles-cross-stack).

Go-specific:

- **`go.gearno.de/kit/log` only** — never `log/slog`, `zap`, `logrus`.
- **Always `*Ctx` variants** (`InfoCtx`, `WarnCtx`, `ErrorCtx`) so
  trace context propagates.
- **Static messages, dynamic fields**: `l.InfoCtx(ctx, "vendor created",
  log.String("vendor_id", v.ID.String()))` — not `l.InfoCtx(ctx,
  fmt.Sprintf("vendor %s created", v.ID))`.
- **Typed field helpers only**: `log.String`, `log.Int`, `log.Error`,
  `log.Duration`, `log.Bool`. `log.Any` is reserved for trusted-proxy
  lists and similar; avoid it for general state.
- **Derive child loggers** at service boundaries with `.Named("subsystem")`
  and per-scope IDs with `.With(log.String("organization_id", id))`.
- **`gid.GID`s log as opaque base64url strings** (their `String()`
  method) — safe to log; never log the underlying tenant or timestamp
  fragments.

---

## 14. UUIDs and randomness

> Source: [`contrib/claude/coredata.md`](../../../contrib/claude/coredata.md).

- Use **`go.gearno.de/crypto/uuid`** (indirect via `pkg/gid`).
- **Never `github.com/google/uuid`** — the project removed it
  intentionally.
- Random hex / random bytes for secrets: **`pkg/crypto/rand`**.
  Never `math/rand`.

---

## 15. License headers

Every `.go` source file starts with the ISC license header. Year range
expands when editing (never overwrite the original year). Full template
in [shared.md § 6](../shared.md#6-license-headers--isc-on-every-source-file).

---

## 16. Compile-time interface assertions

Add at the bottom of every constructor file that implements an
interface:

```go
var _ unit.Configurable = (*Implm)(nil)
var _ unit.Runnable     = (*Implm)(nil)
var _ worker.Handler[coredata.Evidence] = (*evidenceDescriptionHandler)(nil)
var _ worker.StaleRecoverer            = (*evidenceDescriptionHandler)(nil)
```

These cost nothing at runtime and prevent silent interface drift when
a method signature changes.

---

## 17. Review-Enforced Standards (Go-specific subset)

The full table lives in [shared.md § 13](../shared.md#13-code-review-enforced-standards).
The Go-specific rules consistently enforced in code review:

| Rule | Ref |
| --- | --- |
| All SQL in `pkg/coredata` (no inline SQL in services / handlers / workers) | shared.md #1 |
| Wrap errors with context (`cannot ...: %w`) — bare `return err` blocks approval | shared.md #2 |
| Remove redundant comments that restate the code | shared.md #3 |
| Use `pkg/baseurl` for application URL construction | shared.md #7 |
| `New*` constructor naming | shared.md #8 |
| No `json` tags on internal structs | shared.md #9 |
| Auth-sensitive packages (OAuth2, OIDC, PKCE, ID-token) need 100% unit test coverage | shared.md #11 |
| Webhook payloads use DTOs from `pkg/webhook/types`, never raw `coredata` structs | shared.md #13 |
| Tests live in `*_test` package, not inside the package under test | shared.md #14 |
| Extract complex `switch`/`case` into private functions | shared.md #17 |
| Use `http.StatusXxx` constants, never bare integers | shared.md #18 |
| Avoid JOINs in `coredata` when two queries are clearer | shared.md #19 |

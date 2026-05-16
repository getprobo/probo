# Probo — Go Backend — Pitfalls

> Pitfalls below are concrete: they reference a file path, the symptom,
> the cause, and a fix. Cross-stack pitfalls (PII in logs, SSRF on raw
> `http.Client`, secret rotation, URL string-formatting) are in
> [shared.md § 12](../shared.md#12-security-baseline-cross-stack) and
> [shared.md § 14](../shared.md#14-known-drift--active-violations).

---

## 1. `pkg/coredata/agent_run.go:472` — hardcoded `'PENDING'` SQL literal

**What goes wrong**: Adding a new agent-run status (or renaming
`PENDING`) silently breaks `LoadNextPendingForUpdateSkipLocked` because
the SQL string `WHERE status = 'PENDING'` is not a Go constant.

**Why**: This file deviates from the coredata convention of passing enum
values as named parameters. Reviewers should reject new occurrences;
the current site is documented drift to fix opportunistically.

**Fix**: Replace the literal with a named parameter:

```go
args := pgx.StrictNamedArgs{"status": coredata.AgentRunStatusPending}
// ... WHERE status = @status ...
```

**Source**: `pkg-coredata.json` open question; pattern violates
[`contrib/claude/coredata.md`](../../../contrib/claude/coredata.md).

---

## 2. `pkg/iam/policy` — `In()` and `NotIn()` builders documented but missing

**What goes wrong**: Following
`contrib/claude/authorization.md` literally, you write
`statement.When(policy.In("resource.tag", "a", "b"))`. The build fails:
no such function exists.

**Why**: The doc references a planned API that was never implemented.
The actual surface is the `Condition` struct with
`ConditionOperator` `In` / `NotIn`.

**Fix**: Use the struct directly:

```go
policy.NewStatement(...).When(policy.Condition{
    Operator: policy.ConditionOperatorIn,
    Key:      "resource.tag",
    Values:   []string{"a", "b"},
})
```

**Source**: `pkg-iam-policy.json`; doc drift in `contrib/claude/authorization.md`.

---

## 3. `pkg/iam/oidc` — provider `error_description` logged verbatim

**What goes wrong**: An OIDC IdP returns
`error_description=user johndoe@example.com is locked`. The current
code path logs the description as-is, violating the PII-free rule.

**Why**: External error messages are untrusted strings that frequently
contain emails, account names, and internal hostnames.

**Fix**: Log only the sanitized error code (`error=invalid_grant`).
Drop `error_description` at the boundary, or include it only at
`DEBUG` with a `redact` filter applied. See
[shared.md § 8](../shared.md#8-logging-principles-cross-stack).

**Source**: `pkg-iam-oidc.json` pitfall; cross-referenced in
[shared.md § 14](../shared.md#14-known-drift--active-violations).

---

## 4. `pkg/agent/tools/search` — bare `http.Client` (SSRF gap)

**What goes wrong**: The search tool dials user-influenced URLs with
the standard library `http.Client`. A prompt-injected URL pointing at
`http://169.254.169.254/...` (AWS metadata) or an internal Postgres
endpoint will resolve and connect.

**Why**: SSRF protection is opt-in unless you go through
`go.gearno.de/kit/httpclient`.

**Fix**: Replace the client with
`httpclient.DefaultClient(httpclient.WithSSRFProtection())`. See
[shared.md § 12](../shared.md#12-security-baseline-cross-stack).

**Source**: `pkg-agent-tools-search.json` pitfall.

---

## 5. `pkg/agent/tools/security/csp.go` — missing `netcheck.ValidatePublicURL`

**What goes wrong**: The CSP validator fetches a URL for inspection
without checking that the URL resolves to a public address. Same SSRF
shape as #4, in a different code path.

**Fix**: Wrap the URL with `netcheck.ValidatePublicURL` before
dispatching, **and** use `httpclient.WithSSRFProtection()`. The two are
complementary: ValidatePublicURL is a pre-flight guard;
WithSSRFProtection runs at dial time.

**Source**: `pkg-agent-tools-security.json` pitfall;
also called out for `pkg/server/api/csp.go` in
[shared.md § 14](../shared.md#14-known-drift--active-violations).

---

## 6. `pkg/agent` `NavigateToURLTool` — TOCTOU on redirects

**What goes wrong**: The tool validates the *original* URL before
fetching, but follows redirects. A malicious URL on a public host can
302-redirect to `http://10.0.0.1/...` after the validation has passed.

**Why**: Time-of-check / time-of-use gap between
`netcheck.ValidatePublicURL(originalURL)` and the actual dial that
follows redirects.

**Fix**: Either disable redirects on the http client used for tool
calls (`Client.CheckRedirect = func(*http.Request, []*http.Request) error
{ return http.ErrUseLastResponse }`) and validate each hop manually,
**or** use `httpclient.WithSSRFProtection()` which validates each dial.
The latter is preferred.

**Source**: `pkg-agent.json` pitfall.

---

## 7. `pkg/probod` — every child context is `context.Background()`-derived

**What goes wrong**: You add a new worker, call
`wg.Go(func() { worker.Run(workerCtx) })`, but forget to define and
invoke a `stopWorker()` after `<-ctx.Done()`. Process shutdown hangs
(or kills the worker mid-task with a SIGKILL after the grace period).

**Why**: Probod **intentionally** decouples child contexts from the
parent: each worker gets a `context.Background()`-derived context with
its own cancel function. The only way it learns about shutdown is by
the caller invoking that cancel. Forgetting to call it leaves the worker
running.

**Fix**: For every `wg.Go(...)` block, define
`workerCtx, stopWorker := context.WithCancel(context.Background())`
and add `stopWorker()` to the shutdown sequence after `<-ctx.Done()`,
in the correct order (typically: stop new work in, drain, stop senders,
stop pg).

**Source**: `pkg-probod.json` pitfall — explicitly documented at
`pkg/probod/probod.go` per the profile.

---

## 8. `pkg/trust` `GrantByIDs:387` — inverted `shouldSendEmail` condition

**What goes wrong**: Grants either send a notification email when they
should not, or skip it when they should. The branch evaluates
`shouldSendEmail` with the wrong polarity.

**Why**: The flag was added late and the boolean expression was inverted
in the original PR. There is no test that pins the email side effect.

**Fix**: Read `pkg/trust/grant.go` around line 387 and audit the
condition against the API contract. Add an e2e test that asserts the
Mailpit mailbox state for both granting paths.

**Source**: `pkg-trust.json` pitfall.

---

## 9. `pkg/trust` `GenerateLogoURL` — swallows errors

**What goes wrong**: When the underlying URL builder fails, the function
returns an empty string without surfacing the error. Trust pages
silently render with no logo.

**Fix**: Change the signature to return `(string, error)` and propagate
the wrapped error to the caller, who can decide whether to fall back to
the default logo or fail the page render.

**Source**: `pkg-trust.json` pitfall.

---

## 10. `pkg/coredata` — `*string` vs `**string` confusion in updates

**What goes wrong**: An update endpoint exposes "set field to null" but
the request struct uses `*string`. Now there is no way to distinguish
"don't change" (omit the field) from "set to null" (send null) — both
arrive as `nil`.

**Why**: Single-pointer optional fields conflate two semantics.

**Fix**: Use `**string` (or `**gid.GID`, `**time.Time`) on the update
request:

```go
type UpdateVendorRequest struct {
    Name        *string   // single ptr: nil = no change
    Description **string  // double ptr: outer nil = no change, outer non-nil + inner nil = set NULL
}
```

The validator framework auto-derefs both forms; the coredata Update
method understands both via its own check.

**Source**: `pkg-probo.json` pitfall; canonical reference
`pkg/probo/vendor_service.go`.

---

## 11. `pkg/probo` — 4-file checklist when adding a new entity

**What goes wrong**: You add a new domain entity (`Asset` say), wire it
in `pkg/coredata`, write a service, expose a GraphQL field — and IAM
silently denies access for every role. Or worse, allows it for every
role because no policy matched.

**Why**: A new entity touches **four** files outside the coredata pair
and the service file:

| File | What to add |
| --- | --- |
| `pkg/probo/actions.go` | `ActionAssetCreate / Read / Update / Delete / List` constants |
| `pkg/probo/policies.go` | Allow statements per role, with `organizationCondition` |
| `pkg/coredata/entity_type_reg.go` | Next sequential `uint16` constant + `case` in `NewEntityFromID` |
| `pkg/probo/service.go` | New `*AssetService` field on `TenantService` + wiring in `WithTenant` |

Plus, when the new entity should publish webhooks, add a DTO file in
`pkg/webhook/types/asset.go` with a `NewAsset(coredata.Asset)`
constructor.

**Fix**: Use the entity checklist above when scaffolding. The
authorization rule tells you which file you missed: a denied role with
no policy matched the action; an over-permitted action means the
`organizationCondition` is missing.

**Source**: `pkg-probo.json` pitfall and
`contrib/claude/authorization.md` §"New entity IAM wiring".

---

## 12. `pkg/llm/anthropic` — Anthropic requires `MaxTokens`

**What goes wrong**: `ChatCompletion` returns
`llm.ErrContextLength` immediately, with no API call made.

**Why**: Anthropic's API requires `max_tokens` to be set. The Probo
adapter validates this client-side; OpenAI does not require it.

**Fix**: Always populate `ChatCompletionRequest.MaxTokens` for any
request that may be routed to Anthropic. Default in
`pkg/probod/llm.go` is 4096.

**Source**: `pkg-llm.json`; mapped errors in `pkg/llm/anthropic/provider.go`.

---

## 13. `pkg/llm` streams — `stream.Close()` must always be called

**What goes wrong**: A streaming call returns; the caller breaks out of
the loop on an early error and forgets to call `Close()`. The HTTP
connection leaks (held until socket timeout) and the OTel span never
ends.

**Why**: `tracedStream` ends the span via `sync.Once` triggered by
either stream exhaustion **or** `Close()`. Skipping both leaks both.

**Fix**:

```go
stream, err := client.ChatCompletionStream(ctx, req)
if err != nil { return fmt.Errorf("cannot start stream: %w", err) }
defer stream.Close() // <-- mandatory

for stream.Next() { ... }
if err := stream.Err(); err != nil { return err }
```

**Source**: `pkg-llm.json`; pattern in `pkg/llm/trace.go`.

---

## 14. `pkg/webhook` — DTO discipline has no compile-time guard

**What goes wrong**: `webhook.InsertData(ctx, tx, scope, orgID, "vendor:created", v)`
where `v` is a `*coredata.Vendor`. It compiles, runs, and the customer
receives a payload with internal field names (`tenant_id`, `db_*`),
deprecation-prone column names, and accidental sensitive fields.

**Why**: `InsertData` accepts `any` for the payload and JSON-marshals
it. There is no type seam preventing coredata structs from leaking.

**Fix**: Always wrap with `webhooktypes.NewXxx(coredata)`:

```go
webhook.InsertData(ctx, tx, scope, orgID, "vendor:created", webhooktypes.NewVendor(v))
```

If a DTO doesn't exist for a new entity, **add one** in
`pkg/webhook/types/<entity>.go` with a `NewXxx(coredata.Xxx) Xxx`
constructor. Reviewers enforce this — frequency-2 rule (PR #720).

**Source**: `pkg-webhook.json`; review evidence
[shared.md § 13 #13](../shared.md#13-code-review-enforced-standards).

---

## 15. `pkg/connector` — registering a new OAuth2 provider edits 3 maps

**What goes wrong**: You add `LINEAR` to `providerDefinitions` but
forget to wire `ConnectorRegistry.Register("LINEAR", ...)` in
`pkg/probod/probod.go`, or you forget to expose the provider name in
the GraphQL/MCP enum. Either omission yields a runtime "unknown
provider" error.

**Fix**: Three locations, lockstep:

1. `pkg/connector/providers.go` — `providerDefinitions["LINEAR"] = {...}`.
2. `pkg/probod/probod.go` — call `Register("LINEAR", connector)` after
   `ApplyProviderDefaults` with the deployment's client ID/secret.
3. The connector-provider GraphQL enum (and MCP equivalent) — add
   `LINEAR` to keep the four-surface rule (see
   [shared.md § 3](../shared.md#3-the-four-surface-api-rule)).

**Source**: `pkg-connector.json` and `contrib/claude/api-surface.md`.

---

## 16. `pkg/page` — `CursorKey.Value` is `any`, JSON round-trips can change `int` → `float64`

**What goes wrong**: A cursor key is constructed with an
`int` sort value. After serialising to JSON and reading it back from
the client, the `Value` field is `float64`. Comparing
`reflect.DeepEqual(original, decoded)` fails, or worse, the
SQL parameter becomes a float and the keyset comparison silently
broadens.

**Why**: `CursorKey` stores `any`, and Go's `encoding/json`
defaults all numbers to `float64` on decode. The base64-JSON round trip
inherits the loss.

**Fix**: When parsing a cursor key from a client, normalise the value
to the expected concrete type before using it as a SQL parameter:

```go
ck, err := page.ParseCursorKey(s)
// ... type-switch ck.Value to coerce float64 → int when expected ...
```

In practice, sort values are usually `time.Time` (passes through as
RFC3339 string) or `string`; the int trap appears with numeric sort
fields.

**Source**: `pkg-page.json`.

---

## 17. `pkg/probod` — `runTrustCenterServer` errgroup vs top-level WaitGroup

**What goes wrong**: You "fix" the inconsistency by porting
`runTrustCenterServer` to the top-level `sync.WaitGroup` pattern. The
ACME provisioner can now run when the HTTP server has crashed, and
crash-fast semantics are lost.

**Why**: The errgroup is **intentional** for this subsystem. The cert
provisioner, cert renewer, HTTP server, and HTTPS server form a unit:
if any of them dies, the others must die too. `errgroup.WithContext`
encodes that.

**Fix**: Leave the errgroup in place. Document it locally if you touch
the file. Do not adopt errgroup at the top level — the rest of probod
deliberately separates child lifetimes.

**Source**: `pkg-probod.json` pitfall.

---

## 18. `cmd/<binary>` — adding a new binary requires three lockstep edits

**What goes wrong**: You add `cmd/probo-migrate-thing/main.go`, build
locally, and ship. CI passes, the release pipeline runs, and the new
binary is missing from the release archives and the cross-platform
build matrix.

**Why**: Three integration points outside the package:

1. **`cmd/<name>/main.go`** itself.
2. **`GNUmakefile`** — append the binary to the build matrix and the
   default `make build` target.
3. **GoReleaser CI config** — extend the build/archive entries so the
   new binary is signed, packaged, and uploaded.

**Fix**: Use an existing binary's diff (e.g. `probod-bootstrap`'s
introduction commit) as a checklist. See
[shared.md § 7](../shared.md#7-ci--quality-gates).

**Source**: `cmd.json` open question.

---

## 19. `pkg/iam/oidc`, `pkg/iam/oauth2server`, PKCE — 100% coverage required

**What goes wrong**: A new helper in `pkg/iam/oidc/idtoken.go` ships
without unit tests. PR is blocked by a frequency-3 reviewer comment:
*"this file must have unit test 100%"*.

**Why**: Auth-sensitive packages are reviewed line-by-line and
covered exhaustively — token issuance, signing, verification, PKCE,
state cleanup. A regression here is a security incident.

**Fix**: Cover every branch and every error path with table-driven
unit tests. Use `httptest.NewServer` plus
`httpclient.WithSSRFAllowLoopback()` for IdP fakes (the connector tests
are a good template). See [testing.md § 5](./testing.md#5-coverage-expectations).

**Source**: [shared.md § 13 #11](../shared.md#13-code-review-enforced-standards).

---

## 20. `pkg/iam/oauth2server` — failed PKCE must clean up the auth code

**What goes wrong**: The code-challenge check fails. The handler returns
an error to the client but leaves the authorization code valid, so a
second request can succeed by replaying the code without the
challenge.

**Why**: PR #957 reviewer comment: *"Security issue, if the code
challenge failed it will not delete the code."*

**Fix**: On every PKCE failure path, delete the auth-code row inside
the same transaction as the failure response. See
[shared.md § 12](../shared.md#12-security-baseline-cross-stack).

**Source**: PR-mining; cross-stack rule applied to this stack.

---

## 21. Forgetting `default:` in resolver error switches

**What goes wrong**: A new error type introduced by `pkg/probo` falls
through every `case errors.Is(...)` arm. Without `default:`, gqlgen's
default error formatter forwards the original error to the client —
SQL details, file paths, stack info, all on the wire.

**Fix**: Every resolver error switch ends with:

```go
default:
    r.logger.ErrorCtx(ctx, "cannot <verb> <noun>", log.Error(err))
    return nil, gqlutils.Internal(ctx)
}
```

See [patterns.md § 7](./patterns.md#7-graphql-resolver-shape) and
[shared.md § 3](../shared.md#3-the-four-surface-api-rule).

---

## 22. Forgetting `@goModel` on a Connection type

**What goes wrong**: `vendorConnection.totalCount` always returns 0,
or the dispatch panics with a nil `Resolver` field.

**Why**: Without `@goModel`, gqlgen generates a plain struct lacking
the `Resolver` and `ParentID` fields that the totalCount dispatcher
relies on.

**Fix**: In the `.graphql` file for the entity, add:

```graphql
type VendorConnection @goModel(model: "github.com/getprobo/probo/pkg/server/api/console/v1/types.VendorConnection") {
    edges: [VendorEdge!]!
    pageInfo: PageInfo!
    totalCount: Int!
}
```

**Source**: `pkg-server-api-console.json` pitfall.

---

## 23. Forgetting to run `go generate` after a schema/spec change

**What goes wrong**: Resolver compile errors after editing a `.graphql`
file (or `mcp specification.yaml`), or — worse — the build succeeds
but the server returns "field not found" because `schema.go` is stale.

**Fix**: After every edit to `.graphql` schema files or
`pkg/server/api/mcp/v1/specification.yaml`:

```bash
go generate ./pkg/server/api/console/v1   # or connect/v1, trust/v1
go generate ./pkg/server/api/mcp/v1
make relay                                  # for frontend GraphQL ops
```

CI catches stale generated files; commit the regenerated output. See
[shared.md § 2](../shared.md#2-build--toolchain).

---

## 24. `coredata.NewNoScope()` accidentally used for tenant-bound work

**What goes wrong**: A worker that operates per-tenant uses
`coredata.NewNoScope()`. The first INSERT panics at runtime (NoScope's
`GetTenantID()` panics), or — worse — a SELECT returns rows from every
tenant.

**Why**: NoScope is a deliberate escape hatch for cross-tenant admin
queries (e.g. claiming the next pending row across all tenants). It
**panics** on `GetTenantID()` to prevent insert with no tenant.

**Fix**: After claiming a tenant-owned row with NoScope, switch to
`coredata.NewScopeFromObjectID(row.ID)` for any subsequent
per-tenant work. PR #957 explicitly removed accidental NoScope usage:
*"remove use coredata.NewNoScope() where needed"*. See
[shared.md § 9](../shared.md#9-tenant-isolation--cross-stack-architectural-principle).

**Source**: `pkg-coredata.json` and PR-mining.

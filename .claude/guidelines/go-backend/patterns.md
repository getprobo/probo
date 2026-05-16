# Probo — Go Backend — Patterns

> Cross-stack principles (PII-free logging, SSRF defaults, error wrapping
> rule, tenant isolation, GID layout) live in [shared.md](../shared.md).
> This file documents Go-specific patterns universal to all `pkg/*` code,
> with module-specific deviations called out inline.

---

## 1. Service / TenantService — the domain-service shape

**Universal pattern** (all of `pkg/probo`; mirrored by `pkg/iam`,
`pkg/trust`, `pkg/cookiebanner`, `pkg/esign`, `pkg/connector` for their
respective domains).

```
NewService(ctx, encryptionKey, pgClient, s3Client, ...) → *Service
Service.WithTenant(tenantID) → *TenantService
TenantService.Vendors / .Controls / .Risks / ... → *FooService
FooService.Operation(ctx, req) → entity, error
```

- `Service` (root, tenant-agnostic) holds infrastructure: `*pg.Client`,
  `*log.Logger`, S3 client, `*llm.Client`, file manager, esign, connectors,
  cipher key. It also owns cross-tenant workers (e.g. `Service.ExportJob`).
- `TenantService` carries the `coredata.Scoper` (built from `tenantID`)
  and **exposes every entity sub-service as a public field**.
- Each sub-service (`VendorService`, `ControlService`, `FindingService`)
  has a single `svc *TenantService` field. Methods read `s.svc.scope`,
  `s.svc.pg`, `s.svc.logger` — they never construct a new `Scoper`.
- Service methods are **authorization-free**. IAM checks happen in the
  GraphQL/MCP resolver before the service is called
  (see [§ 4 Authorization](#4-authorization)). Adding `authorize()` calls
  inside a `pkg/probo` method is incorrect.

> See `pkg/probo/service.go` and `pkg/probo/vendor_service.go`.

### Workers attached to the root Service

Cross-tenant workers (e.g. `ExportJob`) live on `*Service`, **not** on
`TenantService`, because they need to operate across all tenants. They
must construct their own `Scoper` from the claimed entity:

```go
// pkg/probo/service.go — pattern shown in lockExportJob
scope := coredata.NewScopeFromObjectID(exportJob.ID)
```

---

## 2. Request + Validate

**Universal pattern** for every mutating service method.

```go
// pkg/probo/vendor_service.go (canonical)
type CreateVendorRequest struct {
    Name        string
    Description *string
    // ...
}

func (r CreateVendorRequest) Validate() error {
    v := validator.New()
    v.Check(r.Name, "name", validator.Required(), validator.MaxLen(NameMaxLength))
    v.Check(r.Description, "description", validator.MaxLen(DescriptionMaxLength))
    return v.Error()
}

func (s *VendorService) Create(ctx context.Context, req CreateVendorRequest) (*coredata.Vendor, error) {
    if err := req.Validate(); err != nil {
        return nil, err
    }
    // ... pg.WithTx, coredata.Insert, webhook.InsertData ...
}
```

- `Validate()` is **the first line** of every mutating method. It returns
  a `validator.ValidationErrors` (a typed slice implementing `error`),
  which the GraphQL/MCP/CLI layer detects with `errors.As` and surfaces
  as field-level form errors — see [shared.md § 11](../shared.md#11-error-handling-principles).
- `validator.New()` is allocated **per call** (the validator is a stateful
  accumulator, not a long-lived service).
- The framework auto-dereferences pointers, so `v.Check(r.Description, …)`
  works whether `Description` is `*string`, `**string`, or a plain string.
- Cross-field rules (e.g. *"`risk_id` is required when status =
  risk_accepted"*) live inside `Validate()` — see
  `pkg/probo/finding_service.go`.

### Update requests use double pointers

```go
type UpdateVendorRequest struct {
    ID          gid.GID
    Name        *string   // single pointer: nil = no change, non-nil = set
    Description **string  // double pointer: outer nil = no change, outer non-nil + inner nil = set NULL
}
```

Single-pointer optional fields conflate "not provided" with "set to null".
For nullable columns where the API must support both, use `**T`. The
validator and the `coredata` Update method both understand this.

---

## 3. Worker pattern (poll-based + FOR UPDATE SKIP LOCKED)

**Universal pattern** for every background job: mailer, slack, esign,
evidence describer, export, webhook delivery, custom-domain renewer.

> Source: [`contrib/claude/go-worker.md`](../../../contrib/claude/go-worker.md).
> Library: `go.gearno.de/kit/worker`.

```go
// pkg/probo/evidence_description_worker.go (canonical)
type evidenceDescriptionHandler struct {
    pg          *pg.Client
    files       *filemanager.Service
    describer   *evidencedescriber.Describer
    logger      *log.Logger
    staleAfter  time.Duration
}

// 1. Claim — atomic state transition PENDING → PROCESSING
func (h *evidenceDescriptionHandler) Claim(ctx context.Context) (*coredata.Evidence, error) {
    var ev *coredata.Evidence
    err := h.pg.WithTx(ctx, func(tx pg.Tx) error {
        e, err := coredata.LoadNextPendingDescriptionForUpdateSkipLocked(ctx, tx, coredata.NewNoScope())
        if errors.Is(err, coredata.ErrResourceNotFound) {
            return worker.ErrNoTask // <-- mandatory translation
        }
        if err != nil {
            return fmt.Errorf("cannot claim evidence: %w", err)
        }
        // ... mark PROCESSING and update ...
        ev = e
        return nil
    })
    return ev, err
}

// 2. Process — long work outside any transaction
func (h *evidenceDescriptionHandler) Process(ctx context.Context, ev *coredata.Evidence) error {
    // ... S3 fetch, LLM call, then a fresh tx to commit COMPLETED/FAILED ...
}

// 3. RecoverStale — resets PROCESSING rows older than staleAfter (default 5 min)
func (h *evidenceDescriptionHandler) RecoverStale(ctx context.Context) error { ... }
```

### Worker rules (universal)

1. **Claim must translate `coredata.ErrResourceNotFound` into
   `worker.ErrNoTask`.** The kit's loop interprets `ErrNoTask` as
   "sleep until next poll"; any other error triggers backoff.
2. **Claim runs inside `pg.WithTx`**, uses `LoadNext*ForUpdateSkipLocked`,
   and updates the row's status to a sentinel value (`PROCESSING`).
3. **Process runs outside any transaction** so long external calls
   (S3, LLM, HTTP) do not hold DB locks. Commits use a fresh transaction.
4. **`RecoverStale` is mandatory** — implement `worker.StaleRecoverer`
   to release rows stuck in `PROCESSING` after `staleAfter` (default 5
   minutes). Without it, a process crash leaves rows stuck forever.
5. **Failure path**: on any error in Process, transition status to
   `FAILED` (or equivalent) and `log.Error(err)` *before* returning the
   wrapped error to the kit.

### Webhook sender deviates intentionally

`pkg/webhook/sender.go` runs its own poll loop (not `kit/worker`)
because it fans out one `WebhookData` to N `WebhookEvent` rows; the
generic worker pattern is single-task. Use `kit/worker` for everything
else.

---

## 4. Authorization

**Universal pattern** at every API surface.

> Source: [`contrib/claude/authorization.md`](../../../contrib/claude/authorization.md).

```go
// pkg/server/api/console/v1/vendor_resolvers.go
func (r *queryResolver) Vendor(ctx context.Context, id gid.GID) (*types.Vendor, error) {
    if err := r.authorize(ctx, id, probo.ActionVendorRead); err != nil {
        return nil, err   // gqlutils.Forbidden / gqlutils.Unauthenticated
    }
    // ... fetch and map ...
}
```

- **Every** GraphQL field resolver and **every** MCP tool body starts with
  an authorize call as the first line, before any data access.
- Action constants (`probo.ActionVendorRead`, `probo.ActionVendorUpdate`,
  ...) live in `pkg/probo/actions.go` for the core service, and
  `pkg/iam/iam_actions.go` for IAM operations. Format:
  `service:resource:operation` (e.g. `core:vendor:read`,
  `iam:organization:update`).
- Policies are **Go code, not DB rows**. They are assembled in
  `pkg/probo/policies.go` (and `pkg/iam/iam_policy_set.go`) using the
  fluent builder from `pkg/iam/policy`:

  ```go
  policy.Allow(probo.ActionVendorRead).
      WithResources(policy.NewResourcePattern("...")).
      When(organizationCondition()) // attribute-based scoping
  ```
- The `organizationCondition` matches `resource.organization_id` against
  `principal.organization_id` — this is **mandatory** on every product
  policy. Without it, a policy leaks across organizations.
- Resource entities implement `iam.AuthorizationAttributer`:

  ```go
  func (v *Vendor) AuthorizationAttributes() map[string]string {
      return map[string]string{"organization_id": v.OrganizationID.String()}
  }
  ```

### Documentation drift

The `contrib/claude/authorization.md` doc references `policy.In(...)` and
`policy.NotIn(...)` builder helpers. **They do not exist in the source**
(`pkg/iam/policy/statement.go`). Use the `Condition` struct directly with
`ConditionOperator` `In` / `NotIn`. See [pitfalls.md](./pitfalls.md).

---

## 5. SQL composition (coredata)

**Universal pattern** — and a hard project rule:
**all SQL lives in `pkg/coredata`, one file per entity. Reviewers reject
inline SQL anywhere else** (see [shared.md § 13](../shared.md#13-code-review-enforced-standards)).

```go
// pkg/coredata/cookie_banner.go (canonical)
const queryTemplate = `
SELECT %s
FROM cookie_banners cb
WHERE %s
  AND %s         -- scope.SQLFragment()
  AND %s         -- filter.SQLFragment()
ORDER BY %s
LIMIT %s
`

func (cb *CookieBanner) LoadByOrganizationID(
    ctx context.Context, conn pg.Querier, scope Scoper,
    organizationID gid.GID, filter *CookieBannerFilter, cursor *page.Cursor[CookieBannerOrderField],
) (*page.Page[CookieBanner, CookieBannerOrderField], error) {
    q := fmt.Sprintf(queryTemplate, columns(), filter.SQLFragment(), scope.SQLFragment(), cursor.SQLFragment(), ...)
    args := pgx.StrictNamedArgs{"organization_id": organizationID}
    maps.Copy(args, scope.SQLArguments())
    maps.Copy(args, filter.SQLArguments())
    maps.Copy(args, cursor.SQLArguments())
    rows, err := conn.Query(ctx, q, args)
    // ... pgx.CollectRows + RowToStructByName ...
}
```

### SQL composition rules (universal)

1. **Templates are static, with `%s` placeholders for fragments only.**
   No runtime conditional string building. `fmt.Sprintf` injects scope,
   filter, cursor, and column-list fragments at call time.
2. **`pgx.StrictNamedArgs` only** — never `pgx.NamedArgs`. StrictNamedArgs
   panics on unknown keys at runtime, catching parameter typos.
3. **`maps.Copy(args, scope.SQLArguments())`** — every Scoper /
   Filter / Cursor declares all of its argument keys (set to `nil` when
   inactive); a missing key crashes StrictNamedArgs. See
   `pkg/coredata/cookie_banner_filter.go` for the convention.
4. **Insert uses `pg.Tx`; Read uses `pg.Querier`.** `pg.Querier` is
   satisfied by both a connection and a transaction; `pg.Tx` is required
   for writes (and `FOR UPDATE` queries).
5. **Tenant ID at insert time only**: `tenant_id := scope.GetTenantID()`
   — never store it on the Go struct (sole exception: `Organization`,
   which *is* the tenant).
6. **Never hardcode enum values in SQL.** Use Go constants as named
   parameters (`@filter_state`, `@status`). Drift: `pkg/coredata/agent_run.go:472`
   hardcodes `'PENDING'` — fix opportunistically.
7. **Update with `Exec` + `RowsAffected` check** by default (see
   `cookie_banner.go`). `RETURNING` is reserved for entities that
   genuinely need the refreshed state (`asset.go`, `agent_run.go`);
   don't introduce new ones casually.
8. **`FOR UPDATE SKIP LOCKED`** queries are named
   `LoadNextXxxForUpdateSkipLocked`, require `pg.Tx`, and are the
   foundation of every worker queue.

### Migrations

- Files: `pkg/coredata/migrations/YYYYMMDDTHHMMSSZ.sql` — **UTC date +
  random 6-digit time portion** to avoid collisions when multiple
  developers branch off `main` simultaneously (per user memory).
- Embedded via `embed.FS` from `pkg/coredata/migrations.go`.
- One logical change per file. **No speculative indexes** — only add
  indexes that solve an observed query latency problem in production.
- Sensitive columns are `BYTEA` and follow the three strategies
  documented in [shared.md § 12](../shared.md#12-security-baseline-cross-stack):
  SHA-256 hash for tokens (`Hashed*`), PBKDF2 for passwords
  (`HashedPassword`), AES-256-GCM for decryptable secrets (`Encrypted*`).

### Entity-type registry

Adding a new entity requires updating `pkg/coredata/entity_type_reg.go`:
add the next sequential `uint16` constant, add a `case` in
`NewEntityFromID`, and **never reuse a removed number** — leave a `_`
placeholder with a comment explaining the gap.

---

## 6. Service orchestration (probod)

**Universal pattern** for the composition root.

> Source: [`contrib/claude/go-service.md`](../../../contrib/claude/go-service.md).

```go
// pkg/probod/probod.go (simplified)
func (impl *Implm) Run(ctx context.Context, l *log.Logger, m prometheus.Registerer, tp trace.TracerProvider) error {
    ctx, cancel := context.WithCancelCause(ctx)
    defer cancel(nil)

    // 1. Synchronous startup — DB migrations BEFORE any goroutine.
    if err := migrator.NewMigrator(pgClient, coredata.Migrations, ...).Run(ctx, "migrations"); err != nil {
        return fmt.Errorf("cannot run migrations: %w", err)
    }

    // 2. Build all services (positional injection).
    // ...

    // 3. Launch each subsystem with its own background-derived context.
    var wg sync.WaitGroup
    apiServerCtx, stopApiServer := context.WithCancel(context.Background()) // <-- context.Background, not ctx
    wg.Go(func() {
        if err := apiServer.Run(apiServerCtx); err != nil {
            cancel(fmt.Errorf("api server crashed: %w", err))
        }
    })
    // ... mailerCtx, slackSenderCtx, exportWorkerCtx, etc. ...

    // 4. Block on shutdown.
    <-ctx.Done()

    // 5. Stop in reverse order (or whatever the dependency chain dictates).
    stopApiServer()
    // ... stopMailer(), stopSlackSender(), ...

    wg.Wait()
    pgClient.Close() // last — guarantees no in-flight queries are abandoned
    return context.Cause(ctx)
}
```

### Orchestration rules (universal)

1. **Every child context is derived from `context.Background()`,
   not the parent `ctx`.** Each subsystem gets an independent lifetime;
   the only shutdown signal it receives is its explicit `stopX()` call
   after `<-ctx.Done()`. **If you forget the stop function, the worker
   runs forever after shutdown begins.**
2. **Crash propagation**: every `wg.Go` block ends with
   `cancel(fmt.Errorf("<subsystem> crashed: %w", err))`. The main
   goroutine returns `context.Cause(ctx)` so the wrapped crash error
   surfaces at process exit.
3. **Migrations are synchronous and blocking.** No retry, no rollback —
   fix the migration and restart.
4. **`pgClient.Close()` is the very last call**, after `wg.Wait()`.
5. `runTrustCenterServer` is the **only** subsystem that uses
   `errgroup.WithContext` (for the cert renewer + provisioner + HTTP +
   HTTPS quartet that must die together). Do not adopt errgroup at the
   top level.
6. `var _ unit.Configurable = (*Implm)(nil)` compile-time assertions are
   used to lock in interface satisfaction without runtime cost.

---

## 7. GraphQL resolver shape

**Universal pattern** across `pkg/server/api/{console,trust,connect}/v1`.

> Source: [`contrib/claude/graphql.md`](../../../contrib/claude/graphql.md).
> Generator: `gqlgen` with `layout: follow-schema`.

```go
// pkg/server/api/console/v1/vendor_resolvers.go (canonical)
func (r *mutationResolver) UpdateVendor(ctx context.Context, input UpdateVendorInput) (*types.UpdateVendorPayload, error) {
    if err := r.authorize(ctx, input.ID, probo.ActionVendorUpdate); err != nil {
        return nil, err
    }
    tenantID := input.ID.TenantID()
    vendor, err := r.ProboService(ctx, tenantID).Vendors.Update(ctx, probo.UpdateVendorRequest{
        ID:   input.ID,
        Name: gqlutils.UnwrapOmittable(input.Name),
        // ...
    })
    if err != nil {
        switch {
        case errors.Is(err, coredata.ErrResourceNotFound):
            return nil, gqlutils.NotFound(err)
        case errors.Is(err, coredata.ErrResourceAlreadyExists):
            return nil, gqlutils.Conflict(err)
        case errors.As(err, &validator.ValidationErrors{}):
            return nil, gqlutils.InvalidValidationErrors(err)
        default:
            r.logger.ErrorCtx(ctx, "cannot update vendor", log.Error(err))
            return nil, gqlutils.Internal(ctx) // <-- mandatory default
        }
    }
    return &types.UpdateVendorPayload{Vendor: types.NewVendor(vendor)}, nil
}
```

### Resolver rules (universal)

1. **Authorization first line, always.**
2. **Error switch must include a `default:`** that logs server-side and
   returns `gqlutils.Internal(ctx)`. Forgetting `default:` leaks SQL,
   stack info, etc. to the wire — see [shared.md § 3](../shared.md#3-the-four-surface-api-rule).
3. **`extend type Mutation` only.** Never `extend type Vendor`,
   `extend type Query`, etc. — gqlgen's follow-schema layout puts the
   resolver in the wrong file otherwise.
4. **DataLoader for any related-entity field traversal.** Use
   `dataloader.FromContext(ctx).Vendor.Load(ctx, id)` inside field
   resolvers. Direct service calls cause N+1.
5. **Connection types need `@goModel`** pointing to a hand-written
   struct in `types/` so `totalCount` resolver dispatch can find the
   `Resolver` and `ParentID` fields. Without it, code compiles but
   `totalCount` returns 0 and the dispatcher panics.
6. **Use `types.NewXxx(coredata)`** to map between layers. Never expose
   coredata structs as GraphQL types.
7. **Cursor pagination via `types.NewCursor` + `types.NewXxxConnection`**.
8. **GraphQL fields whose resolvers can fail must NOT be non-null (`!`).**
   Use Relay `@required` on the client side. Frequency-4 reviewer rule —
   see [shared.md § 13](../shared.md#13-code-review-enforced-standards).
9. **`@goField(omittable: true)`** on nullable update inputs; unwrap with
   `gqlutils.UnwrapOmittable(input.Field)` to convert `graphql.Omittable[T]`
   to `*T` for the service layer.

---

## 8. MCP resolver shape

**Universal pattern** in `pkg/server/api/mcp/v1`.

> Source: [`contrib/claude/mcp.md`](../../../contrib/claude/mcp.md).
> Generator: `mcpgen` driven from `specification.yaml`.

- Tool declarations live in `pkg/server/api/mcp/v1/specification.yaml`;
  regenerate with `go generate ./pkg/server/api/mcp/v1`.
- The hand-written tool body sits next to the generated stub.
- Use **`MustAuthorize(ctx, id, action)`** (panicking variant) — the MCP
  framework recovers panics and renders an internal error. This is the
  **only place** in the codebase that uses panicking authorization.
- **Panic on unexpected** is the convention for state assertions (e.g.
  switch defaults that should be unreachable). Errors that the client
  can act on still go through `jsonutil.RenderInternalServerError(w)`.
- Type helpers live in `pkg/server/api/mcp/v1/types/*.go` — one file per
  entity, each defining `NewXxx(coredata.Xxx)` builders.
- **`Omittable[T]`** distinguishes "field absent" from "field set to
  null" in update tool inputs.

---

## 9. CLI command shape

**Universal pattern** for every leaf command under `pkg/cmd/<resource>/<verb>/`.

> Source: [`contrib/claude/cli.md`](../../../contrib/claude/cli.md).

```go
// pkg/cmd/risk/list/list.go (canonical)
const listQuery = `query ListRisks($orgId: ID!, $first: Int, $after: CursorKey) { ... }`

type listResponse struct { /* ... */ }

func NewCmdList(f *cmdutil.Factory) *cobra.Command {
    var opts listOptions
    cmd := &cobra.Command{
        Use:   "list",
        Short: "List risks",
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg, err := f.Config()
            if err != nil { return err }
            host, hc, err := cfg.DefaultHost()
            if err != nil { return err }
            client := api.NewClient(host, hc.Token, hc.TokenEndpoint, opts.timeout, api.TokenRefreshOption(cfg.Save))
            results, err := api.Paginate[Risk](ctx, client, listQuery, vars, opts.limit, extractConn)
            // ... render to f.IOStreams.Out via cmdutil.NewTable or JSON ...
        },
    }
    return cmd
}
```

### CLI rules (universal)

1. **One file per verb**, named after the verb: `list.go`, `create.go`,
   `view.go`, `update.go`, `delete.go`. Group commands wire the verbs;
   the root wires the groups.
2. **`cmdutil.Factory` is the only DI seam** — every leaf takes
   `f *cmdutil.Factory` and reads `IOStreams`, `Version`, `Config()`
   from it.
3. **GraphQL string is a `const` in the file**, plus an unexported
   `*Response` struct describing the expected payload shape. Inline,
   not generated.
4. **Pagination via `api.Paginate[T]`** — generic helper that walks
   `edges/pageInfo/totalCount`.
5. **Interactive prompts via `huh`** (charmbracelet), but every `huh`
   call must be gated by `f.IOStreams.IsInteractive()` so non-TTY runs
   (CI, scripts, `--no-interactive`) fall back to flag-only mode.
6. **Output: tables to `Out`, truncation/info to `ErrOut`** so piping
   to another command stays clean.

---

## 10. Webhook outbox / payload DTOs

**Universal pattern** for outgoing events.

> Module: `pkg/webhook` + `pkg/webhook/types`.

```go
// pkg/probo/vendor_service.go (inside pg.WithTx)
err := webhook.InsertData(ctx, tx, scope, orgID, "vendor:created", webhooktypes.NewVendor(v))
```

Rules:

1. **`webhook.InsertData` must run inside the same `pg.WithTx` as the
   entity mutation.** Outside the transaction, you can lose either the
   entity or the event.
2. **Payloads must be `pkg/webhook/types.NewXxx(coredata)` DTOs — never
   raw coredata structs.** Frequency-2 reviewer rule (PR #720): *"We
   can't use coredata object as payload for webhook. We must consider
   webhook payload as public API."* There is no compile-time guard;
   reviewers spot it. See [pitfalls.md](./pitfalls.md).
3. The Sender goroutine handles HMAC-SHA256 signing
   (`X-Probo-Webhook-Signature`, `X-Probo-Webhook-Timestamp`),
   per-subscription signing-secret cache, and the FAILED/SUCCEEDED
   record on every attempt.

The same outbox shape is used by `pkg/slack` (writes
`slack_messages` rows in-tx, drained by a Sender) and `pkg/mailer`
(writes `emails` rows in-tx, drained by the mailer worker).

---

## 11. Code generation

**Universal**: codegen is driven by `make generate` / `go generate`.

| Generated artefact | Trigger | Generator |
| --- | --- | --- |
| GraphQL schema/exec/types in each `pkg/server/api/{console,connect,trust}/v1` | Edit `*.graphql`, run `go generate ./<api>/v1` | `gqlgen` |
| MCP resolvers in `pkg/server/api/mcp/v1` | Edit `specification.yaml`, run `go generate ./pkg/server/api/mcp/v1` | `mcpgen` |
| `pkg/llm/registry_gen.go` | Run `go generate ./pkg/llm` | `internal/cmd/genmodels` (sources OpenRouter capability data) |

**Never hand-edit a `*_gen.go` file.** Re-run the generator. Generated
files are checked in (so the build does not depend on network access).

---

## 12. Connector OAuth2 framework

**Module-specific** (`pkg/connector`) but used by every third-party
integration.

```
ConnectorRegistry.Register(name, connector)
                .Initiate(provider, orgID, opts, req) → redirectURL
                .CompleteWithState(provider, req)    → Connection + State
```

Three `TokenEndpointAuth` modes for OAuth2 token exchange — choose per
provider:

- `post-form` — credentials in the form body (most providers).
- `basic-form` — Basic auth header + form body.
- `basic-json` — Basic auth header + JSON body (Notion, some others).

Provider configuration lives in `pkg/connector/providers.go`
(`providerDefinitions` map) and is applied via `ApplyProviderDefaults`
before `Register`. **Adding a new OAuth2 provider requires editing three
maps**: `providerDefinitions` (auth/token URLs), the registry registration
in `pkg/probod/probod.go`, and the GraphQL/MCP enum exposing the provider
to clients.

The OAuth2 `state` parameter is a **stateless HMAC-signed token** carrying
`OrganizationID`, `Provider`, `RequestedScopes`, optional `ContinueURL`
and `ConnectorID` — no DB row needed. Validate with
`statelesstoken.Decode`.

The HTTP client used for token exchange and connector calls **must** be
SSRF-protected (`httpclient.WithSSRFProtection()`) — see [shared.md § 12](../shared.md#12-security-baseline-cross-stack).

---

## 13. Driver pattern (accessreview)

**Module-specific** (`pkg/accessreview`) — one of two registry styles in
the codebase, chosen for **explicit, switch-based** dispatch instead of
`init()`-side-effect registration.

```go
// pkg/accessreview/driver.go
func NewDriver(provider string, conn connector.Connection) (Driver, error) {
    switch provider {
    case "GITHUB":
        return github.New(conn), nil
    case "OKTA":
        return okta.New(conn), nil
    // ...
    default:
        return nil, fmt.Errorf("unsupported provider: %s", provider)
    }
}
```

This is the **preferred shape** for adding a new driver of an existing
domain (auditable diff, no hidden init order). The `connector` registry
above is the only place we use post-`init` registration, and only because
provider config must be wired with secrets at startup.

---

## 14. Observability

**Universal pattern** across services and workers.

- **Logging**: `go.gearno.de/kit/log`, always `*Ctx` variants
  (`InfoCtx`, `ErrorCtx`), typed field helpers (`log.String`, `log.Error`,
  `log.Duration`). Loggers are constructor-injected and derived per
  subsystem with `.Named("subsystem")`. **Never** `fmt.Sprintf` into
  the message — keep messages static, dynamic data in fields.
  PII rules in [shared.md § 8](../shared.md#8-logging-principles-cross-stack).
- **Metrics**: Prometheus via `go.gearno.de/kit` — auto-instrumented for
  HTTP servers, pg.Client, and workers. Hand-written counters/histograms
  are rare and live in the subsystem that needs them.
- **Tracing**: OpenTelemetry. Tracer obtained via `tp.Tracer("subsystem")`
  in `pkg/probod`, propagated through context. `pkg/llm/trace.go` adds
  GenAI semantic-convention spans around every LLM call; provider
  errors are recorded with `span.RecordError(err)`.

---

## 15. Type system

- **Strict throughout.** Domain types live in their owning package
  (`coredata.Vendor`, `policy.Statement`); request/response types are
  co-located with the service method that consumes them
  (`probo.CreateVendorRequest` lives in `pkg/probo/vendor_service.go`).
- **Generics are used sparingly and intentionally**: `page.Cursor[T]`,
  `page.Page[T, O]`, `api.Paginate[T]`, `errors.AsType[T]`,
  `types.OrderBy[T]`. Don't introduce generics for "code reuse" alone —
  the existing usages all encode a real type-level invariant.
- **Pointer types for nullable database columns** (`*string`,
  `*time.Time`); double pointers (`**T`) on update-request structs to
  distinguish "no change" from "set to null" (see § 2).
- **Compile-time interface assertions** at the bottom of constructors:
  `var _ Interface = (*Concrete)(nil)`.

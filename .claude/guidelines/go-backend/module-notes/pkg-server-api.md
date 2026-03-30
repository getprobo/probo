# Probo -- Go Backend -- pkg/server/api (GraphQL, MCP, Protocol Handlers)

> Module-specific patterns that differ from stack-wide conventions.
> For stack-wide patterns, see [patterns.md](../patterns.md) and [conventions.md](../conventions.md).

## Three API Surfaces

The server exposes three API surfaces under `/api/`:

| API | Path | Auth | Schema source | Codegen |
|-----|------|------|--------------|---------|
| Console GraphQL | `/api/console/v1/graphql` | Session cookie or API key | `schema.graphql` | `go generate ./pkg/server/api/console/v1` |
| Connect GraphQL | `/api/connect/v1/graphql` | Session cookie or API key | `schema.graphql` | `go generate ./pkg/server/api/connect/v1` |
| Trust GraphQL | `/trust/{slug}/api/trust/v1/graphql` | Session cookie (optional) | `schema.graphql` | `go generate ./pkg/server/api/trust/v1` |
| MCP | `/mcp/v1` | API key only | `specification.yaml` | `go generate ./pkg/server/api/mcp/v1` |

Every feature must be exposed through GraphQL + MCP + CLI. See [shared.md -- Three-interface API surface rule](../shared.md#cross-stack-contracts).

## GraphQL Resolver Pattern (console/v1, connect/v1, trust/v1)

### Schema-first with gqlgen

GraphQL APIs use schema-first code generation via gqlgen. The schema file (`schema.graphql`) is hand-written. Running `go generate` produces:

- `schema/schema.go` -- executable schema (never edit)
- `types/types.go` -- type struct definitions (never edit)
- `v1_resolver.go` -- resolver stubs (hand-edit method bodies only)

### Resolver method sequence

Every resolver follows this strict sequence:

```go
// See: pkg/server/api/console/v1/v1_resolver.go
func (r *mutationResolver) CreateVendor(ctx context.Context, input types.CreateVendorInput) (*types.CreateVendorPayload, error) {
    // 1. Authorize
    if err := r.authorize(ctx, input.OrganizationID, probo.ActionVendorCreate); err != nil {
        return nil, err
    }

    // 2. Get tenant service
    prb := r.ProboService(ctx, input.OrganizationID.TenantID())

    // 3. Call service method
    vendor, err := prb.Vendors.Create(ctx, &probo.CreateVendorRequest{...})

    // 4. Handle errors + map to types
    if err != nil {
        r.logger.ErrorCtx(ctx, "cannot create vendor", log.Error(err))
        return nil, gqlutils.Internal(ctx)
    }
    return &types.CreateVendorPayload{Vendor: types.NewVendor(vendor)}, nil
}
```

### Connection types (Relay cursor pagination)

Each paginated entity has a `*Connection` struct in `types/` with:
- Edges, PageInfo
- Resolver (parent type for TotalCount dispatch)
- ParentID (gid.GID)
- Filter (optional)

```go
// See: pkg/server/api/console/v1/types/vendor.go
type VendorConnection struct {
    Resolver any
    ParentID gid.GID
    Filter   *coredata.VendorFilter
}
```

TotalCount resolvers dispatch on `obj.Resolver.(type)` to handle different parent contexts. Adding a new parent requires updating the type switch.

### Dataloaders (console/v1 only)

Per-request batch loaders solve N+1 queries for 11 entity types. Injected via HTTP middleware.

```go
// See: pkg/server/api/console/v1/dataloader/dataloader.go
loaders := dataloader.FromContext(ctx)
org, err := loaders.Organization.Load(ctx, vendor.OrganizationID)
```

Always use dataloaders for entities that have them. Direct service calls cause N+1 queries.

### Custom scalars

| GraphQL scalar | Go type | Adapter |
|---|---|---|
| `GID` | `gid.GID` | `gqlutils/types/gid` |
| `Datetime` | `time.Time` | stdlib `time.Time` |
| `CursorKey` | `page.CursorKey` | `gqlutils/types/cursorkey` |
| `Duration` | `time.Duration` | stdlib |
| `BigInt` | `int64` | stdlib |
| `EmailAddr` | `mail.Addr` | `gqlutils/types/addr` |

### Error helpers (gqlutils)

All resolver errors must use typed gqlutils helpers, never raw `gqlerror.Error` construction:

| Helper | Extension code | When to use |
|--------|---------------|-------------|
| `Internal(ctx)` | `INTERNAL_SERVER_ERROR` | Unexpected errors (hides details) |
| `NotFoundf(ctx, ...)` | `NOT_FOUND` | Resource not found |
| `Forbidden(ctx, msg)` | `FORBIDDEN` | Permission denied |
| `Invalidf(ctx, ...)` | `INVALID` | Validation failure |
| `Unauthenticatedf(ctx, ...)` | `UNAUTHENTICATED` | Missing auth |
| `AssumptionRequired(ctx)` | `ASSUMPTION_REQUIRED` | Org session not assumed |
| `NDASignatureRequiredf(ctx, ...)` | `NDA_SIGNATURE_REQUIRED` | Trust center NDA required |

## MCP Resolver Pattern

### Schema-first with mcpgen

MCP tools are defined in `specification.yaml` (hand-written YAML). Running `go generate` produces `server/server.go` and `types/types.go`.

### Resolver differences from GraphQL

1. **Authorization uses panic**: `r.MustAuthorize(ctx, id, action)` panics on failure. `RecoveryMiddleware` catches and translates to MCP error.
2. **First return is always nil**: `return nil, output, nil`
3. **API key auth only**: No session cookies.
4. **Type helpers in types/**: One file per entity with `New*` conversion functions (similar to GraphQL types/).
5. **Optional fields use Omittable**: `mcpgen/omittable` type for nullable update fields. Unwrap with `UnwrapOmittable` helper.

```go
// See: pkg/server/api/mcp/v1/schema.resolvers.go
func (r *Resolver) UpdateVendor(ctx context.Context, req mcp.CallToolRequest, input types.UpdateVendorInput) (*mcp.CallToolResult, *types.UpdateVendorOutput, error) {
    r.MustAuthorize(ctx, input.ID, probo.ActionVendorUpdate)
    // ...
    return nil, types.NewUpdateVendorOutput(vendor), nil
}
```

## Authentication Middleware Chain

The middleware chain order is critical (session -> API key -> identity presence):

```go
// See: pkg/server/api/console/v1/resolver.go
mux.Use(authn.NewSessionMiddleware(iamSvc, cookieConfig))
mux.Use(authn.NewAPIKeyMiddleware(iamSvc, tokenSecret))
mux.Use(authn.NewIdentityPresenceMiddleware())
```

- Session and API key are mutually exclusive (checked symmetrically)
- Identity presence must be last (only checks context, does not authenticate)
- Unexpected errors in middleware panic (caught by upstream recovery)

## Trust Center GraphQL (@nda directive)

The trust API has a unique `@nda` directive that gates field resolution behind completed electronic signatures:

```graphql
# See: pkg/server/api/trust/v1/schema.graphql
type Document implements Node @nda { ... }
```

The directive handler checks NDA completion at field resolution time. New types with NDA-gated content must have `@nda` applied on the type or each field individually.

## Connect API (SAML/OIDC/SCIM handlers)

The connect/v1 module mounts protocol handlers alongside its GraphQL API:

| Endpoint | Handler | Protocol |
|----------|---------|----------|
| `/graphql` | gqlgen | GraphQL |
| `/saml/2.0/*` | SAMLHandler | SAML 2.0 SP |
| `/oidc/*/*` | OIDCHandler | OIDC/OAuth2 |
| `/scim/2.0/*` | SCIMHandler | SCIM 2.0 |

SAML and OIDC handlers bypass GraphQL entirely -- they call IAM services directly, set secure cookies, and issue HTTP redirects.

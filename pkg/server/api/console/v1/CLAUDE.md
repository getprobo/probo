# pkg/server/api/console/v1

GraphQL API using `gqlgen`. Schema-first approach.

## Generated vs hand-written

| File | Type | Notes |
|------|------|-------|
| `schema.graphql` | Hand-written | GraphQL schema definition |
| `gqlgen.yaml` | Hand-written | Codegen config |
| `resolver.go` | Hand-written | Root `Resolver` struct and `NewMux` |
| `graphql_handler.go` | Hand-written | Handler setup |
| `v1_resolver.go` | Generated stubs | Resolver method implementations (edit the bodies) |
| `schema/schema.go` | **Generated — DO NOT EDIT** | Executable schema |
| `types/types.go` | **Generated — DO NOT EDIT** | Type definitions |

## Important rules

- **Never edit generated files** (`schema/schema.go`, `types/types.go`). Only edit `schema.graphql` and resolver bodies.
- **After any change to `schema.graphql`**, always run codegen:

```
go generate ./pkg/server/api/console/v1
```

## Resolver pattern

Every resolver method follows this sequence:

1. **Authorize** — `r.authorize(ctx, obj.ID, probo.ActionXxxGet)`
2. **Get service** — `prb := r.ProboService(ctx, tenantID)`
3. **Call service** — `result, err := prb.Foo.Bar(ctx, ...)`
4. **Handle error** — wrap or panic on unexpected errors

## Pagination

Relay cursor pattern:
- `page.Cursor[OrderField]` for cursor handling
- Connection types (`*Connection`) with `ParentID`, `Resolver`, `Filter` fields

## Custom scalars

`ID`, `Datetime`, `CursorKey`, `Duration`, `BigInt`, `EmailAddr` — mapped in `gqlgen.yaml`.

## Authentication middleware

`NewMux()` chains: session → API key → identity presence middlewares.

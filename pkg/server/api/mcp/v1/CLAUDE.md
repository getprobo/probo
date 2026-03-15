# pkg/server/api/mcp/v1

MCP (Model Context Protocol) API. Schema-first approach using `mcpgen`.

## Generated vs hand-written

| File | Type | Notes |
|------|------|-------|
| `specification.yaml` | Hand-written | Tool definitions, input/output schemas |
| `mcpgen.yaml` | Hand-written | Codegen config |
| `resolver.go` | Hand-written | Resolver struct, `MustAuthorize()`, helpers |
| `v1_handler.go` | Hand-written | `NewMux()`, MCP server setup |
| `middleware.go` | Hand-written | API key authentication |
| `helpers.go` | Hand-written | Pagination helpers |
| `schema.resolvers.go` | **Generated (preserved)** | Tool implementations — edit the bodies |
| `server/server.go` | **Generated — DO NOT EDIT** | Tool registration, `ResolverInterface` |
| `types/types.go` | **Generated — DO NOT EDIT** | Type definitions and JSON schemas |
| `types/*.go` (other) | Hand-written | Type conversion helpers (`NewVendor`, etc.) |

## Important rules

- **Never edit generated files** (`server/server.go`, `types/types.go`). Only edit `specification.yaml`, resolver bodies, and hand-written helpers.
- **After any change to `specification.yaml`**, always run codegen:

```
go generate ./pkg/server/api/mcp/v1
```

Reads `specification.yaml` and generates server, types, and resolver stubs.

## Adding a new tool

1. Define the tool in `specification.yaml` under `tools:` with name, description, hints, inputSchema, outputSchema
2. Define input/output schemas under `components/schemas/`
3. Run `go generate ./pkg/server/api/mcp/v1`
4. Implement the tool body in `schema.resolvers.go`
5. Add type conversion helpers in `types/` if needed

## Tool definition format

```yaml
tools:
  - name: listVendors
    description: List all vendors for the organization
    hints:
      readonly: true
      idempotent: true
    inputSchema:
      $ref: "#/components/schemas/ListVendorsInput"
    outputSchema:
      $ref: "#/components/schemas/ListVendorsOutput"
```

## Resolver pattern

```go
func (r *Resolver) ListVendorsTool(ctx context.Context, input types.ListVendorsInput) (*types.ListVendorsOutput, error) {
    r.MustAuthorize(ctx, input.OrganizationID, probo.ActionVendorList)
    prb := r.ProboService(ctx, input.OrganizationID.TenantID())
    // ... service call, type conversion
}
```

- `MustAuthorize()` panics on auth failure — caught by MCP recovery middleware
- Type conversion via `types.New*()` helpers

## Custom type mappings

In `specification.yaml`, map Go types with `go.probo.inc/mcpgen/type`:

```yaml
OrderDirection:
  type: string
  enum: [ASC, DESC]
  go.probo.inc/mcpgen/type: go.probo.inc/probo/pkg/page.OrderDirection
```

## Authentication

API key auth via `authn.NewAPIKeyMiddleware`. Mounted at `/mcp/v1`.

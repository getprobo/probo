# MCP API Patterns

MCP tools are defined in `pkg/server/api/mcp/v1/specification.yaml` and generated with `mcpgen`. The schema is hand-written; Go types, server registration, and resolver stubs are generated.

## File organization

**Hand-written** (edit these):
- `specification.yaml` — tool definitions, input/output schemas, component schemas
- `resolver.go` — `Resolver` struct, `MustAuthorize`, service accessors
- `helpers.go` — pagination helpers, `UnwrapOmittable`
- `types/*.go` (except `types/types.go`) — type conversion helpers (`NewVendor()`, `NewListVendorsOutput()`, etc.)
- `schema.resolvers.go` — tool implementation bodies (stubs generated, you edit the bodies)

**Generated** (do not edit):
- `server/server.go` — tool registration, `ResolverInterface`
- `types/types.go` — type definitions and JSON schemas

After modifying `specification.yaml`, run:
```bash
go generate ./pkg/server/api/mcp/v1
```

## Tool definition in specification.yaml

```yaml
tools:
  - name: listVendors
    description: List all vendors for the organization
    hints:
      readonly: true
      idempotent: true
      destructive: false
    inputSchema:
      $ref: "#/components/schemas/ListVendorsInput"
    outputSchema:
      $ref: "#/components/schemas/ListVendorsOutput"
```

Input/output schemas reference `components/schemas`. Map custom Go types with the `go.probo.inc/mcpgen/type` extension:

```yaml
components:
  schemas:
    GID:
      type: string
      go.probo.inc/mcpgen/type: go.probo.inc/probo/pkg/gid.GID
```

## Resolver signature

Generated stubs follow this pattern:

```go
func (r *Resolver) ListVendorsTool(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input *types.ListVendorsInput,
) (*mcp.CallToolResult, types.ListVendorsOutput, error)
```

First return is always `nil`. Errors are either returned (for recoverable) or panicked (for authorization and unexpected failures).

## Authorization

Use `MustAuthorize` which panics on failure (caught by middleware):

```go
r.MustAuthorize(ctx, input.OrganizationID, probo.ActionVendorList)
```

## Common resolver patterns

**List with pagination:**
```go
func (r *Resolver) ListVendorsTool(ctx context.Context, req *mcp.CallToolRequest, input *types.ListVendorsInput) (*mcp.CallToolResult, types.ListVendorsOutput, error) {
	r.MustAuthorize(ctx, input.OrganizationID, probo.ActionVendorList)

	prb := r.ProboService(ctx, input.OrganizationID)

	pageOrderBy := page.OrderBy[coredata.VendorOrderField]{
		Field:     coredata.VendorOrderFieldCreatedAt,
		Direction: page.OrderDirectionDesc,
	}
	if input.OrderBy != nil {
		pageOrderBy = page.OrderBy[coredata.VendorOrderField]{
			Field:     input.OrderBy.Field,
			Direction: input.OrderBy.Direction,
		}
	}

	cursor := types.NewCursor(input.Size, input.Cursor, pageOrderBy)

	page, err := prb.Vendors.ListForOrganizationID(ctx, input.OrganizationID, cursor, coredata.NewVendorFilter(nil, nil))
	if err != nil {
		panic(fmt.Errorf("cannot list vendors: %w", err))
	}

	return nil, types.NewListVendorsOutput(page), nil
}
```

**Get single resource:**
```go
func (r *Resolver) GetRiskTool(ctx context.Context, req *mcp.CallToolRequest, input *types.GetRiskInput) (*mcp.CallToolResult, types.GetRiskOutput, error) {
	r.MustAuthorize(ctx, input.ID, probo.ActionRiskGet)
	prb := r.ProboService(ctx, input.ID)

	risk, err := prb.Risks.Get(ctx, input.ID)
	if err != nil {
		return nil, types.GetRiskOutput{}, fmt.Errorf("failed to get risk: %w", err)
	}

	return nil, types.GetRiskOutput{Risk: types.NewRisk(risk)}, nil
}
```

**Create:**
```go
func (r *Resolver) AddRiskTool(ctx context.Context, req *mcp.CallToolRequest, input *types.AddRiskInput) (*mcp.CallToolResult, types.AddRiskOutput, error) {
	r.MustAuthorize(ctx, input.OrganizationID, probo.ActionRiskCreate)
	svc := r.ProboService(ctx, input.OrganizationID)

	risk, err := svc.Risks.Create(ctx, probo.CreateRiskRequest{
		OrganizationID: input.OrganizationID,
		Name:           input.Name,
		Description:    input.Description,
	})
	if err != nil {
		return nil, types.AddRiskOutput{}, fmt.Errorf("failed to create risk: %w", err)
	}

	return nil, types.AddRiskOutput{Risk: types.NewRisk(risk)}, nil
}
```

## Optional fields with Omittable

For nullable update fields, use `go.probo.inc/mcpgen/omittable: true` in the schema:

```yaml
description:
  type:
    - string
    - "null"
  go.probo.inc/mcpgen/omittable: true
```

In resolvers, unwrap with `UnwrapOmittable`:

```go
Description: UnwrapOmittable(input.Description),
```

## Type conversion helpers

Live in `types/*.go` (not the generated `types/types.go`). One file per entity:

```go
func NewVendor(v *coredata.Vendor) *Vendor {
	return &Vendor{
		ID:             v.ID,
		OrganizationID: v.OrganizationID,
		Name:           v.Name,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}

func NewListVendorsOutput(vendorPage *page.Page[*coredata.Vendor, coredata.VendorOrderField]) ListVendorsOutput {
	vendors := make([]*Vendor, 0, len(vendorPage.Data))
	for _, v := range vendorPage.Data {
		vendors = append(vendors, NewVendor(v))
	}

	var nextCursor *page.CursorKey
	if len(vendorPage.Data) > 0 {
		cursorKey := vendorPage.Data[len(vendorPage.Data)-1].CursorKey(vendorPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListVendorsOutput{
		NextCursor: nextCursor,
		Vendors:    vendors,
	}
}
```

## Adding a new MCP tool — checklist

1. **Schema** — add input/output schemas and tool definition in `specification.yaml`
2. **Codegen** — `go generate ./pkg/server/api/mcp/v1`
3. **Resolver** — implement the tool body in `schema.resolvers.go` (authorize, call service, convert types)
4. **Type helpers** — add `New<Entity>()` and `New<Output>()` in `types/<entity>.go`
5. **Verify** — tool is automatically registered via generated `server/server.go`

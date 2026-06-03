# MCP API Patterns

MCP tools are defined in `pkg/server/api/mcp/v1/specification.yaml` and generated with `mcpgen`. The schema is hand-written; Go types, server registration, and resolver stubs are generated.

## File organization

**Hand-written** (edit these):
- `specification.yaml` — tool definitions, input/output schemas, component schemas
- `resolver.go` — `Resolver` struct, `Authorize`, service accessors
- `helpers.go` — pagination helpers, `UnwrapOmittable`
- `types/*.go` (except `types/types.go`) — type conversion helpers (`NewThirdParty()`, `NewListThirdPartiesOutput()`, etc.)
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
  - name: listThirdParties
    description: List all thirdParties for the organization
    hints:
      readonly: true
      idempotent: true
      destructive: false
    inputSchema:
      $ref: "#/components/schemas/ListThirdPartiesInput"
    outputSchema:
      $ref: "#/components/schemas/ListThirdPartiesOutput"
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
func (r *Resolver) ListThirdPartiesTool(
	ctx context.Context,
	req *mcp.CallToolRequest,
	input *types.ListThirdPartiesInput,
) (*mcp.CallToolResult, types.ListThirdPartiesOutput, error)
```

First return is always `nil`. Authorization errors are returned and handled like other recoverable tool errors.

## Authorization

Use `Authorize` with an early return:

```go
scope, err := r.Authorize(ctx, input.OrganizationID, probo.ActionThirdPartyList)
if err != nil {
	return nil, types.ListThirdPartiesOutput{}, err
}
```

## Common resolver patterns

**List with pagination:**
```go
func (r *Resolver) ListThirdPartiesTool(ctx context.Context, req *mcp.CallToolRequest, input *types.ListThirdPartiesInput) (*mcp.CallToolResult, types.ListThirdPartiesOutput, error) {
	if _, err := r.Authorize(ctx, input.OrganizationID, probo.ActionThirdPartyList); err != nil {
		return nil, types.ListThirdPartiesOutput{}, err
	}

	prb := r.ProboService(ctx, input.OrganizationID)

	pageOrderBy := page.OrderBy[coredata.ThirdPartyOrderField]{
		Field:     coredata.ThirdPartyOrderFieldCreatedAt,
		Direction: page.OrderDirectionDesc,
	}
	if input.OrderBy != nil {
		pageOrderBy = page.OrderBy[coredata.ThirdPartyOrderField]{
			Field:     input.OrderBy.Field,
			Direction: input.OrderBy.Direction,
		}
	}

	cursor := types.NewCursor(input.Size, input.Cursor, pageOrderBy)

	page, err := prb.ThirdParties.ListForOrganizationID(ctx, input.OrganizationID, cursor, coredata.NewThirdPartyFilter(nil, nil))
	if err != nil {
		panic(fmt.Errorf("cannot list thirdParties: %w", err))
	}

	return nil, types.NewListThirdPartiesOutput(page), nil
}
```

**Get single resource:**
```go
func (r *Resolver) GetRiskTool(ctx context.Context, req *mcp.CallToolRequest, input *types.GetRiskInput) (*mcp.CallToolResult, types.GetRiskOutput, error) {
	if _, err := r.Authorize(ctx, input.ID, probo.ActionRiskGet); err != nil {
		return nil, types.GetRiskOutput{}, err
	}

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
	if _, err := r.Authorize(ctx, input.OrganizationID, probo.ActionRiskCreate); err != nil {
		return nil, types.AddRiskOutput{}, err
	}

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
func NewThirdParty(v *coredata.ThirdParty) *ThirdParty {
	return &ThirdParty{
		ID:             v.ID,
		OrganizationID: v.OrganizationID,
		Name:           v.Name,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
}

func NewListThirdPartiesOutput(thirdPartyPage *page.Page[*coredata.ThirdParty, coredata.ThirdPartyOrderField]) ListThirdPartiesOutput {
	thirdParties := make([]*ThirdParty, 0, len(thirdPartyPage.Data))
	for _, v := range thirdPartyPage.Data {
		thirdParties = append(thirdParties, NewThirdParty(v))
	}

	var nextCursor *page.CursorKey
	if len(thirdPartyPage.Data) > 0 {
		cursorKey := thirdPartyPage.Data[len(thirdPartyPage.Data)-1].CursorKey(thirdPartyPage.Cursor.OrderBy.Field)
		nextCursor = &cursorKey
	}

	return ListThirdPartiesOutput{
		NextCursor: nextCursor,
		ThirdParties:    thirdParties,
	}
}
```

## Adding a new MCP tool — checklist

1. **Schema** — add input/output schemas and tool definition in `specification.yaml`
2. **Codegen** — `go generate ./pkg/server/api/mcp/v1`
3. **Resolver** — implement the tool body in `schema.resolvers.go` (authorize, call service, convert types)
4. **Type helpers** — add `New<Entity>()` and `New<Output>()` in `types/<entity>.go`
5. **Verify** — tool is automatically registered via generated `server/server.go`

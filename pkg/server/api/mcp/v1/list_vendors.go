package v1

import (
	"context"
	"fmt"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
)

type (
	listVendorsArgs struct {
		OrganizationID string
		OrderField     coredata.VendorOrderField
		Cursor         *page.CursorKey
		Size           int
	}

	listVendorsResult struct {
		NextCursor *string
		Result     []struct {
			Name string
			ID   string
		}
	}
)

func (r *resolver) ListVendors(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args *listVendorsArgs,
) (*mcp.CallToolResult, *listVendorsResult, error) {
	// Get MCP context
	mcpCtx := MCPContextFromContext(ctx)
	if mcpCtx == nil {
		r.logger.ErrorCtx(ctx, "ListVendors: missing MCP context")
		return nil, nil, fmt.Errorf("authentication context not found")
	}

	// Parse and validate organization ID
	organizationID, err := gid.ParseGID(args.OrganizationID)
	if err != nil {
		r.logger.WarnCtx(ctx, "ListVendors: invalid organization_id",
			log.Error(err),
			log.String("user_id", mcpCtx.UserID.String()),
			log.String("organization_id", args.OrganizationID),
		)
		return nil, nil, NewValidationError("organizationID", "invalid organization ID format")
	}

	// Validate user has access to the organization
	if err := ValidateOrganizationAccess(ctx, organizationID); err != nil {
		r.logger.WarnCtx(ctx, "ListVendors: access denied",
			log.Error(err),
			log.String("user_id", mcpCtx.UserID.String()),
			log.String("organization_id", organizationID.String()),
		)
		return nil, nil, err
	}

	tenantID := organizationID.TenantID()

	r.logger.InfoCtx(ctx, "ListVendors: listing vendors",
		log.String("tenant_id", tenantID.String()),
		log.String("organization_id", organizationID.String()),
		log.String("user_id", mcpCtx.UserID.String()),
		log.Int("size", args.Size),
		log.String("order_field", string(args.OrderField)),
	)

	// Note: size range and orderField enum validation is done automatically
	// by the MCP SDK against the InputSchema before reaching this handler

	// Get tenant-scoped service
	svc := r.proboSvc.WithTenant(tenantID)

	filter := coredata.NewVendorFilter(nil, nil)
	cursor := page.NewCursor(
		args.Size,
		args.Cursor,
		page.Head,
		page.OrderBy[coredata.VendorOrderField]{
			Field:     args.OrderField,
			Direction: page.OrderDirectionDesc,
		},
	)

	vendors, err := svc.Vendors.ListForOrganizationID(ctx, organizationID, cursor, filter)
	if err != nil {
		r.logger.ErrorCtx(ctx, "ListVendors: failed to list vendors",
			log.Error(err),
			log.String("tenant_id", tenantID.String()),
			log.String("organization_id", organizationID.String()),
		)
		return nil, nil, fmt.Errorf("failed to list vendors: %w", err)
	}

	result := &listVendorsResult{}
	if len(vendors.Data) > 0 {
		nextCursorKey := vendors.Data[len(vendors.Data)-1].CursorKey(args.OrderField).String()
		result.NextCursor = &nextCursorKey
	}

	for _, vendor := range vendors.Data {
		result.Result = append(
			result.Result,
			struct {
				Name string
				ID   string
			}{
				Name: vendor.Name,
				ID:   vendor.ID.String(),
			},
		)
	}

	r.logger.InfoCtx(ctx, "ListVendors: vendors listed successfully",
		log.String("tenant_id", tenantID.String()),
		log.String("organization_id", organizationID.String()),
		log.Int("count", len(result.Result)),
		log.Bool("has_next", result.NextCursor != nil),
	)

	return nil, result, nil
}

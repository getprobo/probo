package v1

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

var (
	ListVendorsTool = &mcp.Tool{
		Name:         "listVendors",
		Title:        "List Vendors",
		Description:  "List all vendors for the organization",
		Annotations:  &mcp.ToolAnnotations{ReadOnlyHint: true},
		InputSchema:  types.ListVendorsInputSchema,
		OutputSchema: types.ListVendorsOutputSchema,
	}
)

func (r *resolver) ListVendors(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args types.ListVendorsInput,
) (*mcp.CallToolResult, types.ListVendorsOutput, error) {
	prb := r.ProboService(ctx, args.OrganizationID.TenantID())

	pageOrderBy := page.OrderBy[coredata.VendorOrderField]{
		Field:     coredata.VendorOrderFieldCreatedAt,
		Direction: page.OrderDirectionDesc,
	}
	if args.OrderBy != nil {
		pageOrderBy = page.OrderBy[coredata.VendorOrderField]{
			Field:     args.OrderBy.Field,
			Direction: page.OrderDirectionDesc,
		}
	}

	cursor := types.NewCursor(args.Size, args.Cursor, pageOrderBy)

	var vendorFilter = coredata.NewVendorFilter(nil, nil)
	if args.Filter != nil {
		vendorFilter = coredata.NewVendorFilter(&args.Filter.SnapshotID, nil)
	}

	page, err := prb.Vendors.ListForOrganizationID(ctx, args.OrganizationID, cursor, vendorFilter)
	if err != nil {
		panic(fmt.Errorf("cannot list organization vendors: %w", err))
	}

	return nil, types.NewListVendorsOutput(page), nil
}

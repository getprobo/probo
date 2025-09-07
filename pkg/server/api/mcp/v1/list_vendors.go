package v1

import (
	"context"
	"fmt"

	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/page"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type (
	listVendorsArgs struct {
		OrderField coredata.VendorOrderField
		Cursor     *page.CursorKey
		Size       int
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

	vendors, err := r.proboSvc.Vendors.ListForOrganizationID(ctx, r.organizationID, cursor, filter)
	if err != nil {
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

	return nil, result, nil
}

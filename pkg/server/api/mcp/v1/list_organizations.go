// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package v1

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.gearno.de/kit/log"
)

type (
	listOrganizationsArgs struct {
		// No arguments needed - returns all organizations user has access to
	}

	listOrganizationsResult struct {
		Result []struct {
			Name     string
			ID       string
			TenantID string
		}
	}
)

func (r *resolver) ListOrganizations(
	ctx context.Context,
	req *mcp.CallToolRequest,
	args *listOrganizationsArgs,
) (*mcp.CallToolResult, *listOrganizationsResult, error) {
	// Get MCP context
	mcpCtx := MCPContextFromContext(ctx)
	if mcpCtx == nil {
		r.logger.ErrorCtx(ctx, "ListOrganizations: missing MCP context")
		return nil, nil, fmt.Errorf("authentication context not found")
	}

	r.logger.InfoCtx(ctx, "ListOrganizations: listing organizations",
		log.String("user_id", mcpCtx.UserID.String()),
	)

	// Use usrmgrSvc to list all organizations the user has access to
	// This handles multi-tenant access internally
	organizations, err := r.usrmgrSvc.ListOrganizationsForUserID(ctx, mcpCtx.UserID)
	if err != nil {
		r.logger.ErrorCtx(ctx, "ListOrganizations: failed to list organizations",
			log.Error(err),
			log.String("user_id", mcpCtx.UserID.String()),
		)
		return nil, nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	result := &listOrganizationsResult{
		Result: make([]struct {
			Name     string
			ID       string
			TenantID string
		}, 0, len(organizations)),
	}

	for _, org := range organizations {
		result.Result = append(result.Result, struct {
			Name     string
			ID       string
			TenantID string
		}{
			Name:     org.Name,
			ID:       org.ID.String(),
			TenantID: org.ID.TenantID().String(),
		})
	}

	r.logger.InfoCtx(ctx, "ListOrganizations: organizations listed successfully",
		log.String("user_id", mcpCtx.UserID.String()),
		log.Int("count", len(result.Result)),
	)

	return nil, result, nil
}

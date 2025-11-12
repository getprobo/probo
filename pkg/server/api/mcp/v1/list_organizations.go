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

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.probo.inc/probo/pkg/server/api/mcp/v1/types"
)

var (
	ListOrganizationsTool = &mcp.Tool{
		Name:        "listOrganizations",
		Title:       "List Organizations",
		Description: "List all organizations the user has access to",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
		InputSchema: &jsonschema.Schema{
			Type:       "object",
			Properties: map[string]*jsonschema.Schema{},
		},
	}
)

func (r *resolver) ListOrganizations(
	ctx context.Context,
	req *mcp.CallToolRequest,
	_ types.ListOrganizationsInput,
) (*mcp.CallToolResult, types.ListOrganizationsOutput, error) {
	mcpCtx := MCPContextFromContext(ctx)
	organizations, err := r.authzSvc.GetAllUserOrganizations(ctx, mcpCtx.UserID)
	if err != nil {
		return nil, types.ListOrganizationsOutput{}, fmt.Errorf("failed to list organizations: %w", err)
	}

	result := types.ListOrganizationsOutput{
		Organizations: make([]types.Organization, 0, len(organizations)),
	}

	for _, org := range organizations {
		result.Organizations = append(result.Organizations, types.NewOrganization(org))
	}

	return nil, result, nil
}

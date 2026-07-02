// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package mcp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMCP_SetResourceAlias(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	documentID := factory.NewDocument(owner).WithTitle("MCP Alias Document").Create()

	var result struct {
		ResourceAlias struct {
			ResourceID string `json:"resource_id"`
			Alias      string `json:"alias"`
		} `json:"resource_alias"`
	}

	mc.CallToolInto("setResourceAlias", map[string]any{
		"resource_id": documentID,
		"alias":       "mcp-privacy-policy",
	}, &result)
	assert.Equal(t, documentID, result.ResourceAlias.ResourceID)
	assert.Equal(t, "mcp-privacy-policy", result.ResourceAlias.Alias)

	var removeResult struct {
		DeletedResourceID string `json:"deleted_resource_id"`
	}

	mc.CallToolInto("removeResourceAlias", map[string]any{
		"resource_id": documentID,
	}, &removeResult)
	assert.Equal(t, documentID, removeResult.DeletedResourceID)
}

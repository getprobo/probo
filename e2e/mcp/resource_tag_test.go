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
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMCP_ResourceTag(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	documentID := factory.NewDocument(owner).WithTitle("MCP Tagged Document").Create()
	orgID := owner.GetOrganizationID().String()

	var createOut struct {
		ResourceTag struct {
			ID    string  `json:"id"`
			Key   string  `json:"key"`
			Value string  `json:"value"`
			Color *string `json:"color"`
		} `json:"resource_tag"`
	}

	mc.CallToolInto("createResourceTag", map[string]any{
		"organization_id": orgID,
		"key":             "mcp-env",
		"value":           "Production",
		"color":           "#abc",
	}, &createOut)
	assert.Equal(t, "mcp-env", createOut.ResourceTag.Key)
	tagID := createOut.ResourceTag.ID

	mc.CallToolInto("attachResourceTag", map[string]any{
		"resource_id": documentID,
		"tag_id":      tagID,
	}, nil)

	var forResourceOut struct {
		ResourceTags []struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		} `json:"resource_tags"`
	}

	mc.CallToolInto("listResourceTagsForResource", map[string]any{
		"resource_id": documentID,
	}, &forResourceOut)
	require.Len(t, forResourceOut.ResourceTags, 1)
	assert.Equal(t, tagID, forResourceOut.ResourceTags[0].ID)

	var listOut struct {
		ResourceTags []struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		} `json:"resource_tags"`
	}

	mc.CallToolInto("listResourceTags", map[string]any{
		"organization_id": orgID,
	}, &listOut)

	found := false
	for _, tag := range listOut.ResourceTags {
		if tag.ID == tagID {
			found = true
		}
	}
	assert.True(t, found)

	mc.CallToolInto("detachResourceTag", map[string]any{
		"resource_id": documentID,
		"tag_id":      tagID,
	}, nil)

	mc.CallToolInto("deleteResourceTag", map[string]any{
		"resource_tag_id": tagID,
	}, nil)
}

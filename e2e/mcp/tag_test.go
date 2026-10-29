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

func TestMCP_Tag(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	documentID := factory.NewDocument(owner).WithTitle("MCP Tagged Document").Create()
	orgID := owner.GetOrganizationID().String()

	var createOut struct {
		Tag struct {
			ID    string  `json:"id"`
			Key   string  `json:"key"`
			Value string  `json:"value"`
			Color *string `json:"color"`
		} `json:"tag"`
	}

	mc.CallToolInto("createTag", map[string]any{
		"organization_id": orgID,
		"key":             "mcp-env",
		"value":           "Production",
		"color":           "#abc",
	}, &createOut)
	assert.Equal(t, "mcp-env", createOut.Tag.Key)
	tagID := createOut.Tag.ID

	mc.CallToolInto("attachTag", map[string]any{
		"resource_id": documentID,
		"tag_id":      tagID,
	}, nil)

	var forResourceOut struct {
		Tags []struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		} `json:"tags"`
	}

	mc.CallToolInto("listTagsForResource", map[string]any{
		"resource_id": documentID,
	}, &forResourceOut)
	require.Len(t, forResourceOut.Tags, 1)
	assert.Equal(t, tagID, forResourceOut.Tags[0].ID)

	var listOut struct {
		Tags []struct {
			ID  string `json:"id"`
			Key string `json:"key"`
		} `json:"tags"`
	}

	mc.CallToolInto("listTags", map[string]any{
		"organization_id": orgID,
	}, &listOut)

	found := false

	for _, tag := range listOut.Tags {
		if tag.ID == tagID {
			found = true
		}
	}

	assert.True(t, found)

	mc.CallToolInto("detachTag", map[string]any{
		"resource_id": documentID,
		"tag_id":      tagID,
	}, nil)

	mc.CallToolInto("deleteTag", map[string]any{
		"tag_id": tagID,
	}, nil)
}

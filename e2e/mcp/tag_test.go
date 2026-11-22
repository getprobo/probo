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

type mcpTag struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Color *string `json:"color"`
}

func TestMCP_Tag(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	documentID := factory.NewDocument(owner).WithTitle("MCP Tagged Document").Create()
	orgID := owner.GetOrganizationID().String()

	createTag := func(name string, color string) mcpTag {
		var out struct {
			Tag mcpTag `json:"tag"`
		}

		args := map[string]any{
			"organization_id": orgID,
			"name":            name,
		}
		if color != "" {
			args["color"] = color
		}

		mc.CallToolInto("createTag", args, &out)

		return out.Tag
	}

	production := createTag("MCP Production", "#abc")
	assert.Equal(t, "MCP Production", production.Name)
	require.NotNil(t, production.Color)

	staging := createTag("MCP Staging", "")

	var updateOut struct {
		Tag mcpTag `json:"tag"`
	}

	mc.CallToolInto("updateTag", map[string]any{
		"id":    production.ID,
		"color": nil,
	}, &updateOut)
	assert.Nil(t, updateOut.Tag.Color, "null color clears it")

	type attachOut struct {
		Tag mcpTag `json:"tag"`
	}

	var attachProduction attachOut

	mc.CallToolInto("attachTag", map[string]any{
		"resource_id": documentID,
		"tag_id":      production.ID,
	}, &attachProduction)
	assert.Equal(t, production.ID, attachProduction.Tag.ID)

	var attachStaging attachOut

	mc.CallToolInto("attachTag", map[string]any{
		"resource_id": documentID,
		"tag_id":      staging.ID,
	}, &attachStaging)
	assert.Equal(t, staging.ID, attachStaging.Tag.ID)

	var forResourceOut struct {
		Tags []mcpTag `json:"tags"`
	}

	mc.CallToolInto("listTagsForResource", map[string]any{
		"resource_id": documentID,
	}, &forResourceOut)
	require.Len(t, forResourceOut.Tags, 2)

	tagIDs := map[string]struct{}{}
	for _, tag := range forResourceOut.Tags {
		tagIDs[tag.ID] = struct{}{}
	}

	assert.Contains(t, tagIDs, production.ID)
	assert.Contains(t, tagIDs, staging.ID)

	var searchOut struct {
		Resources []string `json:"resources"`
	}

	mc.CallToolInto("search", map[string]any{
		"organization_id": orgID,
		"query":           `tag:"MCP Production" tag:"MCP Staging" type:document`,
	}, &searchOut)
	require.Len(t, searchOut.Resources, 1)
	assert.Equal(t, documentID, searchOut.Resources[0])

	var listOut struct {
		Tags []mcpTag `json:"tags"`
	}

	mc.CallToolInto("listTags", map[string]any{
		"organization_id": orgID,
	}, &listOut)

	found := 0

	for _, tag := range listOut.Tags {
		if tag.ID == production.ID || tag.ID == staging.ID {
			found++
		}
	}

	assert.Equal(t, 2, found, "both tags are in the catalog")

	type detachOut struct {
		Tag      mcpTag `json:"tag"`
		Detached bool   `json:"detached"`
	}

	var detached detachOut

	mc.CallToolInto("detachTag", map[string]any{
		"resource_id": documentID,
		"tag_id":      staging.ID,
	}, &detached)
	assert.True(t, detached.Detached)
	assert.Equal(t, staging.ID, detached.Tag.ID)

	var detachedAgain detachOut

	mc.CallToolInto("detachTag", map[string]any{
		"resource_id": documentID,
		"tag_id":      staging.ID,
	}, &detachedAgain)
	assert.False(t, detachedAgain.Detached)

	var deletedProduction struct {
		DeletedTagID string `json:"deleted_tag_id"`
	}

	mc.CallToolInto("deleteTag", map[string]any{
		"tag_id": production.ID,
	}, &deletedProduction)
	assert.Equal(t, production.ID, deletedProduction.DeletedTagID)

	mc.CallToolInto("listTagsForResource", map[string]any{
		"resource_id": documentID,
	}, &forResourceOut)
	assert.Empty(t, forResourceOut.Tags)

	mc.CallToolInto("search", map[string]any{
		"organization_id": orgID,
		"query":           `tag:"MCP Production" type:document`,
	}, &searchOut)
	assert.NotContains(t, searchOut.Resources, documentID)

	var deletedStaging struct {
		DeletedTagID string `json:"deleted_tag_id"`
	}

	mc.CallToolInto("deleteTag", map[string]any{
		"tag_id": staging.ID,
	}, &deletedStaging)
	assert.Equal(t, staging.ID, deletedStaging.DeletedTagID)
}

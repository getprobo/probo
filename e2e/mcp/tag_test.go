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
	Key   string  `json:"key"`
	Value string  `json:"value"`
	Color *string `json:"color"`
}

func TestMCP_Tag(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	documentID := factory.NewDocument(owner).WithTitle("MCP Tagged Document").Create()
	orgID := owner.GetOrganizationID().String()

	createTag := func(key string, value string, color string) mcpTag {
		var out struct {
			Tag mcpTag `json:"tag"`
		}

		args := map[string]any{
			"organization_id": orgID,
			"key":             key,
			"value":           value,
		}
		if color != "" {
			args["color"] = color
		}

		mc.CallToolInto("createTag", args, &out)

		return out.Tag
	}

	production := createTag("mcp-env", "Production", "#abc")
	assert.Equal(t, "mcp-env", production.Key)
	require.NotNil(t, production.Color)

	staging := createTag("mcp-env", "Staging", "")

	var updateOut struct {
		Tag mcpTag `json:"tag"`
	}

	mc.CallToolInto("updateTag", map[string]any{
		"id":    production.ID,
		"color": nil,
	}, &updateOut)
	assert.Nil(t, updateOut.Tag.Color, "null color clears it")

	type attachOut struct {
		Tag                     mcpTag  `json:"tag"`
		TagAssignmentID         string  `json:"tag_assignment_id"`
		ReplacedTagAssignmentID *string `json:"replaced_tag_assignment_id"`
	}

	var attachProduction attachOut

	mc.CallToolInto("attachTag", map[string]any{
		"resource_id": documentID,
		"tag_id":      production.ID,
	}, &attachProduction)
	assert.Equal(t, production.ID, attachProduction.Tag.ID)
	assert.NotEmpty(t, attachProduction.TagAssignmentID)
	assert.Nil(t, attachProduction.ReplacedTagAssignmentID)

	var attachStaging attachOut

	mc.CallToolInto("attachTag", map[string]any{
		"resource_id": documentID,
		"tag_id":      staging.ID,
	}, &attachStaging)
	require.NotNil(t, attachStaging.ReplacedTagAssignmentID)
	assert.Equal(t, attachProduction.TagAssignmentID, *attachStaging.ReplacedTagAssignmentID)

	var forResourceOut struct {
		Tags []mcpTag `json:"tags"`
	}

	mc.CallToolInto("listTagsForResource", map[string]any{
		"resource_id": documentID,
	}, &forResourceOut)
	require.Len(t, forResourceOut.Tags, 1)
	assert.Equal(t, staging.ID, forResourceOut.Tags[0].ID)

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

	assert.Equal(t, 2, found, "both values of the key are in the catalog")

	type detachOut struct {
		Tag                    mcpTag  `json:"tag"`
		DeletedTagAssignmentID *string `json:"deleted_tag_assignment_id"`
	}

	var detached detachOut

	mc.CallToolInto("detachTag", map[string]any{
		"resource_id": documentID,
		"tag_id":      staging.ID,
	}, &detached)
	require.NotNil(t, detached.DeletedTagAssignmentID)
	assert.Equal(t, attachStaging.TagAssignmentID, *detached.DeletedTagAssignmentID)

	var detachedAgain detachOut

	mc.CallToolInto("detachTag", map[string]any{
		"resource_id": documentID,
		"tag_id":      staging.ID,
	}, &detachedAgain)
	assert.Nil(t, detachedAgain.DeletedTagAssignmentID)

	mc.CallToolInto("deleteTag", map[string]any{
		"tag_id": production.ID,
	}, nil)
	mc.CallToolInto("deleteTag", map[string]any{
		"tag_id": staging.ID,
	}, nil)
}

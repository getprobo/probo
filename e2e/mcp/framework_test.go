// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

package mcp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestMCP_Framework_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		Framework struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"framework"`
	}
	mc.CallToolInto("addFramework", map[string]any{
		"organizationId": orgID,
		"name":           factory.SafeName("Framework"),
	}, &addResult)
	require.NotEmpty(t, addResult.Framework.ID)

	// Get
	var getResult struct {
		Framework struct {
			ID string `json:"id"`
		} `json:"framework"`
	}
	mc.CallToolInto("getFramework", map[string]any{
		"id": addResult.Framework.ID,
	}, &getResult)
	assert.Equal(t, addResult.Framework.ID, getResult.Framework.ID)

	// Update
	var updateResult struct {
		Framework struct {
			ID          string `json:"id"`
			Description string `json:"description"`
		} `json:"framework"`
	}
	mc.CallToolInto("updateFramework", map[string]any{
		"id":          addResult.Framework.ID,
		"description": "Updated description",
	}, &updateResult)
	assert.Equal(t, "Updated description", updateResult.Framework.Description)

	// List
	var listResult struct {
		Frameworks []struct {
			ID string `json:"id"`
		} `json:"frameworks"`
	}
	mc.CallToolInto("listFrameworks", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Frameworks)
}

func TestMCP_Control_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)

	frameworkID := factory.CreateFramework(owner)

	// Create
	var addResult struct {
		Control struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"control"`
	}
	mc.CallToolInto("addControl", map[string]any{
		"frameworkId": frameworkID,
		"name":        factory.SafeName("Control"),
	}, &addResult)
	require.NotEmpty(t, addResult.Control.ID)

	// Get
	var getResult struct {
		Control struct {
			ID string `json:"id"`
		} `json:"control"`
	}
	mc.CallToolInto("getControl", map[string]any{
		"id": addResult.Control.ID,
	}, &getResult)
	assert.Equal(t, addResult.Control.ID, getResult.Control.ID)

	// Update
	var updateResult struct {
		Control struct {
			ID          string `json:"id"`
			Description string `json:"description"`
		} `json:"control"`
	}
	mc.CallToolInto("updateControl", map[string]any{
		"id":          addResult.Control.ID,
		"description": "Updated control",
	}, &updateResult)
	assert.Equal(t, "Updated control", updateResult.Control.Description)

	// List
	var listResult struct {
		Controls []struct {
			ID string `json:"id"`
		} `json:"controls"`
	}
	mc.CallToolInto("listControls", map[string]any{
		"frameworkId": frameworkID,
	}, &listResult)
	assert.NotEmpty(t, listResult.Controls)
}

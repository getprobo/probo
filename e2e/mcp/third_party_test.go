// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

func TestMCP_ThirdParty_CRUD(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Create
	var addResult struct {
		ThirdParty struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"third_party"`
	}

	name := factory.SafeName("ThirdParty")
	mc.CallToolInto("addThirdParty", map[string]any{
		"organizationId": orgID,
		"name":           name,
	}, &addResult)
	require.NotEmpty(t, addResult.ThirdParty.ID)
	assert.Equal(t, name, addResult.ThirdParty.Name)

	// Update
	var updateResult struct {
		ThirdParty struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"third_party"`
	}
	mc.CallToolInto("updateThirdParty", map[string]any{
		"id":   addResult.ThirdParty.ID,
		"name": "Updated ThirdParty",
	}, &updateResult)
	assert.Equal(t, "Updated ThirdParty", updateResult.ThirdParty.Name)

	// List
	var listResult struct {
		ThirdParties []struct {
			ID string `json:"id"`
		} `json:"third_parties"`
	}
	mc.CallToolInto("listThirdParties", map[string]any{
		"organizationId": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.ThirdParties)

	// Delete
	var deleteResult struct {
		DeletedThirdPartyID string `json:"deletedThirdPartyId"`
	}
	mc.CallToolInto("deleteThirdParty", map[string]any{
		"id": addResult.ThirdParty.ID,
	}, &deleteResult)
	assert.Equal(t, addResult.ThirdParty.ID, deleteResult.DeletedThirdPartyID)

	// Update deleted thirdParty returns sanitized not-found error
	msg := mc.CallToolExpectToolError("updateThirdParty", map[string]any{
		"id":   addResult.ThirdParty.ID,
		"name": "Should Fail",
	})
	assert.Equal(t, "resource not found", msg)
}

func TestMCP_ThirdParty_UpdatePreservesCategoryWhenOmitted(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	var addResult struct {
		ThirdParty struct {
			ID       string `json:"id"`
			Category string `json:"category"`
		} `json:"third_party"`
	}

	name := factory.SafeName("ThirdParty")
	mc.CallToolInto("addThirdParty", map[string]any{
		"organizationId": orgID,
		"name":           name,
		"category":       "CLOUD_PROVIDER",
	}, &addResult)
	require.NotEmpty(t, addResult.ThirdParty.ID)
	assert.Equal(t, "CLOUD_PROVIDER", addResult.ThirdParty.Category)

	var updateResult struct {
		ThirdParty struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Category string `json:"category"`
		} `json:"third_party"`
	}
	mc.CallToolInto("updateThirdParty", map[string]any{
		"id":   addResult.ThirdParty.ID,
		"name": "Updated ThirdParty",
	}, &updateResult)
	assert.Equal(t, "Updated ThirdParty", updateResult.ThirdParty.Name)
	assert.Equal(t, "CLOUD_PROVIDER", updateResult.ThirdParty.Category)
}

func TestMCP_ThirdParty_ValidationError(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	msg := mc.CallToolExpectToolError("addThirdParty", map[string]any{
		"organizationId": orgID,
		"name":           "",
	})
	assert.Contains(t, msg, "name")
	assert.NotContains(t, msg, "pq:")
	assert.NotContains(t, msg, "sql:")
}

func TestMCP_ThirdParty_PermissionDenied(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	orgID := owner.GetOrganizationID().String()
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	viewerMC := testutil.NewMCPClient(t, viewer)

	msg := viewerMC.CallToolExpectToolError("addThirdParty", map[string]any{
		"organizationId": orgID,
		"name":           factory.SafeName("ThirdParty"),
	})
	assert.Contains(t, msg, "permission denied")
	assert.NotContains(t, msg, "pq:")
	assert.NotContains(t, msg, "sql:")
}

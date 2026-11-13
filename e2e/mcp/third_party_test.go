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
		} `json:"thirdParty"`
	}

	name := factory.SafeName("ThirdParty")
	mc.CallToolInto("addThirdParty", map[string]any{
		"organization_id": orgID,
		"name":            name,
	}, &addResult)
	require.NotEmpty(t, addResult.ThirdParty.ID)
	assert.Equal(t, name, addResult.ThirdParty.Name)

	// Update
	var updateResult struct {
		ThirdParty struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"thirdParty"`
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
		} `json:"thirdParties"`
	}
	mc.CallToolInto("listThirdParties", map[string]any{
		"organization_id": orgID,
	}, &listResult)
	assert.NotEmpty(t, listResult.ThirdParties)

	// Delete
	var deleteResult struct {
		DeletedThirdPartyID string `json:"deleted_third_party_id"`
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
		} `json:"thirdParty"`
	}

	name := factory.SafeName("ThirdParty")
	mc.CallToolInto("addThirdParty", map[string]any{
		"organization_id": orgID,
		"name":            name,
		"category":        "CLOUD_PROVIDER",
	}, &addResult)
	require.NotEmpty(t, addResult.ThirdParty.ID)
	assert.Equal(t, "CLOUD_PROVIDER", addResult.ThirdParty.Category)

	var updateResult struct {
		ThirdParty struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Category string `json:"category"`
		} `json:"thirdParty"`
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
		"organization_id": orgID,
		"name":            "",
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
		"organization_id": orgID,
		"name":            factory.SafeName("ThirdParty"),
	})
	assert.Contains(t, msg, "permission denied")
	assert.NotContains(t, msg, "pq:")
	assert.NotContains(t, msg, "sql:")
}

func TestMCP_ThirdParty_AddToSupplyChain(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	type thirdPartyResult struct {
		ThirdParty struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			Level int    `json:"level"`
		} `json:"thirdParty"`
	}

	var parent thirdPartyResult
	mc.CallToolInto("addThirdParty", map[string]any{
		"organization_id": orgID,
		"name":            factory.SafeName("Vendor"),
	}, &parent)
	require.NotEmpty(t, parent.ThirdParty.ID)
	assert.Equal(t, 1, parent.ThirdParty.Level)

	var child thirdPartyResult

	childName := factory.SafeName("Sub-processor")
	mc.CallToolInto("addThirdParty", map[string]any{
		"organization_id":       orgID,
		"name":                  childName,
		"parent_third_party_id": parent.ThirdParty.ID,
	}, &child)
	require.NotEmpty(t, child.ThirdParty.ID)
	assert.Equal(t, childName, child.ThirdParty.Name)
	assert.Equal(t, 2, child.ThirdParty.Level)

	var children struct {
		ThirdParties []struct {
			ID string `json:"id"`
		} `json:"thirdParties"`
	}
	mc.CallToolInto("listChildThirdParties", map[string]any{
		"parent_third_party_id": parent.ThirdParty.ID,
	}, &children)
	require.Len(t, children.ThirdParties, 1)
	assert.Equal(t, child.ThirdParty.ID, children.ThirdParties[0].ID)
}

func TestMCP_ThirdParty_AddToSupplyChainOfAnotherOrganization(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	otherOwner := testutil.NewClient(t, testutil.RoleOwner)
	otherMC := testutil.NewMCPClient(t, otherOwner)

	var otherParent struct {
		ThirdParty struct {
			ID string `json:"id"`
		} `json:"thirdParty"`
	}
	otherMC.CallToolInto("addThirdParty", map[string]any{
		"organization_id": otherOwner.GetOrganizationID().String(),
		"name":            factory.SafeName("Other vendor"),
	}, &otherParent)
	require.NotEmpty(t, otherParent.ThirdParty.ID)

	msg := mc.CallToolExpectToolError("addThirdParty", map[string]any{
		"organization_id":       orgID,
		"name":                  factory.SafeName("Sub-processor"),
		"parent_third_party_id": otherParent.ThirdParty.ID,
	})
	assert.NotEmpty(t, msg)
	assert.NotContains(t, msg, "pq:")
	assert.NotContains(t, msg, "sql:")

	var children struct {
		ThirdParties []struct {
			ID string `json:"id"`
		} `json:"thirdParties"`
	}
	otherMC.CallToolInto("listChildThirdParties", map[string]any{
		"parent_third_party_id": otherParent.ThirdParty.ID,
	}, &children)
	assert.Empty(t, children.ThirdParties)
}

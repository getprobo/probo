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

type trustCenter struct {
	ID                 string `json:"id"`
	CompanyName        string `json:"companyName"`
	PageTitle          string `json:"pageTitle"`
	TrustCenterVisible bool   `json:"trustCenterVisible"`
}

type trustCenterReference struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Order int    `json:"order"`
}

type complianceExternalURL struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

func TestMCP_GetTrustCenter(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	var result struct {
		TrustCenter trustCenter `json:"trustCenter"`
	}
	mc.CallToolInto("getTrustCenter", map[string]any{
		"organizationId": orgID,
	}, &result)

	assert.NotEmpty(t, result.TrustCenter.ID)
}

func TestMCP_UpdateTrustCenter(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get trust center ID
	var getResult struct {
		TrustCenter trustCenter `json:"trustCenter"`
	}
	mc.CallToolInto("getTrustCenter", map[string]any{
		"organizationId": orgID,
	}, &getResult)
	require.NotEmpty(t, getResult.TrustCenter.ID)

	// Update
	var updateResult struct {
		TrustCenter trustCenter `json:"trustCenter"`
	}
	mc.CallToolInto("updateTrustCenter", map[string]any{
		"id":          getResult.TrustCenter.ID,
		"companyName": "Updated Company",
		"pageTitle":   "Updated Trust Center",
	}, &updateResult)

	assert.Equal(t, getResult.TrustCenter.ID, updateResult.TrustCenter.ID)
	assert.Equal(t, "Updated Company", updateResult.TrustCenter.CompanyName)
	assert.Equal(t, "Updated Trust Center", updateResult.TrustCenter.PageTitle)
}

func TestMCP_AddTrustCenterReference(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get trust center ID
	var getResult struct {
		TrustCenter trustCenter `json:"trustCenter"`
	}
	mc.CallToolInto("getTrustCenter", map[string]any{
		"organizationId": orgID,
	}, &getResult)
	tcID := getResult.TrustCenter.ID

	var result struct {
		TrustCenterReference trustCenterReference `json:"trustCenterReference"`
	}
	mc.CallToolInto("addTrustCenterReference", map[string]any{
		"trustCenterId": tcID,
		"name":          "SOC 2 Report",
		"url":           "https://example.com/soc2",
	}, &result)

	assert.NotEmpty(t, result.TrustCenterReference.ID)
	assert.Equal(t, "SOC 2 Report", result.TrustCenterReference.Name)
}

func TestMCP_UpdateTrustCenterReference(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get trust center ID
	var getResult struct {
		TrustCenter trustCenter `json:"trustCenter"`
	}
	mc.CallToolInto("getTrustCenter", map[string]any{
		"organizationId": orgID,
	}, &getResult)
	tcID := getResult.TrustCenter.ID

	// Create reference
	var addResult struct {
		TrustCenterReference trustCenterReference `json:"trustCenterReference"`
	}
	mc.CallToolInto("addTrustCenterReference", map[string]any{
		"trustCenterId": tcID,
		"name":          "Original Reference",
		"url":           "https://example.com/original",
	}, &addResult)
	require.NotEmpty(t, addResult.TrustCenterReference.ID)

	// Update reference
	var updateResult struct {
		TrustCenterReference trustCenterReference `json:"trustCenterReference"`
	}
	mc.CallToolInto("updateTrustCenterReference", map[string]any{
		"id":   addResult.TrustCenterReference.ID,
		"name": "Updated Reference",
		"url":  "https://example.com/updated",
	}, &updateResult)

	assert.Equal(t, addResult.TrustCenterReference.ID, updateResult.TrustCenterReference.ID)
	assert.Equal(t, "Updated Reference", updateResult.TrustCenterReference.Name)
}

func TestMCP_DeleteTrustCenterReference(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get trust center ID
	var getResult struct {
		TrustCenter trustCenter `json:"trustCenter"`
	}
	mc.CallToolInto("getTrustCenter", map[string]any{
		"organizationId": orgID,
	}, &getResult)
	tcID := getResult.TrustCenter.ID

	// Create reference
	var addResult struct {
		TrustCenterReference trustCenterReference `json:"trustCenterReference"`
	}
	mc.CallToolInto("addTrustCenterReference", map[string]any{
		"trustCenterId": tcID,
		"name":          "Reference to delete",
		"url":           "https://example.com/delete",
	}, &addResult)
	require.NotEmpty(t, addResult.TrustCenterReference.ID)

	// Delete
	var deleteResult struct {
		DeletedTrustCenterReferenceID string `json:"deletedTrustCenterReferenceId"`
	}
	mc.CallToolInto("deleteTrustCenterReference", map[string]any{
		"id": addResult.TrustCenterReference.ID,
	}, &deleteResult)

	assert.Equal(t, addResult.TrustCenterReference.ID, deleteResult.DeletedTrustCenterReferenceID)
}

func TestMCP_ListTrustCenterReferences(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get trust center ID
	var getResult struct {
		TrustCenter trustCenter `json:"trustCenter"`
	}
	mc.CallToolInto("getTrustCenter", map[string]any{
		"organizationId": orgID,
	}, &getResult)
	tcID := getResult.TrustCenter.ID

	// Create references
	for i := range 2 {
		var result struct {
			TrustCenterReference trustCenterReference `json:"trustCenterReference"`
		}
		mc.CallToolInto("addTrustCenterReference", map[string]any{
			"trustCenterId": tcID,
			"name":          factory.SafeName("Ref"),
			"url":           "https://example.com/" + factory.SafeName("path"),
		}, &result)
		require.NotEmpty(t, result.TrustCenterReference.ID)
		_ = i
	}

	// List
	var listResult struct {
		TrustCenterReferences []trustCenterReference `json:"trustCenterReferences"`
	}
	mc.CallToolInto("listTrustCenterReferences", map[string]any{
		"trustCenterId": tcID,
	}, &listResult)

	assert.GreaterOrEqual(t, len(listResult.TrustCenterReferences), 2)
}

func TestMCP_ListTrustCenterFiles(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get trust center ID
	var getResult struct {
		TrustCenter trustCenter `json:"trustCenter"`
	}
	mc.CallToolInto("getTrustCenter", map[string]any{
		"organizationId": orgID,
	}, &getResult)
	tcID := getResult.TrustCenter.ID

	// List files (may be empty, just verify the tool works)
	var listResult struct {
		TrustCenterFiles []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"trustCenterFiles"`
	}
	mc.CallToolInto("listTrustCenterFiles", map[string]any{
		"trustCenterId": tcID,
	}, &listResult)

	// Just assert the call succeeded — files require multipart upload
	assert.NotNil(t, listResult.TrustCenterFiles)
}

func TestMCP_AddComplianceExternalURL(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get trust center ID
	var getResult struct {
		TrustCenter trustCenter `json:"trustCenter"`
	}
	mc.CallToolInto("getTrustCenter", map[string]any{
		"organizationId": orgID,
	}, &getResult)
	tcID := getResult.TrustCenter.ID

	var result struct {
		ComplianceExternalURL complianceExternalURL `json:"complianceExternalUrl"`
	}
	mc.CallToolInto("addComplianceExternalURL", map[string]any{
		"trustCenterId": tcID,
		"name":          "ISO 27001 Certificate",
		"url":           "https://example.com/iso27001",
	}, &result)

	assert.NotEmpty(t, result.ComplianceExternalURL.ID)
	assert.Equal(t, "ISO 27001 Certificate", result.ComplianceExternalURL.Name)
}

func TestMCP_UpdateComplianceExternalURL(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get trust center ID
	var getResult struct {
		TrustCenter trustCenter `json:"trustCenter"`
	}
	mc.CallToolInto("getTrustCenter", map[string]any{
		"organizationId": orgID,
	}, &getResult)
	tcID := getResult.TrustCenter.ID

	// Create
	var addResult struct {
		ComplianceExternalURL complianceExternalURL `json:"complianceExternalUrl"`
	}
	mc.CallToolInto("addComplianceExternalURL", map[string]any{
		"trustCenterId": tcID,
		"name":          "Original URL",
		"url":           "https://example.com/original",
	}, &addResult)
	require.NotEmpty(t, addResult.ComplianceExternalURL.ID)

	// Update
	var updateResult struct {
		ComplianceExternalURL complianceExternalURL `json:"complianceExternalUrl"`
	}
	mc.CallToolInto("updateComplianceExternalURL", map[string]any{
		"id":   addResult.ComplianceExternalURL.ID,
		"name": "Updated URL",
		"url":  "https://example.com/updated",
	}, &updateResult)

	assert.Equal(t, addResult.ComplianceExternalURL.ID, updateResult.ComplianceExternalURL.ID)
	assert.Equal(t, "Updated URL", updateResult.ComplianceExternalURL.Name)
}

func TestMCP_DeleteComplianceExternalURL(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get trust center ID
	var getResult struct {
		TrustCenter trustCenter `json:"trustCenter"`
	}
	mc.CallToolInto("getTrustCenter", map[string]any{
		"organizationId": orgID,
	}, &getResult)
	tcID := getResult.TrustCenter.ID

	// Create
	var addResult struct {
		ComplianceExternalURL complianceExternalURL `json:"complianceExternalUrl"`
	}
	mc.CallToolInto("addComplianceExternalURL", map[string]any{
		"trustCenterId": tcID,
		"name":          "URL to delete",
		"url":           "https://example.com/delete",
	}, &addResult)
	require.NotEmpty(t, addResult.ComplianceExternalURL.ID)

	// Delete
	var deleteResult struct {
		DeletedComplianceExternalURLID string `json:"deletedComplianceExternalUrlId"`
	}
	mc.CallToolInto("deleteComplianceExternalURL", map[string]any{
		"id": addResult.ComplianceExternalURL.ID,
	}, &deleteResult)

	assert.Equal(t, addResult.ComplianceExternalURL.ID, deleteResult.DeletedComplianceExternalURLID)
}

func TestMCP_ListComplianceExternalURLs(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get trust center ID
	var getResult struct {
		TrustCenter trustCenter `json:"trustCenter"`
	}
	mc.CallToolInto("getTrustCenter", map[string]any{
		"organizationId": orgID,
	}, &getResult)
	tcID := getResult.TrustCenter.ID

	// Create URLs
	for i := range 2 {
		var result struct {
			ComplianceExternalURL complianceExternalURL `json:"complianceExternalUrl"`
		}
		mc.CallToolInto("addComplianceExternalURL", map[string]any{
			"trustCenterId": tcID,
			"name":          factory.SafeName("URL"),
			"url":           "https://example.com/" + factory.SafeName("path"),
		}, &result)
		require.NotEmpty(t, result.ComplianceExternalURL.ID)
		_ = i
	}

	// List
	var listResult struct {
		ComplianceExternalURLs []complianceExternalURL `json:"complianceExternalUrls"`
	}
	mc.CallToolInto("listComplianceExternalURLs", map[string]any{
		"trustCenterId": tcID,
	}, &listResult)

	assert.GreaterOrEqual(t, len(listResult.ComplianceExternalURLs), 2)
}

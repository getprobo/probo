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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

type mcpFile struct {
	DownloadURL string `json:"download_url"`
}

type trustCenter struct {
	ID                 string   `json:"id"`
	CompanyName        string   `json:"companyName"`
	PageTitle          string   `json:"pageTitle"`
	TrustCenterVisible bool     `json:"trustCenterVisible"`
	Logo               *mcpFile `json:"logo,omitempty"`
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

	const trustCenterQuery = `
		query($organizationId: ID!) {
			node(id: $organizationId) {
				... on Organization {
					trustCenter {
						id
					}
				}
			}
		}
	`

	var trustCenterLookup struct {
		Node struct {
			TrustCenter struct {
				ID string `json:"id"`
			} `json:"trustCenter"`
		} `json:"node"`
	}

	err := owner.Execute(trustCenterQuery, map[string]any{
		"organizationId": orgID,
	}, &trustCenterLookup)
	require.NoError(t, err)
	require.NotEmpty(t, trustCenterLookup.Node.TrustCenter.ID)

	trustCenterID := trustCenterLookup.Node.TrustCenter.ID

	const uploadMutation = `
		mutation UpdateTrustCenterBrand($input: UpdateTrustCenterBrandInput!) {
			updateTrustCenterBrand(input: $input) {
				trustCenter {
					id
					logo {
						id
						downloadUrl
					}
				}
			}
		}
	`

	pngContent := []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}

	var uploadResult struct {
		UpdateTrustCenterBrand struct {
			TrustCenter struct {
				ID   string `json:"id"`
				Logo *struct {
					ID          string `json:"id"`
					DownloadURL string `json:"downloadUrl"`
				} `json:"logo"`
			} `json:"trustCenter"`
		} `json:"updateTrustCenterBrand"`
	}

	err = owner.ExecuteWithFile(uploadMutation, map[string]any{
		"input": map[string]any{
			"trustCenterId": trustCenterID,
			"logoFile":      nil,
		},
	}, "input.logoFile", testutil.UploadFile{
		Filename:    "mcp-trust-center-logo.png",
		ContentType: "image/png",
		Content:     pngContent,
	}, &uploadResult)
	require.NoError(t, err)
	require.NotNil(t, uploadResult.UpdateTrustCenterBrand.TrustCenter.Logo)

	var result struct {
		TrustCenter trustCenter `json:"trustCenter"`
	}
	mc.CallToolInto("getTrustCenter", map[string]any{
		"organizationId": orgID,
	}, &result)

	assert.NotEmpty(t, result.TrustCenter.ID)
	require.NotNil(t, result.TrustCenter.Logo)
	assert.True(
		t,
		strings.Contains(result.TrustCenter.Logo.DownloadURL, "/api/files/v1/public/"),
		"download_url must route through the public files API, got %q",
		result.TrustCenter.Logo.DownloadURL,
	)
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

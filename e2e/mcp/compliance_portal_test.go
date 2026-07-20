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

type compliancePortal struct {
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	Active bool     `json:"active"`
	Logo   *mcpFile `json:"logo,omitempty"`
}

type compliancePortalReference struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	WebsiteURL  *string `json:"website_url"`
	Description *string `json:"description"`
}

type complianceCustomLink struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

func TestMCP_GetCompliancePortal(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	const compliancePortalQuery = `
		query($organizationId: ID!) {
			node(id: $organizationId) {
				... on Organization {
					compliancePortal {
						id
					}
				}
			}
		}
	`

	var compliancePortalLookup struct {
		Node struct {
			CompliancePortal struct {
				ID string `json:"id"`
			} `json:"compliancePortal"`
		} `json:"node"`
	}

	err := owner.Execute(compliancePortalQuery, map[string]any{
		"organizationId": orgID,
	}, &compliancePortalLookup)
	require.NoError(t, err)
	require.NotEmpty(t, compliancePortalLookup.Node.CompliancePortal.ID)

	compliancePortalID := compliancePortalLookup.Node.CompliancePortal.ID

	const uploadMutation = `
		mutation UpdateCompliancePortalBrand($input: UpdateCompliancePortalBrandInput!) {
			updateCompliancePortalBrand(input: $input) {
				compliancePortal {
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
		UpdateCompliancePortalBrand struct {
			CompliancePortal struct {
				ID   string `json:"id"`
				Logo *struct {
					ID          string `json:"id"`
					DownloadURL string `json:"downloadUrl"`
				} `json:"logo"`
			} `json:"compliancePortal"`
		} `json:"updateCompliancePortalBrand"`
	}

	err = owner.ExecuteWithFile(uploadMutation, map[string]any{
		"input": map[string]any{
			"compliancePortalId": compliancePortalID,
			"logoFile":           nil,
		},
	}, "input.logoFile", testutil.UploadFile{
		Filename:    "mcp-compliance-portal-logo.png",
		ContentType: "image/png",
		Content:     pngContent,
	}, &uploadResult)
	require.NoError(t, err)
	require.NotNil(t, uploadResult.UpdateCompliancePortalBrand.CompliancePortal.Logo)

	var result struct {
		CompliancePortal compliancePortal `json:"compliance_portal"`
	}
	mc.CallToolInto("getCompliancePortal", map[string]any{
		"organization_id": orgID,
	}, &result)

	assert.NotEmpty(t, result.CompliancePortal.ID)
	require.NotNil(t, result.CompliancePortal.Logo)
	assert.True(
		t,
		strings.Contains(result.CompliancePortal.Logo.DownloadURL, "/api/files/v1/public/"),
		"download_url must route through the public files API, got %q",
		result.CompliancePortal.Logo.DownloadURL,
	)
}

func TestMCP_UpdateCompliancePortal(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get compliance portal ID
	var getResult struct {
		CompliancePortal compliancePortal `json:"compliance_portal"`
	}
	mc.CallToolInto("getCompliancePortal", map[string]any{
		"organization_id": orgID,
	}, &getResult)
	require.NotEmpty(t, getResult.CompliancePortal.ID)

	// Update
	var updateResult struct {
		CompliancePortal struct {
			ID                 string  `json:"id"`
			Title              string  `json:"title"`
			Description        *string `json:"description"`
			WebsiteURL         *string `json:"website_url"`
			Email              *string `json:"email"`
			HeadquarterAddress *string `json:"headquarter_address"`
		} `json:"compliance_portal"`
	}
	mc.CallToolInto("updateCompliancePortal", map[string]any{
		"compliance_portal_id": getResult.CompliancePortal.ID,
		"title":                "Acme Security",
		"description":          "We keep your data safe.",
		"website_url":          "https://example.com",
		"email":                "security@example.com",
		"headquarter_address":  "123 Main St, San Francisco, CA 94102",
	}, &updateResult)

	assert.Equal(t, getResult.CompliancePortal.ID, updateResult.CompliancePortal.ID)
	assert.Equal(t, "Acme Security", updateResult.CompliancePortal.Title)
	require.NotNil(t, updateResult.CompliancePortal.Description)
	assert.Equal(t, "We keep your data safe.", *updateResult.CompliancePortal.Description)
	require.NotNil(t, updateResult.CompliancePortal.WebsiteURL)
	assert.Equal(t, "https://example.com", *updateResult.CompliancePortal.WebsiteURL)
	require.NotNil(t, updateResult.CompliancePortal.Email)
	assert.Equal(t, "security@example.com", *updateResult.CompliancePortal.Email)
	require.NotNil(t, updateResult.CompliancePortal.HeadquarterAddress)
	assert.Equal(t, "123 Main St, San Francisco, CA 94102", *updateResult.CompliancePortal.HeadquarterAddress)
}

func TestMCP_AddCompliancePortalReference(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get compliance portal ID
	var getResult struct {
		CompliancePortal compliancePortal `json:"compliance_portal"`
	}
	mc.CallToolInto("getCompliancePortal", map[string]any{
		"organization_id": orgID,
	}, &getResult)
	portalID := getResult.CompliancePortal.ID

	var result struct {
		CompliancePortalReference compliancePortalReference `json:"compliance_portal_reference"`
	}
	mc.CallToolInto("addCompliancePortalReference", map[string]any{
		"compliance_portal_id": portalID,
		"name":                 "SOC 2 Report",
		"website_url":          "https://example.com/soc2",
	}, &result)

	assert.NotEmpty(t, result.CompliancePortalReference.ID)
	assert.Equal(t, "SOC 2 Report", result.CompliancePortalReference.Name)
}

func TestMCP_UpdateCompliancePortalReference(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get compliance portal ID
	var getResult struct {
		CompliancePortal compliancePortal `json:"compliance_portal"`
	}
	mc.CallToolInto("getCompliancePortal", map[string]any{
		"organization_id": orgID,
	}, &getResult)
	portalID := getResult.CompliancePortal.ID

	// Create reference
	var addResult struct {
		CompliancePortalReference compliancePortalReference `json:"compliance_portal_reference"`
	}
	mc.CallToolInto("addCompliancePortalReference", map[string]any{
		"compliance_portal_id": portalID,
		"name":                 "Original Reference",
		"website_url":          "https://example.com/original",
	}, &addResult)
	require.NotEmpty(t, addResult.CompliancePortalReference.ID)

	// Update reference
	var updateResult struct {
		CompliancePortalReference compliancePortalReference `json:"compliance_portal_reference"`
	}
	mc.CallToolInto("updateCompliancePortalReference", map[string]any{
		"id":          addResult.CompliancePortalReference.ID,
		"name":        "Updated Reference",
		"website_url": "https://example.com/updated",
	}, &updateResult)

	assert.Equal(t, addResult.CompliancePortalReference.ID, updateResult.CompliancePortalReference.ID)
	assert.Equal(t, "Updated Reference", updateResult.CompliancePortalReference.Name)
}

func TestMCP_DeleteCompliancePortalReference(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get compliance portal ID
	var getResult struct {
		CompliancePortal compliancePortal `json:"compliance_portal"`
	}
	mc.CallToolInto("getCompliancePortal", map[string]any{
		"organization_id": orgID,
	}, &getResult)
	portalID := getResult.CompliancePortal.ID

	// Create reference
	var addResult struct {
		CompliancePortalReference compliancePortalReference `json:"compliance_portal_reference"`
	}
	mc.CallToolInto("addCompliancePortalReference", map[string]any{
		"compliance_portal_id": portalID,
		"name":                 "Reference to delete",
		"website_url":          "https://example.com/delete",
	}, &addResult)
	require.NotEmpty(t, addResult.CompliancePortalReference.ID)

	// Delete
	var deleteResult struct {
		DeletedCompliancePortalReferenceID string `json:"deleted_compliance_portal_reference_id"`
	}
	mc.CallToolInto("deleteCompliancePortalReference", map[string]any{
		"id": addResult.CompliancePortalReference.ID,
	}, &deleteResult)

	assert.Equal(t, addResult.CompliancePortalReference.ID, deleteResult.DeletedCompliancePortalReferenceID)
}

func TestMCP_ListCompliancePortalReferences(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get compliance portal ID
	var getResult struct {
		CompliancePortal compliancePortal `json:"compliance_portal"`
	}
	mc.CallToolInto("getCompliancePortal", map[string]any{
		"organization_id": orgID,
	}, &getResult)
	portalID := getResult.CompliancePortal.ID

	// Create references
	for i := range 2 {
		var result struct {
			CompliancePortalReference compliancePortalReference `json:"compliance_portal_reference"`
		}
		mc.CallToolInto("addCompliancePortalReference", map[string]any{
			"compliance_portal_id": portalID,
			"name":                 factory.SafeName("Ref"),
			"website_url":          "https://example.com/" + factory.SafeName("path"),
		}, &result)
		require.NotEmpty(t, result.CompliancePortalReference.ID)

		_ = i
	}

	// List
	var listResult struct {
		CompliancePortalReferences []compliancePortalReference `json:"compliance_portal_references"`
	}
	mc.CallToolInto("listCompliancePortalReferences", map[string]any{
		"compliance_portal_id": portalID,
	}, &listResult)

	assert.GreaterOrEqual(t, len(listResult.CompliancePortalReferences), 2)
}

func TestMCP_ListCompliancePortalFiles(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// List files (may be empty, just verify the tool works)
	var listResult struct {
		CompliancePortalFiles []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"compliance_portal_files"`
	}
	mc.CallToolInto("listCompliancePortalFiles", map[string]any{
		"organization_id": orgID,
	}, &listResult)

	// Just assert the call succeeded — files require multipart upload
	assert.NotNil(t, listResult.CompliancePortalFiles)
}

func TestMCP_AddComplianceCustomLink(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get compliance portal ID
	var getResult struct {
		CompliancePortal compliancePortal `json:"compliance_portal"`
	}
	mc.CallToolInto("getCompliancePortal", map[string]any{
		"organization_id": orgID,
	}, &getResult)
	portalID := getResult.CompliancePortal.ID

	var result struct {
		ComplianceCustomLink complianceCustomLink `json:"compliance_custom_link"`
	}
	mc.CallToolInto("addComplianceCustomLink", map[string]any{
		"compliance_portal_id": portalID,
		"name":                 "ISO 27001 Certificate",
		"url":                  "https://example.com/iso27001",
	}, &result)

	assert.NotEmpty(t, result.ComplianceCustomLink.ID)
	assert.Equal(t, "ISO 27001 Certificate", result.ComplianceCustomLink.Name)
}

func TestMCP_UpdateComplianceCustomLink(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get compliance portal ID
	var getResult struct {
		CompliancePortal compliancePortal `json:"compliance_portal"`
	}
	mc.CallToolInto("getCompliancePortal", map[string]any{
		"organization_id": orgID,
	}, &getResult)
	portalID := getResult.CompliancePortal.ID

	// Create
	var addResult struct {
		ComplianceCustomLink complianceCustomLink `json:"compliance_custom_link"`
	}
	mc.CallToolInto("addComplianceCustomLink", map[string]any{
		"compliance_portal_id": portalID,
		"name":                 "Original URL",
		"url":                  "https://example.com/original",
	}, &addResult)
	require.NotEmpty(t, addResult.ComplianceCustomLink.ID)

	// Update
	var updateResult struct {
		ComplianceCustomLink complianceCustomLink `json:"compliance_custom_link"`
	}
	mc.CallToolInto("updateComplianceCustomLink", map[string]any{
		"id":   addResult.ComplianceCustomLink.ID,
		"name": "Updated URL",
		"url":  "https://example.com/updated",
	}, &updateResult)

	assert.Equal(t, addResult.ComplianceCustomLink.ID, updateResult.ComplianceCustomLink.ID)
	assert.Equal(t, "Updated URL", updateResult.ComplianceCustomLink.Name)
}

func TestMCP_DeleteComplianceCustomLink(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get compliance portal ID
	var getResult struct {
		CompliancePortal compliancePortal `json:"compliance_portal"`
	}
	mc.CallToolInto("getCompliancePortal", map[string]any{
		"organization_id": orgID,
	}, &getResult)
	portalID := getResult.CompliancePortal.ID

	// Create
	var addResult struct {
		ComplianceCustomLink complianceCustomLink `json:"compliance_custom_link"`
	}
	mc.CallToolInto("addComplianceCustomLink", map[string]any{
		"compliance_portal_id": portalID,
		"name":                 "URL to delete",
		"url":                  "https://example.com/delete",
	}, &addResult)
	require.NotEmpty(t, addResult.ComplianceCustomLink.ID)

	// Delete
	var deleteResult struct {
		DeletedComplianceCustomLinkID string `json:"deleted_compliance_custom_link_id"`
	}
	mc.CallToolInto("deleteComplianceCustomLink", map[string]any{
		"id": addResult.ComplianceCustomLink.ID,
	}, &deleteResult)

	assert.Equal(t, addResult.ComplianceCustomLink.ID, deleteResult.DeletedComplianceCustomLinkID)
}

func TestMCP_ListComplianceCustomLinks(t *testing.T) {
	t.Parallel()
	owner := testutil.NewClient(t, testutil.RoleOwner)
	mc := testutil.NewMCPClient(t, owner)
	orgID := owner.GetOrganizationID().String()

	// Get compliance portal ID
	var getResult struct {
		CompliancePortal compliancePortal `json:"compliance_portal"`
	}
	mc.CallToolInto("getCompliancePortal", map[string]any{
		"organization_id": orgID,
	}, &getResult)
	portalID := getResult.CompliancePortal.ID

	// Create URLs
	for i := range 2 {
		var result struct {
			ComplianceCustomLink complianceCustomLink `json:"compliance_custom_link"`
		}
		mc.CallToolInto("addComplianceCustomLink", map[string]any{
			"compliance_portal_id": portalID,
			"name":                 factory.SafeName("URL"),
			"url":                  "https://example.com/" + factory.SafeName("path"),
		}, &result)
		require.NotEmpty(t, result.ComplianceCustomLink.ID)

		_ = i
	}

	// List
	var listResult struct {
		ComplianceCustomLinks []complianceCustomLink `json:"compliance_custom_links"`
	}
	mc.CallToolInto("listComplianceCustomLinks", map[string]any{
		"compliance_portal_id": portalID,
	}, &listResult)

	assert.GreaterOrEqual(t, len(listResult.ComplianceCustomLinks), 2)
}

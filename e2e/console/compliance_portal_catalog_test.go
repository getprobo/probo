// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package console_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const compliancePortalCatalogQuery = `
	query CompliancePortalCatalog($id: ID!) {
		node(id: $id) {
			... on CompliancePortal {
				id
				audits(first: 10) {
					totalCount
				}
				thirdParties(first: 10) {
					totalCount
				}
			}
		}
	}
`

const organizationCompliancePortalDocumentsQuery = `
	query OrganizationCompliancePortalDocuments($organizationId: ID!, $compliancePortalId: ID!) {
		node(id: $organizationId) {
			... on Organization {
				documents(
					first: 100
					orderBy: { field: TITLE, direction: ASC }
					filter: { status: [ACTIVE], published: true }
				) {
					edges {
						node {
							id
							compliancePortalDocument(compliancePortalId: $compliancePortalId) {
								id
								visibility
							}
						}
					}
				}
			}
		}
	}
`

func publishDocumentMinor(t *testing.T, owner *testutil.Client, documentID string) {
	t.Helper()

	err := owner.Execute(`
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				documentVersion { status }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":      true,
			"documentId": documentID,
			"changelog":  "Catalog test publish",
		},
	}, nil)
	require.NoError(t, err)
}

func TestCompliancePortal_OrganizationDocumentsExposePortalLink(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := compliancePortalID(t, owner)
	draftDocumentID := factory.NewDocument(owner).Create()
	publishedDocumentID := factory.NewDocument(owner).Create()
	publishDocumentMinor(t, owner, publishedDocumentID)

	type portalDocument struct {
		ID         string `json:"id"`
		Visibility string `json:"visibility"`
		Document   struct {
			ID string `json:"id"`
		} `json:"document"`
	}

	type queryResult struct {
		Node struct {
			Documents struct {
				Edges []struct {
					Node struct {
						ID                       string          `json:"id"`
						CompliancePortalDocument *portalDocument `json:"compliancePortalDocument"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"documents"`
		} `json:"node"`
	}

	loadDocuments := func() queryResult {
		var result queryResult
		err := owner.Execute(
			organizationCompliancePortalDocumentsQuery,
			map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"compliancePortalId": compliancePortalID,
			},
			&result,
		)
		require.NoError(t, err)
		return result
	}

	findDocument := func(result queryResult, documentID string) *portalDocument {
		for _, edge := range result.Node.Documents.Edges {
			if edge.Node.ID == documentID {
				return edge.Node.CompliancePortalDocument
			}
		}
		return nil
	}

	initialResult := loadDocuments()
	publishedDocumentFound := false
	for _, edge := range initialResult.Node.Documents.Edges {
		assert.NotEqual(t, draftDocumentID, edge.Node.ID)
		if edge.Node.ID == publishedDocumentID {
			publishedDocumentFound = true
		}
	}
	assert.True(t, publishedDocumentFound)
	assert.Nil(t, findDocument(initialResult, publishedDocumentID))

	var updateResult struct {
		UpdateCompliancePortalDocumentVisibility struct {
			CatalogDocument portalDocument `json:"catalogDocument"`
		} `json:"updateCompliancePortalDocumentVisibility"`
	}

	err := owner.Execute(`
		mutation($input: UpdateCompliancePortalDocumentVisibilityInput!) {
			updateCompliancePortalDocumentVisibility(input: $input) {
				catalogDocument {
					id
					visibility
					document { id }
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"compliancePortalId":         compliancePortalID,
			"documentId":                 publishedDocumentID,
			"compliancePortalVisibility": "RESTRICTED",
		},
	}, &updateResult)
	require.NoError(t, err)
	assert.Equal(t, publishedDocumentID, updateResult.UpdateCompliancePortalDocumentVisibility.CatalogDocument.Document.ID)

	restrictedDocument := findDocument(loadDocuments(), publishedDocumentID)
	require.NotNil(t, restrictedDocument)
	assert.Equal(t, updateResult.UpdateCompliancePortalDocumentVisibility.CatalogDocument.ID, restrictedDocument.ID)
	assert.Equal(t, "RESTRICTED", restrictedDocument.Visibility)

	err = owner.Execute(`
		mutation($input: UpdateCompliancePortalDocumentVisibilityInput!) {
			updateCompliancePortalDocumentVisibility(input: $input) {
				catalogDocument { id visibility }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"compliancePortalId":         compliancePortalID,
			"documentId":                 publishedDocumentID,
			"compliancePortalVisibility": "PUBLIC",
		},
	}, &updateResult)
	require.NoError(t, err)

	publicDocument := findDocument(loadDocuments(), publishedDocumentID)
	require.NotNil(t, publicDocument)
	assert.Equal(t, "PUBLIC", publicDocument.Visibility)

	err = owner.Execute(`
		mutation($input: DeleteCompliancePortalDocumentInput!) {
			deleteCompliancePortalDocument(input: $input) {
				deletedCompliancePortalDocumentId
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id": updateResult.UpdateCompliancePortalDocumentVisibility.CatalogDocument.ID,
		},
	}, nil)
	require.NoError(t, err)
	assert.Nil(t, findDocument(loadDocuments(), publishedDocumentID))
}

func TestCompliancePortal_CatalogRemoveDocument(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := compliancePortalID(t, owner)
	documentID := factory.NewDocument(owner).Create()
	publishDocumentMinor(t, owner, documentID)

	var addResult struct {
		UpdateCompliancePortalDocumentVisibility struct {
			CatalogDocument struct {
				ID string `json:"id"`
			} `json:"catalogDocument"`
		} `json:"updateCompliancePortalDocumentVisibility"`
	}

	err := owner.Execute(`
		mutation($input: UpdateCompliancePortalDocumentVisibilityInput!) {
			updateCompliancePortalDocumentVisibility(input: $input) {
				catalogDocument { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"compliancePortalId":         compliancePortalID,
			"documentId":                 documentID,
			"compliancePortalVisibility": "RESTRICTED",
		},
	}, &addResult)
	require.NoError(t, err)

	var removeResult struct {
		DeleteCompliancePortalDocument struct {
			DeletedCompliancePortalDocumentID string `json:"deletedCompliancePortalDocumentId"`
		} `json:"deleteCompliancePortalDocument"`
	}

	err = owner.Execute(`
		mutation($input: DeleteCompliancePortalDocumentInput!) {
			deleteCompliancePortalDocument(input: $input) {
				deletedCompliancePortalDocumentId
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id": addResult.UpdateCompliancePortalDocumentVisibility.CatalogDocument.ID,
		},
	}, &removeResult)
	require.NoError(t, err)
	assert.Equal(
		t,
		addResult.UpdateCompliancePortalDocumentVisibility.CatalogDocument.ID,
		removeResult.DeleteCompliancePortalDocument.DeletedCompliancePortalDocumentID,
	)
}

func TestCompliancePortal_CatalogTenantIsolation(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	otherOwner := testutil.NewClient(t, testutil.RoleOwner)

	compliancePortalID := compliancePortalID(t, owner)

	err := otherOwner.ExecuteShouldFail(compliancePortalCatalogQuery, map[string]any{
		"id": compliancePortalID,
	})
	require.Error(t, err)

	otherDocumentID := factory.NewDocument(otherOwner).Create()
	publishDocumentMinor(t, otherOwner, otherDocumentID)

	err = otherOwner.ExecuteShouldFail(
		organizationCompliancePortalDocumentsQuery,
		map[string]any{
			"organizationId":     otherOwner.GetOrganizationID().String(),
			"compliancePortalId": compliancePortalID,
		},
	)
	require.Error(t, err)
}

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
				documents(first: 10) {
					totalCount
					edges {
						node {
							id
							visibility
							document { id }
						}
					}
				}
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

func TestCompliancePortal_CatalogListsOnlyLinkedResources(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := compliancePortalID(t, owner)

	documentID := factory.NewDocument(owner).Create()
	publishDocumentMinor(t, owner, documentID)

	var emptyResult struct {
		Node struct {
			Documents struct {
				TotalCount int `json:"totalCount"`
			} `json:"documents"`
			Audits struct {
				TotalCount int `json:"totalCount"`
			} `json:"audits"`
			ThirdParties struct {
				TotalCount int `json:"totalCount"`
			} `json:"thirdParties"`
		} `json:"node"`
	}

	err := owner.Execute(compliancePortalCatalogQuery, map[string]any{
		"id": compliancePortalID,
	}, &emptyResult)
	require.NoError(t, err)
	assert.Equal(t, 0, emptyResult.Node.Documents.TotalCount)
	assert.Equal(t, 0, emptyResult.Node.Audits.TotalCount)
	assert.Equal(t, 0, emptyResult.Node.ThirdParties.TotalCount)

	var addResult struct {
		UpdateCompliancePortalDocumentVisibility struct {
			CatalogDocument struct {
				ID         string `json:"id"`
				Visibility string `json:"visibility"`
				Document   struct {
					ID string `json:"id"`
				} `json:"document"`
			} `json:"catalogDocument"`
		} `json:"updateCompliancePortalDocumentVisibility"`
	}

	err = owner.Execute(`
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
			"documentId":                 documentID,
			"compliancePortalVisibility": "PUBLIC",
		},
	}, &addResult)
	require.NoError(t, err)
	assert.Equal(t, documentID, addResult.UpdateCompliancePortalDocumentVisibility.CatalogDocument.Document.ID)
	assert.Equal(t, "PUBLIC", addResult.UpdateCompliancePortalDocumentVisibility.CatalogDocument.Visibility)
	assert.NotEmpty(t, addResult.UpdateCompliancePortalDocumentVisibility.CatalogDocument.ID)

	var linkedResult struct {
		Node struct {
			Documents struct {
				TotalCount int `json:"totalCount"`
			} `json:"documents"`
		} `json:"node"`
	}

	err = owner.Execute(compliancePortalCatalogQuery, map[string]any{
		"id": compliancePortalID,
	}, &linkedResult)
	require.NoError(t, err)
	assert.Equal(t, 1, linkedResult.Node.Documents.TotalCount)
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
}

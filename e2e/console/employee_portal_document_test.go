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

package console_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestEmployeePortalDocument_SharePublishedDocument(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	employee := testutil.NewClientInOrg(t, testutil.RoleEmployee, owner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	portalID := factory.DefaultEmployeePortalID(owner)

	draftID := factory.NewDocument(owner).WithTitle("Draft guideline").Create()
	publishedID := factory.NewDocument(owner).WithTitle("Shared guideline").Create()
	versionID := publishEmployeePortalDocument(t, owner, publishedID)
	unsharedID := factory.NewDocument(owner).WithTitle("Private guideline").Create()
	unsharedVersionID := publishEmployeePortalDocument(t, owner, unsharedID)

	err := owner.Execute(`
		mutation($input: AddEmployeePortalDocumentInput!) {
			addEmployeePortalDocument(input: $input) {
				employeePortalDocument { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"employeePortalId": portalID,
			"documentId":       draftID,
		},
	}, nil)
	testutil.RequireErrorCode(t, err, "INVALID", "unpublished documents cannot be shared")

	var added struct {
		AddEmployeePortalDocument struct {
			EmployeePortalDocument struct {
				ID                 string `json:"id"`
				DocumentID         string `json:"documentId"`
				Title              string `json:"title"`
				DocumentType       string `json:"documentType"`
				PublishedVersionID string `json:"publishedVersionId"`
			} `json:"employeePortalDocument"`
			Document struct {
				ID                     string `json:"id"`
				EmployeePortalDocument *struct {
					ID string `json:"id"`
				} `json:"employeePortalDocument"`
			} `json:"document"`
		} `json:"addEmployeePortalDocument"`
	}

	err = owner.Execute(`
		mutation($input: AddEmployeePortalDocumentInput!) {
			addEmployeePortalDocument(input: $input) {
				employeePortalDocument {
					id
					documentId
					title
					documentType
					publishedVersionId
				}
				document {
					id
					employeePortalDocument { id }
				}
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"employeePortalId": portalID,
			"documentId":       publishedID,
		},
	}, &added)
	require.NoError(t, err)

	link := added.AddEmployeePortalDocument.EmployeePortalDocument
	assert.Equal(t, publishedID, link.DocumentID)
	assert.Equal(t, "Shared guideline", link.Title)
	assert.Equal(t, "POLICY", link.DocumentType)
	assert.Equal(t, versionID, link.PublishedVersionID)
	require.NotNil(t, added.AddEmployeePortalDocument.Document.EmployeePortalDocument)
	assert.Equal(t, link.ID, added.AddEmployeePortalDocument.Document.EmployeePortalDocument.ID)

	err = viewer.Execute(`
		mutation($input: AddEmployeePortalDocumentInput!) {
			addEmployeePortalDocument(input: $input) {
				employeePortalDocument { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"employeePortalId": portalID,
			"documentId":       unsharedID,
		},
	}, nil)
	testutil.RequireForbiddenError(t, err, "viewers cannot share documents")

	var listed struct {
		Node struct {
			Documents struct {
				TotalCount int `json:"totalCount"`
				Edges      []struct {
					Node struct {
						ID         string `json:"id"`
						DocumentID string `json:"documentId"`
						Title      string `json:"title"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"documents"`
		} `json:"node"`
	}

	err = employee.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on EmployeePortal {
					documents(first: 20) {
						totalCount
						edges { node { id documentId title } }
					}
				}
			}
		}
	`, map[string]any{"id": portalID}, &listed)
	require.NoError(t, err)
	require.Equal(t, 1, listed.Node.Documents.TotalCount)
	require.Len(t, listed.Node.Documents.Edges, 1)
	assert.Equal(t, publishedID, listed.Node.Documents.Edges[0].Node.DocumentID)
	assert.Equal(t, "Shared guideline", listed.Node.Documents.Edges[0].Node.Title)

	var exported struct {
		ExportEmployeeDocumentVersionPDF struct {
			Data string `json:"data"`
		} `json:"exportEmployeeDocumentVersionPDF"`
	}

	err = employee.Execute(`
		mutation($input: ExportEmployeeDocumentVersionPDFInput!) {
			exportEmployeeDocumentVersionPDF(input: $input) {
				data
			}
		}
	`, map[string]any{
		"input": map[string]any{"documentVersionId": versionID},
	}, &exported)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(exported.ExportEmployeeDocumentVersionPDF.Data, "data:application/pdf;base64,"))

	err = employee.Execute(`
		mutation($input: ExportEmployeeDocumentVersionPDFInput!) {
			exportEmployeeDocumentVersionPDF(input: $input) { data }
		}
	`, map[string]any{
		"input": map[string]any{"documentVersionId": unsharedVersionID},
	}, nil)
	testutil.RequireErrorCode(t, err, "NOT_FOUND", "unshared documents cannot be exported")

	err = owner.Execute(`
		mutation($input: DeleteEmployeePortalDocumentInput!) {
			deleteEmployeePortalDocument(input: $input) {
				deletedEmployeePortalDocumentId
			}
		}
	`, map[string]any{
		"input": map[string]any{"id": link.ID},
	}, nil)
	require.NoError(t, err)

	err = owner.Execute(`
		mutation($input: DeleteEmployeePortalDocumentInput!) {
			deleteEmployeePortalDocument(input: $input) {
				deletedEmployeePortalDocumentId
			}
		}
	`, map[string]any{
		"input": map[string]any{"id": link.ID},
	}, nil)
	require.NoError(t, err)

	err = employee.Execute(`
		query($id: ID!) {
			node(id: $id) {
				... on EmployeePortal {
					documents(first: 20) {
						totalCount
						edges { node { id documentId title } }
					}
				}
			}
		}
	`, map[string]any{"id": portalID}, &listed)
	require.NoError(t, err)
	assert.Equal(t, 0, listed.Node.Documents.TotalCount)
	assert.Empty(t, listed.Node.Documents.Edges)
}

func publishEmployeePortalDocument(t *testing.T, owner *testutil.Client, documentID string) string {
	t.Helper()

	var result struct {
		PublishDocument struct {
			DocumentVersion struct {
				ID string `json:"id"`
			} `json:"documentVersion"`
		} `json:"publishDocument"`
	}

	err := owner.Execute(`
		mutation($input: PublishDocumentInput!) {
			publishDocument(input: $input) {
				documentVersion { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"minor":      true,
			"documentId": documentID,
			"changelog":  "Share with employees",
		},
	}, &result)
	require.NoError(t, err)

	return result.PublishDocument.DocumentVersion.ID
}

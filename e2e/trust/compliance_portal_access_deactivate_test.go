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

package trust_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const deactivateVisitorAccessMutation = `
	mutation($input: DeactivateCompliancePortalAccessInput!) {
		deactivateCompliancePortalAccess(input: $input) {
			compliancePortalAccess { id state }
		}
	}
`

const activateVisitorAccessMutation = `
	mutation($input: ActivateCompliancePortalAccessInput!) {
		activateCompliancePortalAccess(input: $input) {
			compliancePortalAccess { id state }
		}
	}
`

const visitorDocumentGetQuery = `
	query($id: ID!) {
		node(id: $id) {
			... on Document {
				id
				isUserAuthorized
				access { status }
			}
		}
	}
`

const visitorFileGetQuery = `
	query($id: ID!) {
		node(id: $id) {
			... on CompliancePortalFile {
				id
				isUserAuthorized
				access { status }
			}
		}
	}
`

func TestCompliancePortal_DeactivateBlocksRequestAndGrantedGet(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	documentID, compliancePortalID := setupRestrictedPortalDocument(t, owner)
	fileID := setupRestrictedPortalFile(t, owner, compliancePortalID)
	trustHost := lookupTrustHost(t, owner, compliancePortalID)

	visitor := testutil.SelfProvisionCompliancePortalVisitor(t, trustHost)
	accessID := lookupVisitorAccessID(t, owner, compliancePortalID, visitor.GetEmail())

	grantVisitorDocumentAccess(t, owner, accessID, documentID)
	grantVisitorFileAccess(t, owner, accessID, fileID)

	assertVisitorDocumentGet(t, visitor, trustHost, documentID, true, "GRANTED")
	assertVisitorFileGet(t, visitor, trustHost, fileID, true, "GRANTED")

	require.NoError(t, owner.Execute(deactivateVisitorAccessMutation, map[string]any{
		"input": map[string]any{"id": accessID},
	}, nil))

	err := visitor.ExecuteTrust(trustHost, requestAccessesMutation, map[string]any{
		"input": map[string]any{
			"documentIds":             []string{documentID},
			"reportIds":               []string{},
			"compliancePortalFileIds": []string{},
		},
	}, nil)
	require.Error(t, err, "deactivated visitor must not request restricted documents")
	assert.Contains(t, err.Error(), "user inactive")

	err = visitor.ExecuteTrust(trustHost, visitorDocumentGetQuery, map[string]any{
		"id": documentID,
	}, nil)
	require.Error(t, err, "deactivated visitor must not read granted document access")
	assert.Contains(t, err.Error(), "user inactive")

	err = visitor.ExecuteTrust(trustHost, visitorFileGetQuery, map[string]any{
		"id": fileID,
	}, nil)
	require.Error(t, err, "deactivated visitor must not read granted file access")
	assert.Contains(t, err.Error(), "user inactive")

	require.NoError(t, owner.Execute(activateVisitorAccessMutation, map[string]any{
		"input": map[string]any{"id": accessID},
	}, nil))

	assertVisitorDocumentGet(t, visitor, trustHost, documentID, true, "GRANTED")
	assertVisitorFileGet(t, visitor, trustHost, fileID, true, "GRANTED")

	var result requestAccessesResult

	err = visitor.ExecuteTrust(trustHost, requestAccessesMutation, map[string]any{
		"input": map[string]any{
			"documentIds":             []string{documentID},
			"reportIds":               []string{},
			"compliancePortalFileIds": []string{},
		},
	}, &result)
	require.NoError(t, err, "reactivated visitor must be able to request access")
}

func lookupVisitorAccessID(
	t *testing.T,
	owner *testutil.Client,
	compliancePortalID string,
	email string,
) string {
	t.Helper()

	const query = `
		query($id: ID!) {
			node(id: $id) {
				... on CompliancePortal {
					accesses(first: 50) {
						edges {
							node {
								id
								profile { emailAddress }
							}
						}
					}
				}
			}
		}
	`

	var result struct {
		Node struct {
			Accesses struct {
				Edges []struct {
					Node struct {
						ID      string `json:"id"`
						Profile struct {
							EmailAddress string `json:"emailAddress"`
						} `json:"profile"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"accesses"`
		} `json:"node"`
	}

	err := owner.Execute(query, map[string]any{"id": compliancePortalID}, &result)
	require.NoError(t, err)

	for _, edge := range result.Node.Accesses.Edges {
		if edge.Node.Profile.EmailAddress == email {
			return edge.Node.ID
		}
	}

	require.FailNowf(t, "visitor access not found", "email %s", email)

	return ""
}

func grantVisitorDocumentAccess(t *testing.T, owner *testutil.Client, accessID, documentID string) {
	t.Helper()

	err := owner.Execute(`
		mutation($input: UpdateCompliancePortalAccessInput!) {
			updateCompliancePortalAccess(input: $input) {
				compliancePortalAccess { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id": accessID,
			"documents": []map[string]any{
				{"id": documentID, "status": "GRANTED"},
			},
		},
	}, nil)
	require.NoError(t, err)
}

func grantVisitorFileAccess(t *testing.T, owner *testutil.Client, accessID, fileID string) {
	t.Helper()

	err := owner.Execute(`
		mutation($input: UpdateCompliancePortalAccessInput!) {
			updateCompliancePortalAccess(input: $input) {
				compliancePortalAccess { id }
			}
		}
	`, map[string]any{
		"input": map[string]any{
			"id": accessID,
			"compliancePortalFiles": []map[string]any{
				{"id": fileID, "status": "GRANTED"},
			},
		},
	}, nil)
	require.NoError(t, err)
}

func setupRestrictedPortalFile(
	t *testing.T,
	owner *testutil.Client,
	compliancePortalID string,
) string {
	t.Helper()

	var result struct {
		CreateCompliancePortalFile struct {
			CompliancePortalFileEdge struct {
				Node struct {
					ID string `json:"id"`
				} `json:"node"`
			} `json:"compliancePortalFileEdge"`
		} `json:"createCompliancePortalFile"`
	}

	err := owner.ExecuteWithFile(
		`
			mutation($input: CreateCompliancePortalFileInput!) {
				createCompliancePortalFile(input: $input) {
					compliancePortalFileEdge {
						node { id }
					}
				}
			}
		`,
		map[string]any{
			"input": map[string]any{
				"compliancePortalId":         compliancePortalID,
				"name":                       "Restricted portal file",
				"category":                   "Security",
				"file":                       nil,
				"compliancePortalVisibility": "RESTRICTED",
			},
		},
		"input.file",
		testutil.UploadFile{
			Filename:    "restricted.pdf",
			ContentType: "application/pdf",
			Content:     []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF"),
		},
		&result,
	)
	require.NoError(t, err)
	require.NotEmpty(t, result.CreateCompliancePortalFile.CompliancePortalFileEdge.Node.ID)

	return result.CreateCompliancePortalFile.CompliancePortalFileEdge.Node.ID
}

func assertVisitorDocumentGet(
	t *testing.T,
	visitor *testutil.Client,
	trustHost string,
	documentID string,
	authorized bool,
	status string,
) {
	t.Helper()

	var result struct {
		Node struct {
			ID               string `json:"id"`
			IsUserAuthorized bool   `json:"isUserAuthorized"`
			Access           *struct {
				Status string `json:"status"`
			} `json:"access"`
		} `json:"node"`
	}

	err := visitor.ExecuteTrust(trustHost, visitorDocumentGetQuery, map[string]any{
		"id": documentID,
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, authorized, result.Node.IsUserAuthorized)
	require.NotNil(t, result.Node.Access)
	assert.Equal(t, status, result.Node.Access.Status)
}

func assertVisitorFileGet(
	t *testing.T,
	visitor *testutil.Client,
	trustHost string,
	fileID string,
	authorized bool,
	status string,
) {
	t.Helper()

	var result struct {
		Node struct {
			ID               string `json:"id"`
			IsUserAuthorized bool   `json:"isUserAuthorized"`
			Access           *struct {
				Status string `json:"status"`
			} `json:"access"`
		} `json:"node"`
	}

	err := visitor.ExecuteTrust(trustHost, visitorFileGetQuery, map[string]any{
		"id": fileID,
	}, &result)
	require.NoError(t, err)
	assert.Equal(t, authorized, result.Node.IsUserAuthorized)
	require.NotNil(t, result.Node.Access)
	assert.Equal(t, status, result.Node.Access.Status)
}

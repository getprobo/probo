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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const visitorDocumentAccessQuery = `
	query($organizationId: ID!, $compliancePortalId: ID!, $accessId: ID!) {
		node(id: $organizationId) {
			... on Organization {
				documents(
					first: 100
					filter: { status: [ACTIVE], published: true }
				) {
					edges {
						node {
							id
							compliancePortalDocument(compliancePortalId: $compliancePortalId) {
								visibility
							}
							compliancePortalDocumentAccess(compliancePortalAccessId: $accessId) {
								id
								status
							}
						}
					}
				}
			}
		}
	}
`

func TestCompliancePortalAccess_GrantCreatesDocumentAccessWithoutReplacingSiblings(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := compliancePortalID(t, owner)
	firstDocumentID := factory.NewDocument(owner).WithTitle("Visitor access first").Create()
	secondDocumentID := factory.NewDocument(owner).WithTitle("Visitor access second").Create()
	publishDocumentMinor(t, owner, firstDocumentID)
	publishDocumentMinor(t, owner, secondDocumentID)
	restrictCompliancePortalDocument(t, owner, compliancePortalID, firstDocumentID)
	restrictCompliancePortalDocument(t, owner, compliancePortalID, secondDocumentID)

	accessID := seedCompliancePortalAccess(t, owner, compliancePortalID)

	type documentAccess struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}

	type documentNode struct {
		ID                       string `json:"id"`
		CompliancePortalDocument *struct {
			Visibility string `json:"visibility"`
		} `json:"compliancePortalDocument"`
		CompliancePortalDocumentAccess *documentAccess `json:"compliancePortalDocumentAccess"`
	}

	type queryResult struct {
		Node struct {
			Documents struct {
				Edges []struct {
					Node documentNode `json:"node"`
				} `json:"edges"`
			} `json:"documents"`
		} `json:"node"`
	}

	loadDocuments := func() map[string]documentNode {
		var result queryResult

		err := owner.Execute(
			visitorDocumentAccessQuery,
			map[string]any{
				"organizationId":     owner.GetOrganizationID().String(),
				"compliancePortalId": compliancePortalID,
				"accessId":           accessID,
			},
			&result,
		)
		require.NoError(t, err)

		byID := map[string]documentNode{}
		for _, edge := range result.Node.Documents.Edges {
			byID[edge.Node.ID] = edge.Node
		}

		return byID
	}

	initial := loadDocuments()
	require.Contains(t, initial, firstDocumentID)
	require.Contains(t, initial, secondDocumentID)
	require.NotNil(t, initial[firstDocumentID].CompliancePortalDocument)
	require.NotNil(t, initial[secondDocumentID].CompliancePortalDocument)
	assert.Equal(t, "RESTRICTED", initial[firstDocumentID].CompliancePortalDocument.Visibility)
	assert.Nil(t, initial[firstDocumentID].CompliancePortalDocumentAccess)
	assert.Nil(t, initial[secondDocumentID].CompliancePortalDocumentAccess)

	grantDocumentAccess(t, owner, accessID, firstDocumentID)

	afterFirstGrant := loadDocuments()
	require.NotNil(t, afterFirstGrant[firstDocumentID].CompliancePortalDocumentAccess)
	assert.Equal(t, "GRANTED", afterFirstGrant[firstDocumentID].CompliancePortalDocumentAccess.Status)
	accessGID, err := gid.ParseGID(afterFirstGrant[firstDocumentID].CompliancePortalDocumentAccess.ID)
	require.NoError(t, err)
	assert.Equal(t, coredata.CompliancePortalDocumentAccessEntityType, accessGID.EntityType())
	assert.Nil(t, afterFirstGrant[secondDocumentID].CompliancePortalDocumentAccess)

	grantDocumentAccess(t, owner, accessID, secondDocumentID)

	afterSecondGrant := loadDocuments()
	require.NotNil(t, afterSecondGrant[firstDocumentID].CompliancePortalDocumentAccess)
	require.NotNil(t, afterSecondGrant[secondDocumentID].CompliancePortalDocumentAccess)
	assert.Equal(t, "GRANTED", afterSecondGrant[firstDocumentID].CompliancePortalDocumentAccess.Status)
	assert.Equal(t, "GRANTED", afterSecondGrant[secondDocumentID].CompliancePortalDocumentAccess.Status)
	assert.Equal(
		t,
		afterFirstGrant[firstDocumentID].CompliancePortalDocumentAccess.ID,
		afterSecondGrant[firstDocumentID].CompliancePortalDocumentAccess.ID,
	)
}

func restrictCompliancePortalDocument(
	t *testing.T,
	owner *testutil.Client,
	compliancePortalID string,
	documentID string,
) {
	t.Helper()

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
	}, nil)
	require.NoError(t, err)
}

func grantDocumentAccess(t *testing.T, owner *testutil.Client, accessID string, documentID string) {
	t.Helper()

	err := owner.Execute(`
		mutation($input: UpdateCompliancePortalAccessInput!, $accessId: ID!) {
			updateCompliancePortalAccess(input: $input) {
				documents {
					id
					compliancePortalDocumentAccess(compliancePortalAccessId: $accessId) {
						id
						status
					}
				}
			}
		}
	`, map[string]any{
		"accessId": accessID,
		"input": map[string]any{
			"id": accessID,
			"documents": []map[string]any{
				{"id": documentID, "status": "GRANTED"},
			},
		},
	}, nil)
	require.NoError(t, err)
}

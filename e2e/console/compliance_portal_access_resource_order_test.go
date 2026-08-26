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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const listAccessResourcesQuery = `
	query($id: ID!) {
		node(id: $id) {
			... on CompliancePortalAccess {
				resources(first: 10) {
					edges {
						node {
							resourceId
							status
						}
					}
				}
			}
		}
	}
`

func TestCompliancePortalAccess_ResourcesOrder(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	compliancePortalID := compliancePortalID(t, owner)
	accessID := seedCompliancePortalAccessForIdentity(
		t,
		owner,
		compliancePortalID,
		owner.GetUserID(),
		time.Now().UTC(),
	)

	requestedID := createRestrictedPortalDocument(t, owner, compliancePortalID, "Requested resource")
	grantedID := createRestrictedPortalDocument(t, owner, compliancePortalID, "Granted resource")
	noneID := createRestrictedPortalDocument(t, owner, compliancePortalID, "No access resource")
	revokedID := createRestrictedPortalDocument(t, owner, compliancePortalID, "Revoked resource")
	rejectedID := createRestrictedPortalDocument(t, owner, compliancePortalID, "Rejected resource")

	setDocumentAccessStatus(t, owner, accessID, requestedID, "REQUESTED")
	setDocumentAccessStatus(t, owner, accessID, grantedID, "GRANTED")
	setDocumentAccessStatus(t, owner, accessID, revokedID, "REVOKED")
	setDocumentAccessStatus(t, owner, accessID, rejectedID, "REJECTED")

	var result struct {
		Node struct {
			Resources struct {
				Edges []struct {
					Node struct {
						ResourceID string  `json:"resourceId"`
						Status     *string `json:"status"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"resources"`
		} `json:"node"`
	}

	err := owner.Execute(
		listAccessResourcesQuery,
		map[string]any{"id": accessID},
		&result,
	)
	require.NoError(t, err)

	ids := make([]string, 0, len(result.Node.Resources.Edges))
	for _, edge := range result.Node.Resources.Edges {
		ids = append(ids, edge.Node.ResourceID)
	}

	require.Equal(
		t,
		[]string{requestedID, grantedID, noneID, revokedID, rejectedID},
		ids,
	)
	assert.Equal(t, "REQUESTED", *result.Node.Resources.Edges[0].Node.Status)
	assert.Equal(t, "GRANTED", *result.Node.Resources.Edges[1].Node.Status)
	assert.Nil(t, result.Node.Resources.Edges[2].Node.Status)
	assert.Equal(t, "REVOKED", *result.Node.Resources.Edges[3].Node.Status)
	assert.Equal(t, "REJECTED", *result.Node.Resources.Edges[4].Node.Status)
}

func createRestrictedPortalDocument(
	t *testing.T,
	owner *testutil.Client,
	compliancePortalID string,
	title string,
) string {
	t.Helper()

	documentID := factory.NewDocument(owner).WithTitle(title).Create()
	publishDocumentMinor(t, owner, documentID)
	restrictCompliancePortalDocument(t, owner, compliancePortalID, documentID)

	return documentID
}

func setDocumentAccessStatus(
	t *testing.T,
	owner *testutil.Client,
	accessID string,
	documentID string,
	status string,
) {
	t.Helper()

	err := owner.Execute(
		`
		mutation($input: UpdateCompliancePortalAccessInput!) {
			updateCompliancePortalAccess(input: $input) {
				compliancePortalAccess { id }
			}
		}
	`,
		map[string]any{
			"input": map[string]any{
				"id": accessID,
				"documents": []map[string]any{
					{"id": documentID, "status": status},
				},
			},
		},
		nil,
	)
	require.NoError(t, err)
}

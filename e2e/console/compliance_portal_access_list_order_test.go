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
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const listCompliancePortalAccessesQuery = `
	query($id: ID!, $orderBy: CompliancePortalAccessOrder) {
		node(id: $id) {
			... on CompliancePortal {
				accesses(first: 10, orderBy: $orderBy) {
					edges {
						node {
							id
							pendingRequestCount
						}
					}
				}
			}
		}
	}
`

type listAccessesResult struct {
	Node struct {
		Accesses struct {
			Edges []struct {
				Node struct {
					ID                  string `json:"id"`
					PendingRequestCount int    `json:"pendingRequestCount"`
				} `json:"node"`
			} `json:"edges"`
		} `json:"accesses"`
	} `json:"node"`
}

func TestCompliancePortalAccess_ListOrder(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	newerVisitor := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	compliancePortalID := compliancePortalID(t, owner)
	documentID := factory.NewDocument(owner).WithTitle("Visitor list order").Create()
	publishDocumentMinor(t, owner, documentID)
	restrictCompliancePortalDocument(t, owner, compliancePortalID, documentID)

	now := time.Now().UTC()
	olderAccessID := seedCompliancePortalAccessForIdentity(
		t,
		owner,
		compliancePortalID,
		owner.GetUserID(),
		now.Add(-2*time.Hour),
	)
	newerAccessID := seedCompliancePortalAccessForIdentity(
		t,
		owner,
		compliancePortalID,
		newerVisitor.GetUserID(),
		now,
	)
	requestDocumentAccess(t, owner, olderAccessID, documentID)

	t.Run(
		"defaults to visitors with requests first",
		func(t *testing.T) {
			t.Parallel()

			ids := listAccessIDs(t, owner, compliancePortalID, nil)
			require.GreaterOrEqual(t, len(ids), 2)
			assert.Equal(t, olderAccessID, ids[0])
			assert.Equal(t, newerAccessID, ids[1])
		},
	)

	t.Run(
		"join date newest first",
		func(t *testing.T) {
			t.Parallel()

			ids := listAccessIDs(
				t,
				owner,
				compliancePortalID,
				map[string]any{
					"field":     "CREATED_AT",
					"direction": "DESC",
				},
			)
			require.GreaterOrEqual(t, len(ids), 2)
			assert.Equal(t, newerAccessID, ids[0])
			assert.Equal(t, olderAccessID, ids[1])
		},
	)

	t.Run(
		"pending request count descending",
		func(t *testing.T) {
			t.Parallel()

			ids := listAccessIDs(
				t,
				owner,
				compliancePortalID,
				map[string]any{
					"field":     "PENDING_REQUEST_COUNT",
					"direction": "DESC",
				},
			)
			require.GreaterOrEqual(t, len(ids), 2)
			assert.Equal(t, olderAccessID, ids[0])
			assert.Equal(t, newerAccessID, ids[1])
		},
	)
}

func listAccessIDs(
	t *testing.T,
	owner *testutil.Client,
	compliancePortalID string,
	orderBy map[string]any,
) []string {
	t.Helper()

	var result listAccessesResult

	err := owner.Execute(
		listCompliancePortalAccessesQuery,
		map[string]any{
			"id":      compliancePortalID,
			"orderBy": orderBy,
		},
		&result,
	)
	require.NoError(t, err)

	ids := make([]string, 0, len(result.Node.Accesses.Edges))
	for _, edge := range result.Node.Accesses.Edges {
		ids = append(ids, edge.Node.ID)
	}

	return ids
}

func seedCompliancePortalAccessForIdentity(
	t *testing.T,
	owner *testutil.Client,
	compliancePortalID string,
	identityID gid.GID,
	createdAt time.Time,
) string {
	t.Helper()

	portalID, err := gid.ParseGID(compliancePortalID)
	require.NoError(t, err)

	tenantID := owner.GetOrganizationID().TenantID()
	accessID := gid.New(tenantID, coredata.CompliancePortalAccessEntityType)
	client := test.PGClient(t)
	ctx := context.Background()

	err = client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			_, err := conn.Exec(
				ctx,
				`
					INSERT INTO cp_accesses (
						id,
						tenant_id,
						organization_id,
						identity_id,
						compliance_portal_id,
						created_at,
						updated_at
					)
					VALUES ($1, $2, $3, $4, $5, $6, $6)
				`,
				accessID.String(),
				tenantID.String(),
				owner.GetOrganizationID().String(),
				identityID.String(),
				portalID.String(),
				createdAt,
			)

			return err
		},
	)
	require.NoError(t, err, "test setup: cannot seed cp_accesses row")

	return accessID.String()
}

func requestDocumentAccess(t *testing.T, owner *testutil.Client, accessID string, documentID string) {
	t.Helper()

	err := owner.Execute(
		`
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
	`,
		map[string]any{
			"accessId": accessID,
			"input": map[string]any{
				"id": accessID,
				"documents": []map[string]any{
					{"id": documentID, "status": "REQUESTED"},
				},
			},
		},
		nil,
	)
	require.NoError(t, err)
}

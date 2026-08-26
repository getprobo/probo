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
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/gid"
)

const listCompliancePortalAccessesFilterQuery = `
	query($id: ID!, $filter: CompliancePortalAccessFilter) {
		node(id: $id) {
			... on CompliancePortal {
				accesses(first: 10, filter: $filter) {
					edges {
						node {
							id
						}
					}
				}
			}
		}
	}
`

func TestCompliancePortalAccess_ListFilter(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	otherVisitor := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	compliancePortalID := compliancePortalID(t, owner)
	organizationID := owner.GetOrganizationID()

	now := time.Now().UTC()
	matchedAccessID := seedCompliancePortalAccessForIdentity(
		t,
		owner,
		compliancePortalID,
		owner.GetUserID(),
		now,
	)
	otherAccessID := seedCompliancePortalAccessForIdentity(
		t,
		owner,
		compliancePortalID,
		otherVisitor.GetUserID(),
		now.Add(-time.Hour),
	)

	matchedName := factory.SafeName("ZebraVisitor")
	setMembershipProfileFullName(t, organizationID, owner.GetUserID(), matchedName)
	setMembershipProfileFullName(t, organizationID, otherVisitor.GetUserID(), "")

	t.Run(
		"omitting filter returns both visitors",
		func(t *testing.T) {
			t.Parallel()

			ids := listFilteredAccessIDs(t, owner, compliancePortalID, nil)
			assert.Contains(t, ids, matchedAccessID)
			assert.Contains(t, ids, otherAccessID)
		},
	)

	t.Run(
		"query matches by full name",
		func(t *testing.T) {
			t.Parallel()

			ids := listFilteredAccessIDs(
				t,
				owner,
				compliancePortalID,
				map[string]any{"query": matchedName},
			)
			require.Len(t, ids, 1)
			assert.Equal(t, matchedAccessID, ids[0])
		},
	)

	t.Run(
		"query matches by email when full name is empty",
		func(t *testing.T) {
			t.Parallel()

			ids := listFilteredAccessIDs(
				t,
				owner,
				compliancePortalID,
				map[string]any{"query": otherVisitor.GetEmail()},
			)
			require.Len(t, ids, 1)
			assert.Equal(t, otherAccessID, ids[0])
		},
	)
}

func listFilteredAccessIDs(
	t *testing.T,
	owner *testutil.Client,
	compliancePortalID string,
	filter map[string]any,
) []string {
	t.Helper()

	var result listAccessesResult
	err := owner.Execute(
		listCompliancePortalAccessesFilterQuery,
		map[string]any{
			"id":     compliancePortalID,
			"filter": filter,
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

func setMembershipProfileFullName(
	t *testing.T,
	organizationID gid.GID,
	identityID gid.GID,
	fullName string,
) {
	t.Helper()

	client := test.PGClient(t)
	ctx := context.Background()

	err := client.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			tag, err := conn.Exec(
				ctx,
				`
					UPDATE iam_membership_profiles
					SET full_name = $1
					WHERE identity_id = $2
					AND organization_id = $3
				`,
				fullName,
				identityID.String(),
				organizationID.String(),
			)
			if err != nil {
				return err
			}

			if tag.RowsAffected() != 1 {
				return fmt.Errorf("expected 1 profile update, got %d", tag.RowsAffected())
			}

			return nil
		},
	)
	require.NoError(t, err, "test setup: cannot set membership profile full name")
}

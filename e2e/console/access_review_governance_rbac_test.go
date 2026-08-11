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
	"go.probo.inc/probo/e2e/internal/testutil"
)

func accessReviewGovernanceEntryIDForRole(
	fixture accessReviewGovernanceCampaign,
	role testutil.TestRole,
) string {
	switch role {
	case testutil.RoleOwner:
		return fixture.Entries[0].ID
	case testutil.RoleAdmin:
		return fixture.Entries[1].ID
	case testutil.RoleViewer:
		return fixture.Entries[2].ID
	default:
		return fixture.Entries[0].ID
	}
}

func TestAccessReviewGovernance_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	t.Run(
		"flag access review entry",
		func(t *testing.T) {
			t.Parallel()

			owner := org.Client(t, testutil.RoleOwner)
			fixture := startAccessReviewGovernanceCampaign(
				t,
				owner,
				"RBAC flag",
				accessReviewGovernanceCSVMultiRow,
			)
			require.Len(t, fixture.Entries, accessReviewGovernanceExpectedEntries)

			tests := []struct {
				name      string
				role      testutil.TestRole
				forbidden bool
			}{
				{name: "owner can flag", role: testutil.RoleOwner},
				{name: "admin can flag", role: testutil.RoleAdmin},
				{
					name:      "viewer cannot flag",
					role:      testutil.RoleViewer,
					forbidden: true,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						entryID := accessReviewGovernanceEntryIDForRole(fixture, tt.role)
						client := org.Client(t, tt.role)

						if tt.forbidden {
							err := flagAccessReviewEntryExpectError(
								t,
								client,
								entryID,
								[]string{"INACTIVE"},
							)
							testutil.RequireForbiddenError(
								t,
								err,
								"viewer must not flag access review entries",
							)

							return
						}

						flagged := flagAccessReviewEntryLeaf(
							t,
							client,
							entryID,
							[]string{"INACTIVE"},
							[]string{"No recent login"},
						)
						assert.Equal(t, entryID, flagged.ID)
						assert.Contains(t, flagged.Flags, "INACTIVE")
					},
				)
			}
		},
	)

	t.Run(
		"record access review entry decisions",
		func(t *testing.T) {
			t.Parallel()

			owner := org.Client(t, testutil.RoleOwner)
			fixture := startAccessReviewGovernanceCampaign(
				t,
				owner,
				"RBAC bulk decide",
				accessReviewGovernanceCSVMultiRow,
			)
			require.Len(t, fixture.Entries, accessReviewGovernanceExpectedEntries)

			tests := []struct {
				name      string
				role      testutil.TestRole
				forbidden bool
			}{
				{name: "owner can bulk decide", role: testutil.RoleOwner},
				{name: "admin can bulk decide", role: testutil.RoleAdmin},
				{
					name:      "viewer cannot bulk decide",
					role:      testutil.RoleViewer,
					forbidden: true,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						entryID := accessReviewGovernanceEntryIDForRole(fixture, tt.role)
						client := org.Client(t, tt.role)

						decisions := []map[string]any{
							{
								"accessReviewEntryId": entryID,
								"decision":            "APPROVED",
							},
						}

						if tt.forbidden {
							err := recordAccessReviewEntryDecisionsExpectError(
								t,
								client,
								decisions,
							)
							testutil.RequireForbiddenError(
								t,
								err,
								"viewer must not record access review decisions",
							)

							return
						}

						updated := recordAccessReviewEntryDecisionsLeaf(t, client, decisions)
						require.Len(t, updated, 1)
						assert.Equal(t, "APPROVED", updated[0].Decision)
						assert.True(t, updated[0].CanDecide)
					},
				)
			}
		},
	)
}

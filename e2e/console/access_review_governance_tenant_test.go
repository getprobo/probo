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

	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestAccessReviewGovernance_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	fixture := startAccessReviewGovernanceCampaign(
		t,
		org1Owner,
		"Org1 governance tenant",
		accessReviewGovernanceCSVMultiRow,
	)
	entryID := fixture.Entries[0].ID

	t.Run(
		"cannot flag entry in another organization",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := flagAccessReviewEntryExpectError(
				t,
				org2,
				entryID,
				[]string{"ORPHANED"},
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not flag access review entry from another organization",
			)
		},
	)

	t.Run(
		"cannot bulk decide entry in another organization",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := recordAccessReviewEntryDecisionsExpectError(
				t,
				org2,
				[]map[string]any{
					{
						"accessReviewEntryId": entryID,
						"decision":            "REVOKE",
					},
				},
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not record decisions for another organization's entries",
			)
		},
	)

	t.Run(
		"cannot read campaign source details from another organization",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			var result struct {
				Node *struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"node"`
			}

			err := org2.Execute(
				`
					query($id: ID!) {
						node(id: $id) {
							... on AccessReviewCampaignSource {
								id
								name
								entries(first: 1) { totalCount }
							}
						}
					}
				`,
				map[string]any{"id": fixture.CampaignSourceID},
				&result,
			)
			testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "access review campaign source")
		},
	)
}

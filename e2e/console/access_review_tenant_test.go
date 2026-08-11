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

	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestAccessReview_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1ID := org1Owner.GetOrganizationID().String()

	t.Run("cannot create access source in another organization", func(t *testing.T) {
		t.Parallel()

		const query = `
			mutation($input: CreateAccessReviewSourceInput!) {
				createAccessReviewSource(input: $input) {
					accessReviewSourceEdge {
						node { id }
					}
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"organizationId": org1ID,
				"name":           "Unauthorized Source",
			},
		})
		require.Error(t, err, "Should not be able to create access source in another organization")
	})

	t.Run("cannot create campaign in another organization", func(t *testing.T) {
		t.Parallel()

		const query = `
			mutation($input: CreateAccessReviewCampaignInput!) {
				createAccessReviewCampaign(input: $input) {
					accessReviewCampaignEdge {
						node { id }
					}
				}
			}
		`

		_, err := org2Owner.Do(query, map[string]any{
			"input": map[string]any{
				"organizationId": org1ID,
				"name":           "Unauthorized Campaign",
			},
		})
		require.Error(t, err, "Should not be able to create campaign in another organization")
	})
}

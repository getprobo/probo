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

	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestTransferImpactAssessment_PublishList_RBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)

	paID := factory.NewProcessingActivity(owner).
		WithName("TIA RBAC PA").
		WithLawfulBasis("CONSENT").
		Create()
	createTIAForPublish(t, owner, paID, "TIA RBAC")

	const query = `
		mutation($input: PublishTransferImpactAssessmentListInput!) {
			publishTransferImpactAssessmentList(input: $input) {
				documentEdge { node { id } }
			}
		}
	`

	t.Run("viewer cannot publish TIA list", func(t *testing.T) {
		t.Parallel()

		err := viewer.ExecuteShouldFail(
			query,
			map[string]any{
				"input": map[string]any{
					"minor":          false,
					"organizationId": owner.GetOrganizationID(),
				},
			},
		)
		testutil.RequireForbiddenError(t, err)
	})
}

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
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestThirdPartyAgreement_TenantIsolation(t *testing.T) {
	t.Parallel()

	org1Owner := testutil.NewClient(t, testutil.RoleOwner)
	org2Owner := testutil.NewClient(t, testutil.RoleOwner)

	org1ThirdPartyID := factory.NewThirdParty(org1Owner).
		WithName(factory.SafeName("Org1 third party agreements")).
		Create()

	uploaded := uploadThirdPartyAgreement(
		t,
		org1Owner,
		org1ThirdPartyID,
		"Org1 master service agreement",
		"MASTER_SERVICE",
		"2024-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z",
		"org1-msa.pdf",
	)
	require.NotEmpty(t, uploaded.ID)

	t.Run(
		"cannot upload an agreement for another organization's third party",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := uploadThirdPartyAgreementExpectError(
				t,
				org2,
				org1ThirdPartyID,
				"Cross tenant",
				"ANNEX",
				"2024-01-01T00:00:00Z",
				"2025-01-01T00:00:00Z",
				"cross-tenant.pdf",
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not upload an agreement for another organization's third party",
			)
		},
	)

	t.Run(
		"cannot update an agreement owned by another organization",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			_, err := org2.Do(
				`
					mutation($input: UpdateThirdPartyAgreementInput!) {
						updateThirdPartyAgreement(input: $input) {
							thirdPartyAgreement { id }
						}
					}
				`,
				map[string]any{
					"input": map[string]any{
						"id":   uploaded.ID,
						"name": "Stolen name",
					},
				},
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not update an agreement owned by another organization",
			)
		},
	)

	t.Run(
		"cannot delete an agreement owned by another organization",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			_, err := org2.Do(
				`
					mutation($input: DeleteThirdPartyAgreementInput!) {
						deleteThirdPartyAgreement(input: $input) {
							deletedThirdPartyAgreementId
						}
					}
				`,
				map[string]any{
					"input": map[string]any{
						"agreementId": uploaded.ID,
					},
				},
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not delete an agreement owned by another organization",
			)
		},
	)

	t.Run(
		"cannot create risk assessment for another organization's third party",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			_, err := org2.Do(
				`
					mutation($input: CreateThirdPartyRiskAssessmentInput!) {
						createThirdPartyRiskAssessment(input: $input) {
							thirdPartyRiskAssessmentEdge { node { id } }
						}
					}
				`,
				map[string]any{
					"input": map[string]any{
						"thirdPartyId":    org1ThirdPartyID,
						"expiresAt":       thirdPartyRiskAssessmentExpiresAt2030,
						"dataSensitivity": "LOW",
						"businessImpact":  "LOW",
					},
				},
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not create risk assessment for another organization's third party",
			)
		},
	)
}

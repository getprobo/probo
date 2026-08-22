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

	uploadedBAA := uploadThirdPartyAgreement(
		t,
		org1Owner,
		thirdPartyAgreementKindBAA,
		org1ThirdPartyID,
		"2024-01-01T00:00:00Z",
		"2025-01-01T00:00:00Z",
		"org1-baa.pdf",
	)

	uploadedDPA := uploadThirdPartyAgreement(
		t,
		org1Owner,
		thirdPartyAgreementKindDPA,
		org1ThirdPartyID,
		"2024-02-01T00:00:00Z",
		"2025-02-01T00:00:00Z",
		"org1-dpa.pdf",
	)
	require.NotEmpty(t, uploadedBAA.ID)
	require.NotEmpty(t, uploadedDPA.ID)

	t.Run(
		"cannot upload business associate agreement for another organization's third party",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := uploadThirdPartyAgreementExpectError(
				t,
				org2,
				thirdPartyAgreementKindBAA,
				org1ThirdPartyID,
				"2024-01-01T00:00:00Z",
				"2025-01-01T00:00:00Z",
				"cross-tenant-baa.pdf",
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not upload business associate agreement for another organization's third party",
			)
		},
	)

	t.Run(
		"cannot upload data privacy agreement for another organization's third party",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			err := uploadThirdPartyAgreementExpectError(
				t,
				org2,
				thirdPartyAgreementKindDPA,
				org1ThirdPartyID,
				"2024-01-01T00:00:00Z",
				"2025-01-01T00:00:00Z",
				"cross-tenant-dpa.pdf",
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not upload data privacy agreement for another organization's third party",
			)
		},
	)

	t.Run(
		"cannot update business associate agreement for another organization's third party",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			_, err := org2.Do(
				`
					mutation($input: UpdateThirdPartyBusinessAssociateAgreementInput!) {
						updateThirdPartyBusinessAssociateAgreement(input: $input) {
							thirdPartyBusinessAssociateAgreement { id }
						}
					}
				`,
				map[string]any{
					"input": map[string]any{
						"thirdPartyId": org1ThirdPartyID,
						"validity": map[string]any{
							"end": "2026-01-01T00:00:00Z",
						},
					},
				},
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not update business associate agreement for another organization's third party",
			)
		},
	)

	t.Run(
		"cannot update data privacy agreement for another organization's third party",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			_, err := org2.Do(
				`
					mutation($input: UpdateThirdPartyDataPrivacyAgreementInput!) {
						updateThirdPartyDataPrivacyAgreement(input: $input) {
							thirdPartyDataPrivacyAgreement { id }
						}
					}
				`,
				map[string]any{
					"input": map[string]any{
						"thirdPartyId": org1ThirdPartyID,
						"validity": map[string]any{
							"end": "2026-02-01T00:00:00Z",
						},
					},
				},
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not update data privacy agreement for another organization's third party",
			)
		},
	)

	t.Run(
		"cannot delete business associate agreement for another organization's third party",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			_, err := org2.Do(
				`
					mutation($input: DeleteThirdPartyBusinessAssociateAgreementInput!) {
						deleteThirdPartyBusinessAssociateAgreement(input: $input) {
							deletedThirdPartyId
						}
					}
				`,
				map[string]any{
					"input": map[string]any{
						"thirdPartyId": org1ThirdPartyID,
					},
				},
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not delete business associate agreement for another organization's third party",
			)
		},
	)

	t.Run(
		"cannot delete data privacy agreement for another organization's third party",
		func(t *testing.T) {
			t.Parallel()

			org2 := org2Owner.ForTest(t)

			_, err := org2.Do(
				`
					mutation($input: DeleteThirdPartyDataPrivacyAgreementInput!) {
						deleteThirdPartyDataPrivacyAgreement(input: $input) {
							deletedThirdPartyId
						}
					}
				`,
				map[string]any{
					"input": map[string]any{
						"thirdPartyId": org1ThirdPartyID,
					},
				},
			)
			testutil.RequireForbiddenError(
				t,
				err,
				"must not delete data privacy agreement for another organization's third party",
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

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
)

func TestThirdPartyAgreement_RBAC(t *testing.T) {
	t.Parallel()

	org := testutil.NewOrganizationRoles(t)

	t.Run(
		"upload agreement",
		func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name      string
				role      testutil.TestRole
				forbidden bool
			}{
				{
					name: "owner can upload an agreement",
					role: testutil.RoleOwner,
				},
				{
					name: "admin can upload an agreement",
					role: testutil.RoleAdmin,
				},
				{
					name:      "viewer cannot upload an agreement",
					role:      testutil.RoleViewer,
					forbidden: true,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						owner := org.Client(t, testutil.RoleOwner)
						thirdPartyID := factory.NewThirdParty(owner).
							WithName(factory.SafeName("RBAC agreement upload " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						if tt.forbidden {
							err := uploadThirdPartyAgreementExpectError(
								t,
								client,
								thirdPartyID,
								"MASTER_SERVICE",
								"2024-01-01T00:00:00Z",
								"2025-01-01T00:00:00Z",
								"agreement-rbac.pdf",
							)
							testutil.RequireForbiddenError(
								t,
								err,
								"viewer must not upload third party agreements",
							)

							return
						}

						uploaded := uploadThirdPartyAgreement(
							t,
							client,
							thirdPartyID,
							"MASTER_SERVICE",
							"2024-01-01T00:00:00Z",
							"2025-01-01T00:00:00Z",
							"agreement-rbac.pdf",
						)
						assert.NotEmpty(t, uploaded.ID)
					},
				)
			}
		},
	)

	t.Run(
		"create third party risk assessment",
		func(t *testing.T) {
			t.Parallel()

			const mutation = `
				mutation CreateThirdPartyRiskAssessment($input: CreateThirdPartyRiskAssessmentInput!) {
					createThirdPartyRiskAssessment(input: $input) {
						thirdPartyRiskAssessmentEdge {
							node { id }
						}
					}
				}
			`

			tests := []struct {
				name      string
				role      testutil.TestRole
				forbidden bool
			}{
				{
					name: "owner can create third party risk assessment",
					role: testutil.RoleOwner,
				},
				{
					name: "admin can create third party risk assessment",
					role: testutil.RoleAdmin,
				},
				{
					name:      "viewer cannot create third party risk assessment",
					role:      testutil.RoleViewer,
					forbidden: true,
				},
			}

			for _, tt := range tests {
				t.Run(
					tt.name,
					func(t *testing.T) {
						t.Parallel()

						owner := org.Client(t, testutil.RoleOwner)
						thirdPartyID := factory.NewThirdParty(owner).
							WithName(factory.SafeName("RBAC risk assessment " + string(tt.role))).
							Create()

						client := org.Client(t, tt.role)

						var result struct {
							CreateThirdPartyRiskAssessment struct {
								ThirdPartyRiskAssessmentEdge struct {
									Node struct {
										ID string `json:"id"`
									} `json:"node"`
								} `json:"thirdPartyRiskAssessmentEdge"`
							} `json:"createThirdPartyRiskAssessment"`
						}

						err := client.Execute(
							mutation,
							map[string]any{
								"input": map[string]any{
									"thirdPartyId":    thirdPartyID,
									"expiresAt":       thirdPartyRiskAssessmentExpiresAt2030,
									"dataSensitivity": "LOW",
									"businessImpact":  "LOW",
								},
							},
							&result,
						)
						if tt.forbidden {
							testutil.RequireForbiddenError(
								t,
								err,
								"viewer must not create third party risk assessments",
							)

							return
						}

						require.NoError(t, err)
						assert.NotEmpty(t, result.CreateThirdPartyRiskAssessment.ThirdPartyRiskAssessmentEdge.Node.ID)
					},
				)
			}
		},
	)

	t.Run(
		"viewer cannot update or delete an existing agreement",
		func(t *testing.T) {
			t.Parallel()

			owner := org.Client(t, testutil.RoleOwner)
			viewer := org.Client(t, testutil.RoleViewer)
			thirdPartyID := factory.NewThirdParty(owner).
				WithName(factory.SafeName("RBAC agreement mutate")).
				Create()

			uploaded := uploadThirdPartyAgreement(
				t,
				owner,
				thirdPartyID,
				"TERMS",
				"2024-01-01T00:00:00Z",
				"2025-01-01T00:00:00Z",
				"rbac-existing.pdf",
			)

			updateErr := updateThirdPartyAgreementExpectError(
				t,
				viewer,
				uploaded.ID,
				new("ANNEX"),
				new("2024-06-01T00:00:00Z"),
				new("2027-01-01T00:00:00Z"),
			)
			testutil.RequireForbiddenError(
				t,
				updateErr,
				"viewer must not update third party agreements",
			)

			_, deleteErr := viewer.Do(
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
				deleteErr,
				"viewer must not delete third party agreements",
			)
		},
	)
}

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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/factory"
	"go.probo.inc/probo/e2e/internal/testutil"
)

func TestThirdPartyBusinessAssociateAgreement_Lifecycle(t *testing.T) {
	t.Parallel()
	runThirdPartyAgreementLifecycle(t, thirdPartyAgreementKindBAA)
}

func TestThirdPartyDataPrivacyAgreement_Lifecycle(t *testing.T) {
	t.Parallel()
	runThirdPartyAgreementLifecycle(t, thirdPartyAgreementKindDPA)
}

func runThirdPartyAgreementLifecycle(t *testing.T, kind thirdPartyAgreementKind) {
	owner := testutil.NewClient(t, testutil.RoleOwner)
	thirdPartyID := factory.NewThirdParty(owner).
		WithName(factory.SafeName("Agreement lifecycle third party")).
		Create()

	const (
		initialValidFrom  = "2024-01-01T00:00:00Z"
		initialValidUntil = "2025-01-01T00:00:00Z"
		updatedValidFrom  = "2024-06-01T00:00:00Z"
		updatedValidUntil = "2026-06-01T00:00:00Z"
		fileName          = "vendor-agreement.pdf"
	)

	expectedInitialValidFrom := requireParseRFC3339Time(t, initialValidFrom)
	expectedInitialValidUntil := requireParseRFC3339Time(t, initialValidUntil)
	expectedUpdatedValidFrom := requireParseRFC3339Time(t, updatedValidFrom)
	expectedUpdatedValidUntil := requireParseRFC3339Time(t, updatedValidUntil)

	beforeCreate := time.Now().Add(-time.Second)

	uploaded := uploadThirdPartyAgreement(
		t,
		owner,
		kind,
		thirdPartyID,
		initialValidFrom,
		initialValidUntil,
		fileName,
	)

	assert.NotEmpty(t, uploaded.ID)
	assert.Equal(t, thirdPartyID, uploaded.ThirdParty.ID)
	assert.Equal(t, fileName, uploaded.File.FileName)
	assert.NotEmpty(t, uploaded.File.ID)
	require.NotNil(t, uploaded.ValidFrom)
	require.NotNil(t, uploaded.ValidUntil)
	assert.Equal(t, expectedInitialValidFrom, *uploaded.ValidFrom)
	assert.Equal(t, expectedInitialValidUntil, *uploaded.ValidUntil)
	assert.True(t, uploaded.CanUpdate)
	assert.True(t, uploaded.CanDelete)
	testutil.AssertTimestampsOnCreate(t, uploaded.CreatedAt, uploaded.UpdatedAt, beforeCreate)

	queried := queryThirdPartyAgreementField(t, owner, kind, thirdPartyID)
	require.NotNil(t, queried)
	assert.Equal(t, uploaded.ID, queried.ID)
	assert.Equal(t, thirdPartyID, queried.ThirdParty.ID)
	assert.Equal(t, uploaded.File.ID, queried.File.ID)
	assert.Equal(t, fileName, queried.File.FileName)

	updated := updateThirdPartyAgreement(
		t,
		owner,
		kind,
		thirdPartyID,
		new(updatedValidFrom),
		new(updatedValidUntil),
	)

	assert.Equal(t, uploaded.ID, updated.ID)
	require.NotNil(t, updated.ValidFrom)
	require.NotNil(t, updated.ValidUntil)
	assert.Equal(t, expectedUpdatedValidFrom, *updated.ValidFrom)
	assert.Equal(t, expectedUpdatedValidUntil, *updated.ValidUntil)
	testutil.AssertTimestampsOnUpdate(
		t,
		updated.CreatedAt,
		updated.UpdatedAt,
		uploaded.CreatedAt,
		uploaded.UpdatedAt,
	)

	deletedThirdPartyID := deleteThirdPartyAgreement(t, owner, kind, thirdPartyID)
	assert.Equal(t, thirdPartyID, deletedThirdPartyID)

	afterDelete := queryThirdPartyAgreementField(t, owner, kind, thirdPartyID)
	assert.Nil(t, afterDelete)
}

func TestThirdPartyRiskAssessment_CreateAndList(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	thirdPartyID := factory.NewThirdParty(owner).
		WithName(factory.SafeName("Risk assessment third party")).
		Create()

	const (
		dataSensitivity = "HIGH"
		businessImpact  = "MEDIUM"
		notes           = "Annual vendor review"
	)

	expectedExpiresAt := requireParseRFC3339Time(t, thirdPartyRiskAssessmentExpiresAt2030)
	beforeCreate := time.Now().Add(-time.Second)

	created := createThirdPartyRiskAssessment(
		t,
		owner,
		thirdPartyID,
		thirdPartyRiskAssessmentExpiresAt2030,
		dataSensitivity,
		businessImpact,
		notes,
	)

	assert.NotEmpty(t, created.ID)
	assert.Equal(t, thirdPartyID, created.ThirdParty.ID)
	assert.Equal(t, expectedExpiresAt, created.ExpiresAt)
	assert.Equal(t, dataSensitivity, created.DataSensitivity)
	assert.Equal(t, businessImpact, created.BusinessImpact)
	require.NotNil(t, created.Notes)
	assert.Equal(t, notes, *created.Notes)
	assert.True(t, created.CanList)
	testutil.AssertTimestampsOnCreate(t, created.CreatedAt, created.UpdatedAt, beforeCreate)

	listQuery := `
		query($id: ID!) {
			node(id: $id) {
				... on ThirdParty {
					id
					riskAssessments(first: 10) {
						edges {
							node {
								id
								expiresAt
								dataSensitivity
								businessImpact
								notes
								thirdParty { id }
								canList: permission(action: "LIST_ACTION")
							}
						}
						pageInfo {
							hasNextPage
							hasPreviousPage
						}
					}
				}
			}
		}
	`
	listQuery = strings.ReplaceAll(listQuery, "LIST_ACTION", actionThirdPartyRiskAssessmentList)

	var listResult struct {
		Node struct {
			ID              string `json:"id"`
			RiskAssessments struct {
				Edges []struct {
					Node thirdPartyRiskAssessmentWireNode `json:"node"`
				} `json:"edges"`
				PageInfo testutil.PageInfo `json:"pageInfo"`
			} `json:"riskAssessments"`
		} `json:"node"`
	}

	err := owner.Execute(listQuery, map[string]any{"id": thirdPartyID}, &listResult)
	require.NoError(t, err)

	require.Len(t, listResult.Node.RiskAssessments.Edges, 1)
	listed := listResult.Node.RiskAssessments.Edges[0].Node
	assert.Equal(t, created.ID, listed.ID)
	assert.Equal(t, thirdPartyID, listed.ThirdParty.ID)
	assert.Equal(t, expectedExpiresAt, listed.ExpiresAt)
	assert.Equal(t, dataSensitivity, listed.DataSensitivity)
	assert.Equal(t, businessImpact, listed.BusinessImpact)
	require.NotNil(t, listed.Notes)
	assert.Equal(t, notes, *listed.Notes)
	assert.True(t, listed.CanList)
	assert.False(t, listResult.Node.RiskAssessments.PageInfo.HasNextPage)
}

func TestThirdPartyAgreement_ValidDateOrdering(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	tests := []struct {
		name string
		kind thirdPartyAgreementKind
	}{
		{
			name: "business associate agreement rejects validUntil before validFrom on upload",
			kind: thirdPartyAgreementKindBAA,
		},
		{
			name: "data privacy agreement rejects validUntil before validFrom on upload",
			kind: thirdPartyAgreementKindDPA,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				client := owner.ForTest(t)
				thirdPartyID := factory.NewThirdParty(client).
					WithName(factory.SafeName("Agreement upload validation " + string(tt.kind))).
					Create()

				err := uploadThirdPartyAgreementExpectError(
					t,
					client,
					tt.kind,
					thirdPartyID,
					"2025-06-01T00:00:00Z",
					"2024-01-01T00:00:00Z",
					"invalid-dates.pdf",
				)
				testutil.RequireErrorCode(t, err, "INVALID")
			},
		)
	}

	updateTests := []struct {
		name string
		kind thirdPartyAgreementKind
	}{
		{
			name: "business associate agreement rejects validUntil before validFrom on update",
			kind: thirdPartyAgreementKindBAA,
		},
		{
			name: "data privacy agreement rejects validUntil before validFrom on update",
			kind: thirdPartyAgreementKindDPA,
		},
	}

	for _, tt := range updateTests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				client := owner.ForTest(t)
				thirdPartyID := factory.NewThirdParty(client).
					WithName(factory.SafeName("Agreement update validation " + string(tt.kind))).
					Create()

				uploadThirdPartyAgreement(
					t,
					client,
					tt.kind,
					thirdPartyID,
					"2024-01-01T00:00:00Z",
					"2025-01-01T00:00:00Z",
					"valid-agreement.pdf",
				)

				err := updateThirdPartyAgreementExpectError(
					t,
					client,
					tt.kind,
					thirdPartyID,
					new("2026-06-01T00:00:00Z"),
					new("2024-01-01T00:00:00Z"),
				)
				testutil.RequireErrorCode(t, err, "INVALID")
			},
		)
	}
}

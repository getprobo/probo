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

func TestThirdPartyAgreement_Lifecycle(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	thirdPartyID := factory.NewThirdParty(owner).
		WithName(factory.SafeName("Agreement lifecycle third party")).
		Create()

	const (
		initialValidFrom  = "2024-01-01T00:00:00Z"
		initialValidUntil = "2025-01-01T00:00:00Z"
		updatedValidFrom  = "2024-06-01T00:00:00Z"
		updatedValidUntil = "2026-06-01T00:00:00Z"
	)

	expectedInitialValidFrom := requireParseRFC3339Time(t, initialValidFrom)
	expectedInitialValidUntil := requireParseRFC3339Time(t, initialValidUntil)
	expectedUpdatedValidFrom := requireParseRFC3339Time(t, updatedValidFrom)
	expectedUpdatedValidUntil := requireParseRFC3339Time(t, updatedValidUntil)

	beforeCreate := time.Now().Add(-time.Second)

	msa := uploadThirdPartyAgreement(
		t,
		owner,
		thirdPartyID,
		"MASTER_SERVICE",
		initialValidFrom,
		initialValidUntil,
		"msa.pdf",
	)
	annex := uploadThirdPartyAgreement(
		t,
		owner,
		thirdPartyID,
		"ANNEX",
		initialValidFrom,
		initialValidUntil,
		"dora-addendum.pdf",
	)

	assert.NotEmpty(t, msa.ID)
	assert.NotEqual(t, msa.ID, annex.ID)
	assert.Equal(t, thirdPartyID, msa.ThirdParty.ID)
	assert.Equal(t, "MASTER_SERVICE", msa.Type)
	assert.Equal(t, "msa", msa.Title)
	assert.Equal(t, "msa.pdf", msa.File.FileName)
	assert.Equal(t, "ANNEX", annex.Type)
	require.NotNil(t, msa.ValidFrom())
	require.NotNil(t, msa.ValidUntil())
	assert.Equal(t, expectedInitialValidFrom, *msa.ValidFrom())
	assert.Equal(t, expectedInitialValidUntil, *msa.ValidUntil())
	assert.True(t, msa.CanUpdate)
	assert.True(t, msa.CanDelete)
	testutil.AssertTimestampsOnCreate(t, msa.CreatedAt, msa.UpdatedAt, beforeCreate)

	listed := listThirdPartyAgreements(t, owner, thirdPartyID)
	require.Len(t, listed, 2)
	assert.ElementsMatch(t, []string{msa.ID, annex.ID}, []string{listed[0].ID, listed[1].ID})

	updated := updateThirdPartyAgreement(
		t,
		owner,
		msa.ID,
		nil,
		new(updatedValidFrom),
		new(updatedValidUntil),
	)

	assert.Equal(t, msa.ID, updated.ID)
	assert.Equal(t, "msa", updated.Title)
	assert.Equal(t, "MASTER_SERVICE", updated.Type)
	require.NotNil(t, updated.ValidFrom())
	require.NotNil(t, updated.ValidUntil())
	assert.Equal(t, expectedUpdatedValidFrom, *updated.ValidFrom())
	assert.Equal(t, expectedUpdatedValidUntil, *updated.ValidUntil())
	testutil.AssertTimestampsOnUpdate(
		t,
		updated.CreatedAt,
		updated.UpdatedAt,
		msa.CreatedAt,
		msa.UpdatedAt,
	)

	retitled, err := updateThirdPartyAgreementTitle(owner, msa.ID, "Master service agreement 2026")
	require.NoError(t, err)
	assert.Equal(t, "Master service agreement 2026", retitled.Title)
	assert.Equal(t, "MASTER_SERVICE", retitled.Type)

	_, err = updateThirdPartyAgreementTitle(owner, msa.ID, "")
	require.Error(t, err)

	deletedID := deleteThirdPartyAgreement(t, owner, annex.ID)
	assert.Equal(t, annex.ID, deletedID)

	remaining := listThirdPartyAgreements(t, owner, thirdPartyID)
	require.Len(t, remaining, 1)
	assert.Equal(t, msa.ID, remaining[0].ID)
	assert.Equal(t, "Master service agreement 2026", remaining[0].Title)
	assert.Equal(t, "MASTER_SERVICE", remaining[0].Type)
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
	assert.True(t, expectedExpiresAt.Equal(listed.ExpiresAt))
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

	t.Run(
		"upload rejects validUntil before validFrom",
		func(t *testing.T) {
			t.Parallel()

			client := owner.ForTest(t)
			thirdPartyID := factory.NewThirdParty(client).
				WithName(factory.SafeName("Agreement upload validation")).
				Create()

			err := uploadThirdPartyAgreementExpectError(
				t,
				client,
				thirdPartyID,
				"OTHER",
				"2025-06-01T00:00:00Z",
				"2024-01-01T00:00:00Z",
				"invalid-dates.pdf",
			)
			testutil.RequireErrorCode(t, err, "INVALID")
		},
	)

	t.Run(
		"update rejects validUntil before validFrom",
		func(t *testing.T) {
			t.Parallel()

			client := owner.ForTest(t)
			thirdPartyID := factory.NewThirdParty(client).
				WithName(factory.SafeName("Agreement update validation")).
				Create()

			uploaded := uploadThirdPartyAgreement(
				t,
				client,
				thirdPartyID,
				"TERMS",
				"2024-01-01T00:00:00Z",
				"2025-01-01T00:00:00Z",
				"valid-agreement.pdf",
			)

			err := updateThirdPartyAgreementExpectError(
				t,
				client,
				uploaded.ID,
				nil,
				new("2026-06-01T00:00:00Z"),
				new("2024-01-01T00:00:00Z"),
			)
			testutil.RequireErrorCode(t, err, "INVALID")
		},
	)
}

func TestThirdPartyAgreement_RejectsNonPDF(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	client := owner.ForTest(t)
	thirdPartyID := factory.NewThirdParty(client).
		WithName(factory.SafeName("Agreement non-PDF upload")).
		Create()

	err := uploadThirdPartyAgreementFileExpectError(
		t,
		client,
		thirdPartyID,
		testutil.UploadFile{
			Filename:    "scan.png",
			ContentType: "image/png",
			Content:     []byte("\x89PNG\r\n\x1a\n"),
		},
	)
	testutil.RequireErrorCode(t, err, "INVALID")
}

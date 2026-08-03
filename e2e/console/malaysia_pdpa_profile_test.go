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
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/e2e/internal/testutil"
)

const (
	malaysiaPDPAProfileQuery = `
		query GetMalaysiaPDPAProfile($id: ID!) {
			node(id: $id) {
				... on Organization {
					malaysiaPDPAProfile {
						organizationId
						totalDataSubjects
						sensitiveDataSubjects
						regularSystematicMonitoring
						dpoRequired
						dpoRequirementReasons
						assessedByProfileId
						assessedAt
						dpoProfileId
						dpoAppointedAt
						commissionerNotificationDueAt
						commissionerNotifiedAt
						commissionerNotificationReference
						createdAt
						updatedAt
					}
				}
			}
		}
	`

	updateMalaysiaPDPAProfileMutation = `
		mutation UpdateMalaysiaPDPAProfile($input: UpdateMalaysiaPDPAProfileInput!) {
			updateMalaysiaPDPAProfile(input: $input) {
				malaysiaPDPAProfile {
					organizationId
					totalDataSubjects
					sensitiveDataSubjects
					regularSystematicMonitoring
					dpoRequired
					dpoRequirementReasons
					assessedByProfileId
					assessedAt
					dpoProfileId
					dpoAppointedAt
					commissionerNotificationDueAt
					commissionerNotifiedAt
					commissionerNotificationReference
					createdAt
					updatedAt
				}
			}
		}
	`
)

type malaysiaPDPAProfileResult struct {
	OrganizationID                    string     `json:"organizationId"`
	TotalDataSubjects                 int64      `json:"totalDataSubjects"`
	SensitiveDataSubjects             int64      `json:"sensitiveDataSubjects"`
	RegularSystematicMonitoring       bool       `json:"regularSystematicMonitoring"`
	DPORequired                       bool       `json:"dpoRequired"`
	DPORequirementReasons             []string   `json:"dpoRequirementReasons"`
	AssessedByProfileID               *string    `json:"assessedByProfileId"`
	AssessedAt                        *time.Time `json:"assessedAt"`
	DPOProfileID                      *string    `json:"dpoProfileId"`
	DPOAppointedAt                    *time.Time `json:"dpoAppointedAt"`
	CommissionerNotificationDueAt     *time.Time `json:"commissionerNotificationDueAt"`
	CommissionerNotifiedAt            *time.Time `json:"commissionerNotifiedAt"`
	CommissionerNotificationReference *string    `json:"commissionerNotificationReference"`
	CreatedAt                         *time.Time `json:"createdAt"`
	UpdatedAt                         *time.Time `json:"updatedAt"`
}

func TestMalaysiaPDPAProfile_UpdateAssessment(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	appointedAt := time.Date(2026, time.August, 1, 9, 0, 0, 0, time.UTC)
	notifiedAt := appointedAt.Add(24 * time.Hour)

	var result struct {
		UpdateMalaysiaPDPAProfile struct {
			MalaysiaPDPAProfile malaysiaPDPAProfileResult `json:"malaysiaPDPAProfile"`
		} `json:"updateMalaysiaPDPAProfile"`
	}

	err := owner.Execute(
		updateMalaysiaPDPAProfileMutation,
		map[string]any{
			"input": map[string]any{
				"organizationId":                    owner.GetOrganizationID().String(),
				"totalDataSubjects":                 50_000,
				"sensitiveDataSubjects":             12_000,
				"regularSystematicMonitoring":       true,
				"dpoProfileId":                      owner.GetProfileID().String(),
				"dpoAppointedAt":                    appointedAt.Format(time.RFC3339),
				"commissionerNotifiedAt":            notifiedAt.Format(time.RFC3339),
				"commissionerNotificationReference": "SPDP-2026-0001",
			},
		},
		&result,
	)
	require.NoError(t, err)

	profile := result.UpdateMalaysiaPDPAProfile.MalaysiaPDPAProfile
	assert.Equal(t, owner.GetOrganizationID().String(), profile.OrganizationID)
	assert.EqualValues(t, 50_000, profile.TotalDataSubjects)
	assert.EqualValues(t, 12_000, profile.SensitiveDataSubjects)
	assert.True(t, profile.RegularSystematicMonitoring)
	assert.True(t, profile.DPORequired)
	assert.ElementsMatch(
		t,
		[]string{
			"PERSONAL_DATA_VOLUME",
			"SENSITIVE_OR_FINANCIAL_DATA_VOLUME",
			"REGULAR_AND_SYSTEMATIC_MONITORING",
		},
		profile.DPORequirementReasons,
	)
	require.NotNil(t, profile.AssessedByProfileID)
	assert.Equal(t, owner.GetProfileID().String(), *profile.AssessedByProfileID)
	require.NotNil(t, profile.DPOProfileID)
	assert.Equal(t, owner.GetProfileID().String(), *profile.DPOProfileID)
	require.NotNil(t, profile.CommissionerNotificationDueAt)
	assert.Equal(t, appointedAt.AddDate(0, 0, 21), *profile.CommissionerNotificationDueAt)
	require.NotNil(t, profile.CreatedAt)
	require.NotNil(t, profile.UpdatedAt)
}

func TestMalaysiaPDPAProfile_ExactThresholdsDoNotTrigger(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	var result struct {
		UpdateMalaysiaPDPAProfile struct {
			MalaysiaPDPAProfile malaysiaPDPAProfileResult `json:"malaysiaPDPAProfile"`
		} `json:"updateMalaysiaPDPAProfile"`
	}

	err := owner.Execute(
		updateMalaysiaPDPAProfileMutation,
		map[string]any{
			"input": map[string]any{
				"organizationId":              owner.GetOrganizationID().String(),
				"totalDataSubjects":           20_000,
				"sensitiveDataSubjects":       10_000,
				"regularSystematicMonitoring": false,
			},
		},
		&result,
	)
	require.NoError(t, err)

	profile := result.UpdateMalaysiaPDPAProfile.MalaysiaPDPAProfile
	assert.False(t, profile.DPORequired)
	assert.Empty(t, profile.DPORequirementReasons)
}

func TestMalaysiaPDPAProfile_RBAC(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)
	viewer := testutil.NewClientInOrg(t, testutil.RoleViewer, owner)
	admin := testutil.NewClientInOrg(t, testutil.RoleAdmin, owner)

	t.Run(
		"viewer can read",
		func(t *testing.T) {
			t.Parallel()

			var result struct {
				Node *struct {
					MalaysiaPDPAProfile malaysiaPDPAProfileResult `json:"malaysiaPDPAProfile"`
				} `json:"node"`
			}

			err := viewer.Execute(
				malaysiaPDPAProfileQuery,
				map[string]any{"id": owner.GetOrganizationID().String()},
				&result,
			)
			require.NoError(t, err)
			require.NotNil(t, result.Node)
			assert.Equal(t, owner.GetOrganizationID().String(), result.Node.MalaysiaPDPAProfile.OrganizationID)
		},
	)

	t.Run(
		"viewer cannot update",
		func(t *testing.T) {
			t.Parallel()

			_, err := viewer.Do(
				updateMalaysiaPDPAProfileMutation,
				map[string]any{
					"input": map[string]any{
						"organizationId":              owner.GetOrganizationID().String(),
						"totalDataSubjects":           100,
						"sensitiveDataSubjects":       10,
						"regularSystematicMonitoring": false,
					},
				},
			)
			testutil.RequireForbiddenError(t, err, "viewer cannot update Malaysia PDPA profile")
		},
	)

	t.Run(
		"admin can update",
		func(t *testing.T) {
			t.Parallel()

			var result struct {
				UpdateMalaysiaPDPAProfile struct {
					MalaysiaPDPAProfile malaysiaPDPAProfileResult `json:"malaysiaPDPAProfile"`
				} `json:"updateMalaysiaPDPAProfile"`
			}

			err := admin.Execute(
				updateMalaysiaPDPAProfileMutation,
				map[string]any{
					"input": map[string]any{
						"organizationId":              owner.GetOrganizationID().String(),
						"totalDataSubjects":           100,
						"sensitiveDataSubjects":       10,
						"regularSystematicMonitoring": true,
					},
				},
				&result,
			)
			require.NoError(t, err)
			assert.True(t, result.UpdateMalaysiaPDPAProfile.MalaysiaPDPAProfile.DPORequired)
		},
	)
}

func TestMalaysiaPDPAProfile_TenantIsolation(t *testing.T) {
	t.Parallel()

	firstOwner := testutil.NewClient(t, testutil.RoleOwner)
	secondOwner := testutil.NewClient(t, testutil.RoleOwner)

	var result struct {
		Node *struct {
			MalaysiaPDPAProfile malaysiaPDPAProfileResult `json:"malaysiaPDPAProfile"`
		} `json:"node"`
	}

	err := secondOwner.Execute(
		malaysiaPDPAProfileQuery,
		map[string]any{"id": firstOwner.GetOrganizationID().String()},
		&result,
	)
	testutil.AssertNodeNotAccessible(t, err, result.Node == nil, "Malaysia PDPA profile")
}

func TestMalaysiaPDPAProfile_RejectsSensitiveCountAboveTotal(t *testing.T) {
	t.Parallel()

	owner := testutil.NewClient(t, testutil.RoleOwner)

	_, err := owner.Do(
		updateMalaysiaPDPAProfileMutation,
		map[string]any{
			"input": map[string]any{
				"organizationId":              owner.GetOrganizationID().String(),
				"totalDataSubjects":           100,
				"sensitiveDataSubjects":       101,
				"regularSystematicMonitoring": false,
			},
		},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sensitive_data_subjects")
}

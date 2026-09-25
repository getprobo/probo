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

package riskmanagement

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestTreatmentPlanEvent_TreatmentPlan(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	planID := gid.New(tenantID, coredata.TreatmentPlanEntityType)
	riskID := gid.New(tenantID, coredata.RiskEntityType)
	ownerID := gid.New(tenantID, coredata.MembershipProfileEntityType)
	measureID := gid.New(tenantID, coredata.MeasureEntityType)
	analysisID := gid.New(tenantID, coredata.RiskAnalysisEntityType)
	orgID := gid.New(tenantID, coredata.OrganizationEntityType)
	createdAt := time.Unix(1_700_000_000, 0).UTC()
	updatedAt := createdAt.Add(time.Hour)

	tp := &coredata.TreatmentPlan{
		ID:                 planID,
		OrganizationID:     orgID,
		RiskID:             riskID,
		RiskAnalysisID:     analysisID,
		Treatment:          coredata.RiskTreatmentMitigated,
		OwnerID:            ownerID,
		InherentLikelihood: 3,
		InherentImpact:     5,
		ResidualLikelihood: 1,
		ResidualImpact:     2,
		Category:           "Security",
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}

	event := coredata.NewTreatmentPlanEvent(
		tp,
		coredata.TreatmentPlanEventTypeUpdated,
		[]gid.GID{measureID},
		updatedAt,
	)

	assert.Equal(t, []string{measureID.String()}, event.MeasureIDs)
	assert.Equal(t, "Security", event.Category)
	linked, err := event.LinkedMeasureIDs()
	require.NoError(t, err)
	assert.Equal(t, []gid.GID{measureID}, linked)

	reconstructed := event.TreatmentPlan()
	assert.Equal(t, planID, reconstructed.ID)
	assert.Equal(t, ownerID, reconstructed.OwnerID)
	assert.Equal(t, 3, reconstructed.InherentLikelihood)
	assert.Equal(t, 5, reconstructed.InherentImpact)
	assert.Equal(t, 15, reconstructed.InherentRiskScore)
	assert.Equal(t, 2, reconstructed.ResidualRiskScore)
	assert.Equal(t, "Security", reconstructed.Category)
	assert.Equal(t, createdAt, reconstructed.CreatedAt)
	assert.Equal(t, updatedAt, reconstructed.UpdatedAt)

	entry := matrixEntryFromPlan(
		reconstructed,
		progressFromStates(
			linked,
			map[gid.GID]coredata.MeasureState{
				measureID: coredata.MeasureStateImplemented,
			},
		),
	)
	assert.Equal(t, 1, entry.NetLikelihood)
	assert.Equal(t, 2, entry.NetImpact)
}

func TestTreatmentPlanEvent_TreatmentPlan_MeasureLinkedKeepsUpdatedAt(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	createdAt := time.Unix(1_700_000_000, 0).UTC()
	updatedAt := createdAt.Add(time.Hour)
	linkedAt := updatedAt.Add(time.Hour)
	tp := &coredata.TreatmentPlan{
		ID:                 gid.New(tenantID, coredata.TreatmentPlanEntityType),
		OrganizationID:     gid.New(tenantID, coredata.OrganizationEntityType),
		RiskID:             gid.New(tenantID, coredata.RiskEntityType),
		RiskAnalysisID:     gid.New(tenantID, coredata.RiskAnalysisEntityType),
		Treatment:          coredata.RiskTreatmentMitigated,
		OwnerID:            gid.New(tenantID, coredata.MembershipProfileEntityType),
		InherentLikelihood: 2,
		InherentImpact:     3,
		ResidualLikelihood: 1,
		ResidualImpact:     1,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}

	event := coredata.NewTreatmentPlanEvent(
		tp,
		coredata.TreatmentPlanEventTypeMeasureLinked,
		nil,
		linkedAt,
	)
	reconstructed := event.TreatmentPlan()

	assert.Equal(t, createdAt, reconstructed.CreatedAt)
	assert.Equal(t, updatedAt, reconstructed.UpdatedAt)
}

func TestTreatmentPlanEvent_LinkedMeasureIDs_Invalid(t *testing.T) {
	t.Parallel()

	event := &coredata.TreatmentPlanEvent{MeasureIDs: []string{"not-a-gid"}}
	_, err := event.LinkedMeasureIDs()
	require.Error(t, err)
}

func TestMeasureEvent_Measure(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	createdAt := time.Unix(1_700_000_000, 0).UTC()
	updatedAt := createdAt.Add(time.Hour)
	measure := &coredata.Measure{
		ID:             gid.New(tenantID, coredata.MeasureEntityType),
		OrganizationID: gid.New(tenantID, coredata.OrganizationEntityType),
		Name:           "Access control",
		Category:       "Access",
		State:          coredata.MeasureStateImplemented,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	}

	event := coredata.NewMeasureEvent(measure, coredata.MeasureEventTypeUpdated, updatedAt)
	assert.Equal(t, measure.Name, event.Name)
	assert.Equal(t, measure.Category, event.Category)
	assert.Equal(t, measure.State, event.State)
	assert.Equal(t, createdAt, event.MeasureCreatedAt)
	assert.Equal(t, updatedAt, event.CreatedAt)

	reconstructed := event.Measure()
	assert.Equal(t, measure.ID, reconstructed.ID)
	assert.Equal(t, measure.Name, reconstructed.Name)
	assert.Equal(t, measure.Category, reconstructed.Category)
	assert.Equal(t, measure.State, reconstructed.State)
	assert.Equal(t, createdAt, reconstructed.CreatedAt)
	assert.Equal(t, updatedAt, reconstructed.UpdatedAt)
}

func TestMeasureIDStrings_Empty(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{}, coredata.MeasureIDStrings(nil))
	assert.Equal(t, []string{}, coredata.MeasureIDStrings([]gid.GID{}))
}

func TestProgressFromStates(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	doneID := gid.New(tenantID, coredata.MeasureEntityType)
	openID := gid.New(tenantID, coredata.MeasureEntityType)

	progress := progressFromStates(
		[]gid.GID{doneID, openID},
		map[gid.GID]coredata.MeasureState{
			doneID: coredata.MeasureStateImplemented,
			openID: coredata.MeasureStateNotStarted,
		},
	)

	assert.Equal(t, 2, progress.Total)
	assert.Equal(t, 1, progress.Done)
}

func TestProgressFromStates_MissingState(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	doneID := gid.New(tenantID, coredata.MeasureEntityType)
	missingID := gid.New(tenantID, coredata.MeasureEntityType)

	progress := progressFromStates(
		[]gid.GID{doneID, missingID},
		map[gid.GID]coredata.MeasureState{
			doneID: coredata.MeasureStateImplemented,
		},
	)

	assert.Equal(t, 2, progress.Total)
	assert.Equal(t, 1, progress.Done)
	assert.Equal(t, 0, progress.InProgress)
	assert.Equal(t, 0, progress.NotImplemented)
}

func TestCellsFromMatrixEntries(t *testing.T) {
	t.Parallel()

	entries := []RiskAnalysisMatrixEntry{
		{
			InherentLikelihood: 4,
			InherentImpact:     4,
			NetLikelihood:      1,
			NetImpact:          2,
			ResidualLikelihood: 1,
			ResidualImpact:     2,
		},
		{
			InherentLikelihood: 4,
			InherentImpact:     4,
			NetLikelihood:      4,
			NetImpact:          4,
			ResidualLikelihood: 1,
			ResidualImpact:     1,
		},
	}

	cells := cellsFromMatrixEntries(entries)
	require.Len(t, cells, 5)

	countOf := func(scoreType coredata.TreatmentPlanScoreType, likelihood, impact int) int {
		for _, cell := range cells {
			if cell.Type == scoreType && cell.Likelihood == likelihood && cell.Impact == impact {
				return cell.Count
			}
		}

		return 0
	}

	assert.Equal(t, 2, countOf(coredata.TreatmentPlanScoreTypeInherent, 4, 4))
	assert.Equal(t, 1, countOf(coredata.TreatmentPlanScoreTypeNet, 1, 2))
	assert.Equal(t, 1, countOf(coredata.TreatmentPlanScoreTypeNet, 4, 4))
	assert.Equal(t, 1, countOf(coredata.TreatmentPlanScoreTypeResidual, 1, 2))
	assert.Equal(t, 1, countOf(coredata.TreatmentPlanScoreTypeResidual, 1, 1))
}

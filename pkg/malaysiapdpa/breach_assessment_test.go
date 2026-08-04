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

package malaysiapdpa_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/malaysiapdpa"
)

func TestAssessBreachNotification(t *testing.T) {
	t.Parallel()

	awareAt := time.Date(2026, time.August, 4, 8, 30, 0, 0, time.UTC)

	t.Run("does not notify at exact scale threshold", func(t *testing.T) {
		t.Parallel()

		assessment, err := malaysiapdpa.AssessBreachNotification(
			malaysiapdpa.BreachAssessmentInput{
				AwarenessAt:          awareAt,
				AffectedDataSubjects: malaysiapdpa.SignificantScaleThreshold,
			},
		)
		require.NoError(t, err)

		assert.False(t, assessment.SignificantScale)
		assert.False(t, assessment.CommissionerNotificationRequired)
		assert.False(t, assessment.DataSubjectNotificationRequired)
		assert.Equal(t, malaysiapdpa.BreachNotificationRecommendationNotRequired, assessment.Recommendation)
		assert.Empty(t, assessment.Reasons)
		assert.Nil(t, assessment.CommissionerNotificationDueAt)
	})

	t.Run("significant scale requires Commissioner only", func(t *testing.T) {
		t.Parallel()

		assessment, err := malaysiapdpa.AssessBreachNotification(
			malaysiapdpa.BreachAssessmentInput{
				AwarenessAt:          awareAt,
				AffectedDataSubjects: malaysiapdpa.SignificantScaleThreshold + 1,
			},
		)
		require.NoError(t, err)

		assert.True(t, assessment.SignificantScale)
		assert.False(t, assessment.SignificantHarm)
		assert.True(t, assessment.CommissionerNotificationRequired)
		assert.False(t, assessment.DataSubjectNotificationRequired)
		assert.Equal(t, malaysiapdpa.BreachNotificationRecommendationCommissionerOnly, assessment.Recommendation)
		assert.Equal(t, []malaysiapdpa.BreachNotificationReason{malaysiapdpa.BreachNotificationReasonSignificantScale}, assessment.Reasons)
		require.NotNil(t, assessment.CommissionerNotificationDueAt)
		assert.Equal(t, awareAt.Add(72*time.Hour), *assessment.CommissionerNotificationDueAt)
	})

	t.Run("significant harm requires both notifications", func(t *testing.T) {
		t.Parallel()

		commissionerNotifiedAt := awareAt.Add(24 * time.Hour)
		assessment, err := malaysiapdpa.AssessBreachNotification(
			malaysiapdpa.BreachAssessmentInput{
				AwarenessAt:            awareAt,
				AffectedDataSubjects:   25,
				PotentialFinancialLoss: true,
				SensitivePersonalData:  true,
				CommissionerNotifiedAt: &commissionerNotifiedAt,
			},
		)
		require.NoError(t, err)

		assert.True(t, assessment.SignificantHarm)
		assert.False(t, assessment.SignificantScale)
		assert.True(t, assessment.CommissionerNotificationRequired)
		assert.True(t, assessment.DataSubjectNotificationRequired)
		assert.Equal(t, malaysiapdpa.BreachNotificationRecommendationCommissionerAndDataSubjects, assessment.Recommendation)
		assert.Equal(
			t,
			[]malaysiapdpa.BreachNotificationReason{
				malaysiapdpa.BreachNotificationReasonFinancialLoss,
				malaysiapdpa.BreachNotificationReasonSensitivePersonalData,
			},
			assessment.Reasons,
		)
		require.NotNil(t, assessment.DataSubjectNotificationDueAt)
		assert.Equal(t, commissionerNotifiedAt.Add(7*24*time.Hour), *assessment.DataSubjectNotificationDueAt)
		require.NotNil(t, assessment.PhasedInformationDueAt)
		assert.Equal(t, commissionerNotifiedAt.Add(30*24*time.Hour), *assessment.PhasedInformationDueAt)
	})

	t.Run("human confirmation activates deadlines without changing recommendation", func(t *testing.T) {
		t.Parallel()

		commissionerNotifiedAt := awareAt.Add(time.Hour)
		assessment, err := malaysiapdpa.AssessBreachNotification(
			malaysiapdpa.BreachAssessmentInput{
				AwarenessAt:                   awareAt,
				AffectedDataSubjects:          10,
				CommissionerNotifiedAt:        &commissionerNotifiedAt,
				HumanCommissionerNotification: true,
				HumanDataSubjectNotification:  true,
			},
		)
		require.NoError(t, err)

		assert.Equal(t, malaysiapdpa.BreachNotificationRecommendationNotRequired, assessment.Recommendation)
		assert.True(t, assessment.CommissionerNotificationRequired)
		assert.True(t, assessment.DataSubjectNotificationRequired)
		require.NotNil(t, assessment.CommissionerNotificationDueAt)
		require.NotNil(t, assessment.DataSubjectNotificationDueAt)
	})

	t.Run("rejects missing awareness time", func(t *testing.T) {
		t.Parallel()

		_, err := malaysiapdpa.AssessBreachNotification(malaysiapdpa.BreachAssessmentInput{})
		require.Error(t, err)
	})

	t.Run("rejects negative affected-subject count", func(t *testing.T) {
		t.Parallel()

		_, err := malaysiapdpa.AssessBreachNotification(
			malaysiapdpa.BreachAssessmentInput{
				AwarenessAt:          awareAt,
				AffectedDataSubjects: -1,
			},
		)
		require.Error(t, err)
	})
}

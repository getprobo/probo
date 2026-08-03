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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/malaysiapdpa"
)

func TestAssessDPORequirement_ThresholdsAndMonitoring(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		input           malaysiapdpa.DPOAssessmentInput
		required        bool
		expectedReasons []malaysiapdpa.DPORequirementReason
	}{
		{
			name: "counts exactly at both thresholds",
			input: malaysiapdpa.DPOAssessmentInput{
				TotalDataSubjects:     20_000,
				SensitiveDataSubjects: 10_000,
			},
			required:        false,
			expectedReasons: []malaysiapdpa.DPORequirementReason{},
		},
		{
			name: "personal data volume exceeds threshold",
			input: malaysiapdpa.DPOAssessmentInput{
				TotalDataSubjects:     20_001,
				SensitiveDataSubjects: 10_000,
			},
			required: true,
			expectedReasons: []malaysiapdpa.DPORequirementReason{
				malaysiapdpa.DPORequirementReasonPersonalDataVolume,
			},
		},
		{
			name: "sensitive data volume exceeds threshold",
			input: malaysiapdpa.DPOAssessmentInput{
				TotalDataSubjects:     20_000,
				SensitiveDataSubjects: 10_001,
			},
			required: true,
			expectedReasons: []malaysiapdpa.DPORequirementReason{
				malaysiapdpa.DPORequirementReasonSensitiveDataVolume,
			},
		},
		{
			name: "regular and systematic monitoring",
			input: malaysiapdpa.DPOAssessmentInput{
				TotalDataSubjects:           100,
				SensitiveDataSubjects:       10,
				RegularSystematicMonitoring: true,
			},
			required: true,
			expectedReasons: []malaysiapdpa.DPORequirementReason{
				malaysiapdpa.DPORequirementReasonRegularMonitoring,
			},
		},
		{
			name: "all criteria apply",
			input: malaysiapdpa.DPOAssessmentInput{
				TotalDataSubjects:           50_000,
				SensitiveDataSubjects:       12_000,
				RegularSystematicMonitoring: true,
			},
			required: true,
			expectedReasons: []malaysiapdpa.DPORequirementReason{
				malaysiapdpa.DPORequirementReasonPersonalDataVolume,
				malaysiapdpa.DPORequirementReasonSensitiveDataVolume,
				malaysiapdpa.DPORequirementReasonRegularMonitoring,
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assessment, err := malaysiapdpa.AssessDPORequirement(tt.input)

				require.NoError(t, err)
				assert.Equal(t, tt.required, assessment.Required)
				assert.Equal(t, tt.expectedReasons, assessment.Reasons)
			},
		)
	}
}

func TestAssessDPORequirement_InvalidCounts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       malaysiapdpa.DPOAssessmentInput
		expectedErr error
	}{
		{
			name: "negative total count",
			input: malaysiapdpa.DPOAssessmentInput{
				TotalDataSubjects:     -1,
				SensitiveDataSubjects: 0,
			},
			expectedErr: malaysiapdpa.ErrNegativeDataSubjectCount,
		},
		{
			name: "negative sensitive count",
			input: malaysiapdpa.DPOAssessmentInput{
				TotalDataSubjects:     1,
				SensitiveDataSubjects: -1,
			},
			expectedErr: malaysiapdpa.ErrNegativeDataSubjectCount,
		},
		{
			name: "sensitive count exceeds total",
			input: malaysiapdpa.DPOAssessmentInput{
				TotalDataSubjects:     100,
				SensitiveDataSubjects: 101,
			},
			expectedErr: malaysiapdpa.ErrSensitiveExceedsTotal,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assessment, err := malaysiapdpa.AssessDPORequirement(tt.input)

				require.ErrorIs(t, err, tt.expectedErr)
				assert.False(t, assessment.Required)
				assert.Empty(t, assessment.Reasons)
			},
		)
	}
}

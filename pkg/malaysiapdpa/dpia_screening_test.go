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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/malaysiapdpa"
)

func TestAssessDPIAScreening_Recommendation(t *testing.T) {
	t.Parallel()

	other := "large-scale biometric matching"
	tests := []struct {
		name           string
		input          malaysiapdpa.DPIAScreeningInput
		recommendation coredata.MalaysiaPDPADPIARecommendation
		reasons        []malaysiapdpa.DPIAScreeningReason
	}{
		{
			name: "exact numeric thresholds do not trigger",
			input: malaysiapdpa.DPIAScreeningInput{
				TotalDataSubjects:     20_000,
				SensitiveDataSubjects: 10_000,
			},
			recommendation: coredata.MalaysiaPDPADPIARecommendationNotIndicated,
			reasons:        []malaysiapdpa.DPIAScreeningReason{},
		},
		{
			name: "numeric threshold requires DPIA",
			input: malaysiapdpa.DPIAScreeningInput{
				TotalDataSubjects:     20_001,
				SensitiveDataSubjects: 10_001,
			},
			recommendation: coredata.MalaysiaPDPADPIARecommendationRequired,
			reasons: []malaysiapdpa.DPIAScreeningReason{
				malaysiapdpa.DPIAScreeningReasonPersonalDataVolume,
				malaysiapdpa.DPIAScreeningReasonSensitiveDataVolume,
			},
		},
		{
			name: "qualitative criteria require DPO review",
			input: malaysiapdpa.DPIAScreeningInput{
				TotalDataSubjects:               100,
				SensitiveDataSubjects:           10,
				SystematicMonitoring:            true,
				HighRiskAutomatedDecisionMaking: true,
				OtherHighRiskFactors:            &other,
			},
			recommendation: coredata.MalaysiaPDPADPIARecommendationDPOReviewRequired,
			reasons: []malaysiapdpa.DPIAScreeningReason{
				malaysiapdpa.DPIAScreeningReasonSystematicMonitoring,
				malaysiapdpa.DPIAScreeningReasonHighRiskAutomatedDecisions,
				malaysiapdpa.DPIAScreeningReasonOtherHighRiskFactors,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assessment, err := malaysiapdpa.AssessDPIAScreening(tt.input)

			require.NoError(t, err)
			assert.Equal(t, tt.recommendation, assessment.Recommendation)
			assert.Equal(t, tt.reasons, assessment.Reasons)
		})
	}
}

func TestAssessDPIAScreening_InvalidCounts(t *testing.T) {
	t.Parallel()

	_, err := malaysiapdpa.AssessDPIAScreening(malaysiapdpa.DPIAScreeningInput{
		TotalDataSubjects:     100,
		SensitiveDataSubjects: 101,
	})

	require.EqualError(t, err, "sensitive data subjects cannot exceed total data subjects")
}

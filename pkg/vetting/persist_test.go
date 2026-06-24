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

package vetting

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/coredata"
)

func TestBuildRiskAssessmentNotes(t *testing.T) {
	t.Parallel()

	info := ThirdPartyInfo{
		OverallRiskRating:   "Medium",
		OverallRiskScore:    62,
		Recommendation:      "APPROVE_WITH_CONDITIONS",
		SecurityRiskScore:   45,
		PrivacyRiskScore:    70,
		AIRiskScore:         10,
		ThirdPartyType:      "SAAS",
		PrivacyRole:         "PROCESSOR",
		ProcessesPII:        true,
		CrossBorderTransfer: true,
		InvolvesAI:          true,
		AIUseCases:          []string{"content generation", "fraud detection"},
		DPAStatus:           "AVAILABLE",
		DSARCapability:      "Self-service portal",
		DataLocations:       []string{"United States", "EU"},
		HumanOversight:      "Human review on flagged decisions",
		PrivacyClauses:      []string{"72-hour breach notification"},
		AIClauses:           []string{"Customer data not used for training"},
		Certifications:      []string{"SOC 2 Type II", "ISO 27001"},
		BaselineFailures:    []string{"No public DPA"},
		RiskScores: []RiskScore{
			{Category: "Security", Rating: "Medium", Notes: "Missing SOC 2"},
		},
		InformationGaps: []string{"No public DPA", "Sub-processor list inaccessible"},
	}

	notes := buildRiskAssessmentNotes(info)

	assert.Equal(
		t,
		`Automated vetting

Overall risk: 62/100 (Medium)
Recommendation: Approve with conditions

Security 45/100 · Privacy 70/100 · AI 10/100

Classification
· Type: SAAS
· Privacy role: PROCESSOR
· Processes PII: yes
· Cross-border transfers: yes
· AI involvement: yes (content generation, fraud detection)

Risk breakdown
· Security — Medium: Missing SOC 2

Privacy & data processing
· DPA: AVAILABLE
· DSAR: Self-service portal

AI governance
· Human oversight: Human review on flagged decisions

Contractual clauses
· Privacy: 72-hour breach notification
· AI: Customer data not used for training

Minimum baseline not met
· No public DPA

Gaps
· No public DPA
· Sub-processor list inaccessible`,
		notes,
	)
	assert.NotContains(t, notes, "**")
	assert.NotContains(t, notes, "#")
}

func TestBuildRiskAssessmentNotes_LimitsGaps(t *testing.T) {
	t.Parallel()

	gaps := make([]string, maxVettingNotesGaps+2)
	for i := range gaps {
		gaps[i] = "gap"
	}

	notes := buildRiskAssessmentNotes(ThirdPartyInfo{InformationGaps: gaps})

	assert.Equal(t, maxVettingNotesGaps, strings.Count(notes, "· gap"))
}

func TestFormatVettingRecommendation(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "Approve with conditions", formatVettingRecommendation("APPROVE_WITH_CONDITIONS"))
	assert.Equal(t, "Reject", formatVettingRecommendation("reject"))
}

func TestMapVettingRiskLevels(t *testing.T) {
	t.Parallel()

	assert.Equal(
		t,
		coredata.DataSensitivityNone,
		mapVettingDataSensitivity(ThirdPartyInfo{ProcessesPII: false}),
	)
	assert.Equal(
		t,
		coredata.DataSensitivityHigh,
		mapVettingDataSensitivity(ThirdPartyInfo{
			ProcessesPII:     true,
			PrivacyRiskScore: 70,
		}),
	)
	assert.Equal(
		t,
		coredata.BusinessImpactMedium,
		mapVettingBusinessImpact(ThirdPartyInfo{OverallRiskScore: 40}),
	)
	assert.Equal(
		t,
		coredata.BusinessImpactHigh,
		mapVettingBusinessImpact(ThirdPartyInfo{OverallRiskRating: "High"}),
	)
}

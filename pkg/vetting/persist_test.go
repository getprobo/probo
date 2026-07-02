// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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

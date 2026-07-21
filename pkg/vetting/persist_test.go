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

func TestBuildRiskAssessmentNotesFromResult_FiltersDocument(t *testing.T) {
	t.Parallel()

	document := `# Third party Assessment: Acme

## Executive Summary
Approve with conditions.

## Third party Classification
- Name: Acme
- Description: SaaS analytics

## Overall Risk Score
Score 70/100.

## Compliance & Certifications
- SOC 2 Type II

## Privacy & Data Processing
Retention is 30 days.

### Data Classification & Handling
PII is encrypted at rest.

### Sub-Processors
| Name | Country | Purpose |
|------|---------|---------|
| AWS | United States | Hosting |

## Security Posture
TLS looks good.

## Market Presence
Strong brand.
`

	notes := buildRiskAssessmentNotesFromResult(
		Result{
			Document: document,
			Info: ThirdPartyInfo{
				OverallRiskScore: 70,
				Recommendation:   "APPROVE_WITH_CONDITIONS",
			},
		},
	)

	assert.Contains(t, notes, "Executive Summary")
	assert.Contains(t, notes, "Approve with conditions.")
	assert.Contains(t, notes, "Overall Risk Score")
	assert.Contains(t, notes, "Score 70/100.")
	assert.Contains(t, notes, "Privacy & Data Processing")
	assert.Contains(t, notes, "Retention is 30 days.")
	assert.Contains(t, notes, "Data Classification & Handling")
	assert.Contains(t, notes, "PII is encrypted at rest.")
	assert.Contains(t, notes, "Security Posture")
	assert.Contains(t, notes, "Market Presence")

	assert.NotContains(t, notes, "Third party Classification")
	assert.NotContains(t, notes, "Name: Acme")
	assert.NotContains(t, notes, "Compliance & Certifications")
	assert.NotContains(t, notes, "SOC 2 Type II")
	assert.NotContains(t, notes, "Sub-Processors")
	assert.NotContains(t, notes, "AWS")
	assert.NotContains(t, notes, "Automated vetting")
}

func TestBuildRiskAssessmentNotesFromResult_FallsBackWhenDocumentEmpty(t *testing.T) {
	t.Parallel()

	info := ThirdPartyInfo{
		OverallRiskRating: "Medium",
		OverallRiskScore:  62,
		Recommendation:    "APPROVE_WITH_CONDITIONS",
	}

	notes := buildRiskAssessmentNotesFromResult(Result{Info: info})

	assert.Equal(t, buildRiskAssessmentNotes(info), notes)
	assert.Contains(t, notes, "Automated vetting")
}

func TestShouldDropVettingNotesSection(t *testing.T) {
	t.Parallel()

	assert.True(t, shouldDropVettingNotesSection("Third party Classification"))
	assert.True(t, shouldDropVettingNotesSection("Third-Party Classification"))
	assert.True(t, shouldDropVettingNotesSection("Vendor Classification"))
	assert.True(t, shouldDropVettingNotesSection("Vendor-Classification"))
	assert.True(t, shouldDropVettingNotesSection("Compliance & Certifications"))
	assert.True(t, shouldDropVettingNotesSection("Sub-Processors"))
	assert.True(t, shouldDropVettingNotesSection("Subprocessors"))
	assert.False(t, shouldDropVettingNotesSection("Executive Summary"))
	assert.False(t, shouldDropVettingNotesSection("Security Posture"))
	assert.False(t, shouldDropVettingNotesSection("Three-Pillar Risk Assessment"))
	assert.False(t, shouldDropVettingNotesSection("Data Classification & Handling"))
	assert.False(t, shouldDropVettingNotesSection("AI risk classifications"))
}

func TestFilterVettingDocumentNotes_IgnoresHeadingsInFences(t *testing.T) {
	t.Parallel()

	document := `# Assessment

## Security Posture
Looks good.

` + "```" + `
## Third party Classification
This is an example heading inside a fence.
` + "```" + `

## Market Presence
Strong.
`

	notes := filterVettingDocumentNotes(document)

	assert.Contains(t, notes, "Security Posture")
	assert.Contains(t, notes, "Looks good.")
	assert.Contains(t, notes, "Third party Classification")
	assert.Contains(t, notes, "This is an example heading inside a fence.")
	assert.Contains(t, notes, "Market Presence")
	assert.Contains(t, notes, "Strong.")
}

func TestFilterVettingDocumentNotes_IndentedAndTabHeadings(t *testing.T) {
	t.Parallel()

	document := `# Assessment

  ## Third party Classification
- Name: Acme

##` + "\t" + `Security Posture
TLS looks good.

## Market Presence
Strong.
`

	notes := filterVettingDocumentNotes(document)

	assert.NotContains(t, notes, "Third party Classification")
	assert.NotContains(t, notes, "Name: Acme")
	assert.Contains(t, notes, "Security Posture")
	assert.Contains(t, notes, "TLS looks good.")
	assert.Contains(t, notes, "Market Presence")
}

func TestParseMarkdownHeading(t *testing.T) {
	t.Parallel()

	level, title, ok := parseMarkdownHeading("## Executive Summary")
	assert.True(t, ok)
	assert.Equal(t, 2, level)
	assert.Equal(t, "Executive Summary", title)

	level, title, ok = parseMarkdownHeading("  ## Indented")
	assert.True(t, ok)
	assert.Equal(t, 2, level)
	assert.Equal(t, "Indented", title)

	level, title, ok = parseMarkdownHeading("##\tTabbed")
	assert.True(t, ok)
	assert.Equal(t, 2, level)
	assert.Equal(t, "Tabbed", title)

	level, title, ok = parseMarkdownHeading("## Trailing ##")
	assert.True(t, ok)
	assert.Equal(t, 2, level)
	assert.Equal(t, "Trailing", title)

	_, _, ok = parseMarkdownHeading("    ## Too indented")
	assert.False(t, ok)

	_, _, ok = parseMarkdownHeading("##NoSpace")
	assert.False(t, ok)
}

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

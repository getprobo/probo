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

package malaysiapdpa

import (
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

const (
	DPIAScreeningRuleVersion = "MY-PDPA-DPIA-2026-04"
	DPIAScreeningRuleSource  = "https://www.pdp.gov.my/ppdpv1/wp-content/uploads/2026/04/Data-Protection-Impact-Assessment-Guideline-DPIA.pdf"
)

type DPIAScreeningReason string

const (
	DPIAScreeningReasonPersonalDataVolume          DPIAScreeningReason = "PERSONAL_DATA_VOLUME"
	DPIAScreeningReasonSensitiveDataVolume         DPIAScreeningReason = "SENSITIVE_OR_FINANCIAL_DATA_VOLUME"
	DPIAScreeningReasonLegalOrSignificantEffects   DPIAScreeningReason = "LEGAL_OR_SIGNIFICANT_EFFECTS"
	DPIAScreeningReasonSystematicMonitoring        DPIAScreeningReason = "SYSTEMATIC_MONITORING"
	DPIAScreeningReasonInnovativeTechnology        DPIAScreeningReason = "INNOVATIVE_TECHNOLOGY"
	DPIAScreeningReasonRightsRestriction           DPIAScreeningReason = "RIGHTS_RESTRICTION"
	DPIAScreeningReasonLocationOrBehaviourTracking DPIAScreeningReason = "LOCATION_OR_BEHAVIOUR_TRACKING"
	DPIAScreeningReasonVulnerableDataSubjects      DPIAScreeningReason = "VULNERABLE_DATA_SUBJECTS"
	DPIAScreeningReasonHighRiskAutomatedDecisions  DPIAScreeningReason = "HIGH_RISK_AUTOMATED_DECISIONS"
	DPIAScreeningReasonOtherHighRiskFactors        DPIAScreeningReason = "OTHER_HIGH_RISK_FACTORS"
)

type (
	DPIAScreeningInput struct {
		TotalDataSubjects                int64
		SensitiveDataSubjects            int64
		LegalOrSignificantEffects        bool
		SystematicMonitoring             bool
		InnovativeTechnology             bool
		DenialOrRestrictionOfRights      bool
		LocationOrBehaviourTracking      bool
		ChildrenOrVulnerableDataSubjects bool
		HighRiskAutomatedDecisionMaking  bool
		OtherHighRiskFactors             *string
	}

	DPIAScreeningAssessment struct {
		Recommendation coredata.MalaysiaPDPADPIARecommendation
		Reasons        []DPIAScreeningReason
	}
)

func AssessDPIAScreening(input DPIAScreeningInput) (DPIAScreeningAssessment, error) {
	if input.TotalDataSubjects < 0 || input.SensitiveDataSubjects < 0 {
		return DPIAScreeningAssessment{}, fmt.Errorf("data subject counts cannot be negative")
	}

	if input.SensitiveDataSubjects > input.TotalDataSubjects {
		return DPIAScreeningAssessment{}, fmt.Errorf("sensitive data subjects cannot exceed total data subjects")
	}

	reasons := make([]DPIAScreeningReason, 0, 10)
	quantitative := false

	if input.TotalDataSubjects > PersonalDataSubjectThreshold {
		quantitative = true
		reasons = append(reasons, DPIAScreeningReasonPersonalDataVolume)
	}
	if input.SensitiveDataSubjects > SensitiveDataSubjectThreshold {
		quantitative = true
		reasons = append(reasons, DPIAScreeningReasonSensitiveDataVolume)
	}
	if input.LegalOrSignificantEffects {
		reasons = append(reasons, DPIAScreeningReasonLegalOrSignificantEffects)
	}
	if input.SystematicMonitoring {
		reasons = append(reasons, DPIAScreeningReasonSystematicMonitoring)
	}
	if input.InnovativeTechnology {
		reasons = append(reasons, DPIAScreeningReasonInnovativeTechnology)
	}
	if input.DenialOrRestrictionOfRights {
		reasons = append(reasons, DPIAScreeningReasonRightsRestriction)
	}
	if input.LocationOrBehaviourTracking {
		reasons = append(reasons, DPIAScreeningReasonLocationOrBehaviourTracking)
	}
	if input.ChildrenOrVulnerableDataSubjects {
		reasons = append(reasons, DPIAScreeningReasonVulnerableDataSubjects)
	}
	if input.HighRiskAutomatedDecisionMaking {
		reasons = append(reasons, DPIAScreeningReasonHighRiskAutomatedDecisions)
	}
	if input.OtherHighRiskFactors != nil && strings.TrimSpace(*input.OtherHighRiskFactors) != "" {
		reasons = append(reasons, DPIAScreeningReasonOtherHighRiskFactors)
	}

	recommendation := coredata.MalaysiaPDPADPIARecommendationNotIndicated
	if quantitative {
		recommendation = coredata.MalaysiaPDPADPIARecommendationRequired
	} else if len(reasons) > 0 {
		recommendation = coredata.MalaysiaPDPADPIARecommendationDPOReviewRequired
	}

	return DPIAScreeningAssessment{Recommendation: recommendation, Reasons: reasons}, nil
}

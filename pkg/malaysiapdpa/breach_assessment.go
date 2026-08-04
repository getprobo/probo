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
	"time"
)

const (
	BreachNotificationRuleVersion = "MY-PDPA-DBN-2025-02-25"
	BreachNotificationRuleSource  = "https://www.pdp.gov.my/ppdpv1/wp-content/uploads/2025/08/GP_DBN_ENG.pdf"

	SignificantScaleThreshold int64 = 1000

	CommissionerNotificationWindow = 72 * time.Hour
	DataSubjectNotificationWindow  = 7 * 24 * time.Hour
	PhasedInformationWindow        = 30 * 24 * time.Hour
)

type BreachNotificationReason string

const (
	BreachNotificationReasonPhysicalHarm           BreachNotificationReason = "PHYSICAL_HARM"
	BreachNotificationReasonFinancialLoss          BreachNotificationReason = "FINANCIAL_LOSS"
	BreachNotificationReasonCreditOrPropertyDamage BreachNotificationReason = "CREDIT_OR_PROPERTY_DAMAGE"
	BreachNotificationReasonIllegalUse             BreachNotificationReason = "ILLEGAL_USE"
	BreachNotificationReasonSensitivePersonalData  BreachNotificationReason = "SENSITIVE_PERSONAL_DATA"
	BreachNotificationReasonIdentityFraud          BreachNotificationReason = "IDENTITY_FRAUD"
	BreachNotificationReasonSignificantScale       BreachNotificationReason = "SIGNIFICANT_SCALE"
)

type BreachNotificationRecommendation string

const (
	BreachNotificationRecommendationNotRequired                 BreachNotificationRecommendation = "NOT_REQUIRED"
	BreachNotificationRecommendationCommissionerOnly            BreachNotificationRecommendation = "COMMISSIONER_ONLY"
	BreachNotificationRecommendationCommissionerAndDataSubjects BreachNotificationRecommendation = "COMMISSIONER_AND_DATA_SUBJECTS"
)

type (
	BreachAssessmentInput struct {
		AwarenessAt                     time.Time
		AffectedDataSubjects            int64
		PotentialPhysicalHarm           bool
		PotentialFinancialLoss          bool
		PotentialCreditOrPropertyDamage bool
		PotentialIllegalUse             bool
		SensitivePersonalData           bool
		PotentialIdentityFraud          bool
		CommissionerNotifiedAt          *time.Time
		HumanCommissionerNotification   bool
		HumanDataSubjectNotification    bool
	}

	BreachAssessment struct {
		SignificantHarm                  bool
		SignificantScale                 bool
		CommissionerNotificationRequired bool
		DataSubjectNotificationRequired  bool
		Recommendation                   BreachNotificationRecommendation
		Reasons                          []BreachNotificationReason
		CommissionerNotificationDueAt    *time.Time
		DataSubjectNotificationDueAt     *time.Time
		PhasedInformationDueAt           *time.Time
	}
)

func AssessBreachNotification(input BreachAssessmentInput) (BreachAssessment, error) {
	if input.AwarenessAt.IsZero() {
		return BreachAssessment{}, fmt.Errorf("awareness time is required")
	}

	if input.AffectedDataSubjects < 0 {
		return BreachAssessment{}, fmt.Errorf("affected data subjects cannot be negative")
	}

	reasons := make([]BreachNotificationReason, 0, 7)
	significantHarm := false

	if input.PotentialPhysicalHarm {
		significantHarm = true
		reasons = append(reasons, BreachNotificationReasonPhysicalHarm)
	}
	if input.PotentialFinancialLoss {
		significantHarm = true
		reasons = append(reasons, BreachNotificationReasonFinancialLoss)
	}
	if input.PotentialCreditOrPropertyDamage {
		significantHarm = true
		reasons = append(reasons, BreachNotificationReasonCreditOrPropertyDamage)
	}
	if input.PotentialIllegalUse {
		significantHarm = true
		reasons = append(reasons, BreachNotificationReasonIllegalUse)
	}
	if input.SensitivePersonalData {
		significantHarm = true
		reasons = append(reasons, BreachNotificationReasonSensitivePersonalData)
	}
	if input.PotentialIdentityFraud {
		significantHarm = true
		reasons = append(reasons, BreachNotificationReasonIdentityFraud)
	}

	significantScale := input.AffectedDataSubjects > SignificantScaleThreshold
	if significantScale {
		reasons = append(reasons, BreachNotificationReasonSignificantScale)
	}

	commissionerRecommended := significantHarm || significantScale
	dataSubjectsRecommended := significantHarm
	commissionerRequired := commissionerRecommended || input.HumanCommissionerNotification
	dataSubjectsRequired := dataSubjectsRecommended || input.HumanDataSubjectNotification

	recommendation := BreachNotificationRecommendationNotRequired
	if dataSubjectsRecommended {
		recommendation = BreachNotificationRecommendationCommissionerAndDataSubjects
	} else if commissionerRecommended {
		recommendation = BreachNotificationRecommendationCommissionerOnly
	}

	assessment := BreachAssessment{
		SignificantHarm:                  significantHarm,
		SignificantScale:                 significantScale,
		CommissionerNotificationRequired: commissionerRequired,
		DataSubjectNotificationRequired:  dataSubjectsRequired,
		Recommendation:                   recommendation,
		Reasons:                          reasons,
	}

	if commissionerRequired {
		dueAt := input.AwarenessAt.Add(CommissionerNotificationWindow)
		assessment.CommissionerNotificationDueAt = &dueAt
	}

	if input.CommissionerNotifiedAt != nil && commissionerRequired {
		phasedDueAt := input.CommissionerNotifiedAt.Add(PhasedInformationWindow)
		assessment.PhasedInformationDueAt = &phasedDueAt

		if dataSubjectsRequired {
			dataSubjectsDueAt := input.CommissionerNotifiedAt.Add(DataSubjectNotificationWindow)
			assessment.DataSubjectNotificationDueAt = &dataSubjectsDueAt
		}
	}

	return assessment, nil
}

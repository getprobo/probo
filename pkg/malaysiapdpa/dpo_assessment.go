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

import "errors"

type (
	DPORequirementReason string

	DPOAssessmentInput struct {
		TotalDataSubjects           int64
		SensitiveDataSubjects       int64
		RegularSystematicMonitoring bool
	}

	DPOAssessment struct {
		Required bool
		Reasons  []DPORequirementReason
	}
)

const (
	PersonalDataSubjectThreshold            int64                = 20_000
	SensitiveDataSubjectThreshold           int64                = 10_000
	DPORequirementReasonPersonalDataVolume  DPORequirementReason = "PERSONAL_DATA_VOLUME"
	DPORequirementReasonSensitiveDataVolume DPORequirementReason = "SENSITIVE_OR_FINANCIAL_DATA_VOLUME"
	DPORequirementReasonRegularMonitoring   DPORequirementReason = "REGULAR_AND_SYSTEMATIC_MONITORING"
)

var (
	ErrNegativeDataSubjectCount = errors.New("data subject counts cannot be negative")
	ErrSensitiveExceedsTotal    = errors.New("sensitive data subjects cannot exceed total data subjects")
)

func AssessDPORequirement(input DPOAssessmentInput) (DPOAssessment, error) {
	if input.TotalDataSubjects < 0 || input.SensitiveDataSubjects < 0 {
		return DPOAssessment{}, ErrNegativeDataSubjectCount
	}

	if input.SensitiveDataSubjects > input.TotalDataSubjects {
		return DPOAssessment{}, ErrSensitiveExceedsTotal
	}

	reasons := make([]DPORequirementReason, 0, 3)

	if input.TotalDataSubjects > PersonalDataSubjectThreshold {
		reasons = append(reasons, DPORequirementReasonPersonalDataVolume)
	}

	if input.SensitiveDataSubjects > SensitiveDataSubjectThreshold {
		reasons = append(reasons, DPORequirementReasonSensitiveDataVolume)
	}

	if input.RegularSystematicMonitoring {
		reasons = append(reasons, DPORequirementReasonRegularMonitoring)
	}

	return DPOAssessment{
		Required: len(reasons) > 0,
		Reasons:  reasons,
	}, nil
}

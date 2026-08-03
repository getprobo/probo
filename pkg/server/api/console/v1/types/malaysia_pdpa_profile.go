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

package types

import (
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/malaysiapdpa"
)

func NewMalaysiaPDPAProfile(p *coredata.MalaysiaPDPAProfile) *MalaysiaPDPAProfile {
	reasons := make([]malaysiapdpa.DPORequirementReason, len(p.DPORequirementReasons))
	for i, reason := range p.DPORequirementReasons {
		reasons[i] = malaysiapdpa.DPORequirementReason(reason)
	}

	var commissionerNotificationDueAt *time.Time
	if p.DPOAppointedAt != nil {
		dueAt := p.DPOAppointedAt.AddDate(0, 0, 21)
		commissionerNotificationDueAt = &dueAt
	}

	var createdAt *time.Time
	if !p.CreatedAt.IsZero() {
		createdAt = &p.CreatedAt
	}

	var updatedAt *time.Time
	if !p.UpdatedAt.IsZero() {
		updatedAt = &p.UpdatedAt
	}

	return &MalaysiaPDPAProfile{
		OrganizationID:                    p.OrganizationID,
		TotalDataSubjects:                 p.TotalDataSubjects,
		SensitiveDataSubjects:             p.SensitiveDataSubjects,
		RegularSystematicMonitoring:       p.RegularSystematicMonitoring,
		DPORequired:                       p.DPORequired,
		DPORequirementReasons:             reasons,
		AssessedByProfileID:               p.AssessedByProfileID,
		AssessedAt:                        p.AssessedAt,
		DPOProfileID:                      p.DPOProfileID,
		DPOAppointedAt:                    p.DPOAppointedAt,
		CommissionerNotificationDueAt:     commissionerNotificationDueAt,
		CommissionerNotifiedAt:            p.CommissionerNotifiedAt,
		CommissionerNotificationReference: p.CommissionerNotificationReference,
		CreatedAt:                         createdAt,
		UpdatedAt:                         updatedAt,
	}
}

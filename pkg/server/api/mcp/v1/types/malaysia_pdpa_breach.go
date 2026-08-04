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
	"go.probo.inc/probo/pkg/page"
)

func NewMalaysiaPDPABreachIncident(incident *coredata.MalaysiaPDPABreachIncident) *MalaysiaPDPABreachIncident {
	reasons := make([]malaysiapdpa.BreachNotificationReason, len(incident.NotificationReasons))
	for index, reason := range incident.NotificationReasons {
		reasons[index] = malaysiapdpa.BreachNotificationReason(reason)
	}

	assessment, _ := malaysiapdpa.AssessBreachNotification(malaysiapdpa.BreachAssessmentInput{
		AwarenessAt:                     incident.AwarenessAt,
		AffectedDataSubjects:            incident.AffectedDataSubjects,
		PotentialPhysicalHarm:           incident.PotentialPhysicalHarm,
		PotentialFinancialLoss:          incident.PotentialFinancialLoss,
		PotentialCreditOrPropertyDamage: incident.PotentialCreditOrPropertyDamage,
		PotentialIllegalUse:             incident.PotentialIllegalUse,
		SensitivePersonalData:           incident.SensitivePersonalData,
		PotentialIdentityFraud:          incident.PotentialIdentityFraud,
		CommissionerNotifiedAt:          incident.CommissionerNotifiedAt,
		HumanCommissionerNotification: incident.NotificationDecision == coredata.MalaysiaPDPABreachNotificationDecisionCommissionerOnly ||
			incident.NotificationDecision == coredata.MalaysiaPDPABreachNotificationDecisionCommissionerAndDataSubjects,
		HumanDataSubjectNotification: incident.NotificationDecision == coredata.MalaysiaPDPABreachNotificationDecisionCommissionerAndDataSubjects,
	})

	return &MalaysiaPDPABreachIncident{
		ID:                                 incident.ID,
		OrganizationID:                     incident.OrganizationID,
		Title:                              incident.Title,
		Description:                        incident.Description,
		OccurredAt:                         incident.OccurredAt,
		DiscoveredAt:                       incident.DiscoveredAt,
		AwarenessAt:                        incident.AwarenessAt,
		AffectedDataSubjects:               int(incident.AffectedDataSubjects),
		AffectedDataRecords:                int(incident.AffectedDataRecords),
		PersonalDataTypes:                  incident.PersonalDataTypes,
		AffectedSystem:                     incident.AffectedSystem,
		LikelyConsequences:                 incident.LikelyConsequences,
		ContainmentActions:                 incident.ContainmentActions,
		PotentialPhysicalHarm:              incident.PotentialPhysicalHarm,
		PotentialFinancialLoss:             incident.PotentialFinancialLoss,
		PotentialCreditOrPropertyDamage:    incident.PotentialCreditOrPropertyDamage,
		PotentialIllegalUse:                incident.PotentialIllegalUse,
		SensitivePersonalData:              incident.SensitivePersonalData,
		PotentialIdentityFraud:             incident.PotentialIdentityFraud,
		SignificantHarm:                    incident.SignificantHarm,
		SignificantScale:                   incident.SignificantScale,
		NotificationRecommendation:         incident.NotificationRecommendation,
		NotificationReasons:                reasons,
		NotificationDecision:               incident.NotificationDecision,
		DecisionRationale:                  incident.DecisionRationale,
		DecisionEvidence:                   incident.DecisionEvidence,
		AssessedByProfileID:                incident.AssessedByProfileID,
		AssessedAt:                         incident.AssessedAt,
		RuleVersion:                        incident.RuleVersion,
		RuleSource:                         incident.RuleSource,
		CommissionerNotificationDueAt:      assessment.CommissionerNotificationDueAt,
		CommissionerNotificationOverdue:    malaysiaPDPABreachDeadlineOverdue(assessment.CommissionerNotificationDueAt, incident.CommissionerNotifiedAt),
		CommissionerNotifiedAt:             incident.CommissionerNotifiedAt,
		CommissionerNotificationReference:  incident.CommissionerNotificationReference,
		CommissionerConfirmationReceivedAt: incident.CommissionerConfirmationReceivedAt,
		CommissionerConfirmationReference:  incident.CommissionerConfirmationReference,
		PhasedInformationDueAt:             assessment.PhasedInformationDueAt,
		DelayedNotificationReason:          incident.DelayedNotificationReason,
		DelayedNotificationEvidence:        incident.DelayedNotificationEvidence,
		DataSubjectsNotificationDueAt:      assessment.DataSubjectNotificationDueAt,
		DataSubjectsNotificationOverdue:    malaysiaPDPABreachDeadlineOverdue(assessment.DataSubjectNotificationDueAt, incident.DataSubjectsNotifiedAt),
		DataSubjectsNotifiedAt:             incident.DataSubjectsNotifiedAt,
		DataSubjectsNotificationEvidence:   incident.DataSubjectsNotificationEvidence,
		Status:                             incident.Status,
		CreatedByProfileID:                 incident.CreatedByProfileID,
		CreatedAt:                          incident.CreatedAt,
		UpdatedAt:                          incident.UpdatedAt,
	}
}

func NewMalaysiaPDPABreachStatusHistory(history *coredata.MalaysiaPDPABreachStatusHistory) *MalaysiaPDPABreachStatusHistory {
	return &MalaysiaPDPABreachStatusHistory{
		ID:                 history.ID,
		OrganizationID:     history.OrganizationID,
		IncidentID:         history.IncidentID,
		FromStatus:         history.FromStatus,
		ToStatus:           history.ToStatus,
		ChangedByProfileID: history.ChangedByProfileID,
		Reason:             history.Reason,
		CreatedAt:          history.CreatedAt,
	}
}

func NewListMalaysiaPDPABreachIncidentsOutput(
	p *page.Page[*coredata.MalaysiaPDPABreachIncident, coredata.MalaysiaPDPABreachOrderField],
) ListMalaysiaPDPABreachIncidentsOutput {
	incidents := make([]*MalaysiaPDPABreachIncident, len(p.Data))
	for index, incident := range p.Data {
		incidents[index] = NewMalaysiaPDPABreachIncident(incident)
	}

	var nextCursor *page.CursorKey
	if len(p.Data) > 0 {
		key := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &key
	}

	return ListMalaysiaPDPABreachIncidentsOutput{
		Incidents:  incidents,
		NextCursor: nextCursor,
	}
}

func NewListMalaysiaPDPABreachStatusHistoryOutput(
	p *page.Page[*coredata.MalaysiaPDPABreachStatusHistory, coredata.MalaysiaPDPABreachStatusHistoryOrderField],
) ListMalaysiaPDPABreachStatusHistoryOutput {
	history := make([]*MalaysiaPDPABreachStatusHistory, len(p.Data))
	for index, entry := range p.Data {
		history[index] = NewMalaysiaPDPABreachStatusHistory(entry)
	}

	var nextCursor *page.CursorKey
	if len(p.Data) > 0 {
		key := p.Data[len(p.Data)-1].CursorKey(p.Cursor.OrderBy.Field)
		nextCursor = &key
	}

	return ListMalaysiaPDPABreachStatusHistoryOutput{
		History:    history,
		NextCursor: nextCursor,
	}
}

func malaysiaPDPABreachDeadlineOverdue(dueAt, completedAt *time.Time) bool {
	if dueAt == nil {
		return false
	}

	if completedAt != nil {
		return completedAt.After(*dueAt)
	}

	return time.Now().After(*dueAt)
}

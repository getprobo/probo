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
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/malaysiapdpa"
	"go.probo.inc/probo/pkg/page"
)

type (
	MalaysiaPDPABreachOrderBy              OrderBy[coredata.MalaysiaPDPABreachOrderField]
	MalaysiaPDPABreachStatusHistoryOrderBy OrderBy[coredata.MalaysiaPDPABreachStatusHistoryOrderField]

	MalaysiaPDPABreachIncidentConnection struct {
		TotalCount int
		Edges      []*MalaysiaPDPABreachIncidentEdge
		PageInfo   PageInfo

		Resolver any
		ParentID gid.GID
	}

	MalaysiaPDPABreachStatusHistoryConnection struct {
		TotalCount int
		Edges      []*MalaysiaPDPABreachStatusHistoryEdge
		PageInfo   PageInfo

		IncidentID gid.GID
	}
)

func NewMalaysiaPDPABreachIncidentConnection(
	p *page.Page[*coredata.MalaysiaPDPABreachIncident, coredata.MalaysiaPDPABreachOrderField],
	resolver any,
	parentID gid.GID,
) *MalaysiaPDPABreachIncidentConnection {
	edges := make([]*MalaysiaPDPABreachIncidentEdge, len(p.Data))
	for index, incident := range p.Data {
		edges[index] = NewMalaysiaPDPABreachIncidentEdge(incident, p.Cursor.OrderBy.Field)
	}

	return &MalaysiaPDPABreachIncidentConnection{
		Edges:    edges,
		PageInfo: *NewPageInfo(p),
		Resolver: resolver,
		ParentID: parentID,
	}
}

func NewMalaysiaPDPABreachStatusHistoryConnection(
	p *page.Page[*coredata.MalaysiaPDPABreachStatusHistory, coredata.MalaysiaPDPABreachStatusHistoryOrderField],
	incidentID gid.GID,
) *MalaysiaPDPABreachStatusHistoryConnection {
	edges := make([]*MalaysiaPDPABreachStatusHistoryEdge, len(p.Data))
	for index, history := range p.Data {
		edges[index] = NewMalaysiaPDPABreachStatusHistoryEdge(history, p.Cursor.OrderBy.Field)
	}

	return &MalaysiaPDPABreachStatusHistoryConnection{
		Edges:      edges,
		PageInfo:   *NewPageInfo(p),
		IncidentID: incidentID,
	}
}

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
		Title:                              incident.Title,
		Description:                        incident.Description,
		OccurredAt:                         incident.OccurredAt,
		DiscoveredAt:                       incident.DiscoveredAt,
		AwarenessAt:                        incident.AwarenessAt,
		AffectedDataSubjects:               incident.AffectedDataSubjects,
		AffectedDataRecords:                incident.AffectedDataRecords,
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
		CommissionerNotificationOverdue:    deadlineOverdue(assessment.CommissionerNotificationDueAt, incident.CommissionerNotifiedAt),
		CommissionerNotifiedAt:             incident.CommissionerNotifiedAt,
		CommissionerNotificationReference:  incident.CommissionerNotificationReference,
		CommissionerConfirmationReceivedAt: incident.CommissionerConfirmationReceivedAt,
		CommissionerConfirmationReference:  incident.CommissionerConfirmationReference,
		PhasedInformationDueAt:             assessment.PhasedInformationDueAt,
		DelayedNotificationReason:          incident.DelayedNotificationReason,
		DelayedNotificationEvidence:        incident.DelayedNotificationEvidence,
		DataSubjectsNotificationDueAt:      assessment.DataSubjectNotificationDueAt,
		DataSubjectsNotificationOverdue:    deadlineOverdue(assessment.DataSubjectNotificationDueAt, incident.DataSubjectsNotifiedAt),
		DataSubjectsNotifiedAt:             incident.DataSubjectsNotifiedAt,
		DataSubjectsNotificationEvidence:   incident.DataSubjectsNotificationEvidence,
		Status:                             incident.Status,
		CreatedByProfileID:                 incident.CreatedByProfileID,
		CreatedAt:                          incident.CreatedAt,
		UpdatedAt:                          incident.UpdatedAt,
	}
}

func NewMalaysiaPDPABreachIncidentEdge(
	incident *coredata.MalaysiaPDPABreachIncident,
	orderField coredata.MalaysiaPDPABreachOrderField,
) *MalaysiaPDPABreachIncidentEdge {
	return &MalaysiaPDPABreachIncidentEdge{
		Node:   NewMalaysiaPDPABreachIncident(incident),
		Cursor: incident.CursorKey(orderField),
	}
}

func NewMalaysiaPDPABreachStatusHistory(history *coredata.MalaysiaPDPABreachStatusHistory) *MalaysiaPDPABreachStatusHistory {
	return &MalaysiaPDPABreachStatusHistory{
		ID:                 history.ID,
		IncidentID:         history.IncidentID,
		FromStatus:         history.FromStatus,
		ToStatus:           history.ToStatus,
		ChangedByProfileID: history.ChangedByProfileID,
		Reason:             history.Reason,
		CreatedAt:          history.CreatedAt,
	}
}

func NewMalaysiaPDPABreachStatusHistoryEdge(
	history *coredata.MalaysiaPDPABreachStatusHistory,
	orderField coredata.MalaysiaPDPABreachStatusHistoryOrderField,
) *MalaysiaPDPABreachStatusHistoryEdge {
	return &MalaysiaPDPABreachStatusHistoryEdge{
		Node:   NewMalaysiaPDPABreachStatusHistory(history),
		Cursor: history.CursorKey(orderField),
	}
}

func deadlineOverdue(dueAt, completedAt *time.Time) bool {
	if dueAt == nil {
		return false
	}

	if completedAt != nil {
		return completedAt.After(*dueAt)
	}

	return time.Now().After(*dueAt)
}

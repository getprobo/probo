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

package probo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/malaysiapdpa"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/validator"
)

type MalaysiaPDPABreachService struct {
	svc *Service
}

type (
	CreateMalaysiaPDPABreachRequest struct {
		OrganizationID                     gid.GID
		Title                              string
		Description                        *string
		OccurredAt                         *time.Time
		DiscoveredAt                       time.Time
		AwarenessAt                        time.Time
		AffectedDataSubjects               int64
		AffectedDataRecords                int64
		PersonalDataTypes                  string
		AffectedSystem                     *string
		LikelyConsequences                 *string
		ContainmentActions                 *string
		PotentialPhysicalHarm              bool
		PotentialFinancialLoss             bool
		PotentialCreditOrPropertyDamage    bool
		PotentialIllegalUse                bool
		SensitivePersonalData              bool
		PotentialIdentityFraud             bool
		NotificationDecision               coredata.MalaysiaPDPABreachNotificationDecision
		DecisionRationale                  *string
		DecisionEvidence                   *string
		CommissionerNotifiedAt             *time.Time
		CommissionerNotificationReference  *string
		CommissionerConfirmationReceivedAt *time.Time
		CommissionerConfirmationReference  *string
		DelayedNotificationReason          *string
		DelayedNotificationEvidence        *string
		DataSubjectsNotifiedAt             *time.Time
		DataSubjectsNotificationEvidence   *string
		ActorProfileID                     gid.GID
	}

	UpdateMalaysiaPDPABreachRequest struct {
		ID                                 gid.GID
		Title                              *string
		Description                        **string
		OccurredAt                         **time.Time
		DiscoveredAt                       *time.Time
		AwarenessAt                        *time.Time
		AffectedDataSubjects               *int64
		AffectedDataRecords                *int64
		PersonalDataTypes                  *string
		AffectedSystem                     **string
		LikelyConsequences                 **string
		ContainmentActions                 **string
		PotentialPhysicalHarm              *bool
		PotentialFinancialLoss             *bool
		PotentialCreditOrPropertyDamage    *bool
		PotentialIllegalUse                *bool
		SensitivePersonalData              *bool
		PotentialIdentityFraud             *bool
		NotificationDecision               *coredata.MalaysiaPDPABreachNotificationDecision
		DecisionRationale                  **string
		DecisionEvidence                   **string
		CommissionerNotifiedAt             **time.Time
		CommissionerNotificationReference  **string
		CommissionerConfirmationReceivedAt **time.Time
		CommissionerConfirmationReference  **string
		DelayedNotificationReason          **string
		DelayedNotificationEvidence        **string
		DataSubjectsNotifiedAt             **time.Time
		DataSubjectsNotificationEvidence   **string
		ActorProfileID                     gid.GID
	}

	TransitionMalaysiaPDPABreachStatusRequest struct {
		ID             gid.GID
		ToStatus       coredata.MalaysiaPDPABreachStatus
		Reason         *string
		ActorProfileID gid.GID
	}
)

func (req *CreateMalaysiaPDPABreachRequest) Validate() error {
	v := validator.New()

	v.Check(req.OrganizationID, "organization_id", validator.Required(), validator.GID(coredata.OrganizationEntityType))
	v.Check(req.Title, "title", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(req.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(req.AffectedDataSubjects, "affected_data_subjects", validator.Min(0))
	v.Check(req.AffectedDataRecords, "affected_data_records", validator.Min(0))
	v.Check(req.PersonalDataTypes, "personal_data_types", validator.Required(), validator.SafeText(ContentMaxLength))
	v.Check(req.AffectedSystem, "affected_system", validator.SafeText(ContentMaxLength))
	v.Check(req.LikelyConsequences, "likely_consequences", validator.SafeText(ContentMaxLength))
	v.Check(req.ContainmentActions, "containment_actions", validator.SafeText(ContentMaxLength))
	v.Check(req.NotificationDecision, "notification_decision", validator.Required(), validator.OneOfSlice(coredata.MalaysiaPDPABreachNotificationDecisions()))
	v.Check(req.DecisionRationale, "decision_rationale", validator.SafeText(ContentMaxLength))
	v.Check(req.DecisionEvidence, "decision_evidence", validator.SafeText(ContentMaxLength))
	v.Check(req.CommissionerNotificationReference, "commissioner_notification_reference", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(req.CommissionerConfirmationReference, "commissioner_confirmation_reference", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(req.DelayedNotificationReason, "delayed_notification_reason", validator.SafeText(ContentMaxLength))
	v.Check(req.DelayedNotificationEvidence, "delayed_notification_evidence", validator.SafeText(ContentMaxLength))
	v.Check(req.DataSubjectsNotificationEvidence, "data_subjects_notification_evidence", validator.SafeText(ContentMaxLength))
	v.Check(req.ActorProfileID, "actor_profile_id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))

	validateMalaysiaPDPABreachTimeline(
		v,
		req.OccurredAt,
		req.DiscoveredAt,
		req.AwarenessAt,
		req.CommissionerNotifiedAt,
		req.CommissionerNotificationReference,
		req.CommissionerConfirmationReceivedAt,
		req.CommissionerConfirmationReference,
		req.DataSubjectsNotifiedAt,
		req.DataSubjectsNotificationEvidence,
	)

	if req.NotificationDecision != coredata.MalaysiaPDPABreachNotificationDecisionPending && isBlank(req.DecisionRationale) {
		addMalaysiaPDPABreachValidationError(v, req.DecisionRationale, "decision_rationale", "is required after a notification decision is recorded")
	}

	return v.Error()
}

func (req *UpdateMalaysiaPDPABreachRequest) Validate() error {
	v := validator.New()

	v.Check(req.ID, "id", validator.Required(), validator.GID(coredata.MalaysiaPDPABreachIncidentEntityType))
	v.Check(req.Title, "title", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(req.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(req.AffectedDataSubjects, "affected_data_subjects", validator.Min(0))
	v.Check(req.AffectedDataRecords, "affected_data_records", validator.Min(0))
	v.Check(req.PersonalDataTypes, "personal_data_types", validator.SafeText(ContentMaxLength))
	v.Check(req.AffectedSystem, "affected_system", validator.SafeText(ContentMaxLength))
	v.Check(req.LikelyConsequences, "likely_consequences", validator.SafeText(ContentMaxLength))
	v.Check(req.ContainmentActions, "containment_actions", validator.SafeText(ContentMaxLength))
	v.Check(req.NotificationDecision, "notification_decision", validator.OneOfSlice(coredata.MalaysiaPDPABreachNotificationDecisions()))
	v.Check(req.DecisionRationale, "decision_rationale", validator.SafeText(ContentMaxLength))
	v.Check(req.DecisionEvidence, "decision_evidence", validator.SafeText(ContentMaxLength))
	v.Check(req.CommissionerNotificationReference, "commissioner_notification_reference", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(req.CommissionerConfirmationReference, "commissioner_confirmation_reference", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(req.DelayedNotificationReason, "delayed_notification_reason", validator.SafeText(ContentMaxLength))
	v.Check(req.DelayedNotificationEvidence, "delayed_notification_evidence", validator.SafeText(ContentMaxLength))
	v.Check(req.DataSubjectsNotificationEvidence, "data_subjects_notification_evidence", validator.SafeText(ContentMaxLength))
	v.Check(req.ActorProfileID, "actor_profile_id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))

	return v.Error()
}

func (req *TransitionMalaysiaPDPABreachStatusRequest) Validate() error {
	v := validator.New()

	v.Check(req.ID, "id", validator.Required(), validator.GID(coredata.MalaysiaPDPABreachIncidentEntityType))
	v.Check(req.ToStatus, "to_status", validator.Required(), validator.OneOfSlice(coredata.MalaysiaPDPABreachStatuses()))
	v.Check(req.Reason, "reason", validator.SafeText(ContentMaxLength))
	v.Check(req.ActorProfileID, "actor_profile_id", validator.Required(), validator.GID(coredata.MembershipProfileEntityType))

	return v.Error()
}

func (s *MalaysiaPDPABreachService) Get(
	ctx context.Context,
	scope coredata.Scoper,
	id gid.GID,
) (*coredata.MalaysiaPDPABreachIncident, error) {
	incident := &coredata.MalaysiaPDPABreachIncident{}

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		if err := incident.LoadByID(ctx, conn, scope, id); err != nil {
			return fmt.Errorf("cannot load Malaysia PDPA breach incident: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return incident, nil
}

func (s *MalaysiaPDPABreachService) CountByOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		incidents := &coredata.MalaysiaPDPABreachIncidents{}
		var err error
		count, err = incidents.CountByOrganizationID(ctx, conn, scope, organizationID)
		if err != nil {
			return fmt.Errorf("cannot count Malaysia PDPA breach incidents: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *MalaysiaPDPABreachService) ListForOrganizationID(
	ctx context.Context,
	scope coredata.Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[coredata.MalaysiaPDPABreachOrderField],
) (*page.Page[*coredata.MalaysiaPDPABreachIncident, coredata.MalaysiaPDPABreachOrderField], error) {
	var incidents coredata.MalaysiaPDPABreachIncidents

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		organization := &coredata.Organization{}
		if err := organization.LoadByID(ctx, conn, scope, organizationID); err != nil {
			return fmt.Errorf("cannot load organization: %w", err)
		}

		if err := incidents.LoadByOrganizationID(ctx, conn, scope, organizationID, cursor); err != nil {
			return fmt.Errorf("cannot load Malaysia PDPA breach incidents: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return page.NewPage(incidents, cursor), nil
}

func (s *MalaysiaPDPABreachService) Create(
	ctx context.Context,
	scope coredata.Scoper,
	req *CreateMalaysiaPDPABreachRequest,
) (*coredata.MalaysiaPDPABreachIncident, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	assessment, err := assessMalaysiaPDPABreach(
		req.AwarenessAt,
		req.AffectedDataSubjects,
		req.PotentialPhysicalHarm,
		req.PotentialFinancialLoss,
		req.PotentialCreditOrPropertyDamage,
		req.PotentialIllegalUse,
		req.SensitivePersonalData,
		req.PotentialIdentityFraud,
		req.CommissionerNotifiedAt,
		req.NotificationDecision,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot assess Malaysia PDPA breach: %w", err)
	}

	now := time.Now()
	incident := &coredata.MalaysiaPDPABreachIncident{
		ID:                                 gid.New(scope.GetTenantID(), coredata.MalaysiaPDPABreachIncidentEntityType),
		OrganizationID:                     req.OrganizationID,
		Title:                              req.Title,
		Description:                        req.Description,
		OccurredAt:                         req.OccurredAt,
		DiscoveredAt:                       req.DiscoveredAt,
		AwarenessAt:                        req.AwarenessAt,
		AffectedDataSubjects:               req.AffectedDataSubjects,
		AffectedDataRecords:                req.AffectedDataRecords,
		PersonalDataTypes:                  req.PersonalDataTypes,
		AffectedSystem:                     req.AffectedSystem,
		LikelyConsequences:                 req.LikelyConsequences,
		ContainmentActions:                 req.ContainmentActions,
		PotentialPhysicalHarm:              req.PotentialPhysicalHarm,
		PotentialFinancialLoss:             req.PotentialFinancialLoss,
		PotentialCreditOrPropertyDamage:    req.PotentialCreditOrPropertyDamage,
		PotentialIllegalUse:                req.PotentialIllegalUse,
		SensitivePersonalData:              req.SensitivePersonalData,
		PotentialIdentityFraud:             req.PotentialIdentityFraud,
		NotificationDecision:               req.NotificationDecision,
		DecisionRationale:                  req.DecisionRationale,
		DecisionEvidence:                   req.DecisionEvidence,
		AssessedByProfileID:                req.ActorProfileID,
		AssessedAt:                         now,
		RuleVersion:                        malaysiapdpa.BreachNotificationRuleVersion,
		RuleSource:                         malaysiapdpa.BreachNotificationRuleSource,
		CommissionerNotifiedAt:             req.CommissionerNotifiedAt,
		CommissionerNotificationReference:  req.CommissionerNotificationReference,
		CommissionerConfirmationReceivedAt: req.CommissionerConfirmationReceivedAt,
		CommissionerConfirmationReference:  req.CommissionerConfirmationReference,
		DelayedNotificationReason:          req.DelayedNotificationReason,
		DelayedNotificationEvidence:        req.DelayedNotificationEvidence,
		DataSubjectsNotifiedAt:             req.DataSubjectsNotifiedAt,
		DataSubjectsNotificationEvidence:   req.DataSubjectsNotificationEvidence,
		Status:                             coredata.MalaysiaPDPABreachStatusOpen,
		CreatedByProfileID:                 req.ActorProfileID,
		CreatedAt:                          now,
		UpdatedAt:                          now,
	}
	applyMalaysiaPDPABreachAssessment(incident, assessment)

	if err := validateMalaysiaPDPABreachIncident(incident); err != nil {
		return nil, err
	}

	err = s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		organization := &coredata.Organization{}
		if err := organization.LoadByID(ctx, tx, scope, req.OrganizationID); err != nil {
			return fmt.Errorf("cannot load organization: %w", err)
		}

		if err := validateProfileOrganization(ctx, tx, scope, req.ActorProfileID, req.OrganizationID, "breach actor"); err != nil {
			return err
		}

		if err := incident.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert Malaysia PDPA breach incident: %w", err)
		}

		initialHistory := &coredata.MalaysiaPDPABreachStatusHistory{
			ID:                 gid.New(scope.GetTenantID(), coredata.MalaysiaPDPABreachStatusHistoryEntityType),
			OrganizationID:     req.OrganizationID,
			IncidentID:         incident.ID,
			ToStatus:           incident.Status,
			ChangedByProfileID: req.ActorProfileID,
			Reason:             stringPointer("Incident created"),
			CreatedAt:          now,
		}
		if err := initialHistory.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert initial Malaysia PDPA breach status: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return incident, nil
}

func (s *MalaysiaPDPABreachService) Update(
	ctx context.Context,
	scope coredata.Scoper,
	req *UpdateMalaysiaPDPABreachRequest,
) (*coredata.MalaysiaPDPABreachIncident, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	incident := &coredata.MalaysiaPDPABreachIncident{}
	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := incident.LoadByID(ctx, tx, scope, req.ID); err != nil {
			return fmt.Errorf("cannot load Malaysia PDPA breach incident: %w", err)
		}

		if err := validateProfileOrganization(ctx, tx, scope, req.ActorProfileID, incident.OrganizationID, "breach actor"); err != nil {
			return err
		}

		mergeMalaysiaPDPABreachUpdate(incident, req)
		assessment, err := assessMalaysiaPDPABreach(
			incident.AwarenessAt,
			incident.AffectedDataSubjects,
			incident.PotentialPhysicalHarm,
			incident.PotentialFinancialLoss,
			incident.PotentialCreditOrPropertyDamage,
			incident.PotentialIllegalUse,
			incident.SensitivePersonalData,
			incident.PotentialIdentityFraud,
			incident.CommissionerNotifiedAt,
			incident.NotificationDecision,
		)
		if err != nil {
			return fmt.Errorf("cannot assess Malaysia PDPA breach: %w", err)
		}

		applyMalaysiaPDPABreachAssessment(incident, assessment)
		incident.AssessedByProfileID = req.ActorProfileID
		incident.AssessedAt = time.Now()
		incident.RuleVersion = malaysiapdpa.BreachNotificationRuleVersion
		incident.RuleSource = malaysiapdpa.BreachNotificationRuleSource
		incident.UpdatedAt = incident.AssessedAt

		if err := validateMalaysiaPDPABreachIncident(incident); err != nil {
			return err
		}

		if err := incident.Update(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot update Malaysia PDPA breach incident: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return incident, nil
}

func (s *MalaysiaPDPABreachService) TransitionStatus(
	ctx context.Context,
	scope coredata.Scoper,
	req *TransitionMalaysiaPDPABreachStatusRequest,
) (*coredata.MalaysiaPDPABreachIncident, *coredata.MalaysiaPDPABreachStatusHistory, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	incident := &coredata.MalaysiaPDPABreachIncident{}
	var history *coredata.MalaysiaPDPABreachStatusHistory

	err := s.svc.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := incident.LoadByID(ctx, tx, scope, req.ID); err != nil {
			return fmt.Errorf("cannot load Malaysia PDPA breach incident: %w", err)
		}

		if err := validateProfileOrganization(ctx, tx, scope, req.ActorProfileID, incident.OrganizationID, "breach actor"); err != nil {
			return err
		}

		if !isMalaysiaPDPABreachStatusTransitionAllowed(incident.Status, req.ToStatus) {
			v := validator.New()
			addMalaysiaPDPABreachValidationError(
				v,
				req.ToStatus,
				"to_status",
				fmt.Sprintf("cannot transition from %s to %s", incident.Status, req.ToStatus),
			)

			return v.Error()
		}

		if req.ToStatus == coredata.MalaysiaPDPABreachStatusClosed {
			if err := validateMalaysiaPDPABreachClosure(incident); err != nil {
				return err
			}
		}

		fromStatus := incident.Status
		now := time.Now()
		incident.Status = req.ToStatus
		incident.UpdatedAt = now

		if err := incident.UpdateStatus(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot update Malaysia PDPA breach status: %w", err)
		}

		history = &coredata.MalaysiaPDPABreachStatusHistory{
			ID:                 gid.New(scope.GetTenantID(), coredata.MalaysiaPDPABreachStatusHistoryEntityType),
			OrganizationID:     incident.OrganizationID,
			IncidentID:         incident.ID,
			FromStatus:         &fromStatus,
			ToStatus:           req.ToStatus,
			ChangedByProfileID: req.ActorProfileID,
			Reason:             req.Reason,
			CreatedAt:          now,
		}
		if err := history.Insert(ctx, tx, scope); err != nil {
			return fmt.Errorf("cannot insert Malaysia PDPA breach status history: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	return incident, history, nil
}

func (s *MalaysiaPDPABreachService) CountStatusHistory(
	ctx context.Context,
	scope coredata.Scoper,
	incidentID gid.GID,
) (int, error) {
	var count int

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		incident := &coredata.MalaysiaPDPABreachIncident{}
		if err := incident.LoadByID(ctx, conn, scope, incidentID); err != nil {
			return fmt.Errorf("cannot load Malaysia PDPA breach incident: %w", err)
		}

		history := &coredata.MalaysiaPDPABreachStatusHistories{}
		var err error
		count, err = history.CountByIncidentID(ctx, conn, scope, incidentID)
		if err != nil {
			return fmt.Errorf("cannot count Malaysia PDPA breach status history: %w", err)
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *MalaysiaPDPABreachService) ListStatusHistory(
	ctx context.Context,
	scope coredata.Scoper,
	incidentID gid.GID,
	cursor *page.Cursor[coredata.MalaysiaPDPABreachStatusHistoryOrderField],
) (*page.Page[*coredata.MalaysiaPDPABreachStatusHistory, coredata.MalaysiaPDPABreachStatusHistoryOrderField], error) {
	var history coredata.MalaysiaPDPABreachStatusHistories

	err := s.svc.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		incident := &coredata.MalaysiaPDPABreachIncident{}
		if err := incident.LoadByID(ctx, conn, scope, incidentID); err != nil {
			return fmt.Errorf("cannot load Malaysia PDPA breach incident: %w", err)
		}

		if err := history.LoadByIncidentID(ctx, conn, scope, incidentID, cursor); err != nil {
			return fmt.Errorf("cannot load Malaysia PDPA breach status history: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return page.NewPage(history, cursor), nil
}

func mergeMalaysiaPDPABreachUpdate(incident *coredata.MalaysiaPDPABreachIncident, req *UpdateMalaysiaPDPABreachRequest) {
	if req.Title != nil {
		incident.Title = *req.Title
	}
	if req.Description != nil {
		incident.Description = *req.Description
	}
	if req.OccurredAt != nil {
		incident.OccurredAt = *req.OccurredAt
	}
	if req.DiscoveredAt != nil {
		incident.DiscoveredAt = *req.DiscoveredAt
	}
	if req.AwarenessAt != nil {
		incident.AwarenessAt = *req.AwarenessAt
	}
	if req.AffectedDataSubjects != nil {
		incident.AffectedDataSubjects = *req.AffectedDataSubjects
	}
	if req.AffectedDataRecords != nil {
		incident.AffectedDataRecords = *req.AffectedDataRecords
	}
	if req.PersonalDataTypes != nil {
		incident.PersonalDataTypes = *req.PersonalDataTypes
	}
	if req.AffectedSystem != nil {
		incident.AffectedSystem = *req.AffectedSystem
	}
	if req.LikelyConsequences != nil {
		incident.LikelyConsequences = *req.LikelyConsequences
	}
	if req.ContainmentActions != nil {
		incident.ContainmentActions = *req.ContainmentActions
	}
	if req.PotentialPhysicalHarm != nil {
		incident.PotentialPhysicalHarm = *req.PotentialPhysicalHarm
	}
	if req.PotentialFinancialLoss != nil {
		incident.PotentialFinancialLoss = *req.PotentialFinancialLoss
	}
	if req.PotentialCreditOrPropertyDamage != nil {
		incident.PotentialCreditOrPropertyDamage = *req.PotentialCreditOrPropertyDamage
	}
	if req.PotentialIllegalUse != nil {
		incident.PotentialIllegalUse = *req.PotentialIllegalUse
	}
	if req.SensitivePersonalData != nil {
		incident.SensitivePersonalData = *req.SensitivePersonalData
	}
	if req.PotentialIdentityFraud != nil {
		incident.PotentialIdentityFraud = *req.PotentialIdentityFraud
	}
	if req.NotificationDecision != nil {
		incident.NotificationDecision = *req.NotificationDecision
	}
	if req.DecisionRationale != nil {
		incident.DecisionRationale = *req.DecisionRationale
	}
	if req.DecisionEvidence != nil {
		incident.DecisionEvidence = *req.DecisionEvidence
	}
	if req.CommissionerNotifiedAt != nil {
		incident.CommissionerNotifiedAt = *req.CommissionerNotifiedAt
	}
	if req.CommissionerNotificationReference != nil {
		incident.CommissionerNotificationReference = *req.CommissionerNotificationReference
	}
	if req.CommissionerConfirmationReceivedAt != nil {
		incident.CommissionerConfirmationReceivedAt = *req.CommissionerConfirmationReceivedAt
	}
	if req.CommissionerConfirmationReference != nil {
		incident.CommissionerConfirmationReference = *req.CommissionerConfirmationReference
	}
	if req.DelayedNotificationReason != nil {
		incident.DelayedNotificationReason = *req.DelayedNotificationReason
	}
	if req.DelayedNotificationEvidence != nil {
		incident.DelayedNotificationEvidence = *req.DelayedNotificationEvidence
	}
	if req.DataSubjectsNotifiedAt != nil {
		incident.DataSubjectsNotifiedAt = *req.DataSubjectsNotifiedAt
	}
	if req.DataSubjectsNotificationEvidence != nil {
		incident.DataSubjectsNotificationEvidence = *req.DataSubjectsNotificationEvidence
	}
}

func assessMalaysiaPDPABreach(
	awarenessAt time.Time,
	affectedDataSubjects int64,
	potentialPhysicalHarm bool,
	potentialFinancialLoss bool,
	potentialCreditOrPropertyDamage bool,
	potentialIllegalUse bool,
	sensitivePersonalData bool,
	potentialIdentityFraud bool,
	commissionerNotifiedAt *time.Time,
	notificationDecision coredata.MalaysiaPDPABreachNotificationDecision,
) (malaysiapdpa.BreachAssessment, error) {
	return malaysiapdpa.AssessBreachNotification(malaysiapdpa.BreachAssessmentInput{
		AwarenessAt:                     awarenessAt,
		AffectedDataSubjects:            affectedDataSubjects,
		PotentialPhysicalHarm:           potentialPhysicalHarm,
		PotentialFinancialLoss:          potentialFinancialLoss,
		PotentialCreditOrPropertyDamage: potentialCreditOrPropertyDamage,
		PotentialIllegalUse:             potentialIllegalUse,
		SensitivePersonalData:           sensitivePersonalData,
		PotentialIdentityFraud:          potentialIdentityFraud,
		CommissionerNotifiedAt:          commissionerNotifiedAt,
		HumanCommissionerNotification: notificationDecision == coredata.MalaysiaPDPABreachNotificationDecisionCommissionerOnly ||
			notificationDecision == coredata.MalaysiaPDPABreachNotificationDecisionCommissionerAndDataSubjects,
		HumanDataSubjectNotification: notificationDecision == coredata.MalaysiaPDPABreachNotificationDecisionCommissionerAndDataSubjects,
	})
}

func applyMalaysiaPDPABreachAssessment(incident *coredata.MalaysiaPDPABreachIncident, assessment malaysiapdpa.BreachAssessment) {
	incident.SignificantHarm = assessment.SignificantHarm
	incident.SignificantScale = assessment.SignificantScale
	incident.NotificationRecommendation = coredata.MalaysiaPDPABreachNotificationDecision(assessment.Recommendation)
	incident.NotificationReasons = make([]string, len(assessment.Reasons))
	for index, reason := range assessment.Reasons {
		incident.NotificationReasons[index] = string(reason)
	}
}

func validateMalaysiaPDPABreachIncident(incident *coredata.MalaysiaPDPABreachIncident) error {
	v := validator.New()

	v.Check(incident.Title, "title", validator.Required(), validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(incident.Description, "description", validator.SafeText(ContentMaxLength))
	v.Check(incident.AffectedDataSubjects, "affected_data_subjects", validator.Min(0))
	v.Check(incident.AffectedDataRecords, "affected_data_records", validator.Min(0))
	v.Check(incident.PersonalDataTypes, "personal_data_types", validator.Required(), validator.SafeText(ContentMaxLength))
	v.Check(incident.AffectedSystem, "affected_system", validator.SafeText(ContentMaxLength))
	v.Check(incident.LikelyConsequences, "likely_consequences", validator.SafeText(ContentMaxLength))
	v.Check(incident.ContainmentActions, "containment_actions", validator.SafeText(ContentMaxLength))
	v.Check(incident.NotificationDecision, "notification_decision", validator.Required(), validator.OneOfSlice(coredata.MalaysiaPDPABreachNotificationDecisions()))
	v.Check(incident.DecisionRationale, "decision_rationale", validator.SafeText(ContentMaxLength))
	v.Check(incident.DecisionEvidence, "decision_evidence", validator.SafeText(ContentMaxLength))
	v.Check(incident.CommissionerNotificationReference, "commissioner_notification_reference", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(incident.CommissionerConfirmationReference, "commissioner_confirmation_reference", validator.SafeTextNoNewLine(TitleMaxLength))
	v.Check(incident.DelayedNotificationReason, "delayed_notification_reason", validator.SafeText(ContentMaxLength))
	v.Check(incident.DelayedNotificationEvidence, "delayed_notification_evidence", validator.SafeText(ContentMaxLength))
	v.Check(incident.DataSubjectsNotificationEvidence, "data_subjects_notification_evidence", validator.SafeText(ContentMaxLength))

	validateMalaysiaPDPABreachTimeline(
		v,
		incident.OccurredAt,
		incident.DiscoveredAt,
		incident.AwarenessAt,
		incident.CommissionerNotifiedAt,
		incident.CommissionerNotificationReference,
		incident.CommissionerConfirmationReceivedAt,
		incident.CommissionerConfirmationReference,
		incident.DataSubjectsNotifiedAt,
		incident.DataSubjectsNotificationEvidence,
	)

	if incident.NotificationDecision != coredata.MalaysiaPDPABreachNotificationDecisionPending && isBlank(incident.DecisionRationale) {
		addMalaysiaPDPABreachValidationError(v, incident.DecisionRationale, "decision_rationale", "is required after a notification decision is recorded")
	}

	commissionerRequired := incident.NotificationRecommendation != coredata.MalaysiaPDPABreachNotificationDecisionNotRequired ||
		incident.NotificationDecision == coredata.MalaysiaPDPABreachNotificationDecisionCommissionerOnly ||
		incident.NotificationDecision == coredata.MalaysiaPDPABreachNotificationDecisionCommissionerAndDataSubjects
	if commissionerRequired && incident.CommissionerNotifiedAt != nil && incident.CommissionerNotifiedAt.After(incident.AwarenessAt.Add(malaysiapdpa.CommissionerNotificationWindow)) {
		if isBlank(incident.DelayedNotificationReason) {
			addMalaysiaPDPABreachValidationError(v, incident.DelayedNotificationReason, "delayed_notification_reason", "is required for a notification submitted after 72 hours")
		}
		if isBlank(incident.DelayedNotificationEvidence) {
			addMalaysiaPDPABreachValidationError(v, incident.DelayedNotificationEvidence, "delayed_notification_evidence", "is required for a notification submitted after 72 hours")
		}
	}

	return v.Error()
}

func validateMalaysiaPDPABreachTimeline(
	v *validator.Validator,
	occurredAt *time.Time,
	discoveredAt time.Time,
	awarenessAt time.Time,
	commissionerNotifiedAt *time.Time,
	commissionerNotificationReference *string,
	commissionerConfirmationReceivedAt *time.Time,
	commissionerConfirmationReference *string,
	dataSubjectsNotifiedAt *time.Time,
	dataSubjectsNotificationEvidence *string,
) {
	if discoveredAt.IsZero() {
		addMalaysiaPDPABreachValidationError(v, discoveredAt, "discovered_at", "is required")
	}
	if awarenessAt.IsZero() {
		addMalaysiaPDPABreachValidationError(v, awarenessAt, "awareness_at", "is required")
	}
	if occurredAt != nil && !discoveredAt.IsZero() && occurredAt.After(discoveredAt) {
		addMalaysiaPDPABreachValidationError(v, occurredAt, "occurred_at", "must not be after discovered_at")
	}
	if !discoveredAt.IsZero() && !awarenessAt.IsZero() && discoveredAt.After(awarenessAt) {
		addMalaysiaPDPABreachValidationError(v, awarenessAt, "awareness_at", "must not be before discovered_at")
	}
	if commissionerNotifiedAt != nil && !awarenessAt.IsZero() && commissionerNotifiedAt.Before(awarenessAt) {
		addMalaysiaPDPABreachValidationError(v, commissionerNotifiedAt, "commissioner_notified_at", "must not be before awareness_at")
	}
	if commissionerNotifiedAt != nil && isBlank(commissionerNotificationReference) {
		addMalaysiaPDPABreachValidationError(v, commissionerNotificationReference, "commissioner_notification_reference", "is required when commissioner_notified_at is recorded")
	}
	if commissionerNotificationReference != nil && commissionerNotifiedAt == nil {
		addMalaysiaPDPABreachValidationError(v, commissionerNotificationReference, "commissioner_notification_reference", "requires commissioner_notified_at")
	}
	if commissionerConfirmationReceivedAt != nil {
		if commissionerNotifiedAt == nil {
			addMalaysiaPDPABreachValidationError(v, commissionerConfirmationReceivedAt, "commissioner_confirmation_received_at", "requires commissioner_notified_at")
		} else if commissionerConfirmationReceivedAt.Before(*commissionerNotifiedAt) {
			addMalaysiaPDPABreachValidationError(v, commissionerConfirmationReceivedAt, "commissioner_confirmation_received_at", "must not be before commissioner_notified_at")
		}
		if isBlank(commissionerConfirmationReference) {
			addMalaysiaPDPABreachValidationError(v, commissionerConfirmationReference, "commissioner_confirmation_reference", "is required when commissioner_confirmation_received_at is recorded")
		}
	}
	if commissionerConfirmationReference != nil && commissionerConfirmationReceivedAt == nil {
		addMalaysiaPDPABreachValidationError(v, commissionerConfirmationReference, "commissioner_confirmation_reference", "requires commissioner_confirmation_received_at")
	}
	if dataSubjectsNotifiedAt != nil {
		if commissionerNotifiedAt == nil {
			addMalaysiaPDPABreachValidationError(v, dataSubjectsNotifiedAt, "data_subjects_notified_at", "requires commissioner_notified_at")
		} else if dataSubjectsNotifiedAt.Before(*commissionerNotifiedAt) {
			addMalaysiaPDPABreachValidationError(v, dataSubjectsNotifiedAt, "data_subjects_notified_at", "must not be before commissioner_notified_at")
		}
		if isBlank(dataSubjectsNotificationEvidence) {
			addMalaysiaPDPABreachValidationError(v, dataSubjectsNotificationEvidence, "data_subjects_notification_evidence", "is required when data_subjects_notified_at is recorded")
		}
	}
	if dataSubjectsNotificationEvidence != nil && dataSubjectsNotifiedAt == nil {
		addMalaysiaPDPABreachValidationError(v, dataSubjectsNotificationEvidence, "data_subjects_notification_evidence", "requires data_subjects_notified_at")
	}
}

func isMalaysiaPDPABreachStatusTransitionAllowed(from, to coredata.MalaysiaPDPABreachStatus) bool {
	if from == to {
		return false
	}

	switch from {
	case coredata.MalaysiaPDPABreachStatusOpen:
		return to == coredata.MalaysiaPDPABreachStatusAssessing || to == coredata.MalaysiaPDPABreachStatusContained
	case coredata.MalaysiaPDPABreachStatusAssessing:
		return to == coredata.MalaysiaPDPABreachStatusOpen || to == coredata.MalaysiaPDPABreachStatusContained
	case coredata.MalaysiaPDPABreachStatusContained:
		return to == coredata.MalaysiaPDPABreachStatusAssessing || to == coredata.MalaysiaPDPABreachStatusClosed
	case coredata.MalaysiaPDPABreachStatusClosed:
		return to == coredata.MalaysiaPDPABreachStatusAssessing
	}

	return false
}

func validateMalaysiaPDPABreachClosure(incident *coredata.MalaysiaPDPABreachIncident) error {
	v := validator.New()

	commissionerRequired := incident.NotificationRecommendation != coredata.MalaysiaPDPABreachNotificationDecisionNotRequired ||
		incident.NotificationDecision == coredata.MalaysiaPDPABreachNotificationDecisionCommissionerOnly ||
		incident.NotificationDecision == coredata.MalaysiaPDPABreachNotificationDecisionCommissionerAndDataSubjects
	if commissionerRequired && incident.CommissionerNotifiedAt == nil {
		addMalaysiaPDPABreachValidationError(v, incident.CommissionerNotifiedAt, "commissioner_notified_at", "must be recorded before the incident can be closed")
	}

	dataSubjectsRequired := incident.NotificationRecommendation == coredata.MalaysiaPDPABreachNotificationDecisionCommissionerAndDataSubjects ||
		incident.NotificationDecision == coredata.MalaysiaPDPABreachNotificationDecisionCommissionerAndDataSubjects
	if dataSubjectsRequired && incident.DataSubjectsNotifiedAt == nil {
		addMalaysiaPDPABreachValidationError(v, incident.DataSubjectsNotifiedAt, "data_subjects_notified_at", "must be recorded before the incident can be closed")
	}

	return v.Error()
}

func addMalaysiaPDPABreachValidationError(v *validator.Validator, value any, field, message string) {
	v.Check(value, field, func(any) *validator.ValidationError {
		return &validator.ValidationError{Code: validator.ErrorCodeCustom, Message: message}
	})
}

func isBlank(value *string) bool {
	return value == nil || strings.TrimSpace(*value) == ""
}

func stringPointer(value string) *string { return &value }

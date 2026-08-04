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

package coredata

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	MalaysiaPDPABreachIncident struct {
		ID                                 gid.GID                                `db:"id"`
		OrganizationID                     gid.GID                                `db:"organization_id"`
		Title                              string                                 `db:"title"`
		Description                        *string                                `db:"description"`
		OccurredAt                         *time.Time                             `db:"occurred_at"`
		DiscoveredAt                       time.Time                              `db:"discovered_at"`
		AwarenessAt                        time.Time                              `db:"awareness_at"`
		AffectedDataSubjects               int64                                  `db:"affected_data_subjects"`
		AffectedDataRecords                int64                                  `db:"affected_data_records"`
		PersonalDataTypes                  string                                 `db:"personal_data_types"`
		AffectedSystem                     *string                                `db:"affected_system"`
		LikelyConsequences                 *string                                `db:"likely_consequences"`
		ContainmentActions                 *string                                `db:"containment_actions"`
		PotentialPhysicalHarm              bool                                   `db:"potential_physical_harm"`
		PotentialFinancialLoss             bool                                   `db:"potential_financial_loss"`
		PotentialCreditOrPropertyDamage    bool                                   `db:"potential_credit_or_property_damage"`
		PotentialIllegalUse                bool                                   `db:"potential_illegal_use"`
		SensitivePersonalData              bool                                   `db:"sensitive_personal_data"`
		PotentialIdentityFraud             bool                                   `db:"potential_identity_fraud"`
		SignificantHarm                    bool                                   `db:"significant_harm"`
		SignificantScale                   bool                                   `db:"significant_scale"`
		NotificationRecommendation         MalaysiaPDPABreachNotificationDecision `db:"notification_recommendation"`
		NotificationReasons                []string                               `db:"notification_reasons"`
		NotificationDecision               MalaysiaPDPABreachNotificationDecision `db:"notification_decision"`
		DecisionRationale                  *string                                `db:"decision_rationale"`
		DecisionEvidence                   *string                                `db:"decision_evidence"`
		AssessedByProfileID                gid.GID                                `db:"assessed_by_profile_id"`
		AssessedAt                         time.Time                              `db:"assessed_at"`
		RuleVersion                        string                                 `db:"rule_version"`
		RuleSource                         string                                 `db:"rule_source"`
		CommissionerNotifiedAt             *time.Time                             `db:"commissioner_notified_at"`
		CommissionerNotificationReference  *string                                `db:"commissioner_notification_reference"`
		CommissionerConfirmationReceivedAt *time.Time                             `db:"commissioner_confirmation_received_at"`
		CommissionerConfirmationReference  *string                                `db:"commissioner_confirmation_reference"`
		DelayedNotificationReason          *string                                `db:"delayed_notification_reason"`
		DelayedNotificationEvidence        *string                                `db:"delayed_notification_evidence"`
		DataSubjectsNotifiedAt             *time.Time                             `db:"data_subjects_notified_at"`
		DataSubjectsNotificationEvidence   *string                                `db:"data_subjects_notification_evidence"`
		Status                             MalaysiaPDPABreachStatus               `db:"status"`
		CreatedByProfileID                 gid.GID                                `db:"created_by_profile_id"`
		CreatedAt                          time.Time                              `db:"created_at"`
		UpdatedAt                          time.Time                              `db:"updated_at"`
	}

	MalaysiaPDPABreachIncidents []*MalaysiaPDPABreachIncident
)

const malaysiaPDPABreachIncidentColumns = `
    id,
    organization_id,
    title,
    description,
    occurred_at,
    discovered_at,
    awareness_at,
    affected_data_subjects,
    affected_data_records,
    personal_data_types,
    affected_system,
    likely_consequences,
    containment_actions,
    potential_physical_harm,
    potential_financial_loss,
    potential_credit_or_property_damage,
    potential_illegal_use,
    sensitive_personal_data,
    potential_identity_fraud,
    significant_harm,
    significant_scale,
    notification_recommendation,
    notification_reasons,
    notification_decision,
    decision_rationale,
    decision_evidence,
    assessed_by_profile_id,
    assessed_at,
    rule_version,
    rule_source,
    commissioner_notified_at,
    commissioner_notification_reference,
    commissioner_confirmation_received_at,
    commissioner_confirmation_reference,
    delayed_notification_reason,
    delayed_notification_evidence,
    data_subjects_notified_at,
    data_subjects_notification_evidence,
    status,
    created_by_profile_id,
    created_at,
    updated_at
`

func (i *MalaysiaPDPABreachIncident) CursorKey(field MalaysiaPDPABreachOrderField) page.CursorKey {
	switch field {
	case MalaysiaPDPABreachOrderFieldCreatedAt:
		return page.NewCursorKey(i.ID, i.CreatedAt)
	case MalaysiaPDPABreachOrderFieldUpdatedAt:
		return page.NewCursorKey(i.ID, i.UpdatedAt)
	case MalaysiaPDPABreachOrderFieldAwarenessAt:
		return page.NewCursorKey(i.ID, i.AwarenessAt)
	case MalaysiaPDPABreachOrderFieldTitle:
		return page.NewCursorKey(i.ID, i.Title)
	case MalaysiaPDPABreachOrderFieldStatus:
		return page.NewCursorKey(i.ID, i.Status)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (i *MalaysiaPDPABreachIncident) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM malaysia_pdpa_breach_incidents WHERE id = ANY(@resource_ids::text[])`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"resource_ids": resourceIDs})
	if err != nil {
		return nil, fmt.Errorf("cannot query Malaysia PDPA breach authorization attributes: %w", err)
	}
	defer rows.Close()

	attributes := make(policy.AttributesByID)
	for rows.Next() {
		var id, organizationID gid.GID
		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan Malaysia PDPA breach authorization attributes: %w", err)
		}

		attributes[id] = policy.Attributes{"organization_id": organizationID.String()}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate Malaysia PDPA breach authorization attributes: %w", err)
	}

	return attributes, nil
}

func (i *MalaysiaPDPABreachIncident) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `SELECT %s FROM malaysia_pdpa_breach_incidents WHERE %s AND id = @id LIMIT 1;`
	q = fmt.Sprintf(q, malaysiaPDPABreachIncidentColumns, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query Malaysia PDPA breach incident: %w", err)
	}

	incident, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[MalaysiaPDPABreachIncident])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect Malaysia PDPA breach incident: %w", err)
	}

	*i = incident

	return nil
}

func (is *MalaysiaPDPABreachIncidents) CountByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `SELECT COUNT(id) FROM malaysia_pdpa_breach_incidents WHERE %s AND organization_id = @organization_id;`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	var count int
	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count Malaysia PDPA breach incidents: %w", err)
	}

	return count, nil
}

func (is *MalaysiaPDPABreachIncidents) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[MalaysiaPDPABreachOrderField],
) error {
	q := `SELECT %s FROM malaysia_pdpa_breach_incidents WHERE %s AND organization_id = @organization_id AND %s;`
	q = fmt.Sprintf(q, malaysiaPDPABreachIncidentColumns, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query Malaysia PDPA breach incidents: %w", err)
	}

	incidents, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[MalaysiaPDPABreachIncident])
	if err != nil {
		return fmt.Errorf("cannot collect Malaysia PDPA breach incidents: %w", err)
	}

	*is = incidents

	return nil
}

func (i *MalaysiaPDPABreachIncident) Insert(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
INSERT INTO malaysia_pdpa_breach_incidents (
    id, tenant_id, organization_id, title, description, occurred_at,
    discovered_at, awareness_at, affected_data_subjects, affected_data_records,
    personal_data_types, affected_system, likely_consequences,
    containment_actions, potential_physical_harm, potential_financial_loss,
    potential_credit_or_property_damage, potential_illegal_use,
    sensitive_personal_data, potential_identity_fraud, significant_harm,
    significant_scale, notification_recommendation, notification_reasons,
    notification_decision, decision_rationale, decision_evidence,
    assessed_by_profile_id, assessed_at, rule_version, rule_source,
    commissioner_notified_at, commissioner_notification_reference,
    commissioner_confirmation_received_at, commissioner_confirmation_reference,
    delayed_notification_reason, delayed_notification_evidence,
    data_subjects_notified_at, data_subjects_notification_evidence, status,
    created_by_profile_id, created_at, updated_at
) VALUES (
    @id, @tenant_id, @organization_id, @title, @description, @occurred_at,
    @discovered_at, @awareness_at, @affected_data_subjects, @affected_data_records,
    @personal_data_types, @affected_system, @likely_consequences,
    @containment_actions, @potential_physical_harm, @potential_financial_loss,
    @potential_credit_or_property_damage, @potential_illegal_use,
    @sensitive_personal_data, @potential_identity_fraud, @significant_harm,
    @significant_scale, @notification_recommendation, @notification_reasons,
    @notification_decision, @decision_rationale, @decision_evidence,
    @assessed_by_profile_id, @assessed_at, @rule_version, @rule_source,
    @commissioner_notified_at, @commissioner_notification_reference,
    @commissioner_confirmation_received_at, @commissioner_confirmation_reference,
    @delayed_notification_reason, @delayed_notification_evidence,
    @data_subjects_notified_at, @data_subjects_notification_evidence, @status,
    @created_by_profile_id, @created_at, @updated_at
);`

	_, err := conn.Exec(ctx, q, i.insertSQLArguments(scope))
	if err != nil {
		return fmt.Errorf("cannot insert Malaysia PDPA breach incident: %w", err)
	}

	return nil
}

func (i *MalaysiaPDPABreachIncident) Update(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `
UPDATE malaysia_pdpa_breach_incidents SET
    title = @title,
    description = @description,
    occurred_at = @occurred_at,
    discovered_at = @discovered_at,
    awareness_at = @awareness_at,
    affected_data_subjects = @affected_data_subjects,
    affected_data_records = @affected_data_records,
    personal_data_types = @personal_data_types,
    affected_system = @affected_system,
    likely_consequences = @likely_consequences,
    containment_actions = @containment_actions,
    potential_physical_harm = @potential_physical_harm,
    potential_financial_loss = @potential_financial_loss,
    potential_credit_or_property_damage = @potential_credit_or_property_damage,
    potential_illegal_use = @potential_illegal_use,
    sensitive_personal_data = @sensitive_personal_data,
    potential_identity_fraud = @potential_identity_fraud,
    significant_harm = @significant_harm,
    significant_scale = @significant_scale,
    notification_recommendation = @notification_recommendation,
    notification_reasons = @notification_reasons,
    notification_decision = @notification_decision,
    decision_rationale = @decision_rationale,
    decision_evidence = @decision_evidence,
    assessed_by_profile_id = @assessed_by_profile_id,
    assessed_at = @assessed_at,
    rule_version = @rule_version,
    rule_source = @rule_source,
    commissioner_notified_at = @commissioner_notified_at,
    commissioner_notification_reference = @commissioner_notification_reference,
    commissioner_confirmation_received_at = @commissioner_confirmation_received_at,
    commissioner_confirmation_reference = @commissioner_confirmation_reference,
    delayed_notification_reason = @delayed_notification_reason,
    delayed_notification_evidence = @delayed_notification_evidence,
    data_subjects_notified_at = @data_subjects_notified_at,
    data_subjects_notification_evidence = @data_subjects_notification_evidence,
    updated_at = @updated_at
WHERE %s AND id = @id;`
	q = fmt.Sprintf(q, scope.SQLFragment())

	result, err := conn.Exec(ctx, q, i.updateSQLArguments(scope))
	if err != nil {
		return fmt.Errorf("cannot update Malaysia PDPA breach incident: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (i *MalaysiaPDPABreachIncident) UpdateStatus(ctx context.Context, conn pg.Tx, scope Scoper) error {
	q := `UPDATE malaysia_pdpa_breach_incidents SET status = @status, updated_at = @updated_at WHERE %s AND id = @id;`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":         i.ID,
		"status":     i.Status,
		"updated_at": i.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update Malaysia PDPA breach status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (i *MalaysiaPDPABreachIncident) insertSQLArguments(scope Scoper) pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"id":                                    i.ID,
		"tenant_id":                             scope.GetTenantID(),
		"organization_id":                       i.OrganizationID,
		"title":                                 i.Title,
		"description":                           i.Description,
		"occurred_at":                           i.OccurredAt,
		"discovered_at":                         i.DiscoveredAt,
		"awareness_at":                          i.AwarenessAt,
		"affected_data_subjects":                i.AffectedDataSubjects,
		"affected_data_records":                 i.AffectedDataRecords,
		"personal_data_types":                   i.PersonalDataTypes,
		"affected_system":                       i.AffectedSystem,
		"likely_consequences":                   i.LikelyConsequences,
		"containment_actions":                   i.ContainmentActions,
		"potential_physical_harm":               i.PotentialPhysicalHarm,
		"potential_financial_loss":              i.PotentialFinancialLoss,
		"potential_credit_or_property_damage":   i.PotentialCreditOrPropertyDamage,
		"potential_illegal_use":                 i.PotentialIllegalUse,
		"sensitive_personal_data":               i.SensitivePersonalData,
		"potential_identity_fraud":              i.PotentialIdentityFraud,
		"significant_harm":                      i.SignificantHarm,
		"significant_scale":                     i.SignificantScale,
		"notification_recommendation":           i.NotificationRecommendation,
		"notification_reasons":                  i.NotificationReasons,
		"notification_decision":                 i.NotificationDecision,
		"decision_rationale":                    i.DecisionRationale,
		"decision_evidence":                     i.DecisionEvidence,
		"assessed_by_profile_id":                i.AssessedByProfileID,
		"assessed_at":                           i.AssessedAt,
		"rule_version":                          i.RuleVersion,
		"rule_source":                           i.RuleSource,
		"commissioner_notified_at":              i.CommissionerNotifiedAt,
		"commissioner_notification_reference":   i.CommissionerNotificationReference,
		"commissioner_confirmation_received_at": i.CommissionerConfirmationReceivedAt,
		"commissioner_confirmation_reference":   i.CommissionerConfirmationReference,
		"delayed_notification_reason":           i.DelayedNotificationReason,
		"delayed_notification_evidence":         i.DelayedNotificationEvidence,
		"data_subjects_notified_at":             i.DataSubjectsNotifiedAt,
		"data_subjects_notification_evidence":   i.DataSubjectsNotificationEvidence,
		"status":                                i.Status,
		"created_by_profile_id":                 i.CreatedByProfileID,
		"created_at":                            i.CreatedAt,
		"updated_at":                            i.UpdatedAt,
	}

	return args
}

func (i *MalaysiaPDPABreachIncident) updateSQLArguments(scope Scoper) pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"id":                                    i.ID,
		"title":                                 i.Title,
		"description":                           i.Description,
		"occurred_at":                           i.OccurredAt,
		"discovered_at":                         i.DiscoveredAt,
		"awareness_at":                          i.AwarenessAt,
		"affected_data_subjects":                i.AffectedDataSubjects,
		"affected_data_records":                 i.AffectedDataRecords,
		"personal_data_types":                   i.PersonalDataTypes,
		"affected_system":                       i.AffectedSystem,
		"likely_consequences":                   i.LikelyConsequences,
		"containment_actions":                   i.ContainmentActions,
		"potential_physical_harm":               i.PotentialPhysicalHarm,
		"potential_financial_loss":              i.PotentialFinancialLoss,
		"potential_credit_or_property_damage":   i.PotentialCreditOrPropertyDamage,
		"potential_illegal_use":                 i.PotentialIllegalUse,
		"sensitive_personal_data":               i.SensitivePersonalData,
		"potential_identity_fraud":              i.PotentialIdentityFraud,
		"significant_harm":                      i.SignificantHarm,
		"significant_scale":                     i.SignificantScale,
		"notification_recommendation":           i.NotificationRecommendation,
		"notification_reasons":                  i.NotificationReasons,
		"notification_decision":                 i.NotificationDecision,
		"decision_rationale":                    i.DecisionRationale,
		"decision_evidence":                     i.DecisionEvidence,
		"assessed_by_profile_id":                i.AssessedByProfileID,
		"assessed_at":                           i.AssessedAt,
		"rule_version":                          i.RuleVersion,
		"rule_source":                           i.RuleSource,
		"commissioner_notified_at":              i.CommissionerNotifiedAt,
		"commissioner_notification_reference":   i.CommissionerNotificationReference,
		"commissioner_confirmation_received_at": i.CommissionerConfirmationReceivedAt,
		"commissioner_confirmation_reference":   i.CommissionerConfirmationReference,
		"delayed_notification_reason":           i.DelayedNotificationReason,
		"delayed_notification_evidence":         i.DelayedNotificationEvidence,
		"data_subjects_notified_at":             i.DataSubjectsNotifiedAt,
		"data_subjects_notification_evidence":   i.DataSubjectsNotificationEvidence,
		"updated_at":                            i.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	return args
}

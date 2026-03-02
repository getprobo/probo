// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/pkg/page"
)

type (
	ProcessingActivity struct {
		ID                                   gid.GID                                          `db:"id"`
		SnapshotID                           *gid.GID                                         `db:"snapshot_id"`
		SourceID                             *gid.GID                                         `db:"source_id"`
		OrganizationID                       gid.GID                                          `db:"organization_id"`
		Name                                 string                                           `db:"name"`
		Purpose                              *string                                          `db:"purpose"`
		DataSubjectCategory                  *string                                          `db:"data_subject_category"`
		PersonalDataCategory                 *string                                          `db:"personal_data_category"`
		SpecialOrCriminalData                ProcessingActivitySpecialOrCriminalDatum         `db:"special_or_criminal_data"`
		ConsentEvidenceLink                  *string                                          `db:"consent_evidence_link"`
		LawfulBasis                          ProcessingActivityLawfulBasis                    `db:"lawful_basis"`
		Recipients                           *string                                          `db:"recipients"`
		Location                             *string                                          `db:"location"`
		InternationalTransfers               bool                                             `db:"international_transfers"`
		TransferSafeguard                    *ProcessingActivityTransferSafeguard             `db:"transfer_safeguards"`
		RetentionPeriod                      *string                                          `db:"retention_period"`
		SecurityMeasures                     *string                                          `db:"security_measures"`
		DataProtectionImpactAssessmentNeeded ProcessingActivityDataProtectionImpactAssessment `db:"data_protection_impact_assessment_needed"`
		TransferImpactAssessmentNeeded       ProcessingActivityTransferImpactAssessment       `db:"transfer_impact_assessment_needed"`
		LastReviewDate                       *time.Time                                       `db:"last_review_date"`
		NextReviewDate                       *time.Time                                       `db:"next_review_date"`
		Role                                 ProcessingActivityRole                           `db:"role"`
		DataProtectionOfficerID              *gid.GID                                         `db:"dpo_profile_id"`
		CreatedAt                            time.Time                                        `db:"created_at"`
		UpdatedAt                            time.Time                                        `db:"updated_at"`
	}

	ProcessingActivities []*ProcessingActivity
)

func (p *ProcessingActivity) CursorKey(field ProcessingActivityOrderField) page.CursorKey {
	switch field {
	case ProcessingActivityOrderFieldCreatedAt:
		return page.NewCursorKey(p.ID, p.CreatedAt)
	case ProcessingActivityOrderFieldName:
		return page.NewCursorKey(p.ID, p.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (p *ProcessingActivity) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM processing_activities WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, p.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query processing activity authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (p *ProcessingActivity) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	processingActivityID gid.GID,
) error {
	q := `
SELECT
	id,
	snapshot_id,
	source_id,
	organization_id,
	name,
	purpose,
	data_subject_category,
	personal_data_category,
	special_or_criminal_data,
	consent_evidence_link,
	lawful_basis,
	recipients,
	location,
	international_transfers,
	transfer_safeguards,
	retention_period,
	security_measures,
	data_protection_impact_assessment_needed,
	transfer_impact_assessment_needed,
	last_review_date,
	next_review_date,
	role,
	dpo_profile_id,
	created_at,
	updated_at
FROM
	processing_activities
WHERE
	%s
	AND id = @processing_activity_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"processing_activity_id": processingActivityID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query processing activity: %w", err)
	}

	processingActivity, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ProcessingActivity])
	if err != nil {
		return fmt.Errorf("cannot collect processing activity: %w", err)
	}

	*p = processingActivity

	return nil
}

func (p *ProcessingActivities) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *ProcessingActivityFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	processing_activities
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count processing activities: %w", err)
	}

	return count, nil
}

func (p *ProcessingActivities) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ProcessingActivityOrderField],
	filter *ProcessingActivityFilter,
) error {
	q := `
SELECT
	id,
	snapshot_id,
	source_id,
	organization_id,
	name,
	purpose,
	data_subject_category,
	personal_data_category,
	special_or_criminal_data,
	consent_evidence_link,
	lawful_basis,
	recipients,
	location,
	international_transfers,
	transfer_safeguards,
	retention_period,
	security_measures,
	data_protection_impact_assessment_needed,
	transfer_impact_assessment_needed,
	last_review_date,
	next_review_date,
	role,
	dpo_profile_id,
	created_at,
	updated_at
FROM
	processing_activities
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query processing activities: %w", err)
	}

	processingActivities, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ProcessingActivity])
	if err != nil {
		return fmt.Errorf("cannot collect processing activities: %w", err)
	}

	*p = processingActivities

	return nil
}

func (p *ProcessingActivities) LoadAllByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *ProcessingActivityFilter,
) error {
	q := `
SELECT
	id,
	snapshot_id,
	source_id,
	organization_id,
	name,
	purpose,
	data_subject_category,
	personal_data_category,
	special_or_criminal_data,
	consent_evidence_link,
	lawful_basis,
	recipients,
	location,
	international_transfers,
	transfer_safeguards,
	retention_period,
	security_measures,
	data_protection_impact_assessment_needed,
	transfer_impact_assessment_needed,
	last_review_date,
	next_review_date,
	role,
	dpo_profile_id,
	created_at,
	updated_at
FROM
	processing_activities
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
ORDER BY created_at DESC
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query processing activities: %w", err)
	}

	processingActivities, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ProcessingActivity])
	if err != nil {
		return fmt.Errorf("cannot collect processing activities: %w", err)
	}

	*p = processingActivities

	return nil
}

func (p *ProcessingActivity) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO processing_activities (
	id,
	tenant_id,
	snapshot_id,
	source_id,
	organization_id,
	name,
	purpose,
	data_subject_category,
	personal_data_category,
	special_or_criminal_data,
	consent_evidence_link,
	lawful_basis,
	recipients,
	location,
	international_transfers,
	transfer_safeguards,
	retention_period,
	security_measures,
	data_protection_impact_assessment_needed,
	transfer_impact_assessment_needed,
	last_review_date,
	next_review_date,
	role,
	dpo_profile_id,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@snapshot_id,
	@source_id,
	@organization_id,
	@name,
	@purpose,
	@data_subject_category,
	@personal_data_category,
	@special_or_criminal_data,
	@consent_evidence_link,
	@lawful_basis,
	@recipients,
	@location,
	@international_transfers,
	@transfer_safeguards,
	@retention_period,
	@security_measures,
	@data_protection_impact_assessment_needed,
	@transfer_impact_assessment_needed,
	@last_review_date,
	@next_review_date,
	@role,
	@dpo_profile_id,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                       p.ID,
		"tenant_id":                scope.GetTenantID(),
		"snapshot_id":              p.SnapshotID,
		"source_id":                p.SourceID,
		"organization_id":          p.OrganizationID,
		"name":                     p.Name,
		"purpose":                  p.Purpose,
		"data_subject_category":    p.DataSubjectCategory,
		"personal_data_category":   p.PersonalDataCategory,
		"special_or_criminal_data": p.SpecialOrCriminalData,
		"consent_evidence_link":    p.ConsentEvidenceLink,
		"lawful_basis":             p.LawfulBasis,
		"recipients":               p.Recipients,
		"location":                 p.Location,
		"international_transfers":  p.InternationalTransfers,
		"transfer_safeguards":      p.TransferSafeguard,
		"retention_period":         p.RetentionPeriod,
		"security_measures":        p.SecurityMeasures,
		"data_protection_impact_assessment_needed": p.DataProtectionImpactAssessmentNeeded,
		"transfer_impact_assessment_needed":        p.TransferImpactAssessmentNeeded,
		"last_review_date":                         p.LastReviewDate,
		"next_review_date":                         p.NextReviewDate,
		"role":                                     p.Role,
		"dpo_profile_id":                           p.DataProtectionOfficerID,
		"created_at":                               p.CreatedAt,
		"updated_at":                               p.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert processing activity: %w", err)
	}

	return nil
}

func (p *ProcessingActivity) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE processing_activities
SET
	name = @name,
	purpose = @purpose,
	data_subject_category = @data_subject_category,
	personal_data_category = @personal_data_category,
	special_or_criminal_data = @special_or_criminal_data,
	consent_evidence_link = @consent_evidence_link,
	lawful_basis = @lawful_basis,
	recipients = @recipients,
	location = @location,
	international_transfers = @international_transfers,
	transfer_safeguards = @transfer_safeguards,
	retention_period = @retention_period,
	security_measures = @security_measures,
	data_protection_impact_assessment_needed = @data_protection_impact_assessment_needed,
	transfer_impact_assessment_needed = @transfer_impact_assessment_needed,
	last_review_date = @last_review_date,
	next_review_date = @next_review_date,
	role = @role,
	dpo_profile_id = @dpo_profile_id,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                       p.ID,
		"name":                     p.Name,
		"purpose":                  p.Purpose,
		"data_subject_category":    p.DataSubjectCategory,
		"personal_data_category":   p.PersonalDataCategory,
		"special_or_criminal_data": p.SpecialOrCriminalData,
		"consent_evidence_link":    p.ConsentEvidenceLink,
		"lawful_basis":             p.LawfulBasis,
		"recipients":               p.Recipients,
		"location":                 p.Location,
		"international_transfers":  p.InternationalTransfers,
		"transfer_safeguards":      p.TransferSafeguard,
		"retention_period":         p.RetentionPeriod,
		"security_measures":        p.SecurityMeasures,
		"data_protection_impact_assessment_needed": p.DataProtectionImpactAssessmentNeeded,
		"transfer_impact_assessment_needed":        p.TransferImpactAssessmentNeeded,
		"last_review_date":                         p.LastReviewDate,
		"next_review_date":                         p.NextReviewDate,
		"role":                                     p.Role,
		"dpo_profile_id":                           p.DataProtectionOfficerID,
		"updated_at":                               p.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update processing activity: %w", err)
	}

	return nil
}

func (p *ProcessingActivity) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM processing_activities
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": p.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete processing activity: %w", err)
	}

	return nil
}

func (pas ProcessingActivities) Snapshot(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error {
	snapshotters := []ProcessingActivitySnapshotter{ProcessingActivities{}, Vendors{}, ProcessingActivityVendors{}, DataProtectionImpactAssessments{}, TransferImpactAssessments{}}

	for _, snapshotter := range snapshotters {
		if err := snapshotter.InsertProcessingActivitySnapshots(ctx, conn, scope, organizationID, snapshotID); err != nil {
			return fmt.Errorf("cannot create processing activity snapshots: (%T) %w", snapshotter, err)
		}
	}

	return nil
}

func (pas ProcessingActivities) InsertProcessingActivitySnapshots(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
INSERT INTO processing_activities (
	id,
	tenant_id,
	snapshot_id,
	source_id,
	organization_id,
	name,
	purpose,
	data_subject_category,
	personal_data_category,
	special_or_criminal_data,
	consent_evidence_link,
	lawful_basis,
	recipients,
	location,
	international_transfers,
	transfer_safeguards,
	retention_period,
	security_measures,
	data_protection_impact_assessment_needed,
	transfer_impact_assessment_needed,
	last_review_date,
	next_review_date,
	role,
	dpo_profile_id,
	created_at,
	updated_at
)
SELECT
	generate_gid(decode_base64_unpadded(@tenant_id), @processing_activity_entity_type),
	@tenant_id,
	@snapshot_id,
	par.id,
	par.organization_id,
	par.name,
	par.purpose,
	par.data_subject_category,
	par.personal_data_category,
	par.special_or_criminal_data,
	par.consent_evidence_link,
	par.lawful_basis,
	par.recipients,
	par.location,
	par.international_transfers,
	par.transfer_safeguards,
	par.retention_period,
	par.security_measures,
	par.data_protection_impact_assessment_needed,
	par.transfer_impact_assessment_needed,
	par.last_review_date,
	par.next_review_date,
	par.role,
	par.dpo_profile_id,
	par.created_at,
	par.updated_at
FROM processing_activities par
WHERE %s AND par.organization_id = @organization_id AND par.snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":                       scope.GetTenantID(),
		"snapshot_id":                     snapshotID,
		"organization_id":                 organizationID,
		"processing_activity_entity_type": ProcessingActivityEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert processing activity snapshots: %w", err)
	}

	return nil
}

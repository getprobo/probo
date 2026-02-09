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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	TransferImpactAssessment struct {
		ID                    gid.GID   `db:"id"`
		SnapshotID            *gid.GID  `db:"snapshot_id"`
		SourceID              *gid.GID  `db:"source_id"`
		OrganizationID        gid.GID   `db:"organization_id"`
		ProcessingActivityID  gid.GID   `db:"processing_activity_id"`
		DataSubjects          *string   `db:"data_subjects"`
		LegalMechanism        *string   `db:"legal_mechanism"`
		Transfer              *string   `db:"transfer"`
		LocalLawRisk          *string   `db:"local_law_risk"`
		SupplementaryMeasures *string   `db:"supplementary_measures"`
		CreatedAt             time.Time `db:"created_at"`
		UpdatedAt             time.Time `db:"updated_at"`
	}

	TransferImpactAssessments []*TransferImpactAssessment
)

func (tia *TransferImpactAssessment) CursorKey(field TransferImpactAssessmentOrderField) page.CursorKey {
	switch field {
	case TransferImpactAssessmentOrderFieldCreatedAt:
		return page.NewCursorKey(tia.ID, tia.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (tia *TransferImpactAssessment) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM processing_activity_transfer_impact_assessments WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, tia.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}

		return nil, fmt.Errorf("cannot query transfer impact assessment authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (tias *TransferImpactAssessments) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *TransferImpactAssessmentFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	processing_activity_transfer_impact_assessments
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
		return 0, fmt.Errorf("cannot count transfer impact assessments: %w", err)
	}

	return count, nil
}

func (tias *TransferImpactAssessments) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[TransferImpactAssessmentOrderField],
	filter *TransferImpactAssessmentFilter,
) error {
	q := `
SELECT
	id,
	snapshot_id,
	source_id,
	organization_id,
	processing_activity_id,
	data_subjects,
	legal_mechanism,
	transfer,
	local_law_risk,
	supplementary_measures,
	created_at,
	updated_at
FROM
	processing_activity_transfer_impact_assessments
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
		return fmt.Errorf("cannot query transfer impact assessments: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TransferImpactAssessment])
	if err != nil {
		return fmt.Errorf("cannot collect transfer impact assessments: %w", err)
	}

	*tias = results

	return nil
}

func (tias *TransferImpactAssessments) LoadAllByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *TransferImpactAssessmentFilter,
) error {
	q := `
SELECT
	id,
	snapshot_id,
	source_id,
	organization_id,
	processing_activity_id,
	data_subjects,
	legal_mechanism,
	transfer,
	local_law_risk,
	supplementary_measures,
	created_at,
	updated_at
FROM
	processing_activity_transfer_impact_assessments
WHERE
	%s
	AND organization_id = @organization_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query transfer impact assessments: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[TransferImpactAssessment])
	if err != nil {
		return fmt.Errorf("cannot collect transfer impact assessments: %w", err)
	}

	*tias = results

	return nil
}

func (tia *TransferImpactAssessment) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	tiaID gid.GID,
) error {
	q := `
SELECT
	id,
	snapshot_id,
	source_id,
	organization_id,
	processing_activity_id,
	data_subjects,
	legal_mechanism,
	transfer,
	local_law_risk,
	supplementary_measures,
	created_at,
	updated_at
FROM
	processing_activity_transfer_impact_assessments
WHERE
	%s
	AND id = @tia_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"tia_id": tiaID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query transfer impact assessment: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TransferImpactAssessment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect transfer impact assessment: %w", err)
	}

	*tia = result

	return nil
}

func (tia *TransferImpactAssessment) LoadByProcessingActivityID(
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
	processing_activity_id,
	data_subjects,
	legal_mechanism,
	transfer,
	local_law_risk,
	supplementary_measures,
	created_at,
	updated_at
FROM
	processing_activity_transfer_impact_assessments
WHERE
	%s
	AND processing_activity_id = @processing_activity_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"processing_activity_id": processingActivityID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query transfer impact assessment: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[TransferImpactAssessment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect transfer impact assessment: %w", err)
	}

	*tia = result

	return nil
}

func (tia *TransferImpactAssessment) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO processing_activity_transfer_impact_assessments (
	id,
	tenant_id,
	organization_id,
	processing_activity_id,
	data_subjects,
	legal_mechanism,
	transfer,
	local_law_risk,
	supplementary_measures,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@processing_activity_id,
	@data_subjects,
	@legal_mechanism,
	@transfer,
	@local_law_risk,
	@supplementary_measures,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                     tia.ID,
		"tenant_id":              scope.GetTenantID(),
		"organization_id":        tia.OrganizationID,
		"processing_activity_id": tia.ProcessingActivityID,
		"data_subjects":          tia.DataSubjects,
		"legal_mechanism":        tia.LegalMechanism,
		"transfer":               tia.Transfer,
		"local_law_risk":         tia.LocalLawRisk,
		"supplementary_measures": tia.SupplementaryMeasures,
		"created_at":             tia.CreatedAt,
		"updated_at":             tia.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "processing_activity_tias_processing_activity_id_snapshot_id_uniq" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert transfer impact assessment: %w", err)
	}

	return nil
}

func (tia *TransferImpactAssessment) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE processing_activity_transfer_impact_assessments SET
	data_subjects = @data_subjects,
	legal_mechanism = @legal_mechanism,
	transfer = @transfer,
	local_law_risk = @local_law_risk,
	supplementary_measures = @supplementary_measures,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                     tia.ID,
		"data_subjects":          tia.DataSubjects,
		"legal_mechanism":        tia.LegalMechanism,
		"transfer":               tia.Transfer,
		"local_law_risk":         tia.LocalLawRisk,
		"supplementary_measures": tia.SupplementaryMeasures,
		"updated_at":             tia.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update transfer impact assessment: %w", err)
	}

	return nil
}

func (tia *TransferImpactAssessment) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM processing_activity_transfer_impact_assessments
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": tia.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete transfer impact assessment: %w", err)
	}

	return nil
}

func (tias TransferImpactAssessments) InsertProcessingActivitySnapshots(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
INSERT INTO processing_activity_transfer_impact_assessments (
	id,
	tenant_id,
	snapshot_id,
	source_id,
	organization_id,
	processing_activity_id,
	data_subjects,
	legal_mechanism,
	transfer,
	local_law_risk,
	supplementary_measures,
	created_at,
	updated_at
)
SELECT
	generate_gid(decode_base64_unpadded(@tenant_id), @tia_entity_type),
	@tenant_id,
	@snapshot_id,
	tia.id,
	tia.organization_id,
	pa_snapshot.id,
	tia.data_subjects,
	tia.legal_mechanism,
	tia.transfer,
	tia.local_law_risk,
	tia.supplementary_measures,
	tia.created_at,
	tia.updated_at
FROM processing_activity_transfer_impact_assessments tia
INNER JOIN processing_activities pa_source ON tia.processing_activity_id = pa_source.id AND pa_source.snapshot_id IS NULL
INNER JOIN processing_activities pa_snapshot ON pa_source.id = pa_snapshot.source_id AND pa_snapshot.snapshot_id = @snapshot_id
WHERE tia.tenant_id = @tenant_id AND tia.organization_id = @organization_id AND tia.snapshot_id IS NULL
	`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"snapshot_id":     snapshotID,
		"organization_id": organizationID,
		"tia_entity_type": TransferImpactAssessmentEntityType,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert transfer impact assessment snapshots: %w", err)
	}

	return nil
}

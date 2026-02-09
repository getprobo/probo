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
	DataProtectionImpactAssessment struct {
		ID                          gid.GID                                     `db:"id"`
		SnapshotID                  *gid.GID                                    `db:"snapshot_id"`
		SourceID                    *gid.GID                                    `db:"source_id"`
		OrganizationID              gid.GID                                     `db:"organization_id"`
		ProcessingActivityID        gid.GID                                     `db:"processing_activity_id"`
		Description                 *string                                     `db:"description"`
		NecessityAndProportionality *string                                     `db:"necessity_and_proportionality"`
		PotentialRisk               *string                                     `db:"potential_risk"`
		Mitigations                 *string                                     `db:"mitigations"`
		ResidualRisk                *DataProtectionImpactAssessmentResidualRisk `db:"residual_risk"`
		CreatedAt                   time.Time                                   `db:"created_at"`
		UpdatedAt                   time.Time                                   `db:"updated_at"`
	}

	DataProtectionImpactAssessments []*DataProtectionImpactAssessment
)

func (dpia *DataProtectionImpactAssessment) CursorKey(field DataProtectionImpactAssessmentOrderField) page.CursorKey {
	switch field {
	case DataProtectionImpactAssessmentOrderFieldCreatedAt:
		return page.NewCursorKey(dpia.ID, dpia.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (dpia *DataProtectionImpactAssessment) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM processing_activity_data_protection_impact_assessments WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, dpia.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query data protection impact assessment authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (dpias *DataProtectionImpactAssessments) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *DataProtectionImpactAssessmentFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	processing_activity_data_protection_impact_assessments
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
		return 0, fmt.Errorf("cannot count data protection impact assessments: %w", err)
	}

	return count, nil
}

func (dpias *DataProtectionImpactAssessments) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[DataProtectionImpactAssessmentOrderField],
	filter *DataProtectionImpactAssessmentFilter,
) error {
	q := `
SELECT
	id,
	snapshot_id,
	source_id,
	organization_id,
	processing_activity_id,
	description,
	necessity_and_proportionality,
	potential_risk,
	mitigations,
	residual_risk,
	created_at,
	updated_at
FROM
	processing_activity_data_protection_impact_assessments
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
		return fmt.Errorf("cannot query data protection impact assessments: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DataProtectionImpactAssessment])
	if err != nil {
		return fmt.Errorf("cannot collect data protection impact assessments: %w", err)
	}

	*dpias = results

	return nil
}

func (dpias *DataProtectionImpactAssessments) LoadAllByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *DataProtectionImpactAssessmentFilter,
) error {
	q := `
SELECT
	id,
	snapshot_id,
	source_id,
	organization_id,
	processing_activity_id,
	description,
	necessity_and_proportionality,
	potential_risk,
	mitigations,
	residual_risk,
	created_at,
	updated_at
FROM
	processing_activity_data_protection_impact_assessments
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
		return fmt.Errorf("cannot query data protection impact assessments: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DataProtectionImpactAssessment])
	if err != nil {
		return fmt.Errorf("cannot collect data protection impact assessments: %w", err)
	}

	*dpias = results

	return nil
}

func (dpia *DataProtectionImpactAssessment) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	dpiaID gid.GID,
) error {
	q := `
SELECT
	id,
	snapshot_id,
	source_id,
	organization_id,
	processing_activity_id,
	description,
	necessity_and_proportionality,
	potential_risk,
	mitigations,
	residual_risk,
	created_at,
	updated_at
FROM
	processing_activity_data_protection_impact_assessments
WHERE
	%s
	AND id = @dpia_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"dpia_id": dpiaID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query data protection impact assessment: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DataProtectionImpactAssessment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect data protection impact assessment: %w", err)
	}

	*dpia = result

	return nil
}

func (dpia *DataProtectionImpactAssessment) LoadByProcessingActivityID(
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
	description,
	necessity_and_proportionality,
	potential_risk,
	mitigations,
	residual_risk,
	created_at,
	updated_at
FROM
	processing_activity_data_protection_impact_assessments
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
		return fmt.Errorf("cannot query data protection impact assessment: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DataProtectionImpactAssessment])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect data protection impact assessment: %w", err)
	}

	*dpia = result

	return nil
}

func (dpia *DataProtectionImpactAssessment) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO processing_activity_data_protection_impact_assessments (
	id,
	tenant_id,
	organization_id,
	processing_activity_id,
	description,
	necessity_and_proportionality,
	potential_risk,
	mitigations,
	residual_risk,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@processing_activity_id,
	@description,
	@necessity_and_proportionality,
	@potential_risk,
	@mitigations,
	@residual_risk,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                            dpia.ID,
		"tenant_id":                     scope.GetTenantID(),
		"organization_id":               dpia.OrganizationID,
		"processing_activity_id":        dpia.ProcessingActivityID,
		"description":                   dpia.Description,
		"necessity_and_proportionality": dpia.NecessityAndProportionality,
		"potential_risk":                dpia.PotentialRisk,
		"mitigations":                   dpia.Mitigations,
		"residual_risk":                 dpia.ResidualRisk,
		"created_at":                    dpia.CreatedAt,
		"updated_at":                    dpia.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "processing_activity_dpias_processing_activity_id_snapshot_id_uniq" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert data protection impact assessment: %w", err)
	}

	return nil
}

func (dpia *DataProtectionImpactAssessment) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE processing_activity_data_protection_impact_assessments SET
	description = @description,
	necessity_and_proportionality = @necessity_and_proportionality,
	potential_risk = @potential_risk,
	mitigations = @mitigations,
	residual_risk = @residual_risk,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                            dpia.ID,
		"description":                   dpia.Description,
		"necessity_and_proportionality": dpia.NecessityAndProportionality,
		"potential_risk":                dpia.PotentialRisk,
		"mitigations":                   dpia.Mitigations,
		"residual_risk":                 dpia.ResidualRisk,
		"updated_at":                    dpia.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update data protection impact assessment: %w", err)
	}

	return nil
}

func (dpia *DataProtectionImpactAssessment) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM processing_activity_data_protection_impact_assessments
WHERE
	%s
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": dpia.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete data protection impact assessment: %w", err)
	}

	return nil
}

func (dpias DataProtectionImpactAssessments) InsertProcessingActivitySnapshots(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
INSERT INTO processing_activity_data_protection_impact_assessments (
	id,
	tenant_id,
	snapshot_id,
	source_id,
	organization_id,
	processing_activity_id,
	description,
	necessity_and_proportionality,
	potential_risk,
	mitigations,
	residual_risk,
	created_at,
	updated_at
)
SELECT
	generate_gid(decode_base64_unpadded(@tenant_id), @dpia_entity_type),
	@tenant_id,
	@snapshot_id,
	dpia.id,
	dpia.organization_id,
	pa_snapshot.id,
	dpia.description,
	dpia.necessity_and_proportionality,
	dpia.potential_risk,
	dpia.mitigations,
	dpia.residual_risk,
	dpia.created_at,
	dpia.updated_at
FROM processing_activity_data_protection_impact_assessments dpia
INNER JOIN processing_activities pa_source ON dpia.processing_activity_id = pa_source.id AND pa_source.snapshot_id IS NULL
INNER JOIN processing_activities pa_snapshot ON pa_source.id = pa_snapshot.source_id AND pa_snapshot.snapshot_id = @snapshot_id
WHERE dpia.tenant_id = @tenant_id AND dpia.organization_id = @organization_id AND dpia.snapshot_id IS NULL
	`

	args := pgx.StrictNamedArgs{
		"tenant_id":        scope.GetTenantID(),
		"snapshot_id":      snapshotID,
		"organization_id":  organizationID,
		"dpia_entity_type": DataProtectionImpactAssessmentEntityType,
	}

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert data protection impact assessment snapshots: %w", err)
	}

	return nil
}

// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

func (d DataProtectionImpactAssessment) GetGeneratedDocumentID(
	ctx context.Context,
	conn pg.Querier,
	organizationID gid.GID,
) (*gid.GID, error) {
	var documentID *gid.GID

	err := conn.QueryRow(
		ctx,
		`
SELECT
	data_protection_impact_assessments_document_id
FROM
	generated_documents
WHERE
	organization_id = @organization_id
`,
		pgx.NamedArgs{"organization_id": organizationID},
	).Scan(&documentID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cannot get DPIA list document ID: %w", err)
	}

	return documentID, nil
}

func (d DataProtectionImpactAssessment) UpsertGeneratedDocumentID(
	ctx context.Context,
	conn pg.Tx,
	organizationID gid.GID,
	tenantID gid.TenantID,
	documentID gid.GID,
) error {
	now := time.Now()

	_, err := conn.Exec(
		ctx,
		`
INSERT INTO generated_documents (
	organization_id,
	tenant_id,
	data_protection_impact_assessments_document_id,
	created_at,
	updated_at
) VALUES (
	@organization_id,
	@tenant_id,
	@data_protection_impact_assessments_document_id,
	@created_at,
	@updated_at
)
ON CONFLICT (organization_id) DO UPDATE
SET
	data_protection_impact_assessments_document_id = @data_protection_impact_assessments_document_id,
	updated_at = @updated_at
`,
		pgx.NamedArgs{
			"organization_id": organizationID,
			"tenant_id":       tenantID,
			"data_protection_impact_assessments_document_id": documentID,
			"created_at": now,
			"updated_at": now,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot upsert DPIA list document ID: %w", err)
	}

	return nil
}

func (d DataProtectionImpactAssessment) ClearGeneratedDocumentID(
	ctx context.Context,
	conn pg.Tx,
	documentIDs []gid.GID,
) error {
	ids := make([]string, len(documentIDs))
	for i, id := range documentIDs {
		ids[i] = id.String()
	}

	_, err := conn.Exec(
		ctx,
		`
UPDATE
	generated_documents
SET
	data_protection_impact_assessments_document_id = NULL,
	updated_at = @now
WHERE
	data_protection_impact_assessments_document_id = ANY(@ids)
`,
		pgx.NamedArgs{
			"ids": ids,
			"now": time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("cannot clear DPIA list document references: %w", err)
	}

	return nil
}

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
func (dpia *DataProtectionImpactAssessment) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
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
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	processing_activity_data_protection_impact_assessments
WHERE
	%s
	AND organization_id = @organization_id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

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
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[DataProtectionImpactAssessmentOrderField],
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
	AND snapshot_id IS NULL
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
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
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
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
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

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
	conn pg.Querier,
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
	conn pg.Querier,
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
	conn pg.Tx,
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
	conn pg.Tx,
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
	conn pg.Tx,
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

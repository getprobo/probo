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

type ErrProcessingActivityTIANotFound struct {
	Identifier string
}

func (e ErrProcessingActivityTIANotFound) Error() string {
	return fmt.Sprintf("processing activity tia not found: %q", e.Identifier)
}

type (
	ProcessingActivityTIA struct {
		ID                    gid.GID   `db:"id"`
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

	ProcessingActivityTIAs []*ProcessingActivityTIA
)

func (tia *ProcessingActivityTIA) CursorKey(field ProcessingActivityTIAOrderField) page.CursorKey {
	switch field {
	case ProcessingActivityTIAOrderFieldCreatedAt:
		return page.NewCursorKey(tia.ID, tia.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (tias *ProcessingActivityTIAs) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	processing_activity_transfer_impact_assessments
WHERE
	%s
	AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count processing activity tias: %w", err)
	}

	return count, nil
}

func (tias *ProcessingActivityTIAs) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ProcessingActivityTIAOrderField],
) error {
	q := `
SELECT
	id,
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

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query processing activity tias: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ProcessingActivityTIA])
	if err != nil {
		return fmt.Errorf("cannot collect processing activity tias: %w", err)
	}

	*tias = results

	return nil
}

func (tia *ProcessingActivityTIA) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	tiaID gid.GID,
) error {
	q := `
SELECT
	id,
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
		return fmt.Errorf("cannot query processing activity tia: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ProcessingActivityTIA])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrProcessingActivityTIANotFound{Identifier: tiaID.String()}
		}
		return fmt.Errorf("cannot collect processing activity tia: %w", err)
	}

	*tia = result

	return nil
}

func (tia *ProcessingActivityTIA) LoadByProcessingActivityID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	processingActivityID gid.GID,
) error {
	q := `
SELECT
	id,
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
		return fmt.Errorf("cannot query processing activity tia: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ProcessingActivityTIA])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrProcessingActivityTIANotFound{Identifier: processingActivityID.String()}
		}
		return fmt.Errorf("cannot collect processing activity tia: %w", err)
	}

	*tia = result

	return nil
}

func (tia *ProcessingActivityTIA) Insert(
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
		return fmt.Errorf("cannot insert processing activity tia: %w", err)
	}

	return nil
}

func (tia *ProcessingActivityTIA) Update(
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
		return fmt.Errorf("cannot update processing activity tia: %w", err)
	}

	return nil
}

func (tia *ProcessingActivityTIA) Delete(
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
		return fmt.Errorf("cannot delete processing activity tia: %w", err)
	}

	return nil
}

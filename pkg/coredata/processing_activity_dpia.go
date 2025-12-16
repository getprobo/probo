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

type ErrProcessingActivityDPIANotFound struct {
	Identifier string
}

func (e ErrProcessingActivityDPIANotFound) Error() string {
	return fmt.Sprintf("processing activity dpia not found: %q", e.Identifier)
}

type (
	ProcessingActivityDPIA struct {
		ID                          gid.GID                             `db:"id"`
		OrganizationID              gid.GID                             `db:"organization_id"`
		ProcessingActivityID        gid.GID                             `db:"processing_activity_id"`
		Description                 *string                             `db:"description"`
		NecessityAndProportionality *string                             `db:"necessity_and_proportionality"`
		PotentialRisk               *string                             `db:"potential_risk"`
		Mitigations                 *string                             `db:"mitigations"`
		ResidualRisk                *ProcessingActivityDPIAResidualRisk `db:"residual_risk"`
		CreatedAt                   time.Time                           `db:"created_at"`
		UpdatedAt                   time.Time                           `db:"updated_at"`
	}

	ProcessingActivityDPIAs []*ProcessingActivityDPIA
)

func (dpia *ProcessingActivityDPIA) CursorKey(field ProcessingActivityDPIAOrderField) page.CursorKey {
	switch field {
	case ProcessingActivityDPIAOrderFieldCreatedAt:
		return page.NewCursorKey(dpia.ID, dpia.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (dpias *ProcessingActivityDPIAs) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
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
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count processing activity dpias: %w", err)
	}

	return count, nil
}

func (dpias *ProcessingActivityDPIAs) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[ProcessingActivityDPIAOrderField],
) error {
	q := `
SELECT
	id,
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

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query processing activity dpias: %w", err)
	}

	results, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[ProcessingActivityDPIA])
	if err != nil {
		return fmt.Errorf("cannot collect processing activity dpias: %w", err)
	}

	*dpias = results

	return nil
}

func (dpia *ProcessingActivityDPIA) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	dpiaID gid.GID,
) error {
	q := `
SELECT
	id,
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
		return fmt.Errorf("cannot query processing activity dpia: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ProcessingActivityDPIA])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrProcessingActivityDPIANotFound{Identifier: dpiaID.String()}
		}
		return fmt.Errorf("cannot collect processing activity dpia: %w", err)
	}

	*dpia = result

	return nil
}

func (dpia *ProcessingActivityDPIA) LoadByProcessingActivityID(
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
		return fmt.Errorf("cannot query processing activity dpia: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[ProcessingActivityDPIA])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrProcessingActivityDPIANotFound{Identifier: processingActivityID.String()}
		}
		return fmt.Errorf("cannot collect processing activity dpia: %w", err)
	}

	*dpia = result

	return nil
}

func (dpia *ProcessingActivityDPIA) Insert(
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
		return fmt.Errorf("cannot insert processing activity dpia: %w", err)
	}

	return nil
}

func (dpia *ProcessingActivityDPIA) Update(
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
		return fmt.Errorf("cannot update processing activity dpia: %w", err)
	}

	return nil
}

func (dpia *ProcessingActivityDPIA) Delete(
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
		return fmt.Errorf("cannot delete processing activity dpia: %w", err)
	}

	return nil
}

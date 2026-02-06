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
	Nonconformity struct {
		ID                 gid.GID             `db:"id"`
		OrganizationID     gid.GID             `db:"organization_id"`
		SnapshotID         *gid.GID            `db:"snapshot_id"`
		SourceID           *gid.GID            `db:"source_id"`
		ReferenceID        string              `db:"reference_id"`
		Description        *string             `db:"description"`
		AuditID            *gid.GID            `db:"audit_id"`
		DateIdentified     *time.Time          `db:"date_identified"`
		RootCause          string              `db:"root_cause"`
		CorrectiveAction   *string             `db:"corrective_action"`
		OwnerID            gid.GID             `db:"owner_profile_id"`
		DueDate            *time.Time          `db:"due_date"`
		Status             NonconformityStatus `db:"status"`
		EffectivenessCheck *string             `db:"effectiveness_check"`
		CreatedAt          time.Time           `db:"created_at"`
		UpdatedAt          time.Time           `db:"updated_at"`
	}

	Nonconformities []*Nonconformity
)

func (nc *Nonconformity) CursorKey(field NonconformityOrderField) page.CursorKey {
	switch field {
	case NonconformityOrderFieldCreatedAt:
		return page.NewCursorKey(nc.ID, nc.CreatedAt)
	case NonconformityOrderFieldDateIdentified:
		return page.NewCursorKey(nc.ID, nc.DateIdentified)
	case NonconformityOrderFieldDueDate:
		return page.NewCursorKey(nc.ID, nc.DueDate)
	case NonconformityOrderFieldStatus:
		return page.NewCursorKey(nc.ID, nc.Status)
	case NonconformityOrderFieldReferenceId:
		return page.NewCursorKey(nc.ID, nc.ReferenceID)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (nc *Nonconformity) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM nonconformities WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, nc.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query nonconformity authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (nc *Nonconformity) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	nonconformityID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	snapshot_id,
	source_id,
	reference_id,
	description,
	audit_id,
	date_identified,
	root_cause,
	corrective_action,
	owner_profile_id,
	due_date,
	status,
	effectiveness_check,
	created_at,
	updated_at
FROM
	nonconformities
WHERE
	%s
	AND id = @nonconformity_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"nonconformity_id": nonconformityID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query nonconformity: %w", err)
	}

	nonconformity, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Nonconformity])
	if err != nil {
		return fmt.Errorf("cannot collect nonconformity: %w", err)
	}

	*nc = nonconformity

	return nil
}

func (ncs *Nonconformities) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *NonconformityFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	nonconformities
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
		return 0, fmt.Errorf("cannot count nonconformities: %w", err)
	}

	return count, nil
}

func (ncs *Nonconformities) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[NonconformityOrderField],
	filter *NonconformityFilter,
) error {
	q := `
SELECT
	id,
	organization_id,
	snapshot_id,
	source_id,
	reference_id,
	description,
	audit_id,
	date_identified,
	root_cause,
	corrective_action,
	owner_profile_id,
	due_date,
	status,
	effectiveness_check,
	created_at,
	updated_at
FROM
	nonconformities
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
		return fmt.Errorf("cannot query nonconformities: %w", err)
	}

	nonconformities, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Nonconformity])
	if err != nil {
		return fmt.Errorf("cannot collect nonconformities: %w", err)
	}

	*ncs = nonconformities

	return nil
}

func (nc *Nonconformity) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO nonconformities (
	id,
	tenant_id,
	organization_id,
	reference_id,
	description,
	audit_id,
	date_identified,
	root_cause,
	corrective_action,
	owner_profile_id,
	due_date,
	status,
	effectiveness_check,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@reference_id,
	@description,
	@audit_id,
	@date_identified,
	@root_cause,
	@corrective_action,
	@owner_profile_id,
	@due_date,
	@status,
	@effectiveness_check,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                  nc.ID,
		"tenant_id":           scope.GetTenantID(),
		"organization_id":     nc.OrganizationID,
		"reference_id":        nc.ReferenceID,
		"description":         nc.Description,
		"audit_id":            nc.AuditID,
		"date_identified":     nc.DateIdentified,
		"root_cause":          nc.RootCause,
		"corrective_action":   nc.CorrectiveAction,
		"owner_profile_id":    nc.OwnerID,
		"due_date":            nc.DueDate,
		"status":              nc.Status,
		"effectiveness_check": nc.EffectivenessCheck,
		"created_at":          nc.CreatedAt,
		"updated_at":          nc.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert nonconformity: %w", err)
	}

	return nil
}

func (nc *Nonconformity) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE nonconformities
SET
	reference_id = @reference_id,
	description = @description,
	date_identified = @date_identified,
	root_cause = @root_cause,
	corrective_action = @corrective_action,
	due_date = @due_date,
	status = @status,
	effectiveness_check = @effectiveness_check,
	owner_profile_id = @owner_profile_id,
	audit_id = @audit_id,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                  nc.ID,
		"reference_id":        nc.ReferenceID,
		"description":         nc.Description,
		"date_identified":     nc.DateIdentified,
		"root_cause":          nc.RootCause,
		"corrective_action":   nc.CorrectiveAction,
		"due_date":            nc.DueDate,
		"status":              nc.Status,
		"effectiveness_check": nc.EffectivenessCheck,
		"owner_profile_id":    nc.OwnerID,
		"audit_id":            nc.AuditID,
		"updated_at":          nc.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update nonconformity: %w", err)
	}

	return nil
}

func (nc *Nonconformity) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM nonconformities
WHERE
	%s
	AND id = @id AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": nc.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete nonconformity: %w", err)
	}

	return nil
}

func (ncs Nonconformities) Snapshot(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error {
	query := `
INSERT INTO nonconformities (
	id,
	tenant_id,
	snapshot_id,
	source_id,
	organization_id,
	reference_id,
	description,
	audit_id,
	date_identified,
	root_cause,
	corrective_action,
	owner_profile_id,
	due_date,
	status,
	effectiveness_check,
	created_at,
	updated_at
)
SELECT
	generate_gid(decode_base64_unpadded(@tenant_id), @nonconformity_entity_type),
	@tenant_id,
	@snapshot_id,
	nc.id,
	nc.organization_id,
	nc.reference_id,
	nc.description,
	nc.audit_id,
	nc.date_identified,
	nc.root_cause,
	nc.corrective_action,
	nc.owner_profile_id,
	nc.due_date,
	nc.status,
	nc.effectiveness_check,
	nc.created_at,
	nc.updated_at
FROM nonconformities nc
WHERE %s AND nc.organization_id = @organization_id AND nc.snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":                 scope.GetTenantID(),
		"snapshot_id":               snapshotID,
		"organization_id":           organizationID,
		"nonconformity_entity_type": NonconformityEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert data snapshots: %w", err)
	}

	return nil
}

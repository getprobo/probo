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
	"fmt"
	"maps"
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	NonconformityRegistry struct {
		ID                 gid.GID                     `db:"id"`
		OrganizationID     gid.GID                     `db:"organization_id"`
		SnapshotID         *gid.GID                    `db:"snapshot_id"`
		SourceID           *gid.GID                    `db:"source_id"`
		ReferenceID        string                      `db:"reference_id"`
		Description        *string                     `db:"description"`
		AuditID            gid.GID                     `db:"audit_id"`
		DateIdentified     *time.Time                  `db:"date_identified"`
		RootCause          string                      `db:"root_cause"`
		CorrectiveAction   *string                     `db:"corrective_action"`
		OwnerID            gid.GID                     `db:"owner_id"`
		DueDate            *time.Time                  `db:"due_date"`
		Status             NonconformityRegistryStatus `db:"status"`
		EffectivenessCheck *string                     `db:"effectiveness_check"`
		CreatedAt          time.Time                   `db:"created_at"`
		UpdatedAt          time.Time                   `db:"updated_at"`
	}

	NonconformityRegistries []*NonconformityRegistry
)

func (nr *NonconformityRegistry) CursorKey(field NonconformityRegistryOrderField) page.CursorKey {
	switch field {
	case NonconformityRegistryOrderFieldCreatedAt:
		return page.NewCursorKey(nr.ID, nr.CreatedAt)
	case NonconformityRegistryOrderFieldDateIdentified:
		return page.NewCursorKey(nr.ID, nr.DateIdentified)
	case NonconformityRegistryOrderFieldDueDate:
		return page.NewCursorKey(nr.ID, nr.DueDate)
	case NonconformityRegistryOrderFieldStatus:
		return page.NewCursorKey(nr.ID, nr.Status)
	case NonconformityRegistryOrderFieldReferenceId:
		return page.NewCursorKey(nr.ID, nr.ReferenceID)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (nr *NonconformityRegistry) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	nonconformityRegistryID gid.GID,
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
	owner_id,
	due_date,
	status,
	effectiveness_check,
	created_at,
	updated_at
FROM
	nonconformity_registries
WHERE
	%s
	AND id = @nonconformity_registry_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"nonconformity_registry_id": nonconformityRegistryID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query nonconformity registry: %w", err)
	}

	registry, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[NonconformityRegistry])
	if err != nil {
		return fmt.Errorf("cannot collect nonconformity registry: %w", err)
	}

	*nr = registry

	return nil
}

func (nrs *NonconformityRegistries) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *NonconformityRegistryFilter,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	nonconformity_registries
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
		return 0, fmt.Errorf("cannot count nonconformity registries: %w", err)
	}

	return count, nil
}

func (nrs *NonconformityRegistries) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[NonconformityRegistryOrderField],
	filter *NonconformityRegistryFilter,
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
	owner_id,
	due_date,
	status,
	effectiveness_check,
	created_at,
	updated_at
FROM
	nonconformity_registries
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
		return fmt.Errorf("cannot query nonconformity registries: %w", err)
	}

	registries, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[NonconformityRegistry])
	if err != nil {
		return fmt.Errorf("cannot collect nonconformity registries: %w", err)
	}

	*nrs = registries

	return nil
}

func (nr *NonconformityRegistry) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO nonconformity_registries (
	id,
	tenant_id,
	organization_id,
	reference_id,
	description,
	audit_id,
	date_identified,
	root_cause,
	corrective_action,
	owner_id,
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
	@owner_id,
	@due_date,
	@status,
	@effectiveness_check,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                  nr.ID,
		"tenant_id":           scope.GetTenantID(),
		"organization_id":     nr.OrganizationID,
		"reference_id":        nr.ReferenceID,
		"description":         nr.Description,
		"audit_id":            nr.AuditID,
		"date_identified":     nr.DateIdentified,
		"root_cause":          nr.RootCause,
		"corrective_action":   nr.CorrectiveAction,
		"owner_id":            nr.OwnerID,
		"due_date":            nr.DueDate,
		"status":              nr.Status,
		"effectiveness_check": nr.EffectivenessCheck,
		"created_at":          nr.CreatedAt,
		"updated_at":          nr.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert nonconformity registry: %w", err)
	}

	return nil
}

func (nr *NonconformityRegistry) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE nonconformity_registries
SET
	reference_id = @reference_id,
	description = @description,
	date_identified = @date_identified,
	root_cause = @root_cause,
	corrective_action = @corrective_action,
	due_date = @due_date,
	status = @status,
	effectiveness_check = @effectiveness_check,
	owner_id = @owner_id,
	audit_id = @audit_id,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
	AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                  nr.ID,
		"reference_id":        nr.ReferenceID,
		"description":         nr.Description,
		"date_identified":     nr.DateIdentified,
		"root_cause":          nr.RootCause,
		"corrective_action":   nr.CorrectiveAction,
		"due_date":            nr.DueDate,
		"status":              nr.Status,
		"effectiveness_check": nr.EffectivenessCheck,
		"owner_id":            nr.OwnerID,
		"audit_id":            nr.AuditID,
		"updated_at":          nr.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update nonconformity registry: %w", err)
	}

	return nil
}

func (nr *NonconformityRegistry) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM nonconformity_registries
WHERE
	%s
	AND id = @id AND snapshot_id IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": nr.ID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete nonconformity registry: %w", err)
	}

	return nil
}

func (nrs NonconformityRegistries) Snapshot(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error {
	query := `
INSERT INTO nonconformity_registries (
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
	owner_id,
	due_date,
	status,
	effectiveness_check,
	created_at,
	updated_at
)
SELECT
	generate_gid(decode_base64_unpadded(@tenant_id), @nonconformity_registry_entity_type),
	@tenant_id,
	@snapshot_id,
	r.id,
	r.organization_id,
	r.reference_id,
	r.description,
	r.audit_id,
	r.date_identified,
	r.root_cause,
	r.corrective_action,
	r.owner_id,
	r.due_date,
	r.status,
	r.effectiveness_check,
	r.created_at,
	r.updated_at
FROM nonconformity_registries r
WHERE %s AND r.organization_id = @organization_id AND r.snapshot_id IS NULL
	`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":                          scope.GetTenantID(),
		"snapshot_id":                        snapshotID,
		"organization_id":                    organizationID,
		"nonconformity_registry_entity_type": NonconformityRegistryEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert data snapshots: %w", err)
	}

	return nil
}

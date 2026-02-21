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
	StateOfApplicability struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		Name           string    `db:"name"`
		SourceID       *gid.GID  `db:"source_id"`
		SnapshotID     *gid.GID  `db:"snapshot_id"`
		OwnerID        gid.GID   `db:"owner_profile_id"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	StatesOfApplicability []*StateOfApplicability
)

func (s StateOfApplicability) CursorKey(orderBy StateOfApplicabilityOrderField) page.CursorKey {
	switch orderBy {
	case StateOfApplicabilityOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	case StateOfApplicabilityOrderFieldName:
		return page.NewCursorKey(s.ID, s.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *StateOfApplicability) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM states_of_applicability WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, s.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query state of applicability authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (s *StateOfApplicability) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	stateOfApplicabilityID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    source_id,
    snapshot_id,
    owner_profile_id,
    created_at,
    updated_at
FROM
    states_of_applicability
WHERE
    %s
    AND id = @state_of_applicability_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"state_of_applicability_id": stateOfApplicabilityID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query states_of_applicability: %w", err)
	}

	stateOfApplicability, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[StateOfApplicability])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect state_of_applicability: %w", err)
	}

	*s = stateOfApplicability
	return nil
}

func (s *StatesOfApplicability) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[StateOfApplicabilityOrderField],
	filter *StateOfApplicabilityFilter,
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    source_id,
    snapshot_id,
    owner_profile_id,
    created_at,
    updated_at
FROM
    states_of_applicability
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query states_of_applicability: %w", err)
	}

	statesOfApplicability, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[StateOfApplicability])
	if err != nil {
		return fmt.Errorf("cannot collect states_of_applicability: %w", err)
	}

	*s = statesOfApplicability
	return nil
}

func (s *StatesOfApplicability) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	filter *StateOfApplicabilityFilter,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    states_of_applicability
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), filter.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count states_of_applicability: %w", err)
	}

	return count, nil
}

func (s *StateOfApplicability) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    states_of_applicability (
        tenant_id,
        id,
        organization_id,
        name,
        source_id,
        snapshot_id,
        owner_profile_id,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @state_of_applicability_id,
    @organization_id,
    @name,
    @source_id,
    @snapshot_id,
    @owner_profile_id,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":                 scope.GetTenantID(),
		"state_of_applicability_id": s.ID,
		"organization_id":           s.OrganizationID,
		"name":                      s.Name,
		"source_id":                 s.SourceID,
		"snapshot_id":               s.SnapshotID,
		"owner_profile_id":          s.OwnerID,
		"created_at":                s.CreatedAt,
		"updated_at":                s.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot insert state_of_applicability: %w", err)
	}

	return nil
}

func (s *StateOfApplicability) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE states_of_applicability
SET
    name = @name,
    owner_profile_id = @owner_profile_id,
    updated_at = @updated_at
WHERE
    %s
    AND id = @state_of_applicability_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"state_of_applicability_id": s.ID,
		"name":                      s.Name,
		"owner_profile_id":          s.OwnerID,
		"updated_at":                s.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return ErrResourceAlreadyExists
			}
		}
		return fmt.Errorf("cannot update state_of_applicability: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *StateOfApplicability) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM states_of_applicability
WHERE
    %s
    AND id = @state_of_applicability_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"state_of_applicability_id": s.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete state_of_applicability: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (soas StatesOfApplicability) Snapshot(ctx context.Context, conn pg.Conn, scope Scoper, organizationID, snapshotID gid.GID) error {
	if err := soas.insertStateOfApplicabilitySnapshots(ctx, conn, scope, organizationID, snapshotID); err != nil {
		return fmt.Errorf("cannot insert state_of_applicability snapshots: %w", err)
	}

	if err := soas.insertStateOfApplicabilityControlSnapshots(ctx, conn, scope, organizationID, snapshotID); err != nil {
		return fmt.Errorf("cannot insert state_of_applicability_control snapshots: %w", err)
	}

	return nil
}

func (soas StatesOfApplicability) insertStateOfApplicabilitySnapshots(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
INSERT INTO states_of_applicability (
    id,
    tenant_id,
    organization_id,
    name,
    source_id,
    snapshot_id,
    owner_profile_id,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(@tenant_id), @state_of_applicability_entity_type),
    @tenant_id,
    soa.organization_id,
    soa.name,
    soa.id,
    @snapshot_id,
    soa.owner_profile_id,
    soa.created_at,
    soa.updated_at
FROM states_of_applicability soa
WHERE
    %s
    AND soa.organization_id = @organization_id
    AND soa.snapshot_id IS NULL
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":                          scope.GetTenantID(),
		"snapshot_id":                        snapshotID,
		"organization_id":                    organizationID,
		"state_of_applicability_entity_type": StateOfApplicabilityEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert state_of_applicability snapshots: %w", err)
	}

	return nil
}

func (soas StatesOfApplicability) insertStateOfApplicabilityControlSnapshots(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	snapshotID gid.GID,
) error {
	query := `
WITH source_soa AS (
    SELECT id, organization_id
    FROM states_of_applicability
    WHERE
        %s
        AND organization_id = @organization_id
        AND snapshot_id IS NULL
),
snapshot_soa AS (
    SELECT id, source_id
    FROM states_of_applicability
    WHERE snapshot_id = @snapshot_id
)
INSERT INTO applicability_statements (
    id,
    state_of_applicability_id,
    control_id,
    organization_id,
    tenant_id,
    snapshot_id,
    applicability,
    justification,
    created_at,
    updated_at
)
SELECT
    generate_gid(decode_base64_unpadded(@tenant_id), @applicability_statement_entity_type),
    snapshot_soa.id,
    soac.control_id,
    soac.organization_id,
    @tenant_id,
    @snapshot_id,
    soac.applicability,
    soac.justification,
    soac.created_at,
    soac.updated_at
FROM applicability_statements soac
INNER JOIN source_soa
    ON soac.state_of_applicability_id = source_soa.id
INNER JOIN snapshot_soa
    ON snapshot_soa.source_id = source_soa.id
WHERE soac.snapshot_id IS NULL
`

	query = fmt.Sprintf(query, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"tenant_id":                           scope.GetTenantID(),
		"snapshot_id":                         snapshotID,
		"organization_id":                     organizationID,
		"applicability_statement_entity_type": ApplicabilityStatementEntityType,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("cannot insert state_of_applicability_control snapshots: %w", err)
	}

	return nil
}

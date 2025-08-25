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
	Snapshot struct {
		ID             gid.GID       `db:"id"`
		OrganizationID gid.GID       `db:"organization_id"`
		Name           string        `db:"name"`
		Description    *string       `db:"description"`
		Type           SnapshotsType `db:"type"`
		CreatedAt      time.Time     `db:"created_at"`
	}

	Snapshots []*Snapshot
)

func (s *Snapshot) CursorKey(field SnapshotOrderField) page.CursorKey {
	switch field {
	case SnapshotOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	case SnapshotOrderFieldName:
		return page.NewCursorKey(s.ID, s.Name)
	case SnapshotOrderFieldType:
		return page.NewCursorKey(s.ID, s.Type)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (s *Snapshot) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	snapshotID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	description,
	type,
	created_at
FROM
	snapshots
WHERE
	%s
	AND id = @snapshot_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"snapshot_id": snapshotID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query snapshots: %w", err)
	}

	snapshot, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Snapshot])
	if err != nil {
		return fmt.Errorf("cannot collect snapshot: %w", err)
	}

	*s = snapshot

	return nil
}

func (s *Snapshots) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
	COUNT(id)
FROM
	snapshots
WHERE
	%s
	AND organization_id = @organization_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (s *Snapshots) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[SnapshotOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	name,
	description,
	type,
	created_at
FROM
	snapshots
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
		return fmt.Errorf("cannot query snapshots: %w", err)
	}

	snapshots, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Snapshot])
	if err != nil {
		return fmt.Errorf("cannot collect snapshots: %w", err)
	}

	*s = snapshots

	return nil
}

func (s *Snapshot) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO snapshots (
	id,
	tenant_id,
	organization_id,
	name,
	description,
	type,
	created_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@name,
	@description,
	@type,
	@created_at
)
`

	args := pgx.StrictNamedArgs{
		"id":              s.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": s.OrganizationID,
		"name":            s.Name,
		"description":     s.Description,
		"type":            s.Type,
		"created_at":      s.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert snapshot: %w", err)
	}

	return nil
}

func (s *Snapshot) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM snapshots
WHERE
	%s
	AND organization_id = @organization_id
	AND id = @id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": s.ID, "organization_id": s.OrganizationID}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete snapshot: %w", err)
	}

	return nil
}

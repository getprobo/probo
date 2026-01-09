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
	Meeting struct {
		ID             gid.GID   `db:"id"`
		OrganizationID gid.GID   `db:"organization_id"`
		Name           string    `db:"name"`
		Date           time.Time `db:"date"`
		Minutes        *string   `db:"minutes"`
		CreatedAt      time.Time `db:"created_at"`
		UpdatedAt      time.Time `db:"updated_at"`
	}

	Meetings []*Meeting
)

func (m Meeting) CursorKey(orderBy MeetingOrderField) page.CursorKey {
	switch orderBy {
	case MeetingOrderFieldCreatedAt:
		return page.NewCursorKey(m.ID, m.CreatedAt)
	case MeetingOrderFieldDate:
		return page.NewCursorKey(m.ID, m.Date)
	case MeetingOrderFieldName:
		return page.NewCursorKey(m.ID, m.Name)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

// AuthorizationAttributes returns the authorization attributes for policy evaluation.
func (m *Meeting) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `SELECT organization_id FROM meetings WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, m.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query meeting authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (m *Meeting) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	meetingID gid.GID,
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    date,
    minutes,
    created_at,
    updated_at
FROM
    meetings
WHERE
    %s
    AND id = @meeting_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"meeting_id": meetingID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query meetings: %w", err)
	}

	meeting, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Meeting])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect meeting: %w", err)
	}

	*m = meeting
	return nil
}

func (m *Meetings) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
	cursor *page.Cursor[MeetingOrderField],
) error {
	q := `
SELECT
    id,
    organization_id,
    name,
    date,
    minutes,
    created_at,
    updated_at
FROM
    meetings
WHERE
    %s
    AND organization_id = @organization_id
    AND %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.NamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query meetings: %w", err)
	}

	meetings, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Meeting])
	if err != nil {
		return fmt.Errorf("cannot collect meetings: %w", err)
	}

	*m = meetings
	return nil
}

func (m *Meetings) CountByOrganizationID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	organizationID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    meetings
WHERE
    %s
    AND organization_id = @organization_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"organization_id": organizationID,
	}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count meetings: %w", err)
	}

	return count, nil
}

func (m *Meeting) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO
    meetings (
        tenant_id,
        id,
        organization_id,
        name,
        date,
        minutes,
        created_at,
        updated_at
    )
VALUES (
    @tenant_id,
    @meeting_id,
    @organization_id,
    @name,
    @date,
    @minutes,
    @created_at,
    @updated_at
);
`

	args := pgx.StrictNamedArgs{
		"tenant_id":       scope.GetTenantID(),
		"meeting_id":      m.ID,
		"organization_id": m.OrganizationID,
		"name":            m.Name,
		"date":            m.Date,
		"minutes":         m.Minutes,
		"created_at":      m.CreatedAt,
		"updated_at":      m.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)

	if err != nil {
		return fmt.Errorf("cannot insert meeting: %w", err)
	}

	return nil
}

func (m *Meeting) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE meetings
SET
    name = @name,
    date = @date,
    minutes = @minutes,
    updated_at = @updated_at
WHERE %s
    AND id = @meeting_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"meeting_id": m.ID,
		"name":       m.Name,
		"date":       m.Date,
		"minutes":    m.Minutes,
		"updated_at": m.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update meeting: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (m *Meeting) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM meetings
WHERE %s
    AND id = @meeting_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"meeting_id": m.ID,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete meeting: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

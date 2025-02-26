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
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	Session struct {
		ID        gid.GID
		UserID    gid.GID
		ExpiredAt time.Time
		CreatedAt time.Time
		UpdatedAt time.Time
	}
)

func (s Session) CursorKey() page.CursorKey {
	return page.NewCursorKey(s.ID, s.CreatedAt)
}

func (s *Session) scan(r pgx.Row) error {
	return r.Scan(
		&s.ID,
		&s.UserID,
		&s.ExpiredAt,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
}

func (s *Session) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	sessionID gid.GID,
) error {
	q := `
SELECT
    id,
    user_id,
    expired_at,
    created_at,
    updated_at
FROM
    usrmgr_sessions
WHERE
    id = @session_id
LIMIT 1;
`

	args := pgx.NamedArgs{"session_id": sessionID}

	r := conn.QueryRow(ctx, q, args)

	s2 := Session{}
	if err := s2.scan(r); err != nil {
		return err
	}

	*s = s2

	return nil
}

func (s *Session) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO
    usrmgr_sessions (id, user_id, expired_at, created_at, updated_at)
VALUES (
    @session_id,
    @user_id,
    @expired_at,
    @created_at,
    @updated_at
)
`

	args := pgx.NamedArgs{
		"session_id": s.ID,
		"user_id":    s.UserID,
		"expired_at": s.ExpiredAt,
		"created_at": s.CreatedAt,
		"updated_at": s.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (s *Session) Update(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
UPDATE usrmgr_sessions
SET
    expired_at = @expired_at,
    updated_at = @updated_at
WHERE
    id = @session_id
`

	args := pgx.NamedArgs{
		"session_id": s.ID,
		"expired_at": s.ExpiredAt,
		"updated_at": s.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	return err
}

func DeleteSession(
	ctx context.Context,
	conn pg.Conn,
	sessionID gid.GID,
) error {
	q := `
DELETE FROM
    usrmgr_sessions
WHERE
    id = @session_id
`

	args := pgx.NamedArgs{"session_id": sessionID}

	_, err := conn.Exec(ctx, q, args)
	return err
}

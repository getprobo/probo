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
	"net"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	Session struct {
		ID              gid.GID       `db:"id"`
		IdentityID      gid.GID       `db:"identity_id"`
		TenantID        *gid.TenantID `db:"tenant_id"`
		MembershipID    *gid.GID      `db:"membership_id"`
		ParentSessionID *gid.GID      `db:"parent_session_id"`
		Data            SessionData   `db:"data"`
		AuthMethod      AuthMethod    `db:"auth_method"`
		AuthenticatedAt time.Time     `db:"authenticated_at"`
		UserAgent       string        `db:"user_agent"`
		IPAddress       net.IP        `db:"ip_address"`
		ExpireReason    *ExpireReason `db:"expire_reason"`
		ExpiredAt       time.Time     `db:"expired_at"`
		CreatedAt       time.Time     `db:"created_at"`
		UpdatedAt       time.Time     `db:"updated_at"`
	}

	Sessions []*Session

	SessionData struct{}

	AuthMethod string
)

const (
	AuthMethodMagicLink AuthMethod = "MAGIC_LINK"
	AuthMethodPassword  AuthMethod = "PASSWORD"
	AuthMethodSAML      AuthMethod = "SAML"
)

func NewRootSession(identityID gid.GID, method AuthMethod, duration time.Duration) *Session {
	now := time.Now()
	return &Session{
		ID:              gid.New(gid.NilTenant, SessionEntityType),
		IdentityID:      identityID,
		ExpiredAt:       now.Add(duration),
		AuthMethod:      method,
		AuthenticatedAt: now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (s Session) CursorKey(orderBy SessionOrderField) page.CursorKey {
	switch orderBy {
	case SessionOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	case SessionOrderFieldExpiredAt:
		return page.NewCursorKey(s.ID, s.ExpiredAt)
	case SessionOrderFieldUpdatedAt:
		return page.NewCursorKey(s.ID, s.UpdatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *Session) IsRootSession() bool {
	return s.ParentSessionID == nil
}

func (s *Session) IsChildSession() bool {
	return s.ParentSessionID != nil
}

func (s *Session) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	sessionID gid.GID,
) error {
	q := `
SELECT
    id,
    identity_id,
    tenant_id,
    membership_id,
    data,
    parent_session_id,
    auth_method,
    authenticated_at,
    expire_reason,
    user_agent,
    ip_address,
    expired_at,
    created_at,
    updated_at
FROM
    iam_sessions
WHERE
    id = @session_id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"session_id": sessionID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query session: %w", err)
	}

	session, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Session])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect session: %w", err)
	}
	*s = session

	return nil
}

// AuthorizationAttributes loads the minimal authorization attributes for policy condition evaluation.
// It is intentionally lightweight and does not populate the Session struct.
func (s *Session) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := `
SELECT
    identity_id
FROM
    iam_sessions
WHERE
    id = $1
LIMIT 1;
`

	var identityID gid.GID
	if err := conn.QueryRow(ctx, q, s.ID).Scan(&identityID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query session iam attributes: %w", err)
	}

	return map[string]string{"identity_id": identityID.String()}, nil
}

func (s *Session) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO
    iam_sessions (id, identity_id, tenant_id, membership_id, data, parent_session_id, auth_method, authenticated_at, expire_reason, user_agent, ip_address, expired_at, created_at, updated_at)
VALUES (
    @session_id,
    @identity_id,
    @tenant_id,
    @membership_id,
    @data,
    @parent_session_id,
    @auth_method,
    @authenticated_at,
    @expire_reason,
    @user_agent,
    @ip_address,
    @expired_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"session_id":        s.ID,
		"identity_id":       s.IdentityID,
		"tenant_id":         s.TenantID,
		"membership_id":     s.MembershipID,
		"data":              s.Data,
		"parent_session_id": s.ParentSessionID,
		"auth_method":       s.AuthMethod,
		"authenticated_at":  s.AuthenticatedAt,
		"expire_reason":     s.ExpireReason,
		"user_agent":        s.UserAgent,
		"ip_address":        s.IPAddress,
		"expired_at":        s.ExpiredAt,
		"created_at":        s.CreatedAt,
		"updated_at":        s.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	return err
}

func (s *Session) Update(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
UPDATE iam_sessions
SET
    expired_at = @expired_at,
    updated_at = @updated_at,
    user_agent = @user_agent,
    ip_address = @ip_address,
    expire_reason = @expire_reason,
    data = @data
WHERE
    id = @session_id
`

	args := pgx.StrictNamedArgs{
		"session_id":    s.ID,
		"user_agent":    s.UserAgent,
		"ip_address":    s.IPAddress,
		"expire_reason": s.ExpireReason,
		"data":          s.Data,
		"expired_at":    s.ExpiredAt,
		"updated_at":    s.UpdatedAt,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (s *Sessions) LoadByIdentityID(ctx context.Context, conn pg.Conn, identityID gid.GID, cursor *page.Cursor[SessionOrderField]) error {
	q := `
SELECT
    id,
    identity_id,
    tenant_id,
    membership_id,
    data,
    parent_session_id,
    auth_method,
    authenticated_at,
    expire_reason,
    user_agent,
    ip_address,
    expired_at,
    created_at,
    updated_at
FROM
    iam_sessions
WHERE
    identity_id = @identity_id
    AND %s
`

	q = fmt.Sprintf(q, cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"identity_id": identityID}
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query sessions: %w", err)
	}

	sessions, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[Session])
	if err != nil {
		return fmt.Errorf("cannot collect sessions: %w", err)
	}

	*s = sessions

	return nil
}

func (s *Sessions) CountByIdentityID(ctx context.Context, conn pg.Conn, identityID gid.GID) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    iam_sessions
WHERE
    identity_id = @identity_id
`

	args := pgx.StrictNamedArgs{"identity_id": identityID}

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (s *Sessions) ExpireAllForIdentityExceptOneSession(ctx context.Context, conn pg.Conn, identityID gid.GID, sessionID gid.GID) (int64, error) {
	q := `
UPDATE iam_sessions
SET
    expired_at = NOW(),
    updated_at = NOW(),
    expire_reason = 'revoked'
WHERE
    id != @session_id
    AND identity_id = @identity_id
    AND expire_reason IS NULL
`

	args := pgx.StrictNamedArgs{
		"session_id":  sessionID,
		"identity_id": identityID,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot query sessions: %w", err)
	}

	return result.RowsAffected(), nil
}

func (s *Session) LoadByRootSessionIDAndMembershipID(
	ctx context.Context,
	conn pg.Conn,
	rootSessionID gid.GID,
	membershipID gid.GID,
) error {
	q := `
SELECT
    id,
    identity_id,
    tenant_id,
    membership_id,
    data,
    parent_session_id,
    auth_method,
    authenticated_at,
    expire_reason,
    user_agent,
    ip_address,
    expired_at,
    created_at,
    updated_at
FROM
    iam_sessions
WHERE
    parent_session_id = @root_session_id
    AND membership_id = @membership_id
ORDER BY created_at DESC
LIMIT 1
`

	args := pgx.StrictNamedArgs{
		"root_session_id": rootSessionID,
		"membership_id":   membershipID,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query session: %w", err)
	}

	session, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Session])
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect session: %w", err)
	}

	*s = session

	return nil
}

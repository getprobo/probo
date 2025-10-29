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
	"time"

	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/page"
	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
)

type (
	Session struct {
		ID        gid.GID     `db:"id"`
		UserID    gid.GID     `db:"user_id"`
		Data      SessionData `db:"data"`
		ExpiredAt time.Time   `db:"expired_at"`
		CreatedAt time.Time   `db:"created_at"`
		UpdatedAt time.Time   `db:"updated_at"`
	}

	// SessionData stores authentication context for a user session
	// Stored as JSONB in database
	SessionData struct {
		// PasswordAuthenticated indicates if user authenticated with email/password
		// Required for accessing organizations without SAML
		PasswordAuthenticated bool `json:"password_authenticated"`

		// SAMLAuthenticatedOrgs tracks which organizations user has SAML-authenticated for
		// Key: organization ID as string, Value: SAML authentication info
		// Required for accessing organizations with SAML enforcement
		SAMLAuthenticatedOrgs map[string]SAMLAuthInfo `json:"saml_authenticated_orgs,omitempty"`
	}

	// SAMLAuthInfo stores SAML authentication details for an organization
	SAMLAuthInfo struct {
		// AuthenticatedAt is when the user SAML-
		AuthenticatedAt time.Time `json:"authenticated_at"`

		// SAMLConfigID is the SAML configuration used for authentication
		SAMLConfigID gid.GID `json:"saml_config_id"`

		// SAMLSubject is the NameID from the SAML assertion (email address)
		SAMLSubject string `json:"saml_subject"`
	}
)

func (s Session) CursorKey(orderBy SessionOrderField) page.CursorKey {
	switch orderBy {
	case SessionOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
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
	data,
    expired_at,
    created_at,
    updated_at
FROM
    sessions
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
		return fmt.Errorf("cannot collect session: %w", err)
	}
	*s = session

	return nil
}

func (s *Session) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO
    sessions (id, user_id, data, expired_at, created_at, updated_at)
VALUES (
    @session_id,
    @user_id,
    @data,
    @expired_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"session_id": s.ID,
		"user_id":    s.UserID,
		"data":       s.Data,
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
UPDATE sessions
SET
    expired_at = @expired_at,
    updated_at = @updated_at,
    data = @data
WHERE
    id = @session_id
`

	args := pgx.StrictNamedArgs{
		"session_id": s.ID,
		"data":       s.Data,
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
    sessions
WHERE
    id = @session_id
`

	args := pgx.StrictNamedArgs{"session_id": sessionID}

	_, err := conn.Exec(ctx, q, args)
	return err
}

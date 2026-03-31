// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type OAuth2AuthorizationCode struct {
	ID                  string                     `db:"id"`
	ClientID            gid.GID                    `db:"client_id"`
	IdentityID          gid.GID                    `db:"identity_id"`
	RedirectURI         string                     `db:"redirect_uri"`
	Scopes              OAuth2Scopes               `db:"scopes"`
	CodeChallenge       *string                    `db:"code_challenge"`
	CodeChallengeMethod *OAuth2CodeChallengeMethod `db:"code_challenge_method"`
	Nonce               *string                    `db:"nonce"`
	AuthTime            time.Time                  `db:"auth_time"`
	CreatedAt           time.Time                  `db:"created_at"`
	ExpiresAt           time.Time                  `db:"expires_at"`
}

func (c *OAuth2AuthorizationCode) Insert(ctx context.Context, conn pg.Conn) error {
	q := `
INSERT INTO iam_oauth2_authorization_codes (
	id,
	client_id,
	identity_id,
	redirect_uri,
	scopes,
	code_challenge,
	code_challenge_method,
	nonce,
	auth_time,
	created_at,
	expires_at
) VALUES (
	@id,
	@client_id,
	@identity_id,
	@redirect_uri,
	@scopes,
	@code_challenge,
	@code_challenge_method,
	@nonce,
	@auth_time,
	@created_at,
	@expires_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                    c.ID,
		"client_id":             c.ClientID,
		"identity_id":           c.IdentityID,
		"redirect_uri":          c.RedirectURI,
		"scopes":                c.Scopes,
		"code_challenge":        c.CodeChallenge,
		"code_challenge_method": c.CodeChallengeMethod,
		"nonce":                 c.Nonce,
		"auth_time":             c.AuthTime,
		"created_at":            c.CreatedAt,
		"expires_at":            c.ExpiresAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert oauth2_authorization_code: %w", err)
	}

	return nil
}

func (c *OAuth2AuthorizationCode) LoadByIDForUpdate(ctx context.Context, conn pg.Conn, id string) error {
	q := `
SELECT
	id,
	client_id,
	identity_id,
	redirect_uri,
	scopes,
	code_challenge,
	code_challenge_method,
	nonce,
	auth_time,
	created_at,
	expires_at
FROM
	iam_oauth2_authorization_codes
WHERE
	id = @id
FOR UPDATE;
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"id": id})
	if err != nil {
		return fmt.Errorf("cannot query oauth2_authorization_code: %w", err)
	}

	code, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OAuth2AuthorizationCode])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect oauth2_authorization_code: %w", err)
	}

	*c = code
	return nil
}

func (c *OAuth2AuthorizationCode) Delete(ctx context.Context, conn pg.Conn) error {
	q := `
DELETE FROM iam_oauth2_authorization_codes
WHERE
	id = @id
`

	_, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"id": c.ID})
	if err != nil {
		return fmt.Errorf("cannot delete oauth2_authorization_code: %w", err)
	}

	return nil
}

func (c *OAuth2AuthorizationCode) DeleteExpired(ctx context.Context, conn pg.Conn, now time.Time) (int64, error) {
	q := `
DELETE FROM iam_oauth2_authorization_codes
WHERE
	expires_at < @now
`

	result, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"now": now})
	if err != nil {
		return 0, fmt.Errorf("cannot delete expired oauth2_authorization_codes: %w", err)
	}

	return result.RowsAffected(), nil
}

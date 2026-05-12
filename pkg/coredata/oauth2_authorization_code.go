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
	"go.probo.inc/probo/pkg/uri"
)

type OAuth2AuthorizationCode struct {
	ID                  gid.GID                    `db:"id"`
	HashedValue         []byte                     `db:"hashed_value"`
	ClientID            gid.GID                    `db:"client_id"`
	IdentityID          gid.GID                    `db:"identity_id"`
	RedirectURI         uri.URI                    `db:"redirect_uri"`
	Scopes              OAuth2Scopes               `db:"scopes"`
	CodeChallenge       *string                    `db:"code_challenge"`
	CodeChallengeMethod *OAuth2CodeChallengeMethod `db:"code_challenge_method"`
	Nonce               *string                    `db:"nonce"`
	AuthTime            time.Time                  `db:"auth_time"`
	CreatedAt           time.Time                  `db:"created_at"`
	ExpiresAt           time.Time                  `db:"expires_at"`
	RedeemedAt          *time.Time                 `db:"redeemed_at"`
	AccessTokenID       *gid.GID                   `db:"access_token_id"`
}

func (c *OAuth2AuthorizationCode) Insert(ctx context.Context, conn pg.Tx) error {
	q := `
INSERT INTO iam_oauth2_authorization_codes (
	id,
	hashed_value,
	client_id,
	identity_id,
	redirect_uri,
	scopes,
	code_challenge,
	code_challenge_method,
	nonce,
	auth_time,
	created_at,
	expires_at,
	redeemed_at,
	access_token_id
) VALUES (
	@id,
	@hashed_value,
	@client_id,
	@identity_id,
	@redirect_uri,
	@scopes,
	@code_challenge,
	@code_challenge_method,
	@nonce,
	@auth_time,
	@created_at,
	@expires_at,
	@redeemed_at,
	@access_token_id
)
`

	args := pgx.StrictNamedArgs{
		"id":                    c.ID,
		"hashed_value":          c.HashedValue,
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
		"redeemed_at":           c.RedeemedAt,
		"access_token_id":       c.AccessTokenID,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert oauth2_authorization_code: %w", err)
	}

	return nil
}

func (c *OAuth2AuthorizationCode) LoadByHashForUpdate(
	ctx context.Context,
	conn pg.Tx,
	hashedValue []byte,
	clientID gid.GID,
) error {
	q := `
SELECT
	id,
	hashed_value,
	client_id,
	identity_id,
	redirect_uri,
	scopes,
	code_challenge,
	code_challenge_method,
	nonce,
	auth_time,
	created_at,
	expires_at,
	redeemed_at,
	access_token_id
FROM
	iam_oauth2_authorization_codes
WHERE
	hashed_value = @hashed_value
	AND client_id = @client_id
FOR UPDATE;
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{"hashed_value": hashedValue, "client_id": clientID},
	)
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

func (c *OAuth2AuthorizationCode) Redeem(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
	accessTokenID gid.GID,
) error {
	q := `
UPDATE iam_oauth2_authorization_codes
SET
	redeemed_at = @redeemed_at,
	access_token_id = @access_token_id
WHERE
	id = @id
`

	args := pgx.StrictNamedArgs{
		"id":              c.ID,
		"redeemed_at":     now,
		"access_token_id": accessTokenID,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot redeem oauth2_authorization_code: %w", err)
	}

	c.RedeemedAt = &now
	c.AccessTokenID = &accessTokenID

	return nil
}

func (c *OAuth2AuthorizationCode) Delete(ctx context.Context, conn pg.Querier) error {
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

func (c *OAuth2AuthorizationCode) DeleteExpired(ctx context.Context, conn pg.Tx, now time.Time) (int64, error) {
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

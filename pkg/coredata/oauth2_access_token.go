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

type (
	OAuth2AccessToken struct {
		ID          gid.GID      `db:"id"`
		HashedValue []byte       `db:"hashed_value"`
		ClientID    gid.GID      `db:"client_id"`
		IdentityID  gid.GID      `db:"identity_id"`
		Scopes      OAuth2Scopes `db:"scopes"`
		CreatedAt   time.Time    `db:"created_at"`
		ExpiresAt   time.Time    `db:"expires_at"`
	}
)

func (t *OAuth2AccessToken) ExpiresIn(now time.Time) time.Duration {
	return t.ExpiresAt.Sub(now)
}

func (t *OAuth2AccessToken) Insert(ctx context.Context, conn pg.Tx) error {
	q := `
INSERT INTO iam_oauth2_access_tokens (
	id,
	hashed_value,
	client_id,
	identity_id,
	scopes,
	created_at,
	expires_at
) VALUES (
	@id,
	@hashed_value,
	@client_id,
	@identity_id,
	@scopes,
	@created_at,
	@expires_at
)
`

	args := pgx.StrictNamedArgs{
		"id":           t.ID,
		"hashed_value": t.HashedValue,
		"client_id":    t.ClientID,
		"identity_id":  t.IdentityID,
		"scopes":       t.Scopes,
		"created_at":   t.CreatedAt,
		"expires_at":   t.ExpiresAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert oauth2_access_token: %w", err)
	}

	return nil
}

func (t *OAuth2AccessToken) LoadByHashedValue(ctx context.Context, conn pg.Querier, hashedValue []byte) error {
	q := `
SELECT
	id,
	hashed_value,
	client_id,
	identity_id,
	scopes,
	created_at,
	expires_at
FROM
	iam_oauth2_access_tokens
WHERE
	hashed_value = @hashed_value
LIMIT 1;
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"hashed_value": hashedValue})
	if err != nil {
		return fmt.Errorf("cannot query oauth2_access_token: %w", err)
	}

	token, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OAuth2AccessToken])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect oauth2_access_token: %w", err)
	}

	*t = token
	return nil
}

func (t *OAuth2AccessToken) LoadByHashedValueAndClientID(
	ctx context.Context,
	conn pg.Querier,
	hashedValue []byte,
	clientID gid.GID,
) error {
	q := `
SELECT
	id,
	hashed_value,
	client_id,
	identity_id,
	scopes,
	created_at,
	expires_at
FROM
	iam_oauth2_access_tokens
WHERE
	hashed_value = @hashed_value
	AND client_id = @client_id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{
		"hashed_value": hashedValue,
		"client_id":    clientID,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query oauth2_access_token: %w", err)
	}

	token, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OAuth2AccessToken])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect oauth2_access_token: %w", err)
	}

	*t = token
	return nil
}

func (t *OAuth2AccessToken) Delete(ctx context.Context, conn pg.Tx) error {
	q := `
DELETE FROM iam_oauth2_access_tokens
WHERE
	id = @id
`

	_, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"id": t.ID})
	if err != nil {
		return fmt.Errorf("cannot delete oauth2_access_token: %w", err)
	}

	return nil
}

func (t *OAuth2AccessToken) DeleteExpired(ctx context.Context, conn pg.Tx, now time.Time) (int64, error) {
	q := `
DELETE FROM iam_oauth2_access_tokens
WHERE
	expires_at < @now
`

	result, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"now": now})
	if err != nil {
		return 0, fmt.Errorf("cannot delete expired oauth2_access_tokens: %w", err)
	}

	return result.RowsAffected(), nil
}

func (t *OAuth2AccessToken) DeleteByClientAndIdentity(
	ctx context.Context,
	conn pg.Tx,
	clientID gid.GID,
	identityID gid.GID,
) (int64, error) {
	q := `
DELETE FROM iam_oauth2_access_tokens
WHERE
	client_id = @client_id
	AND identity_id = @identity_id
`

	args := pgx.StrictNamedArgs{
		"client_id":   clientID,
		"identity_id": identityID,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot delete oauth2_access_tokens by client and identity: %w", err)
	}

	return result.RowsAffected(), nil
}

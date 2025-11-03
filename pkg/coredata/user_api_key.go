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
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	UserAPIKey struct {
		ID        gid.GID   `db:"id"`
		UserID    gid.GID   `db:"user_id"`
		Name      string    `db:"name"`
		ExpiresAt time.Time `db:"expires_at"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}

	UserAPIKeys []*UserAPIKey

	ErrUserAPIKeyNotFound struct {
		Identifier string
	}
)

func (e ErrUserAPIKeyNotFound) Error() string {
	return fmt.Sprintf("user api key not found: %q", e.Identifier)
}

func (a *UserAPIKey) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	apiKeyID gid.GID,
) error {
	q := `
SELECT
    id,
    user_id,
    name,
    expires_at,
    created_at,
    updated_at
FROM
    auth_user_api_keys
WHERE
    id = @api_key_id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"api_key_id": apiKeyID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query user api key: %w", err)
	}

	apiKey, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[UserAPIKey])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &ErrUserAPIKeyNotFound{Identifier: apiKeyID.String()}
		}

		return fmt.Errorf("cannot collect user api key: %w", err)
	}

	*a = apiKey

	return nil
}

func (a *UserAPIKeys) LoadByUserID(
	ctx context.Context,
	conn pg.Conn,
	userID gid.GID,
) error {
	q := `
SELECT
    id,
    user_id,
    name,
    expires_at,
    created_at,
    updated_at
FROM
    auth_user_api_keys
WHERE
    user_id = @user_id
ORDER BY created_at DESC;
`

	args := pgx.StrictNamedArgs{"user_id": userID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query user api keys: %w", err)
	}

	apiKeys, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[UserAPIKey])
	if err != nil {
		return fmt.Errorf("cannot collect user api keys: %w", err)
	}

	*a = apiKeys

	return nil
}

func (a *UserAPIKey) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO
    auth_user_api_keys (id, user_id, name, expires_at, created_at, updated_at)
VALUES (
    @api_key_id,
    @user_id,
    @name,
    @expires_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"api_key_id": a.ID,
		"user_id":    a.UserID,
		"name":       a.Name,
		"expires_at": a.ExpiresAt,
		"created_at": a.CreatedAt,
		"updated_at": a.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert user api key: %w", err)
	}

	return nil
}

func (a *UserAPIKey) Update(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
UPDATE
    auth_user_api_keys
SET
    name = @name,
    expires_at = @expires_at,
    updated_at = @updated_at
WHERE
    id = @api_key_id
`

	args := pgx.StrictNamedArgs{
		"api_key_id": a.ID,
		"name":       a.Name,
		"expires_at": a.ExpiresAt,
		"updated_at": a.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update user api key: %w", err)
	}

	return nil
}

func (a *UserAPIKey) Delete(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
DELETE FROM
    auth_user_api_keys
WHERE
    id = @api_key_id
`

	args := pgx.StrictNamedArgs{"api_key_id": a.ID}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete user api key: %w", err)
	}

	return nil
}

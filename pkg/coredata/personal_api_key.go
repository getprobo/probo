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
	"go.probo.inc/probo/pkg/page"
)

type (
	PersonalAPIKey struct {
		ID           gid.GID       `db:"id"`
		IdentityID   gid.GID       `db:"identity_id"`
		Name         string        `db:"name"`
		ExpiresAt    time.Time     `db:"expires_at"`
		ExpireReason *ExpireReason `db:"expire_reason"`
		LastUsedAt   *time.Time    `db:"last_used_at"`
		CreatedAt    time.Time     `db:"created_at"`
		UpdatedAt    time.Time     `db:"updated_at"`
	}

	PersonalAPIKeys []*PersonalAPIKey
)

func (a *PersonalAPIKey) CursorKey(orderBy PersonalAPIKeyOrderField) page.CursorKey {
	switch orderBy {
	case PersonalAPIKeyOrderFieldCreatedAt:
		return page.NewCursorKey(a.ID, a.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (a *PersonalAPIKey) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	apiKeyID gid.GID,
) error {
	q := `
SELECT
    id,
    identity_id,
    name,
    expires_at,
    expire_reason,
    last_used_at,
    created_at,
    updated_at
FROM
    iam_personal_api_keys
WHERE
    id = @api_key_id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"api_key_id": apiKeyID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query personal api key: %w", err)
	}

	apiKey, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[PersonalAPIKey])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect personal api key: %w", err)
	}

	*a = apiKey

	return nil
}

func (a *PersonalAPIKey) AuthorizationAttributes(ctx context.Context, conn pg.Conn) (map[string]string, error) {
	q := "SELECT identity_id FROM iam_personal_api_keys WHERE id = $1 LIMIT 1;"

	var identityID gid.GID
	if err := conn.QueryRow(ctx, q, a.ID).Scan(&identityID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query personal api key iam attributes: %w", err)
	}

	return map[string]string{"identity_id": identityID.String()}, nil
}

func (a *PersonalAPIKeys) LoadByIdentityID(
	ctx context.Context,
	conn pg.Conn,
	identityID gid.GID,
) error {
	q := `
SELECT
    id,
    identity_id,
    name,
    expires_at,
    expire_reason,
    last_used_at,
    created_at,
    updated_at
FROM
    iam_personal_api_keys
WHERE
    identity_id = @identity_id
ORDER BY created_at DESC;
`

	args := pgx.StrictNamedArgs{"identity_id": identityID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query personal api keys: %w", err)
	}

	apiKeys, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[PersonalAPIKey])
	if err != nil {
		return fmt.Errorf("cannot collect personal api keys: %w", err)
	}

	*a = apiKeys

	return nil
}

func (a *PersonalAPIKeys) CountByIdentityID(ctx context.Context, conn pg.Conn, identityID gid.GID) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM
    iam_personal_api_keys
WHERE
    identity_id = @identity_id
ORDER BY created_at DESC;
`

	args := pgx.StrictNamedArgs{"identity_id": identityID}
	row := conn.QueryRow(ctx, q, args)
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot scan count: %w", err)
	}

	return count, nil
}

func (a *PersonalAPIKey) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO
    iam_personal_api_keys (id, identity_id, name, expires_at, expire_reason, last_used_at, created_at, updated_at)
VALUES (
    @api_key_id,
    @identity_id,
    @name,
    @expires_at,
    @expire_reason,
    @last_used_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"api_key_id":    a.ID,
		"identity_id":   a.IdentityID,
		"name":          a.Name,
		"expires_at":    a.ExpiresAt,
		"expire_reason": a.ExpireReason,
		"last_used_at":  a.LastUsedAt,
		"created_at":    a.CreatedAt,
		"updated_at":    a.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert personal api key: %w", err)
	}

	return nil
}

func (a *PersonalAPIKey) Update(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
UPDATE
    iam_personal_api_keys
SET
    name = @name,
    expires_at = @expires_at,
    expire_reason = @expire_reason,
    last_used_at = @last_used_at,
    updated_at = @updated_at
WHERE
    id = @api_key_id
`

	args := pgx.StrictNamedArgs{
		"api_key_id":    a.ID,
		"name":          a.Name,
		"expires_at":    a.ExpiresAt,
		"expire_reason": a.ExpireReason,
		"last_used_at":  a.LastUsedAt,
		"updated_at":    a.UpdatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update personal api key: %w", err)
	}

	return nil
}

func (a *PersonalAPIKey) Delete(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
DELETE FROM
    iam_personal_api_keys
WHERE
    id = @api_key_id
`

	args := pgx.StrictNamedArgs{"api_key_id": a.ID}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete personal api key: %w", err)
	}

	return nil
}

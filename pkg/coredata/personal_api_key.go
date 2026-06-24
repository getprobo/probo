// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package coredata

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
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
	conn pg.Querier,
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

func (a *PersonalAPIKey) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `
SELECT
    id,
    identity_id
FROM
    iam_personal_api_keys
WHERE
    id = ANY(@resource_ids::text[])
`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query personal api key authorization attributes: %w", err)
	}
	defer rows.Close()

	attrsByID := make(policy.AttributesByID, len(resourceIDs))

	for rows.Next() {
		var (
			id         gid.GID
			identityID gid.GID
		)

		err = rows.Scan(&id, &identityID)
		if err != nil {
			return nil, fmt.Errorf("cannot scan personal api key authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{"identity_id": identityID.String()}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate personal api key authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (a *PersonalAPIKeys) LoadByIdentityID(
	ctx context.Context,
	conn pg.Querier,
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

func (a *PersonalAPIKeys) CountByIdentityID(ctx context.Context, conn pg.Querier, identityID gid.GID) (int, error) {
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
	conn pg.Tx,
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
	conn pg.Tx,
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
	conn pg.Tx,
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

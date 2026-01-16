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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type Token struct {
	ID          gid.GID   `db:"id"`
	HashedValue []byte    `db:"hashed_value"`
	CreatedAt   time.Time `db:"created_at"`
}

func (t *Token) LoadByHashedValueForUpdate(
	ctx context.Context,
	conn pg.Conn,
	hashedValue []byte,
) error {
	q := `
SELECT
    id,
    hashed_value,
    created_at
FROM
    iam_tokens
WHERE
    hashed_value = @hashed_value
LIMIT 1
FOR UPDATE;
`
	args := pgx.StrictNamedArgs{"hashed_value": hashedValue}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query iam_tokens: %w", err)
	}

	token, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[Token])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect iam_tokens: %w", err)
	}

	*t = token

	return nil
}

func (t *Token) Insert(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
INSERT INTO iam_tokens(
    id,
    hashed_value,
    created_at
) VALUES (
    @id,
    @hashed_value,
    @created_at
)
`

	args := pgx.StrictNamedArgs{
		"id":           t.ID,
		"hashed_value": t.HashedValue,
		"created_at":   t.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "iam_tokens_hashed_value_unique" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert iam_tokens: %w", err)
	}

	return nil
}

func (t *Token) Delete(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
DELETE FROM iam_tokens
WHERE
    id = @id
`

	args := pgx.StrictNamedArgs{"id": t.ID}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete iam_tokens: %w", err)
	}

	return nil
}

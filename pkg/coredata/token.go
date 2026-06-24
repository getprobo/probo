// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	conn pg.Tx,
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
	conn pg.Tx,
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
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
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
	conn pg.Tx,
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

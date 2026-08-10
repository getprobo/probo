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

type SlackIdentityBindTokenUse struct {
	HashedToken []byte    `db:"hashed_token"`
	IdentityID  gid.GID   `db:"identity_id"`
	UsedAt      time.Time `db:"used_at"`
	ExpiresAt   time.Time `db:"expires_at"`
}

func (u *SlackIdentityBindTokenUse) Insert(ctx context.Context, conn pg.Tx) error {
	q := `
INSERT INTO slack_identity_bind_token_uses (
    hashed_token,
    identity_id,
    used_at,
    expires_at
) VALUES (
    @hashed_token,
    @identity_id,
    @used_at,
    @expires_at
)
`

	args := pgx.StrictNamedArgs{
		"hashed_token": u.HashedToken,
		"identity_id":  u.IdentityID,
		"used_at":      u.UsedAt,
		"expires_at":   u.ExpiresAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "slack_identity_bind_token_uses_pkey" {
			return ErrResourceAlreadyExists
		}

		return fmt.Errorf("cannot insert slack identity bind token use: %w", err)
	}

	return nil
}

func (u *SlackIdentityBindTokenUse) LoadByHashedToken(
	ctx context.Context,
	conn pg.Querier,
	hashedToken []byte,
) error {
	q := `
SELECT
    hashed_token,
    identity_id,
    used_at,
    expires_at
FROM
    slack_identity_bind_token_uses
WHERE
    hashed_token = @hashed_token
LIMIT 1
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"hashed_token": hashedToken})
	if err != nil {
		return fmt.Errorf("cannot query slack identity bind token use: %w", err)
	}

	use, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[SlackIdentityBindTokenUse])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect slack identity bind token use: %w", err)
	}

	*u = use

	return nil
}

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

type DeviceEnrollmentToken struct {
	ID          gid.GID      `db:"id"`
	TenantID    gid.TenantID `db:"tenant_id"`
	DeviceID    gid.GID      `db:"device_id"`
	HashedValue []byte       `db:"hashed_value"`
	ExpiresAt   time.Time    `db:"expires_at"`
	CreatedAt   time.Time    `db:"created_at"`
}

func (t *DeviceEnrollmentToken) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO device_enrollment_tokens (
    id,
    tenant_id,
    device_id,
    hashed_value,
    expires_at,
    created_at
) VALUES (
    @id,
    @tenant_id,
    @device_id,
    @hashed_value,
    @expires_at,
    @created_at
)
`

	args := pgx.StrictNamedArgs{
		"id":           t.ID,
		"tenant_id":    scope.GetTenantID(),
		"device_id":    t.DeviceID,
		"hashed_value": t.HashedValue,
		"expires_at":   t.ExpiresAt,
		"created_at":   t.CreatedAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "device_enrollment_tokens_hashed_value_unique" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert device_enrollment_token: %w", err)
	}

	return nil
}

func (t *DeviceEnrollmentToken) LoadByHashedValueForUpdate(
	ctx context.Context,
	conn pg.Tx,
	hashedValue []byte,
) error {
	q := `
SELECT
    id,
    tenant_id,
    device_id,
    hashed_value,
    expires_at,
    created_at
FROM
    device_enrollment_tokens
WHERE
    hashed_value = @hashed_value
LIMIT 1
FOR UPDATE;
`

	args := pgx.StrictNamedArgs{"hashed_value": hashedValue}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query device_enrollment_tokens: %w", err)
	}

	token, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DeviceEnrollmentToken])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect device_enrollment_token: %w", err)
	}

	*t = token

	return nil
}

func (t *DeviceEnrollmentToken) DeleteExpired(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
) (int64, error) {
	q := `
DELETE FROM device_enrollment_tokens
WHERE
    expires_at < @now
`

	result, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"now": now})
	if err != nil {
		return 0, fmt.Errorf("cannot delete expired device_enrollment_tokens: %w", err)
	}

	return result.RowsAffected(), nil
}

func (t *DeviceEnrollmentToken) Delete(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
DELETE FROM device_enrollment_tokens
WHERE
    id = @id
`

	args := pgx.StrictNamedArgs{"id": t.ID}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete device_enrollment_token: %w", err)
	}

	return nil
}

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
	// OAuth2UserCode represents a raw 8-character user code for the device flow.
	OAuth2UserCode string

	OAuth2DeviceCode struct {
		ID             string                 `db:"id"`
		DeviceCodeHash []byte                 `db:"device_code_hash"`
		UserCode       OAuth2UserCode         `db:"user_code"`
		ClientID       gid.GID                `db:"client_id"`
		Scopes         OAuth2Scopes           `db:"scopes"`
		IdentityID     *gid.GID               `db:"identity_id"`
		Status         OAuth2DeviceCodeStatus `db:"status"`
		LastPolledAt   *time.Time             `db:"last_polled_at"`
		PollInterval   int                    `db:"poll_interval"`
		CreatedAt      time.Time              `db:"created_at"`
		ExpiresAt      time.Time              `db:"expires_at"`
	}
)

// Format returns the user code formatted as XXXX-XXXX for display.
func (c OAuth2UserCode) Format() string {
	if len(c) != 8 {
		panic(fmt.Sprintf("invalid user code length: %d", len(c)))
	}

	return string(c[:4]) + "-" + string(c[4:])
}

func (d *OAuth2DeviceCode) Insert(ctx context.Context, conn pg.Tx) error {
	q := `
INSERT INTO iam_oauth2_device_codes (
	id,
	device_code_hash,
	user_code,
	client_id,
	scopes,
	identity_id,
	status,
	last_polled_at,
	poll_interval,
	created_at,
	expires_at
) VALUES (
	@id,
	@device_code_hash,
	@user_code,
	@client_id,
	@scopes,
	@identity_id,
	@status,
	@last_polled_at,
	@poll_interval,
	@created_at,
	@expires_at
)
`

	args := pgx.StrictNamedArgs{
		"id":               d.ID,
		"device_code_hash": d.DeviceCodeHash,
		"user_code":        d.UserCode,
		"client_id":        d.ClientID,
		"scopes":           d.Scopes,
		"identity_id":      d.IdentityID,
		"status":           d.Status,
		"last_polled_at":   d.LastPolledAt,
		"poll_interval":    d.PollInterval,
		"created_at":       d.CreatedAt,
		"expires_at":       d.ExpiresAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert oauth2_device_code: %w", err)
	}

	return nil
}

func (d *OAuth2DeviceCode) LoadByUserCode(
	ctx context.Context,
	conn pg.Querier,
	userCode string,
) error {
	q := `
SELECT
	id,
	device_code_hash,
	user_code,
	client_id,
	scopes,
	identity_id,
	status,
	last_polled_at,
	poll_interval,
	created_at,
	expires_at
FROM
	iam_oauth2_device_codes
WHERE
	user_code = @user_code
LIMIT 1;
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"user_code": userCode})
	if err != nil {
		return fmt.Errorf("cannot query oauth2_device_code: %w", err)
	}

	code, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OAuth2DeviceCode])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect oauth2_device_code: %w", err)
	}

	*d = code
	return nil
}

func (d *OAuth2DeviceCode) LoadByDeviceCodeHashForUpdate(
	ctx context.Context,
	conn pg.Querier,
	hashedValue []byte,
	clientID gid.GID,
) error {
	q := `
SELECT
	id,
	device_code_hash,
	user_code,
	client_id,
	scopes,
	identity_id,
	status,
	last_polled_at,
	poll_interval,
	created_at,
	expires_at
FROM
	iam_oauth2_device_codes
WHERE
	device_code_hash = @device_code_hash
	AND client_id = @client_id
FOR UPDATE;
`

	rows, err := conn.Query(
		ctx,
		q,
		pgx.StrictNamedArgs{
			"device_code_hash": hashedValue,
			"client_id":        clientID,
		},
	)
	if err != nil {
		return fmt.Errorf("cannot query oauth2_device_code: %w", err)
	}

	code, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[OAuth2DeviceCode])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect oauth2_device_code: %w", err)
	}

	*d = code
	return nil
}

func (d *OAuth2DeviceCode) UpdateStatus(
	ctx context.Context,
	conn pg.Tx,
	status OAuth2DeviceCodeStatus,
	identityID *gid.GID,
) error {
	q := `
UPDATE iam_oauth2_device_codes
SET
	status = @status,
	identity_id = @identity_id
WHERE
	id = @id
`

	args := pgx.StrictNamedArgs{
		"id":          d.ID,
		"status":      status,
		"identity_id": identityID,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot update oauth2_device_code status: %w", err)
	}

	return nil
}

func (d *OAuth2DeviceCode) UpdateLastPolledAt(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
	pollInterval int,
) error {
	q := `
UPDATE iam_oauth2_device_codes
SET
	last_polled_at = @last_polled_at,
	poll_interval = @poll_interval
WHERE
	id = @id
`

	args := pgx.StrictNamedArgs{
		"id":             d.ID,
		"last_polled_at": now,
		"poll_interval":  pollInterval,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot update oauth2_device_code last_polled_at: %w", err)
	}

	return nil
}

func (d *OAuth2DeviceCode) Delete(ctx context.Context, conn pg.Tx) error {
	q := `
DELETE FROM iam_oauth2_device_codes
WHERE
	id = @id
`

	args := pgx.StrictNamedArgs{"id": d.ID}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete oauth2_device_code: %w", err)
	}

	return nil
}

func (d *OAuth2DeviceCode) DeleteExpired(ctx context.Context, conn pg.Tx, now time.Time) (int64, error) {
	q := `
DELETE FROM iam_oauth2_device_codes
WHERE
	expires_at < @now
`

	result, err := conn.Exec(ctx, q, pgx.StrictNamedArgs{"now": now})
	if err != nil {
		return 0, fmt.Errorf("cannot delete expired oauth2_device_codes: %w", err)
	}

	return result.RowsAffected(), nil
}

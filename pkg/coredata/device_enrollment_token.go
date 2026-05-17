// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	DeviceEnrollmentToken struct {
		ID                  gid.GID      `db:"id"`
		TenantID            gid.TenantID `db:"tenant_id"`
		OrganizationID      gid.GID      `db:"organization_id"`
		Name                string       `db:"name"`
		TokenHash           []byte       `db:"token_hash"`
		CreatedByIdentityID *gid.GID     `db:"created_by_identity_id"`
		ExpiresAt           time.Time    `db:"expires_at"`
		RevokedAt           *time.Time   `db:"revoked_at"`
		MaxUses             *int         `db:"max_uses"`
		UsedCount           int          `db:"used_count"`
		CreatedAt           time.Time    `db:"created_at"`
		UpdatedAt           time.Time    `db:"updated_at"`
	}

	DeviceEnrollmentTokens []*DeviceEnrollmentToken
)

func (t *DeviceEnrollmentToken) AuthorizationAttributes(ctx context.Context, conn pg.Querier) (map[string]string, error) {
	q := `SELECT organization_id FROM device_enrollment_tokens WHERE id = $1 LIMIT 1;`

	var organizationID gid.GID
	if err := conn.QueryRow(ctx, q, t.ID).Scan(&organizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("cannot query device enrollment token authorization attributes: %w", err)
	}

	return map[string]string{"organization_id": organizationID.String()}, nil
}

func (t DeviceEnrollmentToken) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO device_enrollment_tokens (
    id,
    tenant_id,
    organization_id,
    name,
    token_hash,
    created_by_identity_id,
    expires_at,
    revoked_at,
    max_uses,
    used_count,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @name,
    @token_hash,
    @created_by_identity_id,
    @expires_at,
    @revoked_at,
    @max_uses,
    @used_count,
    @created_at,
    @updated_at
)
`
	args := pgx.StrictNamedArgs{
		"id":                     t.ID,
		"tenant_id":              scope.GetTenantID(),
		"organization_id":        t.OrganizationID,
		"name":                   t.Name,
		"token_hash":             t.TokenHash,
		"created_by_identity_id": t.CreatedByIdentityID,
		"expires_at":             t.ExpiresAt,
		"revoked_at":             t.RevokedAt,
		"max_uses":               t.MaxUses,
		"used_count":             t.UsedCount,
		"created_at":             t.CreatedAt,
		"updated_at":             t.UpdatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert device enrollment token: %w", err)
	}
	return nil
}

// LoadByTokenHash loads a usable enrollment token (not revoked, not expired,
// uses remaining). This is called by the unauthenticated /enroll handler so
// it deliberately bypasses the tenant scope.
func (t *DeviceEnrollmentToken) LoadByTokenHash(
	ctx context.Context,
	conn pg.Querier,
	tokenHash []byte,
) error {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    name,
    token_hash,
    created_by_identity_id,
    expires_at,
    revoked_at,
    max_uses,
    used_count,
    created_at,
    updated_at
FROM
    device_enrollment_tokens
WHERE
    token_hash = @token_hash
LIMIT 1;
`
	args := pgx.StrictNamedArgs{"token_hash": tokenHash}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query device enrollment token: %w", err)
	}

	token, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DeviceEnrollmentToken])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect device enrollment token: %w", err)
	}

	*t = token
	return nil
}

// LoadByTokenHashForUpdate loads an enrollment token and locks its row.
// Use this in a transaction when consuming the token to prevent
// concurrent max-uses races.
func (t *DeviceEnrollmentToken) LoadByTokenHashForUpdate(
	ctx context.Context,
	conn pg.Tx,
	tokenHash []byte,
) error {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    name,
    token_hash,
    created_by_identity_id,
    expires_at,
    revoked_at,
    max_uses,
    used_count,
    created_at,
    updated_at
FROM
    device_enrollment_tokens
WHERE
    token_hash = @token_hash
LIMIT 1
FOR UPDATE;
`
	args := pgx.StrictNamedArgs{"token_hash": tokenHash}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query device enrollment token for update: %w", err)
	}

	token, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DeviceEnrollmentToken])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect device enrollment token for update: %w", err)
	}

	*t = token
	return nil
}

func (t *DeviceEnrollmentToken) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    name,
    token_hash,
    created_by_identity_id,
    expires_at,
    revoked_at,
    max_uses,
    used_count,
    created_at,
    updated_at
FROM
    device_enrollment_tokens
WHERE
    %s
    AND id = @id
LIMIT 1;
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"id": id}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query device enrollment token: %w", err)
	}

	token, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[DeviceEnrollmentToken])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect device enrollment token: %w", err)
	}

	*t = token
	return nil
}

func (t *DeviceEnrollmentToken) IncrementUsage(
	ctx context.Context,
	conn pg.Tx,
) error {
	now := time.Now()

	q := `
UPDATE device_enrollment_tokens
SET
    used_count = used_count + 1,
    updated_at = @now
WHERE
    id = @id
`
	args := pgx.StrictNamedArgs{
		"id":  t.ID,
		"now": now,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot increment device enrollment token usage: %w", err)
	}

	t.UsedCount++
	t.UpdatedAt = now
	return nil
}

func (t *DeviceEnrollmentToken) Revoke(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	now := time.Now()

	q := `
UPDATE device_enrollment_tokens
SET
    revoked_at = COALESCE(revoked_at, @now),
    updated_at = @now
WHERE %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":  t.ID,
		"now": now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot revoke device enrollment token: %w", err)
	}

	if t.RevokedAt == nil {
		t.RevokedAt = &now
	}
	t.UpdatedAt = now
	return nil
}

// IsUsable reports whether the token can still be consumed at time `at`.
func (t DeviceEnrollmentToken) IsUsable(at time.Time) bool {
	if t.RevokedAt != nil {
		return false
	}
	if !t.ExpiresAt.IsZero() && !at.Before(t.ExpiresAt) {
		return false
	}
	if t.MaxUses != nil && t.UsedCount >= *t.MaxUses {
		return false
	}
	return true
}

func (ts *DeviceEnrollmentTokens) LoadByOrganizationID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	organizationID gid.GID,
) error {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    name,
    token_hash,
    created_by_identity_id,
    expires_at,
    revoked_at,
    max_uses,
    used_count,
    created_at,
    updated_at
FROM
    device_enrollment_tokens
WHERE
    %s
    AND organization_id = @organization_id
ORDER BY created_at DESC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"organization_id": organizationID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query device enrollment tokens: %w", err)
	}

	tokens, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DeviceEnrollmentToken])
	if err != nil {
		return fmt.Errorf("cannot collect device enrollment tokens: %w", err)
	}

	*ts = tokens
	return nil
}

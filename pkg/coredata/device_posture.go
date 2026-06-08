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
	"fmt"
	"maps"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	DevicePosture struct {
		ID             gid.GID             `db:"id"`
		TenantID       gid.TenantID        `db:"tenant_id"`
		OrganizationID gid.GID             `db:"organization_id"`
		DeviceID       gid.GID             `db:"device_id"`
		CheckKey       string              `db:"check_key"`
		Status         DevicePostureStatus `db:"status"`
		Evidence       json.RawMessage     `db:"evidence"`
		ObservedAt     time.Time           `db:"observed_at"`
		CreatedAt      time.Time           `db:"created_at"`
	}

	DevicePostures []*DevicePosture
)

func (p DevicePosture) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	evidence := p.Evidence
	if len(evidence) == 0 {
		evidence = emptyJSONObject
	}

	q := `
INSERT INTO device_postures (
    id,
    tenant_id,
    organization_id,
    device_id,
    check_key,
    status,
    evidence,
    observed_at,
    created_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @device_id,
    @check_key,
    @status,
    @evidence,
    @observed_at,
    @created_at
)
`
	args := pgx.StrictNamedArgs{
		"id":              p.ID,
		"tenant_id":       scope.GetTenantID(),
		"organization_id": p.OrganizationID,
		"device_id":       p.DeviceID,
		"check_key":       p.CheckKey,
		"status":          p.Status,
		"evidence":        evidence,
		"observed_at":     p.ObservedAt,
		"created_at":      p.CreatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert device posture: %w", err)
	}
	return nil
}

// LoadLatestByDeviceID loads the latest posture row for each check_key on
// the given device.
func (p *DevicePostures) LoadLatestByDeviceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	deviceID gid.GID,
) error {
	q := `
SELECT DISTINCT ON (check_key)
    id,
    tenant_id,
    organization_id,
    device_id,
    check_key,
    status,
    evidence,
    observed_at,
    created_at
FROM
    device_postures
WHERE
    %s
    AND device_id = @device_id
ORDER BY check_key, observed_at DESC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"device_id": deviceID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query latest device postures: %w", err)
	}

	postures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DevicePosture])
	if err != nil {
		return fmt.Errorf("cannot collect device postures: %w", err)
	}

	*p = postures
	return nil
}

// LoadHistoryByDeviceIDAndCheckKey returns the most recent N entries for one
// (device, check_key) pair, newest first.
func (p *DevicePostures) LoadHistoryByDeviceIDAndCheckKey(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	deviceID gid.GID,
	checkKey string,
	limit int,
) error {
	if limit <= 0 {
		limit = 100
	}

	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    device_id,
    check_key,
    status,
    evidence,
    observed_at,
    created_at
FROM
    device_postures
WHERE
    %s
    AND device_id = @device_id
    AND check_key = @check_key
ORDER BY observed_at DESC
LIMIT @limit
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"device_id": deviceID,
		"check_key": checkKey,
		"limit":     limit,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query device posture history: %w", err)
	}

	postures, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DevicePosture])
	if err != nil {
		return fmt.Errorf("cannot collect device posture history: %w", err)
	}

	*p = postures
	return nil
}

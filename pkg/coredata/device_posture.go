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
	"encoding/json"
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
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

const (
	devicePostureHistoryMaxLimit     = 100
	devicePostureObservedAtClockSkew = 5 * time.Minute
)

func (p DevicePosture) CursorKey(orderBy DevicePostureOrderField) page.CursorKey {
	switch orderBy {
	case DevicePostureOrderFieldCheckKey:
		return page.NewCursorKey(p.ID, p.CheckKey)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (p DevicePosture) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	evidence := p.Evidence
	if len(evidence) == 0 {
		evidence = emptyJSONObject
	}

	now := p.CreatedAt
	if now.IsZero() {
		now = time.Now()
	}

	observedAt := normalizeObservedAt(p.ObservedAt, now)

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
		"observed_at":     observedAt,
		"created_at":      p.CreatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert device posture: %w", err)
	}

	return nil
}

// normalizeObservedAt bounds client-supplied observation times before
// persistence. Zero timestamps default to now; values beyond the clock-skew
// allowance are clamped to now so a far-future observed_at cannot remain the
// latest result indefinitely.
func normalizeObservedAt(observed, now time.Time) time.Time {
	if observed.IsZero() {
		return now
	}

	if observed.After(now.Add(devicePostureObservedAtClockSkew)) {
		return now
	}

	return observed
}

// LoadLatestByDeviceID loads a page of the latest posture row for each
// check_key on the given device.
func (p *DevicePostures) LoadLatestByDeviceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	deviceID gid.GID,
	cursor *page.Cursor[DevicePostureOrderField],
) error {
	q := `
WITH latest AS (
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
)
SELECT * FROM latest WHERE %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"device_id": deviceID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

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
	if limit <= 0 || limit > devicePostureHistoryMaxLimit {
		limit = devicePostureHistoryMaxLimit
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

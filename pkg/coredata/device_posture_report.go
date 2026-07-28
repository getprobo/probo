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
	"fmt"
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/page"
)

type (
	DevicePostureReport struct {
		ID        gid.GID        `db:"id"`
		DeviceID  gid.GID        `db:"device_id"`
		CreatedAt time.Time      `db:"created_at"`
		Postures  DevicePostures `db:"-"`
	}

	DevicePostureReports []*DevicePostureReport
)

func (s DevicePostureReport) CursorKey(
	orderBy DevicePostureReportOrderField,
) page.CursorKey {
	switch orderBy {
	case DevicePostureReportOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

func (s *DevicePostureReports) LoadByDeviceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	deviceID gid.GID,
	cursor *page.Cursor[DevicePostureReportOrderField],
) error {
	q := `
WITH reports AS (
    SELECT
        correlation_id AS id,
        device_id,
        MIN(created_at) AS created_at
    FROM
        device_postures
    WHERE
        %s
        AND device_id = @device_id
    GROUP BY
        device_id,
        correlation_id
)
SELECT
    id,
    device_id,
    created_at
FROM
    reports
WHERE %s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"device_id": deviceID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query device posture reports: %w", err)
	}

	reports, err := pgx.CollectRows(
		rows,
		pgx.RowToAddrOfStructByName[DevicePostureReport],
	)
	if err != nil {
		return fmt.Errorf("cannot collect device posture reports: %w", err)
	}

	*s = reports

	return nil
}

func (s *DevicePostureReports) CountByDeviceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	deviceID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(*)
FROM (
    SELECT
        correlation_id
    FROM
        device_postures
    WHERE
        %s
        AND device_id = @device_id
    GROUP BY
        correlation_id
) AS reports
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"device_id": deviceID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count device posture reports: %w", err)
	}

	return count, nil
}

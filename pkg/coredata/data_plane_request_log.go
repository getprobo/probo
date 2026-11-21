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
)

type (
	DataPlaneRequestLog struct {
		ID              gid.GID   `db:"id"`
		OrganizationID  gid.GID   `db:"organization_id"`
		SigningConfigID *gid.GID  `db:"signing_config_id"`
		RequestPath     string    `db:"request_path"`
		Method          string    `db:"method"`
		StatusCode      int       `db:"status_code"`
		ErrorCode       string    `db:"error_code"`
		RequestID       string    `db:"request_id"`
		OccurredAt      time.Time `db:"occurred_at"`
	}

	DataPlaneRequestLogs []*DataPlaneRequestLog
)

const dataPlaneRequestLogMaxPerConfig = 100

func (l *DataPlaneRequestLog) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO data_plane_request_log (
	id, tenant_id, organization_id, signing_config_id,
	request_path, method, status_code, error_code, request_id, occurred_at
) VALUES (
	@id, @tenant_id, @organization_id, @signing_config_id,
	@request_path, @method, @status_code, @error_code, @request_id, @occurred_at
)
`
	args := pgx.StrictNamedArgs{
		"id":                l.ID,
		"tenant_id":         scope.GetTenantID(),
		"organization_id":   l.OrganizationID,
		"signing_config_id": l.SigningConfigID,
		"request_path":      l.RequestPath,
		"method":            l.Method,
		"status_code":       l.StatusCode,
		"error_code":        l.ErrorCode,
		"request_id":        l.RequestID,
		"occurred_at":       l.OccurredAt,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert data plane request log: %w", err)
	}

	return nil
}

func (ls *DataPlaneRequestLogs) LoadBySigningConfigID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	signingConfigID gid.GID,
	limit int,
) error {
	q := `
SELECT id, organization_id, signing_config_id, request_path, method, status_code, error_code, request_id, occurred_at
FROM data_plane_request_log
WHERE %s AND signing_config_id = @signing_config_id
ORDER BY occurred_at DESC
LIMIT @limit
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"signing_config_id": signingConfigID,
		"limit":             limit,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query data plane request logs: %w", err)
	}

	logs, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[DataPlaneRequestLog])
	if err != nil {
		return fmt.Errorf("cannot collect data plane request logs: %w", err)
	}

	*ls = logs

	return nil
}

func PruneDataPlaneRequestLog(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	signingConfigID gid.GID,
) error {
	q := `
DELETE FROM data_plane_request_log
WHERE %s
	AND signing_config_id = @signing_config_id
	AND id NOT IN (
		SELECT id FROM data_plane_request_log
		WHERE signing_config_id = @signing_config_id
		ORDER BY occurred_at DESC
		LIMIT @max_per_config
	)
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"signing_config_id": signingConfigID,
		"max_per_config":    dataPlaneRequestLogMaxPerConfig,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot prune data plane request log: %w", err)
	}

	return nil
}

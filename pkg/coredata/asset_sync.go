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
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	AssetSync struct {
		ID             gid.GID         `db:"id"`
		TenantID       gid.TenantID    `db:"tenant_id"`
		OrganizationID gid.GID         `db:"organization_id"`
		AssetSourceID  gid.GID         `db:"assets_source_id"`
		Status         AssetSyncStatus `db:"status"`
		Error          *string         `db:"error"`
		CompletedKinds []string        `db:"completed_kinds"`
		AttemptNumber  int             `db:"attempt_number"`
		StartedAt      time.Time       `db:"started_at"`
		CompletedAt    *time.Time      `db:"completed_at"`
		CreatedAt      time.Time       `db:"created_at"`
		UpdatedAt      time.Time       `db:"updated_at"`
	}
)

func (y *AssetSync) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	if y.CompletedKinds == nil {
		y.CompletedKinds = []string{}
	}

	q := `
INSERT INTO assets_syncs (
    id,
    tenant_id,
    organization_id,
    assets_source_id,
    status,
    error,
    completed_kinds,
    attempt_number,
    started_at,
    completed_at,
    created_at,
    updated_at
) VALUES (
    @id,
    @tenant_id,
    @organization_id,
    @assets_source_id,
    @status,
    @error,
    @completed_kinds::text[],
    @attempt_number,
    @started_at,
    @completed_at,
    @created_at,
    @updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":               y.ID,
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  y.OrganizationID,
		"assets_source_id": y.AssetSourceID,
		"status":           y.Status,
		"error":            y.Error,
		"completed_kinds":  y.CompletedKinds,
		"attempt_number":   y.AttemptNumber,
		"started_at":       y.StartedAt,
		"completed_at":     y.CompletedAt,
		"created_at":       y.CreatedAt,
		"updated_at":       y.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert asset sync: %w", err)
	}

	return nil
}

func (y *AssetSync) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	id gid.GID,
) error {
	row, err := queryAssetSync(
		ctx,
		conn,
		scope,
		"AND id = @id",
		pgx.StrictNamedArgs{"id": id},
		"cannot query asset sync",
		"cannot collect asset sync",
	)
	if err != nil {
		return err
	}

	*y = row

	return nil
}

func (y *AssetSync) LoadLatestBySourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sourceID gid.GID,
) error {
	row, err := queryAssetSync(
		ctx,
		conn,
		scope,
		"AND assets_source_id = @assets_source_id ORDER BY created_at DESC",
		pgx.StrictNamedArgs{"assets_source_id": sourceID},
		"cannot query latest asset sync",
		"cannot collect latest asset sync",
	)
	if err != nil {
		return err
	}

	*y = row

	return nil
}

// FinishIfRunning records the outcome when the sync is still running.
// A stale sweep may have failed it already; false leaves that row alone.
func (y *AssetSync) FinishIfRunning(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
	status AssetSyncStatus,
	syncErr *string,
	completedKinds []string,
	completedAt time.Time,
) (bool, error) {
	if completedKinds == nil {
		completedKinds = []string{}
	}

	q := `
UPDATE assets_syncs
SET
    status = @status,
    error = @error,
    completed_kinds = @completed_kinds::text[],
    completed_at = @completed_at,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
    AND status = 'RUNNING'
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":              y.ID,
		"status":          status,
		"error":           syncErr,
		"completed_kinds": completedKinds,
		"completed_at":    completedAt,
		"updated_at":      completedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	tag, err := conn.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot finish asset sync: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return false, nil
	}

	y.Status = status
	y.Error = syncErr
	y.CompletedKinds = completedKinds
	y.CompletedAt = &completedAt
	y.UpdatedAt = completedAt

	return true, nil
}

// FailStaleAssetSyncs fails running syncs that started at or before
// staleBefore. The caller schedules each source.
func FailStaleAssetSyncs(
	ctx context.Context,
	conn pg.Tx,
	staleBefore time.Time,
	message string,
) ([]AssetSync, error) {
	now := time.Now()

	q := `
UPDATE assets_syncs
SET
    status = 'FAILED',
    error = @error,
    completed_at = @completed_at,
    updated_at = @updated_at
WHERE
    status = 'RUNNING'
    AND started_at <= @stale_before
RETURNING
    id,
    tenant_id,
    organization_id,
    assets_source_id,
    status,
    error,
    completed_kinds,
    attempt_number,
    started_at,
    completed_at,
    created_at,
    updated_at
`

	args := pgx.StrictNamedArgs{
		"error":        message,
		"completed_at": now,
		"updated_at":   now,
		"stale_before": staleBefore,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query stale asset syncs: %w", err)
	}

	syncs, err := pgx.CollectRows(rows, pgx.RowToStructByName[AssetSync])
	if err != nil {
		return nil, fmt.Errorf("cannot collect stale asset syncs: %w", err)
	}

	return syncs, nil
}

func queryAssetSync(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	where string,
	args pgx.StrictNamedArgs,
	queryErr string,
	collectErr string,
) (AssetSync, error) {
	q := `
SELECT
    id,
    tenant_id,
    organization_id,
    assets_source_id,
    status,
    error,
    completed_kinds,
    attempt_number,
    started_at,
    completed_at,
    created_at,
    updated_at
FROM
    assets_syncs
WHERE
    %s
    %s
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment(), where)
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return AssetSync{}, fmt.Errorf("%s: %w", queryErr, err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AssetSync])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AssetSync{}, ErrResourceNotFound
		}

		return AssetSync{}, fmt.Errorf("%s: %w", collectErr, err)
	}

	return row, nil
}

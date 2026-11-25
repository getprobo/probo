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
)

// LoadNextDueForInventorySyncSkipLocked claims the next live source whose
// cadence window is open and that has no QUEUED/RUNNING sync run. Cross-tenant
// so the scheduler can sweep every org. NULL next_sync_at means due now.
func (as *AccessReviewSource) LoadNextDueForInventorySyncSkipLocked(
	ctx context.Context,
	conn pg.Tx,
	now time.Time,
) error {
	q := `
SELECT
    id,
    organization_id,
    connector_account_id,
    name,
    csv_data,
    last_successful_sync_at,
    next_sync_at,
    sync_consecutive_failures,
    created_at,
    updated_at
FROM access_review_sources source
WHERE
    (source.next_sync_at IS NULL OR source.next_sync_at <= @now)
    AND NOT EXISTS (
        SELECT 1
        FROM access_source_sync_runs run
        WHERE
            run.access_review_source_id = source.id
            AND run.status IN (@queued_status, @running_status)
    )
ORDER BY source.next_sync_at ASC NULLS FIRST, source.id ASC
LIMIT 1
FOR UPDATE OF source SKIP LOCKED
`
	args := pgx.StrictNamedArgs{
		"now":            now,
		"queued_status":  AccessSourceSyncRunStatusQueued,
		"running_status": AccessSourceSyncRunStatusRunning,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query due access review sources: %w", err)
	}

	source, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessReviewSource])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoAccessReviewSourceInventorySyncDue
		}

		return fmt.Errorf("cannot collect due access review source: %w", err)
	}

	*as = source

	return nil
}

// UpdateInventorySyncSchedule writes freshness and next-run fields without
// touching name / CSV / connector linkage.
func (as *AccessReviewSource) UpdateInventorySyncSchedule(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE access_review_sources
SET
    last_successful_sync_at = @last_successful_sync_at,
    next_sync_at = @next_sync_at,
    sync_consecutive_failures = @sync_consecutive_failures,
    updated_at = @updated_at
WHERE
    %s
    AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                        as.ID,
		"last_successful_sync_at":   as.LastSuccessfulSyncAt,
		"next_sync_at":              as.NextSyncAt,
		"sync_consecutive_failures": as.SyncConsecutiveFailures,
		"updated_at":                as.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update access review source sync schedule: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

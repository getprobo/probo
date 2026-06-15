// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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
	"go.probo.inc/probo/pkg/page"
)

type (
	// AccessReviewCampaignSourceFetchAttempt is a single, append-only fetch run for a
	// campaign source snapshot. Each retry produces a new row, so the error of
	// every attempt is retained. The current state of a snapshot is the latest
	// attempt (most recently created). Terminal rows (SUCCESS / FAILED) are
	// immutable; only the in-flight attempt is updated.
	//
	// TenantID is retained on the struct because the background worker claims
	// rows cross-tenant via LoadNextQueuedForUpdateSkipLocked and needs the
	// tenant to construct a Scope for subsequent operations.
	AccessReviewCampaignSourceFetchAttempt struct {
		ID                           gid.GID                               `db:"id"`
		TenantID                     gid.TenantID                          `db:"tenant_id"`
		AccessReviewCampaignSourceID gid.GID                               `db:"access_review_campaign_source_id"`
		Status                       AccessReviewCampaignSourceFetchStatus `db:"status"`
		FetchedAccountsCount         int                                   `db:"fetched_accounts_count"`
		Error                        *string                               `db:"error"`
		StartedAt                    *time.Time                            `db:"started_at"`
		CompletedAt                  *time.Time                            `db:"completed_at"`
		CreatedAt                    time.Time                             `db:"created_at"`
		UpdatedAt                    time.Time                             `db:"updated_at"`
		AttemptNumber                int                                   `db:"attempt_number"`
	}

	AccessReviewCampaignSourceFetchAttempts []*AccessReviewCampaignSourceFetchAttempt
)

func (a AccessReviewCampaignSourceFetchAttempt) CursorKey(
	orderBy AccessReviewCampaignSourceFetchAttemptOrderField,
) page.CursorKey {
	switch orderBy {
	case AccessReviewCampaignSourceFetchAttemptOrderFieldCreatedAt:
		return page.NewCursorKey(a.ID, a.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", orderBy))
}

var (
	ErrNoAccessReviewCampaignSourceFetchAttemptAvailable = errors.New("no access review source fetch attempt available")
)

// Insert appends a new attempt for the snapshot, assigning the next
// attempt_number atomically. The receiver's AttemptNumber is synced from the
// database.
func (a *AccessReviewCampaignSourceFetchAttempt) Insert(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO access_review_campaign_source_fetch_attempts (
	id,
	tenant_id,
	access_review_campaign_source_id,
	attempt_number,
	status,
	fetched_accounts_count,
	error,
	started_at,
	completed_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@access_review_campaign_source_id,
	COALESCE((
		SELECT MAX(attempt_number)
		FROM access_review_campaign_source_fetch_attempts
		WHERE access_review_campaign_source_id = @access_review_campaign_source_id
	), 0) + 1,
	@status,
	@fetched_accounts_count,
	@error,
	@started_at,
	@completed_at,
	@created_at,
	@updated_at
)
RETURNING attempt_number
`
	args := pgx.StrictNamedArgs{
		"id":                               a.ID,
		"tenant_id":                        scope.GetTenantID(),
		"access_review_campaign_source_id": a.AccessReviewCampaignSourceID,
		"status":                           a.Status,
		"fetched_accounts_count":           a.FetchedAccountsCount,
		"error":                            a.Error,
		"started_at":                       a.StartedAt,
		"completed_at":                     a.CompletedAt,
		"created_at":                       a.CreatedAt,
		"updated_at":                       a.UpdatedAt,
	}

	if err := conn.QueryRow(ctx, q, args).Scan(&a.AttemptNumber); err != nil {
		return fmt.Errorf("cannot insert source fetch attempt: %w", err)
	}

	return nil
}

// Update writes the in-flight attempt's lifecycle fields. It must only be
// called on the attempt that the worker currently owns.
func (a *AccessReviewCampaignSourceFetchAttempt) Update(
	ctx context.Context,
	conn pg.Tx,
	scope Scoper,
) error {
	q := `
UPDATE access_review_campaign_source_fetch_attempts
SET
	status = @status,
	fetched_accounts_count = @fetched_accounts_count,
	error = @error,
	started_at = @started_at,
	completed_at = @completed_at,
	updated_at = @updated_at
WHERE
	%s
	AND id = @id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":                     a.ID,
		"status":                 a.Status,
		"fetched_accounts_count": a.FetchedAccountsCount,
		"error":                  a.Error,
		"started_at":             a.StartedAt,
		"completed_at":           a.CompletedAt,
		"updated_at":             a.UpdatedAt,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update source fetch attempt: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

// LoadNextQueuedForUpdateSkipLocked is intentionally cross-tenant: the
// background worker claims the next available attempt regardless of tenant.
// The caller extracts TenantID from the returned struct to construct a Scope
// for subsequent operations.
func (a *AccessReviewCampaignSourceFetchAttempt) LoadNextQueuedForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Tx,
) error {
	q := `
SELECT
	id,
	tenant_id,
	access_review_campaign_source_id,
	attempt_number,
	status,
	fetched_accounts_count,
	error,
	started_at,
	completed_at,
	created_at,
	updated_at
FROM access_review_campaign_source_fetch_attempts
WHERE status = @status
ORDER BY created_at ASC, id ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`
	args := pgx.StrictNamedArgs{
		"status": AccessReviewCampaignSourceFetchStatusQueued,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query next queued fetch attempt: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessReviewCampaignSourceFetchAttempt])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoAccessReviewCampaignSourceFetchAttemptAvailable
		}

		return fmt.Errorf("cannot collect fetch attempt: %w", err)
	}

	*a = result

	return nil
}

// LoadLatestByCampaignID returns the most recent attempt for every snapshot in
// the campaign, keyed by snapshot ID. Snapshots without any attempt are absent.
func (attempts *AccessReviewCampaignSourceFetchAttempts) LoadLatestByCampaignID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	campaignID gid.GID,
) error {
	q := `
SELECT DISTINCT ON (access_review_campaign_source_id)
	id,
	tenant_id,
	access_review_campaign_source_id,
	attempt_number,
	status,
	fetched_accounts_count,
	error,
	started_at,
	completed_at,
	created_at,
	updated_at
FROM access_review_campaign_source_fetch_attempts
WHERE
	%s
	AND access_review_campaign_source_id IN (
		SELECT id
		FROM access_review_campaign_sources
		WHERE access_review_campaign_id = @campaign_id
	)
ORDER BY access_review_campaign_source_id, attempt_number DESC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"campaign_id": campaignID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query latest fetch attempts: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessReviewCampaignSourceFetchAttempt])
	if err != nil {
		return fmt.Errorf("cannot collect latest fetch attempts: %w", err)
	}

	*attempts = result

	return nil
}

// LoadByCampaignSourceID returns a page of fetch attempts for a snapshot.
func (attempts *AccessReviewCampaignSourceFetchAttempts) LoadByCampaignSourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	campaignSourceID gid.GID,
	cursor *page.Cursor[AccessReviewCampaignSourceFetchAttemptOrderField],
) error {
	q := `
SELECT
	id,
	tenant_id,
	access_review_campaign_source_id,
	status,
	fetched_accounts_count,
	error,
	started_at,
	completed_at,
	created_at,
	updated_at,
	attempt_number
FROM (
	SELECT
		id,
		tenant_id,
		access_review_campaign_source_id,
		status,
		fetched_accounts_count,
		error,
		started_at,
		completed_at,
		created_at,
		updated_at,
		ROW_NUMBER() OVER (ORDER BY created_at ASC, id ASC)::int AS attempt_number
	FROM access_review_campaign_source_fetch_attempts
	WHERE
		%s
		AND access_review_campaign_source_id = @access_review_campaign_source_id
) fetch_attempts
WHERE
	%s
`
	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"access_review_campaign_source_id": campaignSourceID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query fetch attempts: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessReviewCampaignSourceFetchAttempt])
	if err != nil {
		return fmt.Errorf("cannot collect fetch attempts: %w", err)
	}

	*attempts = result

	return nil
}

// LoadAllByCampaignSourceID returns the full attempt history for a snapshot,
// newest first.
func (attempts *AccessReviewCampaignSourceFetchAttempts) LoadAllByCampaignSourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	campaignSourceID gid.GID,
) error {
	q := `
SELECT
	id,
	tenant_id,
	access_review_campaign_source_id,
	attempt_number,
	status,
	fetched_accounts_count,
	error,
	started_at,
	completed_at,
	created_at,
	updated_at
FROM access_review_campaign_source_fetch_attempts
WHERE
	%s
	AND access_review_campaign_source_id = @access_review_campaign_source_id
ORDER BY attempt_number DESC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"access_review_campaign_source_id": campaignSourceID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query fetch attempts: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessReviewCampaignSourceFetchAttempt])
	if err != nil {
		return fmt.Errorf("cannot collect fetch attempts: %w", err)
	}

	*attempts = result

	return nil
}

func (attempts *AccessReviewCampaignSourceFetchAttempts) CountByCampaignSourceID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	campaignSourceID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(*)
FROM access_review_campaign_source_fetch_attempts
WHERE
	%s
	AND access_review_campaign_source_id = @access_review_campaign_source_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"access_review_campaign_source_id": campaignSourceID}
	maps.Copy(args, scope.SQLArguments())

	var count int

	if err := conn.QueryRow(ctx, q, args).Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count fetch attempts: %w", err)
	}

	return count, nil
}

// RecoverStale fails attempts stuck in FETCHING past the threshold and queues a
// fresh retry attempt for each, preserving the stale attempt's history. It is
// intentionally cross-tenant. Returns the number of recovered attempts.
func (attempts *AccessReviewCampaignSourceFetchAttempts) RecoverStale(
	ctx context.Context,
	conn pg.Tx,
	staleThreshold time.Time,
	now time.Time,
) (int, error) {
	q := `
SELECT
	id,
	tenant_id,
	access_review_campaign_source_id,
	attempt_number,
	status,
	fetched_accounts_count,
	error,
	started_at,
	completed_at,
	created_at,
	updated_at
FROM access_review_campaign_source_fetch_attempts
WHERE status = 'FETCHING'
	AND updated_at < @stale_threshold
FOR UPDATE SKIP LOCKED
`

	rows, err := conn.Query(ctx, q, pgx.StrictNamedArgs{"stale_threshold": staleThreshold})
	if err != nil {
		return 0, fmt.Errorf("cannot query stale fetch attempts: %w", err)
	}

	stale, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessReviewCampaignSourceFetchAttempt])
	if err != nil {
		return 0, fmt.Errorf("cannot collect stale fetch attempts: %w", err)
	}

	staleMessage := "fetch timed out"

	for _, attempt := range stale {
		scope := NewScope(attempt.TenantID)

		attempt.Status = AccessReviewCampaignSourceFetchStatusFailed
		attempt.Error = &staleMessage
		attempt.CompletedAt = &now
		attempt.UpdatedAt = now

		if err := attempt.Update(ctx, conn, scope); err != nil {
			return 0, fmt.Errorf("cannot fail stale fetch attempt: %w", err)
		}

		retry := &AccessReviewCampaignSourceFetchAttempt{
			ID:                           gid.New(attempt.TenantID, AccessReviewCampaignSourceFetchAttemptEntityType),
			AccessReviewCampaignSourceID: attempt.AccessReviewCampaignSourceID,
			Status:                       AccessReviewCampaignSourceFetchStatusQueued,
			CreatedAt:                    now,
			UpdatedAt:                    now,
		}

		if err := retry.Insert(ctx, conn, scope); err != nil {
			return 0, fmt.Errorf("cannot queue retry fetch attempt: %w", err)
		}
	}

	return len(stale), nil
}

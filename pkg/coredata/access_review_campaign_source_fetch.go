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
	"maps"
	"time"

	"github.com/jackc/pgx/v5"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
)

type (
	// AccessReviewCampaignSourceFetch tracks per-source fetch lifecycle.
	// TenantID is retained on the struct because the background worker claims
	// rows cross-tenant via LoadNextQueuedForUpdateSkipLocked and needs the
	// tenant to construct a Scope for subsequent operations.
	AccessReviewCampaignSourceFetch struct {
		TenantID               gid.TenantID                          `db:"tenant_id"`
		AccessReviewCampaignID gid.GID                               `db:"access_review_campaign_id"`
		AccessSourceID         gid.GID                               `db:"access_source_id"`
		Status                 AccessReviewCampaignSourceFetchStatus `db:"status"`
		FetchedAccountsCount   int                                   `db:"fetched_accounts_count"`
		AttemptCount           int                                   `db:"attempt_count"`
		LastError              *string                               `db:"last_error"`
		StartedAt              *time.Time                            `db:"started_at"`
		CompletedAt            *time.Time                            `db:"completed_at"`
		CreatedAt              time.Time                             `db:"created_at"`
		UpdatedAt              time.Time                             `db:"updated_at"`
	}

	AccessReviewCampaignSourceFetches []*AccessReviewCampaignSourceFetch
)

var (
	ErrNoAccessReviewCampaignSourceFetchAvailable = errors.New("no access review campaign source fetch available")
)

func (f *AccessReviewCampaignSourceFetch) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO access_review_campaign_source_fetches (
	tenant_id,
	access_review_campaign_id,
	access_source_id,
	status,
	fetched_accounts_count,
	attempt_count,
	last_error,
	started_at,
	completed_at,
	created_at,
	updated_at
) VALUES (
	@tenant_id,
	@access_review_campaign_id,
	@access_source_id,
	@status,
	@fetched_accounts_count,
	@attempt_count,
	@last_error,
	@started_at,
	@completed_at,
	@created_at,
	@updated_at
)
`
	args := pgx.StrictNamedArgs{
		"tenant_id":                 scope.GetTenantID(),
		"access_review_campaign_id": f.AccessReviewCampaignID,
		"access_source_id":          f.AccessSourceID,
		"status":                    f.Status,
		"fetched_accounts_count":    f.FetchedAccountsCount,
		"attempt_count":             f.AttemptCount,
		"last_error":                f.LastError,
		"started_at":                f.StartedAt,
		"completed_at":              f.CompletedAt,
		"created_at":                f.CreatedAt,
		"updated_at":                f.UpdatedAt,
	}
	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert campaign source fetch: %w", err)
	}

	return nil
}

func (f *AccessReviewCampaignSourceFetch) Update(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
UPDATE access_review_campaign_source_fetches
SET
	status = @status,
	fetched_accounts_count = @fetched_accounts_count,
	attempt_count = @attempt_count,
	last_error = @last_error,
	started_at = @started_at,
	completed_at = @completed_at,
	updated_at = @updated_at
WHERE
	%s
	AND access_review_campaign_id = @access_review_campaign_id
	AND access_source_id = @access_source_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"status":                    f.Status,
		"fetched_accounts_count":    f.FetchedAccountsCount,
		"attempt_count":             f.AttemptCount,
		"last_error":                f.LastError,
		"started_at":                f.StartedAt,
		"completed_at":              f.CompletedAt,
		"updated_at":                f.UpdatedAt,
		"access_review_campaign_id": f.AccessReviewCampaignID,
		"access_source_id":          f.AccessSourceID,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot update campaign source fetch: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	return nil
}

func (f *AccessReviewCampaignSourceFetch) LoadByID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	campaignID gid.GID,
	sourceID gid.GID,
) error {
	q := `
SELECT
	tenant_id,
	access_review_campaign_id,
	access_source_id,
	status,
	fetched_accounts_count,
	attempt_count,
	last_error,
	started_at,
	completed_at,
	created_at,
	updated_at
FROM access_review_campaign_source_fetches
WHERE
	%s
	AND access_review_campaign_id = @access_review_campaign_id
	AND access_source_id = @access_source_id
LIMIT 1
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"access_review_campaign_id": campaignID,
		"access_source_id":          sourceID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query campaign source fetch: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessReviewCampaignSourceFetch])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot collect campaign source fetch: %w", err)
	}

	*f = result

	return nil
}

func (fs *AccessReviewCampaignSourceFetches) LoadByCampaignID(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	campaignID gid.GID,
) error {
	q := `
SELECT
	tenant_id,
	access_review_campaign_id,
	access_source_id,
	status,
	fetched_accounts_count,
	attempt_count,
	last_error,
	started_at,
	completed_at,
	created_at,
	updated_at
FROM access_review_campaign_source_fetches
WHERE
	%s
	AND access_review_campaign_id = @access_review_campaign_id
ORDER BY created_at ASC
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"access_review_campaign_id": campaignID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query campaign source fetches: %w", err)
	}

	result, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[AccessReviewCampaignSourceFetch])
	if err != nil {
		return fmt.Errorf("cannot collect campaign source fetches: %w", err)
	}

	*fs = result

	return nil
}

// LoadNextQueuedForUpdateSkipLocked is intentionally cross-tenant: the
// background worker claims the next available fetch regardless of tenant.
// The caller extracts TenantID from the returned struct to construct a
// Scope for subsequent operations.
func (f *AccessReviewCampaignSourceFetch) LoadNextQueuedForUpdateSkipLocked(
	ctx context.Context,
	conn pg.Conn,
) error {
	q := `
SELECT
	tenant_id,
	access_review_campaign_id,
	access_source_id,
	status,
	fetched_accounts_count,
	attempt_count,
	last_error,
	started_at,
	completed_at,
	created_at,
	updated_at
FROM access_review_campaign_source_fetches
WHERE status = @status
ORDER BY created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED
`
	args := pgx.StrictNamedArgs{
		"status": AccessReviewCampaignSourceFetchStatusQueued,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query next queued campaign source fetch: %w", err)
	}

	result, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[AccessReviewCampaignSourceFetch])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoAccessReviewCampaignSourceFetchAvailable
		}
		return fmt.Errorf("cannot collect campaign source fetch: %w", err)
	}

	*f = result

	return nil
}

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

type AccessReviewCampaignScopeSystem struct {
	AccessReviewCampaignID gid.GID `db:"access_review_campaign_id"`
	AccessSourceID         gid.GID `db:"access_source_id"`
}

func (ss AccessReviewCampaignScopeSystem) Insert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO access_review_campaign_scope_systems (access_review_campaign_id, access_source_id, tenant_id)
VALUES (@access_review_campaign_id, @access_source_id, @tenant_id)
`
	args := pgx.StrictNamedArgs{
		"access_review_campaign_id": ss.AccessReviewCampaignID,
		"access_source_id":          ss.AccessSourceID,
		"tenant_id":                 scope.GetTenantID(),
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot insert campaign scope system: %w", err)
	}

	return nil
}

func (ss AccessReviewCampaignScopeSystem) Upsert(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
INSERT INTO access_review_campaign_scope_systems (access_review_campaign_id, access_source_id, tenant_id)
VALUES (@access_review_campaign_id, @access_source_id, @tenant_id)
ON CONFLICT (access_review_campaign_id, access_source_id) DO NOTHING
`
	args := pgx.StrictNamedArgs{
		"access_review_campaign_id": ss.AccessReviewCampaignID,
		"access_source_id":          ss.AccessSourceID,
		"tenant_id":                 scope.GetTenantID(),
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert campaign scope system: %w", err)
	}

	return nil
}

func (ss AccessReviewCampaignScopeSystem) Delete(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
DELETE FROM access_review_campaign_scope_systems
WHERE
    %s
    AND access_review_campaign_id = @access_review_campaign_id
    AND access_source_id = @access_source_id
`
	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"access_review_campaign_id": ss.AccessReviewCampaignID,
		"access_source_id":          ss.AccessSourceID,
	}
	maps.Copy(args, scope.SQLArguments())

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot delete campaign scope system: %w", err)
	}

	return nil
}

func (c *AccessReviewCampaign) LockForUpdate(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
) error {
	q := `
SELECT id
FROM access_review_campaigns
WHERE %s
  AND id = @id
FOR UPDATE
`
	q = fmt.Sprintf(q, scope.SQLFragment())
	args := pgx.StrictNamedArgs{"id": c.ID}
	maps.Copy(args, scope.SQLArguments())

	var id gid.GID
	if err := conn.QueryRow(ctx, q, args).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}
		return fmt.Errorf("cannot lock campaign: %w", err)
	}

	return nil
}

func (f *AccessReviewCampaignSourceFetch) UpsertQueued(
	ctx context.Context,
	conn pg.Conn,
	scope Scoper,
	now time.Time,
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
	@tenant_id, @access_review_campaign_id, @access_source_id,
	'QUEUED', 0, 0, NULL, NULL, NULL, @now, @now
)
ON CONFLICT (access_review_campaign_id, access_source_id) DO UPDATE SET
	status = 'QUEUED',
	fetched_accounts_count = 0,
	attempt_count = 0,
	last_error = NULL,
	started_at = NULL,
	completed_at = NULL,
	updated_at = EXCLUDED.updated_at
`
	args := pgx.StrictNamedArgs{
		"tenant_id":                 scope.GetTenantID(),
		"access_review_campaign_id": f.AccessReviewCampaignID,
		"access_source_id":          f.AccessSourceID,
		"now":                       now,
	}

	_, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot upsert queued source fetch: %w", err)
	}

	return nil
}

// RecoverStale is intentionally cross-tenant: the background worker recovers
// all stale fetches regardless of tenant.
func (fs *AccessReviewCampaignSourceFetches) RecoverStale(
	ctx context.Context,
	conn pg.Conn,
	staleThreshold time.Time,
	now time.Time,
) (int64, error) {
	q := `
UPDATE access_review_campaign_source_fetches
SET
	status = 'QUEUED',
	last_error = 'recovered from stale FETCHING state',
	started_at = NULL,
	completed_at = NULL,
	updated_at = @now
WHERE
	status = 'FETCHING'
	AND updated_at < @stale_threshold
`
	args := pgx.StrictNamedArgs{
		"now":             now,
		"stale_threshold": staleThreshold,
	}

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return 0, fmt.Errorf("cannot recover stale source fetches: %w", err)
	}

	return result.RowsAffected(), nil
}

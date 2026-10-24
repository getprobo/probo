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
	"github.com/jackc/pgx/v5/pgconn"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	CookieBannerScan struct {
		ID             gid.GID                        `db:"id"`
		OrganizationID gid.GID                        `db:"organization_id"`
		CookieBannerID gid.GID                        `db:"cookie_banner_id"`
		State          CookieBannerScanState          `db:"state"`
		FailureReason  *CookieBannerScanFailureReason `db:"failure_reason"`
		Truncated      bool                           `db:"truncated"`
		PagesScanned   int                            `db:"pages_scanned"`
		CookiesFound   int                            `db:"cookies_found"`
		StorageFound   int                            `db:"storage_found"`
		ResourcesFound int                            `db:"resources_found"`
		StartedAt      *time.Time                     `db:"started_at"`
		CompletedAt    *time.Time                     `db:"completed_at"`
		CreatedAt      time.Time                      `db:"created_at"`
		UpdatedAt      time.Time                      `db:"updated_at"`
	}

	CookieBannerScans []*CookieBannerScan
)

func (s *CookieBannerScan) CursorKey(field CookieBannerScanOrderField) page.CursorKey {
	switch field {
	case CookieBannerScanOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (s *CookieBannerScan) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM cookie_banner_scans WHERE id = ANY(@resource_ids::text[])`

	args := pgx.StrictNamedArgs{
		"resource_ids": resourceIDs,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query authorization attributes: %w", err)
	}

	defer rows.Close()

	attrsByID := make(policy.AttributesByID)

	for rows.Next() {
		var id, organizationID gid.GID

		if err := rows.Scan(&id, &organizationID); err != nil {
			return nil, fmt.Errorf("cannot scan authorization attributes: %w", err)
		}

		attrsByID[id] = policy.Attributes{
			"organization_id": organizationID.String(),
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate authorization attributes: %w", err)
	}

	return attrsByID, nil
}

func (s *CookieBannerScan) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	scanID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	state,
	failure_reason,
	truncated,
	pages_scanned,
	cookies_found,
	storage_found,
	resources_found,
	started_at,
	completed_at,
	created_at,
	updated_at
FROM cookie_banner_scans
WHERE
	%s
	AND id = @scan_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"scan_id": scanID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banner scan: %w", err)
	}

	scan, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBannerScan])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect cookie banner scan: %w", err)
	}

	*s = scan

	return nil
}

func (s *CookieBannerScan) LoadLatestByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	state,
	failure_reason,
	truncated,
	pages_scanned,
	cookies_found,
	storage_found,
	resources_found,
	started_at,
	completed_at,
	created_at,
	updated_at
FROM cookie_banner_scans
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
ORDER BY created_at DESC, id DESC
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query latest cookie banner scan: %w", err)
	}

	scan, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBannerScan])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect latest cookie banner scan: %w", err)
	}

	*s = scan

	return nil
}

func (s *CookieBannerScans) LoadByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	cursor *page.Cursor[CookieBannerScanOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	state,
	failure_reason,
	truncated,
	pages_scanned,
	cookies_found,
	storage_found,
	resources_found,
	started_at,
	completed_at,
	created_at,
	updated_at
FROM cookie_banner_scans
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banner scans: %w", err)
	}

	scans, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CookieBannerScan])
	if err != nil {
		return fmt.Errorf("cannot collect cookie banner scans: %w", err)
	}

	*s = scans

	return nil
}

func (s *CookieBannerScans) CountByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM cookie_banner_scans
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count cookie banner scans: %w", err)
	}

	return count, nil
}

// Insert returns ErrResourceAlreadyExists when the banner already has a
// PENDING or RUNNING scan.
func (s *CookieBannerScan) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO cookie_banner_scans (
	id,
	tenant_id,
	organization_id,
	cookie_banner_id,
	state,
	failure_reason,
	truncated,
	pages_scanned,
	cookies_found,
	storage_found,
	resources_found,
	started_at,
	completed_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@state,
	@failure_reason,
	@truncated,
	@pages_scanned,
	@cookies_found,
	@storage_found,
	@resources_found,
	@started_at,
	@completed_at,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":               s.ID,
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  s.OrganizationID,
		"cookie_banner_id": s.CookieBannerID,
		"state":            s.State,
		"failure_reason":   s.FailureReason,
		"truncated":        s.Truncated,
		"pages_scanned":    s.PagesScanned,
		"cookies_found":    s.CookiesFound,
		"storage_found":    s.StorageFound,
		"resources_found":  s.ResourcesFound,
		"started_at":       s.StartedAt,
		"completed_at":     s.CompletedAt,
		"created_at":       s.CreatedAt,
		"updated_at":       s.UpdatedAt,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "cookie_banner_scans_one_active_per_banner_idx" {
				return ErrResourceAlreadyExists
			}
		}

		return fmt.Errorf("cannot insert cookie banner scan: %w", err)
	}

	return nil
}

func (s *CookieBannerScan) LoadNextPendingForUpdateSkipLocked(
	ctx context.Context,
	tx pg.Tx,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	state,
	failure_reason,
	truncated,
	pages_scanned,
	cookies_found,
	storage_found,
	resources_found,
	started_at,
	completed_at,
	created_at,
	updated_at
FROM cookie_banner_scans
WHERE state = @pending_state
ORDER BY created_at ASC, id ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;
`

	args := pgx.StrictNamedArgs{"pending_state": CookieBannerScanStatePending}

	rows, err := tx.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query next pending cookie banner scan: %w", err)
	}

	scan, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBannerScan])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect next pending cookie banner scan: %w", err)
	}

	*s = scan

	return nil
}

// MarkRunning moves a PENDING scan to RUNNING. It returns
// ErrResourceNotFound when the scan is no longer PENDING.
func (s *CookieBannerScan) MarkRunning(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	now time.Time,
) error {
	q := `
UPDATE cookie_banner_scans
SET
	state = @running_state,
	started_at = @now,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND state = @pending_state
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":            s.ID,
		"running_state": CookieBannerScanStateRunning,
		"pending_state": CookieBannerScanStatePending,
		"now":           now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot mark cookie banner scan running: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	s.State = CookieBannerScanStateRunning
	s.StartedAt = &now
	s.UpdatedAt = now

	return nil
}

// MarkCompleted moves a RUNNING scan to COMPLETED with the receiver's
// counts and truncation flag. It returns ErrResourceNotFound when the
// scan is no longer RUNNING, e.g. after stale recovery failed it.
func (s *CookieBannerScan) MarkCompleted(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	now time.Time,
) error {
	q := `
UPDATE cookie_banner_scans
SET
	state = @completed_state,
	truncated = @truncated,
	pages_scanned = @pages_scanned,
	cookies_found = @cookies_found,
	storage_found = @storage_found,
	resources_found = @resources_found,
	completed_at = @now,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND state = @running_state
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":              s.ID,
		"completed_state": CookieBannerScanStateCompleted,
		"running_state":   CookieBannerScanStateRunning,
		"truncated":       s.Truncated,
		"pages_scanned":   s.PagesScanned,
		"cookies_found":   s.CookiesFound,
		"storage_found":   s.StorageFound,
		"resources_found": s.ResourcesFound,
		"now":             now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot mark cookie banner scan completed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	s.State = CookieBannerScanStateCompleted
	s.CompletedAt = &now
	s.UpdatedAt = now

	return nil
}

// MarkFailed moves a RUNNING scan to FAILED. It returns
// ErrResourceNotFound when the scan is no longer RUNNING.
func (s *CookieBannerScan) MarkFailed(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	reason CookieBannerScanFailureReason,
	now time.Time,
) error {
	q := `
UPDATE cookie_banner_scans
SET
	state = @failed_state,
	failure_reason = @failure_reason,
	pages_scanned = @pages_scanned,
	completed_at = @now,
	updated_at = @now
WHERE
	%s
	AND id = @id
	AND state = @running_state
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"id":             s.ID,
		"failed_state":   CookieBannerScanStateFailed,
		"running_state":  CookieBannerScanStateRunning,
		"failure_reason": reason,
		"pages_scanned":  s.PagesScanned,
		"now":            now,
	}
	maps.Copy(args, scope.SQLArguments())

	result, err := conn.Exec(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot mark cookie banner scan failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrResourceNotFound
	}

	s.State = CookieBannerScanStateFailed
	s.FailureReason = &reason
	s.CompletedAt = &now
	s.UpdatedAt = now

	return nil
}

// FailStaleCookieBannerScans fails RUNNING scans started before
// staleAfter, so a crashed worker does not block the banner's next scan.
func FailStaleCookieBannerScans(
	ctx context.Context,
	conn pg.Querier,
	staleAfter time.Duration,
) error {
	now := time.Now()

	q := `
UPDATE cookie_banner_scans
SET
	state = @failed_state,
	failure_reason = @failure_reason,
	completed_at = @now,
	updated_at = @now
WHERE
	state = @running_state
	AND started_at < @stale_threshold
`

	args := pgx.StrictNamedArgs{
		"failed_state":    CookieBannerScanStateFailed,
		"running_state":   CookieBannerScanStateRunning,
		"failure_reason":  CookieBannerScanFailureReasonInterrupted,
		"now":             now,
		"stale_threshold": now.Add(-staleAfter),
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot fail stale cookie banner scans: %w", err)
	}

	return nil
}

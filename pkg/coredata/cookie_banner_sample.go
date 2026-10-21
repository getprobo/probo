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
	"go.probo.inc/probo/pkg/iam/policy"
	"go.probo.inc/probo/pkg/page"
)

type (
	CookieBannerSample struct {
		ID             gid.GID    `db:"id"`
		OrganizationID gid.GID    `db:"organization_id"`
		CookieBannerID gid.GID    `db:"cookie_banner_id"`
		FrozenAt       *time.Time `db:"frozen_at"`
		CreatedAt      time.Time  `db:"created_at"`
		UpdatedAt      time.Time  `db:"updated_at"`
	}

	CookieBannerSamples []*CookieBannerSample
)

func (s *CookieBannerSample) CursorKey(field CookieBannerSampleOrderField) page.CursorKey {
	switch field {
	case CookieBannerSampleOrderFieldCreatedAt:
		return page.NewCursorKey(s.ID, s.CreatedAt)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (s *CookieBannerSample) AuthorizationAttributes(
	ctx context.Context,
	conn pg.Querier,
	resourceIDs []gid.GID,
) (policy.AttributesByID, error) {
	q := `SELECT id, organization_id FROM cookie_banner_samples WHERE id = ANY(@resource_ids::text[])`

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

func (s *CookieBannerSample) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sampleID gid.GID,
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	frozen_at,
	created_at,
	updated_at
FROM cookie_banner_samples
WHERE
	%s
	AND id = @sample_id
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"sample_id": sampleID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banner sample: %w", err)
	}

	sample, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBannerSample])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect cookie banner sample: %w", err)
	}

	*s = sample

	return nil
}

func (s *CookieBannerSample) LoadOpenByCookieBannerID(
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
	frozen_at,
	created_at,
	updated_at
FROM cookie_banner_samples
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND frozen_at IS NULL
LIMIT 1;
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"cookie_banner_id": cookieBannerID}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query open cookie banner sample: %w", err)
	}

	sample, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBannerSample])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect open cookie banner sample: %w", err)
	}

	*s = sample

	return nil
}

func (s *CookieBannerSample) LoadLatestByCookieBannerID(
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
	frozen_at,
	created_at,
	updated_at
FROM cookie_banner_samples
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
		return fmt.Errorf("cannot query latest cookie banner sample: %w", err)
	}

	sample, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CookieBannerSample])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect latest cookie banner sample: %w", err)
	}

	*s = sample

	return nil
}

func (s *CookieBannerSamples) LoadByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	cursor *page.Cursor[CookieBannerSampleOrderField],
) error {
	q := `
SELECT
	id,
	organization_id,
	cookie_banner_id,
	frozen_at,
	created_at,
	updated_at
FROM cookie_banner_samples
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
		return fmt.Errorf("cannot query cookie banner samples: %w", err)
	}

	samples, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CookieBannerSample])
	if err != nil {
		return fmt.Errorf("cannot collect cookie banner samples: %w", err)
	}

	*s = samples

	return nil
}

func (s *CookieBannerSamples) CountByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
) (int, error) {
	q := `
SELECT COUNT(id)
FROM cookie_banner_samples
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
		return 0, fmt.Errorf("cannot count cookie banner samples: %w", err)
	}

	return count, nil
}

func (s *CookieBannerSample) Insert(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) error {
	q := `
INSERT INTO cookie_banner_samples (
	id,
	tenant_id,
	organization_id,
	cookie_banner_id,
	frozen_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@frozen_at,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":               s.ID,
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  s.OrganizationID,
		"cookie_banner_id": s.CookieBannerID,
		"frozen_at":        s.FrozenAt,
		"created_at":       s.CreatedAt,
		"updated_at":       s.UpdatedAt,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert cookie banner sample: %w", err)
	}

	return nil
}

// InsertOpen inserts an unfrozen sample. A concurrent open sample for the
// same banner is left in place and inserted is false, so the caller can
// reload it without aborting the transaction.
func (s *CookieBannerSample) InsertOpen(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
) (bool, error) {
	q := `
INSERT INTO cookie_banner_samples (
	id,
	tenant_id,
	organization_id,
	cookie_banner_id,
	frozen_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	NULL,
	@created_at,
	@updated_at
)
ON CONFLICT (cookie_banner_id) WHERE frozen_at IS NULL DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"id":               s.ID,
		"tenant_id":        scope.GetTenantID(),
		"organization_id":  s.OrganizationID,
		"cookie_banner_id": s.CookieBannerID,
		"created_at":       s.CreatedAt,
		"updated_at":       s.UpdatedAt,
	}

	tag, err := tx.Exec(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot insert open cookie banner sample: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func FreezeOpenCookieBannerSample(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	cookieBannerID gid.GID,
) error {
	now := time.Now()

	q := `
UPDATE cookie_banner_samples
SET
	frozen_at = @now,
	updated_at = @now
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
	AND frozen_at IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"now":              now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot freeze cookie banner sample: %w", err)
	}

	return nil
}

// FreezeCookieBannerSample freezes one sample. Stop uses this instead of
// FreezeOpenCookieBannerSample so a concurrent Start, which closes the
// loaded row and opens a new one, is not undone by a banner-wide update.
func FreezeCookieBannerSample(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	sampleID gid.GID,
) error {
	now := time.Now()

	q := `
UPDATE cookie_banner_samples
SET
	frozen_at = @now,
	updated_at = @now
WHERE
	%s
	AND id = @sample_id
	AND frozen_at IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"sample_id": sampleID,
		"now":       now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot freeze cookie banner sample: %w", err)
	}

	return nil
}

type CookieBannerSampleRef struct {
	ID             gid.GID `db:"id"`
	CookieBannerID gid.GID `db:"cookie_banner_id"`
}

func LoadCookieBannerSampleRefsByIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	sampleIDs []gid.GID,
) ([]CookieBannerSampleRef, error) {
	if len(sampleIDs) == 0 {
		return nil, nil
	}

	q := `
SELECT
	id,
	cookie_banner_id
FROM cookie_banner_samples
WHERE
	%s
	AND id = ANY(@sample_ids)
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"sample_ids": sampleIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query cookie banner samples: %w", err)
	}

	refs, err := pgx.CollectRows(rows, pgx.RowToStructByName[CookieBannerSampleRef])
	if err != nil {
		return nil, fmt.Errorf("cannot collect cookie banner samples: %w", err)
	}

	return refs, nil
}

func LoadLatestCookieBannerSampleRefsByBannerIDs(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	bannerIDs []gid.GID,
) ([]CookieBannerSampleRef, error) {
	if len(bannerIDs) == 0 {
		return nil, nil
	}

	q := `
SELECT DISTINCT ON (cookie_banner_id)
	id,
	cookie_banner_id
FROM cookie_banner_samples
WHERE
	%s
	AND cookie_banner_id = ANY(@banner_ids)
ORDER BY cookie_banner_id, created_at DESC, id DESC
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{"banner_ids": bannerIDs}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query latest cookie banner samples: %w", err)
	}

	refs, err := pgx.CollectRows(rows, pgx.RowToStructByName[CookieBannerSampleRef])
	if err != nil {
		return nil, fmt.Errorf("cannot collect latest cookie banner samples: %w", err)
	}

	return refs, nil
}

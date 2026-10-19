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
	CookieBannerSample struct {
		ID                         gid.GID    `db:"id"`
		OrganizationID             gid.GID    `db:"organization_id"`
		CookieBannerID             gid.GID    `db:"cookie_banner_id"`
		ChromePageLoads            int        `db:"chrome_page_loads"`
		EdgePageLoads              int        `db:"edge_page_loads"`
		FirefoxPageLoads           int        `db:"firefox_page_loads"`
		SafariPageLoads            int        `db:"safari_page_loads"`
		OtherPageLoads             int        `db:"other_page_loads"`
		ChromeConsentingPageLoads  int        `db:"chrome_consenting_page_loads"`
		EdgeConsentingPageLoads    int        `db:"edge_consenting_page_loads"`
		FirefoxConsentingPageLoads int        `db:"firefox_consenting_page_loads"`
		SafariConsentingPageLoads  int        `db:"safari_consenting_page_loads"`
		OtherConsentingPageLoads   int        `db:"other_consenting_page_loads"`
		FrozenAt                   *time.Time `db:"frozen_at"`
		CreatedAt                  time.Time  `db:"created_at"`
		UpdatedAt                  time.Time  `db:"updated_at"`
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
	chrome_page_loads,
	edge_page_loads,
	firefox_page_loads,
	safari_page_loads,
	other_page_loads,
	chrome_consenting_page_loads,
	edge_consenting_page_loads,
	firefox_consenting_page_loads,
	safari_consenting_page_loads,
	other_consenting_page_loads,
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
	chrome_page_loads,
	edge_page_loads,
	firefox_page_loads,
	safari_page_loads,
	other_page_loads,
	chrome_consenting_page_loads,
	edge_consenting_page_loads,
	firefox_consenting_page_loads,
	safari_consenting_page_loads,
	other_consenting_page_loads,
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
	chrome_page_loads,
	edge_page_loads,
	firefox_page_loads,
	safari_page_loads,
	other_page_loads,
	chrome_consenting_page_loads,
	edge_consenting_page_loads,
	firefox_consenting_page_loads,
	safari_consenting_page_loads,
	other_consenting_page_loads,
	frozen_at,
	created_at,
	updated_at
FROM cookie_banner_samples
WHERE
	%s
	AND cookie_banner_id = @cookie_banner_id
ORDER BY created_at DESC
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
	chrome_page_loads,
	edge_page_loads,
	firefox_page_loads,
	safari_page_loads,
	other_page_loads,
	chrome_consenting_page_loads,
	edge_consenting_page_loads,
	firefox_consenting_page_loads,
	safari_consenting_page_loads,
	other_consenting_page_loads,
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
	chrome_page_loads,
	edge_page_loads,
	firefox_page_loads,
	safari_page_loads,
	other_page_loads,
	chrome_consenting_page_loads,
	edge_consenting_page_loads,
	firefox_consenting_page_loads,
	safari_consenting_page_loads,
	other_consenting_page_loads,
	frozen_at,
	created_at,
	updated_at
) VALUES (
	@id,
	@tenant_id,
	@organization_id,
	@cookie_banner_id,
	@chrome_page_loads,
	@edge_page_loads,
	@firefox_page_loads,
	@safari_page_loads,
	@other_page_loads,
	@chrome_consenting_page_loads,
	@edge_consenting_page_loads,
	@firefox_consenting_page_loads,
	@safari_consenting_page_loads,
	@other_consenting_page_loads,
	@frozen_at,
	@created_at,
	@updated_at
)
`

	args := pgx.StrictNamedArgs{
		"id":                            s.ID,
		"tenant_id":                     scope.GetTenantID(),
		"organization_id":               s.OrganizationID,
		"cookie_banner_id":              s.CookieBannerID,
		"chrome_page_loads":             s.ChromePageLoads,
		"edge_page_loads":               s.EdgePageLoads,
		"firefox_page_loads":            s.FirefoxPageLoads,
		"safari_page_loads":             s.SafariPageLoads,
		"other_page_loads":              s.OtherPageLoads,
		"chrome_consenting_page_loads":  s.ChromeConsentingPageLoads,
		"edge_consenting_page_loads":    s.EdgeConsentingPageLoads,
		"firefox_consenting_page_loads": s.FirefoxConsentingPageLoads,
		"safari_consenting_page_loads":  s.SafariConsentingPageLoads,
		"other_consenting_page_loads":   s.OtherConsentingPageLoads,
		"frozen_at":                     s.FrozenAt,
		"created_at":                    s.CreatedAt,
		"updated_at":                    s.UpdatedAt,
	}

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot insert cookie banner sample: %w", err)
	}

	return nil
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

func RecordSamplePageLoads(
	ctx context.Context,
	tx pg.Tx,
	scope Scoper,
	sampleID gid.GID,
	chrome int64,
	edge int64,
	firefox int64,
	safari int64,
	other int64,
	chromeConsenting int64,
	edgeConsenting int64,
	firefoxConsenting int64,
	safariConsenting int64,
	otherConsenting int64,
) error {
	if chrome == 0 && edge == 0 && firefox == 0 && safari == 0 && other == 0 &&
		chromeConsenting == 0 && edgeConsenting == 0 && firefoxConsenting == 0 &&
		safariConsenting == 0 && otherConsenting == 0 {
		return nil
	}

	now := time.Now()

	q := `
UPDATE cookie_banner_samples
SET
	chrome_page_loads = chrome_page_loads + @chrome,
	edge_page_loads = edge_page_loads + @edge,
	firefox_page_loads = firefox_page_loads + @firefox,
	safari_page_loads = safari_page_loads + @safari,
	other_page_loads = other_page_loads + @other,
	chrome_consenting_page_loads = chrome_consenting_page_loads + @chrome_consenting,
	edge_consenting_page_loads = edge_consenting_page_loads + @edge_consenting,
	firefox_consenting_page_loads = firefox_consenting_page_loads + @firefox_consenting,
	safari_consenting_page_loads = safari_consenting_page_loads + @safari_consenting,
	other_consenting_page_loads = other_consenting_page_loads + @other_consenting,
	updated_at = @now
WHERE
	%s
	AND id = @sample_id
	AND frozen_at IS NULL
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"sample_id":          sampleID,
		"chrome":             chrome,
		"edge":               edge,
		"firefox":            firefox,
		"safari":             safari,
		"other":              other,
		"chrome_consenting":  chromeConsenting,
		"edge_consenting":    edgeConsenting,
		"firefox_consenting": firefoxConsenting,
		"safari_consenting":  safariConsenting,
		"other_consenting":   otherConsenting,
		"now":                now,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := tx.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot record sample page loads: %w", err)
	}

	return nil
}

func IsOpenSampleConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == "23505" && pgErr.ConstraintName == "cookie_banner_samples_open_uidx"
}

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
	CookieBannerGVLVendor struct {
		CookieBannerID gid.GID   `db:"cookie_banner_id"`
		IABVendorID    int       `db:"iab_vendor_id"`
		OrganizationID gid.GID   `db:"organization_id"`
		CreatedAt      time.Time `db:"created_at"`
	}

	CookieBannerGVLVendors []*CookieBannerGVLVendor
)

// Upsert links a cookie banner to a GVL vendor. The organization_id stored
// in the junction row is derived from the cookie_banners table inside the
// INSERT, so a caller cannot place the mapping into a different organization
// than the banner actually belongs to. Idempotent: re-linking an existing
// pair is a no-op.
func (l CookieBannerGVLVendor) Upsert(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
) error {
	q := `
INSERT INTO
    cookie_banner_gvl_vendors (
        cookie_banner_id,
        iab_vendor_id,
        organization_id,
        tenant_id,
        created_at
    )
SELECT
    @cookie_banner_id,
    @iab_vendor_id,
    b.organization_id,
    @tenant_id,
    @created_at
FROM
    cookie_banners b
WHERE
    b.id = @cookie_banner_id
    AND b.tenant_id = @tenant_id
ON CONFLICT (cookie_banner_id, iab_vendor_id) DO NOTHING
`

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": l.CookieBannerID,
		"iab_vendor_id":    l.IABVendorID,
		"tenant_id":        scope.GetTenantID(),
		"created_at":       l.CreatedAt,
	}

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot upsert cookie banner gvl vendor: %w", err)
	}

	return nil
}

func (l CookieBannerGVLVendor) Delete(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	iabVendorID int,
) error {
	q := `
DELETE
FROM
    cookie_banner_gvl_vendors
WHERE
    %s
    AND cookie_banner_id = @cookie_banner_id
    AND iab_vendor_id = @iab_vendor_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
		"iab_vendor_id":    iabVendorID,
	}
	maps.Copy(args, scope.SQLArguments())

	if _, err := conn.Exec(ctx, q, args); err != nil {
		return fmt.Errorf("cannot delete cookie banner gvl vendor: %w", err)
	}

	return nil
}

// LoadIABVendorIDsByCookieBannerID returns the disclosed IAB vendor IDs for a
// banner. This is a small curated projection used to freeze the set on a
// version snapshot.
func (l *CookieBannerGVLVendors) LoadIABVendorIDsByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
) ([]int, error) {
	q := `
SELECT
    iab_vendor_id
FROM
    cookie_banner_gvl_vendors
WHERE
    %s
    AND cookie_banner_id = @cookie_banner_id
ORDER BY
    iab_vendor_id
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
	}
	maps.Copy(args, scope.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("cannot query cookie banner gvl vendor ids: %w", err)
	}

	ids, err := pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		return nil, fmt.Errorf("cannot collect cookie banner gvl vendor ids: %w", err)
	}

	if ids == nil {
		ids = []int{}
	}

	return ids, nil
}

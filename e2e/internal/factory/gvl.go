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

package factory

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/internal/test"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// SeedCommonGVLVendor inserts a catalog GVL vendor (and its snapshot) for e2e
// tests, drawing a fresh iab_vendor_id and vendor_list_version from a dedicated
// sequence. Sequence values are never handed out twice, so parallel subtests and
// concurrent test binaries sharing the database never collide.
func SeedCommonGVLVendor(t *testing.T, name string, deleted bool) (iabVendorID int, version int) {
	t.Helper()

	client := test.PGClient(t)
	ctx := context.Background()
	now := time.Now()
	tenantID := gid.NewTenantID()
	snapshotID := gid.New(tenantID, coredata.CommonGVLSnapshotEntityType)
	vendorID := gid.New(tenantID, coredata.CommonGVLVendorEntityType)

	err := client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		// The real GVL uses low vendor ids, so start well above them to keep
		// seeded vendors distinguishable from any imported catalog rows.
		_, err := conn.Exec(ctx, `
CREATE SEQUENCE IF NOT EXISTS e2e_gvl_seq START WITH 10000001
`)
		if err != nil {
			return err
		}

		if err := conn.QueryRow(ctx, `SELECT nextval('e2e_gvl_seq')`).Scan(&iabVendorID); err != nil {
			return err
		}

		version = iabVendorID

		_, err = conn.Exec(ctx, `
INSERT INTO common_gvl_snapshots (
    id,
    vendor_list_version,
    gvl_specification_version,
    tcf_policy_version,
    last_updated,
    fetched_at,
    payload
) VALUES ($1, $2, 3, 5, $3, $3, '{}'::jsonb)
`, snapshotID, version, now)
		if err != nil {
			return err
		}

		var deletedDate any
		if deleted {
			deletedDate = now.Add(-time.Hour)
		}

		_, err = conn.Exec(ctx, `
INSERT INTO common_gvl_vendors (
    id,
    iab_vendor_id,
    vendor_list_version,
    name,
    deleted_date,
    purposes,
    leg_int_purposes,
    flexible_purposes,
    special_purposes,
    features,
    special_features,
    policy_url,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    ARRAY[1, 2]::integer[],
    ARRAY[7]::integer[],
    '{}'::integer[],
    '{}'::integer[],
    '{}'::integer[],
    ARRAY[1]::integer[],
    'https://example.com/privacy',
    $6, $6
)
`, vendorID, iabVendorID, version, name, deletedDate, now)

		return err
	})
	require.NoError(t, err, "test setup: cannot seed common gvl vendor")

	return iabVendorID, version
}

// SeedCommonGVLCatalogState points the singleton catalog pointer at version.
func SeedCommonGVLCatalogState(t *testing.T, vendorListVersion int) {
	t.Helper()

	client := test.PGClient(t)
	ctx := context.Background()
	now := time.Now()

	err := client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(ctx, `
INSERT INTO common_gvl_state (
    singleton,
    latest_vendor_list_version,
    last_fetched_at
) VALUES (TRUE, $1, $2)
ON CONFLICT (singleton) DO UPDATE
SET
    latest_vendor_list_version = EXCLUDED.latest_vendor_list_version,
    last_fetched_at = EXCLUDED.last_fetched_at
`, vendorListVersion, now)

		return err
	})
	require.NoError(t, err, "test setup: cannot seed common gvl catalog state")
}

// EnableCookieBannerTCF flips the hidden tcf capability on a banner.
func EnableCookieBannerTCF(t *testing.T, bannerID string) {
	t.Helper()

	client := test.PGClient(t)
	ctx := context.Background()

	err := client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(ctx, `
UPDATE cookie_banners
SET capabilities = capabilities || '{"tcf": true}'::jsonb
WHERE id = $1
`, bannerID)

		return err
	})
	require.NoError(t, err, "test setup: cannot enable cookie banner tcf")
}

// DisableCookieBannerTCF flips the hidden tcf capability back off on a banner.
func DisableCookieBannerTCF(t *testing.T, bannerID string) {
	t.Helper()

	client := test.PGClient(t)
	ctx := context.Background()

	err := client.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		_, err := conn.Exec(ctx, `
UPDATE cookie_banners
SET capabilities = capabilities || '{"tcf": false}'::jsonb
WHERE id = $1
`, bannerID)

		return err
	})
	require.NoError(t, err, "test setup: cannot disable cookie banner tcf")
}

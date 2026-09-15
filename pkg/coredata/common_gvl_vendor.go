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
	"go.probo.inc/probo/pkg/page"
)

type (
	CommonGVLVendor struct {
		ID                  gid.GID    `db:"id"`
		IABVendorID         int        `db:"iab_vendor_id"`
		VendorListVersion   int        `db:"vendor_list_version"`
		Name                string     `db:"name"`
		DeletedDate         *time.Time `db:"deleted_date"`
		Purposes            []int32    `db:"purposes"`
		LegIntPurposes      []int32    `db:"leg_int_purposes"`
		FlexiblePurposes    []int32    `db:"flexible_purposes"`
		SpecialPurposes     []int32    `db:"special_purposes"`
		Features            []int32    `db:"features"`
		SpecialFeatures     []int32    `db:"special_features"`
		PolicyURL           *string    `db:"policy_url"`
		UsesCookies         *bool      `db:"uses_cookies"`
		CookieRefresh       *bool      `db:"cookie_refresh"`
		UsesNonCookieAccess *bool      `db:"uses_non_cookie_access"`
		CookieMaxAgeSeconds *int       `db:"cookie_max_age_seconds"`
		CreatedAt           time.Time  `db:"created_at"`
		UpdatedAt           time.Time  `db:"updated_at"`
	}

	CommonGVLVendors []*CommonGVLVendor
)

func (v CommonGVLVendor) IsDeleted(now time.Time) bool {
	return v.DeletedDate != nil && !v.DeletedDate.After(now)
}

func (v *CommonGVLVendor) CursorKey(field CommonGVLVendorOrderField) page.CursorKey {
	switch field {
	case CommonGVLVendorOrderFieldName:
		return page.NewCursorKey(v.ID, v.Name)
	case CommonGVLVendorOrderFieldIABVendorID:
		return page.NewCursorKey(v.ID, v.IABVendorID)
	}

	panic(fmt.Sprintf("unsupported order by: %s", field))
}

func (v *CommonGVLVendor) LoadByIABVendorID(
	ctx context.Context,
	conn pg.Querier,
	iabVendorID int,
) error {
	q := `
SELECT
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
    uses_cookies,
    cookie_refresh,
    uses_non_cookie_access,
    cookie_max_age_seconds,
    created_at,
    updated_at
FROM
    common_gvl_vendors
WHERE
    iab_vendor_id = @iab_vendor_id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"iab_vendor_id": iabVendorID}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common gvl vendor: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonGVLVendor])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect common gvl vendor: %w", err)
	}

	*v = row

	return nil
}

func (v *CommonGVLVendor) Upsert(
	ctx context.Context,
	conn pg.Querier,
) (inserted bool, err error) {
	q := `
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
    uses_cookies,
    cookie_refresh,
    uses_non_cookie_access,
    cookie_max_age_seconds,
    created_at,
    updated_at
) VALUES (
    @id,
    @iab_vendor_id,
    @vendor_list_version,
    @name,
    @deleted_date,
    @purposes,
    @leg_int_purposes,
    @flexible_purposes,
    @special_purposes,
    @features,
    @special_features,
    @policy_url,
    @uses_cookies,
    @cookie_refresh,
    @uses_non_cookie_access,
    @cookie_max_age_seconds,
    @created_at,
    @updated_at
)
ON CONFLICT (iab_vendor_id) DO UPDATE
SET
    vendor_list_version    = EXCLUDED.vendor_list_version,
    name                   = EXCLUDED.name,
    deleted_date           = EXCLUDED.deleted_date,
    purposes               = EXCLUDED.purposes,
    leg_int_purposes       = EXCLUDED.leg_int_purposes,
    flexible_purposes      = EXCLUDED.flexible_purposes,
    special_purposes       = EXCLUDED.special_purposes,
    features               = EXCLUDED.features,
    special_features       = EXCLUDED.special_features,
    policy_url             = EXCLUDED.policy_url,
    uses_cookies           = EXCLUDED.uses_cookies,
    cookie_refresh         = EXCLUDED.cookie_refresh,
    uses_non_cookie_access = EXCLUDED.uses_non_cookie_access,
    cookie_max_age_seconds = EXCLUDED.cookie_max_age_seconds,
    updated_at             = EXCLUDED.updated_at
RETURNING
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
    uses_cookies,
    cookie_refresh,
    uses_non_cookie_access,
    cookie_max_age_seconds,
    created_at,
    updated_at
`

	originalID := v.ID

	args := pgx.StrictNamedArgs{
		"id":                     v.ID,
		"iab_vendor_id":          v.IABVendorID,
		"vendor_list_version":    v.VendorListVersion,
		"name":                   v.Name,
		"deleted_date":           v.DeletedDate,
		"purposes":               emptyInt32s(v.Purposes),
		"leg_int_purposes":       emptyInt32s(v.LegIntPurposes),
		"flexible_purposes":      emptyInt32s(v.FlexiblePurposes),
		"special_purposes":       emptyInt32s(v.SpecialPurposes),
		"features":               emptyInt32s(v.Features),
		"special_features":       emptyInt32s(v.SpecialFeatures),
		"policy_url":             v.PolicyURL,
		"uses_cookies":           v.UsesCookies,
		"cookie_refresh":         v.CookieRefresh,
		"uses_non_cookie_access": v.UsesNonCookieAccess,
		"cookie_max_age_seconds": v.CookieMaxAgeSeconds,
		"created_at":             v.CreatedAt,
		"updated_at":             v.UpdatedAt,
	}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return false, fmt.Errorf("cannot upsert common gvl vendor: %w", err)
	}
	defer rows.Close()

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonGVLVendor])
	if err != nil {
		return false, fmt.Errorf("cannot collect upserted common gvl vendor: %w", err)
	}

	*v = row

	return v.ID == originalID, nil
}

func emptyInt32s(values []int32) []int32 {
	if values == nil {
		return []int32{}
	}

	return values
}

func (v *CommonGVLVendor) LoadByID(
	ctx context.Context,
	conn pg.Querier,
	id gid.GID,
) error {
	q := `
SELECT
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
    uses_cookies,
    cookie_refresh,
    uses_non_cookie_access,
    cookie_max_age_seconds,
    created_at,
    updated_at
FROM
    common_gvl_vendors
WHERE
    id = @id
LIMIT 1;
`

	args := pgx.StrictNamedArgs{"id": id}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common gvl vendor: %w", err)
	}

	row, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[CommonGVLVendor])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResourceNotFound
		}

		return fmt.Errorf("cannot collect common gvl vendor: %w", err)
	}

	*v = row

	return nil
}

func (v *CommonGVLVendors) Load(
	ctx context.Context,
	conn pg.Querier,
	cursor *page.Cursor[CommonGVLVendorOrderField],
	filter *CommonGVLVendorFilter,
) error {
	q := `
SELECT
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
    uses_cookies,
    cookie_refresh,
    uses_non_cookie_access,
    cookie_max_age_seconds,
    created_at,
    updated_at
FROM
    common_gvl_vendors
WHERE
    (deleted_date IS NULL OR deleted_date > NOW())
    AND %s
    AND %s
`

	q = fmt.Sprintf(q, filter.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{}
	maps.Copy(args, filter.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common gvl vendors: %w", err)
	}

	vendors, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CommonGVLVendor])
	if err != nil {
		return fmt.Errorf("cannot collect common gvl vendors: %w", err)
	}

	*v = vendors

	return nil
}

func (v *CommonGVLVendors) Count(
	ctx context.Context,
	conn pg.Querier,
	filter *CommonGVLVendorFilter,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    common_gvl_vendors
WHERE
    (deleted_date IS NULL OR deleted_date > NOW())
    AND %s
`

	q = fmt.Sprintf(q, filter.SQLFragment())

	args := pgx.StrictNamedArgs{}
	maps.Copy(args, filter.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count common gvl vendors: %w", err)
	}

	return count, nil
}

func (v *CommonGVLVendors) LoadByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
	cursor *page.Cursor[CommonGVLVendorOrderField],
) error {
	q := `
SELECT
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
    uses_cookies,
    cookie_refresh,
    uses_non_cookie_access,
    cookie_max_age_seconds,
    created_at,
    updated_at
FROM
    common_gvl_vendors
WHERE
    iab_vendor_id IN (
        SELECT
            iab_vendor_id
        FROM
            cookie_banner_gvl_vendors
        WHERE
            %s
            AND cookie_banner_id = @cookie_banner_id
    )
    AND %s
`

	q = fmt.Sprintf(q, scope.SQLFragment(), cursor.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
	}
	maps.Copy(args, scope.SQLArguments())
	maps.Copy(args, cursor.SQLArguments())

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query cookie banner gvl vendors: %w", err)
	}

	vendors, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CommonGVLVendor])
	if err != nil {
		return fmt.Errorf("cannot collect cookie banner gvl vendors: %w", err)
	}

	*v = vendors

	return nil
}

func (v *CommonGVLVendors) CountByCookieBannerID(
	ctx context.Context,
	conn pg.Querier,
	scope Scoper,
	cookieBannerID gid.GID,
) (int, error) {
	q := `
SELECT
    COUNT(id)
FROM
    common_gvl_vendors
WHERE
    iab_vendor_id IN (
        SELECT
            iab_vendor_id
        FROM
            cookie_banner_gvl_vendors
        WHERE
            %s
            AND cookie_banner_id = @cookie_banner_id
    )
`

	q = fmt.Sprintf(q, scope.SQLFragment())

	args := pgx.StrictNamedArgs{
		"cookie_banner_id": cookieBannerID,
	}
	maps.Copy(args, scope.SQLArguments())

	row := conn.QueryRow(ctx, q, args)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("cannot count cookie banner gvl vendors: %w", err)
	}

	return count, nil
}

func (v *CommonGVLVendors) LoadByIABVendorIDs(
	ctx context.Context,
	conn pg.Querier,
	iabVendorIDs []int,
) error {
	if len(iabVendorIDs) == 0 {
		*v = CommonGVLVendors{}
		return nil
	}

	q := `
SELECT
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
    uses_cookies,
    cookie_refresh,
    uses_non_cookie_access,
    cookie_max_age_seconds,
    created_at,
    updated_at
FROM
    common_gvl_vendors
WHERE
    iab_vendor_id = ANY(@iab_vendor_ids)
ORDER BY
    iab_vendor_id
`

	args := pgx.StrictNamedArgs{"iab_vendor_ids": iabVendorIDs}

	rows, err := conn.Query(ctx, q, args)
	if err != nil {
		return fmt.Errorf("cannot query common gvl vendors by iab ids: %w", err)
	}

	vendors, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[CommonGVLVendor])
	if err != nil {
		return fmt.Errorf("cannot collect common gvl vendors by iab ids: %w", err)
	}

	*v = vendors

	return nil
}

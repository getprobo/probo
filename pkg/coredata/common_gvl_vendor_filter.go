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
	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/gid"
)

type CommonGVLVendorFilter struct {
	query          *string
	cookieBannerID *gid.GID
	membership     *CommonGVLVendorMembership
}

func NewCommonGVLVendorFilter(query *string) *CommonGVLVendorFilter {
	return &CommonGVLVendorFilter{query: query}
}

func (f *CommonGVLVendorFilter) WithMembership(
	cookieBannerID *gid.GID,
	membership *CommonGVLVendorMembership,
) *CommonGVLVendorFilter {
	f.cookieBannerID = cookieBannerID
	f.membership = membership

	return f
}

func (f *CommonGVLVendorFilter) SQLFragment() string {
	return `(
	CASE
		WHEN @filter_query::text IS NOT NULL AND @filter_query::text != '' THEN
			(name ILIKE '%' || @filter_query || '%'
			 OR iab_vendor_id::text ILIKE '%' || @filter_query || '%')
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_membership::text = @membership_on_banner
			AND @filter_cookie_banner_id::text IS NOT NULL THEN
			iab_vendor_id IN (
				SELECT
					iab_vendor_id
				FROM
					cookie_banner_gvl_vendors
				WHERE
					cookie_banner_id = @filter_cookie_banner_id
			)
		WHEN @filter_membership::text = @membership_not_on_banner
			AND @filter_cookie_banner_id::text IS NOT NULL THEN
			iab_vendor_id NOT IN (
				SELECT
					iab_vendor_id
				FROM
					cookie_banner_gvl_vendors
				WHERE
					cookie_banner_id = @filter_cookie_banner_id
			)
		ELSE TRUE
	END
)`
}

func (f *CommonGVLVendorFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"filter_query":             nil,
		"filter_cookie_banner_id":  nil,
		"filter_membership":        nil,
		"membership_on_banner":     CommonGVLVendorMembershipOnBanner,
		"membership_not_on_banner": CommonGVLVendorMembershipNotOnBanner,
	}

	if f != nil && f.query != nil {
		args["filter_query"] = *f.query
	}

	if f != nil && f.cookieBannerID != nil && f.membership != nil {
		args["filter_cookie_banner_id"] = *f.cookieBannerID
		args["filter_membership"] = *f.membership
	}

	return args
}

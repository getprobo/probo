// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
	"github.com/jackc/pgx/v5"
)

type (
	ThirdPartyFilter struct {
		showOnTrustCenter *bool
		level             *int
		query             *string
	}
)

func NewThirdPartyFilter(showOnTrustCenter *bool, level *int, query *string) *ThirdPartyFilter {
	return &ThirdPartyFilter{
		showOnTrustCenter: showOnTrustCenter,
		level:             level,
		query:             query,
	}
}

func (f *ThirdPartyFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"show_on_trust_center": nil,
		"filter_query":         nil,
		"level":                nil,
	}

	if f.showOnTrustCenter != nil {
		args["show_on_trust_center"] = *f.showOnTrustCenter
	}

	if f.query != nil && *f.query != "" {
		args["filter_query"] = *f.query
	}

	if f.level != nil {
		args["level"] = *f.level
	}

	return args
}

func (f *ThirdPartyFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @show_on_trust_center::boolean IS NOT NULL THEN
			show_on_trust_center = @show_on_trust_center::boolean
		ELSE TRUE
	END
	AND CASE
		WHEN @level::integer IS NOT NULL THEN
			level = @level::integer
		ELSE TRUE
	END
	AND CASE
		WHEN @filter_query::text IS NOT NULL AND @filter_query::text <> '' THEN
			name ILIKE '%' || @filter_query || '%'
		ELSE TRUE
	END
)`
}

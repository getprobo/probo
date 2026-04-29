// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	RiskFilter struct {
		query *string
	}
)

func NewRiskFilter(query *string) *RiskFilter {
	return &RiskFilter{
		query: query,
	}
}

func (f *RiskFilter) SQLArguments() pgx.StrictNamedArgs {
	return pgx.StrictNamedArgs{
		"query": f.query,
	}
}

func (f *RiskFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @query::text IS NOT NULL AND @query::text != '' THEN
			search_vector @@ (
				SELECT to_tsquery('simple', string_agg(lexeme || ':*', ' & '))
				FROM unnest(regexp_split_to_array(trim(@query), '\s+')) AS lexeme
			)
		ELSE TRUE
	END
)`
}

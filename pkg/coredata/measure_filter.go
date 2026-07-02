// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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
)

type (
	MeasureFilter struct {
		query    *string
		state    *MeasureState
		category *string
	}
)

func NewMeasureFilter(query *string, state *MeasureState, category *string) *MeasureFilter {
	return &MeasureFilter{
		query:    query,
		state:    state,
		category: category,
	}
}

func (f *MeasureFilter) SQLArguments() pgx.NamedArgs {
	return pgx.NamedArgs{
		"query":    f.query,
		"state":    f.state,
		"category": f.category,
	}
}

func (f *MeasureFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @query::text IS NULL OR @query::text = '' THEN
			TRUE
		ELSE
			search_vector @@ (
				SELECT to_tsquery('simple', string_agg(lexeme || ':*', ' & '))
				FROM unnest(regexp_split_to_array(trim(@query), '\s+')) AS lexeme
			)
	END
)
AND
(
	CASE
		WHEN @state::mitigation_state IS NULL THEN
			TRUE
		ELSE
			state = @state::mitigation_state
	END
)
AND
(
	CASE
		WHEN @category::text IS NULL OR @category::text = '' THEN
			TRUE
		ELSE
			category = @category::text
	END
)
	`
}

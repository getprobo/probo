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
	ControlFilter struct {
		query *string
	}
)

func NewControlFilter(query *string) *ControlFilter {
	return &ControlFilter{
		query: query,
	}
}

func (f *ControlFilter) SQLArguments() pgx.NamedArgs {
	args := pgx.NamedArgs{}

	if f.query != nil {
		args["query"] = *f.query
	}

	return args
}

func (f *ControlFilter) SQLFragment() string {
	if f.query == nil || *f.query == "" {
		return "TRUE"
	}

	return `
		search_vector @@ (
			SELECT to_tsquery('simple', string_agg(lexeme || ':*', ' & '))
			FROM unnest(regexp_split_to_array(trim(@query), '\s+')) AS lexeme
		)
	`
}

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
)

type CookieCategoryFilter struct {
	excludeKind *CookieCategoryKind
}

func NewCookieCategoryFilter(excludeKind *CookieCategoryKind) *CookieCategoryFilter {
	return &CookieCategoryFilter{excludeKind: excludeKind}
}

func (f *CookieCategoryFilter) SQLFragment() string {
	if f == nil {
		return "TRUE"
	}

	return `(
	CASE
		WHEN @has_exclude_kind_filter::boolean = false THEN TRUE
		WHEN @has_exclude_kind_filter::boolean = true THEN
			kind != @filter_exclude_kind::cookie_category_kind
		ELSE TRUE
	END
)`
}

func (f *CookieCategoryFilter) SQLArguments() pgx.StrictNamedArgs {
	if f == nil {
		return pgx.StrictNamedArgs{
			"has_exclude_kind_filter": false,
			"filter_exclude_kind":     nil,
		}
	}

	args := pgx.StrictNamedArgs{
		"has_exclude_kind_filter": false,
		"filter_exclude_kind":     nil,
	}

	if f.excludeKind != nil {
		args["has_exclude_kind_filter"] = true
		args["filter_exclude_kind"] = string(*f.excludeKind)
	}

	return args
}

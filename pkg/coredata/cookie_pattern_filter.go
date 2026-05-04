// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/pkg/gid"
)

type CookiePatternFilter struct {
	matchType        *CookiePatternMatchType
	cookieCategoryID *gid.GID
	excluded         *bool
	query            *string
	source           *CookieSource
}

func NewCookiePatternFilter(
	matchType *CookiePatternMatchType,
	cookieCategoryID *gid.GID,
	excluded *bool,
) *CookiePatternFilter {
	return &CookiePatternFilter{
		matchType:        matchType,
		cookieCategoryID: cookieCategoryID,
		excluded:         excluded,
	}
}

func (f *CookiePatternFilter) WithQuery(query *string) *CookiePatternFilter {
	f.query = query
	return f
}

func (f *CookiePatternFilter) WithSource(source *CookieSource) *CookiePatternFilter {
	f.source = source
	return f
}

func (f *CookiePatternFilter) SQLFragment() string {
	if f == nil {
		return "TRUE"
	}

	return `
(
	CASE
		WHEN @has_match_type_filter::boolean = false THEN TRUE
		WHEN @has_match_type_filter::boolean = true THEN
			match_type = @filter_match_type::cookie_pattern_match_type
		ELSE TRUE
	END
	AND
	CASE
		WHEN @has_cookie_category_id_filter::boolean = false THEN TRUE
		WHEN @has_cookie_category_id_filter::boolean = true THEN
			cookie_category_id = @filter_cookie_category_id::text
		ELSE TRUE
	END
	AND
	CASE
		WHEN @has_excluded_filter::boolean = false THEN TRUE
		WHEN @has_excluded_filter::boolean = true THEN
			excluded = @filter_excluded
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_query::text IS NOT NULL AND @filter_query::text != '' THEN
			(display_name ILIKE '%' || @filter_query || '%'
			 OR description ILIKE '%' || @filter_query || '%')
		ELSE TRUE
	END
	AND
	CASE
		WHEN @has_source_filter::boolean = false THEN TRUE
		WHEN @has_source_filter::boolean = true THEN
			source = @filter_source::cookie_source
		ELSE TRUE
	END
)`
}

func (f *CookiePatternFilter) SQLArguments() pgx.StrictNamedArgs {
	if f == nil {
		return pgx.StrictNamedArgs{}
	}

	args := pgx.StrictNamedArgs{
		"has_match_type_filter":         false,
		"filter_match_type":             nil,
		"has_cookie_category_id_filter": false,
		"filter_cookie_category_id":     nil,
		"has_excluded_filter":           false,
		"filter_excluded":               nil,
		"filter_query":                  nil,
		"has_source_filter":             false,
		"filter_source":                 nil,
	}

	if f.matchType != nil {
		args["has_match_type_filter"] = true
		args["filter_match_type"] = string(*f.matchType)
	}

	if f.cookieCategoryID != nil {
		args["has_cookie_category_id_filter"] = true
		args["filter_cookie_category_id"] = *f.cookieCategoryID
	}

	if f.excluded != nil {
		args["has_excluded_filter"] = true
		args["filter_excluded"] = *f.excluded
	}

	if f.query != nil {
		args["filter_query"] = *f.query
	}

	if f.source != nil {
		args["has_source_filter"] = true
		args["filter_source"] = string(*f.source)
	}

	return args
}

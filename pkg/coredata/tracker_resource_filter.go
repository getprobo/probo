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

type TrackerResourceFilter struct {
	cookieCategoryID *gid.GID
	excluded         *bool
	query            *string
	resourceType     *TrackerResourceType
}

func NewTrackerResourceFilter(
	cookieCategoryID *gid.GID,
	excluded *bool,
) *TrackerResourceFilter {
	return &TrackerResourceFilter{
		cookieCategoryID: cookieCategoryID,
		excluded:         excluded,
	}
}

func (f *TrackerResourceFilter) WithQuery(query *string) *TrackerResourceFilter {
	f.query = query
	return f
}

func (f *TrackerResourceFilter) WithResourceType(resourceType *TrackerResourceType) *TrackerResourceFilter {
	f.resourceType = resourceType
	return f
}

func (f *TrackerResourceFilter) SQLFragment() string {
	if f == nil {
		return "TRUE"
	}

	return `
(
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
			(origin ILIKE '%' || @filter_query || '%'
			 OR path ILIKE '%' || @filter_query || '%'
			 OR display_name ILIKE '%' || @filter_query || '%'
			 OR description ILIKE '%' || @filter_query || '%')
		ELSE TRUE
	END
	AND
	CASE
		WHEN @has_resource_type_filter::boolean = false THEN TRUE
		WHEN @has_resource_type_filter::boolean = true THEN
			resource_type = @filter_resource_type::tracker_resource_type
		ELSE TRUE
	END
)`
}

func (f *TrackerResourceFilter) SQLArguments() pgx.StrictNamedArgs {
	if f == nil {
		return pgx.StrictNamedArgs{}
	}

	args := pgx.StrictNamedArgs{
		"has_cookie_category_id_filter": false,
		"filter_cookie_category_id":     nil,
		"has_excluded_filter":           false,
		"filter_excluded":               nil,
		"filter_query":                  nil,
		"has_resource_type_filter":      false,
		"filter_resource_type":          nil,
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

	if f.resourceType != nil {
		args["has_resource_type_filter"] = true
		args["filter_resource_type"] = string(*f.resourceType)
	}

	return args
}

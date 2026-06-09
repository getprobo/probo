// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

type TrackerPatternFilter struct {
	matchType          *TrackerPatternMatchType
	cookieCategoryID   *gid.GID
	excluded           *bool
	query              *string
	source             *CookieSource
	trackerType        *TrackerType
	thirdPartyID       *gid.GID
	commonThirdPartyID *gid.GID
}

func NewTrackerPatternFilter(
	matchType *TrackerPatternMatchType,
	cookieCategoryID *gid.GID,
	excluded *bool,
) *TrackerPatternFilter {
	return &TrackerPatternFilter{
		matchType:        matchType,
		cookieCategoryID: cookieCategoryID,
		excluded:         excluded,
	}
}

func (f *TrackerPatternFilter) WithQuery(query *string) *TrackerPatternFilter {
	f.query = query
	return f
}

func (f *TrackerPatternFilter) WithSource(source *CookieSource) *TrackerPatternFilter {
	f.source = source
	return f
}

func (f *TrackerPatternFilter) WithTrackerType(trackerType *TrackerType) *TrackerPatternFilter {
	f.trackerType = trackerType
	return f
}

func (f *TrackerPatternFilter) WithThirdPartyID(thirdPartyID *gid.GID) *TrackerPatternFilter {
	f.thirdPartyID = thirdPartyID
	return f
}

func (f *TrackerPatternFilter) WithCommonThirdPartyID(id *gid.GID) *TrackerPatternFilter {
	f.commonThirdPartyID = id
	return f
}

func (f *TrackerPatternFilter) SQLFragment() string {
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
	AND
	CASE
		WHEN @has_tracker_type_filter::boolean = false THEN TRUE
		WHEN @has_tracker_type_filter::boolean = true THEN
			tracker_type = @filter_tracker_type::tracker_type
		ELSE TRUE
	END
	AND
	CASE
		WHEN @has_third_party_id_filter::boolean = false THEN TRUE
		WHEN @has_third_party_id_filter::boolean = true THEN
			third_party_id = @filter_third_party_id::text
		ELSE TRUE
	END
	AND
	CASE
		WHEN @has_common_third_party_id_filter::boolean = false THEN TRUE
		WHEN @has_common_third_party_id_filter::boolean = true THEN
			common_tracker_pattern_id IN (
				SELECT id FROM common_tracker_patterns
				WHERE common_third_party_id = @filter_common_third_party_id::text
			)
		ELSE TRUE
	END
)`
}

func (f *TrackerPatternFilter) SQLArguments() pgx.StrictNamedArgs {
	if f == nil {
		return pgx.StrictNamedArgs{}
	}

	args := pgx.StrictNamedArgs{
		"has_match_type_filter":            false,
		"filter_match_type":                nil,
		"has_cookie_category_id_filter":    false,
		"filter_cookie_category_id":        nil,
		"has_excluded_filter":              false,
		"filter_excluded":                  nil,
		"filter_query":                     nil,
		"has_source_filter":                false,
		"filter_source":                    nil,
		"has_tracker_type_filter":          false,
		"filter_tracker_type":              nil,
		"has_third_party_id_filter":        false,
		"filter_third_party_id":            nil,
		"has_common_third_party_id_filter": false,
		"filter_common_third_party_id":     nil,
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

	if f.trackerType != nil {
		args["has_tracker_type_filter"] = true
		args["filter_tracker_type"] = string(*f.trackerType)
	}

	if f.thirdPartyID != nil {
		args["has_third_party_id_filter"] = true
		args["filter_third_party_id"] = *f.thirdPartyID
	}

	if f.commonThirdPartyID != nil {
		args["has_common_third_party_id_filter"] = true
		args["filter_common_third_party_id"] = *f.commonThirdPartyID
	}

	return args
}

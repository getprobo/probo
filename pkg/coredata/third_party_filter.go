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
	ThirdPartyFilter struct {
		showOnTrustCenter *bool
		level             *int
		query             *string
		category          *ThirdPartyCategory
		country           *CountryCode
	}
)

func NewThirdPartyFilter(
	showOnTrustCenter *bool,
	level *int,
	query *string,
	category *ThirdPartyCategory,
	country *CountryCode,
) *ThirdPartyFilter {
	return &ThirdPartyFilter{
		showOnTrustCenter: showOnTrustCenter,
		level:             level,
		query:             query,
		category:          category,
		country:           country,
	}
}

func (f *ThirdPartyFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"show_on_trust_center": nil,
		"filter_query":         nil,
		"level":                nil,
		"filter_category":      nil,
		"filter_country":       nil,
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

	if f.category != nil {
		args["filter_category"] = string(*f.category)
	}

	if f.country != nil {
		args["filter_country"] = string(*f.country)
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
	AND CASE
		WHEN @filter_category::text IS NOT NULL THEN
			category = @filter_category::third_party_category
		ELSE TRUE
	END
	AND CASE
		WHEN @filter_country::text IS NOT NULL THEN
			@filter_country::country_code = ANY(countries)
		ELSE TRUE
	END
)`
}

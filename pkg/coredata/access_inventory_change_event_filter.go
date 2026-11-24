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
	"time"

	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/gid"
)

type AccessInventoryChangeEventFilter struct {
	AccessReviewSourceID *gid.GID
	ChangeTypes          []AccessInventoryChangeType
	ObservedAtOnOrAfter  *time.Time
	ObservedAtOnOrBefore *time.Time
}

func (f *AccessInventoryChangeEventFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @filter_access_review_source_id::text IS NOT NULL THEN
			access_review_source_id = @filter_access_review_source_id
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_change_types::access_inventory_change_type[] IS NOT NULL THEN
			change_type = ANY(@filter_change_types::access_inventory_change_type[])
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_observed_at_on_or_after::timestamptz IS NOT NULL THEN
			observed_at >= @filter_observed_at_on_or_after
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_observed_at_on_or_before::timestamptz IS NOT NULL THEN
			observed_at <= @filter_observed_at_on_or_before
		ELSE TRUE
	END
)`
}

func (f *AccessInventoryChangeEventFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"filter_access_review_source_id":  nil,
		"filter_change_types":             nil,
		"filter_observed_at_on_or_after":  nil,
		"filter_observed_at_on_or_before": nil,
	}

	if f == nil {
		return args
	}

	if f.AccessReviewSourceID != nil {
		args["filter_access_review_source_id"] = *f.AccessReviewSourceID
	}

	if len(f.ChangeTypes) > 0 {
		args["filter_change_types"] = AccessInventoryChangeTypeValues(f.ChangeTypes)
	}

	if f.ObservedAtOnOrAfter != nil {
		args["filter_observed_at_on_or_after"] = *f.ObservedAtOnOrAfter
	}

	if f.ObservedAtOnOrBefore != nil {
		args["filter_observed_at_on_or_before"] = *f.ObservedAtOnOrBefore
	}

	return args
}

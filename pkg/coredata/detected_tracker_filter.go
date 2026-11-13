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
)

type DetectedTrackerFilter struct {
	trackerTypes []TrackerType
	identifiers  []string
	updatedSince *time.Time
}

func NewDetectedTrackerFilter(trackerTypes []TrackerType) *DetectedTrackerFilter {
	return &DetectedTrackerFilter{trackerTypes: trackerTypes}
}

// WithIdentifiers restricts results to detections whose identifier is one
// of the given names. A nil slice disables the filter. An empty slice
// matches nothing.
func (f *DetectedTrackerFilter) WithIdentifiers(identifiers []string) *DetectedTrackerFilter {
	f.identifiers = identifiers
	return f
}

// WithUpdatedSince restricts results to detections updated at or after
// since. A nil time disables the filter.
func (f *DetectedTrackerFilter) WithUpdatedSince(since *time.Time) *DetectedTrackerFilter {
	f.updatedSince = since
	return f
}

func (f *DetectedTrackerFilter) SQLFragment() string {
	return `
(
	CASE
		WHEN @filter_tracker_types::text[] IS NOT NULL THEN
			tracker_type::text = ANY(@filter_tracker_types::text[])
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_identifiers::text[] IS NULL THEN TRUE
		ELSE identifier = ANY(@filter_identifiers::text[])
	END
	AND
	CASE
		WHEN @filter_updated_since::timestamptz IS NULL THEN TRUE
		ELSE updated_at >= @filter_updated_since::timestamptz
	END
)`
}

func (f *DetectedTrackerFilter) SQLArguments() pgx.StrictNamedArgs {
	var (
		trackerTypes []TrackerType
		identifiers  []string
		updatedSince *time.Time
	)
	if f != nil {
		trackerTypes = f.trackerTypes
		identifiers = f.identifiers
		updatedSince = f.updatedSince
	}

	return pgx.StrictNamedArgs{
		"filter_tracker_types": filterStrings(trackerTypes),
		"filter_identifiers":   identifiers,
		"filter_updated_since": updatedSince,
	}
}

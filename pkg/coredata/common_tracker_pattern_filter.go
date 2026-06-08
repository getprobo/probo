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
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/gid"
)

// CommonTrackerPatternEnrichmentState is a synthetic filter over the
// enrichment_requested_at / enriched_at columns. It is not a stored
// column; it classifies a row's position in the enrichment lifecycle.
type CommonTrackerPatternEnrichmentState string

const (
	// CommonTrackerPatternEnrichmentStateQueued: a row armed for the
	// enrichment worker (enrichment_requested_at IS NOT NULL).
	CommonTrackerPatternEnrichmentStateQueued CommonTrackerPatternEnrichmentState = "QUEUED"
	// CommonTrackerPatternEnrichmentStateEnriched: a row whose
	// enrichment has completed (enriched_at IS NOT NULL) and is not
	// re-queued.
	CommonTrackerPatternEnrichmentStateEnriched CommonTrackerPatternEnrichmentState = "ENRICHED"
	// CommonTrackerPatternEnrichmentStateUnenriched: a row never enriched
	// and not currently queued.
	CommonTrackerPatternEnrichmentStateUnenriched CommonTrackerPatternEnrichmentState = "UNENRICHED"
)

func (s CommonTrackerPatternEnrichmentState) IsValid() bool {
	switch s {
	case
		CommonTrackerPatternEnrichmentStateQueued,
		CommonTrackerPatternEnrichmentStateEnriched,
		CommonTrackerPatternEnrichmentStateUnenriched:
		return true
	}

	return false
}

func (s CommonTrackerPatternEnrichmentState) String() string {
	return string(s)
}

func (s CommonTrackerPatternEnrichmentState) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *CommonTrackerPatternEnrichmentState) UnmarshalText(text []byte) error {
	val := CommonTrackerPatternEnrichmentState(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CommonTrackerPatternEnrichmentState value: %q", string(text))
	}

	*s = val

	return nil
}

type CommonTrackerPatternFilter struct {
	ids                []gid.GID
	trackerType        *TrackerType
	matchType          *TrackerPatternMatchType
	commonThirdPartyID *gid.GID
	keyword            *string
	linked             *bool
	described          *bool
	state              *CommonTrackerPatternEnrichmentState
}

func NewCommonTrackerPatternFilter() *CommonTrackerPatternFilter {
	return &CommonTrackerPatternFilter{}
}

// WithIDs restricts the result to the given pattern IDs. A non-nil but
// empty slice matches nothing.
func (f *CommonTrackerPatternFilter) WithIDs(ids []gid.GID) *CommonTrackerPatternFilter {
	f.ids = ids
	return f
}

func (f *CommonTrackerPatternFilter) WithTrackerType(trackerType *TrackerType) *CommonTrackerPatternFilter {
	f.trackerType = trackerType
	return f
}

func (f *CommonTrackerPatternFilter) WithMatchType(matchType *TrackerPatternMatchType) *CommonTrackerPatternFilter {
	f.matchType = matchType
	return f
}

func (f *CommonTrackerPatternFilter) WithCommonThirdPartyID(id *gid.GID) *CommonTrackerPatternFilter {
	f.commonThirdPartyID = id
	return f
}

func (f *CommonTrackerPatternFilter) WithKeyword(keyword *string) *CommonTrackerPatternFilter {
	f.keyword = keyword
	return f
}

func (f *CommonTrackerPatternFilter) WithLinked(linked *bool) *CommonTrackerPatternFilter {
	f.linked = linked
	return f
}

// WithDescribed filters on whether the pattern has a non-empty
// description: true keeps only described rows, false keeps only rows with
// a blank description.
func (f *CommonTrackerPatternFilter) WithDescribed(described *bool) *CommonTrackerPatternFilter {
	f.described = described
	return f
}

func (f *CommonTrackerPatternFilter) WithState(state *CommonTrackerPatternEnrichmentState) *CommonTrackerPatternFilter {
	f.state = state
	return f
}

func (f *CommonTrackerPatternFilter) SQLFragment() string {
	if f == nil {
		return "TRUE"
	}

	return `
(
	CASE
		WHEN @filter_ids::text[] IS NOT NULL THEN
			id = ANY(@filter_ids)
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_tracker_type::text IS NOT NULL THEN
			tracker_type = @filter_tracker_type::tracker_type
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_match_type::text IS NOT NULL THEN
			match_type = @filter_match_type::cookie_pattern_match_type
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_common_third_party_id::text IS NOT NULL THEN
			common_third_party_id = @filter_common_third_party_id::text
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_keyword::text IS NOT NULL AND @filter_keyword::text != '' THEN
			(pattern ILIKE '%' || @filter_keyword || '%'
			 OR description ILIKE '%' || @filter_keyword || '%')
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_linked::boolean IS NULL THEN TRUE
		WHEN @filter_linked::boolean THEN common_third_party_id IS NOT NULL
		ELSE common_third_party_id IS NULL
	END
	AND
	CASE
		WHEN @filter_described::boolean IS NULL THEN TRUE
		WHEN @filter_described::boolean THEN description != ''
		ELSE description = ''
	END
	AND
	CASE
		WHEN @filter_state_queued::boolean THEN enrichment_requested_at IS NOT NULL
		WHEN @filter_state_enriched::boolean THEN
			enrichment_requested_at IS NULL AND enriched_at IS NOT NULL
		WHEN @filter_state_unenriched::boolean THEN
			enrichment_requested_at IS NULL AND enriched_at IS NULL
		ELSE TRUE
	END
)`
}

func (f *CommonTrackerPatternFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"filter_ids":                   nil,
		"filter_tracker_type":          nil,
		"filter_match_type":            nil,
		"filter_common_third_party_id": nil,
		"filter_keyword":               nil,
		"filter_linked":                nil,
		"filter_described":             nil,
		"filter_state_queued":          false,
		"filter_state_enriched":        false,
		"filter_state_unenriched":      false,
	}

	if f == nil {
		return args
	}

	if f.ids != nil {
		args["filter_ids"] = f.ids
	}

	if f.trackerType != nil {
		args["filter_tracker_type"] = string(*f.trackerType)
	}

	if f.matchType != nil {
		args["filter_match_type"] = string(*f.matchType)
	}

	if f.commonThirdPartyID != nil {
		args["filter_common_third_party_id"] = *f.commonThirdPartyID
	}

	if f.keyword != nil {
		args["filter_keyword"] = *f.keyword
	}

	if f.linked != nil {
		args["filter_linked"] = *f.linked
	}

	if f.described != nil {
		args["filter_described"] = *f.described
	}

	if f.state != nil {
		switch *f.state {
		case CommonTrackerPatternEnrichmentStateQueued:
			args["filter_state_queued"] = true
		case CommonTrackerPatternEnrichmentStateEnriched:
			args["filter_state_enriched"] = true
		case CommonTrackerPatternEnrichmentStateUnenriched:
			args["filter_state_unenriched"] = true
		}
	}

	return args
}

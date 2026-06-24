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
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.probo.inc/probo/pkg/gid"
)

// CommonThirdPartyEnrichmentState is a synthetic filter over the
// enrichment_requested_at / enrichment columns. It is not a stored
// column; it classifies a row's position in the enrichment lifecycle. A
// row is "enriched" once it carries an enrichment payload (Process always
// writes one, even on a no-result run).
type CommonThirdPartyEnrichmentState string

const (
	// CommonThirdPartyEnrichmentStateQueued: a row armed for the
	// enrichment worker (enrichment_requested_at IS NOT NULL).
	CommonThirdPartyEnrichmentStateQueued CommonThirdPartyEnrichmentState = "QUEUED"
	// CommonThirdPartyEnrichmentStateEnriched: a row whose enrichment has
	// completed (enrichment IS NOT NULL) and is not re-queued.
	CommonThirdPartyEnrichmentStateEnriched CommonThirdPartyEnrichmentState = "ENRICHED"
	// CommonThirdPartyEnrichmentStateUnenriched: a row never enriched and
	// not currently queued.
	CommonThirdPartyEnrichmentStateUnenriched CommonThirdPartyEnrichmentState = "UNENRICHED"
)

func (s CommonThirdPartyEnrichmentState) IsValid() bool {
	switch s {
	case
		CommonThirdPartyEnrichmentStateQueued,
		CommonThirdPartyEnrichmentStateEnriched,
		CommonThirdPartyEnrichmentStateUnenriched:
		return true
	}

	return false
}

func (s CommonThirdPartyEnrichmentState) String() string {
	return string(s)
}

func (s CommonThirdPartyEnrichmentState) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s *CommonThirdPartyEnrichmentState) UnmarshalText(text []byte) error {
	val := CommonThirdPartyEnrichmentState(text)
	if !val.IsValid() {
		return fmt.Errorf("invalid CommonThirdPartyEnrichmentState value: %q", string(text))
	}

	*s = val

	return nil
}

type CommonThirdPartyFilter struct {
	ids              []gid.GID
	name             *string
	category         *ThirdPartyCategory
	keyword          *string
	state            *CommonThirdPartyEnrichmentState
	enrichmentStatus *string
}

func NewCommonThirdPartyFilter(name *string) *CommonThirdPartyFilter {
	return &CommonThirdPartyFilter{name: name}
}

// WithIDs restricts the result to the given common third party IDs. A
// non-nil but empty slice matches nothing.
func (f *CommonThirdPartyFilter) WithIDs(ids []gid.GID) *CommonThirdPartyFilter {
	f.ids = ids
	return f
}

func (f *CommonThirdPartyFilter) WithCategory(category *ThirdPartyCategory) *CommonThirdPartyFilter {
	f.category = category
	return f
}

func (f *CommonThirdPartyFilter) WithKeyword(keyword *string) *CommonThirdPartyFilter {
	f.keyword = keyword
	return f
}

func (f *CommonThirdPartyFilter) WithState(state *CommonThirdPartyEnrichmentState) *CommonThirdPartyFilter {
	f.state = state
	return f
}

// WithEnrichmentStatus filters on the run-level status recorded in the
// enrichment payload (done, partial, failed). Rows with no payload never
// match.
func (f *CommonThirdPartyFilter) WithEnrichmentStatus(status *string) *CommonThirdPartyFilter {
	f.enrichmentStatus = status
	return f
}

func (f *CommonThirdPartyFilter) SQLFragment() string {
	return `(
	CASE
		WHEN @filter_ids::text[] IS NOT NULL THEN
			id = ANY(@filter_ids)
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_name::text IS NOT NULL THEN
			name ILIKE '%' || @filter_name || '%'
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_category::text IS NOT NULL THEN
			category = @filter_category::third_party_category
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_keyword::text IS NOT NULL AND @filter_keyword::text != '' THEN
			(name ILIKE '%' || @filter_keyword || '%'
			 OR slug ILIKE '%' || @filter_keyword || '%')
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_state_queued::boolean THEN enrichment_requested_at IS NOT NULL
		WHEN @filter_state_enriched::boolean THEN
			enrichment_requested_at IS NULL AND enrichment IS NOT NULL
		WHEN @filter_state_unenriched::boolean THEN
			enrichment_requested_at IS NULL AND enrichment IS NULL
		ELSE TRUE
	END
	AND
	CASE
		WHEN @filter_enrichment_status::text IS NOT NULL THEN
			enrichment->>'status' = @filter_enrichment_status
		ELSE TRUE
	END
)`
}

func (f *CommonThirdPartyFilter) SQLArguments() pgx.StrictNamedArgs {
	args := pgx.StrictNamedArgs{
		"filter_ids":               nil,
		"filter_name":              nil,
		"filter_category":          nil,
		"filter_keyword":           nil,
		"filter_state_queued":      false,
		"filter_state_enriched":    false,
		"filter_state_unenriched":  false,
		"filter_enrichment_status": nil,
	}

	if f.ids != nil {
		args["filter_ids"] = f.ids
	}

	if f.name != nil {
		args["filter_name"] = *f.name
	}

	if f.category != nil {
		args["filter_category"] = string(*f.category)
	}

	if f.keyword != nil {
		args["filter_keyword"] = *f.keyword
	}

	if f.state != nil {
		switch *f.state {
		case CommonThirdPartyEnrichmentStateQueued:
			args["filter_state_queued"] = true
		case CommonThirdPartyEnrichmentStateEnriched:
			args["filter_state_enriched"] = true
		case CommonThirdPartyEnrichmentStateUnenriched:
			args["filter_state_unenriched"] = true
		}
	}

	if f.enrichmentStatus != nil {
		args["filter_enrichment_status"] = *f.enrichmentStatus
	}

	return args
}

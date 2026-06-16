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

package cookiebanner

import (
	"strings"
	"time"
)

// Per-field outcomes recorded in the common tracker pattern enrichment
// payload. They mirror the common-third-party enrichment provenance so the
// proboctl display can compute the same "X/Y resolved" completeness across
// both catalogs.
const (
	commonPatternFieldStatusFound    = "found"
	commonPatternFieldStatusNotFound = "not_found"
	commonPatternFieldStatusExternal = "exists_external"

	// Run-level status recorded at the top of the enrichment payload.
	commonPatternStatusDone     = "done"
	commonPatternStatusPartial  = "partial"
	commonPatternStatusNoResult = "no_result"

	// Field keys recorded in the payload (the enrichment targets).
	commonPatternFieldDescription = "description"
	commonPatternFieldThirdParty  = "third_party"
)

type (
	// CommonPatternFieldMeta is the per-field provenance recorded in the
	// common_tracker_patterns.enrichment JSON column.
	CommonPatternFieldMeta struct {
		Status    string    `json:"status"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	// CommonPatternAttributionMeta records the mapping-agent result the
	// enricher used to attribute (and possibly link) a vendor, so the
	// "mapping" decision is auditable from the enrichment payload.
	CommonPatternAttributionMeta struct {
		ThirdPartyName string  `json:"third_party_name,omitempty"`
		Category       string  `json:"category,omitempty"`
		Confidence     float64 `json:"confidence"`
		Linked         bool    `json:"linked"`
	}

	// CommonPatternEnrichmentMetadata is the full payload stored in the
	// enrichment JSON column: run-level bookkeeping, per-field provenance
	// keyed by the enrichment target, and the vendor attribution.
	CommonPatternEnrichmentMetadata struct {
		Model       string                            `json:"model,omitempty"`
		AttemptedAt time.Time                         `json:"attempted_at"`
		Status      string                            `json:"status"`
		Error       string                            `json:"error,omitempty"`
		Fields      map[string]CommonPatternFieldMeta `json:"fields"`
		Attribution *CommonPatternAttributionMeta     `json:"attribution,omitempty"`
	}
)

// buildCommonPatternEnrichmentMetadata assembles the per-run provenance for
// one common tracker pattern. It records an outcome for both enrichment
// targets — the description and the third-party attribution — so the
// display can report "X/Y resolved". alreadyLinked marks a vendor the
// mapping pipeline resolved before this run (recorded as exists_external),
// versus a vendor this enrichment run resolved (found). attribution, when
// non-nil, carries the mapping-agent decision this run made.
func buildCommonPatternEnrichmentMetadata(
	model string,
	description string,
	attribution *TrackerMappingAgentResult,
	alreadyLinked bool,
	linked bool,
	now time.Time,
) CommonPatternEnrichmentMetadata {
	fields := make(map[string]CommonPatternFieldMeta, 2)

	descStatus := commonPatternFieldStatusNotFound
	if strings.TrimSpace(description) != "" {
		descStatus = commonPatternFieldStatusFound
	}

	fields[commonPatternFieldDescription] = CommonPatternFieldMeta{
		Status:    descStatus,
		UpdatedAt: now,
	}

	thirdPartyStatus := commonPatternFieldStatusNotFound

	switch {
	case alreadyLinked:
		thirdPartyStatus = commonPatternFieldStatusExternal
	case linked:
		thirdPartyStatus = commonPatternFieldStatusFound
	}

	fields[commonPatternFieldThirdParty] = CommonPatternFieldMeta{
		Status:    thirdPartyStatus,
		UpdatedAt: now,
	}

	meta := CommonPatternEnrichmentMetadata{
		Model:       model,
		AttemptedAt: now,
		Status:      commonPatternRunStatus(fields),
		Fields:      fields,
	}

	if attribution != nil {
		meta.Attribution = &CommonPatternAttributionMeta{
			ThirdPartyName: attribution.ThirdPartyName,
			Category:       string(attribution.Category),
			Confidence:     attribution.ThirdPartyConfidence,
			Linked:         linked,
		}
	}

	return meta
}

// commonPatternRunStatus classifies the run from its per-field outcomes:
// done when every field resolved a value, no_result when none did, partial
// otherwise.
func commonPatternRunStatus(fields map[string]CommonPatternFieldMeta) string {
	var resolved int

	for _, f := range fields {
		if commonPatternFieldResolved(f.Status) {
			resolved++
		}
	}

	switch {
	case resolved == 0:
		return commonPatternStatusNoResult
	case resolved == len(fields):
		return commonPatternStatusDone
	default:
		return commonPatternStatusPartial
	}
}

// commonPatternFieldResolved reports whether a field status carries a
// resolved value (found or already present externally) versus an absent
// one.
func commonPatternFieldResolved(status string) bool {
	switch status {
	case commonPatternFieldStatusFound, commonPatternFieldStatusExternal:
		return true
	default:
		return false
	}
}

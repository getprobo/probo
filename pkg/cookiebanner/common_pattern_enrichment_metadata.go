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

	// commonPatternFieldStatusFirstParty records that the artifact was
	// determined to have no third party behind it. It is a definitive
	// answer, not a missing one, so it counts as resolved: a first-party
	// row would otherwise report no_result forever despite the question
	// being settled.
	commonPatternFieldStatusFirstParty = "first_party"

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

		// FirstParty records a terminal verdict that the artifact has no
		// third party behind it, so the decision stays auditable from the
		// payload after the vendor link has been cleared.
		FirstParty bool `json:"first_party,omitempty"`
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

// buildCommonPatternFirstPartyMetadata assembles the provenance for a run
// that ended in a terminal first-party verdict. Both enrichment targets
// are recorded as first_party rather than not_found: the artifact has no
// vendor to name and therefore no vendor-informed description to write, so
// the run resolved both questions rather than failing at them.
func buildCommonPatternFirstPartyMetadata(
	model string,
	now time.Time,
) CommonPatternEnrichmentMetadata {
	fields := map[string]CommonPatternFieldMeta{
		commonPatternFieldDescription: {
			Status:    commonPatternFieldStatusFirstParty,
			UpdatedAt: now,
		},
		commonPatternFieldThirdParty: {
			Status:    commonPatternFieldStatusFirstParty,
			UpdatedAt: now,
		},
	}

	return CommonPatternEnrichmentMetadata{
		Model:       model,
		AttemptedAt: now,
		Status:      commonPatternRunStatus(fields),
		Fields:      fields,
		Attribution: &CommonPatternAttributionMeta{FirstParty: true},
	}
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
	case commonPatternFieldStatusFound,
		commonPatternFieldStatusExternal,
		commonPatternFieldStatusFirstParty:
		return true
	default:
		return false
	}
}

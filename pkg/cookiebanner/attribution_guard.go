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

import "strings"

// attributionContext is the per-call context the guards need beyond the
// agent result.
type attributionContext struct {
	// Pattern is the artifact under judgement, for log context only.
	Pattern string

	// SiteOrigin is the scanned site's origin, or nil when the artifact
	// belongs to no site — a catalog pattern — which makes the scanned-site
	// backstop inapplicable rather than comparing against an empty string.
	SiteOrigin *string
}

// attributionRejection names the guard that discarded an attribution, or
// attributionAccepted when all passed. Named rather than a bool so both call
// sites log one reason vocabulary and tests assert which guard fired.
type attributionRejection string

const (
	attributionAccepted            attributionRejection = ""
	attributionRejectedConfidence  attributionRejection = "below_confidence_threshold"
	attributionRejectedNoEvidence  attributionRejection = "no_concrete_evidence"
	attributionRejectedScannedSite attributionRejection = "scanned_site_as_third_party"
	attributionRejectedAggregator  attributionRejection = "cookie_database_aggregator"
)

// rejectVendorAttribution is the single acceptance bar for creating a common
// third party. Both the tracker mapping worker and the common pattern
// enricher call it, so an attribution one would discard can no longer enter
// the global catalog through the other.
//
// Every rejection leaves the artifact undetermined rather than terminal. These
// guards judge the proposed NAME, not the artifact, so a wrong name on a
// genuine vendor pattern must fall through for a later attempt: a terminal
// verdict would suppress the real vendor permanently. Callers that also got an
// explicit terminal verdict from the agent record it themselves.
//
// No I/O and no logging: callers log with their own context.
func rejectVendorAttribution(
	identification TrackerMappingAgentResult,
	actx attributionContext,
) attributionRejection {
	// Trimmed here so both call sites agree on what an empty name is: the
	// mapping worker forwards the agent's output verbatim.
	name := strings.TrimSpace(identification.ThirdPartyName)

	if name == "" || identification.ThirdPartyConfidence < agentThirdPartyConfidenceThreshold {
		return attributionRejectedConfidence
	}

	// No evidence source means the agent named a vendor from general
	// knowledge, so a wrong precedent never enters the catalog.
	if !evidenceSupportsAttribution(identification.EvidenceSource) {
		return attributionRejectedNoEvidence
	}

	if actx.SiteOrigin != nil && nameMatchesSiteDomain(name, *actx.SiteOrigin) {
		return attributionRejectedScannedSite
	}

	if nameIsCookieDatabaseAggregator(name) {
		return attributionRejectedAggregator
	}

	return attributionAccepted
}

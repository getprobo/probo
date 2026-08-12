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

// attributionContext carries the per-call context the attribution guards
// need beyond the agent result itself.
//
// SiteOrigin is optional. The tracker mapping worker judges a pattern
// observed on a concrete scanned site and supplies it, which arms the
// scanned-site backstop. The common pattern enricher judges a global
// catalog pattern that belongs to no site, so it leaves SiteOrigin nil and
// that one guard is inapplicable — expressed as a missing value rather
// than an empty string that silently compares against nothing.
type attributionContext struct {
	// Pattern is the artifact under judgement, for log context only.
	Pattern string

	// SiteOrigin is the scanned site's origin, or nil when the artifact
	// belongs to no particular site.
	SiteOrigin *string
}

// attributionRejection names the guard that discarded a vendor
// attribution, or attributionAccepted when every guard passed. It is
// returned instead of a bool so both call sites log one reason vocabulary
// and tests can assert which guard fired rather than merely that one did.
type attributionRejection string

const (
	attributionAccepted              attributionRejection = ""
	attributionRejectedConfidence    attributionRejection = "below_confidence_threshold"
	attributionRejectedNoEvidence    attributionRejection = "no_concrete_evidence"
	attributionRejectedScannedSite   attributionRejection = "scanned_site_as_third_party"
	attributionRejectedAggregator    attributionRejection = "cookie_database_aggregator"
	attributionRejectedFirstPartyLib attributionRejection = "bundled_first_party_library"
)

// rejectVendorAttribution applies every catalog-write guard to an agent
// attribution and reports which one rejected it, or attributionAccepted.
//
// It is the single acceptance bar for creating a common third party. Both
// the tracker mapping worker and the common pattern enricher call it, so
// an attribution one would discard can no longer enter the global catalog
// through the other. Guards run cheapest-and-most-decisive first: a
// confident name, then a concrete evidence source, then the three
// name-shape backstops.
//
// Every rejection leaves the artifact undetermined rather than recording a
// first-party verdict. These guards judge the proposed vendor NAME, not
// the artifact, so a hallucinated library or aggregator name on a genuine
// vendor pattern must fall through for a later, better-informed attempt —
// a first-party verdict is terminal and would suppress the real vendor
// permanently. Callers that also received an explicit first-party verdict
// from the agent record it themselves.
//
// It performs no I/O and no logging: callers log with their own context.
func rejectVendorAttribution(
	identification TrackerMappingAgentResult,
	actx attributionContext,
) attributionRejection {
	// Trim here so both call sites agree on what an empty name is: the
	// mapping worker forwards the agent's output verbatim.
	name := strings.TrimSpace(identification.ThirdPartyName)

	// The agent's confidence gauges the attribution (who set the tracker),
	// not whether the artifact is a meaningful tracker. Without a
	// confident vendor there is nothing to catalog.
	if name == "" || identification.ThirdPartyConfidence < agentThirdPartyConfidenceThreshold {
		return attributionRejectedConfidence
	}

	// A vendor is attributed only on concrete evidence (a database match,
	// a meaningful naming convention, or a web/browser result naming the
	// setter). An attribution with no evidence source is a
	// general-knowledge guess, so a wrong precedent never enters the
	// catalog.
	if !evidenceSupportsAttribution(identification.EvidenceSource) {
		return attributionRejectedNoEvidence
	}

	// Backstop for the prompt rule that the scanned site is never a third
	// party of itself: a pattern embedding the site's own domain (a wallet
	// extension key, an owner-set tracker) can lead the agent to attribute
	// the site's own brand.
	if actx.SiteOrigin != nil && nameMatchesSiteDomain(name, *actx.SiteOrigin) {
		return attributionRejectedScannedSite
	}

	// Cookie-database directory sites rank highly in web search only
	// because they catalog cookies, not because they set them.
	if nameIsCookieDatabaseAggregator(name) {
		return attributionRejectedAggregator
	}

	// Bundled client-side libraries write local keys but egress nothing to
	// any endpoint their maintainers operate, so they are the site's own
	// state and never a catalog third party. The prompt's egress test is
	// the primary defence; this catches the cases where the agent names
	// the library that laid the key instead of recognising there is no
	// data recipient.
	if nameIsBundledFirstPartyLibrary(name) {
		return attributionRejectedFirstPartyLib
	}

	return attributionAccepted
}

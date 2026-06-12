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

package thirdparty

import (
	"slices"
	"strings"

	"go.probo.inc/probo/pkg/stringsx"
	"go.probo.inc/probo/pkg/uri"
)

// minLabelOverlap is the shortest label length allowed for a substring
// ownership match. Shorter labels (e.g. "id", "go") match too many
// unrelated domains to be a trustworthy ownership signal, so they only
// match on exact equality.
const minLabelOverlap = 4

// ownedDomain is a vendor-owned eTLD+1 that cleared the ownership gate,
// carrying the provenance the worker records in the enrichment payload.
type ownedDomain struct {
	Domain     string
	Confidence float64
	SourceURL  string
}

// resolveOwnedDomains reduces the domain-discovery agent's candidates to
// the eTLD+1 domains the vendor demonstrably owns, ready to write to
// common_third_party_domains. The vendor's own website eTLD+1 is always
// included. Every other candidate must clear the confidence threshold and
// be ownership-related to the vendor by a label match against the website
// label or the normalized vendor name.
//
// Shared tracker-delivery / CDN infrastructure (uri.FilterSharedInfra...)
// is dropped only when it is NOT owned: a generic provider domain cannot
// be attributed to an unrelated vendor, but when the vendor itself is
// that provider (e.g. enriching Cloudflare or Fastly) its own domain
// passes a stricter exact-label match and is kept. Results are
// deduplicated with insertion order preserved.
func resolveOwnedDomains(
	name string,
	website string,
	result DomainsResult,
	threshold float64,
) []ownedDomain {
	vendorLabels := vendorLabels(name, website)

	var (
		owned []ownedDomain
		seen  = make(map[string]struct{})
	)

	add := func(domain string, confidence float64, sourceURL string) {
		domain = strings.ToLower(strings.TrimSpace(domain))
		if domain == "" {
			return
		}

		if _, ok := seen[domain]; ok {
			return
		}

		seen[domain] = struct{}{}

		owned = append(owned, ownedDomain{
			Domain:     domain,
			Confidence: confidence,
			SourceURL:  sourceURL,
		})
	}

	// The vendor's own website domain is owned by definition.
	if site := normalizeToETLD1(website); site != "" {
		add(site, 1.0, strings.TrimSpace(website))
	}

	for _, candidate := range result.Domains {
		domain := normalizeToETLD1(candidate.Domain)
		if domain == "" {
			continue
		}

		if !isVendorOwnedDomain(domain, candidate.Confidence, threshold, vendorLabels) {
			continue
		}

		add(domain, candidate.Confidence, strings.TrimSpace(candidate.SourceURL))
	}

	return owned
}

// isVendorOwnedDomain reports whether an eTLD+1 candidate clears the
// confidence threshold and is owned by the vendor. Shared-infrastructure
// domains require an exact label match (the vendor must be that provider);
// all other domains accept a substring label overlap.
func isVendorOwnedDomain(
	domain string,
	confidence float64,
	threshold float64,
	vendorLabels []string,
) bool {
	if confidence < threshold {
		return false
	}

	label := uri.DomainLabel(domain)
	if label == "" {
		return false
	}

	if isSharedInfrastructureDomain(domain) {
		return exactLabelMatch(label, vendorLabels)
	}

	return relatedLabelMatch(label, vendorLabels)
}

// isSharedInfrastructureDomain reports whether an eTLD+1 belongs to the
// shared tracker-delivery / CDN infrastructure list, reusing the curated
// set maintained in the uri package.
func isSharedInfrastructureDomain(domain string) bool {
	return len(uri.FilterSharedInfrastructureDomains([]string{domain})) == 0
}

// vendorLabels returns the lowercase ownership labels for a vendor: the
// registrable label of its website and its name reduced to alphanumerics.
func vendorLabels(name, website string) []string {
	var (
		labels []string
		seen   = make(map[string]struct{})
	)

	add := func(label string) {
		if label == "" {
			return
		}

		if _, ok := seen[label]; ok {
			return
		}

		seen[label] = struct{}{}

		labels = append(labels, label)
	}

	add(uri.DomainLabel(website))
	add(stringsx.NormalizeAlnum(name))

	return labels
}

// exactLabelMatch reports whether label equals any vendor label.
func exactLabelMatch(label string, vendorLabels []string) bool {
	return slices.Contains(vendorLabels, label)
}

// relatedLabelMatch reports whether label is the same as, or shares a
// dominant root with, any vendor label. A substring match requires the
// shorter label to be at least minLabelOverlap characters AND to cover at
// least half of the longer label, so a short root (e.g. "meta") no longer
// attributes an unrelated domain (e.g. "metallica") to the vendor.
func relatedLabelMatch(label string, vendorLabels []string) bool {
	for _, vl := range vendorLabels {
		if label == vl {
			return true
		}

		shorter, longer := label, vl
		if len(longer) < len(shorter) {
			shorter, longer = longer, shorter
		}

		if len(shorter) < minLabelOverlap {
			continue
		}

		if len(shorter)*2 < len(longer) {
			continue
		}

		if strings.Contains(longer, shorter) {
			return true
		}
	}

	return false
}

// normalizeToETLD1 reduces a bare domain or full URL to its eTLD+1,
// tolerating candidates the agent returns without a scheme.
func normalizeToETLD1(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}

	return uri.ExtractDomain(raw)
}

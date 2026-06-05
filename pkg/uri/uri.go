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

package uri

import (
	"database/sql/driver"
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// URI is a validated absolute URI (scheme + host required).
type URI string

func Parse(raw string) (URI, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("%q is not a valid URI", raw)
	}

	return URI(raw), nil
}

func (u URI) String() string { return string(u) }

func (u *URI) UnmarshalText(text []byte) error {
	parsed, err := Parse(string(text))
	if err != nil {
		return err
	}

	*u = parsed

	return nil
}

func (u URI) MarshalText() ([]byte, error) {
	return []byte(u), nil
}

func (u *URI) Scan(value any) error {
	var s string

	switch v := value.(type) {
	case string:
		s = v
	case []byte:
		s = string(v)
	default:
		return fmt.Errorf("unsupported type for URI: %T", value)
	}

	parsed, err := Parse(s)
	if err != nil {
		return err
	}

	*u = parsed

	return nil
}

func (u URI) Value() (driver.Value, error) {
	return u.String(), nil
}

// ExtractDomain returns the eTLD+1 (effective top-level domain plus one
// label) from a raw URL string. For example:
//
//	"https://www.googletagmanager.com/gtag/js" → "googletagmanager.com"
//	"https://cdn.segment.io/v1/projects"       → "segment.io"
//
// Returns an empty string when the URL cannot be parsed or has no valid
// hostname (e.g. data: URIs, bare IP addresses without a public suffix).
func ExtractDomain(rawURL string) string {
	u, err := Parse(rawURL)
	if err != nil {
		return ""
	}

	parsed, _ := url.Parse(string(u))
	hostname := strings.ToLower(parsed.Hostname())

	domain, err := publicsuffix.EffectiveTLDPlusOne(hostname)
	if err != nil {
		return ""
	}

	return domain
}

// FilterFirstPartyDomains removes domains that match the eTLD+1 of
// siteOrigin. Tracker scripts loaded through a first-party proxy (e.g.
// t.probo.com proxying PostHog on a probo.com site) share the site's
// eTLD+1 and carry no signal about the actual third party. siteOrigin
// is a full URL such as "https://app.probo.com". The input domains are
// expected to be eTLD+1 strings (as produced by ExtractDomain).
func FilterFirstPartyDomains(domains []string, siteOrigin string) []string {
	siteDomain := ExtractDomain(siteOrigin)
	if siteDomain == "" {
		return domains
	}

	filtered := make([]string, 0, len(domains))

	for _, d := range domains {
		if d != siteDomain {
			filtered = append(filtered, d)
		}
	}

	return filtered
}

// sharedInfrastructureDomains are eTLD+1 hosts that deliver trackers on
// behalf of many unrelated vendors: tag managers, customer-data
// platforms, and generic CDNs / static-asset / app-hosting domains. A
// tracker whose initiator domain is one of these tells us nothing about
// which vendor set it (e.g. a Meta pixel and a LinkedIn tag both loaded
// through Google Tag Manager would otherwise look like the same third
// party), so domain-overlap heuristics must ignore them.
//
// This set is the tuning surface for that exclusion. Two rules keep it
// safe to extend:
//
//   - Only add domains that serve content for many *unrelated* vendors.
//     Omit vendor-specific domains such as google-analytics.com: those
//     are a legitimate same-vendor signal even though Google also runs
//     Tag Manager and gstatic.
//   - Entries are eTLD+1, so they must not collapse onto a vendor. We
//     skip cloudflare.com for this reason (cdnjs.cloudflare.com shares
//     its eTLD+1 with Cloudflare-the-vendor's own properties).
//
// For exhaustive, maintained coverage this would ideally be sourced from
// a community dataset (DuckDuckGo Tracker Radar's `cnames`/CDN entries or
// Disconnect's services list) vendored into the repo, rather than hand
// curated. This list covers the common offenders without that dependency;
// excluding a domain only ever makes grouping more conservative, so
// over-inclusion is the safe failure mode.
var sharedInfrastructureDomains = map[string]struct{}{
	// Tag managers, customer-data platforms, and tag delivery.
	"googletagmanager.com": {},
	"segment.io":           {},
	"segment.com":          {},
	"tealium.com":          {},
	"tiqcdn.com":           {},
	"ensighten.com":        {},
	"adobedtm.com":         {},
	"mparticle.com":        {},
	"rudderlabs.com":       {},
	"rudderstack.com":      {},
	"tagcommander.com":     {},
	"commander1.com":       {},

	// Commercial CDNs and edge networks.
	"cloudfront.net":   {},
	"akamai.net":       {},
	"akamaihd.net":     {},
	"akamaized.net":    {},
	"akamaiedge.net":   {},
	"edgekey.net":      {},
	"edgesuite.net":    {},
	"fastly.net":       {},
	"fastlylb.net":     {},
	"azureedge.net":    {},
	"azurefd.net":      {},
	"edgecastcdn.net":  {},
	"llnwd.net":        {},
	"hwcdn.net":        {},
	"cachefly.net":     {},
	"stackpathdns.com": {},
	"stackpathcdn.com": {},
	"netdna-cdn.com":   {},
	"netdna-ssl.com":   {},
	"kxcdn.com":        {},
	"b-cdn.net":        {},

	// Library, package, and static-asset CDNs.
	"jsdelivr.net":     {},
	"unpkg.com":        {},
	"bootstrapcdn.com": {},
	"maxcdn.com":       {},
	"cdnjs.com":        {},
	"jquery.com":       {},
	"aspnetcdn.com":    {},
	"skypack.dev":      {},
	"esm.sh":           {},
	"googleapis.com":   {},
	"gstatic.com":      {},

	// Font and media CDNs.
	"typekit.net":     {},
	"fontawesome.com": {},
	"cloudinary.com":  {},
	"imgix.net":       {},

	// Object storage and generic app / static hosting.
	"amazonaws.com":         {},
	"github.io":             {},
	"githubusercontent.com": {},
	"herokuapp.com":         {},
	"vercel.app":            {},
	"netlify.app":           {},
	"pages.dev":             {},
	"web.app":               {},
	"firebaseapp.com":       {},
	"wp.com":                {},
}

// FilterSharedInfrastructureDomains removes eTLD+1 domains that belong to
// shared tracker-delivery infrastructure (tag managers, customer-data
// platforms, and generic CDNs). Such domains initiate trackers for many
// unrelated vendors, so a shared initiator domain among them is not a
// same-vendor signal and must not drive domain-overlap grouping. The
// input domains are expected to be eTLD+1 strings (as produced by
// ExtractDomain).
func FilterSharedInfrastructureDomains(domains []string) []string {
	filtered := make([]string, 0, len(domains))

	for _, d := range domains {
		if _, shared := sharedInfrastructureDomains[strings.ToLower(d)]; !shared {
			filtered = append(filtered, d)
		}
	}

	return filtered
}

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

package complianceportal_v1

import (
	"net/http"
	"net/url"
	"slices"
	"strings"

	"go.probo.inc/probo/pkg/iam"
)

const defaultCompliancePortalLocale = "en"

// SEOFromRequest derives html lang, a self-referencing canonical URL, and
// hreflang alternates (including x-default → English) for the SPA shell.
// Portals are host-routed (slug subdomain / custom domain); the request path
// is already relative to the portal root.
func SEOFromRequest(r *http.Request, pageBaseURL string) (htmlLang, canonical string, hreflang []HreflangLink) {
	pathname := r.URL.Path
	if pathname == "" {
		pathname = "/"
	}

	locale, rest := splitLocaleFromAppPath(pathname)

	htmlLang = locale
	canonical = localizedPageURL(pageBaseURL, locale, rest)

	locales := iam.SupportedIdentityLocales

	hreflang = make([]HreflangLink, 0, len(locales)+1)
	for _, loc := range locales {
		hreflang = append(hreflang, HreflangLink{
			Lang: loc,
			Href: localizedPageURL(pageBaseURL, loc, rest),
		})
	}

	hreflang = append(hreflang, HreflangLink{
		Lang: "x-default",
		Href: localizedPageURL(pageBaseURL, defaultCompliancePortalLocale, rest),
	})

	return htmlLang, canonical, hreflang
}

func splitLocaleFromAppPath(appPath string) (locale, rest string) {
	segments := strings.Split(strings.Trim(appPath, "/"), "/")
	if len(segments) == 0 || segments[0] == "" {
		return defaultCompliancePortalLocale, "/"
	}

	if isCompliancePortalLocale(segments[0]) {
		locale = segments[0]
		if len(segments) == 1 {
			return locale, "/"
		}

		return locale, "/" + strings.Join(segments[1:], "/")
	}

	// Unprefixed path — treat content path as-is; default lang for tags.
	return defaultCompliancePortalLocale, appPath
}

func isCompliancePortalLocale(value string) bool {
	return slices.Contains(iam.SupportedIdentityLocales, value)
}

func localizedPageURL(pageBaseURL, locale, rest string) string {
	base := strings.TrimRight(pageBaseURL, "/")
	segments := []string{locale}

	if rest != "/" && rest != "" {
		trimmed := strings.Trim(rest, "/")
		if trimmed != "" {
			segments = append(segments, strings.Split(trimmed, "/")...)
		}
	}

	escaped := make([]string, len(segments))
	for i, segment := range segments {
		escaped[i] = url.PathEscape(segment)
	}

	joined, err := url.JoinPath(base, escaped...)
	if err != nil {
		return base + "/" + strings.Join(escaped, "/")
	}

	return joined
}

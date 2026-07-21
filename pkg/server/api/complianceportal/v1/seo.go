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
	"strings"
)

// Short locale tags used in compliance-portal URL paths. Keep in sync with the
// frontend URL_LOCALES list and iam.supportedIdentityLocales.
var compliancePortalLocales = []string{
	"en", "fr", "de", "es", "id", "it", "ja", "ko", "pl", "pt", "tr", "uk", "zh",
}

const defaultCompliancePortalLocale = "en"

// SEOFromRequest derives html lang, a self-referencing canonical URL, and
// hreflang alternates (including x-default → English) for the SPA shell.
func SEOFromRequest(r *http.Request, pageBaseURL string) (htmlLang, canonical string, hreflang []HreflangLink) {
	appPath := complianceAppPath(r.URL.Path)
	locale, rest := splitLocaleFromAppPath(appPath)

	htmlLang = locale
	canonical = localizedPageURL(pageBaseURL, locale, rest)

	hreflang = make([]HreflangLink, 0, len(compliancePortalLocales)+1)
	for _, loc := range compliancePortalLocales {
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

// complianceAppPath returns the path relative to the portal root: under
// /trust/:slug it strips that prefix; on a custom domain it returns the path as-is.
func complianceAppPath(pathname string) string {
	trimmed := strings.TrimPrefix(pathname, "/")
	if strings.HasPrefix(trimmed, "trust/") {
		parts := strings.SplitN(trimmed, "/", 3)
		if len(parts) < 2 {
			return "/"
		}
		if len(parts) == 2 {
			return "/"
		}
		return "/" + parts[2]
	}
	if pathname == "" {
		return "/"
	}
	return pathname
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

	// Unprefixed legacy path — treat content path as-is; default lang for tags.
	return defaultCompliancePortalLocale, appPath
}

func isCompliancePortalLocale(value string) bool {
	for _, locale := range compliancePortalLocales {
		if locale == value {
			return true
		}
	}
	return false
}

func localizedPageURL(pageBaseURL, locale, rest string) string {
	base := strings.TrimRight(pageBaseURL, "/")
	if rest == "/" || rest == "" {
		return base + "/" + locale
	}
	if !strings.HasPrefix(rest, "/") {
		rest = "/" + rest
	}
	return base + "/" + locale + rest
}

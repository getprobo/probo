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

package connector

import (
	"fmt"
	"net/url"
)

const DatadogProvider = "DATADOG"

// datadogSite describes one Datadog site's web (authorization) host and its
// API domain. The API/token host is always api.<apiDomain>.
type datadogSite struct {
	appHost   string
	apiDomain string
}

// datadogSites is the fixed, exhaustive allow-list of Datadog sites. Every
// site/domain that reaches URL construction MUST be validated against this
// table — both `site` (operator-influenced, at initiate) and `domain`
// (provider-supplied on the callback) feed URL hosts, so an unvetted value
// would be an SSRF vector. Read-only after initialization (effectively
// constant).
var datadogSites = map[string]datadogSite{
	"US1":     {appHost: "app.datadoghq.com", apiDomain: "datadoghq.com"},
	"US3":     {appHost: "us3.datadoghq.com", apiDomain: "us3.datadoghq.com"},
	"US5":     {appHost: "us5.datadoghq.com", apiDomain: "us5.datadoghq.com"},
	"EU1":     {appHost: "app.datadoghq.eu", apiDomain: "datadoghq.eu"},
	"AP1":     {appHost: "ap1.datadoghq.com", apiDomain: "ap1.datadoghq.com"},
	"AP2":     {appHost: "ap2.datadoghq.com", apiDomain: "ap2.datadoghq.com"},
	"US1-FED": {appHost: "app.ddog-gov.com", apiDomain: "ddog-gov.com"},
}

// DatadogAuthorizeURL returns the OAuth2 authorize endpoint for a Datadog
// site key (e.g. "US3"). It errors on any site not in the allow-list.
func DatadogAuthorizeURL(site string) (string, error) {
	s, ok := datadogSites[site]
	if !ok {
		return "", fmt.Errorf("cannot build authorize URL: unknown datadog site")
	}

	u := url.URL{Scheme: "https", Host: s.appHost, Path: "/oauth2/v1/authorize"}

	return u.String(), nil
}

// DatadogTokenURL returns the OAuth2 token endpoint for a Datadog API domain
// (e.g. "us3.datadoghq.com"). It errors on any domain not in the allow-list.
func DatadogTokenURL(domain string) (string, error) {
	if !IsValidDatadogDomain(domain) {
		return "", fmt.Errorf("cannot build token URL: unknown datadog domain")
	}

	u := url.URL{Scheme: "https", Host: "api." + domain, Path: "/oauth2/v1/token"}

	return u.String(), nil
}

// IsValidDatadogDomain reports whether domain is one of Datadog's known API
// domains.
func IsValidDatadogDomain(domain string) bool {
	for _, s := range datadogSites {
		if s.apiDomain == domain {
			return true
		}
	}

	return false
}

// DatadogSiteForDomain reverse-maps a Datadog API domain to its site key.
func DatadogSiteForDomain(domain string) (string, bool) {
	for key, s := range datadogSites {
		if s.apiDomain == domain {
			return key, true
		}
	}

	return "", false
}

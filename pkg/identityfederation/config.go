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

package identityfederation

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/gid"
)

const (
	// PathPrefix is where probod serves the route tree in every deployment. The
	// SaaS edge maps its public apex onto it, so only the advertised issuer
	// differs and it comes from configuration.
	PathPrefix = "/federation"

	// maxIssuerURLLength is the ceiling AWS enforces on an OIDC provider URL.
	maxIssuerURLLength = 255
)

// ResolveIssuerBaseURL returns the issuer base URL to advertise. An empty
// configured value derives {appBaseURL}{PathPrefix}, so self-hosted needs no
// second domain.
//
// The AWS rules apply only to an explicitly configured non-loopback base. A
// derived base is exempt so that http, ports and localhost still start;
// ValidateConfig reports those instead.
func ResolveIssuerBaseURL(
	configured string,
	appBaseURL *baseurl.BaseURL,
) (*baseurl.BaseURL, error) {
	if appBaseURL == nil {
		return nil, fmt.Errorf("cannot resolve identity federation issuer base URL: application base URL is required")
	}

	raw := configured
	explicit := configured != ""

	if !explicit {
		derived, err := url.JoinPath(appBaseURL.String(), PathPrefix)
		if err != nil {
			return nil, fmt.Errorf("cannot derive identity federation issuer base URL: %w", err)
		}

		raw = derived
	}

	issuerBaseURL, err := baseurl.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("cannot parse identity federation issuer base URL: %w", err)
	}

	if err := ValidateIssuerBaseURL(issuerBaseURL, appBaseURL); err != nil {
		return nil, err
	}

	if explicit && !isLoopbackHostname(issuerBaseURL.Hostname()) {
		if err := ValidateAWSIssuerBaseURL(issuerBaseURL); err != nil {
			return nil, err
		}
	}

	return issuerBaseURL, nil
}

// ValidateConfig reports whether this issuer can be registered with one or more
// cloud providers, so callers can surface a deployment that will never federate.
// Loopback returns nil: it is never registered.
//
// The caller decides the severity. probod only warns, because a derived base is
// deliberately exempt from the provider rules so that localhost and CI start.
// Without it, an unusable issuer would surface only when a customer installs the
// connector, long after it is pinned in their cloud resources.
//
// It applies AWS's rules, the strictest of the providers: GCP accepts an
// uploaded JWKS, so a private issuer can still federate there.
func ValidateConfig(issuerBaseURL *baseurl.BaseURL) error {
	if issuerBaseURL == nil || isLoopbackHostname(issuerBaseURL.Hostname()) {
		return nil
	}

	return ValidateAWSIssuerBaseURL(issuerBaseURL)
}

// ValidateIssuerBaseURL enforces the rules that hold in every deployment.
func ValidateIssuerBaseURL(issuerBaseURL, appBaseURL *baseurl.BaseURL) error {
	if issuerBaseURL == nil {
		return fmt.Errorf("cannot validate identity federation issuer base URL: issuer base URL is required")
	}

	if appBaseURL == nil {
		return fmt.Errorf("cannot validate identity federation issuer base URL: application base URL is required")
	}

	parsed, err := url.Parse(issuerBaseURL.String())
	if err != nil {
		return fmt.Errorf("cannot validate identity federation issuer base URL: %w", err)
	}

	// A bare "?" leaves RawQuery empty and sets ForceQuery instead, and it
	// survives url.JoinPath into the per-organization issuer, so refuse both.
	if parsed.RawQuery != "" || parsed.ForceQuery {
		return fmt.Errorf("cannot validate identity federation issuer base URL: query string is not allowed")
	}

	if parsed.Fragment != "" {
		return fmt.Errorf("cannot validate identity federation issuer base URL: fragment is not allowed")
	}

	// The issuer is a public identifier: it is logged at startup, served in the
	// discovery document and minted into the "iss" claim. Credentials in it
	// would leak everywhere it appears, and no provider would accept them.
	if parsed.User != nil {
		return fmt.Errorf("cannot validate identity federation issuer base URL: userinfo is not allowed")
	}

	// The ceiling applies to the issuer a customer registers, not to the base.
	// Measure the serialized URL: summing the parts misses the percent-encoding
	// and the collapsed trailing slash that url.JoinPath applies. Every GID
	// encodes to the same width, so any GID measures the real thing.
	representativeIssuer, err := issuerURL(issuerBaseURL, gid.Nil)
	if err != nil {
		return fmt.Errorf("cannot validate identity federation issuer base URL length: %w", err)
	}

	if len(representativeIssuer) > maxIssuerURLLength {
		return fmt.Errorf(
			"cannot validate identity federation issuer base URL: per-organization issuer would be %d characters, maximum is %d",
			len(representativeIssuer),
			maxIssuerURLLength,
		)
	}

	// probod matches on path only, so an issuer on the application host is
	// reachable at exactly one path. Any other path serves the console SPA's
	// HTML, and a root path serves the OAuth2 server's discovery document with
	// the wrong issuer and key set. Another host is unconstrained: it arrives
	// through an edge that maps its apex onto PathPrefix.
	//
	// DNS is case-insensitive, so casing must not bypass this.
	if !strings.EqualFold(issuerBaseURL.Hostname(), appBaseURL.Hostname()) {
		return nil
	}

	servedPath, err := servedIssuerPath(appBaseURL)
	if err != nil {
		return fmt.Errorf("cannot validate identity federation issuer base URL: %w", err)
	}

	if strings.TrimSuffix(parsed.Path, "/") != servedPath {
		return fmt.Errorf(
			"cannot validate identity federation issuer base URL: issuer on the application host must use the %q path, got %q",
			servedPath,
			parsed.Path,
		)
	}

	return nil
}

// servedIssuerPath returns the only path on the application host where probod
// serves the route tree. Derived like the default issuer base, so an application
// base URL that already carries a path keeps working.
func servedIssuerPath(appBaseURL *baseurl.BaseURL) (string, error) {
	derived, err := url.JoinPath(appBaseURL.String(), PathPrefix)
	if err != nil {
		return "", fmt.Errorf("cannot derive the served identity federation path: %w", err)
	}

	parsed, err := url.Parse(derived)
	if err != nil {
		return "", fmt.Errorf("cannot parse the served identity federation path: %w", err)
	}

	return strings.TrimSuffix(parsed.Path, "/"), nil
}

// ValidateAWSIssuerBaseURL enforces what AWS requires of an OIDC provider URL.
// AWS compares the URL case-sensitively and rejects a port or a query string.
func ValidateAWSIssuerBaseURL(issuerBaseURL *baseurl.BaseURL) error {
	if issuerBaseURL == nil {
		return fmt.Errorf("cannot validate identity federation issuer base URL: issuer base URL is required")
	}

	if issuerBaseURL.Scheme() != "https" {
		return fmt.Errorf(
			"cannot validate identity federation issuer base URL: scheme must be https, got %q",
			issuerBaseURL.Scheme(),
		)
	}

	if issuerBaseURL.Port() != "" {
		return fmt.Errorf("cannot validate identity federation issuer base URL: port is not allowed")
	}

	return nil
}

func isLoopbackHostname(hostname string) bool {
	if hostname == "localhost" || strings.HasSuffix(hostname, ".localhost") {
		return true
	}

	if ip := net.ParseIP(hostname); ip != nil {
		return ip.IsLoopback()
	}

	return false
}

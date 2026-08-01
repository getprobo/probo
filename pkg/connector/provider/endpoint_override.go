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

package provider

import (
	"fmt"
	"net/url"
	"strings"

	"go.probo.inc/probo/pkg/coredata"
)

// EndpointOverrides maps a provider to the endpoint values a deployment
// substitutes for the ones compiled into its Registration, so a non-production
// deployment can point a connector at the provider's sandbox (DocuSign's
// account-d host, for instance) without a code change.
//
// Only non-empty fields override; an empty field leaves the compiled default
// in place.
type EndpointOverrides map[coredata.ConnectorProvider]Endpoints

// Option configures NewBuiltinRegistry.
type Option func(*registryOptions)

type registryOptions struct {
	endpoints EndpointOverrides
}

// WithEndpointOverrides substitutes deployment-supplied endpoints for the
// compiled defaults. Overrides are applied and validated before the
// Registration reaches Register, so every invariant Register enforces — in
// particular that Probe shares a host with APIBase and with Identity — holds
// for the overridden values too.
func WithEndpointOverrides(overrides EndpointOverrides) Option {
	return func(o *registryOptions) { o.endpoints = overrides }
}

// applyEndpointOverride returns reg's endpoints with o's non-empty fields
// substituted, rejecting any override that would not take effect or that would
// weaken the connection.
//
// A field that is empty in code is NOT overridable. An empty Auth or Token
// means the provider builds it per flow (BuildAuthURL, BuildAuthURLForSite,
// BuildTokenURLForDomain, BuildTokenURLForSite); an empty APIBase or Identity
// means the host is customer-supplied through connector settings or
// discovered at runtime. In both cases the override would be silently
// ignored — and for a customer-supplied host, overriding it deployment-wide
// is exactly what the per-provider domain validators exist to prevent.
// Rejecting at startup turns a silent no-op into a boot failure the operator
// can see.
func applyEndpointOverride(p coredata.ConnectorProvider, base Endpoints, o Endpoints) (Endpoints, error) {
	fields := []struct {
		name     string
		current  *string
		override string
		// isBase marks APIBase and Identity: both are the root a driver
		// resolves further data or identity from — APIBase via
		// url.JoinPath, Identity via whatever the provider's response
		// carries (DocuSign's per-account base_uri) — rather than a complete
		// request URL, so they carry the extra restrictions
		// validateEndpointOverride applies only to a base.
		isBase bool
	}{
		{"auth", &base.Auth, o.Auth, false},
		{"token", &base.Token, o.Token, false},
		{"probe", &base.Probe, o.Probe, false},
		{"identity", &base.Identity, o.Identity, true},
		{"api-base", &base.APIBase, o.APIBase, true},
	}

	for _, f := range fields {
		if f.override == "" {
			continue
		}

		if *f.current == "" {
			return Endpoints{}, fmt.Errorf("cannot override %s endpoint for connector provider %q: the provider has no static value for it, so the override would be ignored", f.name, p)
		}

		if err := validateEndpointOverride(f.override, f.isBase); err != nil {
			return Endpoints{}, fmt.Errorf("cannot override %s endpoint for connector provider %q: %w", f.name, p, err)
		}

		*f.current = f.override
	}

	// A Probe override is only safe when the data path can move with it. If
	// neither APIBase nor Identity is set, the driver has no field for the
	// override to carry it to: the health check would move to the override
	// host while every data call stayed on the real provider, turning the
	// connection badge green for an environment the driver never reaches —
	// the DocuSign bug this guards against.
	if o.Probe != "" && base.APIBase == "" && base.Identity == "" {
		return Endpoints{}, fmt.Errorf("cannot override probe endpoint for connector provider %q: the provider has neither api-base nor identity to move with it, so the connection check would move while the driver's data calls stayed on the real provider", p)
	}

	return base, nil
}

// validateEndpointOverride rejects a value that is not an absolute https URL.
// The endpoints reach outbound requests that carry the connection's
// credentials, so an operator typo must not become a plaintext request or a
// request to a host derived from a relative path.
//
// isBase additionally rejects a query string, fragment, or trailing slash —
// restrictions that apply only to APIBase and Identity, the two fields a
// driver joins further path segments onto with url.JoinPath. A query on a
// joined base would propagate onto every path joined from it, and a probe
// builder that appends its own "?"+q.Encode() (buildNeonProbeURL,
// buildScalewayProbeURL) would produce a double-"?" URL at runtime instead of
// failing at boot. Auth, Token and Probe are complete URLs used verbatim, so
// they may legitimately carry a query (Google Workspace's and Apollo's Probe
// both do).
func validateEndpointOverride(raw string, isBase bool) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("cannot parse %q: %w", raw, err)
	}

	if !strings.EqualFold(u.Scheme, "https") {
		return fmt.Errorf("must be an https URL, got %q", raw)
	}

	if u.Host == "" {
		return fmt.Errorf("must include a host, got %q", raw)
	}

	if u.User != nil {
		return fmt.Errorf("must not embed credentials, got %q", u.Redacted())
	}

	if isBase {
		if u.RawQuery != "" || u.Fragment != "" {
			return fmt.Errorf("must not carry a query string or fragment when used as a base, got %q", raw)
		}

		if strings.HasSuffix(u.Path, "/") {
			return fmt.Errorf("must not have a trailing slash, got %q", raw)
		}
	}

	return nil
}

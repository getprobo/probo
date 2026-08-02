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
	"errors"
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
// substituted, rejecting any override that would not take effect, that would
// weaken the connection, or that would move only part of the provider's
// traffic.
//
// Two gates run before any field is substituted:
//
//   - Reach (unsupportedReason): some providers call at least one host that
//     is not expressed anywhere in Endpoints — Vercel's authorize URL and
//     OAuth callback, Crisp's connect-time subscription check, PostHog's
//     Cloud region discovery, Google Workspace's SDK-owned basePath. An
//     override could move what Endpoints DOES describe while those other
//     calls kept going to the real vendor, so such a provider refuses ANY
//     override rather than accept one that can only ever be partial.
//   - Atomicity: once any field is overridden, every field that is non-empty
//     in the provider's COMPILED Registration must be overridden too. This is
//     the fix for the bug an adversarial review found: an override that moved
//     DocuSign's Auth/Token/Probe but left Identity on its compiled value
//     produced a connector whose OAuth handshake and health check ran against
//     the sandbox while every account lookup — which resolves its host from
//     Identity — kept hitting production. The connection badge read healthy
//     for an environment the driver never reached. A field omitted from an
//     override is not "unset", it silently keeps its production value, so a
//     partial override is not a smaller version of a full one — it is a
//     deployment split across two environments that reports as one. For
//     DocuSign an override must set auth+token+probe+identity; for Brex
//     auth+token+probe+api-base; for an API-key provider with only a probe
//     and an api-base (Tally, for instance), those two.
//
// A field that is empty in the compiled Registration is NOT overridable
// regardless of atomicity. An empty Auth or Token means the provider builds
// it per flow (BuildAuthURL, BuildAuthURLForSite, BuildTokenURLForDomain,
// BuildTokenURLForSite); an empty APIBase or Identity means the host is
// customer-supplied through connector settings or discovered at runtime. In
// both cases the override would be silently ignored — and for a
// customer-supplied host, overriding it deployment-wide is exactly what the
// per-provider domain validators exist to prevent. Rejecting at startup turns
// a silent no-op into a boot failure the operator can see.
func applyEndpointOverride(p coredata.ConnectorProvider, unsupportedReason string, base Endpoints, o Endpoints) (Endpoints, error) {
	if o == (Endpoints{}) {
		return base, nil
	}

	if unsupportedReason != "" {
		return Endpoints{}, fmt.Errorf("cannot override endpoints for connector provider %q: %s", p, unsupportedReason)
	}

	fields := []struct {
		name     string
		current  *string
		original string
		override string
		// isBase marks APIBase and Identity: both are the root a driver
		// resolves further data or identity from — APIBase via
		// url.JoinPath, Identity via whatever the provider's response
		// carries (DocuSign's per-account base_uri) — rather than a complete
		// request URL, so they carry the extra restrictions
		// validateEndpointOverride applies only to a base.
		isBase bool
	}{
		{"auth", &base.Auth, base.Auth, o.Auth, false},
		{"token", &base.Token, base.Token, o.Token, false},
		{"probe", &base.Probe, base.Probe, o.Probe, false},
		{"identity", &base.Identity, base.Identity, o.Identity, true},
		{"api-base", &base.APIBase, base.APIBase, o.APIBase, true},
	}

	var missing []string

	for _, f := range fields {
		switch {
		case f.override == "" && f.original == "":
			// Neither compiled nor overridden: nothing to move.
		case f.override == "":
			// Compiled but left out of the override: collected so the
			// atomicity check below can name every field the operator still
			// needs to set, not just the first one found.
			missing = append(missing, f.name)
		case f.original == "":
			return Endpoints{}, fmt.Errorf("cannot override %s endpoint for connector provider %q: the provider has no static value for it, so the override would be ignored", f.name, p)
		default:
			if err := validateEndpointOverride(f.override, f.isBase); err != nil {
				return Endpoints{}, fmt.Errorf("cannot override %s endpoint for connector provider %q: %w", f.name, p, err)
			}

			*f.current = f.override
		}
	}

	if len(missing) > 0 {
		return Endpoints{}, fmt.Errorf("cannot override endpoints for connector provider %q: also override %s: an override must move every endpoint the provider has a static value for, or the field left behind keeps pointing at production", p, strings.Join(missing, ", "))
	}

	// A Probe override is only safe when the data path can move with it. If
	// neither APIBase nor Identity is set, the driver has no field for the
	// override to carry it to: the health check would move to the override
	// host while every data call stayed on the real provider, turning the
	// connection badge green for an environment the driver never reaches —
	// the DocuSign bug this guards against. Atomicity above cannot catch this
	// case by itself when Probe is the provider's only static field (e.g.
	// 1Password): moving that one field is trivially "every field", yet still
	// unsafe, because nothing else in Endpoints exists for it to carry.
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
		// url.Error embeds the URL it failed on, so wrapping it verbatim
		// would put the override back in the message. Unwrap to the cause.
		if urlErr, ok := errors.AsType[*url.Error](err); ok {
			return fmt.Errorf("cannot parse endpoint override: %w", urlErr.Err)
		}

		return fmt.Errorf("cannot parse endpoint override: %w", err)
	}

	// Every message below reports the override back to the operator, and this
	// error is what probod prints when it refuses to boot. An override that
	// embeds credentials must not put them in that log line, and the
	// credentials check is not the branch that fires first — a bad scheme on a
	// credentialed URL reaches the scheme message. Redact once, up front.
	safe := u.Redacted()

	if !strings.EqualFold(u.Scheme, "https") {
		return fmt.Errorf("must be an https URL, got %q", safe)
	}

	// A port makes Host non-empty even with no hostname ("https://:443"), so
	// the emptiness check has to run against Hostname or a hostless override
	// boots and fails later on connector traffic instead.
	if u.Host == "" || u.Hostname() == "" {
		return fmt.Errorf("must include a host, got %q", safe)
	}

	if u.User != nil {
		return fmt.Errorf("must not embed credentials, got %q", safe)
	}

	if isBase {
		// ForceQuery covers a bare "?": it leaves RawQuery empty but still
		// rides along into every path joined from this base.
		if u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
			return fmt.Errorf("must not carry a query string or fragment when used as a base, got %q", safe)
		}

		if strings.HasSuffix(u.Path, "/") {
			return fmt.Errorf("must not have a trailing slash, got %q", safe)
		}
	}

	return nil
}

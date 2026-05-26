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

package provider

import (
	"fmt"
	"maps"

	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
)

// ApplyOAuth2Defaults sets the redirect URI on c and applies static
// provider defaults (auth URL, token URL, extra params, token endpoint
// auth, PKCE) onto an OAuth2Connector, and wires an SSRF-protected
// HTTP client for the token exchange request. Static metadata is
// pulled from r; only ClientID and ClientSecret come from deployment
// config.
//
// Providers whose authorization URL embeds an operator-supplied slug
// (e.g. Vercel) derive it via Registration.BuildAuthURL from
// c.IntegrationSlug; this is a no-op when no slug is configured.
func (r *Registry) ApplyOAuth2Defaults(p string, redirectURI string, c *connector.OAuth2Connector) error {
	c.RedirectURI = redirectURI
	c.HTTPClient = httpclient.DefaultClient(httpclient.WithSSRFProtection())

	reg, ok := r.Get(coredata.ConnectorProvider(p))
	if !ok {
		return nil
	}

	c.AuthURL = reg.AuthURL
	c.TokenURL = reg.TokenURL
	c.TokenEndpointAuth = reg.TokenEndpointAuth
	c.SupportsIncrementalAuth = reg.SupportsIncrementalAuth
	c.RequiresPKCE = reg.RequiresPKCE

	// Deep copy ExtraAuthParams so per-connector mutations (e.g.
	// incremental auth, scope overrides) cannot alias back into the
	// shared registry map.
	if len(reg.ExtraAuthParams) > 0 {
		extra := make(map[string]string, len(reg.ExtraAuthParams))
		maps.Copy(extra, reg.ExtraAuthParams)
		c.ExtraAuthParams = extra
	}

	if reg.BuildAuthURL != nil && c.IntegrationSlug != "" {
		authURL, err := reg.BuildAuthURL(c.IntegrationSlug)
		if err != nil {
			return fmt.Errorf("cannot build %s auth URL: %w", p, err)
		}

		c.AuthURL = authURL
	}

	return nil
}

// ProbeURL returns the registered probe URL for provider p, or the
// empty string if no probe URL is configured.
func (r *Registry) ProbeURL(p string) string {
	reg, ok := r.Get(coredata.ConnectorProvider(p))
	if !ok {
		return ""
	}

	return reg.ProbeURL
}

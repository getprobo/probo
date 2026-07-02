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
	c.BuildAuthURLForSite = reg.BuildAuthURLForSite
	c.BuildTokenURLForDomain = reg.BuildTokenURLForDomain
	c.BuildTokenURLForSite = reg.BuildTokenURLForSite

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

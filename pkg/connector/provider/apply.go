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

	c.AuthURL = reg.Endpoints.Auth
	c.TokenURL = reg.Endpoints.Token

	// A provider with no OAuth2 path has no metadata to apply; the endpoints
	// above are all an OAuth2Connector can take from it.
	oauth2 := reg.OAuth2
	if oauth2 == nil {
		return nil
	}

	c.TokenEndpointAuth = oauth2.TokenEndpointAuth
	c.SupportsIncrementalAuth = oauth2.SupportsIncrementalAuth
	c.ExclusiveScopes = oauth2.ExclusiveScopes
	c.RegisteredScopes = oauth2.Scopes
	c.RequiresPKCE = oauth2.RequiresPKCE
	c.BuildAuthURLForSite = oauth2.BuildAuthURLForSite
	c.BuildTokenURLForDomain = oauth2.BuildTokenURLForDomain
	c.BuildTokenURLForSite = oauth2.BuildTokenURLForSite

	// Deep copy ExtraAuthParams so per-connector mutations (e.g.
	// incremental auth, scope overrides) cannot alias back into the
	// shared registry map.
	if len(oauth2.ExtraAuthParams) > 0 {
		extra := make(map[string]string, len(oauth2.ExtraAuthParams))
		maps.Copy(extra, oauth2.ExtraAuthParams)
		c.ExtraAuthParams = extra
	}

	if oauth2.BuildAuthURL != nil && c.IntegrationSlug != "" {
		authURL, err := oauth2.BuildAuthURL(c.IntegrationSlug)
		if err != nil {
			return fmt.Errorf("cannot build %s auth URL: %w", p, err)
		}

		c.AuthURL = authURL
	}

	return nil
}

// APIKeyFor returns the API key to present on outbound requests. For a
// ManagedAPIKey provider it is the Probo-held key resolved at use time, so
// rotation takes effect without rewriting the connection row. For every
// other provider it is the key stored on conn.
//
// It errors when a managed provider's key is unconfigured (the connector
// was deactivated after the connection was created).
func (r *Registry) APIKeyFor(
	provider coredata.ConnectorProvider,
	conn *connector.APIKeyConnection,
) (string, error) {
	reg, ok := r.Get(provider)
	if !ok || !reg.IsManagedAPIKey() {
		return conn.APIKey, nil
	}

	key, ok := r.ManagedAPIKey(provider)
	if !ok {
		return "", fmt.Errorf("cannot resolve managed API key for provider %q: not configured", provider)
	}

	return key, nil
}

// ProbeURL returns the registered probe URL for provider p, or the
// empty string if no probe URL is configured.
func (r *Registry) ProbeURL(p string) string {
	reg, ok := r.Get(coredata.ConnectorProvider(p))
	if !ok {
		return ""
	}

	return reg.Endpoints.Probe
}

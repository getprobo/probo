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

package provider_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"

	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// TestApplyOAuth2Defaults_AuthURLFromSlug verifies that providers whose
// authorization URL embeds an operator-supplied slug (Vercel) build it
// from c.IntegrationSlug via Registration.BuildAuthURL, with the slug
// percent-escaped. Providers without a BuildAuthURL are unaffected.
func TestApplyOAuth2Defaults_AuthURLFromSlug(t *testing.T) {
	t.Parallel()

	t.Run("auth URL is built when an integration slug is supplied", func(t *testing.T) {
		t.Parallel()

		r := provider.NewBuiltinRegistry()
		c := &connector.OAuth2Connector{
			ClientID:        "id",
			ClientSecret:    "secret",
			IntegrationSlug: "acme",
		}

		require.NoError(t, r.ApplyOAuth2Defaults("VERCEL", "https://example.com/cb", c))

		assert.Equal(t, "https://vercel.com/integrations/acme/new", c.AuthURL)
		assert.Equal(t, "https://api.vercel.com/v2/oauth/access_token", c.TokenURL)
	})

	t.Run("slug with reserved characters is percent-escaped", func(t *testing.T) {
		t.Parallel()

		r := provider.NewBuiltinRegistry()
		c := &connector.OAuth2Connector{
			ClientID:        "id",
			ClientSecret:    "secret",
			IntegrationSlug: "a/b c",
		}

		require.NoError(t, r.ApplyOAuth2Defaults("VERCEL", "https://example.com/cb", c))

		assert.Equal(t, "https://vercel.com/integrations/a%2Fb%20c/new", c.AuthURL)
	})

	t.Run("auth URL is empty when no integration slug is supplied", func(t *testing.T) {
		t.Parallel()

		r := provider.NewBuiltinRegistry()
		c := &connector.OAuth2Connector{
			ClientID:     "id",
			ClientSecret: "secret",
		}

		require.NoError(t, r.ApplyOAuth2Defaults("VERCEL", "https://example.com/cb", c))

		// Vercel has no static AuthURL; without a slug there is nothing
		// to build, so the misconfiguration surfaces at the
		// authorization step rather than being silently masked.
		assert.Empty(t, c.AuthURL)
	})
}

// TestApplyOAuth2Defaults_PKCEDefaults asserts that the registered
// PAGERDUTY provider defaults flip RequiresPKCE on so the downstream
// Initiate/Complete flow generates a verifier and replays it.
func TestApplyOAuth2Defaults_PKCEDefaults(t *testing.T) {
	t.Parallel()

	for _, p := range []string{"PAGERDUTY", "POSTHOG", "RESEND"} {
		t.Run(p, func(t *testing.T) {
			t.Parallel()

			r := provider.NewBuiltinRegistry()
			c := &connector.OAuth2Connector{ClientID: "id", ClientSecret: "secret"}
			require.NoError(t, r.ApplyOAuth2Defaults(p, "https://example.com/cb", c))
			assert.True(t, c.RequiresPKCE,
				"provider %s must enable PKCE so Initiate generates a verifier", p)
		})
	}
}

// TestApplyOAuth2Defaults_PublicClientTokenAuth verifies that CIMD clients
// propagate token_endpoint_auth_method "none" so token exchanges omit a
// client_secret.
func TestApplyOAuth2Defaults_PublicClientTokenAuth(t *testing.T) {
	t.Parallel()

	for _, p := range []string{"POSTHOG", "RESEND"} {
		t.Run(p, func(t *testing.T) {
			t.Parallel()

			r := provider.NewBuiltinRegistry()
			c := &connector.OAuth2Connector{}
			require.NoError(t, r.ApplyOAuth2Defaults(p, "https://example.com/cb", c))

			assert.Equal(t, "none", c.TokenEndpointAuth,
				"provider %s must use token_endpoint_auth_method none (public client)", p)
			assert.True(t, c.RequiresPKCE, "provider %s public client must require PKCE", p)
		})
	}
}

func TestApplyOAuth2Defaults_ResendCIMD(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	c := &connector.OAuth2Connector{}
	require.NoError(t, r.ApplyOAuth2Defaults("RESEND", "https://example.com/cb", c))

	assert.Equal(t, "https://api.resend.com/oauth/authorize", c.AuthURL)
	assert.Equal(t, "https://api.resend.com/oauth/token", c.TokenURL)
	assert.Equal(t, []string{"full_access"}, c.RegisteredScopes)

	publicProviders := make(map[coredata.ConnectorProvider]bool)
	for _, reg := range r.PublicClients() {
		publicProviders[reg.Provider] = true
	}

	assert.True(t, publicProviders[coredata.ConnectorProviderResend])
}

// TestApplyOAuth2Defaults_CopiesSiteClosures verifies the multi-site
// per-provider closures (BuildAuthURLForSite, BuildTokenURLForDomain) are
// copied from the Registration onto the OAuth2Connector.
func TestApplyOAuth2Defaults_CopiesSiteClosures(t *testing.T) {
	t.Parallel()

	r := provider.NewRegistry()
	require.NoError(t, r.Register(&provider.Registration{
		Provider:               coredata.ConnectorProviderDatadog,
		DisplayName:            "Datadog",
		OAuth2Scopes:           []string{"user_access_read"},
		RequiresPKCE:           true,
		BuildAuthURLForSite:    connector.DatadogAuthorizeURL,
		BuildTokenURLForDomain: connector.DatadogTokenURL,
		NewDriver: func(context.Context, *http.Client, *coredata.Connector, *log.Logger, provider.Endpoints) (drivers.Driver, error) {
			return nil, nil
		},
	}))

	var c connector.OAuth2Connector
	require.NoError(t, r.ApplyOAuth2Defaults("DATADOG", "https://probo.example/cb", &c))
	require.NotNil(t, c.BuildAuthURLForSite)
	require.NotNil(t, c.BuildTokenURLForDomain)

	// The copied closures resolve to Datadog's per-site / per-domain hosts.
	authURL, err := c.BuildAuthURLForSite("US3")
	require.NoError(t, err)
	assert.Equal(t, "https://us3.datadoghq.com/oauth2/v1/authorize", authURL)

	tokenURL, err := c.BuildTokenURLForDomain("us3.datadoghq.com")
	require.NoError(t, err)
	assert.Equal(t, "https://api.us3.datadoghq.com/oauth2/v1/token", tokenURL)
}

// TestApplyOAuth2Defaults_CopiesTokenURLForSiteClosure verifies the
// site-carried-in-state token-URL closure (BuildTokenURLForSite) is copied
// from the Registration onto the OAuth2Connector — the Zendesk shape, where
// both the authorize and token hosts are the customer subdomain.
func TestApplyOAuth2Defaults_CopiesTokenURLForSiteClosure(t *testing.T) {
	t.Parallel()

	r := provider.NewRegistry()
	require.NoError(t, r.Register(&provider.Registration{
		Provider:             coredata.ConnectorProviderZendesk,
		DisplayName:          "Zendesk",
		OAuth2Scopes:         []string{"users:read"},
		BuildAuthURLForSite:  connector.ZendeskAuthorizeURL,
		BuildTokenURLForSite: connector.ZendeskTokenURL,
		NewDriver: func(context.Context, *http.Client, *coredata.Connector, *log.Logger, provider.Endpoints) (drivers.Driver, error) {
			return nil, nil
		},
	}))

	var c connector.OAuth2Connector
	require.NoError(t, r.ApplyOAuth2Defaults("ZENDESK", "https://probo.example/cb", &c))
	require.NotNil(t, c.BuildAuthURLForSite)
	require.NotNil(t, c.BuildTokenURLForSite)
	require.Nil(t, c.BuildTokenURLForDomain)

	authURL, err := c.BuildAuthURLForSite("acme")
	require.NoError(t, err)
	assert.Equal(t, "https://acme.zendesk.com/oauth/authorizations/new", authURL)

	tokenURL, err := c.BuildTokenURLForSite("acme")
	require.NoError(t, err)
	assert.Equal(t, "https://acme.zendesk.com/oauth/tokens", tokenURL)
}

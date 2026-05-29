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

	for _, p := range []string{"PAGERDUTY", "POSTHOG"} {
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

// TestApplyOAuth2Defaults_PublicClientTokenAuth verifies that PostHog, a
// public (CIMD) client, propagates token_endpoint_auth_method "none" so the
// token exchange omits a client_secret.
func TestApplyOAuth2Defaults_PublicClientTokenAuth(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	c := &connector.OAuth2Connector{}
	require.NoError(t, r.ApplyOAuth2Defaults("POSTHOG", "https://example.com/cb", c))

	assert.Equal(t, "none", c.TokenEndpointAuth,
		"PostHog must use token_endpoint_auth_method none (public client)")
	assert.True(t, c.RequiresPKCE, "PostHog public client must require PKCE")
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
		NewDriver: func(context.Context, *http.Client, *coredata.Connector, *log.Logger) (drivers.Driver, error) {
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

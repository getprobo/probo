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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"

	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// capturingRoundTripper records the URL of every request it sees and returns
// a canned response, so a test can assert on the outgoing host without
// touching the network.
type capturingRoundTripper struct {
	urls   []string
	status int
}

func (c *capturingRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	c.urls = append(c.urls, r.URL.String())

	status := c.status
	if status == 0 {
		status = http.StatusUnauthorized
	}

	return &http.Response{StatusCode: status, Body: http.NoBody, Header: make(http.Header)}, nil
}

// TestEndpointOverrideRepointsProvider covers the case the feature exists for:
// a staging deployment pointing DocuSign at the vendor's demo hosts. This is
// the regression test for the bug an adversarial review found: DocuSign
// splits identity from data (see Endpoints.Identity), and every account
// lookup — the driver, the name resolver, and the org picker — resolves from
// Identity, not Probe. An override that moved Auth/Token/Probe without also
// moving Identity would leave every one of those calls on the real vendor
// while the connection badge read healthy against the sandbox; Register's
// Probe/Identity host invariant now refuses to boot on that half-migration,
// so Identity must move with Probe here.
func TestEndpointOverrideRepointsProvider(t *testing.T) {
	t.Parallel()

	r, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(provider.EndpointOverrides{
		coredata.ConnectorProviderDocuSign: {
			Auth:     "https://account-d.docusign.com/oauth/auth",
			Token:    "https://account-d.docusign.com/oauth/token",
			Probe:    "https://account-d.docusign.com/oauth/userinfo",
			Identity: "https://account-d.docusign.com/oauth/userinfo",
		},
	}))
	require.NoError(t, err)

	reg, ok := r.Get(coredata.ConnectorProviderDocuSign)
	require.True(t, ok)

	assert.Equal(t, "https://account-d.docusign.com/oauth/auth", reg.Endpoints.Auth)
	assert.Equal(t, "https://account-d.docusign.com/oauth/token", reg.Endpoints.Token)
	assert.Equal(t, "https://account-d.docusign.com/oauth/userinfo", reg.Endpoints.Probe)
	assert.Equal(t, "https://account-d.docusign.com/oauth/userinfo", reg.Endpoints.Identity)
}

// TestEndpointOverrideLeavesOthersAlone guards the blast radius: overriding one
// provider must not perturb any other. The override moves every DocuSign
// field (a partial one is now rejected by the atomicity check — see
// TestEndpointOverridePartialRejected) so this test exercises the accepted
// path rather than an error return.
func TestEndpointOverrideLeavesOthersAlone(t *testing.T) {
	t.Parallel()

	base := provider.NewBuiltinRegistry()

	overridden, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(provider.EndpointOverrides{
		coredata.ConnectorProviderDocuSign: {
			Auth:     "https://account-d.docusign.com/oauth/auth",
			Token:    "https://account-d.docusign.com/oauth/token",
			Probe:    "https://account-d.docusign.com/oauth/userinfo",
			Identity: "https://account-d.docusign.com/oauth/userinfo",
		},
	}))
	require.NoError(t, err)

	for _, want := range base.All() {
		got, ok := overridden.Get(want.Provider)
		require.True(t, ok, "provider %s vanished", want.Provider)

		if want.Provider == coredata.ConnectorProviderDocuSign {
			continue
		}

		assert.Equal(t, want.Endpoints, got.Endpoints, "provider %s was perturbed", want.Provider)
	}
}

// TestEndpointOverridePartialRejected pins the fix for the bug three
// adversarial reviews converged on: an override that set only Auth used to be
// ACCEPTED, silently leaving Token, Probe and Identity on their compiled
// (production) values — a connector whose OAuth handshake ran against the
// sandbox while its health check and every account lookup kept hitting
// production, reporting healthy the whole time. This test previously asserted
// that accept-and-keep-defaults behaviour as correct (as
// TestEndpointOverridePartialKeepsDefaults); it was the bug, not the
// contract, so it is inverted here: a partial override must be rejected, and
// the error must name every field left out.
func TestEndpointOverridePartialRejected(t *testing.T) {
	t.Parallel()

	r, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(provider.EndpointOverrides{
		coredata.ConnectorProviderDocuSign: {Auth: "https://account-d.docusign.com/oauth/auth"},
	}))
	require.Error(t, err)
	assert.Nil(t, r)
	assert.Contains(t, err.Error(), "token")
	assert.Contains(t, err.Error(), "probe")
	assert.Contains(t, err.Error(), "identity")
}

func TestEndpointOverrideRejected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		overrides provider.EndpointOverrides
		wantErr   string
	}{
		{
			name:      "unknown provider",
			overrides: provider.EndpointOverrides{coredata.ConnectorProvider("NOT_A_PROVIDER"): {Auth: "https://example.com/a"}},
			wantErr:   "unknown provider",
		},
		{
			// Datadog builds its authorize URL per site, so a static override
			// would be silently discarded by ApplyOAuth2Defaults.
			name:      "field the provider builds dynamically",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderDatadog: {Auth: "https://example.com/a"}},
			wantErr:   "no static value",
		},
		{
			// Grafana's host comes from connector settings; a deployment-wide
			// override there would bypass the per-connection domain validation.
			name:      "api base that is per-connection",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderGrafana: {APIBase: "https://example.com"}},
			wantErr:   "no static value",
		},
		{
			name:      "plaintext http",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderDocuSign: {Auth: "http://account-d.docusign.com/oauth/auth"}},
			wantErr:   "must be an https URL",
		},
		{
			name:      "relative url",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderDocuSign: {Auth: "/oauth/auth"}},
			wantErr:   "must be an https URL",
		},
		{
			name:      "embedded credentials",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderDocuSign: {Auth: "https://u" + ":p@account-d.docusign.com/oauth/auth"}},
			wantErr:   "must not embed credentials",
		},
		{
			// Register's invariant must see the overridden values, not the
			// compiled ones. Tally's only static fields are Probe and APIBase,
			// so overriding both satisfies the atomicity check above and
			// reaches Register with a genuine host mismatch.
			name: "probe and api base on different hosts",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderTally: {
				Probe:   "https://elsewhere.example.com/me",
				APIBase: "https://api.tally.so",
			}},
			wantErr: "does not match APIBase host",
		},
		{
			// Google Workspace's data host lives in the Google SDK's basePath,
			// which nothing in Endpoints describes: EndpointOverrideUnsupported
			// refuses ANY override before the atomicity/host checks even run.
			name:      "provider whose data host is outside Endpoints",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderGoogleWorkspace: {Probe: "https://admin.googleapis.com/admin/directory/v1/users?customer=other"}},
			wantErr:   "Google SDK's basePath",
		},
		{
			// 1Password's data host is per-connection (regional Users API or
			// customer-hosted SCIM bridge); same reasoning as Google Workspace.
			name:      "probe override rejected for provider with per-connection data host",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderOnePassword: {Probe: "https://events.1password.com/api/v1/other"}},
			wantErr:   "neither api-base nor identity",
		},
		{
			// url.JoinPath would propagate this query onto every path GitHub's
			// driver joins onto APIBase.
			name:      "api base with query string",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderGitHub: {APIBase: "https://api.github.com?x=1"}},
			wantErr:   "must not carry a query string or fragment",
		},
		{
			// The Endpoints doc requires APIBase to carry no trailing slash;
			// url.JoinPath would otherwise produce a double slash.
			name:      "api base with trailing slash",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderGitHub: {APIBase: "https://api.github.com/"}},
			wantErr:   "must not have a trailing slash",
		},
		{
			// DocuSign's Identity carries the same restrictions as APIBase: it
			// is the root the driver resolves accounts from, not a complete
			// request URL.
			name:      "identity with query string",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderDocuSign: {Identity: "https://account-d.docusign.com/oauth/userinfo?x=1"}},
			wantErr:   "must not carry a query string or fragment",
		},
		{
			// Every field is present (so the atomicity check above is
			// satisfied), but Identity was pointed at a different host than
			// Probe: Register must see the OVERRIDDEN values disagree, not
			// the compiled ones — the exact shape the DocuSign bug shipped
			// as, Auth/Token/Probe on the sandbox while account lookups
			// (which resolve from Identity) stayed on a stale host.
			name: "probe and identity moved to different hosts",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderDocuSign: {
				Auth:     "https://account-d.docusign.com/oauth/auth",
				Token:    "https://account-d.docusign.com/oauth/token",
				Probe:    "https://account-d.docusign.com/oauth/userinfo",
				Identity: "https://elsewhere.example.com/oauth/userinfo",
			}},
			wantErr: "does not match Identity host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(tt.overrides))
			require.Error(t, err)
			assert.Nil(t, r)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// TestEndpointOverrideAuthMayCarryQuery pins the asymmetry: the query/fragment
// restriction applies only to APIBase and Identity (a base further paths are
// joined onto), not to Auth/Token/Probe, which are complete URLs used
// verbatim and may legitimately carry one (Google Workspace's and Apollo's
// Probe both do). The override moves every DocuSign field — a partial one
// would now be rejected by the atomicity check — so this only exercises the
// query-string allowance on Auth.
func TestEndpointOverrideAuthMayCarryQuery(t *testing.T) {
	t.Parallel()

	r, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(provider.EndpointOverrides{
		coredata.ConnectorProviderDocuSign: {
			Auth:     "https://account-d.docusign.com/oauth/auth?prompt=login",
			Token:    "https://account-d.docusign.com/oauth/token",
			Probe:    "https://account-d.docusign.com/oauth/userinfo",
			Identity: "https://account-d.docusign.com/oauth/userinfo",
		},
	}))
	require.NoError(t, err)

	reg, ok := r.Get(coredata.ConnectorProviderDocuSign)
	require.True(t, ok)

	assert.Equal(t, "https://account-d.docusign.com/oauth/auth?prompt=login", reg.Endpoints.Auth)
}

// TestDocuSignUsesOverriddenIdentity is the mechanism-level regression test
// for the bug: it builds the driver, the name resolver, and the org picker
// through the registration closures with an overridden Endpoints and asserts
// each one's outgoing request actually hits the overridden identity host,
// not the compiled-in production one. Before the fix, this would have hit
// account.docusign.com regardless of the override.
func TestDocuSignUsesOverriddenIdentity(t *testing.T) {
	t.Parallel()

	const overriddenIdentity = "https://account-d.docusign.com/oauth/userinfo"

	r, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(provider.EndpointOverrides{
		coredata.ConnectorProviderDocuSign: {
			Auth:     "https://account-d.docusign.com/oauth/auth",
			Token:    "https://account-d.docusign.com/oauth/token",
			Probe:    overriddenIdentity,
			Identity: overriddenIdentity,
		},
	}))
	require.NoError(t, err)

	reg, ok := r.Get(coredata.ConnectorProviderDocuSign)
	require.True(t, ok)
	require.Equal(t, overriddenIdentity, reg.Endpoints.Identity)

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderDocuSign}
	require.NoError(t, conn.SetSettings(&coredata.DocuSignConnectorSettings{AccountID: "acct-1"}))

	logger := log.NewLogger(log.WithName("test"))

	t.Run("driver", func(t *testing.T) {
		t.Parallel()

		rt := &capturingRoundTripper{}
		client := &http.Client{Transport: rt}

		driver, err := reg.NewDriver(context.Background(), client, conn, logger, reg.Endpoints)
		require.NoError(t, err)

		// discoverBaseURI's userinfo call fails immediately (401 from the
		// capturing transport), so ListAccounts errors out — the request URL
		// it made on the way there is what this test is about.
		_, err = driver.ListAccounts(context.Background())
		require.Error(t, err)

		require.NotEmpty(t, rt.urls)
		assert.Equal(t, overriddenIdentity, rt.urls[0])
	})

	t.Run("name resolver", func(t *testing.T) {
		t.Parallel()

		rt := &capturingRoundTripper{}
		client := &http.Client{Transport: rt}

		resolver := reg.NewNameResolver(context.Background(), client, conn, logger, reg.Endpoints)
		require.NotNil(t, resolver)

		// A non-2xx userinfo response is the resolver's terminal case: it
		// returns ("", nil) rather than an error. The captured URL is what
		// this test asserts on.
		name, err := resolver.ResolveInstanceName(context.Background())
		require.NoError(t, err)
		assert.Empty(t, name)

		require.NotEmpty(t, rt.urls)
		assert.Equal(t, overriddenIdentity, rt.urls[0])
	})

	t.Run("org picker", func(t *testing.T) {
		t.Parallel()

		rt := &capturingRoundTripper{}
		client := &http.Client{Transport: rt}

		_, err := drivers.ListDocuSignOrganizations(context.Background(), client, reg.Endpoints.Identity)
		require.Error(t, err)

		require.NotEmpty(t, rt.urls)
		assert.Equal(t, overriddenIdentity, rt.urls[0])
	})
}

// TestDocumentedDocuSignOverrideBoots pins the exact override documented in
// .env.example against the registry, so the config an operator is told to copy
// cannot drift from the config that boots. An adversarial review found the
// earlier version of that block moved Auth, Token and Probe but not Identity,
// which probod now refuses — the documentation and the invariant have to agree.
func TestDocumentedDocuSignOverrideBoots(t *testing.T) {
	t.Parallel()

	const sandbox = "https://account-d.docusign.com"

	r, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(provider.EndpointOverrides{
		coredata.ConnectorProviderDocuSign: {
			Auth:     sandbox + "/oauth/auth",
			Token:    sandbox + "/oauth/token",
			Probe:    sandbox + "/oauth/userinfo",
			Identity: sandbox + "/oauth/userinfo",
		},
	}))
	require.NoError(t, err, ".env.example's documented DocuSign override must boot")

	reg, ok := r.Get(coredata.ConnectorProviderDocuSign)
	require.True(t, ok)

	// Every host the connector reaches must have moved together: the OAuth
	// handshake, the connection check, and the identity endpoint the driver
	// resolves each account's data host from.
	for name, got := range map[string]string{
		"auth":     reg.Endpoints.Auth,
		"token":    reg.Endpoints.Token,
		"probe":    reg.Endpoints.Probe,
		"identity": reg.Endpoints.Identity,
	} {
		assert.True(t, strings.HasPrefix(got, sandbox), "%s endpoint still points at %q", name, got)
	}
}

// TestEndpointOverrideRejectedForUnsupportedProviders covers the four
// providers an adversarial review found reach a host that lives outside
// Endpoints entirely, so no override — partial or full — could ever move all
// of their traffic: Vercel's authorize URL and OAuth callback, Crisp's
// connect-time subscription check, PostHog's Cloud region discovery, and
// Google Workspace's SDK-owned basePath. Registration.EndpointOverrideUnsupported
// refuses ANY override for these before the atomicity/reach checks even run,
// and the error must name the reason so the operator understands why.
func TestEndpointOverrideRejectedForUnsupportedProviders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		provider coredata.ConnectorProvider
		override provider.Endpoints
		wantErr  string
	}{
		{
			provider: coredata.ConnectorProviderVercel,
			override: provider.Endpoints{Token: "https://sandbox.example.com/v2/oauth/access_token"},
			wantErr:  "pinned host",
		},
		{
			provider: coredata.ConnectorProviderCrisp,
			override: provider.Endpoints{APIBase: "https://sandbox.example.com"},
			wantErr:  "GetCrispSubscriptionSettings",
		},
		{
			provider: coredata.ConnectorProviderPostHog,
			override: provider.Endpoints{Auth: "https://sandbox.example.com/oauth/authorize/"},
			wantErr:  "us.posthog.com",
		},
		{
			provider: coredata.ConnectorProviderGoogleWorkspace,
			override: provider.Endpoints{Auth: "https://sandbox.example.com/o/oauth2/v2/auth"},
			wantErr:  "Google SDK's basePath",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			t.Parallel()

			r, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(provider.EndpointOverrides{
				tt.provider: tt.override,
			}))
			require.Error(t, err)
			assert.Nil(t, r)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// TestEndpointOverridePartialBrexRejected is the section-E-mandated case for a
// second, non-DocuSign provider: overriding only api-base and leaving
// auth/token/probe on their compiled (production) values must be rejected,
// and the error must list every field the operator still needs to set.
func TestEndpointOverridePartialBrexRejected(t *testing.T) {
	t.Parallel()

	r, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(provider.EndpointOverrides{
		coredata.ConnectorProviderBrex: {APIBase: "https://sandbox.example.com"},
	}))
	require.Error(t, err)
	assert.Nil(t, r)
	assert.Contains(t, err.Error(), "auth")
	assert.Contains(t, err.Error(), "token")
	assert.Contains(t, err.Error(), "probe")
}

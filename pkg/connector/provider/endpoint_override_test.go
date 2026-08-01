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
// provider must not perturb any other.
func TestEndpointOverrideLeavesOthersAlone(t *testing.T) {
	t.Parallel()

	base := provider.NewBuiltinRegistry()

	overridden, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(provider.EndpointOverrides{
		coredata.ConnectorProviderDocuSign: {Auth: "https://account-d.docusign.com/oauth/auth"},
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

// TestEndpointOverridePartialKeepsDefaults verifies an omitted field keeps the
// compiled value rather than blanking it.
func TestEndpointOverridePartialKeepsDefaults(t *testing.T) {
	t.Parallel()

	base := provider.NewBuiltinRegistry()

	original, ok := base.Get(coredata.ConnectorProviderDocuSign)
	require.True(t, ok)

	r, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(provider.EndpointOverrides{
		coredata.ConnectorProviderDocuSign: {Auth: "https://account-d.docusign.com/oauth/auth"},
	}))
	require.NoError(t, err)

	reg, ok := r.Get(coredata.ConnectorProviderDocuSign)
	require.True(t, ok)

	assert.Equal(t, "https://account-d.docusign.com/oauth/auth", reg.Endpoints.Auth)
	assert.Equal(t, original.Endpoints.Token, reg.Endpoints.Token)
	assert.Equal(t, original.Endpoints.Probe, reg.Endpoints.Probe)
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
			// compiled ones: a probe left on the real vendor while the driver
			// moves is exactly the half-migration it exists to prevent.
			name:      "probe and api base on different hosts",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderAsana: {Probe: "https://elsewhere.example.com/api/1.0/users/me"}},
			wantErr:   "does not match APIBase host",
		},
		{
			// Google Workspace's data host lives in the Google SDK's basePath
			// (APIBase and Identity are both empty): moving Probe alone would
			// turn the badge green for a sandbox the SDK never talks to.
			name:      "probe override rejected when provider has no api base or identity",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderGoogleWorkspace: {Probe: "https://admin.googleapis.com/admin/directory/v1/users?customer=other"}},
			wantErr:   "neither api-base nor identity",
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
			// Overriding Probe alone (no matching Identity move) must fail —
			// this is the exact half-migration the bug shipped as: Auth/Token
			// hit the sandbox while every account lookup stayed on production.
			name:      "probe moved without identity",
			overrides: provider.EndpointOverrides{coredata.ConnectorProviderDocuSign: {Probe: "https://account-d.docusign.com/oauth/userinfo"}},
			wantErr:   "does not match Identity host",
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
// Probe both do).
func TestEndpointOverrideAuthMayCarryQuery(t *testing.T) {
	t.Parallel()

	r, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(provider.EndpointOverrides{
		coredata.ConnectorProviderDocuSign: {Auth: "https://account-d.docusign.com/oauth/auth?prompt=login"},
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

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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// TestEndpointOverrideRepointsProvider covers the case the feature exists for:
// a staging deployment pointing DocuSign at the vendor's demo hosts.
func TestEndpointOverrideRepointsProvider(t *testing.T) {
	t.Parallel()

	r, err := provider.NewBuiltinRegistryWith(provider.WithEndpointOverrides(provider.EndpointOverrides{
		coredata.ConnectorProviderDocuSign: {
			Auth:  "https://account-d.docusign.com/oauth/auth",
			Token: "https://account-d.docusign.com/oauth/token",
			Probe: "https://account-d.docusign.com/oauth/userinfo",
		},
	}))
	require.NoError(t, err)

	reg, ok := r.Get(coredata.ConnectorProviderDocuSign)
	require.True(t, ok)

	assert.Equal(t, "https://account-d.docusign.com/oauth/auth", reg.Endpoints.Auth)
	assert.Equal(t, "https://account-d.docusign.com/oauth/token", reg.Endpoints.Token)
	assert.Equal(t, "https://account-d.docusign.com/oauth/userinfo", reg.Endpoints.Probe)
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

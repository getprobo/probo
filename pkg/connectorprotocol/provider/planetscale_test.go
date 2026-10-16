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
	"go.probo.inc/probo/pkg/connectorprotocol"
	"go.probo.inc/probo/pkg/connectorprotocol/provider"
	"go.probo.inc/probo/pkg/coredata"
)

func TestPlanetScaleRegistration(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderPlanetScale)
	require.True(t, ok)

	assert.False(t, reg.SupportsAPIKey())
	assert.True(t, reg.OAuth2.RequiresPKCE)
	assert.False(t, reg.OAuth2.PublicClient)
	assert.Equal(
		t,
		[]string{"read_organizations", "read_user", "read_organization", "read_members"},
		reg.OAuth2.Scopes,
	)
	assert.Equal(t, "https://api.planetscale.com/v1/user", reg.Endpoints.Probe)
	assert.Equal(t, "https://www.probo.com/docs/product/access-review/planetscale", reg.DocumentationURL)

	require.NotNil(t, reg.SetOrganizationSettings)

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderPlanetScale}
	require.NoError(t, reg.SetOrganizationSettings(conn, "acme"))

	settings, err := coredata.ConnectorSettings[coredata.PlanetScaleConnectorSettings](conn)
	require.NoError(t, err)
	assert.Equal(t, "acme", settings.Organization)

	externalID, name, err := reg.ResolveInitialAccount(conn)
	require.NoError(t, err)
	assert.Equal(t, "acme", externalID)
	assert.Equal(t, "acme", name)

	oauthConnector := &connectorprotocol.OAuth2Connector{}
	require.NoError(t, r.ApplyOAuth2Defaults(
		string(coredata.ConnectorProviderPlanetScale),
		"https://example.com/callback",
		oauthConnector,
	))

	assert.Equal(t, "https://app.planetscale.com/oauth/authorize", oauthConnector.AuthURL)
	assert.Equal(t, "https://auth.planetscale.com/oauth/token", oauthConnector.TokenURL)

	assert.Equal(t, []string{"read_organizations", "read_user"}, oauthConnector.RefreshScopes)
}

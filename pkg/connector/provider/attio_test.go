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
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

func TestAttioRegistration(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderAttio)
	require.True(t, ok)

	assert.True(t, reg.SupportsAPIKey())
	assert.False(t, reg.OAuth2.RequiresPKCE)
	assert.False(t, reg.OAuth2.PublicClient)
	assert.Equal(t, []string{"user_management:read"}, reg.OAuth2.Scopes)

	// Attio has nothing to pick and nothing to capture: every credential is
	// bound to one workspace.
	assert.Nil(t, reg.SetOrganizationSettings)

	// The probe must be the members endpoint. /v2/self answers 200 with
	// {"active": false} for a revoked token and 400 for a malformed one, both
	// of which a status-only check reads as connected.
	assert.Equal(t, "https://api.attio.com/v2/workspace_members", reg.Endpoints.Probe)

	// No shape is asserted on a pasted key: Attio documents no prefix.
	assert.Nil(t, reg.APIKey.KeyFormat)

	oauthConnector := &connector.OAuth2Connector{}
	require.NoError(t, r.ApplyOAuth2Defaults(
		string(coredata.ConnectorProviderAttio),
		"https://example.com/callback",
		oauthConnector,
	))

	assert.Equal(t, "https://app.attio.com/authorize", oauthConnector.AuthURL)
	assert.Equal(t, "https://app.attio.com/oauth/token", oauthConnector.TokenURL)
	assert.Equal(t, []string{"user_management:read"}, oauthConnector.RegisteredScopes)
}

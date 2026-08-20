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

func TestCalendlyRegistration(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderCalendly)
	require.True(t, ok)

	assert.True(t, reg.SupportsAPIKey)
	assert.True(t, reg.RequiresPKCE)
	assert.Equal(t, []string{"users:read", "organizations:read"}, reg.OAuth2Scopes)

	oauthConnector := &connector.OAuth2Connector{}
	require.NoError(t, r.ApplyOAuth2Defaults(
		string(coredata.ConnectorProviderCalendly),
		"https://example.com/callback",
		oauthConnector,
	))

	assert.Equal(t, "https://auth.calendly.com/oauth/authorize", oauthConnector.AuthURL)
	assert.Equal(t, "https://auth.calendly.com/oauth/token", oauthConnector.TokenURL)
	assert.Equal(t, []string{"users:read", "organizations:read"}, oauthConnector.RegisteredScopes)
	assert.True(t, oauthConnector.RequiresPKCE)
}

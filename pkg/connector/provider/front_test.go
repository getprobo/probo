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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

func TestFrontRegistrationMetadata(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderFront)
	require.True(t, ok, "front provider must be registered")

	assert.Equal(t, "Front", reg.DisplayName)

	// Both credential paths are offered: OAuth and a company API token.
	assert.Equal(t, "https://app.frontapp.com/oauth/authorize", reg.Endpoints.Auth)
	assert.Equal(t, "https://app.frontapp.com/oauth/token", reg.Endpoints.Token)
	assert.Equal(t, "basic-form", reg.TokenEndpointAuth, "front requires HTTP Basic client credentials on the token exchange")
	assert.True(t, reg.SupportsAPIKey)

	// The Core API lives on a different host from the OAuth endpoints.
	assert.Equal(t, "https://api2.frontapp.com", reg.Endpoints.APIBase)
	assert.Equal(t, "https://api2.frontapp.com/me", reg.Endpoints.Probe)

	// Front resolves an OAuth token's scopes from the app configuration, so
	// none are requested per-authorization.
	assert.Empty(t, reg.OAuth2Scopes)
	// The company API token is a plain Bearer credential and the token is
	// already company-scoped, so there is nothing extra to collect.
	assert.Empty(t, reg.APIKeyExtraSettings)
	assert.Empty(t, reg.APIKeyHeader)
	assert.Empty(t, reg.APIKeyAuthScheme)
	assert.False(t, reg.APIKeyBasicAuth)
	assert.False(t, reg.APIKeyBasicAuthUserPass)
}

func TestFrontFactories(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderFront)
	require.True(t, ok, "front provider must be registered")
	require.NotNil(t, reg.NewDriver, "front NewDriver closure must be wired")
	require.NotNil(t, reg.NewNameResolver, "front NewNameResolver closure must be wired")

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderFront}
	client := httpclient.DefaultClient(httpclient.WithSSRFProtection())

	drv, err := reg.NewDriver(context.Background(), client, conn, nil, reg.Endpoints)
	require.NoError(t, err)
	assert.IsType(t, &drivers.FrontDriver{}, drv)

	resolver := reg.NewNameResolver(context.Background(), client, conn, nil, reg.Endpoints)
	require.NotNil(t, resolver, "front name resolver must be constructed")
}

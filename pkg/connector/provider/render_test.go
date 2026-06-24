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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

func TestRenderRegistrationMetadata(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderRender)
	require.True(t, ok, "render provider must be registered")

	assert.Equal(t, "Render", reg.DisplayName)
	assert.True(t, reg.SupportsAPIKey)
	// Render authenticates with the default Authorization: Bearer scheme.
	assert.Empty(t, reg.APIKeyAuthScheme)
	assert.Empty(t, reg.APIKeyHeader)
	assert.False(t, reg.APIKeyBasicAuth)
	// No OAuth and no picker.
	assert.Empty(t, reg.AuthURL)
	assert.Nil(t, reg.SetOrganizationSettings)

	require.Len(t, reg.ExtraSettings, 1)
	assert.Equal(t, "workspaceId", reg.ExtraSettings[0].Key)
	assert.Equal(t, "Workspace ID", reg.ExtraSettings[0].Label)
	assert.True(t, reg.ExtraSettings[0].Required)
}

func TestRenderNewDriver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderRender)
	require.True(t, ok, "render provider must be registered")
	require.NotNil(t, reg.NewDriver, "render NewDriver closure must be wired")

	t.Run("creates driver with valid owner id", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.RenderConnectorSettings{
			OwnerID: "tea-csp8nlbgbbvc73a8nn9g",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderRender,
			RawSettings: raw,
		}

		drv, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
		require.NoError(t, err)
		assert.IsType(t, &drivers.RenderDriver{}, drv)
	})

	t.Run("errors when owner id is missing", func(t *testing.T) {
		t.Parallel()

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderRender,
			RawSettings: []byte(`{}`),
		}

		_, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "owner_id is required")
	})
}

func TestRenderNewNameResolver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderRender)
	require.True(t, ok, "render provider must be registered")
	require.NotNil(t, reg.NewNameResolver, "render NewNameResolver closure must be wired")

	raw, err := json.Marshal(&coredata.RenderConnectorSettings{
		OwnerID: "tea-csp8nlbgbbvc73a8nn9g",
	})
	require.NoError(t, err)

	conn := &coredata.Connector{
		Provider:    coredata.ConnectorProviderRender,
		RawSettings: raw,
	}

	resolver := reg.NewNameResolver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
	require.NotNil(t, resolver, "render name resolver must be constructed for a valid connector")
}

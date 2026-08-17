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

func TestAuthentikRegistrationMetadata(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderAuthentik)
	require.True(t, ok, "authentik provider must be registered")

	assert.Equal(t, "authentik", reg.DisplayName)
	assert.True(t, reg.SupportsAPIKey)
	assert.Empty(t, reg.APIKeyHeader)
	assert.Empty(t, reg.APIKeyAuthScheme)
	assert.False(t, reg.APIKeyBasicAuth)
	assert.False(t, reg.APIKeyBasicAuthUserPass)
	require.Len(t, reg.APIKeyExtraSettings, 1)
	assert.Equal(t, "baseUrl", reg.APIKeyExtraSettings[0].Key)
	assert.True(t, reg.APIKeyExtraSettings[0].Required)
	assert.Empty(t, reg.Endpoints.APIBase)
	assert.NotNil(t, reg.BuildProbeURL)
	assert.NotNil(t, reg.NewNameResolver)
}

func TestAuthentikNewDriver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderAuthentik)
	require.True(t, ok, "authentik provider must be registered")
	require.NotNil(t, reg.NewDriver, "authentik NewDriver closure must be wired")

	client := httpclient.DefaultClient(httpclient.WithSSRFProtection())

	t.Run("creates driver with valid base_url", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.AuthentikConnectorSettings{
			BaseURL: "https://authentik.example.com/",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderAuthentik,
			RawSettings: raw,
		}

		drv, err := reg.NewDriver(context.Background(), client, conn, nil, reg.Endpoints)
		require.NoError(t, err)
		assert.IsType(t, &drivers.AuthentikDriver{}, drv)

		probeURL, err := reg.BuildProbeURL(conn, reg.Endpoints)
		require.NoError(t, err)
		assert.Equal(t, "https://authentik.example.com/api/v3/core/users/me/", probeURL)
	})

	for name, baseURL := range map[string]string{
		"missing":           "",
		"no scheme":         "authentik.example.com",
		"bad scheme":        "ftp://authentik.example.com",
		"no host":           "https://",
		"unparseable":       "https://authentik.example.com/%zz",
		"port without host": "https://:443",
	} {
		t.Run("rejects "+name+" base_url", func(t *testing.T) {
			t.Parallel()

			raw, err := json.Marshal(&coredata.AuthentikConnectorSettings{BaseURL: baseURL})
			require.NoError(t, err)

			conn := &coredata.Connector{
				Provider:    coredata.ConnectorProviderAuthentik,
				RawSettings: raw,
			}

			_, err = reg.NewDriver(context.Background(), client, conn, nil, reg.Endpoints)
			require.Error(t, err)

			_, err = reg.BuildProbeURL(conn, reg.Endpoints)
			require.Error(t, err)
		})
	}
}

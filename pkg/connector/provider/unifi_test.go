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
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview/drivers"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

func unifiTestRegistration(t *testing.T) *provider.Registration {
	t.Helper()

	reg, ok := provider.NewBuiltinRegistry().Get(coredata.ConnectorProviderUniFi)
	require.True(t, ok, "unifi provider must be registered")

	return reg
}

func TestUniFiRegistrationMetadata(t *testing.T) {
	t.Parallel()

	reg := unifiTestRegistration(t)

	assert.Equal(t, "UniFi", reg.DisplayName)
	assert.True(t, reg.SupportsAPIKey)
	// UniFi rejects a Bearer token; the console API key rides in X-API-KEY.
	assert.Equal(t, "X-API-KEY", reg.APIKeyHeader)
	// There is no OAuth2 program for the Network API.
	assert.Empty(t, reg.OAuth2Scopes)
	assert.Empty(t, reg.Endpoints.Auth)
	assert.Empty(t, reg.Endpoints.Token)
	// The probe is per-connector (it embeds the console ID), so there is no
	// static Probe URL to pin.
	assert.Empty(t, reg.Endpoints.Probe)
	require.NotNil(t, reg.BuildProbeURL)
	assert.Equal(t, "https://api.ui.com/v1/connector/consoles", reg.Endpoints.APIBase)

	require.Len(t, reg.APIKeyExtraSettings, 1)
	assert.Equal(t, "consoleId", reg.APIKeyExtraSettings[0].Key)
	assert.Equal(t, "Console ID", reg.APIKeyExtraSettings[0].Label)
	assert.True(t, reg.APIKeyExtraSettings[0].Required)

	require.NotNil(t, reg.NewDriver)
	require.NotNil(t, reg.NewNameResolver)
}

func TestUniFiNewDriver(t *testing.T) {
	t.Parallel()

	reg := unifiTestRegistration(t)

	t.Run("creates driver with a console id", func(t *testing.T) {
		t.Parallel()

		drv, err := reg.NewDriver(
			context.Background(),
			httpclient.DefaultClient(httpclient.WithSSRFProtection()),
			unifiTestConnector(t, "ABCDEF0123456789:1234567890"),
			log.NewLogger(log.WithName("test")),
			reg.Endpoints,
		)
		require.NoError(t, err)
		assert.IsType(t, &drivers.UniFiDriver{}, drv)
	})

	// Without a console ID the URL would resolve to the console collection
	// itself, so the review would silently read something other than a
	// console. Refuse instead.
	t.Run("errors when the console id is missing", func(t *testing.T) {
		t.Parallel()

		_, err := reg.NewDriver(
			context.Background(),
			httpclient.DefaultClient(httpclient.WithSSRFProtection()),
			&coredata.Connector{
				Provider:    coredata.ConnectorProviderUniFi,
				RawSettings: []byte(`{}`),
			},
			log.NewLogger(log.WithName("test")),
			reg.Endpoints,
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "console_id is required")
	})
}

func TestUniFiNewNameResolver(t *testing.T) {
	t.Parallel()

	reg := unifiTestRegistration(t)

	resolver := reg.NewNameResolver(
		context.Background(),
		httpclient.DefaultClient(httpclient.WithSSRFProtection()),
		unifiTestConnector(t, "ABCDEF0123456789:1234567890"),
		log.NewLogger(log.WithName("test")),
		reg.Endpoints,
	)
	require.NotNil(t, resolver)

	// A connector with no console ID has nothing to resolve against; the
	// worker treats a nil resolver as "keep the generic name".
	assert.Nil(t, reg.NewNameResolver(
		context.Background(),
		httpclient.DefaultClient(httpclient.WithSSRFProtection()),
		&coredata.Connector{
			Provider:    coredata.ConnectorProviderUniFi,
			RawSettings: []byte(`{}`),
		},
		log.NewLogger(log.WithName("test")),
		reg.Endpoints,
	))
}

func unifiTestConnector(t *testing.T, consoleID string) *coredata.Connector {
	t.Helper()

	raw, err := json.Marshal(&coredata.UniFiConnectorSettings{ConsoleID: consoleID})
	require.NoError(t, err)

	return &coredata.Connector{
		Provider:    coredata.ConnectorProviderUniFi,
		RawSettings: raw,
	}
}

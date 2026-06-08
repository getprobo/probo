// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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

func TestSigNozRegistrationMetadata(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderSigNoz)
	require.True(t, ok, "signoz provider must be registered")

	assert.Equal(t, "SigNoz", reg.DisplayName)
	assert.True(t, reg.SupportsAPIKey)
	assert.Equal(t, "SIGNOZ-API-KEY", reg.APIKeyHeader)
	require.Len(t, reg.ExtraSettings, 1)
	assert.Equal(t, "baseUrl", reg.ExtraSettings[0].Key)
	assert.Equal(t, "Base URL", reg.ExtraSettings[0].Label)
	assert.True(t, reg.ExtraSettings[0].Required)
	require.NotNil(t, reg.NewNameResolver, "signoz NewNameResolver closure must be wired")
}

func TestSigNozNewNameResolver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderSigNoz)
	require.True(t, ok, "signoz provider must be registered")
	require.NotNil(t, reg.NewNameResolver, "signoz NewNameResolver closure must be wired")

	raw, err := json.Marshal(&coredata.SigNozConnectorSettings{
		BaseURL: "https://acme.us.signoz.cloud",
	})
	require.NoError(t, err)

	conn := &coredata.Connector{
		Provider:    coredata.ConnectorProviderSigNoz,
		RawSettings: raw,
	}

	resolver := reg.NewNameResolver(
		context.Background(),
		httpclient.DefaultClient(httpclient.WithSSRFProtection()),
		conn,
		nil,
	)
	require.NotNil(t, resolver)
}

func TestSigNozNewDriver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderSigNoz)
	require.True(t, ok, "signoz provider must be registered")
	require.NotNil(t, reg.NewDriver, "signoz NewDriver closure must be wired")

	t.Run("creates driver with valid base_url", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.SigNozConnectorSettings{
			BaseURL: "https://cloud.signoz.io",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderSigNoz,
			RawSettings: raw,
		}

		drv, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
		require.NoError(t, err)
		assert.IsType(t, &drivers.SigNozDriver{}, drv)
	})

	t.Run("errors when base_url is missing", func(t *testing.T) {
		t.Parallel()

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderSigNoz,
			RawSettings: []byte(`{}`),
		}

		_, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base_url is required")
	})

	t.Run("errors when base_url is invalid", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.SigNozConnectorSettings{
			BaseURL: "ftp://cloud.signoz.io",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderSigNoz,
			RawSettings: raw,
		}

		_, err = reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base_url must be an http(s) URL")
	})
}

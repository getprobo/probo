// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

func TestLangfuseRegistrationMetadata(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderLangfuse)
	require.True(t, ok, "langfuse provider must be registered")

	assert.Equal(t, "Langfuse", reg.DisplayName)
	assert.True(t, reg.SupportsAPIKey)
	// Langfuse presents publicKey:secretKey as a full HTTP Basic credential.
	assert.True(t, reg.APIKeyBasicAuthUserPass)
	assert.Empty(t, reg.APIKeyHeader)
	assert.Empty(t, reg.APIKeyAuthScheme)
	require.Len(t, reg.ExtraSettings, 1)
	assert.Equal(t, "baseUrl", reg.ExtraSettings[0].Key)
	assert.Equal(t, "Base URL", reg.ExtraSettings[0].Label)
	assert.True(t, reg.ExtraSettings[0].Required)
	// Single-tenant API-key provider: no picker, no name resolver.
	assert.Nil(t, reg.NewNameResolver, "langfuse must not wire a name resolver")
	assert.Nil(t, reg.SetOrganizationSettings, "langfuse must not wire a picker store")
}

func TestLangfuseNewDriver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderLangfuse)
	require.True(t, ok, "langfuse provider must be registered")
	require.NotNil(t, reg.NewDriver, "langfuse NewDriver closure must be wired")

	t.Run("creates driver with valid base_url", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.LangfuseConnectorSettings{
			BaseURL: "https://cloud.langfuse.com",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderLangfuse,
			RawSettings: raw,
		}

		drv, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
		require.NoError(t, err)
		assert.IsType(t, &drivers.LangfuseDriver{}, drv)
	})

	t.Run("errors when base_url is missing", func(t *testing.T) {
		t.Parallel()

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderLangfuse,
			RawSettings: []byte(`{}`),
		}

		_, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base_url is required")
	})

	t.Run("errors when base_url is invalid", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.LangfuseConnectorSettings{
			BaseURL: "ftp://cloud.langfuse.com",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderLangfuse,
			RawSettings: raw,
		}

		_, err = reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "base_url must be an http(s) URL")
	})
}

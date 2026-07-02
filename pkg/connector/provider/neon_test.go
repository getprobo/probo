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

func TestNeonRegistrationMetadata(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderNeon)
	require.True(t, ok, "neon provider must be registered")

	assert.Equal(t, "Neon", reg.DisplayName)
	assert.True(t, reg.SupportsAPIKey)
	assert.Empty(t, reg.APIKeyAuthScheme, "neon API keys use the default Bearer scheme")
	require.Len(t, reg.ExtraSettings, 1)
	assert.Equal(t, "organizationId", reg.ExtraSettings[0].Key)
	assert.Equal(t, "Organization ID", reg.ExtraSettings[0].Label)
	assert.True(t, reg.ExtraSettings[0].Required)
}

func TestNeonNewDriver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderNeon)
	require.True(t, ok, "neon provider must be registered")
	require.NotNil(t, reg.NewDriver, "neon NewDriver closure must be wired")

	t.Run("creates driver with valid organization_id", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.NeonConnectorSettings{
			OrganizationID: "org-cool-breeze-12345678",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderNeon,
			RawSettings: raw,
		}

		drv, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
		require.NoError(t, err)
		assert.IsType(t, &drivers.NeonDriver{}, drv)
	})

	t.Run("errors when organization_id is missing", func(t *testing.T) {
		t.Parallel()

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderNeon,
			RawSettings: []byte(`{}`),
		}

		_, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "organization_id is required")
	})
}

func TestNeonNewNameResolver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderNeon)
	require.True(t, ok, "neon provider must be registered")
	require.NotNil(t, reg.NewNameResolver, "neon NewNameResolver closure must be wired")

	raw, err := json.Marshal(&coredata.NeonConnectorSettings{
		OrganizationID: "org-cool-breeze-12345678",
	})
	require.NoError(t, err)

	conn := &coredata.Connector{
		Provider:    coredata.ConnectorProviderNeon,
		RawSettings: raw,
	}

	resolver := reg.NewNameResolver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
	require.NotNil(t, resolver, "neon name resolver must be constructed for a valid connector")
}

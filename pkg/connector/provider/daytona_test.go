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

func TestDaytonaRegistrationMetadata(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderDaytona)
	require.True(t, ok, "daytona provider must be registered")

	assert.Equal(t, "Daytona", reg.DisplayName)
	assert.True(t, reg.SupportsAPIKey())
	assert.Equal(t, provider.APIKeyAuthBearer, reg.APIKey.Auth.Mode)
	assert.Empty(t, reg.APIKeyExtraSettings(), "the organization is picked from GET /organizations, not typed")
	assert.NotNil(t, reg.SetOrganizationSettings, "daytona must store the picked organization")
	assert.Equal(t, "https://app.daytona.io/api", reg.Endpoints.APIBase)
	assert.Equal(t, "https://app.daytona.io/api/organizations", reg.Endpoints.Probe)
}

func TestDaytonaSetOrganizationSettings(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderDaytona)
	require.True(t, ok, "daytona provider must be registered")
	require.NotNil(t, reg.SetOrganizationSettings)

	conn := &coredata.Connector{Provider: coredata.ConnectorProviderDaytona}
	require.NoError(t, reg.SetOrganizationSettings(conn, "aaaaaaaa-1111-2222-3333-000000000001"))

	s, err := coredata.ConnectorSettings[coredata.DaytonaConnectorSettings](conn)
	require.NoError(t, err)
	assert.Equal(t, "aaaaaaaa-1111-2222-3333-000000000001", s.OrganizationID)
}

func TestDaytonaNewDriver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderDaytona)
	require.True(t, ok, "daytona provider must be registered")
	require.NotNil(t, reg.NewDriver, "daytona NewDriver closure must be wired")

	t.Run("creates driver with valid organization_id", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.DaytonaConnectorSettings{
			OrganizationID: "aaaaaaaa-1111-2222-3333-000000000001",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderDaytona,
			RawSettings: raw,
		}

		drv, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil, reg.Endpoints)
		require.NoError(t, err)
		assert.IsType(t, &drivers.DaytonaDriver{}, drv)
	})

	t.Run("errors when organization_id is missing", func(t *testing.T) {
		t.Parallel()

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderDaytona,
			RawSettings: []byte(`{}`),
		}

		_, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil, reg.Endpoints)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "organization_id is required")
	})
}

func TestDaytonaNewNameResolver(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderDaytona)
	require.True(t, ok, "daytona provider must be registered")
	require.NotNil(t, reg.NewNameResolver, "daytona NewNameResolver closure must be wired")

	raw, err := json.Marshal(&coredata.DaytonaConnectorSettings{
		OrganizationID: "aaaaaaaa-1111-2222-3333-000000000001",
	})
	require.NoError(t, err)

	conn := &coredata.Connector{
		Provider:    coredata.ConnectorProviderDaytona,
		RawSettings: raw,
	}

	resolver := reg.NewNameResolver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil, reg.Endpoints)
	require.NotNil(t, resolver, "daytona name resolver must be constructed for a valid connector")
}

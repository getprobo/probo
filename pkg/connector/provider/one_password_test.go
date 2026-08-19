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
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// TestOnePasswordRegistrationMetadata pins the per-connect-path settings
// split. 1Password is the only registration offering both paths, and each
// needs different settings because a different driver sits behind each: a
// dialog handed the other path's list would collect fields the create
// resolver rejects, which is exactly how the API-key path was broken while
// the two shapes shared one flat list.
func TestOnePasswordRegistrationMetadata(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderOnePassword)
	require.True(t, ok, "1Password provider must be registered")

	assert.Equal(t, "1Password", reg.DisplayName)
	assert.True(t, reg.SupportsAPIKey)
	assert.True(t, reg.SupportsClientCredentials)

	require.Len(t, reg.APIKeyExtraSettings, 1)
	assert.Equal(t, "scimBridgeUrl", reg.APIKeyExtraSettings[0].Key)
	assert.Equal(t, "SCIM Bridge URL", reg.APIKeyExtraSettings[0].Label)
	assert.True(t, reg.APIKeyExtraSettings[0].Required)

	require.Len(t, reg.ClientCredentialsExtraSettings, 2)
	assert.Equal(t, "accountId", reg.ClientCredentialsExtraSettings[0].Key)
	assert.Equal(t, "Account ID", reg.ClientCredentialsExtraSettings[0].Label)
	assert.True(t, reg.ClientCredentialsExtraSettings[0].Required)
	assert.Equal(t, "region", reg.ClientCredentialsExtraSettings[1].Key)
	assert.Equal(t, "Region", reg.ClientCredentialsExtraSettings[1].Label)
	assert.True(t, reg.ClientCredentialsExtraSettings[1].Required)
}

// TestOnePassword_NewDriver_DispatchByGrantType is the pre-merge gate
// for the 1Password closure. The OnePassword registration dispatches
// between two drivers on GrantType(), which is "" for an API-key
// connection — this test asserts every connection shape reaching the
// closure constructs the driver whose settings the matching
// per-path settings list collects.
func TestOnePassword_NewDriver_DispatchByGrantType(t *testing.T) {
	t.Parallel()

	r := provider.NewBuiltinRegistry()
	reg, ok := r.Get(coredata.ConnectorProviderOnePassword)
	require.True(t, ok, "1Password provider must be registered")
	require.NotNil(t, reg.NewDriver, "1Password NewDriver closure must be wired")

	t.Run("client_credentials uses Users API driver", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.OnePasswordUsersAPISettings{
			AccountID: "test-account",
			Region:    "us",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderOnePassword,
			RawSettings: raw,
			Connection: &connector.OAuth2Connection{
				GrantType: connector.OAuth2GrantTypeClientCredentials,
			},
		}

		drv, err := reg.NewDriver(context.Background(), provider.NewHTTPHandleForTest(reg, conn, httpclient.DefaultClient(httpclient.WithSSRFProtection())), nil)
		require.NoError(t, err)
		assert.IsType(t, &drivers.OnePasswordUsersAPIDriver{}, drv)
	})

	// The production API-key path: an *APIKeyConnection makes GrantType()
	// return "", so it falls through to the SCIM-bridge driver and reads the
	// SCIMBridgeURL that APIKeyExtraSettings collects.
	t.Run("api key uses SCIM-bridge driver", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.OnePasswordConnectorSettings{
			SCIMBridgeURL: "https://scim.example.test",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderOnePassword,
			RawSettings: raw,
			Connection:  &connector.APIKeyConnection{APIKey: "scim-token"},
		}

		drv, err := reg.NewDriver(context.Background(), provider.NewHTTPHandleForTest(reg, conn, httpclient.DefaultClient(httpclient.WithSSRFProtection())), nil)
		require.NoError(t, err)
		assert.IsType(t, &drivers.OnePasswordDriver{}, drv)
	})

	t.Run("authorization_code uses SCIM-bridge driver", func(t *testing.T) {
		t.Parallel()

		raw, err := json.Marshal(&coredata.OnePasswordConnectorSettings{
			SCIMBridgeURL: "https://scim.example.test",
		})
		require.NoError(t, err)

		conn := &coredata.Connector{
			Provider:    coredata.ConnectorProviderOnePassword,
			RawSettings: raw,
			Connection: &connector.OAuth2Connection{
				GrantType: connector.OAuth2GrantTypeAuthorizationCode,
			},
		}

		drv, err := reg.NewDriver(context.Background(), provider.NewHTTPHandleForTest(reg, conn, httpclient.DefaultClient(httpclient.WithSSRFProtection())), nil)
		require.NoError(t, err)
		assert.IsType(t, &drivers.OnePasswordDriver{}, drv)
	})

	t.Run("authorization_code without scim_bridge_url errors", func(t *testing.T) {
		t.Parallel()

		conn := &coredata.Connector{
			Provider: coredata.ConnectorProviderOnePassword,
			Connection: &connector.OAuth2Connection{
				GrantType: connector.OAuth2GrantTypeAuthorizationCode,
			},
		}

		_, err := reg.NewDriver(context.Background(), provider.NewHTTPHandleForTest(reg, conn, httpclient.DefaultClient(httpclient.WithSSRFProtection())), nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "scim_bridge_url is required")
	})
}

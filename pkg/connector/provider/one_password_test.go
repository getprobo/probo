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
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
)

// TestOnePassword_NewDriver_DispatchByGrantType is the pre-merge gate
// for the 1Password closure. The OnePassword registration dispatches
// between two drivers based on the connector's OAuth2 grant type —
// this test asserts both paths construct without error from a
// coredata.Connector shaped for each grant type.
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

		drv, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
		require.NoError(t, err)
		assert.IsType(t, &drivers.OnePasswordUsersAPIDriver{}, drv)
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

		drv, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
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

		_, err := reg.NewDriver(context.Background(), httpclient.DefaultClient(httpclient.WithSSRFProtection()), conn, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "scim_bridge_url is required")
	})
}

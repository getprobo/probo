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

package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/cipher"
)

// Whether a credential rotated is what decides if Use writes the row back, and
// providers that rotate refresh tokens (HubSpot, DocuSign) hand back a new one
// while the access token still stands — so both are snapshotted at open.
func TestHandleCredentialRotated(t *testing.T) {
	t.Parallel()

	// An opener holding no OAuth app credentials keeps the stored tokens
	// rather than refreshing them, leaving each case to rotate what it means
	// to.
	open := func(t *testing.T, dbConnector *coredata.Connector) *Handle {
		t.Helper()

		registry := NewRegistry()
		require.NoError(t, registry.Register(&Registration{
			Provider:    coredata.ConnectorProviderSlack,
			DisplayName: "Slack",
		}))

		handle, err := NewOpener(nil, cipher.EncryptionKey{}, registry, nil, nil).
			Open(context.Background(), dbConnector)
		require.NoError(t, err)

		return handle
	}

	oauth2Connector := func(conn *connector.OAuth2Connection) *coredata.Connector {
		return &coredata.Connector{
			Provider:   coredata.ConnectorProviderSlack,
			Protocol:   coredata.ConnectorProtocolOAuth2,
			Connection: conn,
		}
	}

	t.Run("no token rotated", func(t *testing.T) {
		t.Parallel()

		handle := open(t, oauth2Connector(&connector.OAuth2Connection{
			AccessToken:  "at",
			RefreshToken: "rt",
		}))

		assert.False(t, handle.credentialRotated())
	})

	t.Run("access token rotated", func(t *testing.T) {
		t.Parallel()

		conn := &connector.OAuth2Connection{AccessToken: "at", RefreshToken: "rt"}

		handle := open(t, oauth2Connector(conn))
		conn.AccessToken = "at-2"

		assert.True(t, handle.credentialRotated())
	})

	t.Run("refresh token rotated alone", func(t *testing.T) {
		t.Parallel()

		conn := &connector.OAuth2Connection{AccessToken: "at", RefreshToken: "rt"}

		handle := open(t, oauth2Connector(conn))
		conn.RefreshToken = "rt-2"

		assert.True(t, handle.credentialRotated())
	})

	t.Run("api key connector carries no rotatable token", func(t *testing.T) {
		t.Parallel()

		handle := open(t, &coredata.Connector{
			Provider:   coredata.ConnectorProviderSlack,
			Protocol:   coredata.ConnectorProtocolAPIKey,
			Connection: &connector.APIKeyConnection{APIKey: "secret"},
		})

		assert.False(t, handle.credentialRotated())
	})

	// Nothing is stored for this protocol, so there is never anything to write
	// back after an open.
	t.Run("workload identity connector holds no credential to rotate", func(t *testing.T) {
		t.Parallel()

		handle := &Handle{
			Connector: &coredata.Connector{
				Provider:   coredata.ConnectorProviderSlack,
				Protocol:   coredata.ConnectorProtocolWorkloadIdentity,
				Connection: &connector.WorkloadIdentityConnection{},
			},
		}

		assert.False(t, handle.credentialRotated())
	})
}

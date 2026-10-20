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

package connectorprotocol_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connectorprotocol"
)

func TestConnectionCapabilities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		conn            connectorprotocol.Connection
		wantPicker      bool
		wantScopeGrant  bool
		wantReconnect   bool
		wantHTTP        bool
		wantProbeSuffix string
	}{
		{
			name:           "oauth2",
			conn:           &connectorprotocol.OAuth2Connection{},
			wantPicker:     true,
			wantScopeGrant: true,
			wantReconnect:  true,
			wantHTTP:       true,
		},
		{
			name: "oauth2 client credentials",
			conn: &connectorprotocol.OAuth2Connection{
				GrantType: connectorprotocol.OAuth2GrantTypeClientCredentials,
			},
			wantPicker:     true,
			wantScopeGrant: true,
			wantReconnect:  false,
			wantHTTP:       true,
		},
		{
			name:           "api key",
			conn:           &connectorprotocol.APIKeyConnection{},
			wantPicker:     true,
			wantScopeGrant: false,
			wantReconnect:  false,
			wantHTTP:       true,
		},
		{
			name:            "github app",
			conn:            &connectorprotocol.GitHubAppConnection{},
			wantPicker:      false,
			wantScopeGrant:  false,
			wantReconnect:   true,
			wantHTTP:        true,
			wantProbeSuffix: "/installation/repositories",
		},
		{
			name:           "slack",
			conn:           &connectorprotocol.SlackConnection{},
			wantPicker:     true,
			wantScopeGrant: true,
			wantReconnect:  true,
			wantHTTP:       true,
		},
		{
			name:           "workload identity",
			conn:           &connectorprotocol.WorkloadIdentityConnection{},
			wantPicker:     false,
			wantScopeGrant: false,
			wantReconnect:  false,
			wantHTTP:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantPicker, connectorprotocol.SupportsOrganizationPicker(tt.conn))
			assert.Equal(t, tt.wantScopeGrant, connectorprotocol.SupportsScopeGrantCheck(tt.conn))
			assert.Equal(t, tt.wantReconnect, connectorprotocol.SupportsReconnect(tt.conn))

			// Whether a credential rides on HTTP is answered by the type, not
			// by a capability method: only an HTTPConnection has a Client.
			_, isHTTP := tt.conn.(connectorprotocol.HTTPConnection)
			assert.Equal(t, tt.wantHTTP, isHTTP)

			probeURL, err := connectorprotocol.ResolveProbeURL(
				tt.conn,
				"https://api.github.com",
				"https://api.github.com/user",
			)
			require.NoError(t, err)

			if tt.wantProbeSuffix == "" {
				assert.Equal(t, "https://api.github.com/user", probeURL)
			} else {
				assert.Equal(t, "https://api.github.com"+tt.wantProbeSuffix, probeURL)
			}
		})
	}
}

func TestCapabilityProbeMatchesProtocol(t *testing.T) {
	t.Parallel()

	assert.True(t, connectorprotocol.SupportsOrganizationPickerForProtocol(connectorprotocol.ProtocolOAuth2))
	assert.False(t, connectorprotocol.SupportsOrganizationPickerForProtocol(connectorprotocol.ProtocolGitHubApp))
	assert.True(t, connectorprotocol.SupportsOrganizationPickerForProtocol(connectorprotocol.ProtocolAPIKey))
	assert.False(t, connectorprotocol.SupportsReconnectFor(nil, connectorprotocol.ProtocolOAuth2))
	assert.True(t, connectorprotocol.SupportsReconnectFor(nil, connectorprotocol.ProtocolGitHubApp))
	assert.False(t, connectorprotocol.SupportsScopeGrantCheckFor(nil, connectorprotocol.ProtocolGitHubApp))
	assert.False(t, connectorprotocol.SupportsOrganizationPickerForProtocol(connectorprotocol.ProtocolWorkloadIdentity))
	assert.False(t, connectorprotocol.SupportsReconnectFor(nil, connectorprotocol.ProtocolWorkloadIdentity))
}

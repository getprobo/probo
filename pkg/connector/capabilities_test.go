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

package connector_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/connector"
)

func TestConnectionCapabilities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		conn            connector.Connection
		wantPicker      bool
		wantScopeGrant  bool
		wantReconnect   bool
		wantProbeSuffix string
	}{
		{
			name:           "oauth2",
			conn:           &connector.OAuth2Connection{},
			wantPicker:     true,
			wantScopeGrant: true,
			wantReconnect:  true,
		},
		{
			name:           "api key",
			conn:           &connector.APIKeyConnection{},
			wantPicker:     false,
			wantScopeGrant: false,
			wantReconnect:  false,
		},
		{
			name:            "github app",
			conn:            &connector.GitHubAppConnection{},
			wantPicker:      false,
			wantScopeGrant:  false,
			wantReconnect:   true,
			wantProbeSuffix: "/installation/repositories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantPicker, connector.SupportsOrganizationPicker(tt.conn))
			assert.Equal(t, tt.wantScopeGrant, connector.SupportsScopeGrantCheck(tt.conn))
			assert.Equal(t, tt.wantReconnect, connector.SupportsReconnect(tt.conn))

			probeURL, err := connector.ResolveProbeURL(
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

	assert.True(t, connector.SupportsOrganizationPickerForProtocol(connector.ProtocolOAuth2))
	assert.False(t, connector.SupportsOrganizationPickerForProtocol(connector.ProtocolGitHubApp))
	assert.False(t, connector.SupportsOrganizationPickerForProtocol(connector.ProtocolAPIKey))
	assert.True(t, connector.SupportsReconnectFor(nil, connector.ProtocolGitHubApp))
	assert.False(t, connector.SupportsScopeGrantCheckFor(nil, connector.ProtocolGitHubApp))
}

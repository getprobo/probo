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

package oauth2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2scope"
)

func testCIMDScopeRegistry() *oauth2scope.Registry {
	return oauth2scope.NewRegistry().Register(
		map[coredata.OAuth2Scope][]string{
			"v1:asset":             {"core:asset:list"},
			"v1:business-function": {"core:business-function:list"},
		},
	)
}

func TestResolveCIMDAuthorizationScopes(t *testing.T) {
	t.Parallel()

	reg := testCIMDScopeRegistry()

	t.Run(
		"cimd stale request still offers advertised write scopes",
		func(t *testing.T) {
			t.Parallel()

			client := &coredata.OAuth2Client{
				ExternalClientID: "https://claude.ai/oauth/mcp-oauth-client-metadata",
				Scopes:           coredata.OAuth2Scopes{ScopeOpenID, "v1:asset"},
			}

			allowed, requested := resolveCIMDAuthorizationScopes(
				client,
				coredata.OAuth2Scopes{ScopeOpenID, "v1:asset"},
				reg,
			)

			require.True(t, allowed.ContainsAll(requested.Values()))
			assert.True(t, requested.Contains("v1:business-function"))
			assert.True(t, requested.Contains("v1:asset"))
		},
	)

	t.Run(
		"cimd omitted request uses advertised write scopes",
		func(t *testing.T) {
			t.Parallel()

			client := &coredata.OAuth2Client{
				ExternalClientID: "https://claude.ai/oauth/mcp-oauth-client-metadata",
				Scopes:           coredata.OAuth2Scopes{ScopeOpenID, "v1:asset"},
			}

			allowed, requested := resolveCIMDAuthorizationScopes(
				client,
				nil,
				reg,
			)

			require.True(t, allowed.ContainsAll(requested.Values()))
			assert.True(t, requested.Contains("v1:business-function"))
			assert.True(t, requested.Contains("v1:asset"))
			assert.True(t, requested.Contains(ScopeOpenID))
		},
	)

	t.Run(
		"public dcr client keeps explicit request",
		func(t *testing.T) {
			t.Parallel()

			client := &coredata.OAuth2Client{
				TokenEndpointAuthMethod: coredata.OAuth2ClientTokenEndpointAuthMethodNone,
				Scopes:                  coredata.OAuth2Scopes{ScopeOpenID, "v1:asset"},
			}

			allowed, requested := resolveCIMDAuthorizationScopes(
				client,
				coredata.OAuth2Scopes{ScopeOpenID, "v1:asset"},
				reg,
			)

			assert.Equal(t, client.Scopes, allowed)
			assert.Equal(t, coredata.OAuth2Scopes{ScopeOpenID, "v1:asset"}, requested)
			assert.False(t, requested.Contains("v1:business-function"))
		},
	)
}

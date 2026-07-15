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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2scope"
)

func TestIsCIMDClientID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		raw   string
		valid bool
	}{
		{
			name:  "https with path",
			raw:   "https://chatgpt.com/oauth/client.json",
			valid: true,
		},
		{
			name:  "gid is not cimd",
			raw:   "AAAAAAAAAAAASwAAAAAAAAAAcHJiY2xp",
			valid: false,
		},
		{
			name:  "http rejected",
			raw:   "http://chatgpt.com/oauth/client.json",
			valid: false,
		},
		{
			name:  "https root path rejected",
			raw:   "https://chatgpt.com/",
			valid: false,
		},
		{
			name:  "fragment rejected",
			raw:   "https://chatgpt.com/oauth/client.json#frag",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				assert.Equal(t, tt.valid, IsCIMDClientID(tt.raw))
			},
		)
	}
}

func TestCIMDAllowFromClientIDs(t *testing.T) {
	t.Parallel()

	clientID := "https://chatgpt.com/oauth/client.json"
	allow := CIMDAllowFromClientIDs(nil)

	allowance, err := allow(t.Context(), clientID)
	require.NoError(t, err)
	assert.Equal(t, CIMDAllowanceDenied, allowance)

	allow = CIMDAllowFromClientIDs([]string{})

	allowance, err = allow(t.Context(), clientID)
	require.NoError(t, err)
	assert.Equal(t, CIMDAllowanceDenied, allowance)

	allow = CIMDAllowFromClientIDs([]string{clientID})

	allowance, err = allow(t.Context(), clientID)
	require.NoError(t, err)
	assert.Equal(t, CIMDAllowanceAllowed, allowance)

	allow = CIMDAllowFromClientIDs([]string{"https://other.example.com/oauth/client.json"})

	allowance, err = allow(t.Context(), clientID)
	require.NoError(t, err)
	assert.Equal(t, CIMDAllowanceDenied, allowance)
}

func TestValidateClientMetadataDocument(t *testing.T) {
	t.Parallel()

	clientID := "https://mcp.example.com/oauth/metadata.json"
	doc := ClientMetadataDocument{
		ClientID:                clientID,
		ClientName:              "Example MCP",
		RedirectURIs:            []string{"https://mcp.example.com/callback"},
		TokenEndpointAuthMethod: "none",
	}

	require.NoError(t, validateClientMetadataDocument(clientID, &doc))

	t.Run(
		"http redirect on non-loopback rejected",
		func(t *testing.T) {
			t.Parallel()

			bad := doc
			bad.RedirectURIs = []string{"http://example.com/callback"}

			err := validateClientMetadataDocument(clientID, &bad)
			require.Error(t, err)
		},
	)

	t.Run(
		"http loopback redirect allowed",
		func(t *testing.T) {
			t.Parallel()

			loopback := doc
			loopback.RedirectURIs = []string{"http://127.0.0.1:3000/callback"}

			require.NoError(t, validateClientMetadataDocument(clientID, &loopback))
		},
	)

	t.Run(
		"mismatched client_id",
		func(t *testing.T) {
			t.Parallel()

			bad := doc
			bad.ClientID = "https://other.example.com/oauth/metadata.json"

			err := validateClientMetadataDocument(clientID, &bad)
			require.Error(t, err)
		},
	)
}

func TestCIMDFetcherFetch(t *testing.T) {
	t.Parallel()

	t.Run(
		"caches response with max-age",
		func(t *testing.T) {
			t.Parallel()

			doc := ClientMetadataDocument{
				ClientName:              "Test MCP Client",
				RedirectURIs:            []string{"http://127.0.0.1:3000/callback"},
				TokenEndpointAuthMethod: "none",
			}

			server := httptest.NewTLSServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, _ *http.Request) {
						w.Header().Set("Cache-Control", "max-age=60")
						_ = json.NewEncoder(w).Encode(doc)
					},
				),
			)
			t.Cleanup(server.Close)

			clientID := server.URL + "/oauth/client.json"
			doc.ClientID = clientID

			fetcher := &cimdFetcher{
				httpClient: server.Client(),
				logger:     log.NewLogger(),
			}

			fetched, err := fetcher.fetch(t.Context(), clientID)
			require.NoError(t, err)
			require.Equal(t, doc.ClientName, fetched.ClientName)

			cached, err := fetcher.fetch(t.Context(), clientID)
			require.NoError(t, err)
			require.Equal(t, fetched.ClientName, cached.ClientName)
		},
	)

	t.Run(
		"no-store response is not cached",
		func(t *testing.T) {
			t.Parallel()

			doc := ClientMetadataDocument{
				ClientName:              "Test MCP Client",
				RedirectURIs:            []string{"http://127.0.0.1:3000/callback"},
				TokenEndpointAuthMethod: "none",
			}

			requestCount := 0
			server := httptest.NewTLSServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, _ *http.Request) {
						requestCount++

						w.Header().Set("Cache-Control", "no-store")
						_ = json.NewEncoder(w).Encode(doc)
					},
				),
			)
			t.Cleanup(server.Close)

			clientID := server.URL + "/oauth/client.json"
			doc.ClientID = clientID

			fetcher := &cimdFetcher{
				httpClient: server.Client(),
				logger:     log.NewLogger(),
			}

			_, err := fetcher.fetch(t.Context(), clientID)
			require.NoError(t, err)
			_, err = fetcher.fetch(t.Context(), clientID)
			require.NoError(t, err)
			assert.Equal(t, 2, requestCount)
		},
	)
}

func TestCIMDScopes(t *testing.T) {
	t.Parallel()

	reg := oauth2scope.NewRegistry().Register(
		map[coredata.OAuth2Scope][]string{
			coredata.OAuth2Scope("v1:example:write"): {"example:write"},
		},
	)
	svc := &Service{registry: reg}

	t.Run(
		"defaults to all scopes when metadata omits scope",
		func(t *testing.T) {
			t.Parallel()

			scopes, err := svc.cimdScopes(&ClientMetadataDocument{})
			require.NoError(t, err)
			assert.Contains(t, scopes, ScopeOpenID)
			assert.Contains(t, scopes, coredata.OAuth2Scope("v1:example:write"))
		},
	)

	t.Run(
		"uses scope declared in metadata",
		func(t *testing.T) {
			t.Parallel()

			scopes, err := svc.cimdScopes(
				&ClientMetadataDocument{Scope: "openid profile email"},
			)
			require.NoError(t, err)
			assert.Equal(
				t,
				coredata.OAuth2Scopes{ScopeOpenID, ScopeProfile, ScopeEmail},
				scopes,
			)
		},
	)

	t.Run(
		"rejects unknown scope in metadata",
		func(t *testing.T) {
			t.Parallel()

			_, err := svc.cimdScopes(&ClientMetadataDocument{Scope: "admin"})
			require.Error(t, err)
		},
	)
}

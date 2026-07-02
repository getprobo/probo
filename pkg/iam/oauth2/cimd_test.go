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

package oauth2

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/log"
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

				assert.Equal(t, tt.valid, isCIMDClientID(tt.raw))
			},
		)
	}
}

func TestCIMDClientIDAllowed(t *testing.T) {
	t.Parallel()

	clientID := "https://chatgpt.com/oauth/client.json"

	assert.False(t, cimdClientIDAllowed(clientID, nil))
	assert.False(t, cimdClientIDAllowed(clientID, []string{}))
	assert.True(
		t,
		cimdClientIDAllowed(clientID, []string{clientID}),
	)
	assert.False(
		t,
		cimdClientIDAllowed(clientID, []string{"https://other.example.com/oauth/client.json"}),
	)
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

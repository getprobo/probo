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

package coredata_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/uri"
)

func TestOAuth2Client_IsRedirectURIAllowed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		registered []uri.URI
		requested  string
		want       bool
	}{
		{
			name:       "exact https match",
			registered: []uri.URI{"https://chatgpt.com/connector/oauth/callback"},
			requested:  "https://chatgpt.com/connector/oauth/callback",
			want:       true,
		},
		{
			name:       "https mismatch",
			registered: []uri.URI{"https://chatgpt.com/connector/oauth/callback"},
			requested:  "https://evil.example/callback",
			want:       false,
		},
		{
			name:       "loopback ignores port when registered without port",
			registered: []uri.URI{"http://localhost/callback"},
			requested:  "http://localhost:3118/callback",
			want:       true,
		},
		{
			name:       "loopback ignores port when registered with a different port",
			registered: []uri.URI{"http://127.0.0.1:53682/callback"},
			requested:  "http://127.0.0.1:8080/callback",
			want:       true,
		},
		{
			name:       "loopback host must match (localhost vs 127.0.0.1)",
			registered: []uri.URI{"http://127.0.0.1/callback"},
			requested:  "http://localhost:3118/callback",
			want:       false,
		},
		{
			name:       "loopback path must match",
			registered: []uri.URI{"http://localhost/callback"},
			requested:  "http://localhost:3118/other",
			want:       false,
		},
		{
			name:       "loopback scheme must match",
			registered: []uri.URI{"http://localhost/callback"},
			requested:  "https://localhost:3118/callback",
			want:       false,
		},
		{
			name:       "non-loopback host does not get port flexibility",
			registered: []uri.URI{"https://app.example.com/callback"},
			requested:  "https://app.example.com:8443/callback",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				client := &coredata.OAuth2Client{RedirectURIs: tt.registered}

				assert.Equal(t, tt.want, client.IsRedirectURIAllowed(tt.requested))
			},
		)
	}
}

func TestNewCIMDClient_WebURIs(t *testing.T) {
	t.Parallel()

	t.Run(
		"accepts https client_uri and logo_uri",
		func(t *testing.T) {
			t.Parallel()

			clientURI := "https://mcp.example.com"
			logoURI := "https://mcp.example.com/logo.png"

			client, err := coredata.NewCIMDClient(
				"https://mcp.example.com/oauth/metadata.json",
				"Example MCP",
				[]string{"https://mcp.example.com/callback"},
				nil,
				&logoURI,
				&clientURI,
				time.Now(),
			)
			require.NoError(t, err)
			require.NotNil(t, client.ClientURI)
			require.NotNil(t, client.LogoURI)
			assert.Equal(t, "https://mcp.example.com", client.ClientURI.String())
			assert.Equal(t, "https://mcp.example.com/logo.png", client.LogoURI.String())
		},
	)

	t.Run(
		"rejects non-web client_uri",
		func(t *testing.T) {
			t.Parallel()

			clientURI := "javascript://example.com/%0Aalert(1)"

			_, err := coredata.NewCIMDClient(
				"https://mcp.example.com/oauth/metadata.json",
				"Example MCP",
				[]string{"https://mcp.example.com/callback"},
				nil,
				nil,
				&clientURI,
				time.Now(),
			)
			require.Error(t, err)
		},
	)

	t.Run(
		"rejects non-web logo_uri",
		func(t *testing.T) {
			t.Parallel()

			logoURI := "data://example.com/image"

			_, err := coredata.NewCIMDClient(
				"https://mcp.example.com/oauth/metadata.json",
				"Example MCP",
				[]string{"https://mcp.example.com/callback"},
				nil,
				&logoURI,
				nil,
				time.Now(),
			)
			require.Error(t, err)
		},
	)
}

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

	"github.com/stretchr/testify/assert"
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

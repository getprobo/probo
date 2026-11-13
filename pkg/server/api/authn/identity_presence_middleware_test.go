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

package authn_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/server/api/authn"
)

func TestIdentityPresenceMiddleware_Challenge(t *testing.T) {
	t.Parallel()

	const (
		rootMetadataURL = "https://example.com/.well-known/oauth-protected-resource"
		mcpMetadataURL  = "https://example.com/.well-known/oauth-protected-resource/api/mcp/v1"
	)

	tests := []struct {
		name          string
		middleware    func(next http.Handler) http.Handler
		authorization string
		want          string
	}{
		{
			name:       "base url without credentials",
			middleware: authn.NewIdentityPresenceMiddleware(baseurl.MustParse("https://example.com")),
			want:       `Bearer resource_metadata="` + rootMetadataURL + `"`,
		},
		{
			name:          "base url with invalid bearer token",
			middleware:    authn.NewIdentityPresenceMiddleware(baseurl.MustParse("https://example.com")),
			authorization: "Bearer invalid",
			want:          `Bearer error="invalid_token", resource_metadata="` + rootMetadataURL + `"`,
		},
		{
			name:       "metadata url without credentials",
			middleware: authn.NewIdentityPresenceMiddlewareWithMetadataURL(mcpMetadataURL),
			want:       `Bearer resource_metadata="` + mcpMetadataURL + `"`,
		},
		{
			name:          "metadata url with invalid bearer token",
			middleware:    authn.NewIdentityPresenceMiddlewareWithMetadataURL(mcpMetadataURL),
			authorization: "Bearer invalid",
			want:          `Bearer error="invalid_token", resource_metadata="` + mcpMetadataURL + `"`,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				t.Parallel()

				next := http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						t.Error("next handler must not run without an identity")
					},
				)

				req := httptest.NewRequest(http.MethodPost, "/", nil)
				if tt.authorization != "" {
					req.Header.Set("Authorization", tt.authorization)
				}

				rec := httptest.NewRecorder()

				tt.middleware(next).ServeHTTP(rec, req)

				assert.Equal(t, http.StatusUnauthorized, rec.Code)
				assert.Equal(t, tt.want, rec.Header().Get("WWW-Authenticate"))
			},
		)
	}
}

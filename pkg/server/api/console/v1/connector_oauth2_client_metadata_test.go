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

package console_v1

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/baseurl"
)

// TestHandleConnectorOAuth2ClientMetadata verifies the public CIMD document:
// PostHog fetches it server-to-server during authorization, so client_id,
// redirect_uris (derived from the deployment base URL) and the public-client
// token_endpoint_auth_method must be exactly right or the OAuth flow breaks.
func TestHandleConnectorOAuth2ClientMetadata(t *testing.T) {
	t.Parallel()

	base, err := baseurl.Parse("https://probo.example.com")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	handleConnectorOAuth2ClientMetadata(base)(
		rec,
		httptest.NewRequest(http.MethodGet, "/api/console/v1/connectors/oauth-client-metadata", nil),
	)

	res := rec.Result()

	defer func() { _ = res.Body.Close() }()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var doc struct {
		ClientID                string   `json:"client_id"`
		ClientName              string   `json:"client_name"`
		ClientURI               string   `json:"client_uri"`
		LogoURI                 string   `json:"logo_uri"`
		RedirectURIs            []string `json:"redirect_uris"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
		GrantTypes              []string `json:"grant_types"`
		ResponseTypes           []string `json:"response_types"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&doc))

	// Functional fields are deployment-derived (must match where the OAuth
	// flow actually runs)...
	assert.Equal(t, "https://probo.example.com/api/console/v1/connectors/oauth-client-metadata", doc.ClientID)
	assert.Equal(t, []string{"https://probo.example.com/api/console/v1/connectors/complete"}, doc.RedirectURIs)
	// ...while the brand fields shown on the consent screen are the canonical
	// Probo product identity, NOT the per-tenant deployment URL.
	assert.Equal(t, "Probo", doc.ClientName)
	assert.Equal(t, "https://www.probo.com", doc.ClientURI)
	assert.Equal(t, "https://www.probo.com/probo-logo-only.svg", doc.LogoURI)
	assert.Equal(t, "none", doc.TokenEndpointAuthMethod, "public client must advertise token_endpoint_auth_method none")
	assert.Contains(t, doc.GrantTypes, "authorization_code")
	assert.Contains(t, doc.GrantTypes, "refresh_token")
	assert.Equal(t, []string{"code"}, doc.ResponseTypes)
}

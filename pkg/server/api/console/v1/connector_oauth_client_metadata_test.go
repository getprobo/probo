// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

// TestHandleConnectorOAuthClientMetadata verifies the public CIMD document:
// PostHog fetches it server-to-server during authorization, so client_id,
// redirect_uris (derived from the deployment base URL) and the public-client
// token_endpoint_auth_method must be exactly right or the OAuth flow breaks.
func TestHandleConnectorOAuthClientMetadata(t *testing.T) {
	t.Parallel()

	base, err := baseurl.Parse("https://probo.example.com")
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	handleConnectorOAuthClientMetadata(base)(
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
	assert.Equal(t, "https://www.getprobo.com", doc.ClientURI)
	assert.Equal(t, "https://www.getprobo.com/probo-logo-only.svg", doc.LogoURI)
	assert.Equal(t, "none", doc.TokenEndpointAuthMethod, "public client must advertise token_endpoint_auth_method none")
	assert.Contains(t, doc.GrantTypes, "authorization_code")
	assert.Contains(t, doc.GrantTypes, "refresh_token")
	assert.Equal(t, []string{"code"}, doc.ResponseTypes)
}

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

	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/connector"
)

// oauth2ClientMetadata is the OAuth2 Client ID Metadata Document (CIMD)
// published for public-client connectors. The deployment's
// (baseURL + CIMDMetadataPath) URL is the OAuth client_id; providers such as
// PostHog fetch this document server-to-server during authorization to learn
// the client's identity and allowed redirect URIs, so no app pre-registration
// is needed. Public clients authenticate with PKCE (token_endpoint_auth_method
// "none") rather than a client secret.
// Probo brand fields shown to the end user on the provider's consent screen.
// These describe the Probo product itself (not the per-tenant deployment), so
// they are the canonical brand homepage and logo rather than the baseURL.
const (
	proboBrandURI = "https://www.probo.com"
	proboLogoURI  = "https://www.probo.com/probo-logo-only.svg"
)

type oauth2ClientMetadata struct {
	ClientID                string   `json:"client_id"`
	ClientName              string   `json:"client_name"`
	ClientURI               string   `json:"client_uri"`
	LogoURI                 string   `json:"logo_uri"`
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
}

// handleConnectorOAuth2ClientMetadata serves the public, unauthenticated CIMD
// document. It is intentionally outside the auth middleware group: the OAuth2
// provider fetches it without any Probo credentials.
func handleConnectorOAuth2ClientMetadata(baseURL *baseurl.BaseURL) http.HandlerFunc {
	doc := oauth2ClientMetadata{
		ClientID:                baseURL.WithPath(connector.CIMDMetadataPath).MustString(),
		ClientName:              "Probo",
		ClientURI:               proboBrandURI,
		LogoURI:                 proboLogoURI,
		RedirectURIs:            []string{baseURL.WithPath(connector.CallbackPath).MustString()},
		TokenEndpointAuthMethod: "none",
		GrantTypes:              []string{"authorization_code", "refresh_token"},
		ResponseTypes:           []string{"code"},
	}

	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=300")
		_ = json.NewEncoder(w).Encode(doc)
	}
}

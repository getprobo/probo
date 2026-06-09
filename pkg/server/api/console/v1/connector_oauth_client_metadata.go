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

package console_v1

import (
	"encoding/json"
	"net/http"

	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/connector"
)

// oauthClientMetadata is the OAuth Client ID Metadata Document (CIMD)
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
	proboBrandURI = "https://www.getprobo.com"
	proboLogoURI  = "https://www.getprobo.com/probo-logo-only.svg"
)

type oauthClientMetadata struct {
	ClientID                string   `json:"client_id"`
	ClientName              string   `json:"client_name"`
	ClientURI               string   `json:"client_uri"`
	LogoURI                 string   `json:"logo_uri"`
	RedirectURIs            []string `json:"redirect_uris"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
}

// handleConnectorOAuthClientMetadata serves the public, unauthenticated CIMD
// document. It is intentionally outside the auth middleware group: the OAuth
// provider fetches it without any Probo credentials.
func handleConnectorOAuthClientMetadata(baseURL *baseurl.BaseURL) http.HandlerFunc {
	doc := oauthClientMetadata{
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

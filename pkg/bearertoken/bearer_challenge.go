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

package bearertoken

import (
	"fmt"
	"net/http"
	"strings"

	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	BearerErrInvalidToken      = "invalid_token"
	BearerErrInsufficientScope = "insufficient_scope"
)

func protectedResourceMetadataURL(baseURL *baseurl.BaseURL) string {
	return baseURL.WithPath("/.well-known/oauth-protected-resource").MustString()
}

// BearerChallenge builds the WWW-Authenticate header value per RFC 6750 and RFC 9728.
// Pass errorCode == "" for discovery-only (401, no Bearer attempt).
// Pass scopes only when errorCode == BearerErrInsufficientScope.
func BearerChallenge(
	baseURL *baseurl.BaseURL,
	errorCode string,
	scopes ...coredata.OAuth2Scope,
) string {
	metadataURL := protectedResourceMetadataURL(baseURL)

	var parts []string

	if errorCode != "" {
		parts = append(parts, fmt.Sprintf(`error="%s"`, errorCode))
	}

	if errorCode == BearerErrInsufficientScope && len(scopes) > 0 {
		scopeValues := make([]string, len(scopes))
		for i, scope := range scopes {
			scopeValues[i] = string(scope)
		}

		parts = append(parts, fmt.Sprintf(`scope="%s"`, strings.Join(scopeValues, " ")))
	}

	parts = append(parts, fmt.Sprintf(`resource_metadata="%s"`, metadataURL))

	return "Bearer " + strings.Join(parts, ", ")
}

func SetBearerChallenge(
	w http.ResponseWriter,
	baseURL *baseurl.BaseURL,
	errorCode string,
	scopes ...coredata.OAuth2Scope,
) {
	w.Header().Set("WWW-Authenticate", BearerChallenge(baseURL, errorCode, scopes...))
}

// SetBearerUnauthenticated sets a discovery-only challenge (RFC 9728 resource_metadata).
func SetBearerUnauthenticated(w http.ResponseWriter, baseURL *baseurl.BaseURL) {
	SetBearerChallenge(w, baseURL, "")
}

func SetBearerInvalidToken(w http.ResponseWriter, baseURL *baseurl.BaseURL) {
	SetBearerChallenge(w, baseURL, BearerErrInvalidToken)
}

func SetBearerInsufficientScope(
	w http.ResponseWriter,
	baseURL *baseurl.BaseURL,
	scopes ...coredata.OAuth2Scope,
) {
	SetBearerChallenge(w, baseURL, BearerErrInsufficientScope, scopes...)
}

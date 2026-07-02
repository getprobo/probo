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

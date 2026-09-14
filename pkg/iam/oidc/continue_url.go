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

package oidc

import (
	"net/url"

	iamoauth2 "go.probo.inc/probo/pkg/iam/oauth2"
)

const (
	oauth2AuthorizePath          = "/api/connect/v1/oauth2/authorize"
	signInSourceQueryKey         = "source"
	signInSourceCompliancePortal = "compliance-portal"
)

// allowsPersonalAccounts reports whether this OIDC login is finishing a
// compliance-portal authorize request. Console continue URLs, GID clients,
// and connector CIMD clients must not open the personal-account exception.
func allowsPersonalAccounts(continueURL string) bool {
	parsed, err := url.Parse(continueURL)
	if err != nil {
		return false
	}

	if parsed.Path != oauth2AuthorizePath {
		return false
	}

	query := parsed.Query()
	if query.Get(signInSourceQueryKey) != signInSourceCompliancePortal {
		return false
	}

	return iamoauth2.IsCIMDClientID(query.Get("client_id"))
}

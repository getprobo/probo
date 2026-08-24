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

package oauth2

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/uri"
)

// ProtectedResourceMetadata represents the RFC 9728 protected resource metadata
// document published at /.well-known/oauth-protected-resource.
type ProtectedResourceMetadata struct {
	Resource               uri.URI                `json:"resource"`
	AuthorizationServers   []uri.URI              `json:"authorization_servers"`
	BearerMethodsSupported []string               `json:"bearer_methods_supported"`
	ScopesSupported        []coredata.OAuth2Scope `json:"scopes_supported"`
}

func NewProtectedResourceMetadata(
	resource uri.URI,
	authorizationServer uri.URI,
	writeScopes []coredata.OAuth2Scope,
) *ProtectedResourceMetadata {
	return &ProtectedResourceMetadata{
		Resource:             resource,
		AuthorizationServers: []uri.URI{authorizationServer},
		BearerMethodsSupported: []string{
			"header",
		},
		ScopesSupported: protectedResourceScopes(writeScopes),
	}
}

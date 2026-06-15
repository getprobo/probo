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

package oauth2

import (
	"slices"

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
	apiScopes []coredata.OAuth2Scope,
) *ProtectedResourceMetadata {
	return &ProtectedResourceMetadata{
		Resource:             resource,
		AuthorizationServers: []uri.URI{authorizationServer},
		BearerMethodsSupported: []string{
			"header",
		},
		ScopesSupported: slices.Concat(
			[]coredata.OAuth2Scope{ScopeOpenID},
			apiScopes,
		),
	}
}

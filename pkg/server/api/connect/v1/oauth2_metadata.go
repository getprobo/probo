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

package connect_v1

import (
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2"
	"go.probo.inc/probo/pkg/uri"
)

const (
	oauth2AuthorizePath           = "/oauth2/authorize"
	oauth2TokenPath               = "/oauth2/token"
	oauth2UserinfoPath            = "/oauth2/userinfo"
	oauth2JWKSPath                = "/oauth2/jwks"
	oauth2RegisterPath            = "/oauth2/register"
	oauth2IntrospectPath          = "/oauth2/introspect"
	oauth2RevokePath              = "/oauth2/revoke"
	oauth2DeviceAuthorizationPath = "/oauth2/device"
)

func OAuth2ServerMetadata(
	baseURL *baseurl.BaseURL,
	registeredScopes []coredata.OAuth2Scope,
) *oauth2.ServerMetadata {
	return oauth2.NewMetadata(uri.URI(baseURL.String()), oauth2Endpoints(baseURL), registeredScopes)
}

func oauth2Endpoints(baseURL *baseurl.BaseURL) oauth2.Endpoints {
	return oauth2.Endpoints{
		Authorization:       uri.URI(baseURL.WithPath("/api/connect/v1" + oauth2AuthorizePath).MustString()),
		Token:               uri.URI(baseURL.WithPath("/api/connect/v1" + oauth2TokenPath).MustString()),
		Userinfo:            uri.URI(baseURL.WithPath("/api/connect/v1" + oauth2UserinfoPath).MustString()),
		JWKS:                uri.URI(baseURL.WithPath("/api/connect/v1" + oauth2JWKSPath).MustString()),
		Registration:        uri.URI(baseURL.WithPath("/api/connect/v1" + oauth2RegisterPath).MustString()),
		Introspection:       uri.URI(baseURL.WithPath("/api/connect/v1" + oauth2IntrospectPath).MustString()),
		Revocation:          uri.URI(baseURL.WithPath("/api/connect/v1" + oauth2RevokePath).MustString()),
		DeviceAuthorization: uri.URI(baseURL.WithPath("/api/connect/v1" + oauth2DeviceAuthorizationPath).MustString()),
	}
}

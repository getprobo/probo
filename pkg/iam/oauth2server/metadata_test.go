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

package oauth2server_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2server"
)

func TestNewMetadata(t *testing.T) {
	t.Parallel()

	issuer := "https://auth.example.com"
	endpoints := oauth2server.Endpoints{
		Authorization:       "https://auth.example.com/authorize",
		Token:               "https://auth.example.com/token",
		Userinfo:            "https://auth.example.com/userinfo",
		JWKS:                "https://auth.example.com/.well-known/jwks.json",
		Registration:        "https://auth.example.com/register",
		Introspection:       "https://auth.example.com/introspect",
		Revocation:          "https://auth.example.com/revoke",
		DeviceAuthorization: "https://auth.example.com/device",
	}

	metadata := oauth2server.NewMetadata(issuer, endpoints)
	require.NotNil(t, metadata)

	t.Run(
		"issuer",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, issuer, metadata.Issuer)
		},
	)

	t.Run(
		"endpoints",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, endpoints.Authorization, metadata.AuthorizationEndpoint)
			assert.Equal(t, endpoints.Token, metadata.TokenEndpoint)
			assert.Equal(t, endpoints.Userinfo, metadata.UserinfoEndpoint)
			assert.Equal(t, endpoints.JWKS, metadata.JwksURI)
			assert.Equal(t, endpoints.Registration, metadata.RegistrationEndpoint)
			assert.Equal(t, endpoints.Introspection, metadata.IntrospectionEndpoint)
			assert.Equal(t, endpoints.Revocation, metadata.RevocationEndpoint)
			assert.Equal(t, endpoints.DeviceAuthorization, metadata.DeviceAuthorizationEndpoint)
		},
	)

	t.Run(
		"scopes supported",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				[]coredata.OAuth2Scope{
					coredata.OAuth2ScopeOpenID,
					coredata.OAuth2ScopeProfile,
					coredata.OAuth2ScopeEmail,
				},
				metadata.ScopesSupported,
			)
		},
	)

	t.Run(
		"response types supported",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				[]coredata.OAuth2ResponseType{
					coredata.OAuth2ResponseTypeCode,
				},
				metadata.ResponseTypesSupported,
			)
		},
	)

	t.Run(
		"grant types supported",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				[]coredata.OAuth2GrantType{
					coredata.OAuth2GrantTypeAuthorizationCode,
					coredata.OAuth2GrantTypeRefreshToken,
					coredata.OAuth2GrantTypeDeviceCode,
				},
				metadata.GrantTypesSupported,
			)
		},
	)

	t.Run(
		"token endpoint auth methods supported",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				[]coredata.OAuth2ClientTokenEndpointAuthMethod{
					coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretBasic,
					coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretPost,
					coredata.OAuth2ClientTokenEndpointAuthMethodNone,
				},
				metadata.TokenEndpointAuthMethodsSupported,
			)
		},
	)

	t.Run(
		"subject types supported",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				[]coredata.OAuth2SubjectType{
					coredata.OAuth2SubjectTypePublic,
				},
				metadata.SubjectTypesSupported,
			)
		},
	)

	t.Run(
		"id token signing algorithms supported",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				[]coredata.OAuth2SigningAlgorithm{
					coredata.OAuth2SigningAlgorithmRS256,
				},
				metadata.IDTokenSigningAlgValuesSupported,
			)
		},
	)

	t.Run(
		"code challenge methods supported",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				[]coredata.OAuth2CodeChallengeMethod{
					coredata.OAuth2CodeChallengeMethodS256,
				},
				metadata.CodeChallengeMethodsSupported,
			)
		},
	)

	t.Run(
		"claims supported",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				[]coredata.OAuth2Claim{
					coredata.OAuth2ClaimIssuer,
					coredata.OAuth2ClaimSubject,
					coredata.OAuth2ClaimAudience,
					coredata.OAuth2ClaimExpiration,
					coredata.OAuth2ClaimIssuedAt,
					coredata.OAuth2ClaimAuthTime,
					coredata.OAuth2ClaimNonce,
					coredata.OAuth2ClaimAtHash,
					coredata.OAuth2ClaimEmail,
					coredata.OAuth2ClaimEmailVerified,
					coredata.OAuth2ClaimName,
				},
				metadata.ClaimsSupported,
			)
		},
	)
}

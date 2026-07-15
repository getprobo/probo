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

package oauth2_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.gearno.de/kit/httpclient"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam/oauth2"
	"go.probo.inc/probo/pkg/iam/oauth2scope"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/uri"
)

func TestFetchServerMetadata(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/.well-known/openid-configuration", r.URL.Path)

				_ = json.NewEncoder(w).Encode(
					map[string]string{
						"issuer":                 "https://auth.example.com",
						"authorization_endpoint": "https://auth.example.com/api/connect/v1/oauth2/authorize",
						"token_endpoint":         "https://auth.example.com/api/connect/v1/oauth2/token",
					},
				)
			},
		),
	)
	t.Cleanup(server.Close)

	client := httpclient.DefaultClient(
		httpclient.WithSSRFProtection(),
		httpclient.WithSSRFAllowLoopback(),
	)

	metadata, err := oauth2.FetchServerMetadata(context.Background(), client, server.URL)
	require.NoError(t, err)
	assert.Equal(
		t,
		uri.URI("https://auth.example.com/api/connect/v1/oauth2/authorize"),
		metadata.AuthorizationEndpoint,
	)
}

func TestAuthorizationURLWithQuery(t *testing.T) {
	t.Parallel()

	authorizationEndpoint := uri.URI("https://auth.example.com/api/connect/v1/oauth2/authorize")
	query := url.Values{}
	query.Set("client_id", "https://trust.example.com/.well-known/oauth-client-metadata")
	query.Set("response_type", "code")

	got, err := oauth2.AuthorizationURLWithQuery(authorizationEndpoint, query)
	require.NoError(t, err)
	assert.Equal(
		t,
		"https://auth.example.com/api/connect/v1/oauth2/authorize?client_id=https%3A%2F%2Ftrust.example.com%2F.well-known%2Foauth-client-metadata&response_type=code",
		got,
	)
}

func TestNewMetadata(t *testing.T) {
	t.Parallel()

	reg := oauth2scope.NewRegistry().Register(
		map[coredata.OAuth2Scope][]string{
			probo.ScopeV1DocumentRead: {"core:document:get"},
		},
	)

	issuer := uri.URI("https://auth.example.com")
	endpoints := oauth2.Endpoints{
		Authorization:       "https://auth.example.com/authorize",
		Token:               "https://auth.example.com/token",
		Userinfo:            "https://auth.example.com/userinfo",
		JWKS:                "https://auth.example.com/.well-known/jwks.json",
		Registration:        "https://auth.example.com/register",
		Introspection:       "https://auth.example.com/introspect",
		Revocation:          "https://auth.example.com/revoke",
		DeviceAuthorization: "https://auth.example.com/device",
	}

	metadata := oauth2.NewMetadata(issuer, endpoints, reg.RegisteredScopes())
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

			expectedScopes := slices.Concat(
				[]coredata.OAuth2Scope{
					oauth2.ScopeOpenID,
					oauth2.ScopeProfile,
					oauth2.ScopeEmail,
					oauth2.ScopeOfflineAccess,
				},
				reg.RegisteredScopes(),
			)

			assert.Equal(t, expectedScopes, metadata.ScopesSupported)
			assert.Contains(t, metadata.ScopesSupported, oauth2.ScopeOpenID)
			assert.Contains(t, metadata.ScopesSupported, probo.ScopeV1DocumentRead)
		},
	)

	t.Run(
		"protected resources",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, []uri.URI{issuer}, metadata.ProtectedResources)
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
		"revocation endpoint auth methods supported",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				[]coredata.OAuth2ClientTokenEndpointAuthMethod{
					coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretBasic,
					coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretPost,
					coredata.OAuth2ClientTokenEndpointAuthMethodNone,
				},
				metadata.RevocationEndpointAuthMethodsSupported,
			)
		},
	)

	t.Run(
		"introspection endpoint auth methods supported",
		func(t *testing.T) {
			t.Parallel()

			assert.Equal(
				t,
				[]coredata.OAuth2ClientTokenEndpointAuthMethod{
					coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretBasic,
					coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretPost,
					coredata.OAuth2ClientTokenEndpointAuthMethodNone,
				},
				metadata.IntrospectionEndpointAuthMethodsSupported,
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

	t.Run(
		"cimd supported",
		func(t *testing.T) {
			t.Parallel()

			assert.True(t, metadata.ClientIDMetadataDocumentSupported)
		},
	)
}

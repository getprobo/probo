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

package oauth2server

import (
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/uri"
)

type (
	// ServerMetadata represents the OpenID Connect Discovery 1.0 / RFC 8414
	// authorization server metadata document.
	ServerMetadata struct {
		Issuer                                    uri.URI                                        `json:"issuer"`
		AuthorizationEndpoint                     uri.URI                                        `json:"authorization_endpoint"`
		TokenEndpoint                             uri.URI                                        `json:"token_endpoint"`
		UserinfoEndpoint                          uri.URI                                        `json:"userinfo_endpoint"`
		JwksURI                                   uri.URI                                        `json:"jwks_uri"`
		RegistrationEndpoint                      uri.URI                                        `json:"registration_endpoint"`
		IntrospectionEndpoint                     uri.URI                                        `json:"introspection_endpoint"`
		RevocationEndpoint                        uri.URI                                        `json:"revocation_endpoint"`
		DeviceAuthorizationEndpoint               uri.URI                                        `json:"device_authorization_endpoint"`
		ScopesSupported                           []coredata.OAuth2Scope                         `json:"scopes_supported"`
		ResponseTypesSupported                    []coredata.OAuth2ResponseType                  `json:"response_types_supported"`
		GrantTypesSupported                       []coredata.OAuth2GrantType                     `json:"grant_types_supported"`
		TokenEndpointAuthMethodsSupported         []coredata.OAuth2ClientTokenEndpointAuthMethod `json:"token_endpoint_auth_methods_supported"`
		RevocationEndpointAuthMethodsSupported    []coredata.OAuth2ClientTokenEndpointAuthMethod `json:"revocation_endpoint_auth_methods_supported"`
		IntrospectionEndpointAuthMethodsSupported []coredata.OAuth2ClientTokenEndpointAuthMethod `json:"introspection_endpoint_auth_methods_supported"`
		SubjectTypesSupported                     []coredata.OAuth2SubjectType                   `json:"subject_types_supported"`
		IDTokenSigningAlgValuesSupported          []coredata.OAuth2SigningAlgorithm              `json:"id_token_signing_alg_values_supported"`
		CodeChallengeMethodsSupported             []coredata.OAuth2CodeChallengeMethod           `json:"code_challenge_methods_supported"`
		ClaimsSupported                           []coredata.OAuth2Claim                         `json:"claims_supported"`
	}

	// Endpoints holds the endpoint URLs for the OIDC discovery document.
	Endpoints struct {
		Authorization       uri.URI
		Token               uri.URI
		Userinfo            uri.URI
		JWKS                uri.URI
		Registration        uri.URI
		Introspection       uri.URI
		Revocation          uri.URI
		DeviceAuthorization uri.URI
	}
)

func NewMetadata(issuer uri.URI, endpoints Endpoints) *ServerMetadata {
	return &ServerMetadata{
		Issuer:                      issuer,
		AuthorizationEndpoint:       endpoints.Authorization,
		TokenEndpoint:               endpoints.Token,
		UserinfoEndpoint:            endpoints.Userinfo,
		JwksURI:                     endpoints.JWKS,
		RegistrationEndpoint:        endpoints.Registration,
		IntrospectionEndpoint:       endpoints.Introspection,
		RevocationEndpoint:          endpoints.Revocation,
		DeviceAuthorizationEndpoint: endpoints.DeviceAuthorization,
		ScopesSupported: []coredata.OAuth2Scope{
			coredata.OAuth2ScopeOpenID,
			coredata.OAuth2ScopeProfile,
			coredata.OAuth2ScopeEmail,
			coredata.OAuth2ScopeOfflineAccess,
		},
		ResponseTypesSupported: []coredata.OAuth2ResponseType{
			coredata.OAuth2ResponseTypeCode,
		},
		GrantTypesSupported: []coredata.OAuth2GrantType{
			coredata.OAuth2GrantTypeAuthorizationCode,
			coredata.OAuth2GrantTypeRefreshToken,
			coredata.OAuth2GrantTypeDeviceCode,
		},
		TokenEndpointAuthMethodsSupported: []coredata.OAuth2ClientTokenEndpointAuthMethod{
			coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretBasic,
			coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretPost,
			coredata.OAuth2ClientTokenEndpointAuthMethodNone,
		},
		RevocationEndpointAuthMethodsSupported: []coredata.OAuth2ClientTokenEndpointAuthMethod{
			coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretBasic,
			coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretPost,
			coredata.OAuth2ClientTokenEndpointAuthMethodNone,
		},
		IntrospectionEndpointAuthMethodsSupported: []coredata.OAuth2ClientTokenEndpointAuthMethod{
			coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretBasic,
			coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretPost,
			coredata.OAuth2ClientTokenEndpointAuthMethodNone,
		},
		SubjectTypesSupported: []coredata.OAuth2SubjectType{
			coredata.OAuth2SubjectTypePublic,
		},
		IDTokenSigningAlgValuesSupported: []coredata.OAuth2SigningAlgorithm{
			coredata.OAuth2SigningAlgorithmRS256,
		},
		CodeChallengeMethodsSupported: []coredata.OAuth2CodeChallengeMethod{
			coredata.OAuth2CodeChallengeMethodS256,
		},
		ClaimsSupported: []coredata.OAuth2Claim{
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
	}
}

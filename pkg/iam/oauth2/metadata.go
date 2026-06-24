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
		ProtectedResources                        []uri.URI                                      `json:"protected_resources,omitempty"`
		ResponseTypesSupported                    []coredata.OAuth2ResponseType                  `json:"response_types_supported"`
		GrantTypesSupported                       []coredata.OAuth2GrantType                     `json:"grant_types_supported"`
		TokenEndpointAuthMethodsSupported         []coredata.OAuth2ClientTokenEndpointAuthMethod `json:"token_endpoint_auth_methods_supported"`
		RevocationEndpointAuthMethodsSupported    []coredata.OAuth2ClientTokenEndpointAuthMethod `json:"revocation_endpoint_auth_methods_supported"`
		IntrospectionEndpointAuthMethodsSupported []coredata.OAuth2ClientTokenEndpointAuthMethod `json:"introspection_endpoint_auth_methods_supported"`
		SubjectTypesSupported                     []coredata.OAuth2SubjectType                   `json:"subject_types_supported"`
		IDTokenSigningAlgValuesSupported          []coredata.OAuth2SigningAlgorithm              `json:"id_token_signing_alg_values_supported"`
		CodeChallengeMethodsSupported             []coredata.OAuth2CodeChallengeMethod           `json:"code_challenge_methods_supported"`
		ClaimsSupported                           []coredata.OAuth2Claim                         `json:"claims_supported"`
		ClientIDMetadataDocumentSupported         bool                                           `json:"client_id_metadata_document_supported"`
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

func NewMetadata(issuer uri.URI, endpoints Endpoints, registeredScopes []coredata.OAuth2Scope) *ServerMetadata {
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
		ScopesSupported:             authorizationServerScopes(registeredScopes),
		ProtectedResources:          []uri.URI{issuer},
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
		ClientIDMetadataDocumentSupported: true,
	}
}

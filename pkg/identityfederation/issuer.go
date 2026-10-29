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

// Package identityfederation makes Probo an OIDC identity provider for outbound cloud
// federation: it mints short-lived RS256 tokens asserting "Probo, acting for
// organization X" and publishes the discovery and JWKS documents a cloud
// provider fetches to verify them.
//
// The direction is outbound. Inbound identity (users logging into Probo) lives
// under pkg/iam. The two never share signing keys or issuer URLs.
//
// There is no token endpoint. Tokens are minted in-process and never traverse
// HTTP.
package identityfederation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probodconfig"
)

const (
	// DefaultTokenTTL bounds how long a minted token stays usable. Cloud
	// providers only ever see it during a single credential exchange, so it is
	// kept deliberately short.
	DefaultTokenTTL = 5 * time.Minute

	jwksPathSegment = "jwks"
)

type (
	// Issuer mints identity federation tokens and produces the public documents that
	// describe how to verify them.
	Issuer struct {
		baseURL  *baseurl.BaseURL
		keyRing  *jose.KeyRing
		tokenTTL time.Duration
		logger   *log.Logger
	}
)

// NewIssuer builds the outbound issuer from cfg. A disabled configuration
// returns a nil issuer. logger receives the startup record and each minted
// assertion. appBaseURL is required so an empty cfg.IssuerBaseURL can derive
// {appBaseURL}/federation.
func NewIssuer(
	logger *log.Logger,
	cfg probodconfig.IdentityFederationConfig,
	appBaseURL *baseurl.BaseURL,
) (*Issuer, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	if len(cfg.SigningKeys) == 0 {
		return nil, fmt.Errorf("cannot create identity federation issuer: at least one signing key is required")
	}

	signingKeys := make([]jose.SigningKey, 0, len(cfg.SigningKeys))

	for _, keyCfg := range cfg.SigningKeys {
		kid := keyCfg.KID
		if kid == "" {
			kid = "default"
		}

		signingKeys = append(signingKeys, jose.SigningKey{
			PrivateKey: keyCfg.PrivateKey.PrivateKey(),
			KID:        kid,
			Active:     keyCfg.Active,
		})
	}

	keyRing, err := jose.NewKeyRing(signingKeys)
	if err != nil {
		return nil, fmt.Errorf("cannot create identity federation issuer: %w", err)
	}

	issuerBaseURL, err := ResolveIssuerBaseURL(cfg.IssuerBaseURL, appBaseURL)
	if err != nil {
		return nil, fmt.Errorf("cannot create identity federation issuer: %w", err)
	}

	issuer := &Issuer{
		baseURL:  issuerBaseURL,
		keyRing:  keyRing,
		tokenTTL: DefaultTokenTTL,
		logger:   logger,
	}

	// probod cannot verify that the configured apex actually reaches it, so the
	// effective value is logged for the startup record and covered by a canary.
	logger.Info(
		"identity federation issuer configured",
		log.String("issuer_base_url", issuer.BaseURL()),
	)

	// A derived issuer is not held to the provider rules, so localhost and CI
	// still start. Warn instead of failing, rather than staying silent.
	if err := ValidateConfig(issuerBaseURL); err != nil {
		logger.Warn(
			"identity federation issuer cannot be registered with one or more cloud providers; customers cannot install the connector until it is publicly reachable over https without a port",
			log.String("issuer_base_url", issuer.BaseURL()),
			log.Error(err),
		)
	}

	return issuer, nil
}

// BaseURL returns the advertised issuer base URL.
func (i *Issuer) BaseURL() string {
	return i.baseURL.String()
}

// IssuerURL returns the issuer URL of one organization. This is the string a
// customer registers with their cloud provider, and it is immutable once they
// have done so.
func (i *Issuer) IssuerURL(organizationID gid.GID) (string, error) {
	return issuerURL(i.baseURL, organizationID)
}

// issuerURL builds the per-organization issuer. Startup validation measures its
// output, so the length ceiling cannot drift from what customers register.
func issuerURL(baseURL *baseurl.BaseURL, organizationID gid.GID) (string, error) {
	joined, err := url.JoinPath(
		baseURL.String(),
		url.PathEscape(organizationID.String()),
	)
	if err != nil {
		return "", fmt.Errorf("cannot build identity federation issuer URL: %w", err)
	}

	return joined, nil
}

// JWKSURI returns the key set URL of one organization. The key set is identical
// for every organization; the per-organization path exists so the URI sits
// beneath its own issuer, as OIDC discovery expects.
func (i *Issuer) JWKSURI(organizationID gid.GID) (string, error) {
	jwksURI, err := url.JoinPath(
		i.baseURL.String(),
		url.PathEscape(organizationID.String()),
		jwksPathSegment,
	)
	if err != nil {
		return "", fmt.Errorf("cannot build identity federation jwks URI: %w", err)
	}

	return jwksURI, nil
}

// Token mints a token asserting that Probo acts for the given organization.
//
// The organization must come from the connector row being serviced, never from
// user input. The returned token is a bearer credential for the customer's
// cloud account: never log it and never place it in an error message.
func (i *Issuer) Token(
	ctx context.Context,
	organizationID gid.GID,
	audience string,
) (string, error) {
	if organizationID == gid.Nil {
		return "", fmt.Errorf("cannot mint identity federation token: organization is required")
	}

	if organizationID.EntityType() != coredata.OrganizationEntityType {
		return "", fmt.Errorf("cannot mint identity federation token: identifier is not an organization")
	}

	if audience == "" {
		return "", fmt.Errorf("cannot mint identity federation token: audience is required")
	}

	issuerURL, err := i.IssuerURL(organizationID)
	if err != nil {
		return "", fmt.Errorf("cannot mint identity federation token: %w", err)
	}

	jti, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("cannot generate identity federation token id: %w", err)
	}

	now := time.Now()

	claims := Claims{
		Issuer:    issuerURL,
		Subject:   organizationID.String(),
		Audience:  audience,
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: now.Add(i.tokenTTL).Unix(),
		JTI:       jti.String(),
	}

	token, err := i.keyRing.Sign(claims)
	if err != nil {
		return "", fmt.Errorf("cannot sign identity federation token: %w", err)
	}

	i.logger.InfoCtx(
		ctx,
		"minted identity federation assertion",
		tokenFields(token, i.keyRing.JWKS(), claims)...,
	)

	return token, nil
}

// Metadata returns the OIDC discovery document of one organization's issuer.
//
// Every field comes from configuration. None of it is ever derived from an
// inbound request, which would let an arbitrary Host produce a document
// claiming to be that issuer.
func (i *Issuer) Metadata(organizationID gid.GID) (*Metadata, error) {
	issuerURL, err := i.IssuerURL(organizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot build identity federation metadata: %w", err)
	}

	jwksURI, err := i.JWKSURI(organizationID)
	if err != nil {
		return nil, fmt.Errorf("cannot build identity federation metadata jwks URI: %w", err)
	}

	return &Metadata{
		Issuer:                           issuerURL,
		JWKSURI:                          jwksURI,
		ResponseTypesSupported:           []string{"id_token"},
		SubjectTypesSupported:            []string{"public"},
		IDTokenSigningAlgValuesSupported: []string{"RS256"},
	}, nil
}

// JWKS returns the published key set. Inactive keys are included so that a
// token signed before a rotation still verifies.
func (i *Issuer) JWKS() *jose.JWKS {
	return i.keyRing.JWKS()
}

// tokenFields describes a minted assertion for the debug log. Claims are the
// values that were signed. alg, typ, and kid are read from the token header.
// The compact token also supplies the segment lengths, and the published key
// set checks the signature. The signature and the raw token are never returned.
func tokenFields(token string, jwks *jose.JWKS, claims Claims) []log.Attr {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return []log.Attr{
			log.String("decode_error", "cannot decode identity federation assertion"),
		}
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return []log.Attr{
			log.String("decode_error", "cannot decode identity federation assertion"),
		}
	}

	var header jose.JWTHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return []log.Attr{
			log.String("decode_error", "cannot decode identity federation assertion"),
		}
	}

	_, verifyErr := jose.VerifyJWTWithJWKS(token, jwks)

	fields := []log.Attr{
		log.String("alg", header.Algorithm),
		log.String("typ", header.Type),
		log.String("kid", header.KeyID),
		log.String("iss", claims.Issuer),
		log.String("sub", claims.Subject),
		log.String("aud", claims.Audience),
		log.Int64("iat", claims.IssuedAt),
		log.Int64("nbf", claims.NotBefore),
		log.Int64("exp", claims.ExpiresAt),
		log.String("jti", claims.JTI),
		log.Int("header_length", len(parts[0])),
		log.Int("payload_length", len(parts[1])),
		log.Int("signature_length", len(parts[2])),
		log.Bool("signature_verified", verifyErr == nil),
	}

	if verifyErr != nil {
		fields = append(fields, log.String("verify_error", verifyErr.Error()))
	}

	return fields
}

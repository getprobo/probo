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
	"fmt"
	"net/url"
	"time"

	"go.gearno.de/crypto/uuid"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/gid"
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
	}
)

// NewIssuer returns an issuer advertising baseURL and signing with keyRing.
func NewIssuer(
	baseURL *baseurl.BaseURL,
	keyRing *jose.KeyRing,
	tokenTTL time.Duration,
) (*Issuer, error) {
	if baseURL == nil {
		return nil, fmt.Errorf("cannot create identity federation issuer: base URL is required")
	}

	if keyRing == nil {
		return nil, fmt.Errorf("cannot create identity federation issuer: key ring is required")
	}

	if tokenTTL <= 0 {
		return nil, fmt.Errorf("cannot create identity federation issuer: token TTL must be positive")
	}

	return &Issuer{
		baseURL:  baseURL,
		keyRing:  keyRing,
		tokenTTL: tokenTTL,
	}, nil
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

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
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/uri"
)

type (
	// SigningKey pairs an RSA private key with its key ID. All entries are
	// published in the JWKS endpoint. Keys with Active set to true are
	// used for signing new tokens; when multiple keys are active, the
	// service round-robins between them.
	SigningKey struct {
		PrivateKey *rsa.PrivateKey
		KID        string
		Active     bool
	}

	IDTokenClaims struct {
		Issuer        uri.URI               `json:"iss"`
		Subject       string                `json:"sub"`
		Audience      string                `json:"aud"`
		ExpiresAt     int64                 `json:"exp"`
		IssuedAt      int64                 `json:"iat"`
		AuthTime      int64                 `json:"auth_time"`
		Nonce         string                `json:"nonce,omitempty"`
		AtHash        string                `json:"at_hash,omitempty"`
		Email         string                `json:"email,omitempty"`
		EmailVerified *bool                 `json:"email_verified,omitempty"`
		Name          string                `json:"name,omitempty"`
		Scope         coredata.OAuth2Scopes `json:"-"`
	}

	SigningKeys []SigningKey
)

// ComputeAtHash computes the at_hash claim value for an access token.
// Per OIDC Core §3.1.3.6: left half of SHA-256 hash, base64url-encoded.
func ComputeAtHash(accessToken string) string {
	h := sha256.Sum256([]byte(accessToken))
	return base64.RawURLEncoding.EncodeToString(h[:16])
}

func NewIDTokenClaims(
	issuer uri.URI,
	identityID gid.GID,
	clientID gid.GID,
	authTime time.Time,
	scopes coredata.OAuth2Scopes,
	nonce string,
	accessToken string,
	email string,
	emailVerified bool,
	fullName string,
	ttl time.Duration,
) *IDTokenClaims {
	now := time.Now()

	claims := &IDTokenClaims{
		Issuer:    issuer,
		Subject:   identityID.String(),
		Audience:  clientID.String(),
		ExpiresAt: now.Add(ttl).Unix(),
		IssuedAt:  now.Unix(),
		AuthTime:  authTime.Unix(),
		Scope:     scopes,
	}

	if nonce != "" {
		claims.Nonce = nonce
	}

	if accessToken != "" {
		claims.AtHash = ComputeAtHash(accessToken)
	}

	for _, scope := range scopes {
		switch scope {
		case ScopeEmail:
			claims.Email = email
			claims.EmailVerified = &emailVerified
		case ScopeProfile:
			claims.Name = fullName
		}
	}

	return claims
}

func ParseIDTokenClaims(raw string) (*IDTokenClaims, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("cannot parse id token: invalid format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("cannot parse id token payload: %w", err)
	}

	var claims IDTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("cannot decode id token claims: %w", err)
	}

	return &claims, nil
}

func ParseIDTokenIdentity(
	raw string,
	jwks *jose.JWKS,
	expectedNonce string,
	expectedIssuer uri.URI,
	expectedAudience string,
) (gid.GID, error) {
	if raw == "" {
		return gid.GID{}, fmt.Errorf("cannot parse id token: missing token")
	}

	payload, err := jose.VerifyJWTWithJWKS(raw, jwks)
	if err != nil {
		return gid.GID{}, fmt.Errorf("cannot verify id token: %w", err)
	}

	var claims IDTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return gid.GID{}, fmt.Errorf("cannot decode id token claims: %w", err)
	}

	if claims.Issuer != expectedIssuer {
		return gid.GID{}, fmt.Errorf("cannot validate issuer: unexpected issuer %q", claims.Issuer)
	}

	if claims.Audience != expectedAudience {
		return gid.GID{}, fmt.Errorf("cannot validate audience: mismatch")
	}

	if claims.Nonce != expectedNonce {
		return gid.GID{}, fmt.Errorf("cannot validate nonce: mismatch")
	}

	if time.Now().After(time.Unix(claims.ExpiresAt, 0)) {
		return gid.GID{}, fmt.Errorf("cannot validate id token: token has expired")
	}

	identityID, err := gid.ParseGID(claims.Subject)
	if err != nil {
		return gid.GID{}, fmt.Errorf("cannot parse identity from id token: %w", err)
	}

	return identityID, nil
}

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
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"go.probo.inc/probo/pkg/coredata"
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
		case coredata.OAuth2ScopeEmail:
			claims.Email = email
			claims.EmailVerified = &emailVerified
		case coredata.OAuth2ScopeProfile:
			claims.Name = fullName
		}
	}

	return claims
}

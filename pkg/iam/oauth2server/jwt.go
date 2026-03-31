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
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
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

	// IDTokenClaims represents the claims included in an OIDC ID token.
	IDTokenClaims struct {
		Issuer        string                `json:"iss"`
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

	// JWTHeader represents a JWT header.
	JWTHeader struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
		KeyID     string `json:"kid"`
	}

	// JWK represents a JSON Web Key.
	JWK struct {
		KeyType   string `json:"kty"`
		Use       string `json:"use"`
		Algorithm string `json:"alg"`
		KeyID     string `json:"kid"`
		N         string `json:"n"`
		E         string `json:"e"`
	}

	// JWKS represents a JSON Web Key Set.
	JWKS struct {
		Keys []JWK `json:"keys"`
	}

	// SigningKeys is a collection of signing keys.
	SigningKeys []SigningKey
)

// SignIDToken signs an ID token with the given signing key.
func SignIDToken(sk *SigningKey, claims *IDTokenClaims) (string, error) {
	header := JWTHeader{
		Algorithm: "RS256",
		Type:      "JWT",
		KeyID:     sk.KID,
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("cannot marshal jwt header: %w", err)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("cannot marshal jwt claims: %w", err)
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signingInput := headerB64 + "." + claimsB64

	h := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, sk.PrivateKey, crypto.SHA256, h[:])
	if err != nil {
		return "", fmt.Errorf("cannot sign jwt: %w", err)
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureB64, nil
}

// ComputeAtHash computes the at_hash claim value for an access token.
// Per OIDC Core §3.1.3.6: left half of SHA-256 hash, base64url-encoded.
func ComputeAtHash(accessToken string) string {
	h := sha256.Sum256([]byte(accessToken))
	return base64.RawURLEncoding.EncodeToString(h[:16])
}

// NewIDTokenClaims creates ID token claims for the given parameters.
func NewIDTokenClaims(
	issuer string,
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

// PublicJWK returns the public JWK representation of the signing key.
func (k *SigningKey) PublicJWK() JWK {
	return JWK{
		KeyType:   "RSA",
		Use:       "sig",
		Algorithm: "RS256",
		KeyID:     k.KID,
		N:         base64.RawURLEncoding.EncodeToString(k.PrivateKey.N.Bytes()),
		E:         base64.RawURLEncoding.EncodeToString(big.NewInt(int64(k.PrivateKey.E)).Bytes()),
	}
}

// PublicJWKS returns the JWKS containing the public keys for all signing keys.
func (keys SigningKeys) PublicJWKS() *JWKS {
	jwks := &JWKS{
		Keys: make([]JWK, 0, len(keys)),
	}

	for i := range keys {
		jwks.Keys = append(jwks.Keys, keys[i].PublicJWK())
	}

	return jwks
}

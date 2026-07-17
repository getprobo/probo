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

package jose

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

type (
	// JWK represents a JSON Web Key (RFC 7517).
	JWK struct {
		KeyType   string `json:"kty"`
		Use       string `json:"use"`
		Algorithm string `json:"alg"`
		KeyID     string `json:"kid"`
		N         string `json:"n"`
		E         string `json:"e"`
	}

	// JWKS represents a JSON Web Key Set (RFC 7517).
	JWKS struct {
		Keys []JWK `json:"keys"`
	}

	// JWTHeader represents a JWT header (RFC 7519).
	JWTHeader struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
		KeyID     string `json:"kid"`
	}
)

// RSAPublicKeyToJWK converts an RSA public key to a JWK with the given
// key ID, marked for RS256 signature use.
func RSAPublicKeyToJWK(pub *rsa.PublicKey, kid string) JWK {
	return JWK{
		KeyType:   "RSA",
		Use:       "sig",
		Algorithm: "RS256",
		KeyID:     kid,
		N:         base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		E:         base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
	}
}

// SignJWT signs arbitrary claims as a JWT using RS256 with the given RSA
// private key and key ID. The claims value is JSON-marshaled as the payload.
func SignJWT(privateKey *rsa.PrivateKey, kid string, claims any) (string, error) {
	header := JWTHeader{
		Algorithm: "RS256",
		Type:      "JWT",
		KeyID:     kid,
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

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, h[:])
	if err != nil {
		return "", fmt.Errorf("cannot sign jwt: %w", err)
	}

	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return signingInput + "." + signatureB64, nil
}

// RSAPublicKeyFromJWK reconstructs an RSA public key from a JWK.
func RSAPublicKeyFromJWK(jwk JWK) (*rsa.PublicKey, error) {
	if jwk.KeyType != "RSA" {
		return nil, fmt.Errorf("cannot convert jwk to rsa public key: unsupported key type %q", jwk.KeyType)
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("cannot decode rsa modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("cannot decode rsa exponent: %w", err)
	}

	e := new(big.Int).SetBytes(eBytes)
	if !e.IsInt64() || e.Sign() <= 0 {
		return nil, fmt.Errorf("cannot convert jwk to rsa public key: invalid rsa exponent")
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(e.Int64()),
	}, nil
}

// PublicKeyFromJWKS returns the RSA public key matching the given key ID.
func PublicKeyFromJWKS(jwks *JWKS, kid string) (*rsa.PublicKey, error) {
	if jwks == nil {
		return nil, fmt.Errorf("cannot find signing key %q in jwks: jwks is nil", kid)
	}

	for _, key := range jwks.Keys {
		if key.KeyID == kid {
			return RSAPublicKeyFromJWK(key)
		}
	}

	return nil, fmt.Errorf("cannot find signing key %q in jwks", kid)
}

// VerifyJWT verifies an RS256 JWT signature and returns the decoded payload.
func VerifyJWT(raw string, pubKey *rsa.PublicKey) ([]byte, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("cannot verify jwt: invalid format")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("cannot decode jwt header: %w", err)
	}

	var header JWTHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("cannot parse jwt header: %w", err)
	}

	if header.Algorithm != "RS256" {
		return nil, fmt.Errorf("cannot verify jwt: unsupported algorithm %q", header.Algorithm)
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("cannot decode jwt payload: %w", err)
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("cannot decode jwt signature: %w", err)
	}

	signedContent := parts[0] + "." + parts[1]
	hash := sha256.Sum256([]byte(signedContent))

	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], signature); err != nil {
		return nil, fmt.Errorf("cannot verify jwt signature: %w", err)
	}

	return payload, nil
}

// VerifyJWTWithJWKS verifies an RS256 JWT using the matching key from a JWKS.
func VerifyJWTWithJWKS(raw string, jwks *JWKS) ([]byte, error) {
	if jwks == nil {
		return nil, fmt.Errorf("cannot verify jwt: jwks is nil")
	}

	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("cannot verify jwt: invalid format")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("cannot decode jwt header: %w", err)
	}

	var header JWTHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("cannot parse jwt header: %w", err)
	}

	if header.KeyID == "" {
		return nil, fmt.Errorf("cannot verify jwt: missing key id")
	}

	pubKey, err := PublicKeyFromJWKS(jwks, header.KeyID)
	if err != nil {
		return nil, err
	}

	return VerifyJWT(raw, pubKey)
}

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

// Package emailverification verifies the Email Verification Token (EVT) and
// Key Binding JWT (KB-JWT) that a browser submits through the Email
// Verification API, as specified by draft-hardt-email-verification.
package emailverification

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"

	"go.probo.inc/probo/pkg/dnsclient"
	"go.probo.inc/probo/pkg/mail"
)

type (
	Verifier struct {
		lookupTXT  func(ctx context.Context, name string) ([]string, error)
		httpClient *http.Client
		now        func() time.Time
	}

	Option func(*Verifier)

	// Request is a combined EVT~KB-JWT submitted for an email address.
	Request struct {
		Token    string
		Email    string
		Audience string
		// ValidateNonce accepts the KB-JWT nonce only when the relying party
		// issued it. It must be set.
		ValidateNonce func(nonce string) error
	}

	jwtHeader struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
		Typ string `json:"typ"`
	}

	jwk struct {
		Kty string `json:"kty"`
		Crv string `json:"crv"`
		Kid string `json:"kid"`
		X   string `json:"x"`
		Y   string `json:"y"`
		Alg string `json:"alg"`
	}

	confirmation struct {
		JWK *jwk `json:"jwk"`
	}

	evtClaims struct {
		Iss           string        `json:"iss"`
		IAT           int64         `json:"iat"`
		CNF           *confirmation `json:"cnf"`
		Email         string        `json:"email"`
		EmailVerified bool          `json:"email_verified"`
	}

	kbClaims struct {
		Aud    string `json:"aud"`
		Nonce  string `json:"nonce"`
		IAT    int64  `json:"iat"`
		SDHash string `json:"sd_hash"`
	}

	issuerMetadata struct {
		Issuer                    string   `json:"issuer"`
		JWKSURI                   string   `json:"jwks_uri"`
		SigningAlgValuesSupported []string `json:"signing_alg_values_supported"`
	}

	jwks struct {
		Keys []jwk `json:"keys"`
	}

	compactJWS struct {
		header       jwtHeader
		payload      []byte
		signingInput []byte
		signature    []byte
	}
)

const (
	algEd25519 = "Ed25519"
	algES256   = "ES256"

	evtType = "evt+jwt"
	kbType  = "kb+jwt"

	dnsRecordPrefix = "_email-verification."
	issuerPrefix    = "iss="
	metadataPath    = ".well-known/email-verification"

	maxTokenBytes    = 16 << 10
	maxResponseBytes = 64 << 10
	requestTimeout   = 10 * time.Second
	maxTokenAge      = 15 * time.Minute
	maxClockSkew     = 5 * time.Minute
)

var (
	errSignatureInvalid = errors.New("signature does not verify")
	supportedAlgs       = []string{algEd25519, algES256}
)

func NewVerifier(resolverAddr string, opts ...Option) *Verifier {
	dns := dnsclient.NewClient(resolverAddr)

	v := &Verifier{
		lookupTXT:  dns.LookupTXT,
		httpClient: httpclient.DefaultPooledClient(httpclient.WithSSRFProtection()),
		now:        time.Now,
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

func WithLookupTXT(fn func(ctx context.Context, name string) ([]string, error)) Option {
	return func(v *Verifier) {
		v.lookupTXT = fn
	}
}

func WithHTTPClient(c *http.Client) Option {
	return func(v *Verifier) {
		v.httpClient = c
	}
}

func WithClock(now func() time.Time) Option {
	return func(v *Verifier) {
		v.now = now
	}
}

// Verify returns nil only when the token proves that the browser holds a
// verified email address equal to req.Email, bound to req.Audience and to a
// nonce the relying party issued. Callers treat any error as "not verified"
// and fall back to the regular confirmation flow.
func (v *Verifier) Verify(ctx context.Context, req Request) error {
	if req.ValidateNonce == nil {
		return errors.New("cannot verify email token: nonce validator is required")
	}

	addr, err := mail.ParseAddr(req.Email)
	if err != nil || addr.String() != req.Email {
		return fmt.Errorf("cannot verify email token: invalid email address")
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	evtRaw, kbRaw, err := splitToken(req.Token)
	if err != nil {
		return err
	}

	evt, err := parseCompactJWS(evtRaw)
	if err != nil {
		return fmt.Errorf("cannot parse email verification token: %w", err)
	}

	kb, err := parseCompactJWS(kbRaw)
	if err != nil {
		return fmt.Errorf("cannot parse key binding token: %w", err)
	}

	var claims evtClaims
	if err := json.Unmarshal(evt.payload, &claims); err != nil {
		return fmt.Errorf("cannot parse email verification token claims: %w", err)
	}

	var kbBody kbClaims
	if err := json.Unmarshal(kb.payload, &kbBody); err != nil {
		return fmt.Errorf("cannot parse key binding token claims: %w", err)
	}

	if err := v.checkEVTClaims(evt.header, claims, req.Email); err != nil {
		return err
	}

	cnf := claims.CNF.JWK

	if kb.header.Typ != kbType {
		return fmt.Errorf("cannot verify key binding token: unexpected typ")
	}

	if kb.header.Alg != cnf.Alg {
		return fmt.Errorf("cannot verify key binding token: alg does not match confirmation key")
	}

	if kbBody.Aud != req.Audience {
		return fmt.Errorf("cannot verify key binding token: audience mismatch")
	}

	if err := req.ValidateNonce(kbBody.Nonce); err != nil {
		return fmt.Errorf("cannot verify key binding token: nonce rejected: %w", err)
	}

	if err := v.checkFreshness(claims.IAT); err != nil {
		return fmt.Errorf("cannot verify email verification token: %w", err)
	}

	if err := v.checkFreshness(kbBody.IAT); err != nil {
		return fmt.Errorf("cannot verify key binding token: %w", err)
	}

	sdHash := sha256.Sum256([]byte(evtRaw + "~"))
	if kbBody.SDHash != base64.RawURLEncoding.EncodeToString(sdHash[:]) {
		return fmt.Errorf("cannot verify key binding token: sd_hash mismatch")
	}

	issuer, err := v.issuerForEmail(ctx, req.Email)
	if err != nil {
		return err
	}

	if claims.Iss != issuer {
		return fmt.Errorf("cannot verify email verification token: issuer is not authorized for the email domain")
	}

	metadata, err := v.fetchMetadata(ctx, issuer)
	if err != nil {
		return err
	}

	if !metadata.supportsAlg(evt.header.Alg) {
		return fmt.Errorf("cannot verify email verification token: alg not supported by issuer")
	}

	keys, err := v.fetchJWKS(ctx, metadata.JWKSURI)
	if err != nil {
		return err
	}

	issuerKey, err := signingKeyByKid(keys, evt.header.Kid, evt.header.Alg)
	if err != nil {
		return err
	}

	if err := verifySignature(evt.header.Alg, issuerKey, evt.signingInput, evt.signature); err != nil {
		return fmt.Errorf("cannot verify email verification token: %w", err)
	}

	confirmationKey, err := publicKeyForAlg(cnf.Alg, *cnf)
	if err != nil {
		return fmt.Errorf("cannot verify key binding token: %w", err)
	}

	if err := verifySignature(kb.header.Alg, confirmationKey, kb.signingInput, kb.signature); err != nil {
		return fmt.Errorf("cannot verify key binding token: %w", err)
	}

	return nil
}

func (v *Verifier) checkEVTClaims(header jwtHeader, claims evtClaims, email string) error {
	if header.Typ != evtType {
		return fmt.Errorf("cannot verify email verification token: unexpected typ")
	}

	if !slices.Contains(supportedAlgs, header.Alg) {
		return fmt.Errorf("cannot verify email verification token: unsupported alg")
	}

	if header.Kid == "" {
		return fmt.Errorf("cannot verify email verification token: missing kid")
	}

	if !claims.EmailVerified {
		return fmt.Errorf("cannot verify email verification token: email not verified")
	}

	if claims.Email != email {
		return fmt.Errorf("cannot verify email verification token: email mismatch")
	}

	if claims.CNF == nil || claims.CNF.JWK == nil {
		return fmt.Errorf("cannot verify email verification token: missing confirmation key")
	}

	if !slices.Contains(supportedAlgs, claims.CNF.JWK.Alg) {
		return fmt.Errorf("cannot verify email verification token: unsupported confirmation alg")
	}

	return nil
}

func (v *Verifier) checkFreshness(iat int64) error {
	if iat <= 0 {
		return errors.New("missing iat")
	}

	issued := time.Unix(iat, 0)
	now := v.now()

	if issued.After(now.Add(maxClockSkew)) {
		return errors.New("issued in the future")
	}

	if now.Sub(issued) > maxTokenAge {
		return errors.New("token is too old")
	}

	return nil
}

func (v *Verifier) issuerForEmail(ctx context.Context, email string) (string, error) {
	at := strings.LastIndex(email, "@")
	domain := strings.ToLower(email[at+1:])

	// The DNS error names the queried domain, which is PII-adjacent, so it is
	// deliberately not wrapped.
	values, err := v.lookupTXT(ctx, dnsRecordPrefix+domain)
	if err != nil {
		return "", errors.New("cannot look up email verification record")
	}

	if len(values) != 1 {
		return "", fmt.Errorf("cannot look up email verification record: expected one record, found %d", len(values))
	}

	host, ok := strings.CutPrefix(strings.TrimSpace(values[0]), issuerPrefix)
	if !ok {
		return "", errors.New("cannot look up email verification record: missing iss= prefix")
	}

	return issuerIdentifier(host)
}

// issuerIdentifier derives the HTTPS origin of a DNS-published host. The
// host must be a bare host name without port, path, or credentials.
func issuerIdentifier(host string) (string, error) {
	u, err := url.Parse("https://" + host)
	if err != nil || u.Host != host || u.Port() != "" || u.Path != "" || u.User != nil {
		return "", fmt.Errorf("cannot derive issuer identifier: invalid host")
	}

	return u.String(), nil
}

func (m issuerMetadata) supportsAlg(alg string) bool {
	algs := m.SigningAlgValuesSupported
	if len(algs) == 0 {
		algs = []string{algEd25519}
	}

	return slices.Contains(algs, alg)
}

func (v *Verifier) fetchMetadata(ctx context.Context, issuer string) (*issuerMetadata, error) {
	endpoint, err := url.JoinPath(issuer, metadataPath)
	if err != nil {
		return nil, fmt.Errorf("cannot build metadata URL: %w", err)
	}

	var metadata issuerMetadata
	if err := v.getJSON(ctx, endpoint, &metadata); err != nil {
		return nil, fmt.Errorf("cannot fetch issuer metadata: %w", err)
	}

	if metadata.Issuer != issuer {
		return nil, errors.New("cannot fetch issuer metadata: issuer mismatch")
	}

	jwksURL, err := url.Parse(metadata.JWKSURI)
	if err != nil || jwksURL.Scheme != "https" || jwksURL.Host == "" {
		return nil, errors.New("cannot fetch issuer metadata: invalid jwks_uri")
	}

	return &metadata, nil
}

func (v *Verifier) fetchJWKS(ctx context.Context, jwksURI string) (*jwks, error) {
	var keys jwks
	if err := v.getJSON(ctx, jwksURI, &keys); err != nil {
		return nil, fmt.Errorf("cannot fetch issuer keys: %w", err)
	}

	return &keys, nil
}

func (v *Verifier) getJSON(ctx context.Context, endpoint string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("cannot create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot send request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return fmt.Errorf("cannot read response: %w", err)
	}

	if len(body) > maxResponseBytes {
		return errors.New("response too large")
	}

	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("cannot parse response: %w", err)
	}

	return nil
}

func splitToken(token string) (string, string, error) {
	if len(token) > maxTokenBytes {
		return "", "", errors.New("cannot split email verification token: token too large")
	}

	evt, kb, ok := strings.Cut(token, "~")
	if !ok || evt == "" || kb == "" || strings.Contains(kb, "~") {
		return "", "", errors.New("cannot split email verification token: expected EVT~KB-JWT")
	}

	return evt, kb, nil
}

func parseCompactJWS(raw string) (*compactJWS, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid compact serialization")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("cannot decode header: %w", err)
	}

	var header jwtHeader
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("cannot parse header: %w", err)
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("cannot decode payload: %w", err)
	}

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("cannot decode signature: %w", err)
	}

	return &compactJWS{
		header:       header,
		payload:      payload,
		signingInput: []byte(parts[0] + "." + parts[1]),
		signature:    signature,
	}, nil
}

func signingKeyByKid(keys *jwks, kid string, alg string) (crypto.PublicKey, error) {
	for _, k := range keys.Keys {
		if k.Kid == kid {
			return publicKeyForAlg(alg, k)
		}
	}

	return nil, fmt.Errorf("cannot find signing key %q", kid)
}

func publicKeyForAlg(alg string, k jwk) (crypto.PublicKey, error) {
	switch alg {
	case algEd25519:
		if k.Kty != "OKP" || k.Crv != "Ed25519" {
			return nil, errors.New("key does not match Ed25519")
		}

		x, err := base64.RawURLEncoding.DecodeString(k.X)
		if err != nil || len(x) != ed25519.PublicKeySize {
			return nil, errors.New("invalid Ed25519 public key")
		}

		return ed25519.PublicKey(x), nil

	case algES256:
		if k.Kty != "EC" || k.Crv != "P-256" {
			return nil, errors.New("key does not match ES256")
		}

		x, errX := base64.RawURLEncoding.DecodeString(k.X)

		y, errY := base64.RawURLEncoding.DecodeString(k.Y)
		if errX != nil || errY != nil || len(x) != 32 || len(y) != 32 {
			return nil, errors.New("invalid P-256 public key")
		}

		point := append(append([]byte{0x04}, x...), y...)

		pub, err := ecdsa.ParseUncompressedPublicKey(elliptic.P256(), point)
		if err != nil {
			return nil, fmt.Errorf("invalid P-256 public key: %w", err)
		}

		return pub, nil

	default:
		return nil, fmt.Errorf("unsupported alg %q", alg)
	}
}

func verifySignature(alg string, key crypto.PublicKey, signingInput []byte, signature []byte) error {
	switch alg {
	case algEd25519:
		pub, ok := key.(ed25519.PublicKey)
		if !ok {
			return errors.New("key is not Ed25519")
		}

		if !ed25519.Verify(pub, signingInput, signature) {
			return errSignatureInvalid
		}

		return nil

	case algES256:
		pub, ok := key.(*ecdsa.PublicKey)
		if !ok {
			return errors.New("key is not ECDSA")
		}

		if len(signature) != 64 {
			return errors.New("invalid ES256 signature length")
		}

		digest := sha256.Sum256(signingInput)
		r := new(big.Int).SetBytes(signature[:32])
		s := new(big.Int).SetBytes(signature[32:])

		if !ecdsa.Verify(pub, digest[:], r, s) {
			return errSignatureInvalid
		}

		return nil

	default:
		return fmt.Errorf("unsupported alg %q", alg)
	}
}

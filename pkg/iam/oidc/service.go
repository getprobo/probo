// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package oidc

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/mail"
	"golang.org/x/oauth2"
)

var cryptoRandReader io.Reader = cryptorand.Reader

type (
	ProviderConfig struct {
		ClientID     string
		ClientSecret string
		Enabled      bool
	}

	providerInfo struct {
		oauth2Config      oauth2.Config
		jwksURL           string
		issuerValidator   func(string) bool
		enterpriseChecker func(*idTokenClaims) bool

		// trustProviderEmail indicates that the provider's directory
		// guarantees email verification for enterprise accounts, so
		// the email_verified claim does not need to be present in
		// the ID token.
		trustProviderEmail bool
	}

	UserInfo struct {
		Email    mail.Addr
		FullName string
	}

	Service struct {
		pg         *pg.Client
		baseURL    string
		logger     *log.Logger
		httpClient *http.Client
		providers  map[coredata.OIDCProvider]*providerInfo

		jwksMu    sync.RWMutex
		jwksCache map[string]*jwksEntry
	}

	jwksEntry struct {
		keys      []jwk
		fetchedAt time.Time
	}

	jwk struct {
		Kty string `json:"kty"`
		Kid string `json:"kid"`
		Use string `json:"use"`
		N   string `json:"n"`
		E   string `json:"e"`
		Crv string `json:"crv"`
		X   string `json:"x"`
		Y   string `json:"y"`
	}

	jwksResponse struct {
		Keys []jwk `json:"keys"`
	}

	idTokenClaims struct {
		Issuer        string  `json:"iss"`
		Subject       string  `json:"sub"`
		Audience      any     `json:"aud"`
		ExpiresAt     float64 `json:"exp"`
		Nonce         string  `json:"nonce"`
		Email         string  `json:"email"`
		EmailVerified any     `json:"email_verified"`
		Name          string  `json:"name"`
		HostedDomain  string  `json:"hd"`
	}
)

func (c *idTokenClaims) hasAudience(clientID string) bool {
	switch aud := c.Audience.(type) {
	case string:
		return aud == clientID
	case []any:
		for _, v := range aud {
			if s, ok := v.(string); ok && s == clientID {
				return true
			}
		}
	}
	return false
}

func (c *idTokenClaims) isEmailVerified() bool {
	switch v := c.EmailVerified.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true")
	}
	return false
}

var (
	googleEndpoint = oauth2.Endpoint{
		AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL: "https://oauth2.googleapis.com/token",
	}

	microsoftEndpoint = oauth2.Endpoint{
		AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
		TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
	}
)

const (
	googleJWKSURL             = "https://www.googleapis.com/oauth2/v3/certs"
	microsoftJWKSURL          = "https://login.microsoftonline.com/common/discovery/v2.0/keys"
	microsoftConsumerTenantID = "9188040d-6c67-4c5b-b112-36a304b66dad"
	jwksCacheTTL              = 1 * time.Hour
)

func NewService(
	pgClient *pg.Client,
	baseURL string,
	google ProviderConfig,
	microsoft ProviderConfig,
	logger *log.Logger,
) *Service {
	s := &Service{
		pg:         pgClient,
		baseURL:    baseURL,
		logger:     logger.Named("oidc"),
		httpClient: httpclient.DefaultPooledClient(httpclient.WithLogger(logger)),
		providers:  make(map[coredata.OIDCProvider]*providerInfo),
		jwksCache:  make(map[string]*jwksEntry),
	}

	if google.Enabled {
		s.providers[coredata.OIDCProviderGoogle] = &providerInfo{
			oauth2Config: oauth2.Config{
				ClientID:     google.ClientID,
				ClientSecret: google.ClientSecret,
				Endpoint:     googleEndpoint,
				RedirectURL:  baseURL + "/api/connect/v1/oidc/google/callback",
				Scopes:       []string{"openid", "email", "profile"},
			},
			jwksURL: googleJWKSURL,
			issuerValidator: func(iss string) bool {
				return iss == "https://accounts.google.com"
			},
			enterpriseChecker: func(claims *idTokenClaims) bool {
				// The "hd" (hosted domain) claim is only present for
				// Google Workspace accounts. Personal gmail.com accounts
				// do not have this claim.
				return claims.HostedDomain != ""
			},
		}
	}

	if microsoft.Enabled {
		s.providers[coredata.OIDCProviderMicrosoft] = &providerInfo{
			oauth2Config: oauth2.Config{
				ClientID:     microsoft.ClientID,
				ClientSecret: microsoft.ClientSecret,
				Endpoint:     microsoftEndpoint,
				RedirectURL:  baseURL + "/api/connect/v1/oidc/microsoft/callback",
				Scopes:       []string{"openid", "email", "profile"},
			},
			jwksURL:            microsoftJWKSURL,
			trustProviderEmail: true,
			issuerValidator: func(iss string) bool {
				return strings.HasPrefix(iss, "https://login.microsoftonline.com/") &&
					strings.HasSuffix(iss, "/v2.0")
			},
			enterpriseChecker: func(claims *idTokenClaims) bool {
				// Personal Microsoft accounts (live.com, outlook.com,
				// hotmail.com) use the consumer tenant ID. Reject them.
				return !strings.Contains(
					claims.Issuer,
					microsoftConsumerTenantID,
				)
			},
		}
	}

	return s
}

func (s *Service) Run(ctx context.Context) error {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(context.Canceled)

	gcCtx, stopGC := context.WithCancel(context.WithoutCancel(ctx))
	gc := NewGarbageCollector(s.pg, s.logger)
	wg.Go(
		func() {
			if err := gc.Run(gcCtx); err != nil {
				cancel(fmt.Errorf("oidc garbage collector crashed: %w", err))
			}
		},
	)

	<-ctx.Done()

	stopGC()

	wg.Wait()

	return context.Cause(ctx)
}

func (s *Service) IsProviderEnabled(provider coredata.OIDCProvider) bool {
	_, ok := s.providers[provider]
	return ok
}

func (s *Service) EnabledProviders() []coredata.OIDCProvider {
	providers := make([]coredata.OIDCProvider, 0, len(s.providers))
	for p := range s.providers {
		providers = append(providers, p)
	}
	slices.Sort(providers)
	return providers
}

func (s *Service) InitiateLogin(
	ctx context.Context,
	provider coredata.OIDCProvider,
	continueURL string,
) (string, error) {
	info, ok := s.providers[provider]
	if !ok {
		return "", NewProviderNotEnabledError(provider)
	}

	state, err := generateRandomString(32)
	if err != nil {
		return "", fmt.Errorf("cannot generate state: %w", err)
	}

	nonce, err := generateRandomString(32)
	if err != nil {
		return "", fmt.Errorf("cannot generate nonce: %w", err)
	}

	codeVerifier, err := generateRandomString(64)
	if err != nil {
		return "", fmt.Errorf("cannot generate code verifier: %w", err)
	}

	now := time.Now()
	oidcState := &coredata.OIDCState{
		ID:           state,
		Provider:     provider,
		Nonce:        nonce,
		CodeVerifier: codeVerifier,
		ContinueURL:  continueURL,
		CreatedAt:    now,
		ExpiresAt:    now.Add(10 * time.Minute),
	}

	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := oidcState.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot store oidc state: %w", err)
			}
			return nil
		},
	)
	if err != nil {
		return "", err
	}

	codeChallenge := computeCodeChallenge(codeVerifier)

	authURL := info.oauth2Config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	return authURL, nil
}

func (s *Service) HandleCallback(
	ctx context.Context,
	provider coredata.OIDCProvider,
	stateParam string,
	code string,
) (*coredata.Identity, string, error) {
	info, ok := s.providers[provider]
	if !ok {
		return nil, "", NewProviderNotEnabledError(provider)
	}

	var oidcState coredata.OIDCState
	err := s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			if err := oidcState.LoadByIDForUpdate(ctx, tx, stateParam); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewInvalidStateError()
				}
				return fmt.Errorf("cannot load oidc state: %w", err)
			}

			if err := oidcState.Delete(ctx, tx); err != nil {
				return fmt.Errorf("cannot delete oidc state: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, "", err
	}

	if time.Now().After(oidcState.ExpiresAt) {
		return nil, "", NewInvalidStateError()
	}

	if oidcState.Provider != provider {
		return nil, "", NewInvalidStateError()
	}

	token, err := info.oauth2Config.Exchange(
		ctx,
		code,
		oauth2.SetAuthURLParam("code_verifier", oidcState.CodeVerifier),
	)
	if err != nil {
		return nil, "", NewCodeExchangeError(err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, "", NewIDTokenMissingError()
	}

	claims, err := s.verifyAndParseIDToken(ctx, info, rawIDToken, oidcState.Nonce)
	if err != nil {
		return nil, "", fmt.Errorf("cannot verify id token: %w", err)
	}

	if claims.Email == "" {
		return nil, "", NewMissingEmailClaimError()
	}

	if !info.trustProviderEmail && !claims.isEmailVerified() {
		return nil, "", NewEmailNotVerifiedError()
	}

	if !info.enterpriseChecker(claims) {
		return nil, "", NewPersonalAccountNotAllowedError()
	}

	email, err := mail.ParseAddr(claims.Email)
	if err != nil {
		return nil, "", fmt.Errorf("cannot parse email from id token: %w", err)
	}

	var identity *coredata.Identity
	now := time.Now()

	err = s.pg.WithTx(
		ctx,
		func(tx pg.Conn) error {
			identity = &coredata.Identity{}
			err := identity.LoadByEmail(ctx, tx, email)
			if err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load identity by email: %w", err)
				}

				identity = &coredata.Identity{
					ID:                   gid.New(gid.NilTenant, coredata.IdentityEntityType),
					EmailAddress:         email,
					FullName:             claims.Name,
					EmailAddressVerified: true,
					CreatedAt:            now,
					UpdatedAt:            now,
				}

				if err := identity.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot insert identity: %w", err)
				}

				return nil
			}

			if !identity.EmailAddressVerified {
				identity.EmailAddressVerified = true
				identity.UpdatedAt = now

				if err := identity.Update(ctx, tx); err != nil {
					return fmt.Errorf("cannot update identity: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, "", err
	}

	return identity, oidcState.ContinueURL, nil
}

func (s *Service) verifyAndParseIDToken(ctx context.Context, info *providerInfo, rawIDToken string, expectedNonce string) (*idTokenClaims, error) {
	parts := strings.Split(rawIDToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("cannot parse id token: invalid format")
	}

	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("cannot decode id token header: %w", err)
	}

	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, fmt.Errorf("cannot parse id token header: %w", err)
	}

	key, err := s.getSigningKey(ctx, info.jwksURL, header.Kid)
	if err != nil {
		return nil, fmt.Errorf("cannot get signing key: %w", err)
	}

	signedContent := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("cannot decode signature: %w", err)
	}

	if err := verifySignature(header.Alg, key, []byte(signedContent), signature); err != nil {
		return nil, fmt.Errorf("cannot verify id token signature: %w", err)
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("cannot decode id token payload: %w", err)
	}

	var claims idTokenClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, fmt.Errorf("cannot parse id token claims: %w", err)
	}

	if !info.issuerValidator(claims.Issuer) {
		return nil, fmt.Errorf("cannot validate issuer: unexpected issuer %q", claims.Issuer)
	}

	if !claims.hasAudience(info.oauth2Config.ClientID) {
		return nil, fmt.Errorf("cannot validate audience: expected %q", info.oauth2Config.ClientID)
	}

	if claims.Nonce != expectedNonce {
		return nil, fmt.Errorf("cannot validate nonce: mismatch")
	}

	if time.Now().After(time.Unix(int64(claims.ExpiresAt), 0)) {
		return nil, fmt.Errorf("cannot validate id token: token has expired")
	}

	return &claims, nil
}

func (s *Service) getSigningKey(ctx context.Context, jwksURL string, kid string) (crypto.PublicKey, error) {
	s.jwksMu.RLock()
	entry, ok := s.jwksCache[jwksURL]
	s.jwksMu.RUnlock()

	if !ok || time.Since(entry.fetchedAt) > jwksCacheTTL {
		keys, err := fetchJWKS(ctx, s.httpClient, jwksURL)
		if err != nil {
			return nil, err
		}

		entry = &jwksEntry{keys: keys, fetchedAt: time.Now()}
		s.jwksMu.Lock()
		s.jwksCache[jwksURL] = entry
		s.jwksMu.Unlock()
	}

	for _, k := range entry.keys {
		if k.Kid == kid {
			return parseJWK(k)
		}
	}

	// Key not found in cache, try refreshing
	keys, err := fetchJWKS(ctx, s.httpClient, jwksURL)
	if err != nil {
		return nil, err
	}

	entry = &jwksEntry{keys: keys, fetchedAt: time.Now()}
	s.jwksMu.Lock()
	s.jwksCache[jwksURL] = entry
	s.jwksMu.Unlock()

	for _, k := range entry.keys {
		if k.Kid == kid {
			return parseJWK(k)
		}
	}

	return nil, fmt.Errorf("cannot find signing key %q in JWKS", kid)
}

func fetchJWKS(ctx context.Context, httpClient *http.Client, jwksURL string) ([]jwk, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create jwks request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch jwks: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("cannot read jwks response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch jwks: unexpected status %d", resp.StatusCode)
	}

	var jwksResp jwksResponse
	if err := json.Unmarshal(body, &jwksResp); err != nil {
		return nil, fmt.Errorf("cannot parse jwks response: %w", err)
	}

	return jwksResp.Keys, nil
}

func parseJWK(k jwk) (crypto.PublicKey, error) {
	switch k.Kty {
	case "RSA":
		nBytes, err := base64.RawURLEncoding.DecodeString(k.N)
		if err != nil {
			return nil, fmt.Errorf("cannot decode RSA modulus: %w", err)
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(k.E)
		if err != nil {
			return nil, fmt.Errorf("cannot decode RSA exponent: %w", err)
		}

		n := new(big.Int).SetBytes(nBytes)
		e := 0
		for _, b := range eBytes {
			e = e<<8 + int(b)
		}

		return &rsa.PublicKey{N: n, E: e}, nil

	case "EC":
		var curve elliptic.Curve
		switch k.Crv {
		case "P-256":
			curve = elliptic.P256()
		case "P-384":
			curve = elliptic.P384()
		case "P-521":
			curve = elliptic.P521()
		default:
			return nil, fmt.Errorf("unsupported EC curve: %s", k.Crv)
		}

		xBytes, err := base64.RawURLEncoding.DecodeString(k.X)
		if err != nil {
			return nil, fmt.Errorf("cannot decode EC X: %w", err)
		}
		yBytes, err := base64.RawURLEncoding.DecodeString(k.Y)
		if err != nil {
			return nil, fmt.Errorf("cannot decode EC Y: %w", err)
		}

		return &ecdsa.PublicKey{
			Curve: curve,
			X:     new(big.Int).SetBytes(xBytes),
			Y:     new(big.Int).SetBytes(yBytes),
		}, nil

	default:
		return nil, fmt.Errorf("unsupported key type: %s", k.Kty)
	}
}

func verifySignature(alg string, key crypto.PublicKey, signedContent []byte, signature []byte) error {
	hash := sha256.Sum256(signedContent)

	switch alg {
	case "RS256":
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return fmt.Errorf("cannot verify RS256 signature: expected RSA public key")
		}
		return rsa.VerifyPKCS1v15(rsaKey, crypto.SHA256, hash[:], signature)

	case "ES256":
		ecKey, ok := key.(*ecdsa.PublicKey)
		if !ok {
			return fmt.Errorf("cannot verify ES256 signature: expected ECDSA public key")
		}

		// JWS (RFC 7515) encodes ECDSA signatures as raw R||S
		// concatenation (2x32 bytes for P-256), not ASN.1 DER.
		// Convert to ASN.1 for ecdsa.VerifyASN1.
		keySize := (ecKey.Curve.Params().BitSize + 7) / 8
		if len(signature) != 2*keySize {
			return fmt.Errorf("cannot verify ES256 signature: invalid length %d, expected %d", len(signature), 2*keySize)
		}

		r := new(big.Int).SetBytes(signature[:keySize])
		sigS := new(big.Int).SetBytes(signature[keySize:])

		derSig, err := asn1.Marshal(struct {
			R, S *big.Int
		}{r, sigS})
		if err != nil {
			return fmt.Errorf("cannot encode ECDSA signature to ASN.1: %w", err)
		}

		if !ecdsa.VerifyASN1(ecKey, hash[:], derSig) {
			return fmt.Errorf("cannot verify ECDSA signature")
		}
		return nil

	default:
		return fmt.Errorf("cannot verify signature: unsupported algorithm %s", alg)
	}
}

func computeCodeChallenge(codeVerifier string) string {
	hash := sha256.Sum256([]byte(codeVerifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := io.ReadFull(cryptoRandReader, b); err != nil {
		return "", fmt.Errorf("cannot generate random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

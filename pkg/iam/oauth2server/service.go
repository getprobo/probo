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
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"slices"
	"sync/atomic"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/gid"
)

type (
	Service struct {
		pg                   *pg.Client
		signingKeys          SigningKeys
		activeSigningIdx     []int
		rrCounter            atomic.Uint64
		baseURL              string
		logger               *log.Logger
		gc                   *GarbageCollector
		accessTokenDuration  time.Duration
		refreshTokenDuration time.Duration
	}

	Option func(*Service)

	// AuthorizeRequest contains the parsed parameters for an authorization request.
	AuthorizeRequest struct {
		IdentityID          gid.GID
		SessionID           gid.GID
		ResponseType        string
		ClientID            gid.GID
		RedirectURI         string
		Scopes              coredata.OAuth2Scopes
		CodeChallenge       string
		CodeChallengeMethod coredata.OAuth2CodeChallengeMethod
		Nonce               string
		State               string
		AuthTime            time.Time
	}

	// IntrospectionResult is returned by IntrospectToken on success.
	IntrospectionResult struct {
		Scopes     coredata.OAuth2Scopes
		ClientID   gid.GID
		IdentityID gid.GID
		ExpiresAt  time.Time
		CreatedAt  time.Time
	}

	// ConsentApprovalRequest contains the parameters for consent approval.
	ConsentApprovalRequest struct {
		ConsentID  gid.GID
		IdentityID gid.GID
		Approved   bool
		AuthTime   time.Time
	}

	// RegisterClientRequest contains the parameters for dynamic client registration.
	RegisterClientRequest struct {
		IdentityID              gid.GID  `json:"-"`
		OrganizationID          string   `json:"organization_id"`
		ClientName              string   `json:"client_name"`
		Visibility              string   `json:"visibility"`
		RedirectURIs            []string `json:"redirect_uris"`
		GrantTypes              []string `json:"grant_types"`
		ResponseTypes           []string `json:"response_types"`
		TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
		LogoURI                 *string  `json:"logo_uri"`
		ClientURI               *string  `json:"client_uri"`
		Contacts                []string `json:"contacts"`
		Scopes                  []string `json:"scopes"`
	}

	// TokenResponse is the JSON response returned by the token endpoint.
	TokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
		RefreshToken string `json:"refresh_token,omitempty"`
		IDToken      string `json:"id_token,omitempty"`
		Scope        string `json:"scope,omitempty"`
	}
)

func WithAccessTokenDuration(d time.Duration) Option {
	return func(s *Service) {
		s.accessTokenDuration = d
	}
}

func WithRefreshTokenDuration(d time.Duration) Option {
	return func(s *Service) {
		s.refreshTokenDuration = d
	}
}

func NewService(
	pgClient *pg.Client,
	signingKeys SigningKeys,
	baseURL string,
	logger *log.Logger,
	opts ...Option,
) *Service {
	var activeIdx []int
	for i, k := range signingKeys {
		if k.Active {
			activeIdx = append(activeIdx, i)
		}
	}

	s := &Service{
		pg:                   pgClient,
		signingKeys:          signingKeys,
		activeSigningIdx:     activeIdx,
		baseURL:              baseURL,
		logger:               logger,
		accessTokenDuration:  1 * time.Hour,
		refreshTokenDuration: 30 * 24 * time.Hour,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.gc = NewGarbageCollector(pgClient, logger)

	return s
}

// signingKey returns the next active signing key using round-robin.
func (s *Service) signingKey() *SigningKey {
	n := s.rrCounter.Add(1)
	idx := s.activeSigningIdx[n%uint64(len(s.activeSigningIdx))]
	return &s.signingKeys[idx]
}

func (s *Service) Run(ctx context.Context) error {
	return s.gc.Run(ctx)
}

// Metadata returns the OIDC discovery document.
func (s *Service) Metadata(endpoints Endpoints) *ServerMetadata {
	return NewMetadata(s.baseURL, endpoints)
}

// JWKS returns the public key set.
func (s *Service) JWKS() *jose.JWKS {
	jwks := &jose.JWKS{
		Keys: make([]jose.JWK, 0, len(s.signingKeys)),
	}

	for _, sk := range s.signingKeys {
		jwks.Keys = append(jwks.Keys, jose.RSAPublicKeyToJWK(&sk.PrivateKey.PublicKey, sk.KID))
	}

	return jwks
}

// hashToken computes SHA-256 hash of a token value.
func hashToken(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}

// hashTokenHex computes SHA-256 hash of a token value as hex string.
func hashTokenHex(token string) string {
	return hex.EncodeToString(hashToken(token))
}

// generateRandomToken generates a cryptographically random token of the given byte length,
// returned as a base64url-encoded string.
func generateRandomToken(byteLen int) (string, error) {
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("cannot generate random token: %w", err)
	}

	return hex.EncodeToString(b), nil
}

// verifyClientSecret verifies a plaintext secret against a stored hash
// using constant-time comparison.
func verifyClientSecret(storedHash []byte, secret string) bool {
	h := sha256.Sum256([]byte(secret))
	return subtle.ConstantTimeCompare(storedHash, h[:]) == 1
}

// validateScopes checks that all requested scopes are a subset of the allowed scopes.
func validateScopes(requested, allowed coredata.OAuth2Scopes) bool {
	for _, scope := range requested {
		if !slices.Contains(allowed, scope) {
			return false
		}
	}
	return true
}

// CreateAccessToken generates an opaque access token, stores it hashed,
// and returns the plaintext token value.
func (s *Service) CreateAccessToken(
	ctx context.Context,
	clientID gid.GID,
	identityID gid.GID,
	scopes coredata.OAuth2Scopes,
) (string, *coredata.OAuth2AccessToken, error) {
	tokenValue, err := generateRandomToken(32)
	if err != nil {
		return "", nil, err
	}

	now := time.Now()
	token := &coredata.OAuth2AccessToken{
		ID:          fmt.Sprintf("oat_%s", tokenValue[:16]),
		HashedValue: hashToken(tokenValue),
		ClientID:    clientID,
		IdentityID:  identityID,
		Scopes:      scopes,
		CreatedAt:   now,
		ExpiresAt:   now.Add(s.accessTokenDuration),
	}

	if err = s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			return token.Insert(ctx, conn)
		},
	); err != nil {
		return "", nil, fmt.Errorf("cannot create access token: %w", err)
	}

	return tokenValue, token, nil
}

// CreateRefreshToken generates an opaque refresh token, stores it hashed,
// and returns the plaintext token value.
func (s *Service) CreateRefreshToken(
	ctx context.Context,
	clientID gid.GID,
	identityID gid.GID,
	scopes coredata.OAuth2Scopes,
	accessTokenID string,
) (string, *coredata.OAuth2RefreshToken, error) {
	tokenValue, err := generateRandomToken(48)
	if err != nil {
		return "", nil, fmt.Errorf("cannot generate refresh token: %w", err)
	}

	now := time.Now()
	token := &coredata.OAuth2RefreshToken{
		ID:            fmt.Sprintf("ort_%s", tokenValue[:16]),
		HashedValue:   hashToken(tokenValue),
		ClientID:      clientID,
		IdentityID:    identityID,
		Scopes:        scopes,
		AccessTokenID: accessTokenID,
		CreatedAt:     now,
		ExpiresAt:     now.Add(s.refreshTokenDuration),
	}

	if err = s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := token.Insert(ctx, conn); err != nil {
				return fmt.Errorf("cannot create refresh token: %w", err)
			}
			return nil
		},
	); err != nil {
		return "", nil, err
	}

	return tokenValue, token, nil
}

// GetClientByID loads an OAuth2 client by ID without scope restriction.
func (s *Service) GetClientByID(ctx context.Context, clientID gid.GID) (*coredata.OAuth2Client, error) {
	var client coredata.OAuth2Client

	if err := s.pg.WithConn(
		ctx,
		func(conn pg.Conn) error {
			if err := client.LoadByID(ctx, conn, coredata.NewNoScope(), clientID); err != nil {
				return fmt.Errorf("cannot get oauth2 client: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &client, nil
}

// ExchangeAuthorizationCode validates and exchanges an authorization code for tokens.
func (s *Service) ExchangeAuthorizationCode(
	ctx context.Context,
	client *coredata.OAuth2Client,
	codeValue string,
	redirectURI string,
	codeVerifier string,
) (*TokenResponse, error) {
	var (
		codeHash          = hashTokenHex(codeValue)
		codeScopes        coredata.OAuth2Scopes
		identityID        gid.GID
		authTime          time.Time
		nonce             string
		accessToken       *coredata.OAuth2AccessToken
		accessTokenValue  string
		refreshTokenValue string
	)

	if err := s.pg.WithTx(ctx, func(tx pg.Conn) error {
		var code coredata.OAuth2AuthorizationCode
		if err := code.LoadByIDForUpdate(ctx, tx, codeHash); err != nil {
			return fmt.Errorf("cannot load authorization code: %w", err)
		}

		if time.Now().After(code.ExpiresAt) {
			return fmt.Errorf("authorization code expired")
		}

		if code.ClientID != client.ID {
			return fmt.Errorf("client_id mismatch")
		}

		if code.RedirectURI != redirectURI {
			return fmt.Errorf("redirect_uri mismatch")
		}

		// Single-use: delete after validation.
		if err := code.Delete(ctx, tx); err != nil {
			return fmt.Errorf("cannot delete authorization code: %w", err)
		}

		// PKCE validation.
		if code.CodeChallenge != nil {
			if codeVerifier == "" {
				return fmt.Errorf("code_verifier required")
			}
			method := coredata.OAuth2CodeChallengeMethodS256
			if code.CodeChallengeMethod != nil {
				method = *code.CodeChallengeMethod
			}
			if !ValidateCodeChallenge(codeVerifier, *code.CodeChallenge, method) {
				return fmt.Errorf("invalid code_verifier")
			}
		}

		codeScopes = code.Scopes
		identityID = code.IdentityID
		authTime = code.AuthTime
		if code.Nonce != nil {
			nonce = *code.Nonce
		}

		// Issue access token.
		tokenValue, err := generateRandomToken(32)
		if err != nil {
			return fmt.Errorf("cannot generate access token: %w", err)
		}

		now := time.Now()
		accessToken = &coredata.OAuth2AccessToken{
			ID:          fmt.Sprintf("oat_%s", tokenValue[:16]),
			HashedValue: hashToken(tokenValue),
			ClientID:    client.ID,
			IdentityID:  identityID,
			Scopes:      codeScopes,
			CreatedAt:   now,
			ExpiresAt:   now.Add(s.accessTokenDuration),
		}

		if err := accessToken.Insert(ctx, tx); err != nil {
			return fmt.Errorf("cannot create access token: %w", err)
		}

		accessTokenValue = tokenValue

		// Issue refresh token if granted.
		if slices.Contains(client.GrantTypes, coredata.OAuth2GrantTypeRefreshToken) {
			rtValue, err := generateRandomToken(48)
			if err != nil {
				return fmt.Errorf("cannot generate refresh token: %w", err)
			}

			rt := &coredata.OAuth2RefreshToken{
				ID:            fmt.Sprintf("ort_%s", rtValue[:16]),
				HashedValue:   hashToken(rtValue),
				ClientID:      client.ID,
				IdentityID:    identityID,
				Scopes:        codeScopes,
				AccessTokenID: accessToken.ID,
				CreatedAt:     now,
				ExpiresAt:     now.Add(s.refreshTokenDuration),
			}

			if err := rt.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot create refresh token: %w", err)
			}

			refreshTokenValue = rtValue
		}

		return nil
	}); err != nil {
		return nil, err
	}

	resp := &TokenResponse{
		AccessToken:  accessTokenValue,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(accessToken.ExpiresAt).Seconds()),
		RefreshToken: refreshTokenValue,
		Scope:        codeScopes.String(),
	}

	if slices.Contains(codeScopes, coredata.OAuth2ScopeOpenID) {
		idTokenClaims := NewIDTokenClaims(
			s.baseURL,
			identityID,
			client.ID,
			authTime,
			codeScopes,
			nonce,
			accessTokenValue,
			"", false, "",
			s.accessTokenDuration,
		)

		sk := s.signingKey()
		idToken, err := jose.SignJWT(sk.PrivateKey, sk.KID, idTokenClaims)
		if err != nil {
			return nil, fmt.Errorf("cannot sign id token: %w", err)
		}

		resp.IDToken = idToken
	}

	return resp, nil
}

// RefreshToken exchanges a refresh token for new tokens with rotation.
func (s *Service) RefreshToken(
	ctx context.Context,
	client *coredata.OAuth2Client,
	refreshTokenValue string,
) (*TokenResponse, error) {
	// Generate new token values before the transaction so that token
	// creation and old-token revocation happen atomically.
	accessTokenValue, err := generateRandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("cannot generate access token: %w", err)
	}

	refreshTokenValueNew, err := generateRandomToken(48)
	if err != nil {
		return nil, fmt.Errorf("cannot generate refresh token: %w", err)
	}

	var (
		hashedValue = hashToken(refreshTokenValue)
		accessToken *coredata.OAuth2AccessToken
		identityID  gid.GID
		scopes      coredata.OAuth2Scopes
	)

	if err := s.pg.WithTx(ctx, func(tx pg.Conn) error {
		var oldToken coredata.OAuth2RefreshToken
		if err := oldToken.LoadByHashedValueForUpdate(ctx, tx, hashedValue); err != nil {
			return fmt.Errorf("cannot load refresh token: %w", err)
		}

		if oldToken.ClientID != client.ID {
			return fmt.Errorf("client_id mismatch")
		}

		// Replay detection: if already revoked, revoke ALL tokens for this client+identity.
		if oldToken.RevokedAt != nil {
			s.logger.WarnCtx(ctx, "refresh token replay detected, revoking all tokens",
				log.String("client_id", client.ID.String()),
				log.String("identity_id", oldToken.IdentityID.String()),
			)

			now := time.Now()
			var at coredata.OAuth2AccessToken
			if _, err := at.DeleteByClientAndIdentity(ctx, tx, client.ID, oldToken.IdentityID); err != nil {
				return fmt.Errorf("cannot delete access tokens: %w", err)
			}
			var rt coredata.OAuth2RefreshToken
			if _, err := rt.RevokeByClientAndIdentity(ctx, tx, client.ID, oldToken.IdentityID, now); err != nil {
				return fmt.Errorf("cannot revoke refresh tokens: %w", err)
			}
			return fmt.Errorf("refresh token replay detected")
		}

		if time.Now().After(oldToken.ExpiresAt) {
			return fmt.Errorf("refresh token expired")
		}

		identityID = oldToken.IdentityID
		scopes = oldToken.Scopes

		// Revoke old refresh token (rotation).
		if err := oldToken.Revoke(ctx, tx, time.Now()); err != nil {
			return fmt.Errorf("cannot revoke old refresh token: %w", err)
		}

		// Delete old access token (best effort).
		oldAT := &coredata.OAuth2AccessToken{ID: oldToken.AccessTokenID}
		_ = oldAT.Delete(ctx, tx)

		// Create new access token in the same transaction.
		now := time.Now()
		accessToken = &coredata.OAuth2AccessToken{
			ID:          fmt.Sprintf("oat_%s", accessTokenValue[:16]),
			HashedValue: hashToken(accessTokenValue),
			ClientID:    client.ID,
			IdentityID:  identityID,
			Scopes:      scopes,
			CreatedAt:   now,
			ExpiresAt:   now.Add(s.accessTokenDuration),
		}
		if err := accessToken.Insert(ctx, tx); err != nil {
			return fmt.Errorf("cannot create access token: %w", err)
		}

		// Create new refresh token in the same transaction.
		newRT := &coredata.OAuth2RefreshToken{
			ID:            fmt.Sprintf("ort_%s", refreshTokenValueNew[:16]),
			HashedValue:   hashToken(refreshTokenValueNew),
			ClientID:      client.ID,
			IdentityID:    identityID,
			Scopes:        scopes,
			AccessTokenID: accessToken.ID,
			CreatedAt:     now,
			ExpiresAt:     now.Add(s.refreshTokenDuration),
		}
		if err := newRT.Insert(ctx, tx); err != nil {
			return fmt.Errorf("cannot create refresh token: %w", err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	resp := &TokenResponse{
		AccessToken:  accessTokenValue,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(accessToken.ExpiresAt).Seconds()),
		RefreshToken: refreshTokenValueNew,
		Scope:        scopes.String(),
	}

	if slices.Contains(scopes, coredata.OAuth2ScopeOpenID) {
		claims := NewIDTokenClaims(
			s.baseURL,
			identityID,
			client.ID,
			time.Now(),
			scopes,
			"",
			accessTokenValue,
			"", false, "",
			s.accessTokenDuration,
		)
		sk := s.signingKey()
		idToken, err := jose.SignJWT(sk.PrivateKey, sk.KID, claims)
		if err != nil {
			return nil, fmt.Errorf("cannot sign id token: %w", err)
		}
		resp.IDToken = idToken
	}

	return resp, nil
}

// DeviceCodeResult contains the result of a device authorization request.
type DeviceCodeResult struct {
	DeviceCode string
	UserCode   coredata.OAuth2UserCode
	ExpiresIn  int
	Interval   int
}

// CreateDeviceCode validates the client, requested scopes, and creates a new
// device authorization request.
func (s *Service) CreateDeviceCode(
	ctx context.Context,
	clientID gid.GID,
	scope string,
) (*DeviceCodeResult, error) {
	client, err := s.GetClientByID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("%w: unknown client_id", ErrInvalidRequest)
	}

	if !slices.Contains(client.GrantTypes, coredata.OAuth2GrantTypeDeviceCode) {
		return nil, fmt.Errorf("%w: client not authorized for device flow", ErrUnauthorizedClient)
	}

	var parsedScopes coredata.OAuth2Scopes
	_ = parsedScopes.UnmarshalText([]byte(scope))

	requestedScopes := parsedScopes.OrDefault(client.Scopes)
	if !validateScopes(requestedScopes, client.Scopes) {
		return nil, fmt.Errorf("%w: requested scope exceeds client registration", ErrInvalidScope)
	}

	deviceCodeValue, err := generateRandomToken(32)
	if err != nil {
		return nil, fmt.Errorf("cannot generate device code: %w", err)
	}

	var userCode coredata.OAuth2UserCode
	now := time.Now()

	err = s.pg.WithConn(ctx, func(conn pg.Conn) error {
		// Retry up to 3 times on user code collision.
		for range 3 {
			uc, err := GenerateUserCode()
			if err != nil {
				return fmt.Errorf("cannot generate user code: %w", err)
			}

			dc := &coredata.OAuth2DeviceCode{
				ID:             fmt.Sprintf("odc_%s", deviceCodeValue[:16]),
				DeviceCodeHash: hashToken(deviceCodeValue),
				UserCode:       uc,
				ClientID:       client.ID,
				Scopes:         requestedScopes,
				Status:         coredata.OAuth2DeviceCodeStatusPending,
				PollInterval:   5,
				CreatedAt:      now,
				ExpiresAt:      now.Add(10 * time.Minute),
			}

			if err := dc.Insert(ctx, conn); err != nil {
				continue // retry on collision
			}

			userCode = uc
			return nil
		}

		return fmt.Errorf("cannot generate unique user code after 3 attempts")
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrServerError, err)
	}

	return &DeviceCodeResult{
		DeviceCode: deviceCodeValue,
		UserCode:   userCode,
		ExpiresIn:  600,
		Interval:   5,
	}, nil
}

// PollDeviceCode checks the status of a device code and returns tokens if authorized.
// Returns sentinel errors (ErrSlowDown, ErrAuthorizationPending, ErrExpiredToken,
// ErrAccessDenied, ErrInvalidGrant) for the handler to map to OAuth2 responses.
func (s *Service) PollDeviceCode(
	ctx context.Context,
	clientID gid.GID,
	deviceCodeValue string,
) (*TokenResponse, error) {
	hashedValue := hashToken(deviceCodeValue)

	var dc coredata.OAuth2DeviceCode

	err := s.pg.WithTx(ctx, func(tx pg.Conn) error {
		if err := dc.LoadByDeviceCodeHashForUpdate(ctx, tx, hashedValue); err != nil {
			return err
		}

		if dc.ClientID != clientID {
			return fmt.Errorf("client_id mismatch")
		}

		// Rate limiting.
		now := time.Now()
		if dc.LastPolledAt != nil {
			elapsed := now.Sub(*dc.LastPolledAt)
			if elapsed < time.Duration(dc.PollInterval)*time.Second {
				newInterval := dc.PollInterval + 5
				if err := dc.UpdateLastPolledAt(ctx, tx, now, newInterval); err != nil {
					return fmt.Errorf("cannot update poll interval: %w", err)
				}
				return ErrSlowDown
			}
		}
		if err := dc.UpdateLastPolledAt(ctx, tx, now, dc.PollInterval); err != nil {
			return fmt.Errorf("cannot update last polled at: %w", err)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, ErrSlowDown) {
			return nil, ErrSlowDown
		}
		return nil, fmt.Errorf("%w: invalid device code", ErrInvalidGrant)
	}

	if time.Now().After(dc.ExpiresAt) {
		return nil, ErrExpiredToken
	}

	switch dc.Status {
	case coredata.OAuth2DeviceCodeStatusPending:
		return nil, ErrAuthorizationPending
	case coredata.OAuth2DeviceCodeStatusDenied:
		return nil, ErrAccessDenied
	case coredata.OAuth2DeviceCodeStatusAuthorized:
		// Continue to issue tokens.
	default:
		return nil, fmt.Errorf("%w: invalid device code status", ErrInvalidGrant)
	}

	if dc.IdentityID == nil {
		return nil, fmt.Errorf("%w: device code authorized but no identity", ErrServerError)
	}

	identityID := *dc.IdentityID

	accessTokenValue, accessToken, err := s.CreateAccessToken(ctx, clientID, identityID, dc.Scopes)
	if err != nil {
		return nil, fmt.Errorf("%w: cannot create access token", ErrServerError)
	}

	resp := &TokenResponse{
		AccessToken: accessTokenValue,
		TokenType:   "Bearer",
		ExpiresIn:   int64(time.Until(accessToken.ExpiresAt).Seconds()),
		Scope:       dc.Scopes.String(),
	}

	if slices.Contains(dc.Scopes, coredata.OAuth2ScopeOpenID) {
		claims := NewIDTokenClaims(
			s.baseURL,
			identityID,
			clientID,
			time.Now(),
			dc.Scopes,
			"",
			accessTokenValue,
			"", false, "",
			s.accessTokenDuration,
		)
		sk := s.signingKey()
		idToken, signErr := jose.SignJWT(sk.PrivateKey, sk.KID, claims)
		if signErr != nil {
			return nil, fmt.Errorf("%w: cannot sign id token", ErrServerError)
		}
		resp.IDToken = idToken
	}

	// Load client to check if refresh tokens are allowed.
	client, loadErr := s.GetClientByID(ctx, clientID)
	if loadErr == nil {
		if slices.Contains(client.GrantTypes, coredata.OAuth2GrantTypeRefreshToken) {
			refreshTokenValue, _, err := s.CreateRefreshToken(
				ctx,
				clientID,
				identityID,
				dc.Scopes,
				accessToken.ID,
			)
			if err == nil {
				resp.RefreshToken = refreshTokenValue
			}
		}
	}

	// Clean up device code after successful token issuance.
	_ = s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return dc.Delete(ctx, conn)
	})

	return resp, nil
}

// AuthorizeDevice authorizes a device code by user code.
func (s *Service) AuthorizeDevice(
	ctx context.Context,
	identityID gid.GID,
	userCode string,
) error {
	return s.pg.WithTx(ctx, func(tx pg.Conn) error {
		var dc coredata.OAuth2DeviceCode
		if err := dc.LoadByUserCode(ctx, tx, userCode); err != nil {
			return fmt.Errorf("cannot find device code: %w", err)
		}

		if time.Now().After(dc.ExpiresAt) {
			return fmt.Errorf("device code expired")
		}

		if dc.Status != coredata.OAuth2DeviceCodeStatusPending {
			return fmt.Errorf("device code already %s", dc.Status)
		}

		return dc.UpdateStatus(ctx, tx, coredata.OAuth2DeviceCodeStatusAuthorized, &identityID)
	})
}

// RegisterClient creates a new OAuth2 client for the given organization.
// Returns (clientID, plaintext_secret, error). Secret is empty for public clients.
func (s *Service) RegisterClient(
	ctx context.Context,
	req *RegisterClientRequest,
) (gid.GID, string, error) {
	orgID, err := gid.ParseGID(req.OrganizationID)
	if err != nil {
		return gid.GID{}, "", fmt.Errorf("%w: invalid organization_id", ErrInvalidRequest)
	}

	// Check org membership.
	var membership coredata.Membership
	err = s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return membership.LoadActiveByIdentityIDAndOrganizationID(
			ctx,
			conn,
			req.IdentityID,
			orgID,
		)
	})
	if err != nil {
		return gid.GID{}, "", fmt.Errorf("%w: not a member of the organization", ErrAccessDenied)
	}

	// Apply defaults.
	grantTypes := req.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []string{coredata.OAuth2GrantTypeAuthorizationCode.String()}
	}

	responseTypes := req.ResponseTypes
	if len(responseTypes) == 0 {
		responseTypes = []string{coredata.OAuth2ResponseTypeCode.String()}
	}

	authMethod := req.TokenEndpointAuthMethod
	if authMethod == "" {
		authMethod = coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretBasic.String()
	}

	visibility := req.Visibility
	if visibility == "" {
		visibility = coredata.OAuth2ClientVisibilityPrivate.String()
	}

	scopeStrs := req.Scopes
	if len(scopeStrs) == 0 {
		scopeStrs = []string{
			coredata.OAuth2ScopeOpenID.String(),
			coredata.OAuth2ScopeProfile.String(),
			coredata.OAuth2ScopeEmail.String(),
		}
	}

	scopes := make(coredata.OAuth2Scopes, len(scopeStrs))
	for i, s := range scopeStrs {
		scopes[i] = coredata.OAuth2Scope(s)
	}

	parsedGrantTypes := make([]coredata.OAuth2GrantType, len(grantTypes))
	for i, s := range grantTypes {
		parsedGrantTypes[i] = coredata.OAuth2GrantType(s)
	}

	parsedResponseTypes := make([]coredata.OAuth2ResponseType, len(responseTypes))
	for i, s := range responseTypes {
		parsedResponseTypes[i] = coredata.OAuth2ResponseType(s)
	}

	tenantID := orgID.TenantID()
	clientID := gid.New(tenantID, coredata.OAuth2ClientEntityType)

	var (
		secretHash      []byte
		plaintextSecret string
	)

	parsedAuthMethod := coredata.OAuth2ClientTokenEndpointAuthMethod(authMethod)
	if parsedAuthMethod != coredata.OAuth2ClientTokenEndpointAuthMethodNone {
		secret, err := generateRandomToken(32)
		if err != nil {
			return gid.GID{}, "", fmt.Errorf("cannot generate client secret: %w", err)
		}
		plaintextSecret = secret
		secretHash = hashToken(secret)
	}

	now := time.Now()
	client := &coredata.OAuth2Client{
		ID:                      clientID,
		OrganizationID:          orgID,
		ClientSecretHash:        secretHash,
		ClientName:              req.ClientName,
		Visibility:              coredata.OAuth2ClientVisibility(visibility),
		RedirectURIs:            req.RedirectURIs,
		Scopes:                  scopes,
		GrantTypes:              parsedGrantTypes,
		ResponseTypes:           parsedResponseTypes,
		TokenEndpointAuthMethod: parsedAuthMethod,
		LogoURI:                 req.LogoURI,
		ClientURI:               req.ClientURI,
		Contacts:                req.Contacts,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	scope := coredata.NewScopeFromObjectID(orgID)

	err = s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return client.Insert(ctx, conn, scope)
	})
	if err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot insert oauth2 client: %w", err)
	}

	return clientID, plaintextSecret, nil
}

// LoadAccessToken loads an access token by its plaintext value and checks
// that it has not expired. Returns the token or an error if the token is
// unknown or expired.
func (s *Service) LoadAccessToken(ctx context.Context, tokenValue string) (*coredata.OAuth2AccessToken, error) {
	hashedValue := hashToken(tokenValue)
	var token coredata.OAuth2AccessToken

	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return token.LoadByHashedValue(ctx, conn, hashedValue)
	})
	if err != nil {
		return nil, err
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("access token expired")
	}

	return &token, nil
}

// IntrospectToken loads an access token by its plaintext value and validates
// that it belongs to the given client and has not expired.
func (s *Service) IntrospectToken(ctx context.Context, clientID gid.GID, tokenValue string) (*IntrospectionResult, error) {
	hashedValue := hashToken(tokenValue)
	var token coredata.OAuth2AccessToken

	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return token.LoadByHashedValue(ctx, conn, hashedValue)
	})
	if err != nil {
		return nil, fmt.Errorf("cannot load access token: %w", err)
	}

	if token.ClientID != clientID {
		return nil, fmt.Errorf("cannot introspect token: client mismatch")
	}

	if time.Now().After(token.ExpiresAt) {
		return nil, fmt.Errorf("cannot introspect token: expired")
	}

	return &IntrospectionResult{
		Scopes:     token.Scopes,
		ClientID:   token.ClientID,
		IdentityID: token.IdentityID,
		ExpiresAt:  token.ExpiresAt,
		CreatedAt:  token.CreatedAt,
	}, nil
}

// RevokeToken revokes a token (access or refresh) that belongs to the given
// client. Per RFC 7009, this always succeeds even if the token is invalid or
// does not belong to the client.
func (s *Service) RevokeToken(ctx context.Context, clientID gid.GID, tokenValue string) {
	hashedValue := hashToken(tokenValue)

	// Try access token first.
	var accessToken coredata.OAuth2AccessToken
	err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return accessToken.LoadByHashedValue(ctx, conn, hashedValue)
	})
	if err == nil {
		if accessToken.ClientID == clientID {
			_ = s.pg.WithConn(ctx, func(conn pg.Conn) error {
				return accessToken.Delete(ctx, conn)
			})
		}
		return
	}

	// Try refresh token.
	var refreshToken coredata.OAuth2RefreshToken
	err = s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return refreshToken.LoadByHashedValue(ctx, conn, hashedValue)
	})
	if err == nil {
		if refreshToken.ClientID == clientID {
			_ = s.pg.WithConn(ctx, func(conn pg.Conn) error {
				return refreshToken.Revoke(ctx, conn, time.Now())
			})
		}
	}
}

// Authorize validates an authorization request, checks existing consent,
// and either returns an authorization code or signals that consent is needed.
//
// Authorize validates the client, redirect URI, and membership, then either
// issues an authorization code or returns a ConsentRequiredError.
func (s *Service) Authorize(
	ctx context.Context,
	req *AuthorizeRequest,
) (string, error) {
	client, err := s.GetClientByID(ctx, req.ClientID)
	if err != nil {
		return "", fmt.Errorf("cannot load client: %w", ErrInvalidClient)
	}

	if !client.IsRedirectURIValid(req.RedirectURI) {
		return "", fmt.Errorf("%w", ErrInvalidRedirectURI)
	}

	if client.Visibility == coredata.OAuth2ClientVisibilityPrivate {
		var membership coredata.Membership
		err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
			return membership.LoadActiveByIdentityIDAndOrganizationID(
				ctx,
				conn,
				req.IdentityID,
				client.OrganizationID,
			)
		})
		if err != nil {
			return "", fmt.Errorf(
				"%w: client is private and user is not a member of the organization",
				ErrUnauthorizedClient,
			)
		}
	}

	if req.ResponseType != string(coredata.OAuth2ResponseTypeCode) {
		return "", fmt.Errorf(
			"%w: unsupported response_type",
			ErrInvalidRequest,
		)
	}

	requestedScopes := req.Scopes.OrDefault(client.Scopes)
	if !validateScopes(requestedScopes, client.Scopes) {
		return "", fmt.Errorf(
			"%w: requested scope exceeds client registration",
			ErrInvalidScope,
		)
	}

	codeChallengeMethod := req.CodeChallengeMethod
	if client.TokenEndpointAuthMethod == coredata.OAuth2ClientTokenEndpointAuthMethodNone && req.CodeChallenge == "" {
		return "", fmt.Errorf(
			"%w: code_challenge required for public clients",
			ErrInvalidRequest,
		)
	}

	if codeChallengeMethod != "" && codeChallengeMethod != coredata.OAuth2CodeChallengeMethodS256 {
		return "", fmt.Errorf(
			"%w: only S256 code_challenge_method is supported",
			ErrInvalidRequest,
		)
	}

	if req.CodeChallenge != "" && codeChallengeMethod == "" {
		codeChallengeMethod = coredata.OAuth2CodeChallengeMethodS256
	}

	// Check existing approved consent with matching scopes.
	var existingConsent coredata.OAuth2Consent
	consentErr := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return existingConsent.LoadMatchingConsent(
			ctx,
			conn,
			req.IdentityID,
			client.ID,
			requestedScopes,
		)
	})
	if consentErr == nil {
		code, err := s.issueAuthorizationCode(
			ctx,
			client,
			req.IdentityID,
			req.RedirectURI,
			requestedScopes,
			req.CodeChallenge,
			codeChallengeMethod,
			req.Nonce,
			req.AuthTime,
		)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrServerError, err)
		}

		return code, nil
	}

	// Create a pending consent record storing the full authorization request.
	now := time.Now()
	pendingConsent := &coredata.OAuth2Consent{
		ID:                  gid.New(client.ID.TenantID(), coredata.OAuth2ConsentEntityType),
		IdentityID:          req.IdentityID,
		SessionID:           req.SessionID,
		ClientID:            client.ID,
		Scopes:              requestedScopes,
		RedirectURI:         req.RedirectURI,
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Nonce:               req.Nonce,
		State:               req.State,
		Approved:            false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return pendingConsent.Insert(ctx, conn)
	}); err != nil {
		return "", fmt.Errorf("%w: cannot create pending consent", ErrServerError)
	}

	return "", &ConsentRequiredError{
		ConsentID: pendingConsent.ID,
		Client:    client,
		Scopes:    requestedScopes,
	}
}

// ApproveConsent loads the pending consent by ID, validates ownership, and
// either issues an authorization code (when approved) or returns ErrAccessDenied.
// Returns (code, redirect_uri, state, error).
func (s *Service) ApproveConsent(
	ctx context.Context,
	req *ConsentApprovalRequest,
) (string, string, string, error) {
	var consent coredata.OAuth2Consent
	if err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return consent.LoadByID(ctx, conn, req.ConsentID)
	}); err != nil {
		return "", "", "", fmt.Errorf("%w: consent not found", ErrInvalidRequest)
	}

	if consent.IdentityID != req.IdentityID {
		return "", "", "", fmt.Errorf("%w: consent does not belong to this identity", ErrAccessDenied)
	}

	if consent.Approved {
		return "", "", "", fmt.Errorf("%w: consent already processed", ErrInvalidRequest)
	}

	client, err := s.GetClientByID(ctx, consent.ClientID)
	if err != nil {
		return "", "", "", fmt.Errorf("cannot load client: %w", ErrInvalidClient)
	}

	if !client.IsRedirectURIValid(consent.RedirectURI) {
		return "", "", "", fmt.Errorf("%w", ErrInvalidRedirectURI)
	}

	if !req.Approved {
		// Clean up the pending consent.
		_ = s.pg.WithConn(ctx, func(conn pg.Conn) error {
			return consent.Delete(ctx, conn)
		})
		return "", consent.RedirectURI, consent.State, fmt.Errorf("%w: user denied the request", ErrAccessDenied)
	}

	// Mark consent as approved.
	consent.Approved = true
	consent.UpdatedAt = time.Now()
	if err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return consent.Update(ctx, conn)
	}); err != nil {
		return "", consent.RedirectURI, consent.State, fmt.Errorf("%w: cannot approve consent", ErrServerError)
	}

	code, err := s.issueAuthorizationCode(
		ctx,
		client,
		consent.IdentityID,
		consent.RedirectURI,
		consent.Scopes,
		consent.CodeChallenge,
		consent.CodeChallengeMethod,
		consent.Nonce,
		req.AuthTime,
	)
	if err != nil {
		return "", consent.RedirectURI, consent.State, err
	}

	return code, consent.RedirectURI, consent.State, nil
}

// AuthenticateClient verifies client credentials and returns the client.
// Public clients (token_endpoint_auth_method=none) do not require a secret.
func (s *Service) AuthenticateClient(
	ctx context.Context,
	clientID gid.GID,
	clientSecret string,
) (*coredata.OAuth2Client, error) {
	client, err := s.GetClientByID(ctx, clientID)
	if err != nil {
		return nil, fmt.Errorf("cannot load client: %w", ErrInvalidClient)
	}

	if client.TokenEndpointAuthMethod == coredata.OAuth2ClientTokenEndpointAuthMethodNone {
		return client, nil
	}

	if clientSecret == "" {
		return nil, fmt.Errorf("missing client_secret: %w", ErrInvalidClient)
	}

	if !verifyClientSecret(client.ClientSecretHash, clientSecret) {
		return nil, fmt.Errorf("invalid client_secret: %w", ErrInvalidClient)
	}

	return client, nil
}

// issueAuthorizationCode generates and persists an authorization code,
// returning the plaintext code value.
func (s *Service) issueAuthorizationCode(
	ctx context.Context,
	client *coredata.OAuth2Client,
	identityID gid.GID,
	redirectURI string,
	scopes coredata.OAuth2Scopes,
	codeChallenge string,
	codeChallengeMethod coredata.OAuth2CodeChallengeMethod,
	nonce string,
	authTime time.Time,
) (string, error) {
	codeValue, err := generateRandomToken(32)
	if err != nil {
		return "", fmt.Errorf("cannot generate authorization code: %w", err)
	}

	codeHash := hashTokenHex(codeValue)
	now := time.Now()

	code := &coredata.OAuth2AuthorizationCode{
		ID:          codeHash,
		ClientID:    client.ID,
		IdentityID:  identityID,
		RedirectURI: redirectURI,
		Scopes:      scopes,
		AuthTime:    authTime,
		CreatedAt:   now,
		ExpiresAt:   now.Add(10 * time.Minute),
	}

	if codeChallenge != "" {
		code.CodeChallenge = &codeChallenge
		code.CodeChallengeMethod = &codeChallengeMethod
	}

	if nonce != "" {
		code.Nonce = &nonce
	}

	if err := s.pg.WithConn(ctx, func(conn pg.Conn) error {
		return code.Insert(ctx, conn)
	}); err != nil {
		return "", fmt.Errorf("cannot save authorization code: %w", err)
	}

	return codeValue, nil
}

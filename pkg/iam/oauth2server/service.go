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
	"crypto/subtle"
	"errors"
	"fmt"
	"net/url"
	"sync/atomic"
	"time"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.gearno.de/x/ref"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/hash"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/crypto/rand"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/net"
	"go.probo.inc/probo/pkg/uri"
)

const (
	tokenByteLength        = 32
	refreshTokenByteLength = 48
	tokenTypeBearer        = "Bearer"

	// userCodeAlphabet excludes ambiguous characters: 0/O, 1/I/L.
	userCodeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
)

type (
	Service struct {
		pg                        *pg.Client
		signingKeys               SigningKeys
		activeSigningIdx          []int
		rrCounter                 atomic.Uint64
		baseURL                   uri.URI
		logger                    *log.Logger
		gc                        *GarbageCollector
		accessTokenDuration       time.Duration
		refreshTokenDuration      time.Duration
		authorizationCodeDuration time.Duration
		deviceCodeDuration        time.Duration
	}

	Option func(*Service)

	AuthorizeRequest struct {
		IdentityID          gid.GID
		SessionID           gid.GID
		ResponseType        coredata.OAuth2ResponseType
		ClientID            gid.GID
		RedirectURI         string
		Scopes              coredata.OAuth2Scopes
		CodeChallenge       string
		CodeChallengeMethod coredata.OAuth2CodeChallengeMethod
		Nonce               string
		State               string
		AuthTime            time.Time
	}

	ConsentApprovalRequest struct {
		ConsentID  gid.GID
		IdentityID gid.GID
		Approved   bool
		AuthTime   time.Time
	}

	RegisterClientRequest struct {
		IdentityID              gid.GID
		OrganizationID          gid.GID
		ClientName              string
		Visibility              coredata.OAuth2ClientVisibility
		RedirectURIs            []uri.URI
		GrantTypes              []coredata.OAuth2GrantType
		ResponseTypes           []coredata.OAuth2ResponseType
		TokenEndpointAuthMethod coredata.OAuth2ClientTokenEndpointAuthMethod
		LogoURI                 *uri.URI
		ClientURI               *uri.URI
		Contacts                []string
		Scopes                  coredata.OAuth2Scopes
	}

	TokenResult struct {
		AccessToken  string
		TokenType    string
		ExpiresIn    int64
		RefreshToken string
		IDToken      string
		Scope        string
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

func WithAuthorizationCodeDuration(d time.Duration) Option {
	return func(s *Service) {
		s.authorizationCodeDuration = d
	}
}

func WithDeviceCodeDuration(d time.Duration) Option {
	return func(s *Service) {
		s.deviceCodeDuration = d
	}
}

func NewService(
	pgClient *pg.Client,
	signingKeys SigningKeys,
	baseURL uri.URI,
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
		pg:                        pgClient,
		signingKeys:               signingKeys,
		activeSigningIdx:          activeIdx,
		baseURL:                   baseURL,
		logger:                    logger,
		accessTokenDuration:       1 * time.Hour,
		refreshTokenDuration:      30 * 24 * time.Hour,
		authorizationCodeDuration: 10 * time.Minute,
		deviceCodeDuration:        10 * time.Minute,
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
		jwks.Keys = append(
			jwks.Keys,
			jose.RSAPublicKeyToJWK(&sk.PrivateKey.PublicKey, sk.KID),
		)
	}

	return jwks
}

func (s *Service) CreateAccessToken(
	ctx context.Context,
	clientID gid.GID,
	identityID gid.GID,
	scopes coredata.OAuth2Scopes,
) (string, *coredata.OAuth2AccessToken, error) {
	tokenValue := rand.MustHexString(tokenByteLength)

	now := time.Now()
	token := &coredata.OAuth2AccessToken{
		ID:          fmt.Sprintf("oat_%s", tokenValue[:16]),
		HashedValue: hash.SHA256String(tokenValue),
		ClientID:    clientID,
		IdentityID:  identityID,
		Scopes:      scopes,
		CreatedAt:   now,
		ExpiresAt:   now.Add(s.accessTokenDuration),
	}

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := token.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot create access token: %w", err)
			}

			return nil
		},
	); err != nil {
		return "", nil, err
	}

	return tokenValue, token, nil
}

func (s *Service) GetClientByID(ctx context.Context, clientID gid.GID) (*coredata.OAuth2Client, error) {
	client := coredata.OAuth2Client{}

	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := client.LoadByID(ctx, conn, coredata.NewNoScope(), clientID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewError(ErrInvalidClient, WithDescription("client not found"))
				}

				return fmt.Errorf("cannot load oauth2 client: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &client, nil
}

func (s *Service) ExchangeAuthorizationCode(
	ctx context.Context,
	client *coredata.OAuth2Client,
	codeValue, redirectURI, codeVerifier string,
) (*TokenResult, error) {
	var (
		code                 = coredata.OAuth2AuthorizationCode{}
		now                  = time.Now()
		accessTokenExpiresAt = now.Add(s.accessTokenDuration)
		accessTokenValue     = rand.MustHexString(tokenByteLength)
		refreshTokenValue    = rand.MustHexString(refreshTokenByteLength)
		idToken              string
		codeHash             = hash.SHA256HexString(codeValue)
	)

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := code.LoadByIDForUpdate(ctx, tx, codeHash, client.ID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewError(ErrInvalidGrant, WithDescription("authorization code not found"))
				}

				return fmt.Errorf("cannot load authorization code: %w", err)
			}

			if err := code.Delete(ctx, tx); err != nil {
				return fmt.Errorf("cannot delete authorization code: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	if now.After(code.ExpiresAt) {
		return nil, NewError(
			ErrInvalidGrant,
			WithDescription("authorization code expired"),
		)
	}

	if code.RedirectURI.String() != redirectURI {
		return nil, NewError(
			ErrInvalidRedirectURI,
			WithDescription("redirect_uri mismatch"),
		)
	}

	if code.CodeChallenge != nil {
		if codeVerifier == "" {
			return nil, NewError(
				ErrInvalidRequest,
				WithDescription("code_verifier required"),
			)
		}

		if !ValidateCodeChallenge(codeVerifier, *code.CodeChallenge, *code.CodeChallengeMethod) {
			return nil, NewError(
				ErrInvalidRequest,
				WithDescription("invalid code_verifier"),
			)
		}
	}

	if code.Scopes.Contains(coredata.OAuth2ScopeOpenID) {
		var (
			idTokenClaims = NewIDTokenClaims(
				s.baseURL,
				code.IdentityID,
				client.ID,
				code.AuthTime,
				code.Scopes,
				ref.UnrefOrZero(code.Nonce),
				accessTokenValue,
				"", false, "",
				s.accessTokenDuration,
			)
			sk  = s.signingKey()
			err error
		)

		idToken, err = jose.SignJWT(sk.PrivateKey, sk.KID, idTokenClaims)
		if err != nil {
			return nil, fmt.Errorf("cannot sign id token: %w", err)
		}
	}

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			accessToken := &coredata.OAuth2AccessToken{
				ID:          fmt.Sprintf("oat_%s", accessTokenValue[:16]),
				HashedValue: hash.SHA256String(accessTokenValue),
				ClientID:    client.ID,
				IdentityID:  code.IdentityID,
				Scopes:      code.Scopes,
				CreatedAt:   now,
				ExpiresAt:   accessTokenExpiresAt,
			}

			if err := accessToken.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot create access token: %w", err)
			}

			if client.HasGrantType(coredata.OAuth2GrantTypeRefreshToken) {
				refreshToken := &coredata.OAuth2RefreshToken{
					ID:            fmt.Sprintf("ort_%s", refreshTokenValue[:16]),
					HashedValue:   hash.SHA256String(refreshTokenValue),
					ClientID:      client.ID,
					IdentityID:    code.IdentityID,
					Scopes:        code.Scopes,
					AccessTokenID: accessToken.ID,
					CreatedAt:     now,
					ExpiresAt:     now.Add(s.refreshTokenDuration),
				}

				if err := refreshToken.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot create refresh token: %w", err)
				}
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &TokenResult{
		AccessToken:  accessTokenValue,
		TokenType:    tokenTypeBearer,
		ExpiresIn:    int64(time.Until(accessTokenExpiresAt).Seconds()),
		RefreshToken: refreshTokenValue,
		Scope:        code.Scopes.String(),
		IDToken:      idToken,
	}, nil
}

func (s *Service) RefreshToken(
	ctx context.Context,
	client *coredata.OAuth2Client,
	refreshTokenValue string,
) (*TokenResult, error) {
	var (
		accessTokenValue     = rand.MustHexString(tokenByteLength)
		refreshTokenValueNew = rand.MustHexString(refreshTokenByteLength)
		hashedValue          = hash.SHA256String(refreshTokenValue)
		now                  = time.Now()
		accessTokenExpiresAt = now.Add(s.accessTokenDuration)
		idToken              string
		previousRefreshToken = coredata.OAuth2RefreshToken{}
	)

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := previousRefreshToken.LoadByHashedValueForUpdate(
				ctx,
				tx,
				hashedValue,
				client.ID,
			); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewError(
						ErrInvalidGrant,
						WithDescription("refresh token not found"),
					)
				}

				return fmt.Errorf("cannot load refresh token: %w", err)
			}

			if previousRefreshToken.RevokedAt != nil {
				s.logger.WarnCtx(
					ctx,
					"refresh token replay detected, revoking all tokens",
					log.String("client_id", client.ID.String()),
					log.String("identity_id", previousRefreshToken.IdentityID.String()),
				)

				accessToken := &coredata.OAuth2AccessToken{}
				if _, err := accessToken.DeleteByClientAndIdentity(
					ctx,
					tx,
					client.ID,
					previousRefreshToken.IdentityID,
				); err != nil {
					s.logger.ErrorCtx(
						ctx,
						"cannot delete access tokens",
						log.String("access_token_id", previousRefreshToken.AccessTokenID),
						log.Error(err),
					)
				}

				refreshToken := &coredata.OAuth2RefreshToken{}
				if _, err := refreshToken.RevokeByClientAndIdentity(
					ctx,
					tx,
					client.ID,
					previousRefreshToken.IdentityID,
					now,
				); err != nil {
					s.logger.ErrorCtx(
						ctx,
						"cannot revoke refresh tokens",
						log.String("refresh_token_id", previousRefreshToken.ID),
						log.Error(err),
					)
				}

				return pg.NoRollback(
					NewError(
						ErrInvalidGrant,
						WithDescription("refresh token replay detected"),
					),
				)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	if now.After(previousRefreshToken.ExpiresAt) {
		return nil, NewError(
			ErrInvalidGrant,
			WithDescription("refresh token expired"),
		)
	}

	if previousRefreshToken.Scopes.Contains(coredata.OAuth2ScopeOpenID) {
		var (
			claims = NewIDTokenClaims(
				s.baseURL,
				previousRefreshToken.IdentityID,
				client.ID,
				time.Now(),
				previousRefreshToken.Scopes,
				"",
				accessTokenValue,
				"", false, "",
				s.accessTokenDuration,
			)
			sk  = s.signingKey()
			err error
		)

		idToken, err = jose.SignJWT(sk.PrivateKey, sk.KID, claims)
		if err != nil {
			return nil, fmt.Errorf("cannot sign id token: %w", err)
		}
	}

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := previousRefreshToken.Revoke(ctx, tx, now); err != nil {
				return fmt.Errorf("cannot revoke previous refresh token: %w", err)
			}

			// Attempt to delete the previous (legacy) access token.
			// If this fails, ignore the error; access tokens are short-lived and already
			// unlinked from refresh tokens.
			legacyAccessToken := coredata.OAuth2AccessToken{ID: previousRefreshToken.AccessTokenID}
			if err := legacyAccessToken.Delete(ctx, tx); err != nil {
				s.logger.ErrorCtx(
					ctx,
					"cannot delete legacy access token",
					log.String("access_token_id", previousRefreshToken.AccessTokenID),
					log.Error(err),
				)
			}

			accessToken := &coredata.OAuth2AccessToken{
				ID:          fmt.Sprintf("oat_%s", accessTokenValue[:16]),
				HashedValue: hash.SHA256String(accessTokenValue),
				ClientID:    client.ID,
				IdentityID:  previousRefreshToken.IdentityID,
				Scopes:      previousRefreshToken.Scopes,
				CreatedAt:   now,
				ExpiresAt:   accessTokenExpiresAt,
			}
			if err := accessToken.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot create access token: %w", err)
			}

			refreshToken := &coredata.OAuth2RefreshToken{
				ID:            fmt.Sprintf("ort_%s", refreshTokenValueNew[:16]),
				HashedValue:   hash.SHA256String(refreshTokenValueNew),
				ClientID:      client.ID,
				IdentityID:    previousRefreshToken.IdentityID,
				Scopes:        previousRefreshToken.Scopes,
				AccessTokenID: accessToken.ID,
				CreatedAt:     now,
				ExpiresAt:     now.Add(s.refreshTokenDuration),
			}
			if err := refreshToken.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot create refresh token: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &TokenResult{
		AccessToken:  accessTokenValue,
		TokenType:    tokenTypeBearer,
		ExpiresIn:    int64(time.Until(accessTokenExpiresAt).Seconds()),
		RefreshToken: refreshTokenValueNew,
		Scope:        previousRefreshToken.Scopes.String(),
		IDToken:      idToken,
	}, nil
}

func (s *Service) CreateDeviceCode(
	ctx context.Context,
	clientID gid.GID,
	scopes coredata.OAuth2Scopes,
) (string, *coredata.OAuth2DeviceCode, error) {
	var (
		deviceCodeValue = rand.MustHexString(tokenByteLength)
		deviceCode      *coredata.OAuth2DeviceCode
		now             = time.Now()
	)

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			client := coredata.OAuth2Client{}
			if err := client.LoadByID(ctx, tx, coredata.NewNoScope(), clientID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewError(
						ErrInvalidRequest,
						WithDescription("unknown client_id"),
					)
				}

				return fmt.Errorf("cannot load oauth2 client: %w", err)
			}

			if !client.HasGrantType(coredata.OAuth2GrantTypeDeviceCode) {
				return NewError(
					ErrUnauthorizedClient,
					WithDescription("client not authorized for device flow"),
				)
			}

			requestedScopes := scopes.OrDefault(client.Scopes)
			if !client.AreScopesAllowed(requestedScopes) {
				return NewError(
					ErrInvalidScope,
					WithDescription("requested scope exceeds client registration"),
				)
			}

			// Try up to 3 times to generate a unique user code, retrying if we detect a collision on insertion.
			// This minimizes the (rare) chance of user code collisions due to the limited keyspace.
			for range 3 {
				userCode := rand.MustStringFromAlphabet(userCodeAlphabet, 8)

				candidate := &coredata.OAuth2DeviceCode{
					ID:             fmt.Sprintf("odc_%s", deviceCodeValue[:16]),
					DeviceCodeHash: hash.SHA256String(deviceCodeValue),
					UserCode:       coredata.OAuth2UserCode(userCode),
					ClientID:       client.ID,
					Scopes:         requestedScopes,
					Status:         coredata.OAuth2DeviceCodeStatusPending,
					PollInterval:   5,
					CreatedAt:      now,
					ExpiresAt:      now.Add(s.deviceCodeDuration),
				}

				if err := candidate.Insert(ctx, tx); err != nil {
					if errors.Is(err, coredata.ErrResourceAlreadyExists) {
						continue
					}

					return fmt.Errorf("cannot insert device code: %w", err)
				}

				deviceCode = candidate

				return nil
			}

			return fmt.Errorf("cannot generate unique user code after 3 attempts")
		},
	); err != nil {
		return "", nil, err
	}

	return deviceCodeValue, deviceCode, nil
}

func (s *Service) PollDeviceCode(
	ctx context.Context,
	clientID gid.GID,
	deviceCodeValue string,
) (*TokenResult, error) {
	var (
		hashedValue = hash.SHA256String(deviceCodeValue)
		deviceCode  = coredata.OAuth2DeviceCode{}
		now         = time.Now()
		client      = &coredata.OAuth2Client{}
	)

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := deviceCode.LoadByDeviceCodeHashForUpdate(ctx, tx, hashedValue, clientID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewError(
						ErrInvalidGrant,
						WithDescription("invalid device code"),
					)
				}

				return fmt.Errorf("cannot load device code: %w", err)
			}

			if err := client.LoadByID(ctx, tx, coredata.NewNoScope(), clientID); err != nil {
				return fmt.Errorf("cannot load client: %w", err)
			}

			// Rate limiting.
			var slowDown bool
			if deviceCode.LastPolledAt != nil {
				elapsed := now.Sub(ref.UnrefOrZero(deviceCode.LastPolledAt))
				if elapsed < time.Duration(deviceCode.PollInterval)*time.Second {
					deviceCode.PollInterval += 5
					slowDown = true
				}
			}

			deviceCode.LastPolledAt = &now

			if err := deviceCode.Update(ctx, tx); err != nil {
				return fmt.Errorf("cannot update device code: %w", err)
			}

			if slowDown {
				return NewError(
					ErrSlowDown,
					WithDescription("slow down"),
				)
			}

			// Ensure code is deleted whehever what is happening next the code must not be used again.
			if deviceCode.Status == coredata.OAuth2DeviceCodeStatusAuthorized {
				if err := deviceCode.Delete(ctx, tx); err != nil {
					return fmt.Errorf("cannot delete device code: %w", err)
				}
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	if now.After(deviceCode.ExpiresAt) {
		return nil, NewError(
			ErrExpiredToken,
			WithDescription("expired token"),
		)
	}

	switch deviceCode.Status {
	case coredata.OAuth2DeviceCodeStatusPending:
		return nil, NewError(
			ErrAuthorizationPending,
			WithDescription("authorization pending"),
		)
	case coredata.OAuth2DeviceCodeStatusDenied:
		return nil, NewError(
			ErrAccessDenied,
			WithDescription("access denied"),
		)
	case coredata.OAuth2DeviceCodeStatusAuthorized:
		// Continue to issue tokens.
	case coredata.OAuth2DeviceCodeStatusExpired:
		return nil, NewError(
			ErrExpiredToken,
			WithDescription("expired token"),
		)
	default:
		return nil, fmt.Errorf("invalid device code status: %q", deviceCode.Status)
	}

	var (
		accessTokenValue     = rand.MustHexString(tokenByteLength)
		refreshTokenValue    = rand.MustHexString(refreshTokenByteLength)
		accessTokenExpiresAt = now.Add(s.accessTokenDuration)
		idToken              string
	)

	if deviceCode.Scopes.Contains(coredata.OAuth2ScopeOpenID) {
		var (
			claims = NewIDTokenClaims(
				s.baseURL,
				*deviceCode.IdentityID,
				clientID,
				now,
				deviceCode.Scopes,
				"",
				accessTokenValue,
				"", false, "",
				s.accessTokenDuration,
			)
			sk  = s.signingKey()
			err error
		)

		idToken, err = jose.SignJWT(sk.PrivateKey, sk.KID, claims)
		if err != nil {
			return nil, fmt.Errorf("cannot sign id token: %w", err)
		}
	}

	if err = s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			accessToken := &coredata.OAuth2AccessToken{
				ID:          fmt.Sprintf("oat_%s", accessTokenValue[:16]),
				HashedValue: hash.SHA256String(accessTokenValue),
				ClientID:    clientID,
				IdentityID:  *deviceCode.IdentityID,
				Scopes:      deviceCode.Scopes,
				CreatedAt:   now,
				ExpiresAt:   accessTokenExpiresAt,
			}
			if err := accessToken.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot create access token: %w", err)
			}

			if client.HasGrantType(coredata.OAuth2GrantTypeRefreshToken) {
				refreshToken := &coredata.OAuth2RefreshToken{
					ID:            fmt.Sprintf("ort_%s", refreshTokenValue[:16]),
					HashedValue:   hash.SHA256String(refreshTokenValue),
					ClientID:      clientID,
					IdentityID:    *deviceCode.IdentityID,
					Scopes:        deviceCode.Scopes,
					AccessTokenID: accessToken.ID,
					CreatedAt:     now,
					ExpiresAt:     now.Add(s.refreshTokenDuration),
				}

				if err := refreshToken.Insert(ctx, tx); err != nil {
					return fmt.Errorf("cannot create refresh token: %w", err)
				}
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &TokenResult{
		AccessToken:  accessTokenValue,
		TokenType:    tokenTypeBearer,
		ExpiresIn:    int64(accessTokenExpiresAt.Sub(now).Seconds()),
		RefreshToken: refreshTokenValue,
		Scope:        deviceCode.Scopes.String(),
		IDToken:      idToken,
	}, nil
}

func (s *Service) AuthorizeDevice(
	ctx context.Context,
	identityID gid.GID,
	userCode string,
) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			deviceCode := coredata.OAuth2DeviceCode{}

			if err := deviceCode.LoadByUserCodeForUpdate(ctx, tx, userCode); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewError(
						ErrInvalidGrant,
						WithDescription("invalid user code"),
					)
				}

				return fmt.Errorf("cannot load device code: %w", err)
			}

			if time.Now().After(deviceCode.ExpiresAt) {
				return NewError(
					ErrExpiredToken,
					WithDescription("expired token"),
				)
			}

			if deviceCode.Status != coredata.OAuth2DeviceCodeStatusPending {
				return NewError(
					ErrInvalidGrant,
					WithDescription(fmt.Sprintf("device code already %s", deviceCode.Status)),
				)
			}

			deviceCode.Status = coredata.OAuth2DeviceCodeStatusAuthorized
			deviceCode.IdentityID = &identityID

			if err := deviceCode.Update(ctx, tx); err != nil {
				return fmt.Errorf("cannot update device code: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) RegisterClient(
	ctx context.Context,
	req *RegisterClientRequest,
) (gid.GID, string, error) {
	for _, u := range req.RedirectURIs {
		parsed, _ := url.Parse(u.String())

		switch req.Visibility {
		case coredata.OAuth2ClientVisibilityPublic:
			if parsed.Scheme != "https" {
				return gid.Nil,
					"",
					NewError(
						ErrInvalidRequest,
						WithDescription("public clients require https redirect_uris"),
					)
			}
		case coredata.OAuth2ClientVisibilityPrivate:
			if parsed.Scheme == "http" {
				if !net.IsLoopback(parsed.Hostname()) {
					return gid.Nil,
						"",
						NewError(
							ErrInvalidRequest,
							WithDescription("http redirect_uris are only allowed for localhost"),
						)
				}
			} else if parsed.Scheme != "https" {
				return gid.Nil,
					"",
					NewError(
						ErrInvalidRequest,
						WithDescription(fmt.Sprintf("unsupported redirect_uri scheme: %s", parsed.Scheme)),
					)
			}
		}
	}

	var (
		plaintextSecret string
		secretHash      []byte
	)

	if req.TokenEndpointAuthMethod != coredata.OAuth2ClientTokenEndpointAuthMethodNone {
		plaintextSecret = rand.MustHexString(tokenByteLength)
		secretHash = hash.SHA256String(plaintextSecret)
	}

	var (
		now    = time.Now()
		scope  = coredata.NewScopeFromObjectID(req.OrganizationID)
		client = &coredata.OAuth2Client{
			ID:                      gid.New(scope.GetTenantID(), coredata.OAuth2ClientEntityType),
			OrganizationID:          req.OrganizationID,
			ClientSecretHash:        secretHash,
			ClientName:              req.ClientName,
			Visibility:              req.Visibility,
			RedirectURIs:            req.RedirectURIs,
			Scopes:                  req.Scopes,
			GrantTypes:              req.GrantTypes,
			ResponseTypes:           req.ResponseTypes,
			TokenEndpointAuthMethod: req.TokenEndpointAuthMethod,
			LogoURI:                 req.LogoURI,
			ClientURI:               req.ClientURI,
			Contacts:                req.Contacts,
			CreatedAt:               now,
			UpdatedAt:               now,
		}
	)

	// Check org membership.
	var membership coredata.Membership
	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := membership.LoadActiveByIdentityIDAndOrganizationID(
				ctx,
				tx,
				req.IdentityID,
				req.OrganizationID,
			); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewError(
						ErrAccessDenied,
						WithDescription("not a member of the organization"),
					)
				}

				return fmt.Errorf("cannot load membership: %w", err)
			}

			if err := client.Insert(ctx, tx, scope); err != nil {
				return fmt.Errorf("cannot insert oauth2 client: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return gid.Nil, "", err
	}

	return client.ID, plaintextSecret, nil
}

func (s *Service) LoadAccessToken(ctx context.Context, tokenValue string) (*coredata.OAuth2AccessToken, error) {
	var (
		hashedValue = hash.SHA256String(tokenValue)
		token       coredata.OAuth2AccessToken
		now         = time.Now()
	)

	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, tx pg.Querier) error {
			if err := token.LoadByHashedValue(ctx, tx, hashedValue); err != nil {
				return fmt.Errorf("cannot load access token: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	if now.After(token.ExpiresAt) {
		return nil, fmt.Errorf("access token expired")
	}

	return &token, nil
}

func (s *Service) IntrospectToken(ctx context.Context, clientID gid.GID, tokenValue string) (*coredata.OAuth2AccessToken, error) {
	var (
		hashedValue = hash.SHA256String(tokenValue)
		token       = coredata.OAuth2AccessToken{}
		now         = time.Now()
	)

	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := token.LoadByHashedValueAndClientID(ctx, conn, hashedValue, clientID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load access token: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	if token.ID == "" || now.After(token.ExpiresAt) {
		return nil, nil
	}

	return &token, nil
}

func (s *Service) UserInfo(
	ctx context.Context,
	identityID gid.GID,
	scopes coredata.OAuth2Scopes,
) (map[string]any, error) {
	identity := &coredata.Identity{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := identity.LoadByID(ctx, conn, identityID); err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	claims := map[string]any{
		"sub": identity.ID.String(),
	}

	for _, scope := range scopes {
		switch scope {
		case coredata.OAuth2ScopeEmail:
			claims["email"] = identity.EmailAddress.String()
			claims["email_verified"] = identity.EmailAddressVerified
		case coredata.OAuth2ScopeProfile:
			claims["name"] = identity.FullName
		}
	}

	return claims, nil
}

func (s *Service) RevokeToken(ctx context.Context, clientID gid.GID, tokenValue string) error {
	hashedValue := hash.SHA256String(tokenValue)

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			accessToken := coredata.OAuth2AccessToken{}
			if err := accessToken.LoadByHashedValueAndClientID(ctx, tx, hashedValue, clientID); err != nil {
				if !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load access token: %w", err)
				}
			} else {
				if err := accessToken.Delete(ctx, tx); err != nil {
					return fmt.Errorf("cannot delete access token: %w", err)
				}

				return nil
			}

			refreshToken := coredata.OAuth2RefreshToken{}
			if err := refreshToken.LoadByHashedValueAndClientID(ctx, tx, hashedValue, clientID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load refresh token: %w", err)
			}

			if err := refreshToken.Revoke(ctx, tx, time.Now()); err != nil {
				return fmt.Errorf("cannot revoke refresh token: %w", err)
			}

			return nil
		},
	); err != nil {
		return err
	}

	return nil
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
		return "", NewError(ErrInvalidClient, WithDescription("cannot load client"))
	}

	if !client.IsRedirectURIAllowed(req.RedirectURI) {
		return "", ErrInvalidRedirectURI
	}

	if client.Visibility == coredata.OAuth2ClientVisibilityPrivate {
		var membership coredata.Membership
		err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
			return membership.LoadActiveByIdentityIDAndOrganizationID(
				ctx,
				conn,
				req.IdentityID,
				client.OrganizationID,
			)
		})
		if err != nil {
			return "", NewError(ErrUnauthorizedClient, WithDescription("client is private and user is not a member of the organization"))
		}
	}

	if req.ResponseType != coredata.OAuth2ResponseTypeCode {
		return "", NewError(ErrInvalidRequest, WithDescription("unsupported response_type"))
	}

	requestedScopes := req.Scopes.OrDefault(client.Scopes)
	if !client.AreScopesAllowed(requestedScopes) {
		return "", NewError(ErrInvalidScope, WithDescription("requested scope exceeds client registration"))
	}

	codeChallengeMethod := req.CodeChallengeMethod
	if client.TokenEndpointAuthMethod == coredata.OAuth2ClientTokenEndpointAuthMethodNone && req.CodeChallenge == "" {
		return "", NewError(ErrInvalidRequest, WithDescription("code_challenge required for public clients"))
	}

	if codeChallengeMethod != "" && codeChallengeMethod != coredata.OAuth2CodeChallengeMethodS256 {
		return "", NewError(ErrInvalidRequest, WithDescription("only S256 code_challenge_method is supported"))
	}

	if req.CodeChallenge != "" && codeChallengeMethod == "" {
		codeChallengeMethod = coredata.OAuth2CodeChallengeMethodS256
	}

	// Check existing approved consent with matching scopes.
	var existingConsent coredata.OAuth2Consent
	consentErr := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
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
			uri.URI(req.RedirectURI),
			requestedScopes,
			req.CodeChallenge,
			codeChallengeMethod,
			req.Nonce,
			req.AuthTime,
		)
		if err != nil {
			return "", NewError(ErrServerError, WithError(err))
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
		RedirectURI:         uri.URI(req.RedirectURI),
		CodeChallenge:       req.CodeChallenge,
		CodeChallengeMethod: codeChallengeMethod,
		Nonce:               req.Nonce,
		State:               req.State,
		Approved:            false,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return pendingConsent.Insert(ctx, tx)
	}); err != nil {
		return "", NewError(ErrServerError, WithDescription("cannot create pending consent"))
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
	if err := s.pg.WithConn(ctx, func(ctx context.Context, conn pg.Querier) error {
		return consent.LoadByID(ctx, conn, req.ConsentID)
	}); err != nil {
		return "", "", "", NewError(ErrInvalidRequest, WithDescription("consent not found"))
	}

	if consent.IdentityID != req.IdentityID {
		return "", "", "", NewError(ErrAccessDenied, WithDescription("consent does not belong to this identity"))
	}

	if consent.Approved {
		return "", "", "", NewError(ErrInvalidRequest, WithDescription("consent already processed"))
	}

	client, err := s.GetClientByID(ctx, consent.ClientID)
	if err != nil {
		return "", "", "", NewError(ErrInvalidClient, WithDescription("cannot load client"))
	}

	if !client.IsRedirectURIAllowed(string(consent.RedirectURI)) {
		return "", "", "", ErrInvalidRedirectURI
	}

	redirectURI := string(consent.RedirectURI)

	if !req.Approved {
		// Clean up the pending consent.
		_ = s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			return consent.Delete(ctx, tx)
		})
		return "", redirectURI, consent.State, NewError(ErrAccessDenied, WithDescription("user denied the request"))
	}

	// Mark consent as approved.
	consent.Approved = true
	consent.UpdatedAt = time.Now()
	if err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return consent.Update(ctx, tx)
	}); err != nil {
		return "", redirectURI, consent.State, NewError(ErrServerError, WithDescription("cannot approve consent"))
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
		return "", redirectURI, consent.State, err
	}

	return code, redirectURI, consent.State, nil
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
		return nil, NewError(ErrInvalidClient, WithDescription("cannot load client"))
	}

	if client.TokenEndpointAuthMethod == coredata.OAuth2ClientTokenEndpointAuthMethodNone {
		return client, nil
	}

	if clientSecret == "" {
		return nil, NewError(ErrInvalidClient, WithDescription("missing client_secret"))
	}

	if subtle.ConstantTimeCompare(client.ClientSecretHash, hash.SHA256String(clientSecret)) != 1 {
		return nil, NewError(ErrInvalidClient, WithDescription("invalid client_secret"))
	}

	return client, nil
}

// issueAuthorizationCode generates and persists an authorization code,
// returning the plaintext code value.
func (s *Service) issueAuthorizationCode(
	ctx context.Context,
	client *coredata.OAuth2Client,
	identityID gid.GID,
	redirectURI uri.URI,
	scopes coredata.OAuth2Scopes,
	codeChallenge string,
	codeChallengeMethod coredata.OAuth2CodeChallengeMethod,
	nonce string,
	authTime time.Time,
) (string, error) {
	codeValue := rand.MustHexString(tokenByteLength)

	codeHash := hash.SHA256HexString(codeValue)
	now := time.Now()

	code := &coredata.OAuth2AuthorizationCode{
		ID:          codeHash,
		ClientID:    client.ID,
		IdentityID:  identityID,
		RedirectURI: redirectURI,
		Scopes:      scopes,
		AuthTime:    authTime,
		CreatedAt:   now,
		ExpiresAt:   now.Add(s.authorizationCodeDuration),
	}

	if codeChallenge != "" {
		code.CodeChallenge = &codeChallenge
		code.CodeChallengeMethod = &codeChallengeMethod
	}

	if nonce != "" {
		code.Nonce = &nonce
	}

	if err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return code.Insert(ctx, tx)
	}); err != nil {
		return "", fmt.Errorf("cannot save authorization code: %w", err)
	}

	return codeValue, nil
}

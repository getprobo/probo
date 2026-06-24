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
	"go.probo.inc/probo/pkg/iam/oauth2scope"
	"go.probo.inc/probo/pkg/netx"
	"go.probo.inc/probo/pkg/page"
	"go.probo.inc/probo/pkg/uri"
)

// CLIClientID is the well-known OAuth2 client ID for the Probo CLI.
// It is inserted into every Probo database via migration and hardcoded
// in the CLI binary for the device authorization flow.
var CLIClientID = gid.MustParseGID("AAAAAAAAAAAASwAAAAAAAAAAcHJiY2xp")

const (
	tokenByteLength           = 32
	refreshTokenByteLength    = 48
	tokenTypeBearer           = "Bearer"
	oauthGrantAccessTokenName = "OAuth grant"

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
		cimd                      *cimdFetcher
		cimdAllowedClientIDs      []string
		registry                  *oauth2scope.Registry
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
		ClientIDRaw         string
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
		SessionID  gid.GID
		Approved   bool
		AuthTime   time.Time
	}

	RegisterClientRequest struct {
		IdentityID              gid.GID
		OrganizationID          *gid.GID
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

	IntrospectResult struct {
		ClientID   gid.GID
		IdentityID gid.GID
		Scopes     coredata.OAuth2Scopes
		IssuedAt   time.Time
		ExpiresAt  time.Time
		TokenType  string
	}

	CreateManualAccessTokenRequest struct {
		IdentityID gid.GID
		Name       string
		ExpiresAt  time.Time
		Scopes     coredata.OAuth2Scopes
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

func WithRegistry(registry *oauth2scope.Registry) Option {
	return func(s *Service) {
		s.registry = registry
	}
}

func WithCIMDAllowedClientIDs(clientIDs []string) Option {
	return func(s *Service) {
		s.cimdAllowedClientIDs = clientIDs
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
	s.cimd = newCIMDFetcher(logger.Named("cimd"))

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
		ID:          gid.New(clientID.TenantID(), coredata.OAuth2AccessTokenEntityType),
		Name:        oauthGrantAccessTokenName,
		HashedValue: hash.SHA256String(tokenValue),
		ClientID:    new(clientID),
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
		identity             = coredata.Identity{}
		now                  = time.Now()
		accessTokenExpiresAt = now.Add(s.accessTokenDuration)
		accessTokenValue     = rand.MustHexString(tokenByteLength)
		accessTokenID        = gid.New(client.ID.TenantID(), coredata.OAuth2AccessTokenEntityType)
		refreshTokenValue    string
		idToken              string
	)

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := code.LoadByHashForUpdate(ctx, tx, hash.SHA256String(codeValue), client.ID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewError(ErrInvalidGrant, WithDescription("authorization code not found"))
				}

				return fmt.Errorf("cannot load authorization code: %w", err)
			}

			// RFC 6819 §5.2.1.1: if the code was already redeemed, this is
			// a replay attack. Revoke all tokens derived from this code.
			if code.RedeemedAt != nil {
				s.logger.WarnCtx(
					ctx,
					"authorization code replay detected, revoking derived tokens",
					log.String("client_id", client.ID.String()),
					log.String("identity_id", code.IdentityID.String()),
				)

				if code.AccessTokenID != nil {
					derivedAccessToken := coredata.OAuth2AccessToken{ID: *code.AccessTokenID}
					if err := derivedAccessToken.Delete(ctx, tx); err != nil {
						s.logger.ErrorCtx(
							ctx,
							"cannot delete derived access token",
							log.String("access_token_id", code.AccessTokenID.String()),
							log.Error(err),
						)
					}

					derivedRefreshToken := &coredata.OAuth2RefreshToken{}
					if _, err := derivedRefreshToken.RevokeByAccessTokenID(ctx, tx, *code.AccessTokenID, now); err != nil {
						s.logger.ErrorCtx(
							ctx,
							"cannot revoke derived refresh tokens",
							log.String("access_token_id", code.AccessTokenID.String()),
							log.Error(err),
						)
					}
				}

				return pg.NoRollback(
					NewError(
						ErrInvalidGrant,
						WithDescription("authorization code already redeemed"),
					),
				)
			}

			if err := identity.LoadByID(ctx, tx, code.IdentityID); err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
			}

			if err := code.Redeem(ctx, tx, now, accessTokenID); err != nil {
				return fmt.Errorf("cannot redeem authorization code: %w", err)
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

	if code.Scopes.Contains(ScopeOpenID) {
		var (
			idTokenClaims = NewIDTokenClaims(
				s.baseURL,
				code.IdentityID,
				client.ID,
				code.AuthTime,
				code.Scopes,
				ref.UnrefOrZero(code.Nonce),
				accessTokenValue,
				identity.EmailAddress.String(),
				identity.EmailAddressVerified,
				identity.FullName,
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
				ID:          accessTokenID,
				Name:        oauthGrantAccessTokenName,
				HashedValue: hash.SHA256String(accessTokenValue),
				ClientID:    new(client.ID),
				IdentityID:  code.IdentityID,
				Scopes:      code.Scopes,
				CreatedAt:   now,
				ExpiresAt:   accessTokenExpiresAt,
			}

			if err := accessToken.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot create access token: %w", err)
			}

			if client.HasGrantType(coredata.OAuth2GrantTypeRefreshToken) && code.Scopes.Contains(ScopeOfflineAccess) {
				refreshTokenValue = rand.MustHexString(refreshTokenByteLength)

				refreshToken := &coredata.OAuth2RefreshToken{
					ID:            gid.New(client.ID.TenantID(), coredata.OAuth2RefreshTokenEntityType),
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
		identity             = coredata.Identity{}
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

			if err := identity.LoadByID(ctx, tx, previousRefreshToken.IdentityID); err != nil {
				return fmt.Errorf("cannot load identity: %w", err)
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
						log.String("access_token_id", previousRefreshToken.AccessTokenID.String()),
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
						log.String("refresh_token_id", previousRefreshToken.ID.String()),
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

	if previousRefreshToken.Scopes.Contains(ScopeOpenID) {
		var (
			claims = NewIDTokenClaims(
				s.baseURL,
				previousRefreshToken.IdentityID,
				client.ID,
				time.Now(),
				previousRefreshToken.Scopes,
				"",
				accessTokenValue,
				identity.EmailAddress.String(),
				identity.EmailAddressVerified,
				identity.FullName,
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
					log.String("access_token_id", previousRefreshToken.AccessTokenID.String()),
					log.Error(err),
				)
			}

			accessToken := &coredata.OAuth2AccessToken{
				ID:          gid.New(client.ID.TenantID(), coredata.OAuth2AccessTokenEntityType),
				Name:        oauthGrantAccessTokenName,
				HashedValue: hash.SHA256String(accessTokenValue),
				ClientID:    new(client.ID),
				IdentityID:  previousRefreshToken.IdentityID,
				Scopes:      previousRefreshToken.Scopes,
				CreatedAt:   now,
				ExpiresAt:   accessTokenExpiresAt,
			}
			if err := accessToken.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot create access token: %w", err)
			}

			refreshToken := &coredata.OAuth2RefreshToken{
				ID:            gid.New(client.ID.TenantID(), coredata.OAuth2RefreshTokenEntityType),
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

			if requestedScopes.Contains(ScopeOfflineAccess) && !client.HasGrantType(coredata.OAuth2GrantTypeRefreshToken) {
				return NewError(
					ErrInvalidScope,
					WithDescription("offline_access requires the refresh_token grant type"),
				)
			}

			// Try up to 3 times to generate a unique user code, retrying if we detect a collision on insertion.
			// This minimizes the (rare) chance of user code collisions due to the limited keyspace.
			for range 3 {
				userCode := rand.MustStringFromAlphabet(userCodeAlphabet, 8)

				candidate := &coredata.OAuth2DeviceCode{
					ID:             gid.New(client.ID.TenantID(), coredata.OAuth2DeviceCodeEntityType),
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
		identity    = coredata.Identity{}
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

			if deviceCode.IdentityID != nil {
				if err := identity.LoadByID(ctx, tx, *deviceCode.IdentityID); err != nil {
					return fmt.Errorf("cannot load identity: %w", err)
				}
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
		refreshTokenValue    string
		accessTokenExpiresAt = now.Add(s.accessTokenDuration)
		idToken              string
	)

	if deviceCode.Scopes.Contains(ScopeOpenID) {
		var (
			claims = NewIDTokenClaims(
				s.baseURL,
				*deviceCode.IdentityID,
				clientID,
				now,
				deviceCode.Scopes,
				"",
				accessTokenValue,
				identity.EmailAddress.String(),
				identity.EmailAddressVerified,
				identity.FullName,
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
				ID:          gid.New(clientID.TenantID(), coredata.OAuth2AccessTokenEntityType),
				Name:        oauthGrantAccessTokenName,
				HashedValue: hash.SHA256String(accessTokenValue),
				ClientID:    new(clientID),
				IdentityID:  *deviceCode.IdentityID,
				Scopes:      deviceCode.Scopes,
				CreatedAt:   now,
				ExpiresAt:   accessTokenExpiresAt,
			}
			if err := accessToken.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot create access token: %w", err)
			}

			if client.HasGrantType(coredata.OAuth2GrantTypeRefreshToken) && deviceCode.Scopes.Contains(ScopeOfflineAccess) {
				refreshTokenValue = rand.MustHexString(refreshTokenByteLength)

				refreshToken := &coredata.OAuth2RefreshToken{
					ID:            gid.New(clientID.TenantID(), coredata.OAuth2RefreshTokenEntityType),
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
	sessionID gid.GID,
	userCode string,
) error {
	var (
		deviceCode coredata.OAuth2DeviceCode
		client     coredata.OAuth2Client
	)

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
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

			if err := client.LoadByID(ctx, tx, coredata.NewNoScope(), deviceCode.ClientID); err != nil {
				return fmt.Errorf("cannot load oauth2 client: %w", err)
			}

			// RFC 6819 §5.2.3.2 / §5.2.4.1: public clients must always
			// require explicit user consent since they cannot be strongly
			// authenticated.
			if client.TokenEndpointAuthMethod != coredata.OAuth2ClientTokenEndpointAuthMethodNone {
				var existingConsent coredata.OAuth2Consent
				if err := existingConsent.LoadMatchingConsent(
					ctx,
					tx,
					identityID,
					client.ID,
					deviceCode.Scopes,
				); err == nil {
					deviceCode.Status = coredata.OAuth2DeviceCodeStatusAuthorized
					deviceCode.IdentityID = &identityID

					if err := deviceCode.Update(ctx, tx); err != nil {
						return fmt.Errorf("cannot update device code: %w", err)
					}

					return nil
				}
			}

			now := time.Now()
			pendingConsent := &coredata.OAuth2Consent{
				ID:           gid.New(client.ID.TenantID(), coredata.OAuth2ConsentEntityType),
				IdentityID:   identityID,
				SessionID:    sessionID,
				ClientID:     client.ID,
				Scopes:       deviceCode.Scopes,
				DeviceCodeID: &deviceCode.ID,
				Approved:     false,
				CreatedAt:    now,
				UpdatedAt:    now,
			}

			if err := pendingConsent.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert pending consent: %w", err)
			}

			return pg.NoRollback(
				&ConsentRequiredError{
					ConsentID: pendingConsent.ID,
					Client:    &client,
					Scopes:    deviceCode.Scopes,
				},
			)
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
				if !netx.IsLoopback(parsed.Hostname()) {
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

	if req.OrganizationID == nil {
		return gid.Nil, "", NewError(
			ErrInvalidRequest,
			WithDescription("organization_id is required"),
		)
	}

	var (
		now    = time.Now()
		scope  = coredata.NewScopeFromObjectID(*req.OrganizationID)
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

	var membership coredata.Membership

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := membership.LoadActiveByIdentityIDAndOrganizationID(
				ctx,
				tx,
				req.IdentityID,
				*req.OrganizationID,
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

func (s *Service) IntrospectToken(
	ctx context.Context,
	clientID gid.GID,
	tokenValue string,
	tokenTypeHint *coredata.OAuth2TokenTypeHint,
) (*IntrospectResult, error) {
	var (
		hashedValue  = hash.SHA256String(tokenValue)
		now          = time.Now()
		accessToken  = coredata.OAuth2AccessToken{}
		refreshToken = coredata.OAuth2RefreshToken{}
		hasAccess    bool
		hasRefresh   bool
	)

	loadAccess := func(ctx context.Context, conn pg.Querier) error {
		if err := accessToken.LoadByHashedValueAndClientID(ctx, conn, hashedValue, clientID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return nil
			}

			return fmt.Errorf("cannot load access token: %w", err)
		}

		hasAccess = true

		return nil
	}

	loadRefresh := func(ctx context.Context, conn pg.Querier) error {
		if err := refreshToken.LoadByHashedValueAndClientID(ctx, conn, hashedValue, clientID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return nil
			}

			return fmt.Errorf("cannot load refresh token: %w", err)
		}

		hasRefresh = true

		return nil
	}

	preferRefresh := tokenTypeHint != nil && *tokenTypeHint == coredata.OAuth2TokenTypeHintRefreshToken

	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if preferRefresh {
				if err := loadRefresh(ctx, conn); err != nil {
					return err
				}

				if hasRefresh {
					return nil
				}

				return loadAccess(ctx, conn)
			}

			if err := loadAccess(ctx, conn); err != nil {
				return err
			}

			if hasAccess {
				return nil
			}

			return loadRefresh(ctx, conn)
		},
	); err != nil {
		return nil, err
	}

	switch {
	case hasAccess:
		if now.After(accessToken.ExpiresAt) {
			return nil, nil
		}

		var resultClientID gid.GID
		if accessToken.ClientID != nil {
			resultClientID = *accessToken.ClientID
		}

		return &IntrospectResult{
			ClientID:   resultClientID,
			IdentityID: accessToken.IdentityID,
			Scopes:     accessToken.Scopes,
			IssuedAt:   accessToken.CreatedAt,
			ExpiresAt:  accessToken.ExpiresAt,
			TokenType:  tokenTypeBearer,
		}, nil
	case hasRefresh:
		if refreshToken.RevokedAt != nil || now.After(refreshToken.ExpiresAt) {
			return nil, nil
		}

		return &IntrospectResult{
			ClientID:   refreshToken.ClientID,
			IdentityID: refreshToken.IdentityID,
			Scopes:     refreshToken.Scopes,
			IssuedAt:   refreshToken.CreatedAt,
			ExpiresAt:  refreshToken.ExpiresAt,
		}, nil
	default:
		return nil, nil
	}
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
		case ScopeEmail:
			claims["email"] = identity.EmailAddress.String()
			claims["email_verified"] = identity.EmailAddressVerified
		case ScopeProfile:
			claims["name"] = identity.FullName
		}
	}

	return claims, nil
}

func (s *Service) RevokeToken(
	ctx context.Context,
	clientID gid.GID,
	tokenValue string,
	tokenTypeHint *coredata.OAuth2TokenTypeHint,
) error {
	if tokenValue == "" {
		return nil
	}

	hashedValue := hash.SHA256String(tokenValue)

	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if tokenTypeHint != nil && *tokenTypeHint == coredata.OAuth2TokenTypeHintRefreshToken {
				refreshToken := coredata.OAuth2RefreshToken{}

				err := refreshToken.LoadByHashedValueAndClientID(ctx, tx, hashedValue, clientID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load refresh token: %w", err)
				}

				if err == nil {
					now := time.Now()
					if err := refreshToken.Revoke(ctx, tx, now); err != nil {
						return fmt.Errorf("cannot revoke refresh token: %w", err)
					}

					if refreshToken.AccessTokenID != gid.Nil {
						at := coredata.OAuth2AccessToken{ID: refreshToken.AccessTokenID}
						if err := at.Delete(ctx, tx); err != nil {
							return fmt.Errorf("cannot delete linked access token: %w", err)
						}
					}

					return nil
				}

				accessToken := coredata.OAuth2AccessToken{}

				err = accessToken.LoadByHashedValueAndClientID(ctx, tx, hashedValue, clientID)
				if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
					return fmt.Errorf("cannot load access token: %w", err)
				}

				if err == nil {
					if err := accessToken.Delete(ctx, tx); err != nil {
						return fmt.Errorf("cannot delete access token: %w", err)
					}
				}

				return nil
			}

			accessToken := coredata.OAuth2AccessToken{}

			err := accessToken.LoadByHashedValueAndClientID(ctx, tx, hashedValue, clientID)
			if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load access token: %w", err)
			}

			if err == nil {
				if err := accessToken.Delete(ctx, tx); err != nil {
					return fmt.Errorf("cannot delete access token: %w", err)
				}

				return nil
			}

			refreshToken := coredata.OAuth2RefreshToken{}

			err = refreshToken.LoadByHashedValueAndClientID(ctx, tx, hashedValue, clientID)
			if err != nil && !errors.Is(err, coredata.ErrResourceNotFound) {
				return fmt.Errorf("cannot load refresh token: %w", err)
			}

			if err == nil {
				now := time.Now()
				if err := refreshToken.Revoke(ctx, tx, now); err != nil {
					return fmt.Errorf("cannot revoke refresh token: %w", err)
				}

				if refreshToken.AccessTokenID != gid.Nil {
					at := coredata.OAuth2AccessToken{ID: refreshToken.AccessTokenID}
					if err := at.Delete(ctx, tx); err != nil {
						return fmt.Errorf("cannot delete linked access token: %w", err)
					}
				}
			}

			return nil
		},
	)
}

func (s *Service) Authorize(
	ctx context.Context,
	req *AuthorizeRequest,
) (string, error) {
	var code string

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			client, err := s.resolveClient(ctx, tx, req.ClientIDRaw)
			if err != nil {
				return err
			}

			if !client.IsRedirectURIAllowed(req.RedirectURI) {
				return ErrInvalidRedirectURI
			}

			if client.Visibility == coredata.OAuth2ClientVisibilityPrivate {
				if client.OrganizationID == nil {
					return fmt.Errorf("cannot authorize: private client has no organization")
				}

				var membership coredata.Membership
				if err := membership.LoadActiveByIdentityIDAndOrganizationID(
					ctx,
					tx,
					req.IdentityID,
					*client.OrganizationID,
				); err != nil {
					if errors.Is(err, coredata.ErrResourceNotFound) {
						return ErrUnauthorizedMember
					}

					return fmt.Errorf("cannot check membership: %w", err)
				}
			}

			if req.ResponseType != coredata.OAuth2ResponseTypeCode {
				return fmt.Errorf("cannot authorize: unsupported response_type")
			}

			requestedScopes := req.Scopes.OrDefault(client.Scopes)
			if !client.AreScopesAllowed(requestedScopes) {
				return fmt.Errorf("cannot authorize: requested scope exceeds client registration")
			}

			if requestedScopes.Contains(ScopeOfflineAccess) && !client.HasGrantType(coredata.OAuth2GrantTypeRefreshToken) {
				return NewError(
					ErrInvalidScope,
					WithDescription("offline_access requires the refresh_token grant type"),
				)
			}

			codeChallengeMethod := req.CodeChallengeMethod
			if client.TokenEndpointAuthMethod == coredata.OAuth2ClientTokenEndpointAuthMethodNone && req.CodeChallenge == "" {
				return fmt.Errorf("cannot authorize: code_challenge required for public clients")
			}

			if codeChallengeMethod != "" && codeChallengeMethod != coredata.OAuth2CodeChallengeMethodS256 {
				return fmt.Errorf("cannot authorize: only S256 code_challenge_method is supported")
			}

			if req.CodeChallenge != "" && codeChallengeMethod == "" {
				codeChallengeMethod = coredata.OAuth2CodeChallengeMethodS256
			}

			// RFC 6819 §5.2.3.2 / §5.2.4.1: public clients must always require
			// explicit user consent since they cannot be strongly authenticated.
			if client.TokenEndpointAuthMethod != coredata.OAuth2ClientTokenEndpointAuthMethodNone {
				var existingConsent coredata.OAuth2Consent
				if err := existingConsent.LoadMatchingConsent(
					ctx,
					tx,
					req.IdentityID,
					client.ID,
					requestedScopes,
				); err == nil {
					var err error

					code, err = s.issueAuthorizationCode(
						ctx,
						tx,
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
						return fmt.Errorf("cannot issue authorization code: %w", err)
					}

					return nil
				}
			}

			now := time.Now()
			pendingConsent := &coredata.OAuth2Consent{
				ID:                  gid.New(client.ID.TenantID(), coredata.OAuth2ConsentEntityType),
				IdentityID:          req.IdentityID,
				SessionID:           req.SessionID,
				ClientID:            client.ID,
				Scopes:              requestedScopes,
				RedirectURI:         new(uri.URI(req.RedirectURI)),
				CodeChallenge:       req.CodeChallenge,
				CodeChallengeMethod: codeChallengeMethod,
				Nonce:               req.Nonce,
				State:               req.State,
				Approved:            false,
				CreatedAt:           now,
				UpdatedAt:           now,
			}

			if err := pendingConsent.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot create pending consent: %w", err)
			}

			return pg.NoRollback(
				&ConsentRequiredError{
					ConsentID: pendingConsent.ID,
					Client:    client,
					Scopes:    requestedScopes,
				},
			)
		},
	); err != nil {
		if _, ok := errors.AsType[*ConsentRequiredError](err); ok {
			return "", err
		}

		return "", err
	}

	return code, nil
}

func (s *Service) GetConsentByID(
	ctx context.Context,
	consentID gid.GID,
) (*coredata.OAuth2Consent, error) {
	var consent coredata.OAuth2Consent

	if err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := consent.LoadByID(ctx, conn, consentID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return NewError(ErrInvalidRequest, WithDescription("consent not found"))
				}

				return fmt.Errorf("cannot load consent: %w", err)
			}

			return nil
		},
	); err != nil {
		return nil, err
	}

	if consent.Approved {
		return nil, NewError(
			ErrInvalidRequest,
			WithDescription("consent already processed"),
		)
	}

	return &consent, nil
}

type ConsentApprovalResult struct {
	// Authorization code flow fields.
	Code        string
	RedirectURI string
	State       string

	// Device flow: true when the consent was for a device code grant.
	IsDeviceFlow bool

	// Denied is true when the user denied the consent request.
	Denied bool
}

func (s *Service) ApproveConsent(
	ctx context.Context,
	req *ConsentApprovalRequest,
) (*ConsentApprovalResult, error) {
	var (
		consent coredata.OAuth2Consent
		result  ConsentApprovalResult
	)

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := consent.LoadByIDForSessionForUpdate(ctx, tx, req.ConsentID, req.IdentityID, req.SessionID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrConsentNotFound
				}

				return fmt.Errorf("cannot load consent: %w", err)
			}

			if consent.Approved {
				return NewError(
					ErrInvalidRequest,
					WithDescription("consent already processed"),
				)
			}

			var client coredata.OAuth2Client
			if err := client.LoadByID(ctx, tx, coredata.NewNoScope(), consent.ClientID); err != nil {
				return fmt.Errorf("cannot load client: %w", err)
			}

			isDeviceFlow := consent.DeviceCodeID != nil
			redirectURI := string(ref.UnrefOrZero(consent.RedirectURI))

			if !isDeviceFlow && !client.IsRedirectURIAllowed(redirectURI) {
				return ErrInvalidRedirectURI
			}

			var deviceCode coredata.OAuth2DeviceCode
			if isDeviceFlow {
				if err := deviceCode.LoadByIDForUpdate(ctx, tx, *consent.DeviceCodeID); err != nil {
					return fmt.Errorf("cannot load device code: %w", err)
				}
			}

			if !req.Approved {
				if isDeviceFlow {
					deviceCode.Status = coredata.OAuth2DeviceCodeStatusDenied
					deviceCode.IdentityID = &consent.IdentityID

					if err := deviceCode.Update(ctx, tx); err != nil {
						return fmt.Errorf("cannot deny device code: %w", err)
					}
				}

				if err := consent.Delete(ctx, tx); err != nil {
					return fmt.Errorf("cannot delete consent: %w", err)
				}

				result.Denied = true
				result.IsDeviceFlow = isDeviceFlow
				result.RedirectURI = redirectURI
				result.State = consent.State

				return nil
			}

			consent.Approved = true
			consent.UpdatedAt = time.Now()

			if err := consent.Update(ctx, tx); err != nil {
				return fmt.Errorf("cannot approve consent: %w", err)
			}

			if isDeviceFlow {
				if deviceCode.Status != coredata.OAuth2DeviceCodeStatusPending {
					return ErrDeviceCodeNotPending
				}

				deviceCode.Status = coredata.OAuth2DeviceCodeStatusAuthorized
				deviceCode.IdentityID = &consent.IdentityID

				if err := deviceCode.Update(ctx, tx); err != nil {
					return fmt.Errorf("cannot update device code: %w", err)
				}

				result.IsDeviceFlow = true

				return nil
			}

			code, err := s.issueAuthorizationCode(
				ctx,
				tx,
				&client,
				consent.IdentityID,
				ref.UnrefOrZero(consent.RedirectURI),
				consent.Scopes,
				consent.CodeChallenge,
				consent.CodeChallengeMethod,
				consent.Nonce,
				req.AuthTime,
			)
			if err != nil {
				return fmt.Errorf("cannot issue authorization code: %w", err)
			}

			result.Code = code
			result.RedirectURI = redirectURI
			result.State = consent.State

			return nil
		},
	); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *Service) AuthenticateClient(
	ctx context.Context,
	clientIDRaw string,
	clientSecret string,
) (*coredata.OAuth2Client, error) {
	client, err := s.resolveClient(ctx, nil, clientIDRaw)
	if err != nil {
		return nil, err
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

func (s *Service) issueAuthorizationCode(
	ctx context.Context,
	tx pg.Tx,
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
	now := time.Now()

	code := &coredata.OAuth2AuthorizationCode{
		ID:          gid.New(client.ID.TenantID(), coredata.OAuth2AuthorizationCodeEntityType),
		HashedValue: hash.SHA256String(codeValue),
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

	if err := code.Insert(ctx, tx); err != nil {
		return "", fmt.Errorf("cannot insert authorization code: %w", err)
	}

	return codeValue, nil
}

func (s *Service) GetAccessTokenByID(ctx context.Context, accessTokenID gid.GID) (*coredata.OAuth2AccessToken, error) {
	token := &coredata.OAuth2AccessToken{}

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := token.LoadByID(ctx, conn, accessTokenID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return coredata.ErrResourceNotFound
				}

				return fmt.Errorf("cannot load oauth2 access token: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (s *Service) ListAccessTokensByIdentityID(
	ctx context.Context,
	identityID gid.GID,
	cursor *page.Cursor[coredata.OAuth2AccessTokenOrderField],
) (*page.Page[*coredata.OAuth2AccessToken, coredata.OAuth2AccessTokenOrderField], error) {
	var tokens coredata.OAuth2AccessTokens

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			if err := tokens.LoadByIdentityID(ctx, conn, identityID, cursor); err != nil {
				return fmt.Errorf("cannot load oauth2 access tokens: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return page.NewPage(tokens, cursor), nil
}

func (s *Service) CountAccessTokensByIdentityID(
	ctx context.Context,
	identityID gid.GID,
) (int, error) {
	var count int

	err := s.pg.WithConn(
		ctx,
		func(ctx context.Context, conn pg.Querier) error {
			var tokens coredata.OAuth2AccessTokens

			var err error

			count, err = tokens.CountByIdentityID(ctx, conn, identityID)
			if err != nil {
				return fmt.Errorf("cannot count oauth2 access tokens: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Service) RevokeAccessToken(ctx context.Context, accessTokenID gid.GID) error {
	return s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			token := &coredata.OAuth2AccessToken{}

			if err := token.LoadByID(ctx, tx, accessTokenID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load oauth2 access token: %w", err)
			}

			if err := token.Delete(ctx, tx); err != nil {
				return fmt.Errorf("cannot revoke oauth2 access token: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) CreateManualAccessToken(
	ctx context.Context,
	req *CreateManualAccessTokenRequest,
) (string, *coredata.OAuth2AccessToken, error) {
	if req.Name == "" {
		return "", nil, NewError(ErrInvalidRequest, WithDescription("name is required"))
	}

	now := time.Now()
	if !req.ExpiresAt.After(now) {
		return "", nil, NewError(ErrInvalidRequest, WithDescription("expires_at must be in the future"))
	}

	if len(req.Scopes) == 0 {
		return "", nil, NewError(ErrInvalidRequest, WithDescription("scopes are required"))
	}

	if err := s.registry.ValidateScopes(req.Scopes); err != nil {
		return "", nil, NewError(ErrInvalidScope, WithDescription(err.Error()))
	}

	tokenValue := rand.MustHexString(tokenByteLength)

	accessToken := &coredata.OAuth2AccessToken{
		ID:          gid.New(req.IdentityID.TenantID(), coredata.OAuth2AccessTokenEntityType),
		Name:        req.Name,
		HashedValue: hash.SHA256String(tokenValue),
		ClientID:    nil,
		IdentityID:  req.IdentityID,
		Scopes:      req.Scopes,
		CreatedAt:   now,
		ExpiresAt:   req.ExpiresAt,
	}

	err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := accessToken.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot insert oauth2 access token: %w", err)
			}

			return nil
		},
	)
	if err != nil {
		return "", nil, err
	}

	return tokenValue, accessToken, nil
}

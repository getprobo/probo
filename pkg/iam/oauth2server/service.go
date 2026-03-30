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
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/crypto/hash"
	"go.probo.inc/probo/pkg/crypto/jose"
	"go.probo.inc/probo/pkg/crypto/rand"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/uri"
)

const (
	tokenByteLength        = 32
	refreshTokenByteLength = 48
	tokenTypeBearer        = "Bearer"
)

type (
	Service struct {
		pg                   *pg.Client
		signingKeys          SigningKeys
		activeSigningIdx     []int
		rrCounter            atomic.Uint64
		baseURL              uri.URI
		logger               *log.Logger
		gc                   *GarbageCollector
		accessTokenDuration  time.Duration
		refreshTokenDuration time.Duration
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
		HashedValue: hash.SHA256([]byte(tokenValue)),
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

func (s *Service) CreateRefreshToken(
	ctx context.Context,
	clientID gid.GID,
	identityID gid.GID,
	scopes coredata.OAuth2Scopes,
	accessTokenID string,
) (string, *coredata.OAuth2RefreshToken, error) {
	var (
		tokenValue = rand.MustHexString(refreshTokenByteLength)
		now        = time.Now()
		token      = &coredata.OAuth2RefreshToken{
			ID:            fmt.Sprintf("ort_%s", tokenValue[:16]),
			HashedValue:   hash.SHA256([]byte(tokenValue)),
			ClientID:      clientID,
			IdentityID:    identityID,
			Scopes:        scopes,
			AccessTokenID: accessTokenID,
			CreatedAt:     now,
			ExpiresAt:     now.Add(s.refreshTokenDuration),
		}
	)

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := token.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot create refresh token: %w", err)
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
					return ErrInvalidClient.WithDescription("client not found")
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
		accessTokenValue     string
		refreshTokenValue    string
		idToken              string

		authTime time.Time
		codeHash = hash.SHA256Hex([]byte(codeValue))
		nonce    string
	)

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			if err := code.LoadByIDForUpdate(ctx, tx, codeHash, client.ID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrInvalidGrant.WithDescription("authorization code not found")
				}

				return fmt.Errorf("cannot load authorization code: %w", err)
			}

			if err := tx.Savepoint(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					if err := code.Delete(ctx, tx); err != nil {
						return fmt.Errorf("cannot delete authorization code: %w", err)
					}

					return nil
				},
			); err != nil {
				return err
			}

			if now.After(code.ExpiresAt) {
				return ErrInvalidGrant.WithDescription("authorization code expired")
			}

			if code.RedirectURI.String() != redirectURI {
				return ErrInvalidRedirectURI.WithDescription("redirect_uri mismatch")
			}

			if code.CodeChallenge != nil {
				if codeVerifier == "" {
					return ErrInvalidRequest.WithDescription("code_verifier required")
				}

				if !ValidateCodeChallenge(codeVerifier, *code.CodeChallenge, *code.CodeChallengeMethod) {
					return ErrInvalidRequest.WithDescription("invalid code_verifier")
				}
			}

			if code.Nonce != nil {
				nonce = *code.Nonce
			}

			accessTokenValue = rand.MustHexString(tokenByteLength)
			accessToken := &coredata.OAuth2AccessToken{
				ID:          fmt.Sprintf("oat_%s", accessTokenValue[:16]),
				HashedValue: hash.SHA256([]byte(accessTokenValue)),
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
				refreshTokenValue = rand.MustHexString(refreshTokenByteLength)
				refreshToken := &coredata.OAuth2RefreshToken{
					ID:            fmt.Sprintf("ort_%s", refreshTokenValue[:16]),
					HashedValue:   hash.SHA256([]byte(refreshTokenValue)),
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

			if code.Scopes.Contains(coredata.OAuth2ScopeOpenID) {
				idTokenClaims := NewIDTokenClaims(
					s.baseURL,
					code.IdentityID,
					client.ID,
					authTime,
					code.Scopes,
					nonce,
					accessTokenValue,
					"", false, "",
					s.accessTokenDuration,
				)

				sk := s.signingKey()
				var signErr error
				idToken, signErr = jose.SignJWT(sk.PrivateKey, sk.KID, idTokenClaims)
				if signErr != nil {
					return fmt.Errorf("cannot sign id token: %w", signErr)
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
		hashedValue          = hash.SHA256([]byte(refreshTokenValue))
		now                  = time.Now()
		scopes               coredata.OAuth2Scopes
		accessTokenExpiresAt = now.Add(s.accessTokenDuration)
		idToken              string
	)

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			previousRefreshToken := coredata.OAuth2RefreshToken{}
			if err := previousRefreshToken.LoadByHashedValueForUpdate(
				ctx,
				tx,
				hashedValue,
				client.ID,
			); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrInvalidGrant.WithDescription("refresh token not found")
				}

				return fmt.Errorf("cannot load refresh token: %w", err)
			}

			// Replay detection: if already revoked, revoke ALL tokens for this client+identity.
			if previousRefreshToken.RevokedAt != nil {
				s.logger.WarnCtx(
					ctx,
					"refresh token replay detected, revoking all tokens",
					log.String("client_id", client.ID.String()),
					log.String("identity_id", previousRefreshToken.IdentityID.String()),
				)

				if err := tx.Savepoint(
					ctx,
					func(ctx context.Context, tx pg.Tx) error {
						accessToken := coredata.OAuth2AccessToken{}
						if _, err := accessToken.DeleteByClientAndIdentity(
							ctx,
							tx,
							client.ID,
							previousRefreshToken.IdentityID,
						); err != nil {
							return fmt.Errorf("cannot delete access tokens: %w", err)
						}

						refreshToken := coredata.OAuth2RefreshToken{}
						if _, err := refreshToken.RevokeByClientAndIdentity(
							ctx,
							tx,
							client.ID,
							previousRefreshToken.IdentityID,
							now,
						); err != nil {
							return fmt.Errorf("cannot revoke refresh tokens: %w", err)
						}

						return nil
					},
				); err != nil {
					return err
				}

				return fmt.Errorf("refresh token replay detected")
			}

			if now.After(previousRefreshToken.ExpiresAt) {
				return ErrInvalidGrant.WithDescription("refresh token expired")
			}

			if err := tx.Savepoint(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					if err := previousRefreshToken.Revoke(ctx, tx, now); err != nil {
						return fmt.Errorf("cannot revoke previous refresh token: %w", err)
					}

					return nil
				},
			); err != nil {
				return err
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
				HashedValue: hash.SHA256([]byte(accessTokenValue)),
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
				HashedValue:   hash.SHA256([]byte(refreshTokenValueNew)),
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

			if previousRefreshToken.Scopes.Contains(coredata.OAuth2ScopeOpenID) {
				claims := NewIDTokenClaims(
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
				sk := s.signingKey()
				var signErr error
				idToken, signErr = jose.SignJWT(sk.PrivateKey, sk.KID, claims)
				if signErr != nil {
					return fmt.Errorf("cannot sign id token: %w", signErr)
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
		RefreshToken: refreshTokenValueNew,
		Scope:        scopes.String(),
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
		now             = time.Now()
		deviceCode      *coredata.OAuth2DeviceCode
	)

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			client := coredata.OAuth2Client{}
			if err := client.LoadByID(ctx, tx, coredata.NewNoScope(), clientID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return ErrInvalidRequest.WithDescription("unknown client_id")
				}

				return fmt.Errorf("cannot load oauth2 client: %w", err)
			}

			if !client.HasGrantType(coredata.OAuth2GrantTypeDeviceCode) {
				return ErrUnauthorizedClient.WithDescription("client not authorized for device flow")
			}

			requestedScopes := scopes.OrDefault(client.Scopes)
			if !client.AreScopesAllowed(requestedScopes) {
				return ErrInvalidScope.WithDescription("requested scope exceeds client registration")
			}

			for range 3 {
				uc, err := GenerateUserCode()
				if err != nil {
					return fmt.Errorf("cannot generate user code: %w", err)
				}

				candidate := &coredata.OAuth2DeviceCode{
					ID:             fmt.Sprintf("odc_%s", deviceCodeValue[:16]),
					DeviceCodeHash: hash.SHA256([]byte(deviceCodeValue)),
					UserCode:       uc,
					ClientID:       client.ID,
					Scopes:         requestedScopes,
					Status:         coredata.OAuth2DeviceCodeStatusPending,
					PollInterval:   5,
					CreatedAt:      now,
					ExpiresAt:      now.Add(10 * time.Minute),
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

			return ErrServerError.WithDescription("cannot generate unique user code after 3 attempts")
		},
	); err != nil {
		return "", nil, ErrServerError.Wrap(err)
	}

	return deviceCodeValue, deviceCode, nil
}

// PollDeviceCode checks the status of a device code and returns tokens if authorized.
// Returns sentinel errors (ErrSlowDown, ErrAuthorizationPending, ErrExpiredToken,
// ErrAccessDenied, ErrInvalidGrant) for the handler to map to OAuth2 responses.
func (s *Service) PollDeviceCode(
	ctx context.Context,
	clientID gid.GID,
	deviceCodeValue string,
) (*TokenResult, error) {

	var (
		hashedValue = hash.SHA256([]byte(deviceCodeValue))
		dc          = coredata.OAuth2DeviceCode{}
	)

	err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		if err := dc.LoadByDeviceCodeHashForUpdate(ctx, tx, hashedValue, clientID); err != nil {
			return err
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
		return nil, ErrInvalidGrant.WithDescription("invalid device code")
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
		return nil, ErrInvalidGrant.WithDescription("invalid device code status")
	}

	if dc.IdentityID == nil {
		return nil, ErrServerError.WithDescription("device code authorized but no identity")
	}

	var (
		identityID        = *dc.IdentityID
		now               = time.Now()
		accessTokenValue  = rand.MustHexString(tokenByteLength)
		accessToken       *coredata.OAuth2AccessToken
		refreshTokenValue string
		idToken           string
	)

	if err = s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		accessToken = &coredata.OAuth2AccessToken{
			ID:          fmt.Sprintf("oat_%s", accessTokenValue[:16]),
			HashedValue: hash.SHA256([]byte(accessTokenValue)),
			ClientID:    clientID,
			IdentityID:  identityID,
			Scopes:      dc.Scopes,
			CreatedAt:   now,
			ExpiresAt:   now.Add(s.accessTokenDuration),
		}

		if err := accessToken.Insert(ctx, tx); err != nil {
			return fmt.Errorf("cannot create access token: %w", err)
		}

		var client coredata.OAuth2Client
		if err := client.LoadByID(ctx, tx, coredata.NewNoScope(), clientID); err != nil {
			return fmt.Errorf("cannot load client: %w", err)
		}

		if client.HasGrantType(coredata.OAuth2GrantTypeRefreshToken) {
			refreshTokenValue = rand.MustHexString(refreshTokenByteLength)
			rt := &coredata.OAuth2RefreshToken{
				ID:            fmt.Sprintf("ort_%s", refreshTokenValue[:16]),
				HashedValue:   hash.SHA256([]byte(refreshTokenValue)),
				ClientID:      clientID,
				IdentityID:    identityID,
				Scopes:        dc.Scopes,
				AccessTokenID: accessToken.ID,
				CreatedAt:     now,
				ExpiresAt:     now.Add(s.refreshTokenDuration),
			}

			if err := rt.Insert(ctx, tx); err != nil {
				return fmt.Errorf("cannot create refresh token: %w", err)
			}
		}

		if dc.Scopes.Contains(coredata.OAuth2ScopeOpenID) {
			claims := NewIDTokenClaims(
				s.baseURL,
				identityID,
				clientID,
				now,
				dc.Scopes,
				"",
				accessTokenValue,
				"", false, "",
				s.accessTokenDuration,
			)
			sk := s.signingKey()
			var signErr error
			idToken, signErr = jose.SignJWT(sk.PrivateKey, sk.KID, claims)
			if signErr != nil {
				return fmt.Errorf("cannot sign id token: %w", signErr)
			}
		}

		return dc.Delete(ctx, tx)
	}); err != nil {
		return nil, ErrServerError.Wrap(err)
	}

	return &TokenResult{
		AccessToken:  accessTokenValue,
		TokenType:    tokenTypeBearer,
		ExpiresIn:    int64(time.Until(accessToken.ExpiresAt).Seconds()),
		RefreshToken: refreshTokenValue,
		Scope:        dc.Scopes.String(),
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
			dc := coredata.OAuth2DeviceCode{}

			if err := dc.LoadByUserCode(ctx, tx, userCode); err != nil {
				return fmt.Errorf("cannot find device code: %w", err)
			}

			if time.Now().After(dc.ExpiresAt) {
				return fmt.Errorf("device code expired")
			}

			if dc.Status != coredata.OAuth2DeviceCodeStatusPending {
				return fmt.Errorf("device code already %s", dc.Status)
			}

			if err := dc.UpdateStatus(
				ctx,
				tx,
				coredata.OAuth2DeviceCodeStatusAuthorized,
				&identityID,
			); err != nil {
				return fmt.Errorf("cannot authorize device code: %w", err)
			}

			return nil
		},
	)
}

func (s *Service) RegisterClient(
	ctx context.Context,
	req *RegisterClientRequest,
) (gid.GID, string, error) {
	// Check org membership.
	var membership coredata.Membership
	err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return membership.LoadActiveByIdentityIDAndOrganizationID(
			ctx,
			tx,
			req.IdentityID,
			req.OrganizationID,
		)
	})
	if err != nil {
		return gid.GID{}, "", ErrAccessDenied.WithDescription("not a member of the organization")
	}

	// Apply defaults.
	grantTypes := req.GrantTypes
	if len(grantTypes) == 0 {
		grantTypes = []coredata.OAuth2GrantType{coredata.OAuth2GrantTypeAuthorizationCode}
	}

	responseTypes := req.ResponseTypes
	if len(responseTypes) == 0 {
		responseTypes = []coredata.OAuth2ResponseType{coredata.OAuth2ResponseTypeCode}
	}

	authMethod := req.TokenEndpointAuthMethod
	if authMethod == "" {
		authMethod = coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretBasic
	}

	visibility := req.Visibility
	if visibility == "" {
		visibility = coredata.OAuth2ClientVisibilityPrivate
	}

	for _, u := range req.RedirectURIs {
		parsed, _ := url.Parse(string(u))

		switch visibility {
		case coredata.OAuth2ClientVisibilityPublic:
			if parsed.Scheme != "https" {
				return gid.GID{}, "", ErrInvalidRequest.WithDescription(
					"public clients require https redirect_uris",
				)
			}
		case coredata.OAuth2ClientVisibilityPrivate:
			if parsed.Scheme == "http" {
				host := parsed.Hostname()
				if host != "localhost" && host != "127.0.0.1" && host != "::1" {
					return gid.GID{}, "", ErrInvalidRequest.WithDescription(
						"http redirect_uris are only allowed for localhost",
					)
				}
			} else if parsed.Scheme != "https" {
				return gid.GID{}, "", ErrInvalidRequest.WithDescription(
					fmt.Sprintf("unsupported redirect_uri scheme: %s", parsed.Scheme),
				)
			}
		}
	}

	scopes := req.Scopes
	if len(scopes) == 0 {
		scopes = coredata.OAuth2Scopes{
			coredata.OAuth2ScopeOpenID,
			coredata.OAuth2ScopeProfile,
			coredata.OAuth2ScopeEmail,
		}
	}

	tenantID := req.OrganizationID.TenantID()
	clientID := gid.New(tenantID, coredata.OAuth2ClientEntityType)

	var (
		secretHash      []byte
		plaintextSecret string
	)

	if authMethod != coredata.OAuth2ClientTokenEndpointAuthMethodNone {
		plaintextSecret = rand.MustHexString(tokenByteLength)
		secretHash = hash.SHA256([]byte(plaintextSecret))
	}

	now := time.Now()
	client := &coredata.OAuth2Client{
		ID:                      clientID,
		OrganizationID:          req.OrganizationID,
		ClientSecretHash:        secretHash,
		ClientName:              req.ClientName,
		Visibility:              visibility,
		RedirectURIs:            req.RedirectURIs,
		Scopes:                  scopes,
		GrantTypes:              grantTypes,
		ResponseTypes:           responseTypes,
		TokenEndpointAuthMethod: authMethod,
		LogoURI:                 req.LogoURI,
		ClientURI:               req.ClientURI,
		Contacts:                req.Contacts,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	scope := coredata.NewScopeFromObjectID(req.OrganizationID)

	err = s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return client.Insert(ctx, tx, scope)
	})
	if err != nil {
		return gid.GID{}, "", fmt.Errorf("cannot insert oauth2 client: %w", err)
	}

	return clientID, plaintextSecret, nil
}

func (s *Service) LoadAccessToken(ctx context.Context, tokenValue string) (*coredata.OAuth2AccessToken, error) {
	var (
		hashedValue = hash.SHA256([]byte(tokenValue))
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
		hashedValue = hash.SHA256([]byte(tokenValue))
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
	hashedValue := hash.SHA256([]byte(tokenValue))

	if err := s.pg.WithTx(
		ctx,
		func(ctx context.Context, tx pg.Tx) error {
			accessToken := coredata.OAuth2AccessToken{}
			if err := accessToken.LoadByHashedValueAndClientID(ctx, tx, hashedValue, clientID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load access token: %w", err)
			}

			if err := tx.Savepoint(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					if err := accessToken.Delete(ctx, tx); err != nil {
						return fmt.Errorf("cannot delete access token: %w", err)
					}

					return nil
				},
			); err != nil {
				return err
			}

			refreshToken := coredata.OAuth2RefreshToken{}
			if err := refreshToken.LoadByHashedValueAndClientID(ctx, tx, hashedValue, clientID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return nil
				}

				return fmt.Errorf("cannot load refresh token: %w", err)
			}

			if err := tx.Savepoint(
				ctx,
				func(ctx context.Context, tx pg.Tx) error {
					if err := refreshToken.Revoke(ctx, tx, time.Now()); err != nil {
						return fmt.Errorf("cannot revoke refresh token: %w", err)
					}

					return nil
				},
			); err != nil {
				return err
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
		return "", ErrInvalidClient.WithDescription("cannot load client")
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
			return "", ErrUnauthorizedClient.WithDescription("client is private and user is not a member of the organization")
		}
	}

	if req.ResponseType != coredata.OAuth2ResponseTypeCode {
		return "", ErrInvalidRequest.WithDescription("unsupported response_type")
	}

	requestedScopes := req.Scopes.OrDefault(client.Scopes)
	if !client.AreScopesAllowed(requestedScopes) {
		return "", ErrInvalidScope.WithDescription("requested scope exceeds client registration")
	}

	codeChallengeMethod := req.CodeChallengeMethod
	if client.TokenEndpointAuthMethod == coredata.OAuth2ClientTokenEndpointAuthMethodNone && req.CodeChallenge == "" {
		return "", ErrInvalidRequest.WithDescription("code_challenge required for public clients")
	}

	if codeChallengeMethod != "" && codeChallengeMethod != coredata.OAuth2CodeChallengeMethodS256 {
		return "", ErrInvalidRequest.WithDescription("only S256 code_challenge_method is supported")
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
			return "", ErrServerError.Wrap(err)
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
		return "", ErrServerError.WithDescription("cannot create pending consent")
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
		return "", "", "", ErrInvalidRequest.WithDescription("consent not found")
	}

	if consent.IdentityID != req.IdentityID {
		return "", "", "", ErrAccessDenied.WithDescription("consent does not belong to this identity")
	}

	if consent.Approved {
		return "", "", "", ErrInvalidRequest.WithDescription("consent already processed")
	}

	client, err := s.GetClientByID(ctx, consent.ClientID)
	if err != nil {
		return "", "", "", ErrInvalidClient.WithDescription("cannot load client")
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
		return "", redirectURI, consent.State, ErrAccessDenied.WithDescription("user denied the request")
	}

	// Mark consent as approved.
	consent.Approved = true
	consent.UpdatedAt = time.Now()
	if err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return consent.Update(ctx, tx)
	}); err != nil {
		return "", redirectURI, consent.State, ErrServerError.WithDescription("cannot approve consent")
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
		return nil, ErrInvalidClient.WithDescription("cannot load client")
	}

	if client.TokenEndpointAuthMethod == coredata.OAuth2ClientTokenEndpointAuthMethodNone {
		return client, nil
	}

	if clientSecret == "" {
		return nil, ErrInvalidClient.WithDescription("missing client_secret")
	}

	if subtle.ConstantTimeCompare(client.ClientSecretHash, hash.SHA256([]byte(clientSecret))) != 1 {
		return nil, ErrInvalidClient.WithDescription("invalid client_secret")
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

	codeHash := hash.SHA256Hex([]byte(codeValue))
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

	if err := s.pg.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
		return code.Insert(ctx, tx)
	}); err != nil {
		return "", fmt.Errorf("cannot save authorization code: %w", err)
	}

	return codeValue, nil
}

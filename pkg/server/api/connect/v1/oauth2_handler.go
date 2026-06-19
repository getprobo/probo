// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package connect_v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/oauth2"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/connect/v1/types"
	"go.probo.inc/probo/pkg/uri"
)

type OAuth2Handler struct {
	iam           *iam.Service
	sessionCookie *authn.Cookie
	baseURL       *baseurl.BaseURL
	logger        *log.Logger
}

func NewOAuth2Handler(
	svc *iam.Service,
	cookieConfig securecookie.Config,
	baseURL *baseurl.BaseURL,
	logger *log.Logger,
) *OAuth2Handler {
	return &OAuth2Handler{
		iam:           svc,
		sessionCookie: authn.NewCookie(&cookieConfig),
		baseURL:       baseURL,
		logger:        logger.Named("oauth2"),
	}
}

// ClientAuthMiddleware authenticates the OAuth2 client from HTTP Basic auth
// or POST body credentials and stores it in the request context.
func (h *OAuth2Handler) ClientAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client, err := h.authenticateClient(r)
		if err != nil {
			h.renderOAuth2ErrorResponse(w, r, oauth2.ErrInvalidClient)
			return
		}

		ctx := oauth2.ContextWithClient(r.Context(), client)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// BearerTokenMiddleware validates the OAuth2 bearer token from the
// Authorization header and stores the access token in the request context.
func (h *OAuth2Handler) BearerTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenValue, err := bearertoken.Parse(r.Header.Get("Authorization"))
		if err != nil {
			bearertoken.SetBearerInvalidToken(w, h.baseURL)
			http.Error(w, "unauthorized", http.StatusUnauthorized)

			return
		}

		accessToken, err := h.iam.OAuth2ServerService.LoadAccessToken(r.Context(), tokenValue)
		if err != nil {
			bearertoken.SetBearerInvalidToken(w, h.baseURL)
			http.Error(w, "unauthorized", http.StatusUnauthorized)

			return
		}

		ctx := oauth2.ContextWithAccessToken(r.Context(), accessToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *OAuth2Handler) endpoints() oauth2.Endpoints {
	api := h.baseURL.String() + "/api/connect/v1"

	return oauth2.Endpoints{
		Authorization:       uri.URI(api + "/oauth2/authorize"),
		Token:               uri.URI(api + "/oauth2/token"),
		Userinfo:            uri.URI(api + "/oauth2/userinfo"),
		JWKS:                uri.URI(api + "/oauth2/jwks"),
		Registration:        uri.URI(api + "/oauth2/register"),
		Introspection:       uri.URI(api + "/oauth2/introspect"),
		Revocation:          uri.URI(api + "/oauth2/revoke"),
		DeviceAuthorization: uri.URI(api + "/oauth2/device"),
	}
}

// --- Handlers ---

// DiscoveryHandler serves the OpenID Connect Discovery document.
// GET /.well-known/openid-configuration
func (h *OAuth2Handler) DiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	metadata := h.iam.OAuth2ServerMetadata(h.endpoints())

	PublicCache(w, 1*time.Hour)
	httpserver.RenderJSON(w, http.StatusOK, metadata)
}

// JWKSHandler serves the JSON Web Key Set.
// GET /oauth2/jwks
func (h *OAuth2Handler) JWKSHandler(w http.ResponseWriter, r *http.Request) {
	jwks := h.iam.OAuth2ServerService.JWKS()

	PublicCache(w, 1*time.Hour)
	httpserver.RenderJSON(w, http.StatusOK, jwks)
}

// AuthorizeHandler handles the authorization endpoint.
// GET /oauth2/authorize
func (h *OAuth2Handler) AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	identity := authn.IdentityFromContext(r.Context())
	if identity == nil {
		continueURL := h.baseURL.WithPath("/api/connect/v1/oauth2/authorize").
			WithQueryValues(r.URL.Query()).
			MustString()
		loginURL := h.baseURL.WithPath("/auth/login").
			WithQuery("continue", continueURL).
			MustString()
		http.Redirect(w, r, loginURL, http.StatusFound)

		return
	}

	var in types.OAuth2AuthorizeInput
	if err := in.DecodeQuery(r.URL.Query()); err != nil {
		h.handleAuthorizeError(w, r, oauth2.NewError(oauth2.ErrInvalidRequest, oauth2.WithError(err)), "", "")
		return
	}

	session := authn.SessionFromContext(r.Context())

	authTime := time.Now()
	if session != nil {
		authTime = session.CreatedAt
	}

	code, err := h.iam.OAuth2ServerService.Authorize(
		r.Context(),
		&oauth2.AuthorizeRequest{
			IdentityID:          identity.ID,
			SessionID:           session.ID,
			ResponseType:        in.ResponseType,
			ClientIDRaw:         in.ClientIDRaw,
			RedirectURI:         in.RedirectURI,
			Scopes:              in.Scopes,
			CodeChallenge:       in.CodeChallenge,
			CodeChallengeMethod: in.CodeChallengeMethod,
			Nonce:               in.Nonce,
			State:               in.State,
			AuthTime:            authTime,
		},
	)

	if consentErr, ok := errors.AsType[*oauth2.ConsentRequiredError](err); ok {
		consentURL := h.baseURL.WithPath("/auth/consent").
			WithQuery("consent_id", consentErr.ConsentID.String()).
			MustString()
		http.Redirect(w, r, consentURL, http.StatusFound)

		return
	}

	if err != nil {
		oauthErr := toOAuth2Error(err)
		h.handleAuthorizeError(w, r, oauthErr, in.RedirectURI, in.State)

		return
	}

	redirectWithCode(w, r, in.RedirectURI, code, in.State)
}

func (h *OAuth2Handler) TokenHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderOAuth2ErrorResponse(w, r, oauth2.NewError(oauth2.ErrInvalidRequest, oauth2.WithDescription("invalid form data")))
		return
	}

	var (
		grantType coredata.OAuth2GrantType
		value     = r.FormValue("grant_type")
	)

	if err := grantType.UnmarshalText([]byte(value)); err != nil {
		h.renderOAuth2ErrorResponse(w, r, oauth2.ErrUnsupportedGrantType)
		return
	}

	switch grantType {
	case coredata.OAuth2GrantTypeAuthorizationCode:
		h.handleAuthorizationCodeGrant(w, r)
	case coredata.OAuth2GrantTypeRefreshToken:
		h.handleRefreshTokenGrant(w, r)
	case coredata.OAuth2GrantTypeDeviceCode:
		h.handleDeviceCodeGrant(w, r)
	default:
		panic(fmt.Sprintf("unsupported grant type: %s", grantType))
	}
}

func (h *OAuth2Handler) IntrospectHandler(w http.ResponseWriter, r *http.Request) {
	client, ok := oauth2.ClientFromContext(r.Context())
	if !ok {
		h.renderOAuth2ErrorResponse(w, r, oauth2.ErrInvalidClient)
		return
	}

	in := types.OAuth2IntrospectInput{}
	if err := in.DecodeForm(r); err != nil {
		h.renderOAuth2ErrorResponse(w, r, oauth2.NewError(oauth2.ErrInvalidRequest, oauth2.WithError(err)))
		return
	}

	result, err := h.iam.OAuth2ServerService.IntrospectToken(
		r.Context(),
		client.ID,
		in.Token,
		in.TokenTypeHint,
	)
	if err != nil || result == nil {
		httpserver.RenderJSON(w, http.StatusOK, types.InactiveIntrospectResponse())
		return
	}

	httpserver.RenderJSON(w, http.StatusOK, types.ActiveIntrospectResponse(result))
}

func (h *OAuth2Handler) RevokeHandler(w http.ResponseWriter, r *http.Request) {
	var (
		client, _ = oauth2.ClientFromContext(r.Context())
		in        = types.OAuth2RevokeInput{}
	)

	if err := in.DecodeForm(r); err != nil {
		h.renderOAuth2ErrorResponse(w, r, oauth2.NewError(oauth2.ErrInvalidRequest, oauth2.WithError(err)))
		return
	}

	if err := h.iam.OAuth2ServerService.RevokeToken(
		r.Context(),
		client.ID,
		in.Token,
		in.TokenTypeHint,
	); err != nil {
		h.logger.ErrorCtx(r.Context(), "cannot revoke token", log.Error(err))
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusServiceUnavailable)

		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeviceAuthHandler handles the device authorization endpoint (RFC 8628).
// POST /oauth2/device
func (h *OAuth2Handler) DeviceAuthHandler(w http.ResponseWriter, r *http.Request) {
	in := types.OAuth2DeviceAuthInput{}
	if err := in.DecodeForm(r); err != nil {
		h.renderOAuth2ErrorResponse(w, r, oauth2.NewError(oauth2.ErrInvalidRequest, oauth2.WithError(err)))
		return
	}

	deviceCodeValue, dc, err := h.iam.OAuth2ServerService.CreateDeviceCode(
		r.Context(),
		in.ClientID,
		in.Scopes,
	)
	if err != nil {
		h.renderOAuth2ErrorResponse(w, r, err)
		return
	}

	verificationURI := uri.URI(h.baseURL.WithPath("/auth/device").MustString())
	verificationURIComplete := uri.URI(
		h.baseURL.WithPath("/auth/device").
			WithQuery("user_code", string(dc.UserCode)).
			MustString(),
	)

	httpserver.RenderJSON(
		w,
		http.StatusOK,
		&types.OAuth2DeviceAuthResponse{
			DeviceCode:              deviceCodeValue,
			UserCode:                dc.UserCode.Format(),
			VerificationURI:         verificationURI,
			VerificationURIComplete: verificationURIComplete,
			ExpiresIn:               int(time.Until(dc.ExpiresAt).Seconds()),
			Interval:                dc.PollInterval,
		},
	)
}

// RegisterHandler handles dynamic client registration (RFC 7591).
// POST /oauth2/register
func (h *OAuth2Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	identity := authn.IdentityFromContext(r.Context())

	var in types.OAuth2RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		h.renderOAuth2ErrorResponse(
			w,
			r,
			oauth2.NewError(oauth2.ErrInvalidRequest, oauth2.WithDescription("invalid JSON body")),
		)

		return
	}

	if len(in.GrantTypes) == 0 {
		in.GrantTypes = []coredata.OAuth2GrantType{coredata.OAuth2GrantTypeAuthorizationCode}
	}

	if len(in.ResponseTypes) == 0 {
		in.ResponseTypes = []coredata.OAuth2ResponseType{coredata.OAuth2ResponseTypeCode}
	}

	if in.TokenEndpointAuthMethod == "" {
		in.TokenEndpointAuthMethod = coredata.OAuth2ClientTokenEndpointAuthMethodClientSecretBasic
	}

	if in.Visibility == "" {
		in.Visibility = coredata.OAuth2ClientVisibilityPrivate
	}

	if len(in.Scopes) == 0 {
		in.Scopes = coredata.OAuth2Scopes{
			oauth2.ScopeOpenID,
			oauth2.ScopeProfile,
			oauth2.ScopeEmail,
		}
	}

	clientID, clientSecret, err := h.iam.OAuth2ServerService.RegisterClient(
		r.Context(),
		&oauth2.RegisterClientRequest{
			IdentityID:              identity.ID,
			OrganizationID:          in.OrganizationID,
			ClientName:              in.ClientName,
			Visibility:              in.Visibility,
			RedirectURIs:            in.RedirectURIs,
			GrantTypes:              in.GrantTypes,
			ResponseTypes:           in.ResponseTypes,
			TokenEndpointAuthMethod: in.TokenEndpointAuthMethod,
			LogoURI:                 in.LogoURI,
			ClientURI:               in.ClientURI,
			Contacts:                in.Contacts,
			Scopes:                  in.Scopes,
		},
	)
	if err != nil {
		h.renderOAuth2ErrorResponse(w, r, err)
		return
	}

	httpserver.RenderJSON(
		w,
		http.StatusCreated,
		&types.OAuth2RegisterResponse{
			ClientID:                clientID.String(),
			ClientSecret:            clientSecret,
			ClientName:              in.ClientName,
			Visibility:              in.Visibility,
			RedirectURIs:            in.RedirectURIs,
			GrantTypes:              in.GrantTypes,
			ResponseTypes:           in.ResponseTypes,
			TokenEndpointAuthMethod: in.TokenEndpointAuthMethod,
			Scopes:                  in.Scopes,
		},
	)
}

// UserInfoHandler serves the OIDC UserInfo endpoint.
// GET /oauth2/userinfo
func (h *OAuth2Handler) UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	accessToken, ok := oauth2.AccessTokenFromContext(r.Context())
	if !ok {
		bearertoken.SetBearerInvalidToken(w, h.baseURL)
		http.Error(w, "unauthorized", http.StatusUnauthorized)

		return
	}

	claims, err := h.iam.OAuth2ServerService.UserInfo(
		r.Context(),
		accessToken.IdentityID,
		accessToken.Scopes,
	)
	if err != nil {
		h.renderOAuth2ErrorResponse(w, r, oauth2.ErrServerError)
		return
	}

	httpserver.RenderJSON(w, http.StatusOK, claims)
}

// --- Internal helpers ---

func (h *OAuth2Handler) authenticateClient(r *http.Request) (*coredata.OAuth2Client, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	var clientIDStr, clientSecret string

	// Try HTTP Basic auth first.
	if username, password, ok := r.BasicAuth(); ok {
		clientIDStr = username
		clientSecret = password
	} else {
		// Fall back to POST body.
		clientIDStr = r.FormValue("client_id")
		clientSecret = r.FormValue("client_secret")
	}

	if clientIDStr == "" {
		return nil, oauth2.ErrInvalidClient
	}

	return h.iam.OAuth2ServerService.AuthenticateClient(r.Context(), clientIDStr, clientSecret)
}

func (h *OAuth2Handler) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	client, err := h.authenticateClient(r)
	if err != nil {
		h.renderOAuth2ErrorResponse(w, r, oauth2.ErrInvalidClient)
		return
	}

	var in types.OAuth2AuthorizationCodeGrantInput
	if err := in.DecodeForm(r); err != nil {
		h.renderOAuth2ErrorResponse(w, r, oauth2.NewError(oauth2.ErrInvalidGrant, oauth2.WithError(err)))
		return
	}

	result, err := h.iam.OAuth2ServerService.ExchangeAuthorizationCode(
		r.Context(),
		client,
		in.Code,
		in.RedirectURI,
		in.CodeVerifier,
	)
	if err != nil {
		h.renderOAuth2ErrorResponse(w, r, oauth2.NewError(oauth2.ErrInvalidGrant, oauth2.WithDescription("invalid or expired code")))
		return
	}

	NoCache(w)
	httpserver.RenderJSON(w, http.StatusOK, tokenResultToResponse(result))
}

func (h *OAuth2Handler) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	client, err := h.authenticateClient(r)
	if err != nil {
		h.renderOAuth2ErrorResponse(w, r, oauth2.ErrInvalidClient)
		return
	}

	var in types.OAuth2RefreshTokenGrantInput
	if err := in.DecodeForm(r); err != nil {
		h.renderOAuth2ErrorResponse(w, r, oauth2.NewError(oauth2.ErrInvalidGrant, oauth2.WithError(err)))
		return
	}

	result, err := h.iam.OAuth2ServerService.RefreshToken(r.Context(), client, in.RefreshToken)
	if err != nil {
		h.renderOAuth2ErrorResponse(w, r, oauth2.NewError(oauth2.ErrInvalidGrant, oauth2.WithDescription("invalid or expired refresh token")))
		return
	}

	NoCache(w)
	httpserver.RenderJSON(w, http.StatusOK, tokenResultToResponse(result))
}

func (h *OAuth2Handler) handleDeviceCodeGrant(w http.ResponseWriter, r *http.Request) {
	var in types.OAuth2DeviceCodeGrantInput
	if err := in.DecodeForm(r); err != nil {
		h.renderOAuth2ErrorResponse(w, r, oauth2.NewError(oauth2.ErrInvalidRequest, oauth2.WithError(err)))
		return
	}

	result, err := h.iam.OAuth2ServerService.PollDeviceCode(
		r.Context(),
		in.ClientID,
		in.DeviceCode,
	)
	if err != nil {
		h.renderOAuth2ErrorResponse(w, r, err)
		return
	}

	NoCache(w)
	httpserver.RenderJSON(w, http.StatusOK, tokenResultToResponse(result))
}

func tokenResultToResponse(r *oauth2.TokenResult) *types.OAuth2TokenResponse {
	return &types.OAuth2TokenResponse{
		AccessToken:  r.AccessToken,
		TokenType:    r.TokenType,
		ExpiresIn:    r.ExpiresIn,
		RefreshToken: r.RefreshToken,
		IDToken:      r.IDToken,
		Scope:        r.Scope,
	}
}

func redirectWithCode(w http.ResponseWriter, r *http.Request, redirectURI, code, state string) {
	u, _ := url.Parse(redirectURI)
	q := u.Query()
	q.Set("code", code)

	if state != "" {
		q.Set("state", state)
	}

	u.RawQuery = q.Encode()

	http.Redirect(w, r, u.String(), http.StatusFound)
}

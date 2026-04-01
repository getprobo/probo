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

package connect_v1

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/oauth2server"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
)

//go:embed templates/consent.html.tmpl
var consentTmplHTML string

//go:embed templates/device_verify.html.tmpl
var deviceVerifyTmplHTML string

//go:embed templates/device_verify_failed.html.tmpl
var deviceVerifyFailedTmplHTML string

//go:embed templates/device_verify_success.html.tmpl
var deviceVerifySuccessTmplHTML string

var (
	consentTmpl             = template.Must(template.New("consent").Parse(consentTmplHTML))
	deviceVerifyTmpl        = template.Must(template.New("device_verify").Parse(deviceVerifyTmplHTML))
	deviceVerifyFailedTmpl  = template.Must(template.New("device_verify_failed").Parse(deviceVerifyFailedTmplHTML))
	deviceVerifySuccessTmpl = template.Must(template.New("device_verify_success").Parse(deviceVerifySuccessTmplHTML))

	oauth2ClientContextKey      = &ctxKey{name: "oauth2_client"}
	oauth2AccessTokenContextKey = &ctxKey{name: "oauth2_access_token"}
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

// oauth2ClientFromContext returns the authenticated OAuth2 client from context.
func oauth2ClientFromContext(r *http.Request) *coredata.OAuth2Client {
	client, _ := r.Context().Value(oauth2ClientContextKey).(*coredata.OAuth2Client)
	return client
}

// oauth2AccessTokenFromContext returns the validated OAuth2 access token from context.
func oauth2AccessTokenFromContext(r *http.Request) *coredata.OAuth2AccessToken {
	token, _ := r.Context().Value(oauth2AccessTokenContextKey).(*coredata.OAuth2AccessToken)
	return token
}

// ClientAuthMiddleware authenticates the OAuth2 client from HTTP Basic auth
// or POST body credentials and stores it in the request context.
func (h *OAuth2Handler) ClientAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client, err := h.authenticateClient(r)
		if err != nil {
			h.writeOAuth2Error(w, r, oauth2server.ErrInvalidClient)
			return
		}

		ctx := context.WithValue(r.Context(), oauth2ClientContextKey, client)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// BearerTokenMiddleware validates the OAuth2 bearer token from the
// Authorization header and stores the access token in the request context.
func (h *OAuth2Handler) BearerTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenValue, err := bearertoken.Parse(r.Header.Get("Authorization"))
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		accessToken, err := h.iam.OAuth2ServerService.LoadAccessToken(r.Context(), tokenValue)
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), oauth2AccessTokenContextKey, accessToken)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *OAuth2Handler) endpoints() oauth2server.Endpoints {
	api := h.baseURL.String() + "/api/connect/v1"

	return oauth2server.Endpoints{
		Authorization:       api + "/oauth2/authorize",
		Token:               api + "/oauth2/token",
		Userinfo:            api + "/oauth2/userinfo",
		JWKS:                api + "/oauth2/jwks",
		Registration:        api + "/oauth2/register",
		Introspection:       api + "/oauth2/introspect",
		Revocation:          api + "/oauth2/revoke",
		DeviceAuthorization: api + "/oauth2/device",
	}
}

// --- Handlers ---

// DiscoveryHandler serves the OpenID Connect Discovery document.
// GET /.well-known/openid-configuration
func (h *OAuth2Handler) DiscoveryHandler(w http.ResponseWriter, r *http.Request) {
	metadata := h.iam.OAuth2ServerService.Metadata(h.endpoints())

	w.Header().Set("Cache-Control", "public, max-age=3600")
	httpserver.RenderJSON(w, http.StatusOK, metadata)
}

// JWKSHandler serves the JSON Web Key Set.
// GET /oauth2/jwks
func (h *OAuth2Handler) JWKSHandler(w http.ResponseWriter, r *http.Request) {
	jwks := h.iam.OAuth2ServerService.JWKS()

	w.Header().Set("Cache-Control", "public, max-age=3600")
	httpserver.RenderJSON(w, http.StatusOK, jwks)
}

// AuthorizeHandler handles the authorization endpoint.
// GET /oauth2/authorize
func (h *OAuth2Handler) AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	identity := authn.IdentityFromContext(r.Context())
	if identity == nil {
		// Prevent infinite redirect loop: if we already redirected to login
		// and came back without a valid session, return an error.
		if r.URL.Query().Get("_login_redirect") == "1" {
			h.writeOAuth2Error(w, r, oauth2server.ErrInvalidRequest.WithDescription("authentication required"))
			return
		}

		continueURL := h.baseURL.WithPath("/api/connect/v1/oauth2/authorize").
			WithQueryValues(r.URL.Query()).
			WithQuery("_login_redirect", "1").
			MustString()
		loginURL := h.baseURL.WithPath("/login").
			WithQuery("continue", continueURL).
			MustString()
		http.Redirect(w, r, loginURL, http.StatusFound)
		return
	}

	q := r.URL.Query()
	redirectURI := q.Get("redirect_uri")
	state := q.Get("state")

	clientIDStr := q.Get("client_id")
	if clientIDStr == "" {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidRequest.WithDescription("missing client_id"))
		return
	}

	clientID, err := gid.ParseGID(clientIDStr)
	if err != nil {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidRequest.WithDescription("invalid client_id"))
		return
	}

	session := authn.SessionFromContext(r.Context())
	authTime := time.Now()
	if session != nil {
		authTime = session.CreatedAt
	}

	scopes, err := parseScopes(q.Get("scope"))
	if err != nil {
		h.handleAuthorizeError(w, r, err, redirectURI, state)
		return
	}

	code, err := h.iam.OAuth2ServerService.Authorize(
		r.Context(),
		&oauth2server.AuthorizeRequest{
			IdentityID:          identity.ID,
			SessionID:           session.ID,
			ResponseType:        q.Get("response_type"),
			ClientID:            clientID,
			RedirectURI:         redirectURI,
			Scopes:              scopes,
			CodeChallenge:       q.Get("code_challenge"),
			CodeChallengeMethod: coredata.OAuth2CodeChallengeMethod(q.Get("code_challenge_method")),
			Nonce:               q.Get("nonce"),
			State:               state,
			AuthTime:            authTime,
		},
	)

	if consentErr, ok := errors.AsType[*oauth2server.ConsentRequiredError](err); ok {
		h.renderConsentPage(
			w,
			consentErr.ConsentID,
			consentErr.Client,
			consentErr.Scopes,
		)
		return
	}

	if err != nil {
		h.handleAuthorizeError(w, r, err, redirectURI, state)
		return
	}

	redirectWithCode(w, r, redirectURI, code, state)
}

// AuthorizeConsentHandler handles consent form submission.
// POST /oauth2/authorize
func (h *OAuth2Handler) AuthorizeConsentHandler(w http.ResponseWriter, r *http.Request) {
	identity := authn.IdentityFromContext(r.Context())

	if err := r.ParseForm(); err != nil {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidRequest.WithDescription("invalid form data"))
		return
	}

	consentID, err := gid.ParseGID(r.FormValue("consent_id"))
	if err != nil {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidRequest.WithDescription("invalid consent_id"))
		return
	}

	session := authn.SessionFromContext(r.Context())
	authTime := time.Now()
	if session != nil {
		authTime = session.CreatedAt
	}

	code, redirectURI, state, err := h.iam.OAuth2ServerService.ApproveConsent(
		r.Context(),
		&oauth2server.ConsentApprovalRequest{
			ConsentID:  consentID,
			IdentityID: identity.ID,
			Approved:   r.FormValue("action") != "deny",
			AuthTime:   authTime,
		},
	)
	if err != nil {
		h.handleAuthorizeError(w, r, err, redirectURI, state)
		return
	}

	redirectWithCode(w, r, redirectURI, code, state)
}

// TokenHandler handles the token endpoint.
// POST /oauth2/token
func (h *OAuth2Handler) TokenHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidRequest.WithDescription("invalid form data"))
		return
	}

	grantType := r.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		h.handleAuthorizationCodeGrant(w, r)
	case "refresh_token":
		h.handleRefreshTokenGrant(w, r)
	case "urn:ietf:params:oauth:grant-type:device_code":
		h.handleDeviceCodeGrant(w, r)
	default:
		h.writeOAuth2Error(w, r, oauth2server.ErrUnsupportedGrantType)
	}
}

// IntrospectHandler handles token introspection (RFC 7662).
// POST /oauth2/introspect
func (h *OAuth2Handler) IntrospectHandler(w http.ResponseWriter, r *http.Request) {
	client := oauth2ClientFromContext(r)

	if err := r.ParseForm(); err != nil {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidRequest.WithDescription("invalid form data"))
		return
	}

	token := r.FormValue("token")
	if token == "" {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidRequest.WithDescription("missing token parameter"))
		return
	}

	result, err := h.iam.OAuth2ServerService.IntrospectToken(r.Context(), client.ID, token)
	if err != nil {
		httpserver.RenderJSON(w, http.StatusOK, map[string]any{"active": false})
		return
	}

	httpserver.RenderJSON(
		w,
		http.StatusOK,
		map[string]any{
			"active":     true,
			"scope":      result.Scopes,
			"client_id":  result.ClientID,
			"sub":        result.IdentityID,
			"exp":        result.ExpiresAt.Unix(),
			"iat":        result.CreatedAt.Unix(),
			"token_type": "Bearer",
		},
	)
}

// RevokeHandler handles token revocation (RFC 7009).
// POST /oauth2/revoke
func (h *OAuth2Handler) RevokeHandler(w http.ResponseWriter, r *http.Request) {
	client := oauth2ClientFromContext(r)

	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	token := r.FormValue("token")
	if token == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	h.iam.OAuth2ServerService.RevokeToken(r.Context(), client.ID, token)

	// Always return 200 per RFC 7009.
	w.WriteHeader(http.StatusOK)
}

// DeviceAuthHandler handles the device authorization endpoint (RFC 8628).
// POST /oauth2/device
func (h *OAuth2Handler) DeviceAuthHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidRequest.WithDescription("invalid form data"))
		return
	}

	clientID, err := gid.ParseGID(r.FormValue("client_id"))
	if err != nil {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidRequest.WithDescription("invalid client_id"))
		return
	}

	var scopes coredata.OAuth2Scopes
	if scopeStr := r.FormValue("scope"); scopeStr != "" {
		if err := scopes.UnmarshalText([]byte(scopeStr)); err != nil {
			h.writeOAuth2Error(w, r, oauth2server.ErrInvalidRequest.WithDescription("invalid scope"))
			return
		}
	}

	result, err := h.iam.OAuth2ServerService.CreateDeviceCode(
		r.Context(),
		clientID,
		scopes,
	)
	if err != nil {
		h.writeOAuth2Error(w, r, err)
		return
	}

	verificationURI := h.baseURL.WithPath("/api/connect/v1/oauth2/device/verify").MustString()
	verificationURIComplete := h.baseURL.WithPath("/api/connect/v1/oauth2/device/verify").
		WithQuery("user_code", string(result.UserCode)).
		MustString()

	httpserver.RenderJSON(w, http.StatusOK, map[string]any{
		"device_code":               result.DeviceCode,
		"user_code":                 result.UserCode.Format(),
		"verification_uri":          verificationURI,
		"verification_uri_complete": verificationURIComplete,
		"expires_in":                result.ExpiresIn,
		"interval":                  result.Interval,
	})
}

// RegisterHandler handles dynamic client registration (RFC 7591).
// POST /oauth2/register
func (h *OAuth2Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	identity := authn.IdentityFromContext(r.Context())

	var req oauth2server.RegisterClientRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidRequest.WithDescription("invalid JSON body"))
		return
	}

	req.IdentityID = identity.ID

	clientID, clientSecret, err := h.iam.OAuth2ServerService.RegisterClient(r.Context(), &req)
	if err != nil {
		h.writeOAuth2Error(w, r, err)
		return
	}

	resp := map[string]any{
		"client_id":                  clientID.String(),
		"client_name":                req.ClientName,
		"visibility":                 req.Visibility,
		"redirect_uris":              req.RedirectURIs,
		"grant_types":                req.GrantTypes,
		"response_types":             req.ResponseTypes,
		"token_endpoint_auth_method": req.TokenEndpointAuthMethod,
		"scopes":                     req.Scopes,
	}

	if clientSecret != "" {
		resp["client_secret"] = clientSecret
	}

	httpserver.RenderJSON(w, http.StatusCreated, resp)
}

// UserInfoHandler serves the OIDC UserInfo endpoint.
// GET /oauth2/userinfo
func (h *OAuth2Handler) UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	accessToken := oauth2AccessTokenFromContext(r)

	claims, err := h.iam.OAuth2ServerService.UserInfo(
		r.Context(),
		accessToken.IdentityID,
		accessToken.Scopes,
	)
	if err != nil {
		h.writeOAuth2Error(w, r, oauth2server.ErrServerError)
		return
	}

	httpserver.RenderJSON(w, http.StatusOK, claims)
}

// DeviceVerifyPage renders the device code verification page.
// GET /oauth2/device/verify
func (h *OAuth2Handler) DeviceVerifyPage(w http.ResponseWriter, r *http.Request) {
	identity := authn.IdentityFromContext(r.Context())
	if identity == nil {
		if r.URL.Query().Get("_login_redirect") == "1" {
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}

		continueURL := h.baseURL.WithPath("/api/connect/v1/oauth2/device/verify").
			WithQueryValues(r.URL.Query()).
			WithQuery("_login_redirect", "1").
			MustString()
		loginURL := h.baseURL.WithPath("/login").
			WithQuery("continue", continueURL).
			MustString()
		http.Redirect(w, r, loginURL, http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	_ = deviceVerifyTmpl.Execute(w, struct {
		UserCode string
	}{
		UserCode: r.URL.Query().Get("user_code"),
	})
}

// DeviceVerifySubmit handles the device code verification form submission.
// POST /oauth2/device/verify
func (h *OAuth2Handler) DeviceVerifySubmit(w http.ResponseWriter, r *http.Request) {
	identity := authn.IdentityFromContext(r.Context())

	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	userCode := strings.ReplaceAll(r.FormValue("user_code"), "-", "")
	userCode = strings.ToUpper(strings.TrimSpace(userCode))

	if err := h.iam.OAuth2ServerService.AuthorizeDevice(r.Context(), identity.ID, userCode); err != nil {
		h.logger.ErrorCtx(r.Context(), "cannot authorize device", log.Error(err))
		w.Header().Set("Content-Type", "text/html;charset=UTF-8")
		_ = deviceVerifyFailedTmpl.Execute(w, nil)
		return
	}

	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	_ = deviceVerifySuccessTmpl.Execute(w, nil)
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
		return nil, oauth2server.ErrInvalidClient
	}

	clientID, err := gid.ParseGID(clientIDStr)
	if err != nil {
		return nil, oauth2server.ErrInvalidClient
	}

	return h.iam.OAuth2ServerService.AuthenticateClient(r.Context(), clientID, clientSecret)
}

func (h *OAuth2Handler) handleAuthorizationCodeGrant(w http.ResponseWriter, r *http.Request) {
	client, err := h.authenticateClient(r)
	if err != nil {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidClient)
		return
	}

	code := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")
	codeVerifier := r.FormValue("code_verifier")

	if code == "" {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidGrant.WithDescription("missing code"))
		return
	}

	tokenResponse, err := h.iam.OAuth2ServerService.ExchangeAuthorizationCode(
		r.Context(),
		client,
		code,
		redirectURI,
		codeVerifier,
	)
	if err != nil {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidGrant.WithDescription("invalid or expired code"))
		return
	}

	writeTokenResponse(w, tokenResponse)
}

func (h *OAuth2Handler) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	client, err := h.authenticateClient(r)
	if err != nil {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidClient)
		return
	}

	refreshTokenValue := r.FormValue("refresh_token")
	if refreshTokenValue == "" {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidGrant.WithDescription("missing refresh_token"))
		return
	}

	tokenResponse, err := h.iam.OAuth2ServerService.RefreshToken(r.Context(), client, refreshTokenValue)
	if err != nil {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidGrant.WithDescription("invalid or expired refresh token"))
		return
	}

	writeTokenResponse(w, tokenResponse)
}

func (h *OAuth2Handler) handleDeviceCodeGrant(w http.ResponseWriter, r *http.Request) {
	clientIDStr := r.FormValue("client_id")
	deviceCodeValue := r.FormValue("device_code")

	if clientIDStr == "" || deviceCodeValue == "" {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidRequest.WithDescription("missing client_id or device_code"))
		return
	}

	clientID, err := gid.ParseGID(clientIDStr)
	if err != nil {
		h.writeOAuth2Error(w, r, oauth2server.ErrInvalidRequest.WithDescription("invalid client_id"))
		return
	}

	tokenResponse, err := h.iam.OAuth2ServerService.PollDeviceCode(
		r.Context(),
		clientID,
		deviceCodeValue,
	)
	if err != nil {
		h.writeOAuth2Error(w, r, err)
		return
	}

	writeTokenResponse(w, tokenResponse)
}

func (h *OAuth2Handler) renderConsentPage(
	w http.ResponseWriter,
	consentID gid.GID,
	client *coredata.OAuth2Client,
	scopes coredata.OAuth2Scopes,
) {
	w.Header().Set("Content-Type", "text/html;charset=UTF-8")
	_ = consentTmpl.Execute(w, struct {
		ConsentID  string
		ClientName string
		Scopes     coredata.OAuth2Scopes
	}{
		ConsentID:  consentID.String(),
		ClientName: client.ClientName,
		Scopes:     scopes,
	})
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

func writeTokenResponse(w http.ResponseWriter, resp *oauth2server.TokenResponse) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	httpserver.RenderJSON(w, http.StatusOK, resp)
}

func parseScopes(s string) (coredata.OAuth2Scopes, error) {
	var scopes coredata.OAuth2Scopes
	if err := scopes.UnmarshalText([]byte(s)); err != nil {
		return nil, oauth2server.ErrInvalidScope.WithDescription(err.Error())
	}
	return scopes, nil
}

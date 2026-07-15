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

package complianceportal_v1

import (
	"errors"
	"net/http"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/complianceportal/visitor"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/oauth2"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/complianceportal"
)

var (
	errNotFound            = errors.New("not found")
	errInvalidContinueURL  = errors.New("invalid continue URL")
	errInternal            = errors.New("internal server error")
	errInvalidOAuthRequest = errors.New("invalid oauth request")
)

type OAuthCallbackHandler struct {
	iam           *iam.Service
	visitor       *visitor.Service
	sessionCookie *authn.Cookie
	safeRedirect  *saferedirect.SafeRedirect
	logger        *log.Logger
}

func NewOAuthCallbackHandler(
	iamSvc *iam.Service,
	visitorSvc *visitor.Service,
	cookieConfig securecookie.Config,
	allowedHost saferedirect.AllowedHostFunc,
	logger *log.Logger,
) *OAuthCallbackHandler {
	return &OAuthCallbackHandler{
		iam:           iamSvc,
		visitor:       visitorSvc,
		sessionCookie: authn.NewCookie(&cookieConfig),
		safeRedirect:  saferedirect.New(allowedHost),
		logger:        logger,
	}
}

func (h *OAuthCallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if oauthErr := r.URL.Query().Get("error"); oauthErr != "" {
		h.logger.WarnCtx(
			ctx,
			"oauth callback returned error",
			log.String("error", oauthErr),
			log.String("error_description", r.URL.Query().Get("error_description")),
		)
		httpserver.RenderError(w, http.StatusBadRequest, errInvalidOAuthRequest)

		return
	}

	code := r.URL.Query().Get("code")
	stateToken := r.URL.Query().Get("state")
	if code == "" || stateToken == "" {
		httpserver.RenderError(w, http.StatusBadRequest, errInvalidOAuthRequest)
		return
	}

	state, err := h.visitor.ConsumeOAuthState(ctx, stateToken)
	if err != nil {
		h.logger.WarnCtx(ctx, "invalid oauth state", log.Error(err))
		httpserver.RenderError(w, http.StatusBadRequest, errInvalidOAuthRequest)

		return
	}

	portal := complianceportal.CompliancePageFromContext(ctx)
	portalBaseURL := complianceportal.CompliancePageBaseURLFromContext(ctx)
	if portal == nil || portalBaseURL == nil {
		httpserver.RenderError(w, http.StatusNotFound, errNotFound)
		return
	}

	clientID, err := complianceportal.CIMDClientIDURL(*portalBaseURL)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot build cimd client_id", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, errInternal)

		return
	}

	redirectURI, err := complianceportal.OAuthCallbackURL(*portalBaseURL)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot build oauth redirect_uri", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, errInternal)

		return
	}

	tokenResult, err := h.iam.OAuth2ServerService.ExchangeAuthorizationCode(
		ctx,
		clientID,
		code,
		redirectURI,
		state.CodeVerifier,
	)
	if err != nil {
		h.logger.WarnCtx(ctx, "cannot exchange authorization code", log.Error(err))
		httpserver.RenderError(w, http.StatusBadRequest, errInvalidOAuthRequest)

		return
	}

	identityID, err := oauth2.ParseIDTokenIdentity(
		tokenResult.IDToken,
		h.iam.OAuth2ServerService.JWKS(),
		state.Nonce,
		h.iam.OAuth2ServerService.Issuer(),
		tokenResult.ClientID.String(),
	)
	if err != nil {
		h.logger.WarnCtx(ctx, "cannot validate id token", log.Error(err))
		httpserver.RenderError(w, http.StatusBadRequest, errInvalidOAuthRequest)

		return
	}

	host, ok := complianceportal.TrustedRequestHost(r)
	if !ok {
		httpserver.RenderError(w, http.StatusBadRequest, errInvalidOAuthRequest)
		return
	}

	session, err := h.iam.AuthService.OpenRootSession(
		ctx,
		identityID,
		coredata.AuthMethodOIDC,
		coredata.SessionDataForHost(host),
	)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot open session", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, errInternal)

		return
	}

	if _, err := h.visitor.ProvisionPortalMember(ctx, portal.ID, identityID); err != nil {
		h.logger.ErrorCtx(ctx, "cannot provision portal member", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, errInternal)

		return
	}

	h.sessionCookie.Set(w, session)

	continueURL := state.ContinueURL
	if continueURL == "" {
		continueURL = "/"
	}

	h.safeRedirect.Redirect(w, r, continueURL, "/", http.StatusFound)
}

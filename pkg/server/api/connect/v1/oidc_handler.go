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

package connect_v1

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
)

type OIDCHandler struct {
	iam           *iam.Service
	sessionCookie *authn.Cookie
	logger        *log.Logger
	safeRedirect  *saferedirect.SafeRedirect
}

func NewOIDCHandler(
	iam *iam.Service,
	cookieConfig securecookie.Config,
	logger *log.Logger,
	allowedHost saferedirect.AllowedHostFunc,
) *OIDCHandler {
	return &OIDCHandler{
		iam:           iam,
		sessionCookie: authn.NewCookie(&cookieConfig),
		logger:        logger,
		safeRedirect:  saferedirect.New(allowedHost),
	}
}

func (h *OIDCHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	provider, err := parseOIDCProvider(chi.URLParam(r, "provider"))
	if err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid provider"))
		return
	}

	if !h.iam.OIDCService.IsProviderEnabled(provider) {
		httpserver.RenderError(w, http.StatusNotFound, errors.New("provider not enabled"))
		return
	}

	continueURL := r.URL.Query().Get("continue")

	var organizationID *gid.GID

	if organizationIDParam := r.URL.Query().Get("organization_id"); organizationIDParam != "" {
		parsedOrganizationID, err := gid.ParseGID(organizationIDParam)
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid organization_id parameter"))
			return
		}

		organizationID = &parsedOrganizationID
	}

	authURL, err := h.iam.OIDCService.InitiateLogin(ctx, provider, continueURL, organizationID)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot initiate OIDC login", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))

		return
	}

	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *OIDCHandler) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	provider, err := parseOIDCProvider(chi.URLParam(r, "provider"))
	if err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid provider"))
		return
	}

	errParam := r.URL.Query().Get("error")
	if errParam != "" {
		h.logger.WarnCtx(
			ctx,
			"OIDC provider returned error",
			log.String("error", errParam),
			log.String("error_description", r.URL.Query().Get("error_description")),
		)
		httpserver.RenderError(w, http.StatusUnauthorized, errors.New("authentication failed"))

		return
	}

	stateParam := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if stateParam == "" || code == "" {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("missing state or code"))
		return
	}

	identity, continueURL, organizationID, err := h.iam.OIDCService.HandleCallback(ctx, provider, stateParam, code)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot handle OIDC callback", log.Error(err))
		httpserver.RenderError(w, http.StatusUnauthorized, errors.New("authentication failed"))

		return
	}

	rootSession := authn.SessionFromContext(ctx)

	switch {
	case rootSession == nil:
		rootSession, err = h.iam.AuthService.OpenSessionWithOIDC(ctx, identity.ID, coredata.AuthMethodOIDC)
		if err != nil {
			h.logger.ErrorCtx(ctx, "cannot open root session", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))

			return
		}
	case rootSession.IdentityID != identity.ID:
		err = h.iam.SessionService.CloseSession(ctx, rootSession.ID)
		if err != nil {
			h.logger.ErrorCtx(ctx, "cannot close session", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))

			return
		}

		rootSession, err = h.iam.AuthService.OpenSessionWithOIDC(ctx, identity.ID, coredata.AuthMethodOIDC)
		if err != nil {
			h.logger.ErrorCtx(ctx, "cannot open root session", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))

			return
		}
	}

	if organizationID != nil {
		_, _, err = h.iam.SessionService.OpenOIDCChildSessionForOrganization(ctx, rootSession.ID, *organizationID)
		if err != nil {
			if _, ok := errors.AsType[*iam.ErrMembershipNotFound](err); ok {
				httpserver.RenderError(w, http.StatusNotFound, errors.New("not found"))

				return
			}

			if _, ok := errors.AsType[*iam.ErrProfileNotFound](err); ok {
				httpserver.RenderError(w, http.StatusNotFound, errors.New("not found"))

				return
			}

			if _, ok := errors.AsType[*iam.ErrUserInactive](err); ok {
				httpserver.RenderError(w, http.StatusNotFound, errors.New("not found"))

				return
			}

			h.logger.ErrorCtx(ctx, "cannot open OIDC child session", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))

			return
		}
	}

	h.sessionCookie.Set(w, rootSession)

	defaultRedirect := "/"
	if organizationID != nil {
		defaultRedirect = "/organizations/" + organizationID.String()
	}

	redirectURL := h.safeRedirect.GetSafeRedirectURL(ctx, continueURL, defaultRedirect)
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func parseOIDCProvider(s string) (coredata.OIDCProvider, error) {
	switch strings.ToLower(s) {
	case "google":
		return coredata.OIDCProviderGoogle, nil
	case "microsoft":
		return coredata.OIDCProviderMicrosoft, nil
	default:
		return "", errors.New("unknown provider")
	}
}

type MagicLinkHandler struct {
	iam           *iam.Service
	proboBaseURL  *baseurl.BaseURL
	sessionCookie *authn.Cookie
	safeRedirect  *saferedirect.SafeRedirect
	logger        *log.Logger
}

func NewMagicLinkHandler(
	iamSvc *iam.Service,
	proboBaseURL *baseurl.BaseURL,
	cookieConfig securecookie.Config,
	logger *log.Logger,
	allowedHost saferedirect.AllowedHostFunc,
) *MagicLinkHandler {
	return &MagicLinkHandler{
		iam:           iamSvc,
		proboBaseURL:  proboBaseURL,
		sessionCookie: authn.NewCookie(&cookieConfig),
		safeRedirect:  saferedirect.New(allowedHost),
		logger:        logger,
	}
}

func (h *MagicLinkHandler) SendHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid form data"))
		return
	}

	emailAddr, err := mail.ParseAddr(r.FormValue("email"))
	if err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid email"))
		return
	}

	continueParam := r.FormValue("continue")
	if continueParam == "" {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid magic link parameters"))
		return
	}

	safeContinue, ok := h.safeRedirect.Validate(ctx, continueParam)
	if !ok {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid continue URL"))
		return
	}

	proboURL := h.proboBaseURL.String()

	req := &iam.SendMagicLinkRequest{
		Email:            emailAddr,
		URLPath:          "/api/connect/v1/magic-link/verify",
		MagicLinkBaseURL: &proboURL,
		Continue:         &safeContinue,
	}

	if clientID := oauth2ClientIDFromContinueURL(safeContinue); clientID != "" {
		req.OAuth2ClientIDRaw = &clientID
	}

	if err := h.iam.AuthService.SendMagicLink(ctx, req); err != nil {
		h.logger.ErrorCtx(ctx, "cannot send magic link", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MagicLinkHandler) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := r.URL.Query().Get("token")
	if token == "" {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("missing token"))
		return
	}

	identity, session, continueURL, err := h.iam.AuthService.OpenSessionWithMagicLink(ctx, token)
	if err != nil {
		if _, ok := errors.AsType[*iam.ErrExpiredToken](err); ok {
			http.Redirect(w, r, "/auth/magic-link-expired", http.StatusFound)
			return
		}

		if _, ok := errors.AsType[*iam.ErrTokenAlreadyUsed](err); ok {
			http.Redirect(w, r, "/auth/magic-link-already-used", http.StatusFound)
			return
		}

		if _, ok := errors.AsType[*iam.ErrInvalidToken](err); ok {
			httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid token"))
			return
		}

		h.logger.ErrorCtx(ctx, "cannot open session with magic link", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))

		return
	}

	_ = identity

	h.sessionCookie.Set(w, session)

	metadata := OAuth2ServerMetadata(
		h.proboBaseURL,
		h.iam.OAuth2ScopeRegistry.RegisteredScopes(),
	)

	redirectURL := metadata.AuthorizationEndpoint.String()
	if continueURL != nil && *continueURL != "" {
		redirectURL = *continueURL
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

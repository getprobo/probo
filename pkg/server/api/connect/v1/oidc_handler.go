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
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/oidc"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
)

const (
	oidcLoginCookieName = "OIDC_LOGIN"
	oidcLoginCookiePath = "/api/connect/v1/oidc"
)

type OIDCHandler struct {
	iam           *iam.Service
	sessionCookie *authn.Cookie
	logger        *log.Logger
	safeRedirect  *saferedirect.SafeRedirect
	cookieSecure  bool
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
		cookieSecure:  cookieConfig.Secure,
	}
}

func (h *OIDCHandler) redirectAuthError(w http.ResponseWriter, r *http.Request, code string, continueURL string) {
	safeContinue := ""

	if continueURL != "" {
		if validated, ok := h.safeRedirect.Validate(r.Context(), continueURL); ok {
			safeContinue = validated
		}
	}

	redirectAuthError(w, r, code, safeContinue)
}

func (h *OIDCHandler) continueURLFromCallbackState(
	ctx context.Context,
	provider coredata.OIDCProvider,
	stateParam string,
	browserNonce string,
) (string, error) {
	return h.iam.OIDCService.ContinueURLFromCallbackState(ctx, provider, stateParam, browserNonce)
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

	authURL, browserNonce, err := h.iam.OIDCService.InitiateLogin(ctx, provider, continueURL, organizationID)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot initiate OIDC login", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, errors.New("internal server error"))

		return
	}

	h.setOIDCLoginCookie(w, browserNonce)
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (h *OIDCHandler) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	provider, err := parseOIDCProvider(chi.URLParam(r, "provider"))
	if err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid provider"))
		return
	}

	browserNonce := oidcLoginNonce(r)

	errParam := r.URL.Query().Get("error")
	if errParam != "" {
		h.logger.ErrorCtx(
			ctx,
			"OIDC provider returned error",
			log.String("error", errParam),
			log.String("error_description", r.URL.Query().Get("error_description")),
		)

		continueURL, stateErr := h.continueURLFromCallbackState(ctx, provider, r.URL.Query().Get("state"), browserNonce)
		if h.rejectUnboundOIDCLogin(w, stateErr) {
			return
		}

		h.redirectAuthError(w, r, authErrorAuthenticationFailed, continueURL)

		return
	}

	stateParam := r.URL.Query().Get("state")
	code := r.URL.Query().Get("code")

	if stateParam == "" || code == "" {
		continueURL := ""

		if stateParam != "" {
			var stateErr error

			continueURL, stateErr = h.continueURLFromCallbackState(ctx, provider, stateParam, browserNonce)
			if h.rejectUnboundOIDCLogin(w, stateErr) {
				return
			}
		}

		h.redirectAuthError(w, r, authErrorInvalidState, continueURL)

		return
	}

	identity, continueURL, organizationID, err := h.iam.OIDCService.HandleCallback(
		ctx,
		provider,
		stateParam,
		code,
		browserNonce,
	)
	if h.rejectUnboundOIDCLogin(w, err) {
		return
	}

	if err != nil {
		if _, ok := errors.AsType[*oidc.ErrPersonalAccountNotAllowed](err); ok {
			h.redirectAuthError(w, r, authErrorPersonalAccountNotAllowed, continueURL)

			return
		}

		if _, ok := errors.AsType[*oidc.ErrEmailNotVerified](err); ok {
			h.redirectAuthError(w, r, authErrorEmailNotVerified, continueURL)

			return
		}

		if _, ok := errors.AsType[*oidc.ErrInvalidState](err); ok {
			h.redirectAuthError(w, r, authErrorInvalidState, continueURL)

			return
		}

		h.logger.ErrorCtx(ctx, "cannot handle OIDC callback", log.Error(err))
		h.redirectAuthError(w, r, authErrorAuthenticationFailed, continueURL)

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

func (h *OIDCHandler) rejectUnboundOIDCLogin(w http.ResponseWriter, err error) bool {
	h.clearOIDCLoginCookieIfConsumed(w, err)

	if _, ok := errors.AsType[*oidc.ErrBrowserBinding](err); ok {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid oidc login"))

		return true
	}

	return false
}

func (h *OIDCHandler) clearOIDCLoginCookieIfConsumed(w http.ResponseWriter, err error) {
	if _, ok := errors.AsType[*oidc.ErrBrowserBinding](err); ok {
		return
	}

	if _, ok := errors.AsType[*oidc.ErrInvalidState](err); ok {
		return
	}

	h.clearOIDCLoginCookie(w)
}

func (h *OIDCHandler) setOIDCLoginCookie(w http.ResponseWriter, nonce string) {
	http.SetCookie(
		w,
		&http.Cookie{
			Name:     oidcLoginCookieName,
			Value:    nonce,
			Path:     oidcLoginCookiePath,
			MaxAge:   int(oidc.LoginAttemptTTL.Seconds()),
			Secure:   h.cookieSecure,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
	)
}

func (h *OIDCHandler) clearOIDCLoginCookie(w http.ResponseWriter) {
	http.SetCookie(
		w,
		&http.Cookie{
			Name:     oidcLoginCookieName,
			Value:    "",
			Path:     oidcLoginCookiePath,
			MaxAge:   -1,
			Secure:   h.cookieSecure,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
	)
}

func oidcLoginNonce(r *http.Request) string {
	cookie, err := r.Cookie(oidcLoginCookieName)
	if err != nil {
		return ""
	}

	return cookie.Value
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

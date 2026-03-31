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
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
)

// IsTrustCenterDomainFunc checks whether a given host is a trust center
// custom domain.
type IsTrustCenterDomainFunc func(ctx context.Context, host string) bool

type OIDCHandler struct {
	iam                  *iam.Service
	sessionCookie        *authn.Cookie
	cookieSecret         string
	logger               *log.Logger
	safeRedirect         *saferedirect.SafeRedirect
	isTrustCenterDomain  IsTrustCenterDomainFunc
}

func NewOIDCHandler(
	iam *iam.Service,
	cookieConfig securecookie.Config,
	logger *log.Logger,
	allowedHost saferedirect.AllowedHostFunc,
	isTrustCenterDomain IsTrustCenterDomainFunc,
) *OIDCHandler {
	return &OIDCHandler{
		iam:                 iam,
		sessionCookie:       authn.NewCookie(&cookieConfig),
		cookieSecret:        cookieConfig.Secret,
		logger:              logger,
		safeRedirect:        saferedirect.New(allowedHost),
		isTrustCenterDomain: isTrustCenterDomain,
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

	authURL, err := h.iam.OIDCService.InitiateLogin(ctx, provider, continueURL)
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

	identity, continueURL, err := h.iam.OIDCService.HandleCallback(ctx, provider, stateParam, code)
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

	h.sessionCookie.Set(w, rootSession)

	redirectURL := h.safeRedirect.GetSafeRedirectURL(ctx, continueURL, "/")

	if transferURL, ok := h.buildSessionTransferURL(ctx, redirectURL, rootSession.ID.String()); ok {
		http.Redirect(w, r, transferURL, http.StatusFound)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

// buildSessionTransferURL returns a session-transfer URL on the target trust
// center custom domain. This lets the custom domain set its own cookie for the
// session.
func (h *OIDCHandler) buildSessionTransferURL(ctx context.Context, redirectURL string, sessionID string) (string, bool) {
	parsed, err := url.Parse(redirectURL)
	if err != nil || !parsed.IsAbs() {
		return "", false
	}

	if !h.isTrustCenterDomain(ctx, parsed.Host) {
		return "", false
	}

	token, err := authn.SignSessionTransfer(sessionID, redirectURL, h.cookieSecret)
	if err != nil {
		h.logger.Error("cannot sign session transfer token", log.Error(err))
		return "", false
	}

	transferURL := &url.URL{
		Scheme: parsed.Scheme,
		Host:   parsed.Host,
		Path:   "/api/trust/v1/session-transfer",
	}

	q := transferURL.Query()
	q.Set("token", token)
	transferURL.RawQuery = q.Encode()

	return transferURL.String(), true
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

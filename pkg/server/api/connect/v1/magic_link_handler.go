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

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mail"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
)

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

func (h *MagicLinkHandler) redirectAuthError(w http.ResponseWriter, r *http.Request, code string, token string) {
	safeContinue := ""

	if token != "" {
		continueURL, err := h.iam.AuthService.MagicLinkContinueFromToken(token)
		if err == nil && continueURL != nil && *continueURL != "" {
			if validated, ok := h.safeRedirect.Validate(r.Context(), *continueURL); ok {
				safeContinue = validated
			}
		}
	}

	redirectAuthError(w, r, code, safeContinue)
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
		h.redirectAuthError(w, r, authErrorMagicLinkInvalid, "")

		return
	}

	identity, session, continueURL, err := h.iam.AuthService.OpenSessionWithMagicLink(ctx, token)
	if err != nil {
		if _, ok := errors.AsType[*iam.ErrExpiredToken](err); ok {
			h.redirectAuthError(w, r, authErrorMagicLinkExpired, token)

			return
		}

		if _, ok := errors.AsType[*iam.ErrTokenAlreadyUsed](err); ok {
			h.redirectAuthError(w, r, authErrorMagicLinkAlreadyUsed, token)

			return
		}

		if _, ok := errors.AsType[*iam.ErrInvalidToken](err); ok {
			h.redirectAuthError(w, r, authErrorMagicLinkInvalid, token)

			return
		}

		h.logger.ErrorCtx(ctx, "cannot open session with magic link", log.Error(err))
		h.redirectAuthError(w, r, authErrorAuthenticationFailed, token)

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

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
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS
// ACTION, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE
// OR OTHER DEALINGS IN THE SOFTWARE.

package connect_v1

import (
	"errors"
	"net/http"
	"net/url"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
)

const compliancePortalInviteConfirmPath = "/auth/compliance-portal-invite"

type CompliancePortalInviteHandler struct {
	iam           *iam.Service
	proboBaseURL  *baseurl.BaseURL
	sessionCookie *authn.Cookie
	safeRedirect  *saferedirect.SafeRedirect
	logger        *log.Logger
}

func NewCompliancePortalInviteHandler(
	iamSvc *iam.Service,
	proboBaseURL *baseurl.BaseURL,
	cookieConfig securecookie.Config,
	logger *log.Logger,
	allowedHost saferedirect.AllowedHostFunc,
) *CompliancePortalInviteHandler {
	return &CompliancePortalInviteHandler{
		iam:           iamSvc,
		proboBaseURL:  proboBaseURL,
		sessionCookie: authn.NewCookie(&cookieConfig),
		safeRedirect:  saferedirect.New(allowedHost),
		logger:        logger,
	}
}

func (h *CompliancePortalInviteHandler) redirectAuthError(
	w http.ResponseWriter,
	r *http.Request,
	code string,
	token string,
) {
	safeContinue := ""

	if token != "" {
		continueURL, err := h.iam.AuthService.CompliancePortalInviteContinueFromToken(token)
		if err == nil && continueURL != nil && *continueURL != "" {
			if validated, ok := h.safeRedirect.Validate(r.Context(), *continueURL); ok {
				safeContinue = validated
			}
		}
	}

	redirectAuthError(w, r, code, safeContinue)
}

// ConfirmRedirectHandler is GET /compliance-portal-invite/verify. Scanners
// prefetch GET URLs, so this must not consume the token.
func (h *CompliancePortalInviteHandler) ConfirmRedirectHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		h.redirectAuthError(w, r, authErrorMagicLinkInvalid, "")

		return
	}

	redirectURL := url.URL{
		Path:     compliancePortalInviteConfirmPath,
		RawQuery: url.Values{"token": {token}}.Encode(),
	}

	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}

func (h *CompliancePortalInviteHandler) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		h.redirectAuthError(w, r, authErrorMagicLinkInvalid, "")

		return
	}

	token := r.FormValue("token")
	if token == "" {
		h.redirectAuthError(w, r, authErrorMagicLinkInvalid, "")

		return
	}

	identity, session, continueURL, err := h.iam.AuthService.OpenSessionWithCompliancePortalInvite(ctx, token)
	if err != nil {
		if _, ok := errors.AsType[*iam.ErrExpiredToken](err); ok {
			h.redirectAuthError(w, r, authErrorMagicLinkExpired, token)

			return
		}

		if _, ok := errors.AsType[*iam.ErrInvalidToken](err); ok {
			h.redirectAuthError(w, r, authErrorMagicLinkInvalid, token)

			return
		}

		h.logger.ErrorCtx(ctx, "cannot open session with compliance portal invite", log.Error(err))
		h.redirectAuthError(w, r, authErrorAuthenticationFailed, token)

		return
	}

	_ = identity

	h.sessionCookie.Set(w, session)

	redirectURL := h.proboBaseURL.String()
	if continueURL != nil && *continueURL != "" {
		redirectURL = *continueURL
	}

	http.Redirect(w, r, redirectURL, http.StatusFound)
}

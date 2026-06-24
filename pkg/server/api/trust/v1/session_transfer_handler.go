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

package trust_v1

import (
	"errors"
	"net/http"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
)

type SessionTransferHandler struct {
	iam           *iam.Service
	sessionCookie *authn.Cookie
	cookieSecret  string
	safeRedirect  *saferedirect.SafeRedirect
	logger        *log.Logger
}

func NewSessionTransferHandler(
	iamSvc *iam.Service,
	cookieConfig securecookie.Config,
	allowedHost saferedirect.AllowedHostFunc,
	logger *log.Logger,
) *SessionTransferHandler {
	return &SessionTransferHandler{
		iam:           iamSvc,
		sessionCookie: authn.NewCookie(&cookieConfig),
		cookieSecret:  cookieConfig.Secret,
		safeRedirect:  saferedirect.New(allowedHost),
		logger:        logger,
	}
}

func (h *SessionTransferHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := r.URL.Query().Get("token")
	if token == "" {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("missing token"))
		return
	}

	claims, err := authn.VerifySessionTransfer(token, h.cookieSecret)
	if err != nil {
		h.logger.WarnCtx(ctx, "invalid session transfer token", log.Error(err))
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid or expired token"))

		return
	}

	continueURL := claims.ContinueURL
	if continueURL == "" {
		continueURL = "/"
	}

	sessionID, err := gid.ParseGID(claims.SessionID)
	if err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid token"))
		return
	}

	session, err := h.iam.SessionService.GetSession(ctx, sessionID)
	if err != nil {
		h.logger.ErrorCtx(ctx, "cannot get session for transfer", log.Error(err))
		httpserver.RenderError(w, http.StatusBadRequest, errors.New("invalid or expired token"))

		return
	}

	h.sessionCookie.Set(w, session)

	h.safeRedirect.Redirect(w, r, continueURL, "/", http.StatusFound)
}

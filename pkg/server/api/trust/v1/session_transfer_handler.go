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

package trust_v1

import (
	"errors"
	"net/http"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
)

type SessionTransferHandler struct {
	iam           *iam.Service
	sessionCookie *authn.Cookie
	cookieSecret  string
	logger        *log.Logger
}

func NewSessionTransferHandler(
	iamSvc *iam.Service,
	cookieConfig securecookie.Config,
	logger *log.Logger,
) *SessionTransferHandler {
	return &SessionTransferHandler{
		iam:           iamSvc,
		sessionCookie: authn.NewCookie(&cookieConfig),
		cookieSecret:  cookieConfig.Secret,
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

	http.Redirect(w, r, continueURL, http.StatusFound)
}

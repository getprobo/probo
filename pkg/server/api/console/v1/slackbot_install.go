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

package console_v1

import (
	"errors"
	"fmt"
	"net/http"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	slackchannel "go.probo.inc/probo/pkg/probot/channel/slack"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/server/api/authn"
)

func handleSlackbotInstallInitiate(
	logger *log.Logger,
	iamSvc *iam.Service,
	installations *slackchannel.InstallationService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if installations == nil {
			httpserver.RenderError(w, http.StatusNotFound, fmt.Errorf("slack app is disabled"))
			return
		}

		organizationID, err := gid.ParseGID(r.URL.Query().Get("organization_id"))
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("invalid organization_id parameter"))
			return
		}

		identity := authn.IdentityFromContext(r.Context())

		session := authn.SessionFromContext(r.Context())
		if identity == nil || session == nil {
			httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("authentication required"))
			return
		}

		if _, err := iamSvc.Authorizer.Authorize(
			r.Context(),
			iam.AuthorizeParams{
				Principal: identity.ID,
				Resource:  organizationID,
				Session:   &session.ID,
				Action:    probo.ActionConnectorInitiate,
			},
		); err != nil {
			httpserver.RenderError(w, http.StatusForbidden, err)
			return
		}

		redirectURL, err := installations.InitiateURL(
			organizationID,
			identity.ID,
			r.URL.Query().Get("continue"),
		)
		if err != nil {
			logger.ErrorCtx(r.Context(), "cannot initiate Slack app installation", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

			return
		}

		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}
}

func handleSlackbotInstallComplete(
	logger *log.Logger,
	baseURL *baseurl.BaseURL,
	installations *slackchannel.InstallationService,
	safeRedirect *saferedirect.SafeRedirect,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if installations == nil {
			httpserver.RenderError(w, http.StatusNotFound, fmt.Errorf("slack app is disabled"))
			return
		}

		if oauthError := r.URL.Query().Get("error"); oauthError != "" {
			logger.WarnCtx(
				r.Context(),
				"Slack app installation was rejected",
				log.String("oauth_error", oauthError),
			)
			safeRedirect.Redirect(w, r, baseURL.String(), "/", http.StatusSeeOther)

			return
		}

		state := r.URL.Query().Get("state")

		code := r.URL.Query().Get("code")
		if state == "" || code == "" {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("missing Slack OAuth callback parameters"))
			return
		}

		result, err := installations.Complete(r.Context(), state, code)
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, slackchannel.ErrSlackbotStateAlreadyUsed) {
				status = http.StatusConflict
			}

			logger.ErrorCtx(r.Context(), "cannot complete Slack app installation", log.Error(err))
			httpserver.RenderError(w, status, fmt.Errorf("cannot complete Slack app installation"))

			return
		}

		safeRedirect.Redirect(
			w,
			r,
			result.ContinueURL,
			"/",
			http.StatusSeeOther,
		)
	}
}

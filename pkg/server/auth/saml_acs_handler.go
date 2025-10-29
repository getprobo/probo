// Copyright (c) 2025 Probo Inc <hello@getprobo.com>.
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

package auth

import (
	"fmt"
	"net/http"
	"time"

	authsvc "github.com/getprobo/probo/pkg/auth"
	"github.com/getprobo/probo/pkg/authz"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/securecookie"
	"go.gearno.de/kit/log"
)

func getSessionIDFromCookie(r *http.Request, authCfg RoutesConfig) (gid.GID, error) {
	cookieValue, err := securecookie.Get(r, securecookie.DefaultConfig(
		authCfg.CookieName,
		authCfg.CookieSecret,
	))
	if err != nil {
		return gid.GID{}, err
	}

	return gid.ParseGID(cookieValue)
}

func SAMLACSHandler(samlSvc *authsvc.SAMLService, authSvc *authsvc.Service, authzSvc *authz.Service, authCfg RoutesConfig, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if err := r.ParseForm(); err != nil {
			logger.ErrorCtx(ctx, "failed to parse form", log.Error(err))
			http.Error(w, "failed to parse form", http.StatusBadRequest)
			return
		}

		if r.FormValue("SAMLResponse") == "" {
			logger.WarnCtx(ctx, "missing SAMLResponse")
			http.Error(w, "missing SAMLResponse", http.StatusBadRequest)
			return
		}

		if r.FormValue("RelayState") == "" {
			logger.WarnCtx(ctx, "missing RelayState")
			http.Error(w, "missing RelayState", http.StatusBadRequest)
			return
		}

		userInfo, err := samlSvc.HandleSAMLAssertion(ctx, r)
		if err != nil {
			logger.ErrorCtx(ctx, "SAML authentication failed", log.Error(err))
			http.Error(w, "SAML authentication failed", http.StatusUnauthorized)
			return
		}

		user, err := authSvc.CreateOrGetSAMLUser(ctx, userInfo.Email, userInfo.FullName, userInfo.SAMLSubject)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot create or get SAML user", log.Error(err), log.String("email", userInfo.Email))
			http.Error(w, "failed to create user", http.StatusInternalServerError)
			return
		}

		err = authzSvc.EnsureSAMLMembership(ctx, userInfo.TenantID, user.ID, userInfo.OrganizationID, userInfo.Role)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot ensure membership", log.Error(err), log.String("user_id", user.ID.String()), log.String("org_id", userInfo.OrganizationID.String()))
			http.Error(w, "failed to create membership", http.StatusInternalServerError)
			return
		}

		var session *coredata.Session
		if existingSessionID, err := getSessionIDFromCookie(r, authCfg); err == nil {
			if existingSession, err := authSvc.GetSession(ctx, existingSessionID); err == nil && existingSession.UserID == user.ID {
				session = existingSession
			}
		}

		if session == nil {
			session, err = authSvc.CreateSessionForUser(ctx, user.ID, authCfg.SessionDuration)
			if err != nil {
				logger.ErrorCtx(ctx, "cannot create session", log.Error(err), log.String("user_id", user.ID.String()))
				http.Error(w, "failed to create session", http.StatusInternalServerError)
				return
			}
		}

		if session.Data.SAMLAuthenticatedOrgs == nil {
			session.Data.SAMLAuthenticatedOrgs = make(map[string]coredata.SAMLAuthInfo)
		}
		session.Data.SAMLAuthenticatedOrgs[userInfo.OrganizationID.String()] = coredata.SAMLAuthInfo{
			AuthenticatedAt: time.Now(),
			SAMLConfigID:    userInfo.SAMLConfigID,
			SAMLSubject:     userInfo.SAMLSubject,
		}

		err = authSvc.UpdateSessionData(ctx, session.ID, session.Data)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot update session data", log.Error(err), log.String("session_id", session.ID.String()))
			http.Error(w, "failed to update session", http.StatusInternalServerError)
			return
		}

		securecookie.Set(
			w,
			securecookie.DefaultConfig(
				authCfg.CookieName,
				authCfg.CookieSecret,
			),
			session.ID.String(),
		)

		logger.InfoCtx(ctx, "SAML login successful", log.String("user_id", user.ID.String()), log.String("org_id", userInfo.OrganizationID.String()))

		redirectURL := fmt.Sprintf("/organizations/%s", userInfo.OrganizationID)
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}

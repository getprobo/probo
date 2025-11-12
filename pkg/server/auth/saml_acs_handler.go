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
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.gearno.de/kit/log"
	authsvc "go.probo.inc/probo/pkg/auth"
	"go.probo.inc/probo/pkg/authz"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/securecookie"
)

func getSessionIDFromCookie(r *http.Request, cookieName string, cookieSecret string, cookieSecure bool) (gid.GID, error) {
	cookieValue, err := securecookie.Get(
		r,
		securecookie.DefaultConfig(
			cookieName,
			cookieSecret,
			cookieSecure,
		),
	)
	if err != nil {
		return gid.GID{}, err
	}

	return gid.ParseGID(cookieValue)
}

func SAMLACSHandler(samlSvc *authsvc.SAMLService, authSvc *authsvc.Service, authzSvc *authz.Service, cookieName string, cookieSecret string, cookieSecure bool, sessionDuration time.Duration, logger *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if err := r.ParseForm(); err != nil {
			logger.ErrorCtx(ctx, "cannot parse form", log.Error(err))
			http.Error(w, "cannot parse form", http.StatusBadRequest)
			return
		}

		if r.FormValue("SAMLResponse") == "" {
			logger.WarnCtx(ctx, "missing SAMLResponse")
			http.Error(w, "missing SAMLResponse", http.StatusBadRequest)
			return
		}

		relayState := r.FormValue("RelayState")
		samlConfigID := r.URL.Query().Get("c")

		if relayState != "" {
			logger.InfoCtx(ctx, "processing SP-initiated SAML login", log.String("relay_state", relayState))
		} else {
			logger.InfoCtx(ctx, "processing IDP-initiated SAML login", log.String("config_id", samlConfigID))
		}

		userInfo, err := samlSvc.HandleSAMLAssertion(ctx, r)
		if err != nil {
			logger.ErrorCtx(ctx, "SAML authentication failed", log.Error(err))
			http.Error(w, "SAML authentication failed", http.StatusUnauthorized)
			return
		}

		var existingSession *coredata.Session
		if existingSessionID, err := getSessionIDFromCookie(r, cookieName, cookieSecret, cookieSecure); err == nil {
			if session, err := authSvc.GetSession(ctx, existingSessionID); err == nil {
				existingSession = session
			}
		}

		session, user, err := authSvc.ProvisionSAMLUser(
			ctx,
			userInfo.SAMLConfigID,
			userInfo.OrganizationID,
			userInfo.Email,
			userInfo.FullName,
			userInfo.SAMLSubject,
			existingSession,
			sessionDuration,
		)
		if err != nil {
			var autoSignupDisabledErr *authsvc.ErrSAMLAutoSignupDisabled
			if errors.As(err, &autoSignupDisabledErr) {
				logger.WarnCtx(ctx, "SAML auto-signup is disabled")
				http.Error(w, "User does not exist and auto-signup is disabled for this organization", http.StatusForbidden)
				return
			}
			logger.ErrorCtx(ctx, "cannot provision SAML user", log.Error(err))
			http.Error(w, "cannot provision user", http.StatusInternalServerError)
			return
		}

		tenantAuthzSvc := authzSvc.WithTenant(userInfo.OrganizationID.TenantID())
		err = tenantAuthzSvc.EnsureSAMLMembership(ctx, user.ID, userInfo.OrganizationID, userInfo.Role)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot ensure membership", log.Error(err), log.String("user_id", user.ID.String()), log.String("org_id", userInfo.OrganizationID.String()))
			http.Error(w, "cannot create membership", http.StatusInternalServerError)
			return
		}

		securecookie.Set(
			w,
			securecookie.DefaultConfig(
				cookieName,
				cookieSecret,
				cookieSecure,
			),
			session.ID.String(),
		)

		logger.InfoCtx(ctx, "SAML login successful", log.String("user_id", user.ID.String()), log.String("org_id", userInfo.OrganizationID.String()))

		redirectURL := fmt.Sprintf("/organizations/%s", userInfo.OrganizationID)
		http.Redirect(w, r, redirectURL, http.StatusFound)
	}
}

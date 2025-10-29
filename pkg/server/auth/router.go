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
	"time"

	authsvc "github.com/getprobo/probo/pkg/auth"
	"github.com/getprobo/probo/pkg/authz"
	"github.com/getprobo/probo/pkg/filemanager"
	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
)

type RoutesConfig struct {
	CookieName      string
	CookieDomain    string
	SessionDuration time.Duration
	CookieSecret    string
	FileManager     *filemanager.Service
	PGClient        *pg.Client
}

func MountRoutes(
	r chi.Router,
	authSvc *authsvc.Service,
	authzSvc *authz.Service,
	samlSvc *authsvc.SAMLService,
	authCfg RoutesConfig,
	logger *log.Logger,
) {
	r.Post("/register", SignUpHandler(authSvc, authCfg))
	r.Post("/login", SignInHandler(authSvc, authCfg))
	r.Delete("/logout", SignOutHandler(authSvc, authCfg))
	r.Post("/signup-from-invitation", SignupFromInvitationHandler(authSvc, authCfg))
	r.Post("/forget-password", ForgetPasswordHandler(authSvc, authCfg))
	r.Post("/reset-password", ResetPasswordHandler(authSvc, authCfg))
	r.Post("/check-sso", SAMLCheckSSOHandler(authSvc, logger))
	r.Get("/organizations", ListOrganizationsHandler(authSvc, authzSvc, authCfg))
	r.Get("/invitations", ListInvitationsHandler(authSvc, authzSvc, authCfg))
	r.Post("/invitations/accept", AcceptInvitationHandler(authSvc, authzSvc, authCfg))

	// SAML routes
	r.Get("/saml/login/{samlConfigID}", SAMLLoginHandler(samlSvc, authSvc, logger))
	r.Post("/saml/consume", SAMLACSHandler(samlSvc, authSvc, authzSvc, authCfg, logger))
	r.Get("/saml/metadata", SAMLMetadataHandler(samlSvc))
}

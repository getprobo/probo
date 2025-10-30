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
	"net/http"
	"time"

	authsvc "github.com/getprobo/probo/pkg/auth"
	"github.com/getprobo/probo/pkg/authz"
	"github.com/getprobo/probo/pkg/filemanager"
	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
)

type Config struct {
	Auth            *authsvc.Service
	Authz           *authz.Service
	SAML            *authsvc.SAMLService
	CookieName      string
	CookieDomain    string
	SessionDuration time.Duration
	CookieSecret    string
	FileManager     *filemanager.Service
	Logger          *log.Logger
}

type Server struct {
	router *chi.Mux
}

func NewServer(cfg Config) (*Server, error) {
	router := chi.NewRouter()

	router.Post("/register", SignUpHandler(cfg.Auth, cfg.CookieName, cfg.CookieSecret))
	router.Post("/login", SignInHandler(cfg.Auth, cfg.CookieName, cfg.CookieSecret))
	router.Delete("/logout", SignOutHandler(cfg.Auth, cfg.CookieName, cfg.CookieSecret))
	router.Post("/signup-from-invitation", SignupFromInvitationHandler(cfg.Auth, cfg.CookieName, cfg.CookieSecret))
	router.Post("/forget-password", ForgetPasswordHandler(cfg.Auth))
	router.Post("/reset-password", ResetPasswordHandler(cfg.Auth))
	router.Post("/check-sso", SAMLCheckSSOHandler(cfg.Auth, cfg.Logger))
	router.Get("/organizations", RequireAuth(cfg.Auth, cfg.Authz, cfg.CookieName, cfg.CookieSecret, ListOrganizationsHandler(cfg.Auth, cfg.Authz)))
	router.Get("/organizations/{organizationID}/logo", RequireAuth(cfg.Auth, cfg.Authz, cfg.CookieName, cfg.CookieSecret, OrganizationLogoHandler(cfg.Auth, cfg.FileManager)))
	router.Get("/invitations", RequireAuth(cfg.Auth, cfg.Authz, cfg.CookieName, cfg.CookieSecret, ListInvitationsHandler(cfg.Authz)))
	router.Post("/invitations/accept", AcceptInvitationHandler(cfg.Auth, cfg.Authz, cfg.CookieName, cfg.CookieSecret))

	router.Get("/saml/login/{samlConfigID}", SAMLLoginHandler(cfg.SAML, cfg.Auth, cfg.Logger))
	router.Post("/saml/consume", SAMLACSHandler(cfg.SAML, cfg.Auth, cfg.Authz, cfg.CookieName, cfg.CookieSecret, cfg.SessionDuration, cfg.Logger))
	router.Get("/saml/metadata", SAMLMetadataHandler(cfg.SAML))

	return &Server{
		router: router,
	}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

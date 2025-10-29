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
	"go.gearno.de/kit/pg"
)

type Config struct {
	Auth             *authsvc.Service
	Authz            *authz.Service
	SAML             *authsvc.SAMLService
	CookieName       string
	CookieDomain     string
	SessionDuration  time.Duration
	CookieSecret     string
	FileManager      *filemanager.Service
	PGClient         *pg.Client
	Logger           *log.Logger
}

type Server struct {
	router *chi.Mux
}

func NewServer(cfg Config) (*Server, error) {
	router := chi.NewRouter()

	MountRoutes(
		router,
		cfg.Auth,
		cfg.Authz,
		cfg.SAML,
		RoutesConfig{
			CookieName:      cfg.CookieName,
			CookieDomain:    cfg.CookieDomain,
			SessionDuration: cfg.SessionDuration,
			CookieSecret:    cfg.CookieSecret,
			FileManager:     cfg.FileManager,
			PGClient:        cfg.PGClient,
		},
		cfg.Logger,
	)

	return &Server{
		router: router,
	}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

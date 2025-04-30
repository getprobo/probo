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

package api

import (
	"errors"
	"net/http"

	"github.com/getprobo/probo/pkg/connector"
	"github.com/getprobo/probo/pkg/probo"
	"github.com/getprobo/probo/pkg/saferedirect"
	console_v1 "github.com/getprobo/probo/pkg/server/api/console/v1"
	"github.com/getprobo/probo/pkg/usrmgr"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.gearno.de/kit/httpserver"
)

type (
	Config struct {
		AllowedOrigins    []string
		Probo             *probo.Service
		Usrmgr            *usrmgr.Service
		Auth              console_v1.AuthConfig
		ConnectorRegistry *connector.ConnectorRegistry
		SafeRedirect      *saferedirect.SafeRedirect
	}

	Server struct {
		cfg Config
	}
)

var (
	ErrMissingProboService  = errors.New("server configuration requires a valid probo.Service instance")
	ErrMissingUsrmgrService = errors.New("server configuration requires a valid usrmgr.Service instance")
)

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	httpserver.RenderJSON(
		w,
		http.StatusMethodNotAllowed,
		map[string]string{
			"error": "method not allowed",
		},
	)
}

func notFound(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	httpserver.RenderJSON(
		w,
		http.StatusNotFound,
		map[string]string{
			"error": "not found",
		},
	)
}

func NewServer(cfg Config) (*Server, error) {
	if cfg.Probo == nil {
		return nil, ErrMissingProboService
	}

	if cfg.Usrmgr == nil {
		return nil, ErrMissingUsrmgrService
	}

	return &Server{
		cfg: cfg,
	}, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	corsOpts := cors.Options{
		AllowedOrigins:     s.cfg.AllowedOrigins,
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "HEAD"},
		AllowedHeaders:     []string{"content-type", "traceparent", "authorization"},
		ExposedHeaders:     []string{"x-Request-id"},
		AllowCredentials:   true,
		MaxAge:             600, // 10 minutes (chrome >= 76 maximum value c.f. https://source.chromium.org/chromium/chromium/src/+/main:services/network/public/cpp/cors/preflight_result.cc;drc=52002151773d8cd9ffc5f557cd7cc880fddcae3e;l=36)
		OptionsPassthrough: false,
		Debug:              false,
	}

	router := chi.NewRouter()
	router.MethodNotAllowed(methodNotAllowed)
	router.NotFound(notFound)

	router.Use(cors.Handler(corsOpts))

	// Mount the console API with authentication
	router.Mount("/console/v1", console_v1.NewMux(s.cfg.Probo, s.cfg.Usrmgr, s.cfg.Auth, s.cfg.ConnectorRegistry, s.cfg.SafeRedirect))

	router.ServeHTTP(w, r)
}

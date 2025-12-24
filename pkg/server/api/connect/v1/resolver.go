//go:generate go run github.com/99designs/gqlgen generate

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

package connect_v1

import (
	"time"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securecookie"
)

type (
	Resolver struct {
		logger       *log.Logger
		iam          *iam.Service
		baseURL      *baseurl.BaseURL
		cookieConfig securecookie.Config
	}
)

func (r *Resolver) sessionCookieConfig(maxAge time.Duration) securecookie.Config {
	return securecookie.Config{
		Name:     r.cookieConfig.Name,
		Secret:   r.cookieConfig.Secret,
		Secure:   r.cookieConfig.Secure,
		HTTPOnly: r.cookieConfig.HTTPOnly,
		SameSite: r.cookieConfig.SameSite,
		Path:     r.cookieConfig.Path,
		Domain:   r.cookieConfig.Domain,
		MaxAge:   int(maxAge.Seconds()),
	}
}

func NewMux(logger *log.Logger, svc *iam.Service, cookieConfig securecookie.Config, baseURL *baseurl.BaseURL) *chi.Mux {
	r := chi.NewMux()

	r.Use(HTTPContextMiddleware)

	sessionMiddleware := NewSessionMiddleware(svc, cookieConfig)
	graphqlHandler := NewGraphQLHandler(svc, logger, baseURL, cookieConfig)
	samlHandler := NewSAMLHandler(svc, cookieConfig, baseURL)

	router := r.With(sessionMiddleware)

	router.Handle("/graphql", graphqlHandler)
	router.Get("/saml/2.0/metadata", samlHandler.MetadataHandler)
	router.Post("/saml/2.0/consume", samlHandler.ConsumeHandler)
	router.Get("/saml/2.0/{samlConfigID}", samlHandler.LoginHandler)

	return r
}

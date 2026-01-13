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
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/authz"
	"go.probo.inc/probo/pkg/server/api/connect/v1/types"
)

type (
	Resolver struct {
		authorize     authz.AuthorizeFunc
		logger        *log.Logger
		iam           *iam.Service
		baseURL       *baseurl.BaseURL
		sessionCookie *authn.Cookie
	}
)

func NewMux(logger *log.Logger, svc *iam.Service, cookieConfig securecookie.Config, tokenSecret string, baseURL *baseurl.BaseURL) *chi.Mux {
	r := chi.NewMux()

	sessionMiddleware := authn.NewSessionMiddleware(svc, cookieConfig)
	apiKeyMiddleware := authn.NewAPIKeyMiddleware(svc, tokenSecret)
	graphqlHandler := NewGraphQLHandler(svc, logger, baseURL, cookieConfig)
	samlHandler := NewSAMLHandler(svc, cookieConfig, baseURL, logger)
	scimHandler := NewSCIMHandler(svc, logger.Named("scim"))

	router := r.With(sessionMiddleware, apiKeyMiddleware)

	router.Handle("/graphql", graphqlHandler)
	router.Get("/saml/2.0/metadata", samlHandler.MetadataHandler)
	router.Post("/saml/2.0/consume", samlHandler.ConsumeHandler)
	router.Get("/saml/2.0/{samlConfigID}", samlHandler.LoginHandler)

	// SCIM 2.0 endpoints - these use their own bearer token authentication
	scimServer := NewSCIMServer(scimHandler)
	r.Mount("/scim/2.0", http.StripPrefix("/scim/2.0", scimHandler.BearerTokenMiddleware(scimServer)))

	return r
}

func (r *Resolver) Permission(ctx context.Context, obj types.Node, action string) (bool, error) {
	return r.authorize(ctx, obj.GetID(), action) == nil, nil
}

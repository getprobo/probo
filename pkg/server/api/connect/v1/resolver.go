// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

//go:generate go run github.com/99designs/gqlgen generate

// Copyright (c) 2025 Probo Inc <hello@probo.com>.
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

package connect_v1

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/filemanager"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/iam/oauth2scope"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/authz"
	"go.probo.inc/probo/pkg/server/api/connect/v1/types"
	"go.probo.inc/probo/pkg/server/gqlutils"
)

type (
	Resolver struct {
		authorize      authz.AuthorizeFunc
		batchAuthorize authz.BatchAuthorizeFunc
		logger         *log.Logger
		iam            *iam.Service
		scopeRegistry  *oauth2scope.Registry
		fileManager    *filemanager.Service
		baseURL        *baseurl.BaseURL
		sessionCookie  *authn.Cookie
	}
)

func NewMux(
	logger *log.Logger,
	svc *iam.Service,
	cookieConfig securecookie.Config,
	tokenSecret string,
	fileManagerSvc *filemanager.Service,
	baseURL *baseurl.BaseURL,
	allowedRedirectHost saferedirect.AllowedHostFunc,
	isTrustCenterDomain IsTrustCenterDomainFunc,
	graphqlLimits gqlutils.Limits,
) *chi.Mux {
	r := chi.NewMux()

	sessionMiddleware := authn.NewSessionMiddleware(svc, cookieConfig)
	apiKeyMiddleware := authn.NewAPIKeyMiddleware(svc, tokenSecret)
	oauth2Middleware := authn.NewOAuth2AccessTokenMiddleware(svc)
	identityPresenceMiddleware := authn.NewIdentityPresenceMiddleware(baseURL)
	graphqlHandler := NewGraphQLHandler(svc, logger, fileManagerSvc, baseURL, cookieConfig, graphqlLimits)
	samlHandler := NewSAMLHandler(svc, cookieConfig, baseURL, logger)
	scimHandler := NewSCIMHandler(svc, logger.Named("scim"))

	router := r.With(
		sessionMiddleware,
		apiKeyMiddleware,
		oauth2Middleware,
	)

	oidcHandler := NewOIDCHandler(svc, cookieConfig, logger, allowedRedirectHost, isTrustCenterDomain)

	router.Handle("/graphql", graphqlHandler)
	router.Get("/saml/2.0/metadata", samlHandler.MetadataHandler)
	router.Post("/saml/2.0/consume", samlHandler.ConsumeHandler)
	router.Get("/saml/2.0/{samlConfigID}", samlHandler.LoginHandler)
	router.Get("/oidc/{provider}/login", oidcHandler.LoginHandler)
	router.Get("/oidc/{provider}/callback", oidcHandler.CallbackHandler)

	// SCIM 2.0 endpoints - these use their own bearer token authentication
	scimServer := NewSCIMServer(scimHandler)
	r.Mount("/scim/2.0", http.StripPrefix("/scim/2.0", scimHandler.BearerTokenMiddleware(scimServer)))

	// OAuth2 / OpenID Connect server endpoints.
	oauth2Handler := NewOAuth2Handler(svc, cookieConfig, baseURL, logger)

	// Public endpoints (no authentication).
	r.Get("/oauth2/jwks", oauth2Handler.JWKSHandler)
	r.Post("/oauth2/token", oauth2Handler.TokenHandler)
	r.Post("/oauth2/device", oauth2Handler.DeviceAuthHandler)

	// Bearer-token authenticated endpoints.
	bearerAuth := r.With(oauth2Handler.BearerTokenMiddleware)
	bearerAuth.Get("/oauth2/userinfo", oauth2Handler.UserInfoHandler)

	// Client-authenticated endpoints.
	clientAuth := r.With(oauth2Handler.ClientAuthMiddleware)
	clientAuth.Post("/oauth2/introspect", oauth2Handler.IntrospectHandler)
	clientAuth.Post("/oauth2/revoke", oauth2Handler.RevokeHandler)

	// Session-authenticated endpoints.
	router.Get("/oauth2/authorize", oauth2Handler.AuthorizeHandler)

	requireIdentity := router.With(identityPresenceMiddleware)
	requireIdentity.Post("/oauth2/register", oauth2Handler.RegisterHandler)

	return r
}

func (r *Resolver) Permission(ctx context.Context, obj types.Node, action string) (bool, error) {
	_, err := r.authorize(ctx, obj.GetID(), action, authz.WithDryRun())
	return err == nil, nil
}

func (r *Resolver) SSOLoginURL(samlConfigID gid.GID) string {
	return r.baseURL.WithPath("/api/connect/v1/saml/2.0/" + samlConfigID.String()).MustString()
}

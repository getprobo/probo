//go:generate go tool github.com/99designs/gqlgen generate

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

package console_v1

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/esign"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/mailman"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/securecookie"
	"go.probo.inc/probo/pkg/server/api/authn"
	"go.probo.inc/probo/pkg/server/api/authz"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
)

type (
	Resolver struct {
		authorize         authz.AuthorizeFunc
		bulkAuthorize     authz.BulkAuthorizeFunc
		probo             *probo.Service
		iam               *iam.Service
		esign             *esign.Service
		mailman           *mailman.Service
		logger            *log.Logger
		customDomainCname string
	}
)

func NewMux(
	logger *log.Logger,
	proboSvc *probo.Service,
	iamSvc *iam.Service,
	esignSvc *esign.Service,
	mailmanSvc *mailman.Service,
	cookieConfig securecookie.Config,
	tokenSecret string,
	connectorRegistry *connector.ConnectorRegistry,
	baseURL *baseurl.BaseURL,
	customDomainCname string,
) *chi.Mux {
	r := chi.NewMux()

	safeRedirect := &saferedirect.SafeRedirect{AllowedHost: baseURL.Host()}

	graphqlHandler := NewGraphQLHandler(iamSvc, proboSvc, esignSvc, mailmanSvc, customDomainCname, logger)

	r.Group(func(r chi.Router) {
		r.Use(authn.NewSessionMiddleware(iamSvc, cookieConfig))
		r.Use(authn.NewAPIKeyMiddleware(iamSvc, tokenSecret))
		r.Use(authn.NewIdentityPresenceMiddleware())

		r.Handle("/graphql", graphqlHandler)

		r.Get("/connectors/initiate", func(w http.ResponseWriter, r *http.Request) {
			provider := r.URL.Query().Get("provider")
			if provider != "SLACK" && provider != "GOOGLE_WORKSPACE" {
				httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("unsupported provider"))
				return
			}

			organizationID, err := gid.ParseGID(r.URL.Query().Get("organization_id"))
			if err != nil {
				panic(fmt.Errorf("cannot parse organization id: %w", err))
			}

			apiKey := authn.APIKeyFromContext(r.Context())
			if apiKey != nil {
				httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("api key authentication cannot be used for this endpoint"))
				return
			}

			identity := authn.IdentityFromContext(r.Context())
			if identity == nil {
				httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("authentication required"))
				return
			}
			session := authn.SessionFromContext(r.Context())
			if session == nil {
				httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("authentication required"))
				return
			}

			if err := iamSvc.Authorizer.Authorize(r.Context(), iam.AuthorizeParams{
				Principal: identity.ID,
				Resource:  organizationID,
				Session:   &session.ID,
				Action:    probo.ActionConnectorInitiate,
			}); err != nil {
				httpserver.RenderError(w, http.StatusForbidden, err)
				return
			}

			redirectURL, err := connectorRegistry.Initiate(r.Context(), provider, organizationID, r)
			if err != nil {
				panic(fmt.Errorf("cannot initiate connector: %w", err))
			}

			// Allow external redirects for OAuth providers
			var oauthSafeRedirect *saferedirect.SafeRedirect
			switch provider {
			case "SLACK":
				oauthSafeRedirect = &saferedirect.SafeRedirect{AllowedHost: "slack.com"}
			case "GOOGLE_WORKSPACE":
				oauthSafeRedirect = &saferedirect.SafeRedirect{AllowedHost: "accounts.google.com"}
			}
			oauthSafeRedirect.Redirect(w, r, redirectURL, "/", http.StatusSeeOther)
		})

		r.Get("/connectors/complete", func(w http.ResponseWriter, r *http.Request) {
			provider := r.URL.Query().Get("provider")
			if provider == "" {
				httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("missing provider parameter"))
				return
			}

			var connectorProvider coredata.ConnectorProvider
			switch provider {
			case "SLACK":
				connectorProvider = coredata.ConnectorProviderSlack
			case "GOOGLE_WORKSPACE":
				connectorProvider = coredata.ConnectorProviderGoogleWorkspace
			default:
				httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("unsupported provider"))
				return
			}

			stateToken := r.URL.Query().Get("state")
			if stateToken == "" {
				httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("missing state parameter"))
				return
			}

			connection, organizationID, continueURL, err := connectorRegistry.Complete(r.Context(), provider, r)
			if err != nil {
				panic(fmt.Errorf("cannot complete connector: %w", err))
			}

			svc := proboSvc.WithTenant(organizationID.TenantID())

			connector, err := svc.Connectors.Create(
				r.Context(),
				probo.CreateConnectorRequest{
					OrganizationID: *organizationID,
					Provider:       connectorProvider,
					Protocol:       coredata.ConnectorProtocol(connection.Type()),
					Connection:     connection,
				},
			)
			if err != nil {
				panic(fmt.Errorf("cannot create or update connector: %w", err))
			}

			// Append connector_id to the redirect URL so frontend can create the bridge
			redirectURL := continueURL
			if redirectURL == "" {
				redirectURL = baseURL.WithPath("/organizations/" + organizationID.String()).MustString()
			}

			parsedURL, err := url.Parse(redirectURL)
			if err != nil {
				logger.ErrorCtx(r.Context(), "cannot parse redirect URL", log.Error(err))
				parsedURL, _ = url.Parse(baseURL.WithPath("/organizations/" + organizationID.String()).MustString())
			}
			q := parsedURL.Query()
			q.Set("connector_id", connector.ID.String())
			parsedURL.RawQuery = q.Encode()

			safeRedirect.Redirect(w, r, parsedURL.String(), "/", http.StatusSeeOther)
		})
	})

	return r
}

func (r *Resolver) ProboService(ctx context.Context, tenantID gid.TenantID) *probo.TenantService {
	return r.probo.WithTenant(tenantID)
}

func (r *Resolver) Permission(ctx context.Context, obj types.Node, action string) (bool, error) {
	return r.authorize(ctx, obj.GetID(), action) == nil, nil
}

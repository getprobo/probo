// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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
	"go.probo.inc/probo/pkg/accessreview"
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
	"go.probo.inc/probo/pkg/server/api/console/v1/dataloader"
	"go.probo.inc/probo/pkg/server/api/console/v1/types"
)

type (
	Resolver struct {
		authorize         authz.AuthorizeFunc
		probo             *probo.Service
		iam               *iam.Service
		esign             *esign.Service
		accessReview      *accessreview.Service
		mailman           *mailman.Service
		connectorRegistry *connector.ConnectorRegistry
		logger            *log.Logger
		customDomainCname string
	}
)

func NewMux(
	logger *log.Logger,
	proboSvc *probo.Service,
	iamSvc *iam.Service,
	esignSvc *esign.Service,
	accessReviewSvc *accessreview.Service,
	mailmanSvc *mailman.Service,
	cookieConfig securecookie.Config,
	tokenSecret string,
	connectorRegistry *connector.ConnectorRegistry,
	baseURL *baseurl.BaseURL,
	customDomainCname string,
) *chi.Mux {
	r := chi.NewMux()

	safeRedirect := saferedirect.New(saferedirect.StaticHosts(baseURL.Host()))

	graphqlHandler := NewGraphQLHandler(iamSvc, proboSvc, esignSvc, accessReviewSvc, mailmanSvc, connectorRegistry, customDomainCname, logger)

	r.Group(func(r chi.Router) {
		r.Use(authn.NewSessionMiddleware(iamSvc, cookieConfig))
		r.Use(authn.NewAPIKeyMiddleware(iamSvc, tokenSecret))
		r.Use(authn.NewIdentityPresenceMiddleware())
		r.Use(dataloader.NewMiddleware(proboSvc, iamSvc))

		r.Handle("/graphql", graphqlHandler)

		r.Get("/connectors/initiate", func(w http.ResponseWriter, r *http.Request) {
			provider := r.URL.Query().Get("provider")
			if provider == "" {
				httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("missing provider parameter"))
				return
			}

			if _, err := connectorRegistry.Get(provider); err != nil {
				httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("unsupported provider: %q", provider))
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

			http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		})

		r.Get("/connectors/complete", handleConnectorComplete(
			logger,
			baseURL,
			proboSvc,
			connectorRegistry,
			safeRedirect,
		))
	})

	return r
}

func handleConnectorComplete(
	logger *log.Logger,
	baseURL *baseurl.BaseURL,
	proboSvc *probo.Service,
	connectorRegistry *connector.ConnectorRegistry,
	safeRedirect *saferedirect.SafeRedirect,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if oauthErr := query.Get("error"); oauthErr != "" {
			handleConnectorOAuth2Error(w, r, logger, baseURL, safeRedirect, query)
			return
		}

		stateToken := query.Get("state")
		if stateToken == "" {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("missing state parameter"))
			return
		}

		provider, err := connector.ExtractProviderFromState(stateToken)
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot extract provider from state: %w", err))
			return
		}

		var connectorProvider coredata.ConnectorProvider
		if err := connectorProvider.Scan(provider); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("unsupported provider: %q", provider))
			return
		}

		connection, state, err := connectorRegistry.CompleteWithState(r.Context(), provider, r)
		if err != nil {
			panic(fmt.Errorf("cannot complete connector: %w", err))
		}

		organizationID, err := gid.ParseGID(state.OrganizationID)
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot parse organization ID from state: %w", err))
			return
		}

		svc := proboSvc.WithTenant(organizationID.TenantID())

		var cnnctr *coredata.Connector

		// If a connector_id was passed in the state, this is a
		// reconnection — update the existing connector's token.
		if state.ConnectorID != "" {
			connectorID, err := gid.ParseGID(state.ConnectorID)
			if err != nil {
				httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot parse connector ID from state: %w", err))
				return
			}

			cnnctr, err = svc.Connectors.Reconnect(r.Context(), connectorID, connection)
			if err != nil {
				panic(fmt.Errorf("cannot reconnect connector: %w", err))
			}
		} else {
			cnnctr, err = svc.Connectors.Create(
				r.Context(),
				probo.CreateConnectorRequest{
					OrganizationID: organizationID,
					Provider:       connectorProvider,
					Protocol:       coredata.ConnectorProtocol(connection.Type()),
					Connection:     connection,
				},
			)
			if err != nil {
				panic(fmt.Errorf("cannot create connector: %w", err))
			}
		}

		redirectURL := state.ContinueURL
		if redirectURL == "" {
			redirectURL = baseURL.WithPath("/organizations/" + organizationID.String()).MustString()
		}

		parsedURL, err := url.Parse(redirectURL)
		if err != nil {
			logger.ErrorCtx(r.Context(), "cannot parse redirect URL", log.Error(err))
			parsedURL, _ = url.Parse(baseURL.WithPath("/organizations/" + organizationID.String()).MustString())
		}
		q := parsedURL.Query()
		q.Set("connector_id", cnnctr.ID.String())
		q.Set("provider", string(connectorProvider))
		parsedURL.RawQuery = q.Encode()

		safeRedirect.Redirect(w, r, parsedURL.String(), "/", http.StatusSeeOther)
	}
}

func handleConnectorOAuth2Error(
	w http.ResponseWriter,
	r *http.Request,
	logger *log.Logger,
	baseURL *baseurl.BaseURL,
	safeRedirect *saferedirect.SafeRedirect,
	query url.Values,
) {
	oauthErr := query.Get("error")
	oauthErrDesc := query.Get("error_description")

	provider := "unknown"
	redirectURL := baseURL.String()
	if stateToken := query.Get("state"); stateToken != "" {
		if payload, err := connector.DecodeOAuth2StatePayload(stateToken); err == nil {
			if payload.Data.Provider != "" {
				provider = payload.Data.Provider
			}
			if payload.Data.ContinueURL != "" {
				redirectURL = payload.Data.ContinueURL
			}
		}
	}

	logger.WarnCtx(r.Context(), "OAuth2 callback returned error",
		log.String("provider", provider),
		log.String("error", oauthErr),
		log.String("error_description", oauthErrDesc),
	)

	parsedURL, _ := url.Parse(redirectURL)
	q := parsedURL.Query()
	q.Set("error", oauthErr)
	if oauthErrDesc != "" {
		q.Set("error_description", oauthErrDesc)
	}
	parsedURL.RawQuery = q.Encode()

	safeRedirect.Redirect(w, r, parsedURL.String(), "/", http.StatusSeeOther)
}

func (r *Resolver) ProboService(ctx context.Context, tenantID gid.TenantID) *probo.TenantService {
	return r.probo.WithTenant(tenantID)
}

func (r *Resolver) Permission(ctx context.Context, obj types.Node, action string) (bool, error) {
	return r.authorize(ctx, obj.GetID(), action, authz.WithDryRun()) == nil, nil
}

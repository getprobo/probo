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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/go-chi/chi/v5"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/saferedirect"
	"go.probo.inc/probo/pkg/securecookie"
	connect_v1 "go.probo.inc/probo/pkg/server/api/connect/v1"
	"go.probo.inc/probo/pkg/server/api/console/v1/schema"
	"go.probo.inc/probo/pkg/server/gqlutils"
	"go.probo.inc/probo/pkg/statelesstoken"
)

type (
	Resolver struct {
		probo             *probo.Service
		iam               *iam.Service
		customDomainCname string
	}
)

func ensureAuthenticated(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	identity := connect_v1.IdentityFromContext(ctx)

	if identity == nil {
		return func(ctx context.Context) *graphql.Response {
			return &graphql.Response{
				Errors: gqlerror.List{
					gqlutils.Unauthorized(),
				},
			}
		}
	}

	return next(ctx)
}

func NewMux(
	logger *log.Logger,
	proboSvc *probo.Service,
	iamSvc *iam.Service,
	cookieConfig securecookie.Config,
	tokenSecret string,
	connectorRegistry *connector.ConnectorRegistry,
	baseURL *baseurl.BaseURL,
	customDomainCname string,
) *chi.Mux {
	r := chi.NewMux()

	safeRedirect := &saferedirect.SafeRedirect{AllowedHost: baseURL.Host()}

	r.Use(connect_v1.NewSessionMiddleware(iamSvc, cookieConfig))
	r.Use(connect_v1.NewAPIKeyMiddleware(iamSvc, tokenSecret))

	config := schema.Config{
		Resolvers: &Resolver{
			probo:             proboSvc,
			iam:               iamSvc,
			customDomainCname: customDomainCname,
		},
	}
	es := schema.NewExecutableSchema(config)
	h := gqlutils.NewHandler(es, logger)
	h.AroundOperations(ensureAuthenticated)

	r.Handle("/graphql", h)

	r.Get(
		"/documents/signing-requests",
		func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "token is required", http.StatusUnauthorized)
				return
			}

			token = strings.TrimPrefix(token, "Bearer ")
			data, err := statelesstoken.ValidateToken[probo.SigningRequestData](tokenSecret, probo.TokenTypeSigningRequest, token)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			svc := proboSvc.WithTenant(data.Data.OrganizationID.TenantID())

			requests, err := svc.Documents.ListSigningRequests(r.Context(), data.Data.OrganizationID, data.Data.PeopleID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(requests)
		},
	)

	r.Get(
		"/documents/signing-requests/{document_version_id}/pdf",
		func(w http.ResponseWriter, r *http.Request) {
			token := r.URL.Query().Get("token")
			if token == "" {
				http.Error(w, "token is required", http.StatusUnauthorized)
				return
			}

			data, err := statelesstoken.ValidateToken[probo.SigningRequestData](tokenSecret, probo.TokenTypeSigningRequest, token)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			documentVersionID, err := gid.ParseGID(chi.URLParam(r, "document_version_id"))
			if err != nil {
				http.Error(w, "invalid document version id", http.StatusBadRequest)
				return
			}

			svc := proboSvc.WithTenant(data.Data.OrganizationID.TenantID())

			// Get the people to get their email for watermark
			people, err := svc.Peoples.Get(r.Context(), data.Data.PeopleID)
			if err != nil {
				http.Error(w, "cannot get user", http.StatusInternalServerError)
				return
			}

			// Generate PDF with watermark
			pdfData, err := svc.Documents.ExportPDF(r.Context(), documentVersionID, probo.ExportPDFOptions{
				WithWatermark:  true,
				WatermarkEmail: &people.PrimaryEmailAddress,
				WithSignatures: false,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			uuid, err := uuid.NewV7()
			if err != nil {
				http.Error(w, "cannot generate uuid", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s.pdf\"", uuid.String()))
			w.WriteHeader(http.StatusOK)
			w.Write(pdfData)
		},
	)

	r.Post(
		"/documents/signing-requests/{document_version_id}/sign",
		func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "token is required", http.StatusUnauthorized)
				return
			}

			token = strings.TrimPrefix(token, "Bearer ")
			data, err := statelesstoken.ValidateToken[probo.SigningRequestData](tokenSecret, probo.TokenTypeSigningRequest, token)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			documentVersionID, err := gid.ParseGID(chi.URLParam(r, "document_version_id"))
			if err != nil {
				http.Error(w, "invalid document version id", http.StatusBadRequest)
				return
			}

			svc := proboSvc.WithTenant(data.Data.OrganizationID.TenantID())

			if err := svc.Documents.SignDocumentVersion(r.Context(), documentVersionID, data.Data.PeopleID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
		},
	)

	r.Get("/connectors/initiate", func(w http.ResponseWriter, r *http.Request) {
		provider := r.URL.Query().Get("provider")
		if provider != "SLACK" {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("unsupported provider"))
			return
		}

		organizationID, err := gid.ParseGID(r.URL.Query().Get("organization_id"))
		if err != nil {
			panic(fmt.Errorf("cannot parse organization id: %w", err))
		}

		apiKey := connect_v1.APIKeyFromContext(r.Context())
		if apiKey != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("api key authentication cannot be used for this endpoint"))
			return
		}

		identity := connect_v1.IdentityFromContext(r.Context())
		if identity == nil {
			httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("authentication required"))
			return
		}

		if err := iamSvc.Authorizer.Authorize(r.Context(), iam.AuthorizeParams{
			Principal: identity.ID,
			Resource:  organizationID,
			Action:    probo.ActionConnectorInitiate,
		}); err != nil {
			httpserver.RenderError(w, http.StatusForbidden, err)
			return
		}

		redirectURL, err := connectorRegistry.Initiate(r.Context(), provider, organizationID, r)
		if err != nil {
			panic(fmt.Errorf("cannot initiate connector: %w", err))
		}

		// Allow external redirects for Slack OAuth only for now
		slackSafeRedirect := &saferedirect.SafeRedirect{AllowedHost: "slack.com"}
		slackSafeRedirect.Redirect(w, r, redirectURL, "/", http.StatusSeeOther)
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
		default:
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("unsupported provider"))
			return
		}

		stateToken := r.URL.Query().Get("state")
		if stateToken == "" {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("missing state parameter"))
			return
		}

		connection, organizationID, err := connectorRegistry.Complete(r.Context(), provider, r)
		if err != nil {
			panic(fmt.Errorf("cannot complete connector: %w", err))
		}

		continueURL := r.URL.Query().Get("continue")

		svc := proboSvc.WithTenant(organizationID.TenantID())

		_, err = svc.Connectors.Create(
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

		if continueURL != "" {
			safeRedirect.Redirect(w, r, continueURL, "/", http.StatusSeeOther)
		} else {
			redirectURL := baseURL.WithPath("/organizations/" + organizationID.String()).MustString()
			safeRedirect.Redirect(w, r, redirectURL, "/", http.StatusSeeOther)
		}
	})

	return r
}

func (r *Resolver) ProboService(ctx context.Context, tenantID gid.TenantID) *probo.TenantService {
	return r.probo.WithTenant(tenantID)
}

func (r *Resolver) MustAuthorize(ctx context.Context, entityID gid.GID, action iam.Action) {
	identity := connect_v1.IdentityFromContext(ctx)

	err := r.iam.Authorizer.Authorize(
		ctx,
		iam.AuthorizeParams{
			Principal: identity.ID,
			Resource:  entityID,
			Action:    action,
		},
	)
	if err != nil {
		panic(err)
	}
}

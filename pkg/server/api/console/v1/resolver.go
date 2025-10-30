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

package console_v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/getprobo/probo/pkg/auth"
	"github.com/getprobo/probo/pkg/authz"
	"github.com/getprobo/probo/pkg/connector"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/probo"
	"github.com/getprobo/probo/pkg/saferedirect"
	"github.com/getprobo/probo/pkg/server/api/console/v1/schema"
	gqlutils "github.com/getprobo/probo/pkg/server/graphql"
	"github.com/getprobo/probo/pkg/server/session"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"github.com/go-chi/chi/v5"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
)

type (
	AuthConfig struct {
		CookieName      string
		CookieDomain    string
		SessionDuration time.Duration
		CookieSecret    string
	}

	Resolver struct {
		proboSvc          *probo.Service
		authSvc           *auth.Service
		authzSvc          *authz.Service
		samlSvc           *auth.SAMLService
		authCfg           AuthConfig
		customDomainCname string
	}

	ctxKey struct{ name string }

	userTenantAccess struct {
		tenantIDs  []gid.TenantID
		authErrors map[gid.TenantID]error
	}
)

var (
	sessionContextKey    = &ctxKey{name: "session"}
	userContextKey       = &ctxKey{name: "user"}
	userTenantContextKey = &ctxKey{name: "user_tenants"}
)

func SessionFromContext(ctx context.Context) *coredata.Session {
	session, _ := ctx.Value(sessionContextKey).(*coredata.Session)
	return session
}

func UserFromContext(ctx context.Context) *coredata.User {
	user, _ := ctx.Value(userContextKey).(*coredata.User)
	return user
}

func NewMux(
	logger *log.Logger,
	proboSvc *probo.Service,
	authSvc *auth.Service,
	authzSvc *authz.Service,
	authCfg AuthConfig,
	connectorRegistry *connector.ConnectorRegistry,
	safeRedirect *saferedirect.SafeRedirect,
	customDomainCname string,
	samlSvc *auth.SAMLService,
) *chi.Mux {
	r := chi.NewMux()

	r.Get(
		"/documents/signing-requests",
		func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "token is required", http.StatusUnauthorized)
				return
			}

			token = strings.TrimPrefix(token, "Bearer ")
			data, err := statelesstoken.ValidateToken[probo.SigningRequestData](authCfg.CookieSecret, probo.TokenTypeSigningRequest, token)
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

			data, err := statelesstoken.ValidateToken[probo.SigningRequestData](authCfg.CookieSecret, probo.TokenTypeSigningRequest, token)
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
			data, err := statelesstoken.ValidateToken[probo.SigningRequestData](authCfg.CookieSecret, probo.TokenTypeSigningRequest, token)
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

	r.Get("/connectors/initiate", WithSession(authSvc, authzSvc, authCfg, func(w http.ResponseWriter, r *http.Request) {
		provider := r.URL.Query().Get("provider")
		if provider != "SLACK" {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("unsupported provider"))
			return
		}

		organizationID, err := gid.ParseGID(r.URL.Query().Get("organization_id"))
		if err != nil {
			panic(fmt.Errorf("cannot parse organization id: %w", err))
		}

		_ = GetTenantService(r.Context(), proboSvc, organizationID.TenantID())

		redirectURL, err := connectorRegistry.Initiate(r.Context(), provider, organizationID, r)
		if err != nil {
			panic(fmt.Errorf("cannot initiate connector: %w", err))
		}

		// Allow external redirects for Slack OAuth only for now
		slackSafeRedirect := &saferedirect.SafeRedirect{AllowedHost: "slack.com"}
		slackSafeRedirect.Redirect(w, r, redirectURL, "/", http.StatusSeeOther)
	}))

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
			redirectURL := fmt.Sprintf("/organizations/%s", organizationID.String())
			safeRedirect.Redirect(w, r, redirectURL, "/", http.StatusSeeOther)
		}
	})

	r.Get("/", playground.Handler("GraphQL", "/api/console/v1/query"))
	r.Post("/query", graphqlHandler(logger, proboSvc, authSvc, authzSvc, samlSvc, authCfg, customDomainCname))

	return r
}

func graphqlHandler(logger *log.Logger, proboSvc *probo.Service, authSvc *auth.Service, authzSvc *authz.Service, samlSvc *auth.SAMLService, authCfg AuthConfig, customDomainCname string) http.HandlerFunc {
	var mb int64 = 1 << 20

	es := schema.NewExecutableSchema(
		schema.Config{
			Resolvers: &Resolver{
				proboSvc:          proboSvc,
				authSvc:           authSvc,
				authzSvc:          authzSvc,
				samlSvc:           samlSvc,
				authCfg:           authCfg,
				customDomainCname: customDomainCname,
			},
		},
	)
	srv := handler.New(es)
	srv.AddTransport(transport.POST{})
	srv.AddTransport(
		transport.MultipartForm{
			MaxMemory:     32 * mb,
			MaxUploadSize: 50 * mb,
		},
	)
	srv.Use(extension.Introspection{})
	srv.Use(gqlutils.NewTracingExtension(logger))
	srv.SetRecoverFunc(gqlutils.RecoverFunc)

	srv.AroundOperations(
		func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
			user := UserFromContext(ctx)

			if user == nil {
				return func(ctx context.Context) *graphql.Response {
					return &graphql.Response{
						Errors: gqlerror.List{
							&gqlerror.Error{
								Message: "authentication required",
								Extensions: map[string]any{
									"code": "UNAUTHENTICATED",
								},
							},
						},
					}
				}
			}

			return next(ctx)
		},
	)

	return WithSession(authSvc, authzSvc, authCfg, srv.ServeHTTP)
}

func WithSession(authSvc *auth.Service, authzSvc *authz.Service, authCfg AuthConfig, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sessionAuthCfg := session.AuthConfig{
			CookieName:   authCfg.CookieName,
			CookieSecret: authCfg.CookieSecret,
		}

		errorHandler := session.ErrorHandler{
			OnCookieError: func(err error) {
				panic(fmt.Errorf("cannot get session: %w", err))
			},
			OnParseError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
				session.ClearCookie(w, authCfg)
			},
			OnSessionError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
				session.ClearCookie(w, authCfg)
			},
			OnUserError: func(w http.ResponseWriter, authCfg session.AuthConfig) {
				session.ClearCookie(w, authCfg)
			},
			OnTenantError: func(err error) {
				panic(fmt.Errorf("cannot list tenants for user: %w", err))
			},
		}

		authResult := session.TryAuth(ctx, w, r, authSvc, authzSvc, sessionAuthCfg, errorHandler)
		if authResult == nil {
			next(w, r)
			return
		}

		ctx = context.WithValue(ctx, sessionContextKey, authResult.Session)
		ctx = context.WithValue(ctx, userContextKey, authResult.User)
		ctx = context.WithValue(ctx, userTenantContextKey, &userTenantAccess{
			tenantIDs:  authResult.TenantIDs,
			authErrors: authResult.AuthErrors,
		})

		next(w, r.WithContext(ctx))

		// Update session after the handler completes
		if _, err := authSvc.UpdateSession(ctx, authResult.Session.ID); err != nil {
			panic(fmt.Errorf("cannot update session: %w", err))
		}
	}
}

func (r *Resolver) ProboService(ctx context.Context, tenantID gid.TenantID) *probo.TenantService {
	return GetTenantService(ctx, r.proboSvc, tenantID)
}

func (r *Resolver) AuthzService(ctx context.Context, tenantID gid.TenantID) *authz.TenantAuthzService {
	return GetTenantAuthzService(ctx, r.authzSvc, tenantID)
}

func (r *Resolver) AuthService(ctx context.Context, tenantID gid.TenantID) *auth.TenantAuthService {
	return GetTenantAuthService(ctx, r.authSvc, tenantID)
}

func UnwrapOmittable[T any](field graphql.Omittable[T]) *T {
	if !field.IsSet() {
		return nil
	}
	value := field.Value()
	return &value
}

func GetTenantService(ctx context.Context, proboSvc *probo.Service, tenantID gid.TenantID) *probo.TenantService {
	validateTenantAccess(ctx, tenantID)
	return proboSvc.WithTenant(tenantID)
}

func GetTenantAuthzService(ctx context.Context, authzSvc *authz.Service, tenantID gid.TenantID) *authz.TenantAuthzService {
	validateTenantAccess(ctx, tenantID)
	return authzSvc.WithTenant(tenantID)
}

func GetTenantAuthService(ctx context.Context, authSvc *auth.Service, tenantID gid.TenantID) *auth.TenantAuthService {
	validateTenantAccess(ctx, tenantID)
	return authSvc.WithTenant(tenantID)
}

func validateTenantAccess(ctx context.Context, tenantID gid.TenantID) {
	access, _ := ctx.Value(userTenantContextKey).(*userTenantAccess)

	if access == nil {
		panic(fmt.Errorf("tenant not found"))
	}

	if !slices.Contains(access.tenantIDs, tenantID) {
		if access.authErrors != nil {
			if authErr := access.authErrors[tenantID]; authErr != nil {
				panic(authErr)
			}
		}

		panic(fmt.Errorf("access denied to tenant"))
	}
}

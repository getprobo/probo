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
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/getprobo/probo/pkg/connector"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/probo"
	"github.com/getprobo/probo/pkg/saferedirect"
	"github.com/getprobo/probo/pkg/server/api/console/v1/schema"
	gqlutils "github.com/getprobo/probo/pkg/server/graphql"
	"github.com/getprobo/probo/pkg/server/session"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"github.com/getprobo/probo/pkg/usrmgr"
	"github.com/go-chi/chi/v5"
	"github.com/vektah/gqlparser/v2/gqlerror"
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
		proboSvc  *probo.Service
		usrmgrSvc *usrmgr.Service
		authCfg   AuthConfig
	}

	ctxKey struct{ name string }
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
	usrmgrSvc *usrmgr.Service,
	authCfg AuthConfig,
	connectorRegistry *connector.ConnectorRegistry,
	safeRedirect *saferedirect.SafeRedirect,
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

	r.Post("/auth/register", SignUpHandler(usrmgrSvc, authCfg))
	r.Post("/auth/login", SignInHandler(usrmgrSvc, authCfg))
	r.Delete("/auth/logout", SignOutHandler(usrmgrSvc, authCfg))
	r.Post("/auth/invitation", InvitationConfirmationHandler(usrmgrSvc, proboSvc, authCfg))
	r.Post("/auth/forget-password", ForgetPasswordHandler(usrmgrSvc, authCfg))
	r.Post("/auth/reset-password", ResetPasswordHandler(usrmgrSvc, authCfg))

	r.Get("/connectors/initiate", WithSession(usrmgrSvc, authCfg, func(w http.ResponseWriter, r *http.Request) {
		connectorID := r.URL.Query().Get("connector_id")
		organizationID, err := gid.ParseGID(r.URL.Query().Get("organization_id"))
		if err != nil {
			panic(fmt.Errorf("failed to parse organization id: %w", err))
		}

		_ = GetTenantService(r.Context(), proboSvc, organizationID.TenantID())

		redirectURL, err := connectorRegistry.Initiate(r.Context(), connectorID, organizationID, r)
		if err != nil {
			panic(fmt.Errorf("cannot initiate connector: %w", err))
		}

		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}))

	r.Get("/connectors/complete", WithSession(usrmgrSvc, authCfg, func(w http.ResponseWriter, r *http.Request) {
		connectorID := r.URL.Query().Get("connector_id")
		organizationID, err := gid.ParseGID(r.URL.Query().Get("organization_id"))
		if err != nil {
			panic(fmt.Errorf("failed to parse organization id: %w", err))
		}

		connection, err := connectorRegistry.Complete(r.Context(), connectorID, organizationID, r)
		if err != nil {
			panic(fmt.Errorf("failed to complete connector: %w", err))
		}

		svc := GetTenantService(r.Context(), proboSvc, organizationID.TenantID())

		_, err = svc.Connectors.CreateOrUpdate(
			r.Context(),
			probo.CreateOrUpdateConnectorRequest{
				OrganizationID: organizationID,
				Name:           connectorID,
				Type:           connector.ProtocolType(connection.Type()),
				Connection:     connection,
			},
		)
		if err != nil {
			panic(fmt.Errorf("failed to create or update connector: %w", err))
		}

		safeRedirect.RedirectFromQuery(w, r, "continue", "/", http.StatusSeeOther)
	}))

	r.Get("/", playground.Handler("GraphQL", "/api/console/v1/query"))
	r.Post("/query", graphqlHandler(logger, proboSvc, usrmgrSvc, authCfg))

	return r
}

func graphqlHandler(logger *log.Logger, proboSvc *probo.Service, usrmgrSvc *usrmgr.Service, authCfg AuthConfig) http.HandlerFunc {
	var mb int64 = 1 << 20

	es := schema.NewExecutableSchema(
		schema.Config{
			Resolvers: &Resolver{
				proboSvc:  proboSvc,
				usrmgrSvc: usrmgrSvc,
				authCfg:   authCfg,
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
	srv.Use(tracingExtension{})
	srv.SetRecoverFunc(gqlutils.RecoverFunc)
	srv.SetErrorPresenter(gqlutils.ErrorPresenter)

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

	return WithSession(usrmgrSvc, authCfg, srv.ServeHTTP)
}

func WithSession(usrmgrSvc *usrmgr.Service, authCfg AuthConfig, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		sessionAuthCfg := session.AuthConfig{
			CookieName:   authCfg.CookieName,
			CookieSecret: authCfg.CookieSecret,
		}

		errorHandler := session.ErrorHandler{
			OnCookieError: func(err error) {
				panic(fmt.Errorf("failed to get session: %w", err))
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
				panic(fmt.Errorf("failed to list tenants for user: %w", err))
			},
		}

		authResult := session.TryAuth(ctx, w, r, usrmgrSvc, sessionAuthCfg, errorHandler)
		if authResult == nil {
			next(w, r)
			return
		}

		ctx = context.WithValue(ctx, sessionContextKey, authResult.Session)
		ctx = context.WithValue(ctx, userContextKey, authResult.User)
		ctx = context.WithValue(ctx, userTenantContextKey, &authResult.TenantIDs)

		next(w, r.WithContext(ctx))

		// Update session after the handler completes
		if err := usrmgrSvc.UpdateSession(ctx, authResult.Session); err != nil {
			panic(fmt.Errorf("failed to update session: %w", err))
		}
	}
}

func (r *Resolver) ProboService(ctx context.Context, tenantID gid.TenantID) *probo.TenantService {
	return GetTenantService(ctx, r.proboSvc, tenantID)
}

func GetTenantService(ctx context.Context, proboSvc *probo.Service, tenantID gid.TenantID) *probo.TenantService {
	tenantIDs, _ := ctx.Value(userTenantContextKey).(*[]gid.TenantID)

	if tenantIDs == nil {
		panic(fmt.Errorf("tenant not found"))
	}

	for _, id := range *tenantIDs {
		if id == tenantID {
			return proboSvc.WithTenant(tenantID)
		}
	}

	panic(fmt.Errorf("tenant not found"))
}

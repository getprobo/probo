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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/getprobo/probo/pkg/connector"
	"github.com/getprobo/probo/pkg/coredata"
	"github.com/getprobo/probo/pkg/crypto/cipher"
	"github.com/getprobo/probo/pkg/gid"
	"github.com/getprobo/probo/pkg/probo"
	"github.com/getprobo/probo/pkg/saferedirect"
	"github.com/getprobo/probo/pkg/securecookie"
	"github.com/getprobo/probo/pkg/server/api/console/v1/schema"
	"github.com/getprobo/probo/pkg/statelesstoken"
	"github.com/getprobo/probo/pkg/trust"
	"github.com/getprobo/probo/pkg/usrmgr"
	"github.com/go-chi/chi/v5"
	"github.com/vektah/gqlparser/v2/gqlerror"
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
		proboSvc       *probo.Service
		usrmgrSvc      *usrmgr.Service
		trustCenterSvc *trust.Service
		authCfg        AuthConfig
	}

	ctxKey struct{ name string }
)

var (
	sessionContextKey     = &ctxKey{name: "session"}
	userContextKey        = &ctxKey{name: "user"}
	userTenantContextKey  = &ctxKey{name: "user_tenants"}
	tokenAccessContextKey = &ctxKey{name: "token_access"}
)

func SessionFromContext(ctx context.Context) *coredata.Session {
	session, _ := ctx.Value(sessionContextKey).(*coredata.Session)
	return session
}

func UserFromContext(ctx context.Context) *coredata.User {
	user, _ := ctx.Value(userContextKey).(*coredata.User)
	return user
}

type TokenAccessData struct {
	TrustCenterID gid.GID
	Email         string
	TenantID      gid.TenantID
	Scope         string
}

const (
	TokenScopeTrustCenterReadOnly = "trust_center_readonly"
)

func TokenAccessFromContext(ctx context.Context) *TokenAccessData {
	tokenAccess, _ := ctx.Value(tokenAccessContextKey).(*TokenAccessData)
	return tokenAccess
}

func NewMux(
	logger *log.Logger,
	proboSvc *probo.Service,
	usrmgrSvc *usrmgr.Service,
	trustSvc *trust.Service,
	authCfg AuthConfig,
	connectorRegistry *connector.ConnectorRegistry,
	safeRedirect *saferedirect.SafeRedirect,
) *chi.Mux {
	r := chi.NewMux()

	encryptionKey := proboSvc.GetEncryptionKey()

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
	r.Post("/trust-center-access/authenticate", authTokenHandler(proboSvc, trustSvc, authCfg, encryptionKey))
	r.Delete("/trust-center-access/logout", trustCenterLogoutHandler(authCfg))

	r.Get("/connectors/initiate", WithSession(usrmgrSvc, proboSvc, authCfg, encryptionKey, func(w http.ResponseWriter, r *http.Request) {
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

	r.Get("/connectors/complete", WithSession(usrmgrSvc, proboSvc, authCfg, encryptionKey, func(w http.ResponseWriter, r *http.Request) {
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
	r.Post("/query", graphqlHandler(logger, proboSvc, usrmgrSvc, trustSvc, authCfg, encryptionKey))

	return r
}

func graphqlHandler(logger *log.Logger, proboSvc *probo.Service, usrmgrSvc *usrmgr.Service, trustSvc *trust.Service, authCfg AuthConfig, encryptionKey cipher.EncryptionKey) http.HandlerFunc {
	var mb int64 = 1 << 20

	es := schema.NewExecutableSchema(
		schema.Config{
			Resolvers: &Resolver{
				proboSvc:       proboSvc,
				usrmgrSvc:      usrmgrSvc,
				trustCenterSvc: trustSvc,
				authCfg:        authCfg,
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
	srv.SetRecoverFunc(func(ctx context.Context, err any) error {
		logger := httpserver.LoggerFromContext(ctx)
		logger.Error("resolver panic", log.Any("error", err), log.Any("stack", string(debug.Stack())))

		return errors.New("internal server error")
	})

	srv.AroundOperations(
		func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
			user := UserFromContext(ctx)
			tokenAccess := TokenAccessFromContext(ctx)

			if user == nil && tokenAccess == nil {
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

	srv.AroundFields(
		func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
			user := UserFromContext(ctx)
			tokenAccess := TokenAccessFromContext(ctx)

			if tokenAccess != nil && user == nil {
				fc := graphql.GetFieldContext(ctx)
				if fc != nil {
					path := fc.Path()
					if len(path) == 1 {
						allowedOperations := map[string]bool{
							"trustCenterBySlug":        true,
							"exportDocumentVersionPDF": true,
						}

						if !allowedOperations[fc.Field.Name] {
							return nil, fmt.Errorf("access denied: trust center access tokens can only query trustCenterBySlug and exportDocumentVersionPDF")
						}
					}
				}
			}

			return next(ctx)
		},
	)

	return WithSession(usrmgrSvc, proboSvc, authCfg, encryptionKey, srv.ServeHTTP)
}

func WithSession(usrmgrSvc *usrmgr.Service, proboSvc *probo.Service, authCfg AuthConfig, encryptionKey cipher.EncryptionKey, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		cookieValue, err := securecookie.Get(r, securecookie.DefaultConfig(
			authCfg.CookieName,
			authCfg.CookieSecret,
		))

		if err == nil {
			sessionID, err := gid.ParseGID(cookieValue)
			if err == nil {
				session, err := usrmgrSvc.GetSession(ctx, sessionID)
				if err == nil {
					user, err := usrmgrSvc.GetUserBySession(ctx, sessionID)
					if err == nil {
						tenantIDs, err := usrmgrSvc.ListTenantsForUserID(ctx, user.ID)
						if err == nil {
							ctx = context.WithValue(ctx, sessionContextKey, session)
							ctx = context.WithValue(ctx, userContextKey, user)
							ctx = context.WithValue(ctx, userTenantContextKey, &tenantIDs)

							next(w, r.WithContext(ctx))

							if err := usrmgrSvc.UpdateSession(ctx, session); err != nil {
								panic(fmt.Errorf("failed to update session: %w", err))
							}
							return
						}
					}
				}
			}

			securecookie.Clear(w, securecookie.DefaultConfig(
				authCfg.CookieName,
				authCfg.CookieSecret,
			))
		}

		tokenCookieValue, err := securecookie.Get(r, securecookie.Config{
			Name:   TokenCookieName,
			Secret: authCfg.CookieSecret,
		})

		if err == nil {
			encryptedData, err := base64.StdEncoding.DecodeString(tokenCookieValue)
			if err == nil {
				decryptedData, err := cipher.Decrypt(encryptedData, encryptionKey)
				if err == nil {
					var tokenData TrustCenterTokenData
					if err := json.Unmarshal(decryptedData, &tokenData); err == nil {
						if time.Now().Before(tokenData.ExpiresAt) {
							tenantSvc := proboSvc.WithTenant(tokenData.TenantID)
							isActive, err := tenantSvc.TrustCenterAccesses.IsAccessActive(ctx, tokenData.TrustCenterID, tokenData.Email)

							if err == nil && isActive {
								tokenAccess := &TokenAccessData{
									TrustCenterID: tokenData.TrustCenterID,
									Email:         tokenData.Email,
									TenantID:      tokenData.TenantID,
									Scope:         tokenData.Scope,
								}

								ctx = context.WithValue(ctx, tokenAccessContextKey, tokenAccess)
								next(w, r.WithContext(ctx))
								return
							} else {
								securecookie.Clear(w, securecookie.Config{
									Name:   TokenCookieName,
									Secret: authCfg.CookieSecret,
								})
							}
						} else {
							securecookie.Clear(w, securecookie.Config{
								Name:   TokenCookieName,
								Secret: authCfg.CookieSecret,
							})
						}
					}
				}
			}
		}

		next(w, r)
	}
}

func (r *Resolver) ProboService(ctx context.Context, tenantID gid.TenantID) *probo.TenantService {
	return GetTenantService(ctx, r.proboSvc, tenantID)
}

// ProboServiceFromContext gets the tenant service from context using either user or token access
func (r *Resolver) ProboServiceFromContext(ctx context.Context) *probo.TenantService {
	user := UserFromContext(ctx)
	if user != nil {
		userTenants, _ := ctx.Value(userTenantContextKey).(*[]gid.TenantID)
		if userTenants != nil && len(*userTenants) > 0 {
			return r.ProboService(ctx, (*userTenants)[0])
		}
	}

	tokenAccess := TokenAccessFromContext(ctx)
	if tokenAccess != nil {
		return r.ProboService(ctx, tokenAccess.TenantID)
	}

	return nil
}

func GetTenantService(ctx context.Context, proboSvc *probo.Service, tenantID gid.TenantID) *probo.TenantService {
	tokenAccess := TokenAccessFromContext(ctx)
	if tokenAccess != nil {
		if tokenAccess.TenantID == tenantID {
			return proboSvc.WithTenant(tenantID)
		}
		panic(fmt.Errorf("tenant not found"))
	}

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

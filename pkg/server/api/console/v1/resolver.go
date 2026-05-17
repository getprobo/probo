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

package console_v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpclient"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/cookiebanner"
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
	"go.probo.inc/probo/pkg/thirdparty"
)

type (
	Resolver struct {
		authorize         authz.AuthorizeFunc
		probo             *probo.Service
		iam               *iam.Service
		esign             *esign.Service
		accessReview      *accessreview.Service
		mailman           *mailman.Service
		cookieBanner      *cookiebanner.Service
		connectorRegistry *connector.ConnectorRegistry
		thirdParty        *thirdparty.Service
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
	cookieBannerSvc *cookiebanner.Service,
	cookieConfig securecookie.Config,
	tokenSecret string,
	connectorRegistry *connector.ConnectorRegistry,
	baseURL *baseurl.BaseURL,
	customDomainCname string,
	thirdPartySvc *thirdparty.Service,
) *chi.Mux {
	r := chi.NewMux()

	safeRedirect := saferedirect.New(saferedirect.StaticHosts(baseURL.Host()))

	graphqlHandler := NewGraphQLHandler(
		iamSvc,
		proboSvc,
		esignSvc,
		accessReviewSvc,
		mailmanSvc,
		cookieBannerSvc,
		connectorRegistry,
		customDomainCname,
		logger,
		thirdPartySvc,
	)

	r.Group(func(r chi.Router) {
		r.Use(authn.NewSessionMiddleware(iamSvc, cookieConfig))
		r.Use(authn.NewAPIKeyMiddleware(iamSvc, tokenSecret))
		r.Use(authn.NewOAuth2AccessTokenMiddleware(iamSvc))
		r.Use(authn.NewIdentityPresenceMiddleware())
		r.Use(dataloader.NewMiddleware(proboSvc, iamSvc, cookieBannerSvc))

		r.Handle("/graphql", graphqlHandler)

		r.Get(
			"/connectors/initiate",
			handleConnectorInitiate(logger, proboSvc, iamSvc, connectorRegistry),
		)

		r.Get(
			"/connectors/complete",
			handleConnectorComplete(
				logger,
				baseURL,
				proboSvc,
				connectorRegistry,
				safeRedirect,
			),
		)
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
			logger.ErrorCtx(r.Context(), "cannot complete connector", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
			return
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

			cnnctr, err = svc.Connectors.Reconnect(
				r.Context(),
				probo.ReconnectConnectorRequest{
					ConnectorID:    connectorID,
					OrganizationID: organizationID,
					Provider:       connectorProvider,
					Connection:     connection,
				},
			)
			if err != nil {
				logger.ErrorCtx(r.Context(), "cannot reconnect connector", log.Error(err))
				httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
				return
			}
		} else {
			createReq := probo.CreateConnectorRequest{
				OrganizationID: organizationID,
				Provider:       connectorProvider,
				Protocol:       coredata.ConnectorProtocol(connection.Type()),
				Connection:     connection,
			}

			// PagerDuty Scoped OAuth surfaces the customer's subdomain as
			// a `subdomain` query parameter on the redirect URL (not in
			// the token response body). Persist it on the connector
			// settings so the driver and name resolver can read it.
			if connectorProvider == coredata.ConnectorProviderPagerDuty {
				subdomain := query.Get("subdomain")
				if subdomain == "" {
					// Fall back to ProviderMetadata for older OAuth flows
					// that may have surfaced the subdomain through the
					// token response body.
					subdomain = state.ProviderMetadata["subdomain"]
				}
				// The subdomain comes from an attacker-influenceable
				// callback parameter; refuse anything that isn't a valid
				// DNS label so it cannot be smuggled into URLs or logs.
				if subdomain != "" && !isValidPagerDutySubdomain(subdomain) {
					logger.WarnCtx(r.Context(), "rejecting invalid pagerduty subdomain",
						log.String("provider", string(connectorProvider)),
					)
					subdomain = ""
				}
				if subdomain != "" {
					createReq.PagerDutySettings = &coredata.PagerDutyConnectorSettings{
						Subdomain: subdomain,
					}
				}
			}

			// Vercel surfaces the customer's team_id as an OAuth callback
			// query parameter (not in the token response body). When the
			// install targets a personal account no team_id is sent — fall
			// back to /v2/user.id as a synthetic TeamID; the v3 members
			// endpoint accepts personal-account UIDs.
			if connectorProvider == coredata.ConnectorProviderVercel {
				teamID := query.Get("team_id")
				if teamID == "" {
					if oauth2Conn, ok := connection.(*connector.OAuth2Connection); ok && oauth2Conn.AccessToken != "" {
						if uid, err := fetchVercelUserID(r.Context(), oauth2Conn.AccessToken); err == nil {
							teamID = uid
						} else {
							logger.WarnCtx(r.Context(), "cannot fetch vercel user id for personal-account fallback", log.Error(err))
						}
					}
				}
				if teamID != "" {
					createReq.VercelSettings = &coredata.VercelConnectorSettings{
						TeamID: teamID,
					}
				}
			}

			cnnctr, err = svc.Connectors.Create(r.Context(), createReq)
			if err != nil {
				logger.ErrorCtx(r.Context(), "cannot create connector", log.Error(err))
				httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
				return
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

	// Provider error_description fields routinely carry PII (user emails,
	// account names) and must never reach logs or the client redirect URL.
	// Forward only the standardized error code.
	logger.WarnCtx(r.Context(), "OAuth2 callback returned error",
		log.String("provider", provider),
		log.String("error", oauthErr),
	)

	parsedURL, _ := url.Parse(redirectURL)
	q := parsedURL.Query()
	q.Set("error", oauthErr)
	parsedURL.RawQuery = q.Encode()

	safeRedirect.Redirect(w, r, parsedURL.String(), "/", http.StatusSeeOther)
}

// fetchVercelUserID calls Vercel's /v2/user with the freshly-minted access
// token to retrieve the user's UID. This is used as a synthetic TeamID
// when the OAuth callback omits team_id (personal-account installs).
func fetchVercelUserID(ctx context.Context, accessToken string) (string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, "https://api.vercel.com/v2/user", nil)
	if err != nil {
		return "", fmt.Errorf("cannot create vercel user request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := httpclient.DefaultClient(httpclient.WithSSRFProtection())
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute vercel user request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch vercel user: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("cannot decode vercel user response: %w", err)
	}

	return body.User.ID, nil
}

// isValidPagerDutySubdomain reports whether s is a single DNS label
// (RFC 1035 §2.3.1). PagerDuty subdomains are tenant identifiers that
// will be embedded in API URLs; the OAuth callback is the only place
// where a malformed value can enter the system.
func isValidPagerDutySubdomain(s string) bool {
	if s == "" || len(s) > 63 {
		return false
	}
	for _, c := range s {
		switch {
		case c >= 'a' && c <= 'z':
		case c >= 'A' && c <= 'Z':
		case c >= '0' && c <= '9':
		case c == '-':
		default:
			return false
		}
	}
	return true
}

func (r *Resolver) ProboService(ctx context.Context, tenantID gid.TenantID) *probo.TenantService {
	return r.probo.WithTenant(tenantID)
}

func (r *Resolver) Permission(ctx context.Context, obj types.Node, action string) (bool, error) {
	return r.authorize(ctx, obj.GetID(), action, authz.WithDryRun()) == nil, nil
}

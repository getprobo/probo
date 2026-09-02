// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package console_v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/accessreview"
	"go.probo.inc/probo/pkg/baseurl"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/saferedirect"
)

func handleConnectorComplete(
	logger *log.Logger,
	baseURL *baseurl.BaseURL,
	proboSvc *probo.Service,
	accessReviewSvc *accessreview.Service,
	connectorRegistry *connector.Registry,
	safeRedirect *saferedirect.SafeRedirect,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if oauthErr := query.Get("error"); oauthErr != "" {
			handleConnectorCallbackError(
				w,
				r,
				logger,
				baseURL,
				connectorRegistry,
				safeRedirect,
				query,
			)

			return
		}

		completion, err := connectorRegistry.CompleteFromState(r.Context(), r)
		if err != nil {
			if redirectToGitHubAppInstall(w, r, logger, connectorRegistry, err) {
				return
			}

			logger.ErrorCtx(r.Context(), "cannot complete connector", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

			return
		}

		rawSettings, ok := connectorRawSettings(
			logger,
			w,
			r,
			completion,
			query,
		)
		if !ok {
			return
		}

		finishConnectorCompletion(
			w,
			r,
			logger,
			baseURL,
			proboSvc,
			accessReviewSvc,
			safeRedirect,
			completion,
			query,
			rawSettings,
		)
	}
}

func handleConnectorGitHubAppComplete(
	logger *log.Logger,
	baseURL *baseurl.BaseURL,
	proboSvc *probo.Service,
	accessReviewSvc *accessreview.Service,
	connectorRegistry *connector.Registry,
	safeRedirect *saferedirect.SafeRedirect,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		if callbackErr := query.Get("error"); callbackErr != "" {
			handleConnectorCallbackError(
				w,
				r,
				logger,
				baseURL,
				connectorRegistry,
				safeRedirect,
				query,
			)

			return
		}

		completion, err := connectorRegistry.CompleteGitHubAppFromRequest(r.Context(), r)
		if err != nil {
			if redirectToGitHubAppInstall(w, r, logger, connectorRegistry, err) {
				return
			}

			logger.ErrorCtx(r.Context(), "cannot complete github app connector", log.Error(err))
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot complete GitHub App installation"))

			return
		}

		rawSettings, ok := githubAppConnectorRawSettings(logger, w, r, completion)
		if !ok {
			return
		}

		finishConnectorCompletion(
			w,
			r,
			logger,
			baseURL,
			proboSvc,
			accessReviewSvc,
			safeRedirect,
			completion,
			r.URL.Query(),
			rawSettings,
		)
	}
}

func connectorRawSettings(
	logger *log.Logger,
	w http.ResponseWriter,
	r *http.Request,
	completion *connector.CompletionState,
	query url.Values,
) (json.RawMessage, bool) {
	if completion.Protocol == connector.ProtocolGitHubApp {
		return githubAppConnectorRawSettings(logger, w, r, completion)
	}

	return oauthConnectorRawSettings(logger, w, r, completion, query)
}

func githubAppConnectorRawSettings(
	logger *log.Logger,
	w http.ResponseWriter,
	r *http.Request,
	completion *connector.CompletionState,
) (json.RawMessage, bool) {
	org := completion.ProviderMetadata[connector.CompletionMetadataGitHubOrganization]

	raw, err := json.Marshal(&coredata.GitHubConnectorSettings{
		Organization: org,
	})
	if err != nil {
		logger.ErrorCtx(r.Context(), "cannot marshal github app settings", log.Error(err))
		httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

		return nil, false
	}

	return raw, true
}

func oauthConnectorRawSettings(
	logger *log.Logger,
	w http.ResponseWriter,
	r *http.Request,
	completion *connector.CompletionState,
	query url.Values,
) (json.RawMessage, bool) {
	var connectorProvider coredata.ConnectorProvider
	if err := connectorProvider.UnmarshalText([]byte(completion.Provider)); err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("unsupported provider: %q", completion.Provider))
		return nil, false
	}

	var rawSettings json.RawMessage

	if connectorProvider == coredata.ConnectorProviderDatadog {
		domain := query.Get("domain")
		if !connector.IsValidDatadogDomain(domain) {
			logger.WarnCtx(r.Context(), "rejecting invalid datadog domain",
				log.String("provider", string(connectorProvider)),
			)
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("invalid domain"))

			return nil, false
		}

		region, _ := connector.DatadogSiteForDomain(domain)

		raw, err := json.Marshal(&coredata.DatadogConnectorSettings{
			Region: region,
			Domain: domain,
		})
		if err != nil {
			logger.ErrorCtx(r.Context(), "cannot marshal datadog settings", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

			return nil, false
		}

		rawSettings = raw
	}

	if connectorProvider == coredata.ConnectorProviderZendesk {
		if !connector.IsValidZendeskSubdomain(completion.Site) {
			logger.WarnCtx(r.Context(), "rejecting invalid zendesk subdomain",
				log.String("provider", string(connectorProvider)),
			)
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("invalid subdomain"))

			return nil, false
		}

		raw, err := json.Marshal(&coredata.ZendeskConnectorSettings{
			Subdomain: completion.Site,
		})
		if err != nil {
			logger.ErrorCtx(r.Context(), "cannot marshal zendesk settings", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

			return nil, false
		}

		rawSettings = raw
	}

	if connectorProvider == coredata.ConnectorProviderPagerDuty {
		subdomain := query.Get("subdomain")
		if subdomain == "" {
			subdomain = completion.ProviderMetadata["subdomain"]
		}

		if subdomain != "" && !isValidPagerDutySubdomain(subdomain) {
			logger.WarnCtx(r.Context(), "rejecting invalid pagerduty subdomain",
				log.String("provider", string(connectorProvider)),
			)
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("invalid subdomain"))

			return nil, false
		}

		if subdomain != "" {
			raw, err := json.Marshal(&coredata.PagerDutyConnectorSettings{
				Subdomain: subdomain,
			})
			if err != nil {
				logger.ErrorCtx(r.Context(), "cannot marshal pagerduty settings", log.Error(err))
				httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

				return nil, false
			}

			rawSettings = raw
		}
	}

	return rawSettings, true
}

// continueRedirectURL parses the OAuth state's continue URL, falling
// back to the organization page when it is absent or unparsable.
func continueRedirectURL(
	ctx context.Context,
	logger *log.Logger,
	baseURL *baseurl.BaseURL,
	continueURL string,
	organizationID gid.GID,
) *url.URL {
	redirectURL := continueURL
	if redirectURL == "" {
		redirectURL = baseURL.WithPath("/organizations/" + organizationID.String()).MustString()
	}

	parsedURL, err := url.Parse(redirectURL)
	if err != nil {
		logger.ErrorCtx(ctx, "cannot parse redirect URL", log.Error(err))

		parsedURL, _ = url.Parse(baseURL.WithPath("/organizations/" + organizationID.String()).MustString())
	}

	return parsedURL
}

func finishConnectorCompletion(
	w http.ResponseWriter,
	r *http.Request,
	logger *log.Logger,
	baseURL *baseurl.BaseURL,
	proboSvc *probo.Service,
	accessReviewSvc *accessreview.Service,
	safeRedirect *saferedirect.SafeRedirect,
	completion *connector.CompletionState,
	query url.Values,
	rawSettings json.RawMessage,
) {
	var connectorProvider coredata.ConnectorProvider
	if err := connectorProvider.UnmarshalText([]byte(completion.Provider)); err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("unsupported provider: %q", completion.Provider))
		return
	}

	connection := completion.Connection

	organizationID, err := gid.ParseGID(completion.OrganizationID)
	if err != nil {
		httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot parse organization ID from state: %w", err))
		return
	}

	scope := coredata.NewScopeFromObjectID(organizationID)

	var cnnctr *coredata.Connector

	if completion.ConnectorID != "" {
		connectorID, err := gid.ParseGID(completion.ConnectorID)
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot parse connector ID from state: %w", err))
			return
		}

		cnnctr, err = proboSvc.Connectors.Reconnect(
			r.Context(),
			scope,
			probo.ReconnectConnectorRequest{
				ConnectorID:    connectorID,
				OrganizationID: organizationID,
				Provider:       connectorProvider,
				Connection:     connection,
				RawSettings:    rawSettings,
			},
		)
		if err != nil {
			logger.ErrorCtx(r.Context(), "cannot reconnect connector", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

			return
		}

		if err := accessReviewSvc.ResetSourceNameSyncForConnector(r.Context(), scope, cnnctr.ID); err != nil {
			logger.WarnCtx(r.Context(), "cannot reset access source name sync after reconnect", log.Error(err))
		}
	} else {
		// The source referencing this connector is created by the
		// console after the redirect; a flow abandoned in between
		// strands the connector row.
		createReq := probo.CreateConnectorRequest{
			OrganizationID: organizationID,
			Provider:       connectorProvider,
			Protocol:       coredata.ConnectorProtocol(connection.Type()),
			Connection:     connection,
		}

		if connectorProvider == coredata.ConnectorProviderVercel {
			teamID := vercelCallbackTeamID(query)
			if teamID == "" {
				if oauth2Conn, ok := connection.(*connector.OAuth2Connection); ok && oauth2Conn.AccessToken != "" {
					if uid, err := connector.FetchVercelUserID(r.Context(), oauth2Conn.AccessToken); err == nil {
						teamID = uid
					} else {
						logger.WarnCtx(r.Context(), "cannot fetch vercel user id for personal-account fallback", log.Error(err))
					}
				}
			}

			if teamID != "" {
				raw, err := json.Marshal(&coredata.VercelConnectorSettings{
					TeamID: teamID,
				})
				if err != nil {
					logger.ErrorCtx(r.Context(), "cannot marshal vercel settings", log.Error(err))
					httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

					return
				}

				createReq.RawSettings = raw
			}
		}

		if rawSettings != nil {
			createReq.RawSettings = rawSettings
		}

		cnnctr, err = proboSvc.Connectors.Create(r.Context(), scope, createReq)
		if err != nil {
			logger.ErrorCtx(r.Context(), "cannot create connector", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

			return
		}
	}

	parsedURL := continueRedirectURL(r.Context(), logger, baseURL, completion.ContinueURL, organizationID)

	q := parsedURL.Query()
	q.Set("connector_id", cnnctr.ID.String())
	q.Set("provider", string(connectorProvider))

	if strings.Contains(completion.ContinueURL, "/access-reviews/connections") {
		missing, err := accessReviewSvc.SourceMissingOAuthScopes(r.Context(), scope, cnnctr.ID)
		if err != nil {
			logger.WarnCtx(r.Context(), "cannot determine missing OAuth scopes after connector callback", log.Error(err))
		} else if len(missing) > 0 {
			q.Set("error", accessreview.NewMissingOAuthScopesError(missing).Error())
		}
	}

	parsedURL.RawQuery = q.Encode()

	safeRedirect.Redirect(w, r, parsedURL.String(), "/", http.StatusSeeOther)
}

func redirectToGitHubAppInstall(
	w http.ResponseWriter,
	r *http.Request,
	logger *log.Logger,
	connectorRegistry *connector.Registry,
	err error,
) bool {
	if !errors.Is(err, connector.ErrGitHubAppInstallationRequired) {
		return false
	}

	installURL, urlErr := connectorRegistry.GitHubAppInstallationURL(r.URL.Query().Get("state"))
	if urlErr != nil {
		logger.ErrorCtx(r.Context(), "cannot build github app installation URL", log.Error(urlErr))
		httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

		return true
	}

	http.Redirect(w, r, installURL, http.StatusSeeOther)

	return true
}

func handleConnectorCallbackError(
	w http.ResponseWriter,
	r *http.Request,
	logger *log.Logger,
	baseURL *baseurl.BaseURL,
	connectorRegistry *connector.Registry,
	safeRedirect *saferedirect.SafeRedirect,
	query url.Values,
) {
	oauthErr := query.Get("error")

	provider := "unknown"
	redirectURL := baseURL.String()

	if stateToken := query.Get("state"); stateToken != "" {
		if connector.IsGitHubAppState(stateToken) {
			if state, err := connectorRegistry.ValidateGitHubAppState(stateToken); err == nil {
				provider = connector.GitHubProvider

				if state.ContinueURL != "" {
					redirectURL = state.ContinueURL
				}
			}
		} else if payload, err := connector.DecodeOAuth2StatePayload(stateToken); err == nil {
			if payload.Data.Provider != "" {
				provider = payload.Data.Provider
			}

			if payload.Data.ContinueURL != "" {
				redirectURL = payload.Data.ContinueURL
			}
		}
	}

	logger.WarnCtx(r.Context(), "connector callback returned error",
		log.String("provider", provider),
		log.String("error", oauthErr),
	)

	parsedURL, _ := url.Parse(redirectURL)
	q := parsedURL.Query()
	q.Set("error", oauthErr)
	parsedURL.RawQuery = q.Encode()

	safeRedirect.Redirect(w, r, parsedURL.String(), "/", http.StatusSeeOther)
}

// vercelCallbackTeamID returns the team identifier from Vercel's OAuth
// callback. Vercel uses the camelCase `teamId` query param (not snake_case
// `team_id`); the name is pinned by a test so it cannot silently regress.
func vercelCallbackTeamID(query url.Values) string {
	return query.Get("teamId")
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

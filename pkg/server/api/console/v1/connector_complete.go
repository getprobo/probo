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
	"encoding/json"
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
	connectorRegistry *connector.ConnectorRegistry,
	safeRedirect *saferedirect.SafeRedirect,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()

		if oauthErr := query.Get("error"); oauthErr != "" {
			handleConnectorOAuth2Error(w, r, logger, baseURL, safeRedirect, query)
			return
		}

		completion, err := connectorRegistry.CompleteOAuth2FromRequest(r.Context(), r)
		if err != nil {
			logger.ErrorCtx(r.Context(), "cannot complete oauth2 connector", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

			return
		}

		rawSettings, ok := oauthConnectorRawSettings(
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
	connectorRegistry *connector.ConnectorRegistry,
	safeRedirect *saferedirect.SafeRedirect,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		completion, err := connectorRegistry.CompleteGitHubAppFromRequest(r.Context(), r)
		if err != nil {
			logger.ErrorCtx(r.Context(), "cannot complete github app connector", log.Error(err))
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot complete GitHub App installation"))

			return
		}

		org := completion.ProviderMetadata[connector.CompletionMetadataGitHubOrganization]

		rawSettings, err := json.Marshal(&coredata.GitHubConnectorSettings{
			Organization: org,
		})
		if err != nil {
			logger.ErrorCtx(r.Context(), "cannot marshal github app settings", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

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

	return rawSettings, true
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
		createReq := probo.CreateConnectorRequest{
			OrganizationID: organizationID,
			Provider:       connectorProvider,
			Protocol:       coredata.ConnectorProtocol(connection.Type()),
			Connection:     connection,
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

				subdomain = ""
			}

			if subdomain != "" {
				raw, err := json.Marshal(&coredata.PagerDutyConnectorSettings{
					Subdomain: subdomain,
				})
				if err != nil {
					logger.ErrorCtx(r.Context(), "cannot marshal pagerduty settings", log.Error(err))
					httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

					return
				}

				createReq.RawSettings = raw
			}
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

	redirectURL := completion.ContinueURL
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

	if completion.Protocol != connector.ProtocolGitHubApp &&
		strings.Contains(completion.ContinueURL, "/access-reviews/connections") {
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

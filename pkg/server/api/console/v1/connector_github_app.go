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

func handleGitHubAppComplete(
	logger *log.Logger,
	baseURL *baseurl.BaseURL,
	proboSvc *probo.Service,
	accessReviewSvc *accessreview.Service,
	connectorRegistry *connector.ConnectorRegistry,
	safeRedirect *saferedirect.SafeRedirect,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connection, state, err := connectorRegistry.CompleteGitHubApp(r.Context(), r)
		if err != nil {
			logger.ErrorCtx(r.Context(), "cannot complete github app connector", log.Error(err))
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot complete GitHub App installation"))
			return
		}

		organizationID, err := gid.ParseGID(state.OrganizationID)
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("invalid organization in state"))
			return
		}

		rawSettings, err := json.Marshal(
			&coredata.GitHubConnectorSettings{
				Organization: state.Organization,
			},
		)
		if err != nil {
			logger.ErrorCtx(r.Context(), "cannot marshal github app settings", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
			return
		}

		scope := coredata.NewScopeFromObjectID(organizationID)
		var cnnctr *coredata.Connector

		if state.ConnectorID != "" {
			connectorID, err := gid.ParseGID(state.ConnectorID)
			if err != nil {
				httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("invalid connector in state"))
				return
			}

			cnnctr, err = proboSvc.Connectors.Reconnect(
				r.Context(),
				scope,
				probo.ReconnectConnectorRequest{
					ConnectorID:    connectorID,
					OrganizationID: organizationID,
					Provider:       coredata.ConnectorProviderGitHub,
					Connection:     connection,
					RawSettings:    rawSettings,
				},
			)
			if err != nil {
				logger.ErrorCtx(r.Context(), "cannot reconnect github app connector", log.Error(err))
				httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
				return
			}

			if err := accessReviewSvc.ResetSourceNameSyncForConnector(
				r.Context(),
				scope,
				cnnctr.ID,
			); err != nil {
				logger.WarnCtx(
					r.Context(),
					"cannot reset access source name sync after github app reconnect",
					log.Error(err),
				)
			}
		} else {
			cnnctr, err = proboSvc.Connectors.Create(
				r.Context(),
				scope,
				probo.CreateConnectorRequest{
					OrganizationID: organizationID,
					Provider:       coredata.ConnectorProviderGitHub,
					Protocol:       coredata.ConnectorProtocolGitHubApp,
					Connection:     connection,
					RawSettings:    rawSettings,
				},
			)
			if err != nil {
				logger.ErrorCtx(r.Context(), "cannot create github app connector", log.Error(err))
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
			logger.ErrorCtx(r.Context(), "cannot parse github app redirect URL", log.Error(err))
			parsedURL, _ = url.Parse(
				baseURL.WithPath("/organizations/" + organizationID.String()).MustString(),
			)
		}

		q := parsedURL.Query()
		q.Set("connector_id", cnnctr.ID.String())
		q.Set("provider", string(coredata.ConnectorProviderGitHub))
		parsedURL.RawQuery = q.Encode()

		safeRedirect.Redirect(w, r, parsedURL.String(), "/", http.StatusSeeOther)
	}
}

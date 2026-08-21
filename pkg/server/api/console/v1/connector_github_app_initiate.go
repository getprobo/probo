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
	"errors"
	"fmt"
	"net/http"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
)

func handleConnectorGitHubAppInitiate(
	logger *log.Logger,
	proboSvc *probo.Service,
	iamSvc *iam.Service,
	connectorRegistry *connector.ConnectorRegistry,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationID, err := gid.ParseGID(r.URL.Query().Get("organization_id"))
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("invalid organization_id parameter"))
			return
		}

		if authn.APIKeyFromContext(r.Context()) != nil {
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

		scope, err := iamSvc.Authorizer.Authorize(
			r.Context(),
			iam.AuthorizeParams{
				Principal: identity.ID,
				Resource:  organizationID,
				Session:   &session.ID,
				Action:    probo.ActionConnectorInitiate,
			},
		)
		if err != nil {
			httpserver.RenderError(w, http.StatusForbidden, err)
			return
		}

		if _, err := connectorRegistry.GetProtocol(connector.GitHubProvider, connector.ProtocolGitHubApp); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("github app is not configured"))
			return
		}

		opts := connector.InitiateOptions{}

		if r.URL.Query().Get("connector_id") != "" {
			existing, err := loadExistingGitHubAppConnector(r, proboSvc, scope)
			if err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot reconnect: connector not found"))
					return
				}

				if errors.Is(err, errInvalidReconnectConnector) {
					httpserver.RenderError(w, http.StatusBadRequest, err)
					return
				}

				logger.ErrorCtx(r.Context(), "cannot look up existing github app connector", log.Error(err))
				httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

				return
			}

			opts.ConnectorID = existing.ID.String()
		}

		redirectURL, err := connectorRegistry.InitiateGitHubApp(
			r.Context(),
			organizationID,
			opts,
			r,
		)
		if err != nil {
			logger.ErrorCtx(r.Context(), "cannot initiate github app connector", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

			return
		}

		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}
}

func loadExistingGitHubAppConnector(
	r *http.Request,
	prb *probo.Service,
	scope coredata.Scoper,
) (*coredata.Connector, error) {
	parsedID, err := gid.ParseGID(r.URL.Query().Get("connector_id"))
	if err != nil {
		return nil, fmt.Errorf("%w: cannot parse connector id: %w", errInvalidReconnectConnector, err)
	}

	found, err := prb.Connectors.GetWithConnection(r.Context(), scope, parsedID)
	if err != nil {
		return nil, err
	}

	if found.Provider != coredata.ConnectorProviderGitHub ||
		found.Protocol != coredata.ConnectorProtocolGitHubApp {
		return nil, fmt.Errorf("%w: connector is not a github app", errInvalidReconnectConnector)
	}

	return found, nil
}

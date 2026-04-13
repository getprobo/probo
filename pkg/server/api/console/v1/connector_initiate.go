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

var errInvalidReconnectConnector = errors.New("invalid reconnect connector")

func handleConnectorInitiate(
	logger *log.Logger,
	proboSvc *probo.Service,
	iamSvc *iam.Service,
	connectorRegistry *connector.ConnectorRegistry,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		if err := iamSvc.Authorizer.Authorize(r.Context(), iam.AuthorizeParams{
			Principal: identity.ID,
			Resource:  organizationID,
			Session:   &session.ID,
			Action:    probo.ActionConnectorInitiate,
		}); err != nil {
			httpserver.RenderError(w, http.StatusForbidden, err)
			return
		}

		requestedScopes := r.URL.Query()["scope"]
		prb := proboSvc.WithTenant(organizationID.TenantID())

		// Look up any existing connector so we can union its stored scopes
		// into the new auth request. Cross-org/provider/protocol mismatches
		// are caught inside Reconnect at callback time; this handler only
		// needs the scope set.
		existing, err := loadExistingConnector(r, prb, organizationID, provider)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("cannot reconnect: connector not found"))
				return
			}
			if errors.Is(err, errInvalidReconnectConnector) {
				httpserver.RenderError(w, http.StatusBadRequest, err)
				return
			}
			logger.ErrorCtx(r.Context(), "cannot look up existing connector", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))
			return
		}

		// Always request the union of (old granted ∪ new requested).
		// Union-not-delta because most providers replace rather than
		// merge. No short-circuit: every reconnect runs the full OAuth
		// flow so revoked or stale tokens are never silently reused.
		opts := connector.InitiateOptions{Scopes: requestedScopes}
		if existing != nil {
			opts.Scopes = connector.UnionScopes(existing.Connection.Scopes(), requestedScopes)
			opts.IncludeGrantedScopes = true
			opts.ConnectorID = existing.ID.String()
		}

		redirectURL, err := connectorRegistry.Initiate(r.Context(), provider, organizationID, opts, r)
		if err != nil {
			panic(fmt.Errorf("cannot initiate connector: %w", err))
		}

		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
	}
}

// loadExistingConnector returns the connector the initiate handler
// should reconnect, or nil if this is a fresh install. An explicit
// `connector_id` query parameter selects a specific row; otherwise the
// handler falls back to the widest-scope (org, provider) row. Callers
// must distinguish ErrResourceNotFound (explicit id not found — 400)
// from nil (no existing row — fresh install path).
func loadExistingConnector(
	r *http.Request,
	prb *probo.TenantService,
	organizationID gid.GID,
	provider string,
) (*coredata.Connector, error) {
	if explicitID := r.URL.Query().Get("connector_id"); explicitID != "" {
		parsedID, err := gid.ParseGID(explicitID)
		if err != nil {
			return nil, fmt.Errorf("%w: cannot parse connector id: %w", errInvalidReconnectConnector, err)
		}

		found, err := prb.Connectors.GetWithConnection(r.Context(), parsedID)
		if err != nil {
			return nil, err
		}
		return found, nil
	}

	found, err := prb.Connectors.GetByOrganizationIDAndProvider(
		r.Context(),
		organizationID,
		coredata.ConnectorProvider(provider),
	)
	if errors.Is(err, coredata.ErrResourceNotFound) {
		return nil, nil
	}
	return found, err
}


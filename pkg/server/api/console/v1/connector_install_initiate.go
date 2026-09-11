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
	"fmt"
	"net/http"

	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/connector"
	"go.probo.inc/probo/pkg/connector/provider"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/iam"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/api/authn"
)

// handleConnectorInstallInitiate sends the customer to the vendor to install
// Probo's app. It is the authenticated leg: the identity it authorizes here is
// signed into the state and re-checked on the callback, which is public.
func handleConnectorInstallInitiate(
	logger *log.Logger,
	iamSvc *iam.Service,
	providerRegistry *provider.Registry,
	installStateKey string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var p coredata.ConnectorProvider
		if err := p.UnmarshalText([]byte(r.URL.Query().Get("provider"))); err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("invalid provider parameter"))
			return
		}

		reg, ok := providerRegistry.Get(p)
		if !ok || !reg.SupportsInstall() {
			httpserver.RenderError(w, http.StatusNotFound, fmt.Errorf("connector install is not available"))
			return
		}

		// A deployment with no vendor app configured has no ceremony to offer,
		// and the callback it would redirect to answers 404 for the same reason.
		if !providerRegistry.ManagedConnectorReady(p) {
			httpserver.RenderError(w, http.StatusNotFound, fmt.Errorf("connector is disabled"))
			return
		}

		organizationID, err := gid.ParseGID(r.URL.Query().Get("organization_id"))
		if err != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("invalid organization_id parameter"))
			return
		}

		// The flow ends at a browser redirect carrying a session cookie, so an
		// API key can never complete what it starts here.
		if authn.APIKeyFromContext(ctx) != nil {
			httpserver.RenderError(w, http.StatusBadRequest, fmt.Errorf("api key authentication cannot be used for this endpoint"))
			return
		}

		identity := authn.IdentityFromContext(ctx)
		if identity == nil {
			httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("authentication required"))
			return
		}

		session := authn.SessionFromContext(ctx)
		if session == nil {
			httpserver.RenderError(w, http.StatusUnauthorized, fmt.Errorf("authentication required"))
			return
		}

		// Authorize through iamSvc, never r.authorize: the latter dereferences
		// the identity with no nil check and panics outside a GraphQL request.
		//
		// The authorizer's own error is logged, never rendered: it reports a
		// storage failure the same way it reports a denial, so rendering it
		// would both answer 403 for an internal fault and hand the caller the
		// policy detail behind the decision.
		if _, err := iamSvc.Authorizer.Authorize(
			ctx,
			iam.AuthorizeParams{
				Principal: identity.ID,
				Resource:  organizationID,
				Session:   &session.ID,
				Action:    probo.ActionConnectorInitiate,
			},
		); err != nil {
			status, rendered := installAuthorizationFailure(err)

			logger.WarnCtx(
				ctx,
				"rejecting unauthorized connector install initiate",
				log.String("provider", string(p)),
				log.Error(err),
			)
			httpserver.RenderError(w, status, rendered)

			return
		}

		// The identity in the state is load-bearing, not audit metadata: the
		// callback rejects a browser whose session belongs to anyone else.
		state, err := connector.NewInstallState(
			installStateKey,
			string(p),
			organizationID.String(),
			identity.ID.String(),
		)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot mint connector install state", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

			return
		}

		installURL, err := providerRegistry.InstallURL(p, state)
		if err != nil {
			logger.ErrorCtx(ctx, "cannot build connector install URL", log.Error(err))
			httpserver.RenderError(w, http.StatusInternalServerError, fmt.Errorf("internal error"))

			return
		}

		http.Redirect(w, r, installURL, http.StatusSeeOther)
	}
}

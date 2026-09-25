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

package identityfederation

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/identityfederation"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/server/httpx"
	"go.probo.inc/probo/pkg/server/jsonx"
)

// documentCacheMaxAge lets the CDN and the cloud provider absorb the traffic, so
// these unauthenticated database reads stay near zero.
const documentCacheMaxAge = 1 * time.Hour

type handler struct {
	logger        *log.Logger
	issuer        *identityfederation.Issuer
	organizations *probo.OrganizationService
}

// discovery serves the OIDC discovery document of one organization's issuer.
func (h *handler) discovery(w http.ResponseWriter, r *http.Request) {
	organizationID, err := gid.ParseGID(chi.URLParam(r, "organizationID"))
	if err != nil || organizationID == gid.Nil || organizationID.EntityType() != coredata.OrganizationEntityType {
		jsonx.RenderNotFound(w, fmt.Errorf("not found"))

		return
	}

	// A mistyped identifier must fail here, when the customer registers the
	// provider, rather than register a provider that can never work.
	exists, err := h.organizations.Exists(r.Context(), coredata.NewNoScope(), organizationID)
	if err != nil {
		h.logger.ErrorCtx(
			r.Context(),
			"cannot resolve organization for identity federation discovery",
			log.String("organization_id", organizationID.String()),
			log.Error(err),
		)
		jsonx.RenderInternalServerError(w)

		return
	}

	if !exists {
		jsonx.RenderNotFound(w, fmt.Errorf("not found"))

		return
	}

	metadata, err := h.issuer.Metadata(organizationID)
	if err != nil {
		h.logger.ErrorCtx(
			r.Context(),
			"cannot build identity federation discovery document",
			log.String("organization_id", organizationID.String()),
			log.Error(err),
		)
		jsonx.RenderInternalServerError(w)

		return
	}

	httpx.PublicCache(w, documentCacheMaxAge)

	httpserver.RenderJSON(w, http.StatusOK, metadata)
}

// jwks serves the published key set. The database is not consulted: the key set
// is identical for every organization, and serving it under a well-formed but
// unknown path discloses nothing. A non-organization identifier is still
// rejected so the path cannot advertise a connector or framework as an issuer.
func (h *handler) jwks(w http.ResponseWriter, r *http.Request) {
	organizationID, err := gid.ParseGID(chi.URLParam(r, "organizationID"))
	if err != nil || organizationID == gid.Nil || organizationID.EntityType() != coredata.OrganizationEntityType {
		jsonx.RenderNotFound(w, fmt.Errorf("not found"))

		return
	}

	httpx.PublicCache(w, documentCacheMaxAge)

	httpserver.RenderJSON(w, http.StatusOK, h.issuer.JWKS())
}

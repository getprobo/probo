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

// Package identityfederation serves the two public documents a cloud provider fetches
// to verify a Probo identity federation token. All routes are mounted at
// /federation with no API version prefix: their paths and formats are defined by
// OIDC Discovery, not by us.
//
//	GET  /federation/{organizationID}/.well-known/openid-configuration
//	GET  /federation/.well-known/openid-configuration/{organizationID}
//	GET  /federation/{organizationID}/jwks
//
// The first discovery path is OpenID Connect Discovery 1.0, which appends the
// well-known suffix to the issuer. The second is the same document at the
// RFC 8414 path, which inserts the suffix between the host and the issuer
// path. Entra fetches that form. AWS and GCP fetch the appended form.
//
// All three are unauthenticated, cacheable, and contain no secrets. There is no
// token endpoint here; tokens are minted in-process.
package identityfederation

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/identityfederation"
	"go.probo.inc/probo/pkg/probo"
)

// NewMux returns the identity federation route tree. It is mounted with the /federation
// prefix stripped, so the routes below are relative to it.
func NewMux(
	logger *log.Logger,
	issuer *identityfederation.Issuer,
	organizations *probo.OrganizationService,
) http.Handler {
	h := &handler{
		logger:        logger,
		issuer:        issuer,
		organizations: organizations,
	}

	router := chi.NewRouter()

	// Registered before the organization wildcard so the static prefix is not
	// captured as an organization id. The document is built from configuration,
	// so this URL and the appended form return the same issuer and jwks_uri.
	router.Get("/.well-known/openid-configuration/{organizationID}", h.discovery)

	router.Route(
		"/{organizationID}",
		func(r chi.Router) {
			r.Get("/.well-known/openid-configuration", h.discovery)
			r.Get("/jwks", h.jwks)
		},
	)

	return router
}

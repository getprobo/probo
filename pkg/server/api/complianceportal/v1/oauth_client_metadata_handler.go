// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package complianceportal_v1

import (
	"encoding/json"
	"net/http"

	"go.gearno.de/kit/httpserver"
	portal "go.probo.inc/probo/pkg/complianceportal"
	"go.probo.inc/probo/pkg/server/api/complianceportal"
)

type oauthClientMetadataHandler struct{}

func NewOAuthClientMetadataHandler() http.Handler {
	return &oauthClientMetadataHandler{}
}

func (h *oauthClientMetadataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	compliancePage := complianceportal.CompliancePageFromContext(r.Context())
	baseURL := complianceportal.CompliancePageBaseURLFromContext(r.Context())

	if compliancePage == nil || baseURL == nil {
		httpserver.RenderError(w, http.StatusNotFound, errNotFound)
		return
	}

	doc, err := portal.BuildClientMetadataDocument(compliancePage, *baseURL)
	if err != nil {
		httpserver.RenderError(w, http.StatusInternalServerError, errInternal)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=300")
	_ = json.NewEncoder(w).Encode(doc)
}

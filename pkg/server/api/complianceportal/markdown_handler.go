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

package complianceportal

import (
	"net/http"

	"go.probo.inc/probo/pkg/complianceportal/visitor"
	"go.probo.inc/probo/pkg/coredata"
)

type Handler struct {
	trustService *visitor.Service
}

func NewHandler(trustService *visitor.Service) *Handler {
	return &Handler{trustService: trustService}
}

func (h *Handler) HandleLLMsTxt(w http.ResponseWriter, r *http.Request) {
	tc := CompliancePageFromContext(r.Context())
	if tc == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	scope := coredata.NewScopeFromObjectID(tc.ID)

	if err := h.trustService.RenderCompliancePageMarkdown(r.Context(), w, tc.ID, scope); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) HandleRobotsTxt(w http.ResponseWriter, r *http.Request) {
	tc := CompliancePageFromContext(r.Context())
	if tc == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	baseURL := CompliancePageBaseURLFromContext(r.Context())
	if baseURL == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if err := h.trustService.RenderRobotsTxt(r.Context(), w, tc.SearchEngineIndexing, *baseURL); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) HandleSitemap(w http.ResponseWriter, r *http.Request) {
	tc := CompliancePageFromContext(r.Context())
	if tc == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	baseURL := CompliancePageBaseURLFromContext(r.Context())
	if baseURL == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	scope := coredata.NewScopeFromObjectID(tc.ID)

	if err := h.trustService.RenderSitemap(r.Context(), w, tc.ID, scope, *baseURL); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}

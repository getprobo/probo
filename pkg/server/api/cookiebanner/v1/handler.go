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

package v1

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	cookiebanner "go.probo.inc/probo/packages/cookie-banner"
	"go.probo.inc/probo/pkg/consent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/cookiebanner/v1/types"
)

type handler struct {
	consentService *consent.Service
	logger         *log.Logger
}

func NewMux(logger *log.Logger, consentService *consent.Service) http.Handler {
	h := &handler{
		consentService: consentService,
		logger:         logger,
	}

	r := chi.NewRouter()

	r.Use(
		cors.Handler(
			cors.Options{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
				AllowedHeaders:   []string{"content-type"},
				AllowCredentials: false,
				MaxAge:           600,
			},
		),
	)

	r.Get("/widget.js", h.serveWidget)
	r.Get("/{bannerID}/config", h.getConfig)
	r.Post("/{bannerID}/consents", h.recordConsent)

	return r
}

func (h *handler) serveWidget(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("If-None-Match") == cookiebanner.WidgetETag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "public, no-cache")
	w.Header().Set("ETag", cookiebanner.WidgetETag)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(cookiebanner.WidgetBundle)
}

func (h *handler) getConfig(w http.ResponseWriter, r *http.Request) {
	bannerIDStr := chi.URLParam(r, "bannerID")
	bannerID, err := gid.ParseGID(bannerIDStr)
	if err != nil {
		httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid banner ID"})
		return
	}

	config, err := h.consentService.GetPublishedBannerConfig(r.Context(), bannerID)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "cannot get banner config", log.Error(err))
		httpserver.RenderJSON(w, http.StatusNotFound, map[string]string{"error": "banner not found"})
		return
	}

	resp := types.NewConfigResponse(config.Banner, config.Categories)

	w.Header().Set("Cache-Control", "public, max-age=300")
	httpserver.RenderJSON(w, http.StatusOK, resp)
}

func (h *handler) recordConsent(w http.ResponseWriter, r *http.Request) {
	bannerIDStr := chi.URLParam(r, "bannerID")
	bannerID, err := gid.ParseGID(bannerIDStr)
	if err != nil {
		httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid banner ID"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB limit
	if err != nil {
		httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{"error": "cannot read request body"})
		return
	}

	var req struct {
		VisitorID   string          `json:"visitor_id"`
		ConsentData json.RawMessage `json:"consent_data"`
		Action      string          `json:"action"`
	}

	if err := json.Unmarshal(body, &req); err != nil {
		httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
		return
	}

	if req.VisitorID == "" {
		httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{"error": "visitor_id is required"})
		return
	}

	var action coredata.ConsentAction
	if err := action.Scan(req.Action); err != nil {
		httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid action"})
		return
	}

	ipAddress := r.RemoteAddr

	userAgent := r.UserAgent()

	err = h.consentService.RecordConsent(
		r.Context(),
		bannerID,
		consent.RecordConsentRequest{
			VisitorID:   req.VisitorID,
			ConsentData: req.ConsentData,
			Action:      action,
		},
		ipAddress,
		userAgent,
	)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "cannot record consent", log.Error(err))
		httpserver.RenderJSON(w, http.StatusInternalServerError, map[string]string{"error": "cannot record consent"})
		return
	}

	httpserver.RenderJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

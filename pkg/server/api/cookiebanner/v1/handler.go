// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package cookiebanner_v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/clientip"
	"go.probo.inc/probo/pkg/server/jsonutil"
)

type Handler struct {
	logger          *log.Logger
	cookieBannerSvc *cookiebanner.Service
}

func NewMux(
	logger *log.Logger,
	cookieBannerSvc *cookiebanner.Service,
) *chi.Mux {
	h := &Handler{
		logger:          logger,
		cookieBannerSvc: cookieBannerSvc,
	}

	r := chi.NewMux()
	r.Route("/{bannerID}", func(r chi.Router) {
		r.Use(newCORSMiddleware(logger, cookieBannerSvc))
		r.Get("/config", h.handleGetConfig)
		r.Get("/consents/{visitorID}", h.handleGetConsent)
		r.Post("/consents", h.handlePostConsent)
	})

	return r
}

func (h *Handler) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	bannerID, err := gid.ParseGID(chi.URLParam(r, "bannerID"))
	if err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("invalid banner id"))
		return
	}

	config, err := h.cookieBannerSvc.GetActiveBannerConfig(r.Context(), bannerID)
	if err != nil {
		if errors.Is(err, cookiebanner.ErrBannerNotFound) {
			jsonutil.RenderNotFound(w, fmt.Errorf("banner not found"))
			return
		}
		if errors.Is(err, cookiebanner.ErrNoPublishedVersion) {
			jsonutil.RenderNotFound(w, fmt.Errorf("no published version"))
			return
		}
		h.logger.ErrorCtx(r.Context(), "cannot get banner config", log.Error(err))
		jsonutil.RenderInternalServerError(w)
		return
	}

	httpserver.RenderJSON(w, http.StatusOK, config)
}

func (h *Handler) handleGetConsent(w http.ResponseWriter, r *http.Request) {
	bannerID, err := gid.ParseGID(chi.URLParam(r, "bannerID"))
	if err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("invalid banner id"))
		return
	}

	visitorID := chi.URLParam(r, "visitorID")
	if visitorID == "" {
		jsonutil.RenderBadRequest(w, fmt.Errorf("missing visitor id"))
		return
	}

	consent, err := h.cookieBannerSvc.GetVisitorConsent(r.Context(), bannerID, visitorID)
	if err != nil {
		if errors.Is(err, cookiebanner.ErrBannerNotFound) {
			jsonutil.RenderNotFound(w, fmt.Errorf("banner not found"))
			return
		}
		if errors.Is(err, cookiebanner.ErrConsentNotFound) {
			jsonutil.RenderNotFound(w, fmt.Errorf("consent not found"))
			return
		}
		h.logger.ErrorCtx(r.Context(), "cannot get visitor consent", log.Error(err))
		jsonutil.RenderInternalServerError(w)
		return
	}

	httpserver.RenderJSON(w, http.StatusOK, consent)
}

type (
	postConsentBody struct {
		VisitorID   string                       `json:"visitor_id"`
		Version     int                          `json:"version"`
		Action      coredata.CookieConsentAction `json:"action"`
		ConsentData json.RawMessage              `json:"consent_data"`
	}

	postConsentResponse struct {
		ID        string    `json:"id"`
		VisitorID string    `json:"visitor_id"`
		Action    string    `json:"action"`
		CreatedAt time.Time `json:"created_at"`
	}
)

func (h *Handler) handlePostConsent(w http.ResponseWriter, r *http.Request) {
	bannerID, err := gid.ParseGID(chi.URLParam(r, "bannerID"))
	if err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("invalid banner id"))
		return
	}

	var body postConsentBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("invalid request body"))
		return
	}

	ip := clientip.Extract(r)
	ua := r.UserAgent()

	req := cookiebanner.RecordConsentRequest{
		Version:     body.Version,
		VisitorID:   body.VisitorID,
		IPAddress:   &ip,
		UserAgent:   &ua,
		ConsentData: body.ConsentData,
		Action:      body.Action,
	}

	record, err := h.cookieBannerSvc.RecordConsent(r.Context(), bannerID, req)
	if err != nil {
		if errors.Is(err, cookiebanner.ErrBannerNotFound) {
			jsonutil.RenderNotFound(w, fmt.Errorf("banner not found"))
			return
		}
		if errors.Is(err, cookiebanner.ErrVersionNotFound) || errors.Is(err, cookiebanner.ErrVersionNotPublished) {
			jsonutil.RenderBadRequest(w, fmt.Errorf("invalid version"))
			return
		}
		h.logger.ErrorCtx(r.Context(), "cannot record consent", log.Error(err))
		jsonutil.RenderInternalServerError(w)
		return
	}

	httpserver.RenderJSON(
		w,
		http.StatusCreated,
		postConsentResponse{
			ID:        record.ID.String(),
			VisitorID: record.VisitorID,
			Action:    string(record.Action),
			CreatedAt: record.CreatedAt,
		},
	)
}

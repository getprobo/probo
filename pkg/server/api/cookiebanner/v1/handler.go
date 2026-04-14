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
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
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
	r.Use(newCORSMiddleware(logger, cookieBannerSvc))
	r.Get("/{bannerID}/config", h.handleGetConfig)
	r.Get("/{bannerID}/consents/{visitorID}", h.handleGetConsent)
	r.Post("/{bannerID}/consents", h.handlePostConsent)

	return r
}

func (h *Handler) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	bannerID, err := gid.ParseGID(chi.URLParam(r, "bannerID"))
	if err != nil {
		httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid banner id"})
		return
	}

	config, err := h.cookieBannerSvc.GetActiveBannerConfig(r.Context(), bannerID)
	if err != nil {
		if errors.Is(err, cookiebanner.ErrBannerNotFound) {
			httpserver.RenderJSON(w, http.StatusNotFound, map[string]string{"error": "banner not found"})
			return
		}
		if errors.Is(err, cookiebanner.ErrNoPublishedVersion) {
			httpserver.RenderJSON(w, http.StatusNotFound, map[string]string{"error": "no published version"})
			return
		}
		h.logger.ErrorCtx(r.Context(), "cannot get banner config", log.Error(err))
		httpserver.RenderJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	httpserver.RenderJSON(w, http.StatusOK, config)
}

func (h *Handler) handleGetConsent(w http.ResponseWriter, r *http.Request) {
	bannerID, err := gid.ParseGID(chi.URLParam(r, "bannerID"))
	if err != nil {
		httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid banner id"})
		return
	}

	visitorID := chi.URLParam(r, "visitorID")
	if visitorID == "" {
		httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{"error": "missing visitor id"})
		return
	}

	consent, err := h.cookieBannerSvc.GetVisitorConsent(r.Context(), bannerID, visitorID)
	if err != nil {
		if errors.Is(err, cookiebanner.ErrBannerNotFound) {
			httpserver.RenderJSON(w, http.StatusNotFound, map[string]string{"error": "banner not found"})
			return
		}
		if errors.Is(err, cookiebanner.ErrConsentNotFound) {
			httpserver.RenderJSON(w, http.StatusNotFound, map[string]string{"error": "consent not found"})
			return
		}
		h.logger.ErrorCtx(r.Context(), "cannot get visitor consent", log.Error(err))
		httpserver.RenderJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	httpserver.RenderJSON(w, http.StatusOK, consent)
}

type postConsentBody struct {
	VisitorID   string                       `json:"visitor_id"`
	Version     int                          `json:"version"`
	Action      coredata.CookieConsentAction `json:"action"`
	ConsentData json.RawMessage              `json:"consent_data"`
}

func (h *Handler) handlePostConsent(w http.ResponseWriter, r *http.Request) {
	bannerID, err := gid.ParseGID(chi.URLParam(r, "bannerID"))
	if err != nil {
		httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid banner id"})
		return
	}

	var body postConsentBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	ip := clientIP(r)
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
			httpserver.RenderJSON(w, http.StatusNotFound, map[string]string{"error": "banner not found"})
			return
		}
		if errors.Is(err, cookiebanner.ErrVersionNotFound) || errors.Is(err, cookiebanner.ErrVersionNotPublished) {
			httpserver.RenderJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid version"})
			return
		}
		h.logger.ErrorCtx(r.Context(), "cannot record consent", log.Error(err))
		httpserver.RenderJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	httpserver.RenderJSON(w, http.StatusCreated, map[string]string{
		"id":         record.ID.String(),
		"visitor_id": record.VisitorID,
		"action":     string(record.Action),
		"created_at": record.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip, _, err := net.SplitHostPort(xff); err == nil {
			return ip
		}
		return xff
	}

	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}

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
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/cookiebanner"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/geoloc"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/clientip"
	"go.probo.inc/probo/pkg/server/jsonutil"
)

type Handler struct {
	logger          *log.Logger
	cookieBannerSvc *cookiebanner.Service
	geolocSvc       *geoloc.Service
}

func NewMux(
	logger *log.Logger,
	cookieBannerSvc *cookiebanner.Service,
	geolocSvc *geoloc.Service,
) *chi.Mux {
	h := &Handler{
		logger:          logger,
		cookieBannerSvc: cookieBannerSvc,
		geolocSvc:       geolocSvc,
	}

	r := chi.NewMux()
	r.Route("/{bannerID}", func(r chi.Router) {
		r.Use(newCORSMiddleware(logger, cookieBannerSvc))
		r.Get("/config", h.handleGetConfig)
		r.Get("/consents/{visitorID}", h.handleGetConsent)
		r.Post("/consents", h.handlePostConsent)
		r.Post("/detected-cookies", h.handleReportDetectedCookies)
		r.Post("/report", h.handleReportDetectedTrackers)
	})

	return r
}

type configResponse struct {
	*cookiebanner.BannerConfig
	Regulation cookiebanner.Regulation `json:"regulation"`
}

func (h *Handler) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	bannerID, err := gid.ParseGID(chi.URLParam(r, "bannerID"))
	if err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("invalid banner id"))
		return
	}

	lang := r.URL.Query().Get("lang")

	config, err := h.cookieBannerSvc.GetActiveBannerConfig(r.Context(), bannerID, lang)
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

	regulation := h.resolveRegulation(r)
	if cm := cookiebanner.ConsentModeForRegulation(regulation); cm != "" {
		config.ConsentMode = cm
	}

	httpserver.RenderJSON(w, http.StatusOK, configResponse{
		BannerConfig: config,
		Regulation:   regulation,
	})
}

func (h *Handler) resolveRegulation(r *http.Request) cookiebanner.Regulation {
	ip := clientip.Extract(r)

	cc, err := h.geolocSvc.LookupCountry(r.Context(), ip)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "cannot resolve country for IP", log.Error(err))
		return cookiebanner.RegulationNone
	}

	if cc == "" {
		return cookiebanner.RegulationNone
	}

	return cookiebanner.RegulationForCountry(cc)
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
	sdkVersion := r.Header.Get("X-SDK-Version")

	req := cookiebanner.RecordConsentRequest{
		Version:     body.Version,
		VisitorID:   body.VisitorID,
		IPAddress:   &ip,
		UserAgent:   &ua,
		ConsentData: body.ConsentData,
		Action:      body.Action,
		SdkVersion:  sdkVersion,
		Regulation:  h.resolveRegulation(r),
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

type detectedCookieEntry struct {
	Name          string `json:"name"`
	MaxAgeSeconds *int   `json:"max_age_seconds"`
	Source        string `json:"source"`
}

type reportDetectedCookiesBody struct {
	Cookies []detectedCookieEntry `json:"cookies"`
}

const maxDetectedCookiesPerRequest = 100

func (h *Handler) handleReportDetectedCookies(w http.ResponseWriter, r *http.Request) {
	bannerID, err := gid.ParseGID(chi.URLParam(r, "bannerID"))
	if err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("invalid banner id"))
		return
	}

	var body reportDetectedCookiesBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("invalid request body"))
		return
	}

	if len(body.Cookies) == 0 {
		jsonutil.RenderBadRequest(w, fmt.Errorf("cookies list is empty"))
		return
	}

	if len(body.Cookies) > maxDetectedCookiesPerRequest {
		jsonutil.RenderBadRequest(w, fmt.Errorf("too many cookies, maximum is %d", maxDetectedCookiesPerRequest))
		return
	}

	detected := make([]cookiebanner.DetectedCookie, 0, len(body.Cookies))
	for _, c := range body.Cookies {
		name := strings.TrimSpace(c.Name)
		if name == "" {
			continue
		}

		var source coredata.CookieSource
		switch strings.TrimSpace(c.Source) {
		case "pre-existing":
			source = coredata.CookieSourcePreExisting
		default:
			source = coredata.CookieSourceScript
		}

		detected = append(
			detected,
			cookiebanner.DetectedCookie{
				Name:          name,
				MaxAgeSeconds: c.MaxAgeSeconds,
				Source:        source,
			},
		)
	}

	if len(detected) == 0 {
		jsonutil.RenderBadRequest(w, fmt.Errorf("no valid cookie names provided"))
		return
	}

	req := cookiebanner.ReportDetectedCookiesRequest{
		Cookies: detected,
	}

	if err := h.cookieBannerSvc.ReportDetectedCookies(r.Context(), bannerID, req); err != nil {
		if errors.Is(err, cookiebanner.ErrBannerNotFound) {
			jsonutil.RenderNotFound(w, fmt.Errorf("banner not found"))
			return
		}

		h.logger.ErrorCtx(r.Context(), "cannot report detected cookies", log.Error(err))
		jsonutil.RenderInternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type detectedStorageEntry struct {
	Key         string `json:"key"`
	StorageType string `json:"storage_type"`
	ValueSize   *int   `json:"value_size"`
}

type detectedResourceEntry struct {
	Origin       string `json:"origin"`
	ResourceType string `json:"resource_type"`
}

type reportDetectedTrackersBody struct {
	Cookies   []detectedCookieEntry   `json:"cookies"`
	Storage   []detectedStorageEntry  `json:"storage"`
	Resources []detectedResourceEntry `json:"resources"`
}

const maxDetectedTrackersPerRequest = 100

func (h *Handler) handleReportDetectedTrackers(w http.ResponseWriter, r *http.Request) {
	bannerID, err := gid.ParseGID(chi.URLParam(r, "bannerID"))
	if err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("invalid banner id"))
		return
	}

	var body reportDetectedTrackersBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("invalid request body"))
		return
	}

	total := len(body.Cookies) + len(body.Storage) + len(body.Resources)
	if total == 0 {
		jsonutil.RenderBadRequest(w, fmt.Errorf("no items provided"))
		return
	}

	if total > maxDetectedTrackersPerRequest {
		jsonutil.RenderBadRequest(w, fmt.Errorf("too many items, maximum is %d", maxDetectedTrackersPerRequest))
		return
	}

	var req cookiebanner.ReportDetectedTrackersRequest

	for _, c := range body.Cookies {
		name := strings.TrimSpace(c.Name)
		if name == "" {
			continue
		}

		var source coredata.CookieSource
		switch strings.TrimSpace(c.Source) {
		case "pre-existing":
			source = coredata.CookieSourcePreExisting
		default:
			source = coredata.CookieSourceScript
		}

		req.Cookies = append(
			req.Cookies,
			cookiebanner.DetectedCookie{
				Name:          name,
				MaxAgeSeconds: c.MaxAgeSeconds,
				Source:        source,
			},
		)
	}

	for _, s := range body.Storage {
		key := strings.TrimSpace(s.Key)
		if key == "" {
			continue
		}

		var storageType coredata.TrackerType
		switch strings.TrimSpace(s.StorageType) {
		case "local_storage":
			storageType = coredata.TrackerTypeLocalStorage
		case "session_storage":
			storageType = coredata.TrackerTypeSessionStorage
		case "indexed_db":
			storageType = coredata.TrackerTypeIndexedDB
		default:
			continue
		}

		req.Storage = append(
			req.Storage,
			cookiebanner.DetectedStorageItem{
				Key:         key,
				StorageType: storageType,
				ValueSize:   s.ValueSize,
			},
		)
	}

	for _, res := range body.Resources {
		origin := strings.TrimSpace(res.Origin)
		if origin == "" {
			continue
		}

		var resourceType coredata.TrackerType
		switch strings.TrimSpace(res.ResourceType) {
		case "script":
			resourceType = coredata.TrackerTypeScript
		case "iframe":
			resourceType = coredata.TrackerTypeIframe
		default:
			continue
		}

		req.Resources = append(
			req.Resources,
			cookiebanner.DetectedResourceItem{
				Origin:       origin,
				ResourceType: resourceType,
			},
		)
	}

	if len(req.Cookies)+len(req.Storage)+len(req.Resources) == 0 {
		jsonutil.RenderBadRequest(w, fmt.Errorf("no valid items provided"))
		return
	}

	if err := h.cookieBannerSvc.ReportDetectedTrackers(r.Context(), bannerID, req); err != nil {
		if errors.Is(err, cookiebanner.ErrBannerNotFound) {
			jsonutil.RenderNotFound(w, fmt.Errorf("banner not found"))
			return
		}

		h.logger.ErrorCtx(r.Context(), "cannot report detected trackers", log.Error(err))
		jsonutil.RenderInternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

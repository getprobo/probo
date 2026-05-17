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

// Package agent_v1 exposes the REST surface that the probo-agent binary
// uses to enrol, heartbeat, and push device posture results.
//
// All endpoints speak JSON; agents should not need a GraphQL client.
package agent_v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/probo"
)

const (
	// DefaultHeartbeatIntervalSeconds is the cadence the server tells the
	// agent to heartbeat at. Tunable in future via per-org overrides.
	DefaultHeartbeatIntervalSeconds = 300

	// DefaultPostureIntervalSeconds is how often the agent should run the
	// posture check set.
	DefaultPostureIntervalSeconds = 3600
)

// Handler bundles the dependencies for the /api/agent/v1 router.
type Handler struct {
	logger      *log.Logger
	proboSvc    *probo.Service
	agentServer string
}

// NewMux returns a chi router mounted at /api/agent/v1.
func NewMux(logger *log.Logger, proboSvc *probo.Service, agentServer string) *chi.Mux {
	h := &Handler{
		logger:      logger,
		proboSvc:    proboSvc,
		agentServer: agentServer,
	}

	r := chi.NewRouter()
	r.Post("/enroll", h.handleEnroll)

	r.Group(func(r chi.Router) {
		r.Use(h.deviceAuthMiddleware)
		r.Post("/heartbeat", h.handleHeartbeat)
		r.Post("/postures", h.handlePostures)
		r.Post("/unenroll", h.handleUnenroll)
	})

	return r
}

// ------------------------------------------------------------------
// Request/response payloads
// ------------------------------------------------------------------

type (
	enrollRequest struct {
		EnrollmentToken string                  `json:"enrollment_token"`
		HardwareUUID    string                  `json:"hardware_uuid"`
		SerialNumber    *string                 `json:"serial_number,omitempty"`
		Hostname        string                  `json:"hostname"`
		Platform        coredata.DevicePlatform `json:"platform"`
		OSVersion       string                  `json:"os_version"`
		AgentVersion    string                  `json:"agent_version"`
	}

	enrollResponse struct {
		DeviceID         string `json:"device_id"`
		APIKey           string `json:"api_key"`
		HeartbeatSeconds int    `json:"heartbeat_interval_seconds"`
		PostureSeconds   int    `json:"posture_interval_seconds"`
		ServerTime       string `json:"server_time"`
	}

	heartbeatRequest struct {
		AgentVersion string `json:"agent_version,omitempty"`
		Hostname     string `json:"hostname,omitempty"`
		OSVersion    string `json:"os_version,omitempty"`
		UptimeSec    int64  `json:"uptime_seconds,omitempty"`
	}

	heartbeatResponse struct {
		HeartbeatSeconds int    `json:"heartbeat_interval_seconds"`
		PostureSeconds   int    `json:"posture_interval_seconds"`
		ServerTime       string `json:"server_time"`
	}

	postureResult struct {
		CheckKey   string                       `json:"check_key"`
		Status     coredata.DevicePostureStatus `json:"status"`
		Evidence   json.RawMessage              `json:"evidence,omitempty"`
		ObservedAt time.Time                    `json:"observed_at"`
	}

	postureRequest struct {
		Results []postureResult `json:"results"`
	}

	errorBody struct {
		Error string `json:"error"`
	}
)

// ------------------------------------------------------------------
// Handlers
// ------------------------------------------------------------------

func (h *Handler) handleEnroll(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	var req enrollRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		httpserver.RenderJSON(w, http.StatusBadRequest, errorBody{Error: "invalid json"})
		return
	}
	if req.EnrollmentToken == "" || req.HardwareUUID == "" || req.Hostname == "" || req.Platform == "" {
		httpserver.RenderJSON(w, http.StatusBadRequest, errorBody{Error: "missing required field"})
		return
	}
	if !req.Platform.IsValid() {
		httpserver.RenderJSON(w, http.StatusBadRequest, errorBody{Error: "invalid platform"})
		return
	}

	result, err := h.proboSvc.EnrollDevice(r.Context(), probo.EnrollDeviceRequest{
		EnrollmentSecret: req.EnrollmentToken,
		HardwareUUID:     req.HardwareUUID,
		SerialNumber:     req.SerialNumber,
		Hostname:         req.Hostname,
		Platform:         req.Platform,
		OSVersion:        req.OSVersion,
		AgentVersion:     req.AgentVersion,
	})
	if err != nil {
		if errors.Is(err, probo.ErrEnrollmentTokenInvalid) {
			httpserver.RenderJSON(w, http.StatusUnauthorized, errorBody{Error: "enrollment token is invalid"})
			return
		}
		h.logger.ErrorCtx(r.Context(), "cannot enroll device", log.Error(err))
		httpserver.RenderJSON(w, http.StatusInternalServerError, errorBody{Error: "enrollment failed"})
		return
	}

	httpserver.RenderJSON(w, http.StatusOK, enrollResponse{
		DeviceID:         result.Device.ID.String(),
		APIKey:           result.APIKey,
		HeartbeatSeconds: DefaultHeartbeatIntervalSeconds,
		PostureSeconds:   DefaultPostureIntervalSeconds,
		ServerTime:       time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	device := deviceFromContext(r.Context())
	if device == nil {
		httpserver.RenderJSON(w, http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}

	var req heartbeatRequest
	_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<14)).Decode(&req)

	if err := h.proboSvc.RecordHeartbeat(
		r.Context(),
		device.ID,
		req.Hostname,
		req.OSVersion,
		req.AgentVersion,
	); err != nil {
		if errors.Is(err, probo.ErrDeviceRevoked) {
			httpserver.RenderJSON(w, http.StatusUnauthorized, errorBody{Error: "device revoked"})
			return
		}
		h.logger.ErrorCtx(r.Context(), "cannot record heartbeat", log.Error(err))
		httpserver.RenderJSON(w, http.StatusInternalServerError, errorBody{Error: "heartbeat failed"})
		return
	}

	httpserver.RenderJSON(w, http.StatusOK, heartbeatResponse{
		HeartbeatSeconds: DefaultHeartbeatIntervalSeconds,
		PostureSeconds:   DefaultPostureIntervalSeconds,
		ServerTime:       time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *Handler) handlePostures(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	device := deviceFromContext(r.Context())
	if device == nil {
		httpserver.RenderJSON(w, http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}

	var req postureRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		httpserver.RenderJSON(w, http.StatusBadRequest, errorBody{Error: "invalid json"})
		return
	}
	if len(req.Results) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if len(req.Results) > 200 {
		httpserver.RenderJSON(w, http.StatusBadRequest, errorBody{Error: "too many results in one request"})
		return
	}

	results := make([]probo.RecordPostureResult, 0, len(req.Results))
	for _, pr := range req.Results {
		if pr.CheckKey == "" || !pr.Status.IsValid() {
			httpserver.RenderJSON(w, http.StatusBadRequest, errorBody{Error: "invalid posture entry"})
			return
		}
		results = append(results, probo.RecordPostureResult{
			CheckKey:   pr.CheckKey,
			Status:     pr.Status,
			Evidence:   pr.Evidence,
			ObservedAt: pr.ObservedAt,
		})
	}

	if err := h.proboSvc.RecordPostures(r.Context(), device.ID, results); err != nil {
		if errors.Is(err, probo.ErrDeviceRevoked) {
			httpserver.RenderJSON(w, http.StatusUnauthorized, errorBody{Error: "device revoked"})
			return
		}
		h.logger.ErrorCtx(r.Context(), "cannot record postures", log.Error(err))
		httpserver.RenderJSON(w, http.StatusInternalServerError, errorBody{Error: "posture push failed"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleUnenroll(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	device := deviceFromContext(r.Context())
	if device == nil {
		httpserver.RenderJSON(w, http.StatusUnauthorized, errorBody{Error: "unauthorized"})
		return
	}

	if err := h.proboSvc.UnenrollDevice(r.Context(), device.ID); err != nil {
		h.logger.ErrorCtx(r.Context(), "cannot unenroll device", log.Error(err))
		httpserver.RenderJSON(w, http.StatusInternalServerError, errorBody{Error: "unenroll failed"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ------------------------------------------------------------------
// Middleware
// ------------------------------------------------------------------

type ctxKey struct{ name string }

var deviceContextKey = &ctxKey{name: "device"}

func deviceFromContext(ctx context.Context) *coredata.Device {
	v := ctx.Value(deviceContextKey)
	d, _ := v.(*coredata.Device)
	return d
}

func (h *Handler) deviceAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			httpserver.RenderJSON(w, http.StatusUnauthorized, errorBody{Error: "missing authorization"})
			return
		}
		token, err := bearertoken.Parse(auth)
		if err != nil {
			httpserver.RenderJSON(w, http.StatusUnauthorized, errorBody{Error: "invalid bearer token"})
			return
		}

		device, err := h.proboSvc.AuthenticateDevice(r.Context(), token)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) || errors.Is(err, probo.ErrDeviceRevoked) {
				httpserver.RenderJSON(w, http.StatusUnauthorized, errorBody{Error: "unauthorized"})
				return
			}
			h.logger.ErrorCtx(r.Context(), "cannot authenticate device", log.Error(err))
			httpserver.RenderJSON(w, http.StatusInternalServerError, errorBody{Error: "auth failure"})
			return
		}

		ctx := contextWithDevice(r.Context(), device)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

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
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/bearertoken"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/itam"
	"go.probo.inc/probo/pkg/server/api/agent/v1/types"
	"go.probo.inc/probo/pkg/server/jsonutil"
)

type Handler struct {
	logger      *log.Logger
	itamSvc     *itam.Service
	agentServer string
}

func NewMux(logger *log.Logger, itamSvc *itam.Service, agentServer string) *chi.Mux {
	h := &Handler{
		logger:      logger,
		itamSvc:     itamSvc,
		agentServer: agentServer,
	}

	r := chi.NewRouter()
	r.Post("/enroll", h.handleEnroll)

	r.Group(
		func(r chi.Router) {
			r.Use(h.deviceAuthMiddleware)
			r.Post("/heartbeat", h.handleHeartbeat)
			r.Post("/postures", h.handlePostures)
			r.Post("/unenroll", h.handleUnenroll)
		},
	)

	return r
}

func (h *Handler) handleEnroll(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	var req types.EnrollRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&req); err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("cannot decode request body: %w", err))
		return
	}

	if err := req.Validate(); err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	result, err := h.itamSvc.EnrollDevice(
		r.Context(),
		itam.EnrollDeviceRequest{
			EnrollmentSecret: req.EnrollmentToken,
			HardwareUUID:     req.HardwareUUID,
			SerialNumber:     req.SerialNumber,
			Hostname:         req.Hostname,
			Platform:         req.Platform,
			OSVersion:        req.OSVersion,
			AgentVersion:     req.AgentVersion,
		},
	)
	if err != nil {
		if errors.Is(err, itam.ErrEnrollmentTokenInvalid) {
			jsonutil.RenderUnauthorized(w, errors.New("enrollment token is invalid"))
			return
		}

		h.logger.ErrorCtx(r.Context(), "cannot enroll device", log.Error(err))
		jsonutil.RenderInternalServerError(w)
		return
	}

	httpserver.RenderJSON(w, http.StatusOK, types.NewEnrollResponse(result))
}

func (h *Handler) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	dev := deviceFromContext(r.Context())
	if dev == nil {
		jsonutil.RenderUnauthorized(w, errors.New("unauthorized"))
		return
	}

	var req types.HeartbeatRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<14)).Decode(&req); err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("cannot decode request body: %w", err))
		return
	}

	if err := req.Validate(); err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	scope := coredata.NewScopeFromObjectID(dev.ID)

	if err := h.itamSvc.RecordHeartbeat(
		r.Context(),
		scope,
		dev.ID,
		req.Hostname,
		req.OSVersion,
		req.AgentVersion,
	); err != nil {
		if errors.Is(err, itam.ErrDeviceRevoked) {
			jsonutil.RenderUnauthorized(w, errors.New("device revoked"))
			return
		}

		h.logger.ErrorCtx(r.Context(), "cannot record heartbeat", log.Error(err))
		jsonutil.RenderInternalServerError(w)
		return
	}

	httpserver.RenderJSON(w, http.StatusOK, types.NewHeartbeatResponse())
}

func (h *Handler) handlePostures(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	dev := deviceFromContext(r.Context())
	if dev == nil {
		jsonutil.RenderUnauthorized(w, errors.New("unauthorized"))
		return
	}

	var req types.PostureRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("cannot decode request body: %w", err))
		return
	}

	if err := req.Validate(); err != nil {
		jsonutil.RenderBadRequest(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	if len(req.Results) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	results := make([]itam.RecordPostureResult, 0, len(req.Results))
	for _, pr := range req.Results {
		results = append(
			results,
			itam.RecordPostureResult{
				CheckKey:   pr.CheckKey,
				Status:     pr.Status,
				Evidence:   pr.Evidence,
				ObservedAt: pr.ObservedAt,
			},
		)
	}

	scope := coredata.NewScopeFromObjectID(dev.ID)

	if err := h.itamSvc.RecordPostures(r.Context(), scope, dev.ID, results); err != nil {
		if errors.Is(err, itam.ErrDeviceRevoked) {
			jsonutil.RenderUnauthorized(w, errors.New("device revoked"))
			return
		}

		h.logger.ErrorCtx(r.Context(), "cannot record postures", log.Error(err))
		jsonutil.RenderInternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleUnenroll(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	dev := deviceFromContext(r.Context())
	if dev == nil {
		jsonutil.RenderUnauthorized(w, errors.New("unauthorized"))
		return
	}

	scope := coredata.NewScopeFromObjectID(dev.ID)

	if err := h.itamSvc.UnenrollDevice(r.Context(), scope, dev.ID); err != nil {
		h.logger.ErrorCtx(r.Context(), "cannot unenroll device", log.Error(err))
		jsonutil.RenderInternalServerError(w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

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
			jsonutil.RenderUnauthorized(w, errors.New("missing authorization"))
			return
		}
		token, err := bearertoken.Parse(auth)
		if err != nil {
			jsonutil.RenderUnauthorized(w, errors.New("invalid bearer token"))
			return
		}

		dev, err := h.itamSvc.AuthenticateDevice(r.Context(), token)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) || errors.Is(err, itam.ErrDeviceRevoked) {
				jsonutil.RenderUnauthorized(w, errors.New("unauthorized"))
				return
			}
			h.logger.ErrorCtx(r.Context(), "cannot authenticate device", log.Error(err))
			jsonutil.RenderInternalServerError(w)
			return
		}

		ctx := contextWithDevice(r.Context(), dev)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

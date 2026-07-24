// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

// Package agent_v1 exposes the REST surface that the probo-agent binary
// uses to heartbeat and push device posture results.
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
	"go.probo.inc/probo/pkg/server/jsonx"
)

type Handler struct {
	logger  *log.Logger
	itamSvc *itam.Service
}

func NewMux(logger *log.Logger, itamSvc *itam.Service) *chi.Mux {
	h := &Handler{
		logger:  logger,
		itamSvc: itamSvc,
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

func (h *Handler) handleEnroll(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	var req types.EnrollRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<14)).Decode(&req); err != nil {
		jsonx.RenderBadRequest(w, fmt.Errorf("cannot decode request body: %w", err))
		return
	}

	if err := req.Validate(); err != nil {
		jsonx.RenderBadRequest(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	apiKey, err := h.itamSvc.ExchangeEnrollmentToken(r.Context(), req.Token)
	if err != nil {
		switch {
		case errors.Is(err, itam.ErrEnrollmentTokenExpired),
			errors.Is(err, itam.ErrEnrollmentTokenAlreadyUsed),
			errors.Is(err, itam.ErrEnrollmentTokenInvalid):
			jsonx.RenderUnauthorized(w, errors.New("unauthorized"))
		default:
			h.logger.ErrorCtx(r.Context(), "cannot exchange enrollment token", log.Error(err))
			jsonx.RenderInternalServerError(w)
		}

		return
	}

	httpserver.RenderJSON(w, http.StatusOK, types.EnrollResponse{APIKey: apiKey})
}

func (h *Handler) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	dev := deviceFromContext(r.Context())
	if dev == nil {
		jsonx.RenderUnauthorized(w, errors.New("unauthorized"))
		return
	}

	var req types.HeartbeatRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<14)).Decode(&req); err != nil {
		jsonx.RenderBadRequest(w, fmt.Errorf("cannot decode request body: %w", err))
		return
	}

	if err := req.Validate(); err != nil {
		jsonx.RenderBadRequest(w, fmt.Errorf("invalid request body: %w", err))
		return
	}

	scope := coredata.NewScopeFromObjectID(dev.ID)

	device, err := h.itamSvc.RecordHeartbeat(
		r.Context(),
		scope,
		dev.ID,
		itam.RecordHeartbeatRequest{
			HardwareUUID: req.HardwareUUID,
			SerialNumber: req.SerialNumber,
			Hostname:     req.Hostname,
			Platform:     req.Platform,
			OSVersion:    req.OSVersion,
			AgentVersion: req.AgentVersion,
		},
	)
	if err != nil {
		if errors.Is(err, itam.ErrDeviceRevoked) {
			jsonx.RenderUnauthorized(w, errors.New("device revoked"))
			return
		}

		if errors.Is(err, itam.ErrDeviceHardwareConflict) {
			jsonx.RenderBadRequest(w, errors.New("device hardware uuid already enrolled"))
			return
		}

		h.logger.ErrorCtx(r.Context(), "cannot record heartbeat", log.Error(err))
		jsonx.RenderInternalServerError(w)

		return
	}

	httpserver.RenderJSON(w, http.StatusOK, types.NewHeartbeatResponse(device))
}

func (h *Handler) handlePostures(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	dev := deviceFromContext(r.Context())
	if dev == nil {
		jsonx.RenderUnauthorized(w, errors.New("unauthorized"))
		return
	}

	var req types.PostureRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		jsonx.RenderBadRequest(w, fmt.Errorf("cannot decode request body: %w", err))
		return
	}

	if err := req.Validate(); err != nil {
		jsonx.RenderBadRequest(w, fmt.Errorf("invalid request body: %w", err))
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
			jsonx.RenderUnauthorized(w, errors.New("device revoked"))
			return
		}

		h.logger.ErrorCtx(r.Context(), "cannot record postures", log.Error(err))
		jsonx.RenderInternalServerError(w)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleUnenroll(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	dev := deviceFromContext(r.Context())
	if dev == nil {
		jsonx.RenderUnauthorized(w, errors.New("unauthorized"))
		return
	}

	scope := coredata.NewScopeFromObjectID(dev.ID)

	if err := h.itamSvc.UnenrollDevice(r.Context(), scope, dev.ID); err != nil {
		h.logger.ErrorCtx(r.Context(), "cannot unenroll device", log.Error(err))
		jsonx.RenderInternalServerError(w)

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
			jsonx.RenderUnauthorized(w, errors.New("missing authorization"))
			return
		}

		token, err := bearertoken.Parse(auth)
		if err != nil {
			jsonx.RenderUnauthorized(w, errors.New("invalid bearer token"))
			return
		}

		dev, err := h.itamSvc.AuthenticateDevice(r.Context(), token)
		if err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				jsonx.RenderUnauthorized(w, errors.New("unauthorized"))
				return
			}

			h.logger.ErrorCtx(r.Context(), "cannot authenticate device", log.Error(err))
			jsonx.RenderInternalServerError(w)

			return
		}

		ctx := contextWithDevice(r.Context(), dev)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

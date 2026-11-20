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

// Package agreements_v1 provides the data-plane REST API for recording and
// querying clickwrap agreement acceptances.
//
// Authentication: Bearer sk_live_... or sk_test_... secret key.
// All endpoints are scoped to the authenticated site.
package agreements_v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agreement"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/clientip"
	"go.probo.inc/probo/pkg/server/jsonx"
)

type Handler struct {
	logger       *log.Logger
	agreementSvc *agreement.Service
}

func NewMux(
	logger *log.Logger,
	agreementSvc *agreement.Service,
) *chi.Mux {
	h := &Handler{
		logger:       logger,
		agreementSvc: agreementSvc,
	}

	r := chi.NewMux()
	r.Use(h.authenticate)

	r.Post("/acceptances", h.handlePostAcceptances)
	r.Get("/subjects/{externalUserID}/status", h.handleGetSubjectStatus)

	return r
}

// authenticate extracts and verifies the Bearer secret key, then stores the
// authenticated site in the request context.
func (h *Handler) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			jsonx.RenderUnauthorized(w, fmt.Errorf("missing bearer token"))
			return
		}

		key := strings.TrimPrefix(auth, "Bearer ")
		if !strings.HasPrefix(key, "sk_live_") && !strings.HasPrefix(key, "sk_test_") {
			jsonx.RenderUnauthorized(w, fmt.Errorf("invalid secret key format"))
			return
		}

		site, err := h.agreementSvc.AuthenticateSite(r.Context(), key)
		if err != nil {
			if errors.Is(err, agreement.ErrSiteNotFound) {
				jsonx.RenderUnauthorized(w, fmt.Errorf("invalid secret key"))
				return
			}

			h.logger.ErrorCtx(r.Context(), "cannot authenticate agreement site", log.Error(err))
			jsonx.RenderInternalServerError(w)
			return
		}

		ctx := withSite(r.Context(), site)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// POST /acceptances
//
// Body:
//
//	{
//	  "acceptances": [
//	    {
//	      "agreementId": "...",
//	      "externalUserId": "user-123",
//	      "accountId": "acme-corp",        // optional
//	      "email": "alice@example.com",
//	      "consentText": "I agree to the Terms of Service",
//	      "documentVersionId": "...",
//	      "contentSha256": "abc..."        // optional; validated if provided
//	    }
//	  ],
//	  "sdkVersion": "1.0.0",              // optional
//	  "ipAddress": "1.2.3.4",             // optional; server-side override
//	  "userAgent": "Mozilla/5.0 ..."      // optional
//	}
func (h *Handler) handlePostAcceptances(w http.ResponseWriter, r *http.Request) {
	site := siteFromContext(r.Context())

	var body struct {
		Acceptances []struct {
			AgreementID       string  `json:"agreementId"`
			ExternalUserID    string  `json:"externalUserId"`
			AccountID         *string `json:"accountId"`
			Email             string  `json:"email"`
			ConsentText       string  `json:"consentText"`
			DocumentVersionID string  `json:"documentVersionId"`
			ContentSHA256     string  `json:"contentSha256"`
		} `json:"acceptances"`
		SDKVersion *string `json:"sdkVersion"`
		IPAddress  *string `json:"ipAddress"`
		UserAgent  *string `json:"userAgent"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		jsonx.RenderBadRequest(w, fmt.Errorf("invalid request body"))
		return
	}

	if len(body.Acceptances) == 0 {
		jsonx.RenderBadRequest(w, fmt.Errorf("acceptances must not be empty"))
		return
	}

	ipAddr := body.IPAddress
	if ipAddr == nil {
		ip := clientip.Extract(r)
		if ip != "" {
			ipAddr = &ip
		}
	}

	ua := body.UserAgent
	if ua == nil {
		if rawUA := r.Header.Get("User-Agent"); rawUA != "" {
			ua = &rawUA
		}
	}

	var reqs []agreement.RecordAcceptanceRequest
	for i, a := range body.Acceptances {
		agreementID, err := gid.ParseGID(a.AgreementID)
		if err != nil {
			jsonx.RenderBadRequest(w, fmt.Errorf("acceptances[%d].agreementId: invalid GID", i))
			return
		}

		versionID, err := gid.ParseGID(a.DocumentVersionID)
		if err != nil {
			jsonx.RenderBadRequest(w, fmt.Errorf("acceptances[%d].documentVersionId: invalid GID", i))
			return
		}

		reqs = append(reqs, agreement.RecordAcceptanceRequest{
			AgreementID:       agreementID,
			ExternalUserID:    a.ExternalUserID,
			AccountID:         a.AccountID,
			Email:             a.Email,
			IPAddress:         ipAddr,
			UserAgent:         ua,
			SDKVersion:        body.SDKVersion,
			ConsentText:       a.ConsentText,
			DocumentVersionID: versionID,
			ContentSHA256:     a.ContentSHA256,
		})
	}

	results, err := h.agreementSvc.RecordAcceptances(r.Context(), &agreement.RecordAcceptancesRequest{
		Site:        site,
		Acceptances: reqs,
	})
	if err != nil {
		switch {
		case errors.Is(err, agreement.ErrAgreementNotFound):
			jsonx.RenderNotFound(w, fmt.Errorf("agreement not found"))
		case errors.Is(err, agreement.ErrVersionNotPresentable):
			jsonx.RenderBadRequest(w, fmt.Errorf("document version has no PDF yet; try again shortly"))
		case errors.Is(err, agreement.ErrContentHashMismatch):
			jsonx.RenderBadRequest(w, fmt.Errorf("content_sha256 does not match the stored version hash"))
		default:
			h.logger.ErrorCtx(r.Context(), "cannot record agreement acceptances", log.Error(err))
			jsonx.RenderInternalServerError(w)
		}

		return
	}

	type acceptanceResponse struct {
		ID                    string `json:"id"`
		AgreementID           string `json:"agreementId"`
		SubjectID             string `json:"subjectId"`
		DocumentVersionID     string `json:"documentVersionId"`
		ElectronicSignatureID string `json:"electronicSignatureId"`
		AcceptedAt            string `json:"acceptedAt"`
	}

	var out []acceptanceResponse
	for _, result := range results {
		out = append(out, acceptanceResponse{
			ID:                    result.Acceptance.ID.String(),
			AgreementID:           result.Acceptance.AgreementID.String(),
			SubjectID:             result.Acceptance.SubjectID.String(),
			DocumentVersionID:     result.Acceptance.DocumentVersionID.String(),
			ElectronicSignatureID: result.Acceptance.ElectronicSignatureID.String(),
			AcceptedAt:            result.Acceptance.AcceptedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}

	renderJSON(w, http.StatusCreated, map[string]any{"acceptances": out})
}

// GET /subjects/{externalUserID}/status?agreementIds=id1,id2
func (h *Handler) handleGetSubjectStatus(w http.ResponseWriter, r *http.Request) {
	site := siteFromContext(r.Context())
	externalUserID := chi.URLParam(r, "externalUserID")

	rawIDs := r.URL.Query().Get("agreementIds")
	if rawIDs == "" {
		jsonx.RenderBadRequest(w, fmt.Errorf("agreementIds query param is required"))
		return
	}

	var agreementIDs []gid.GID
	for _, rawID := range strings.Split(rawIDs, ",") {
		id, err := gid.ParseGID(strings.TrimSpace(rawID))
		if err != nil {
			jsonx.RenderBadRequest(w, fmt.Errorf("invalid agreement ID: %s", rawID))
			return
		}

		agreementIDs = append(agreementIDs, id)
	}

	statuses, err := h.agreementSvc.GetSubjectStatus(r.Context(), site, externalUserID, agreementIDs)
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "cannot get subject status", log.Error(err))
		jsonx.RenderInternalServerError(w)
		return
	}

	type statusResponse struct {
		AgreementID        string  `json:"agreementId"`
		CurrentVersionID   *string `json:"currentVersionId"`
		HasAccepted        bool    `json:"hasAccepted"`
		AcceptedVersionID  *string `json:"acceptedVersionId,omitempty"`
		RequiresAcceptance bool    `json:"requiresAcceptance"`
	}

	var out []statusResponse
	for _, st := range statuses {
		sr := statusResponse{
			AgreementID:        st.AgreementID.String(),
			HasAccepted:        st.HasAccepted,
			RequiresAcceptance: st.RequiresAcceptance,
		}

		if st.CurrentVersionID != nil {
			s := st.CurrentVersionID.String()
			sr.CurrentVersionID = &s
		}

		if st.AcceptedVersionID != nil {
			s := st.AcceptedVersionID.String()
			sr.AcceptedVersionID = &s
		}

		out = append(out, sr)
	}

	renderJSON(w, http.StatusOK, map[string]any{"statuses": out})
}

func renderJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		return
	}
}

// Context helpers.

type ctxKey struct{ name string }

var siteCtxKey = &ctxKey{"agreement_site"}

func withSite(ctx context.Context, site *coredata.AgreementSite) context.Context {
	return context.WithValue(ctx, siteCtxKey, site)
}

func siteFromContext(ctx context.Context) *coredata.AgreementSite {
	site, _ := ctx.Value(siteCtxKey).(*coredata.AgreementSite)
	return site
}

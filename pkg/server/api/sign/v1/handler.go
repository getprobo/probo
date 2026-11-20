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

package sign_v1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/crypto/uuid"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/geoloc"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/server/api/clientip"
	"go.probo.inc/probo/pkg/sign"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	logger    *log.Logger
	signSvc   *sign.Service
	pgClient  *pg.Client
	geolocSvc *geoloc.Service
}

func NewMux(
	logger *log.Logger,
	signSvc *sign.Service,
	pgClient *pg.Client,
	geolocSvc *geoloc.Service,
) *chi.Mux {
	h := &Handler{
		logger:    logger,
		signSvc:   signSvc,
		pgClient:  pgClient,
		geolocSvc: geolocSvc,
	}

	r := chi.NewMux()

	r.Get("/documents/{signingConfigID}/content", h.handleGetDocumentContent)
	r.Get("/documents/{signingConfigID}/subjects/{externalUserID}/status", h.handleGetSubjectStatus)
	r.Post("/subjects/{externalUserID}/status", h.handleBatchSubjectStatus)
	r.Post("/documents/{signingConfigID}/acceptances", h.handleRecordAcceptance)
	r.Post("/acceptances/batch", h.handleBatchRecordAcceptances)
	r.Post("/acceptances/{acceptanceID}/revoke", h.handleRevokeAcceptance)

	return r
}

// resolveRequestID returns the X-Request-ID from the request or generates one.
func resolveRequestID(r *http.Request) string {
	id := r.Header.Get("X-Request-ID")
	if id == "" {
		id = uuid.MustNewV4().String()
	}

	return id
}

// authResult holds the resolved authorization context for a data plane request.
type authResult struct {
	organizationID gid.GID
	mode           coredata.DocumentSigningConfigMode
	isSecret       bool
	signingConfig  *coredata.DocumentSigningConfig
}

// resolveAuth authenticates a data plane request by its Authorization bearer
// key (publishable or secret). It resolves the signing config for publishable
// keys, and verifies the bcrypt hash for secret keys.
func (h *Handler) resolveAuth(
	ctx context.Context,
	conn pg.Querier,
	r *http.Request,
	signingConfigIDParam string,
) (*authResult, int, string) {
	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		return nil, http.StatusUnauthorized, codeMissingAuthorization
	}

	parts := strings.SplitN(authorization, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return nil, http.StatusUnauthorized, codeMissingAuthorization
	}

	key := parts[1]

	mode, ok := keyModeFromKey(key)
	if !ok {
		return nil, http.StatusUnauthorized, codeInvalidKey
	}

	if isPublishableKey(key) {
		return h.resolvePublishableKey(ctx, conn, key, mode, r, signingConfigIDParam)
	}

	return h.resolveSecretKey(ctx, conn, key, mode, signingConfigIDParam)
}

func isPublishableKey(key string) bool {
	return strings.HasPrefix(key, "pk_live_") || strings.HasPrefix(key, "pk_test_")
}

func keyModeFromKey(key string) (coredata.DocumentSigningConfigMode, bool) {
	switch {
	case strings.HasPrefix(key, "pk_live_"), strings.HasPrefix(key, "sk_live_"):
		return coredata.DocumentSigningConfigModeLive, true
	case strings.HasPrefix(key, "pk_test_"), strings.HasPrefix(key, "sk_test_"):
		return coredata.DocumentSigningConfigModeTest, true
	}

	return "", false
}

func (h *Handler) resolvePublishableKey(
	ctx context.Context,
	conn pg.Querier,
	key string,
	mode coredata.DocumentSigningConfigMode,
	r *http.Request,
	signingConfigIDParam string,
) (*authResult, int, string) {
	var cfg coredata.DocumentSigningConfig
	if err := cfg.LoadByPublishableKey(ctx, conn, key); err != nil {
		return nil, http.StatusUnauthorized, codeInvalidKey
	}

	if signingConfigIDParam != "" {
		reqID, err := gid.ParseGID(signingConfigIDParam)
		if err != nil || reqID != cfg.ID {
			return nil, http.StatusNotFound, codeConfigNotFound
		}
	}

	if cfg.State == coredata.DocumentSigningConfigStateArchived {
		return nil, http.StatusForbidden, codeConfigNotActive
	}

	if cfg.State != coredata.DocumentSigningConfigStateActive {
		return nil, http.StatusForbidden, codeConfigNotActive
	}

	if cfg.Mode == coredata.DocumentSigningConfigModeLive &&
		mode == coredata.DocumentSigningConfigModeTest {
		return nil, http.StatusForbidden, codeTestKeyOnLiveConfig
	}

	origin := r.Header.Get("Origin")
	if origin != "" && !originAllowed(origin, cfg.AllowedOrigins) {
		return nil, http.StatusForbidden, codeOriginNotAllowed
	}

	return &authResult{
		organizationID: cfg.OrganizationID,
		mode:           mode,
		isSecret:       false,
		signingConfig:  &cfg,
	}, 0, ""
}

func (h *Handler) resolveSecretKey(
	ctx context.Context,
	conn pg.Querier,
	key string,
	mode coredata.DocumentSigningConfigMode,
	signingConfigIDParam string,
) (*authResult, int, string) {
	// Secret keys are org-scoped; we need a signing config to derive the org.
	// For requests that have a signingConfigID param, load the config first
	// to get the organizationID, then verify the secret against the org.
	if signingConfigIDParam == "" {
		return nil, http.StatusUnauthorized, codeInvalidKey
	}

	cfgID, err := gid.ParseGID(signingConfigIDParam)
	if err != nil {
		return nil, http.StatusNotFound, codeConfigNotFound
	}

	scope := coredata.NewScopeFromObjectID(cfgID)

	var cfg coredata.DocumentSigningConfig
	if err := cfg.LoadByID(ctx, conn, scope, cfgID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil, http.StatusNotFound, codeConfigNotFound
		}

		return nil, http.StatusUnauthorized, codeInvalidKey
	}

	orgScope := coredata.NewScopeFromObjectID(cfg.OrganizationID)

	var secret coredata.OrganizationSigningSecret
	if err := secret.LoadByOrganizationID(ctx, conn, orgScope, cfg.OrganizationID); err != nil {
		return nil, http.StatusUnauthorized, codeInvalidKey
	}

	if err := verifySecretKey(key, secret.SecretKeyHash); err != nil {
		return nil, http.StatusUnauthorized, codeInvalidKey
	}

	if cfg.Mode == coredata.DocumentSigningConfigModeLive &&
		mode == coredata.DocumentSigningConfigModeTest {
		return nil, http.StatusForbidden, codeTestKeyOnLiveConfig
	}

	return &authResult{
		organizationID: cfg.OrganizationID,
		mode:           mode,
		isSecret:       true,
		signingConfig:  &cfg,
	}, 0, ""
}

func originAllowed(origin string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}

	for _, a := range allowed {
		if strings.EqualFold(a, origin) {
			return true
		}
	}

	return false
}

func verifySecretKey(plaintext, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
}

// ─── Handlers ───────────────────────────────────────────────────────────────

type documentContentResponse struct {
	DocumentVersionID string `json:"document_version_id"`
	Major             int    `json:"major"`
	Minor             int    `json:"minor"`
	ContentHTML       string `json:"content_html"`
	ContentSha256     string `json:"content_sha256"`
	PdfURL            string `json:"pdf_url"`
	ConsentText       string `json:"consent_text"`
	Kind              string `json:"kind"`
}

func (h *Handler) handleGetDocumentContent(w http.ResponseWriter, r *http.Request) {
	reqID := resolveRequestID(r)
	w.Header().Set("X-Request-ID", reqID)

	signingConfigID := chi.URLParam(r, "signingConfigID")

	var (
		auth       *authResult
		statusCode int
		errCode    string
	)

	err := h.pgClient.WithConn(r.Context(), func(ctx context.Context, conn pg.Querier) error {
		var err error
		auth, statusCode, errCode = h.resolveAuth(ctx, conn, r, signingConfigID)

		return err
	})
	if err != nil || auth == nil {
		h.logAndRecordRejection(r, signingConfigID, http.StatusUnauthorized, errCode, reqID)
		renderError(w, statusCode, errCode, errCode, reqID)

		return
	}

	setCORSHeaders(w, r, auth.signingConfig)

	var dv *coredata.DocumentVersion
	err = h.pgClient.WithConn(r.Context(), func(ctx context.Context, conn pg.Querier) error {
		scope := coredata.NewScopeFromObjectID(auth.signingConfig.DocumentID)
		var loadErr error
		dv, loadErr = h.signSvc.GetCurrentPresentableVersion(ctx, conn, scope, auth.signingConfig)

		return loadErr
	})
	if err != nil {
		if errors.Is(err, sign.ErrDocumentNotPresentable) {
			h.logAndRecordRejection(r, signingConfigID, http.StatusNotFound, codeDocumentNotPresentable, reqID)
			renderError(w, http.StatusNotFound, codeDocumentNotPresentable,
				"The document has no published PDF yet.", reqID)
			return
		}

		h.logger.ErrorCtx(r.Context(), "cannot get document content", log.Error(err))
		renderError(w, http.StatusInternalServerError, "internal_error", "Internal server error.", reqID)

		return
	}

	pdfURL := ""

	if dv.FileID != nil {
		url, err := h.signSvc.GeneratePDFURL(r.Context(), *dv.FileID, time.Hour)
		if err != nil {
			h.logger.ErrorCtx(r.Context(), "cannot generate PDF URL", log.Error(err))
		} else {
			pdfURL = url
		}
	}

	consentText := buildConsentText(auth.signingConfig, dv)

	resp := documentContentResponse{
		DocumentVersionID: dv.ID.String(),
		Major:             dv.Major,
		Minor:             dv.Minor,
		ContentHTML:       dv.Content,
		ContentSha256:     ptrStr(dv.ContentSha256),
		PdfURL:            pdfURL,
		ConsentText:       consentText,
		Kind:              auth.signingConfig.Kind.String(),
	}

	httpserver.RenderJSON(w, http.StatusOK, resp)
}

type subjectStatusResponse struct {
	SubjectID           string     `json:"subject_id"`
	ExternalUserID      string     `json:"external_user_id"`
	Status              string     `json:"status"`
	AcceptedVersionID   *string    `json:"accepted_version_id,omitempty"`
	AcceptedAt          *time.Time `json:"accepted_at,omitempty"`
	CurrentVersionID    *string    `json:"current_version_id,omitempty"`
	CurrentVersionMajor int        `json:"current_version_major"`
	CurrentVersionMinor int        `json:"current_version_minor"`
}

func (h *Handler) handleGetSubjectStatus(w http.ResponseWriter, r *http.Request) {
	reqID := resolveRequestID(r)
	w.Header().Set("X-Request-ID", reqID)

	signingConfigID := chi.URLParam(r, "signingConfigID")
	externalUserID := chi.URLParam(r, "externalUserID")

	var (
		auth       *authResult
		statusCode int
		errCode    string
	)

	err := h.pgClient.WithConn(r.Context(), func(ctx context.Context, conn pg.Querier) error {
		auth, statusCode, errCode = h.resolveAuth(ctx, conn, r, signingConfigID)

		return nil
	})
	if err != nil || auth == nil {
		h.logAndRecordRejection(r, signingConfigID, statusCode, errCode, reqID)
		renderError(w, statusCode, errCode, errCode, reqID)
		return
	}

	setCORSHeaders(w, r, auth.signingConfig)

	var resp *subjectStatusResponse

	err = h.pgClient.WithConn(r.Context(), func(ctx context.Context, conn pg.Querier) error {
		scope := coredata.NewScopeFromObjectID(auth.organizationID)

		var subject coredata.ClickwrapSubject
		if err := subject.LoadByExternalUserID(ctx, conn, scope, auth.organizationID, externalUserID); err != nil {
			if errors.Is(err, coredata.ErrResourceNotFound) {
				return sign.ErrSubjectNotFound
			}

			return fmt.Errorf("cannot load subject: %w", err)
		}

		status, err := h.signSvc.ComputeSubjectStatus(ctx, conn, scope, auth.signingConfig, &subject)
		if err != nil {
			return fmt.Errorf("cannot compute status: %w", err)
		}

		resp = &subjectStatusResponse{
			SubjectID:      subject.ID.String(),
			ExternalUserID: externalUserID,
			Status:         string(status.Status),
		}

		if status.AcceptedVersionID != nil {
			s := status.AcceptedVersionID.String()
			resp.AcceptedVersionID = &s
		}

		resp.AcceptedAt = status.AcceptedAt

		if status.CurrentVersionID != nil {
			s := status.CurrentVersionID.String()
			resp.CurrentVersionID = &s
			resp.CurrentVersionMajor = status.CurrentVersionMajor
			resp.CurrentVersionMinor = status.CurrentVersionMinor
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, sign.ErrSubjectNotFound) {
			renderError(w, http.StatusNotFound, codeSubjectNotFound, "Subject not found.", reqID)
			return
		}

		h.logger.ErrorCtx(r.Context(), "cannot get subject status", log.Error(err))
		renderError(w, http.StatusInternalServerError, "internal_error", "Internal server error.", reqID)

		return
	}

	httpserver.RenderJSON(w, http.StatusOK, resp)
}

type batchStatusRequest struct {
	SigningConfigIDs []string `json:"signing_config_ids"`
	AccountID        *string  `json:"account_id,omitempty"`
}

func (h *Handler) handleBatchSubjectStatus(w http.ResponseWriter, r *http.Request) {
	reqID := resolveRequestID(r)
	w.Header().Set("X-Request-ID", reqID)

	externalUserID := chi.URLParam(r, "externalUserID")

	var body batchStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		renderError(w, http.StatusBadRequest, "invalid_request", "Invalid request body.", reqID)
		return
	}

	if len(body.SigningConfigIDs) == 0 {
		renderError(w, http.StatusBadRequest, "invalid_request", "signing_config_ids is required.", reqID)
		return
	}

	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		renderError(w, http.StatusUnauthorized, codeMissingAuthorization, "Missing authorization.", reqID)
		return
	}

	parts := strings.SplitN(authorization, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		renderError(w, http.StatusUnauthorized, codeInvalidKey, "Invalid authorization.", reqID)
		return
	}

	key := parts[1]

	statuses := make(map[string]*subjectStatusResponse)

	err := h.pgClient.WithConn(r.Context(), func(ctx context.Context, conn pg.Querier) error {
		for _, cfgIDStr := range body.SigningConfigIDs {
			cfgID, err := gid.ParseGID(cfgIDStr)
			if err != nil {
				return fmt.Errorf("invalid signing config ID %q: %w", cfgIDStr, err)
			}

			cfgScope := coredata.NewScopeFromObjectID(cfgID)

			var cfg coredata.DocumentSigningConfig
			if err := cfg.LoadByID(ctx, conn, cfgScope, cfgID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					return sign.ErrConfigNotFound
				}

				return fmt.Errorf("cannot load signing config: %w", err)
			}

			if !keyBelongsToOrg(key, &cfg) {
				return sign.ErrInvalidKey
			}

			scope := coredata.NewScopeFromObjectID(cfg.OrganizationID)

			var subject coredata.ClickwrapSubject

			if err := subject.LoadByExternalUserID(ctx, conn, scope, cfg.OrganizationID, externalUserID); err != nil {
				if errors.Is(err, coredata.ErrResourceNotFound) {
					statuses[cfgIDStr] = &subjectStatusResponse{
						ExternalUserID: externalUserID,
						Status:         string(sign.AcceptanceStatusRequired),
					}
					continue
				}

				return fmt.Errorf("cannot load subject: %w", err)
			}

			status, err := h.signSvc.ComputeSubjectStatus(ctx, conn, scope, &cfg, &subject)
			if err != nil {
				return fmt.Errorf("cannot compute status: %w", err)
			}

			resp := &subjectStatusResponse{
				SubjectID:      subject.ID.String(),
				ExternalUserID: externalUserID,
				Status:         string(status.Status),
				AcceptedAt:     status.AcceptedAt,
			}

			if status.AcceptedVersionID != nil {
				s := status.AcceptedVersionID.String()
				resp.AcceptedVersionID = &s
			}

			if status.CurrentVersionID != nil {
				s := status.CurrentVersionID.String()
				resp.CurrentVersionID = &s
				resp.CurrentVersionMajor = status.CurrentVersionMajor
				resp.CurrentVersionMinor = status.CurrentVersionMinor
			}

			statuses[cfgIDStr] = resp
		}

		return nil
	})
	if err != nil {
		h.logger.ErrorCtx(r.Context(), "cannot get batch subject status", log.Error(err))
		renderError(w, http.StatusInternalServerError, "internal_error", "Internal server error.", reqID)

		return
	}

	httpserver.RenderJSON(w, http.StatusOK, statuses)
}

type recordAcceptanceRequest struct {
	ExternalUserID    string  `json:"external_user_id"`
	Email             string  `json:"email"`
	FullName          *string `json:"full_name,omitempty"`
	AccountID         *string `json:"account_id,omitempty"`
	DocumentVersionID string  `json:"document_version_id"`
	ConsentText       string  `json:"consent_text"`
	IPAddress         string  `json:"ip_address"`
	UserAgent         string  `json:"user_agent"`
	IdempotencyKey    *string `json:"idempotency_key,omitempty"`
	SdkVersion        *string `json:"sdk_version,omitempty"`
}

type recordAcceptanceResponse struct {
	AcceptanceID      string    `json:"acceptance_id"`
	SubjectID         string    `json:"subject_id"`
	DocumentVersionID string    `json:"document_version_id"`
	AcceptedAt        time.Time `json:"accepted_at"`
	Mode              string    `json:"mode"`
	Status            string    `json:"status"`
}

func (h *Handler) handleRecordAcceptance(w http.ResponseWriter, r *http.Request) {
	reqID := resolveRequestID(r)
	w.Header().Set("X-Request-ID", reqID)

	signingConfigIDStr := chi.URLParam(r, "signingConfigID")

	var body recordAcceptanceRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		renderError(w, http.StatusBadRequest, "invalid_request", "Invalid request body.", reqID)
		return
	}

	var (
		auth       *authResult
		statusCode int
		errCode    string
	)

	connErr := h.pgClient.WithConn(r.Context(), func(ctx context.Context, conn pg.Querier) error {
		auth, statusCode, errCode = h.resolveAuth(ctx, conn, r, signingConfigIDStr)

		return nil
	})
	if connErr != nil || auth == nil {
		h.logAndRecordRejection(r, signingConfigIDStr, statusCode, errCode, reqID)
		renderError(w, statusCode, errCode, errCode, reqID)
		return
	}

	if !auth.isSecret {
		h.logAndRecordRejection(r, signingConfigIDStr, http.StatusForbidden, codeInvalidKey, reqID)
		renderError(w, http.StatusForbidden, codeInvalidKey, "Secret key required for recording acceptances.", reqID)
		return
	}

	docVersionID, err := gid.ParseGID(body.DocumentVersionID)
	if err != nil {
		renderError(w, http.StatusBadRequest, "invalid_request", "Invalid document_version_id.", reqID)
		return
	}

	countryCode := h.resolveCountryCode(r)

	scope := coredata.NewScopeFromObjectID(auth.organizationID)

	result, err := h.signSvc.RecordAcceptance(r.Context(), scope, sign.RecordAcceptanceRequest{
		OrganizationID:    auth.organizationID,
		SigningConfigID:   auth.signingConfig.ID,
		ExternalUserID:    body.ExternalUserID,
		Email:             body.Email,
		FullName:          body.FullName,
		AccountID:         body.AccountID,
		DocumentVersionID: docVersionID,
		ConsentText:       body.ConsentText,
		IPAddress:         body.IPAddress,
		UserAgent:         body.UserAgent,
		SdkVersion:        body.SdkVersion,
		CountryCode:       countryCode,
		IdempotencyKey:    body.IdempotencyKey,
		Mode:              auth.mode,
	})
	if err != nil {
		h.handleAcceptanceError(w, r, err, signingConfigIDStr, reqID)
		return
	}

	statusHTTP := http.StatusCreated

	if !result.IsNew {
		statusHTTP = http.StatusOK
	}

	httpserver.RenderJSON(w, statusHTTP, recordAcceptanceResponse{
		AcceptanceID:      result.Acceptance.ID.String(),
		SubjectID:         result.Acceptance.SubjectID.String(),
		DocumentVersionID: result.Acceptance.DocumentVersionID.String(),
		AcceptedAt:        result.Acceptance.AcceptedAt,
		Mode:              result.Acceptance.Mode.String(),
		Status:            "pending_seal",
	})
}

type batchAcceptanceItem struct {
	SigningConfigID   string `json:"signing_config_id"`
	DocumentVersionID string `json:"document_version_id"`
	ConsentText       string `json:"consent_text"`
}

type batchRecordRequest struct {
	ExternalUserID string                `json:"external_user_id"`
	Email          string                `json:"email"`
	FullName       *string               `json:"full_name,omitempty"`
	AccountID      *string               `json:"account_id,omitempty"`
	IPAddress      string                `json:"ip_address"`
	UserAgent      string                `json:"user_agent"`
	IdempotencyKey *string               `json:"idempotency_key,omitempty"`
	SdkVersion     *string               `json:"sdk_version,omitempty"`
	Acceptances    []batchAcceptanceItem `json:"acceptances"`
}

func (h *Handler) handleBatchRecordAcceptances(w http.ResponseWriter, r *http.Request) {
	reqID := resolveRequestID(r)
	w.Header().Set("X-Request-ID", reqID)

	var body batchRecordRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		renderError(w, http.StatusBadRequest, "invalid_request", "Invalid request body.", reqID)
		return
	}

	if len(body.Acceptances) == 0 {
		renderError(w, http.StatusBadRequest, "invalid_request", "acceptances list is required.", reqID)
		return
	}

	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		renderError(w, http.StatusUnauthorized, codeMissingAuthorization, "Missing authorization.", reqID)
		return
	}

	parts := strings.SplitN(authorization, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || !isSecretKeyStr(parts[1]) {
		renderError(w, http.StatusForbidden, codeInvalidKey, "Secret key required for batch recording.", reqID)
		return
	}

	countryCode := h.resolveCountryCode(r)
	mode, _ := keyModeFromKey(parts[1])

	type batchResult struct {
		AcceptanceID    string    `json:"acceptance_id"`
		SigningConfigID string    `json:"signing_config_id"`
		SubjectID       string    `json:"subject_id"`
		AcceptedAt      time.Time `json:"accepted_at"`
		Mode            string    `json:"mode"`
		Status          string    `json:"status"`
	}

	results := make([]batchResult, 0, len(body.Acceptances))

	for _, item := range body.Acceptances {
		cfgID, err := gid.ParseGID(item.SigningConfigID)
		if err != nil {
			renderError(w, http.StatusBadRequest, codeBatchValidationFailed,
				fmt.Sprintf("invalid signing_config_id: %s", item.SigningConfigID), reqID)
			return
		}

		docVersionID, err := gid.ParseGID(item.DocumentVersionID)
		if err != nil {
			renderError(w, http.StatusBadRequest, codeBatchValidationFailed,
				fmt.Sprintf("invalid document_version_id: %s", item.DocumentVersionID), reqID)
			return
		}

		scope := coredata.NewScopeFromObjectID(cfgID)

		result, err := h.signSvc.RecordAcceptance(r.Context(), scope, sign.RecordAcceptanceRequest{
			OrganizationID:    cfgID,
			SigningConfigID:   cfgID,
			ExternalUserID:    body.ExternalUserID,
			Email:             body.Email,
			FullName:          body.FullName,
			AccountID:         body.AccountID,
			DocumentVersionID: docVersionID,
			ConsentText:       item.ConsentText,
			IPAddress:         body.IPAddress,
			UserAgent:         body.UserAgent,
			SdkVersion:        body.SdkVersion,
			CountryCode:       countryCode,
			IdempotencyKey:    body.IdempotencyKey,
			Mode:              mode,
		})
		if err != nil {
			renderError(w, http.StatusBadRequest, codeBatchValidationFailed,
				fmt.Sprintf("acceptance for config %s failed: %s", item.SigningConfigID, err.Error()), reqID)
			return
		}

		results = append(results, batchResult{
			AcceptanceID:    result.Acceptance.ID.String(),
			SigningConfigID: item.SigningConfigID,
			SubjectID:       result.Acceptance.SubjectID.String(),
			AcceptedAt:      result.Acceptance.AcceptedAt,
			Mode:            result.Acceptance.Mode.String(),
			Status:          "pending_seal",
		})
	}

	httpserver.RenderJSON(w, http.StatusCreated, map[string]any{"acceptances": results})
}

type revokeRequest struct {
	Reason *string `json:"reason,omitempty"`
}

func (h *Handler) handleRevokeAcceptance(w http.ResponseWriter, r *http.Request) {
	reqID := resolveRequestID(r)
	w.Header().Set("X-Request-ID", reqID)

	acceptanceIDStr := chi.URLParam(r, "acceptanceID")

	acceptanceID, err := gid.ParseGID(acceptanceIDStr)
	if err != nil {
		renderError(w, http.StatusNotFound, codeAcceptanceNotFound, "Acceptance not found.", reqID)
		return
	}

	authorization := r.Header.Get("Authorization")
	if authorization == "" {
		renderError(w, http.StatusUnauthorized, codeMissingAuthorization, "Missing authorization.", reqID)
		return
	}

	parts := strings.SplitN(authorization, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || !isSecretKeyStr(parts[1]) {
		renderError(w, http.StatusForbidden, codeInvalidKey, "Secret key required for revocation.", reqID)
		return
	}

	var body revokeRequest
	_ = json.NewDecoder(r.Body).Decode(&body)

	scope := coredata.NewScopeFromObjectID(acceptanceID)

	acceptance, err := h.signSvc.RevokeAcceptance(r.Context(), scope, acceptanceID, body.Reason)
	if err != nil {
		if errors.Is(err, sign.ErrAcceptanceNotFound) {
			renderError(w, http.StatusNotFound, codeAcceptanceNotFound, "Acceptance not found.", reqID)
			return
		}

		if errors.Is(err, sign.ErrAlreadyRevoked) {
			renderError(w, http.StatusConflict, codeAlreadyRevoked, "Acceptance is already revoked.", reqID)
			return
		}

		h.logger.ErrorCtx(r.Context(), "cannot revoke acceptance", log.Error(err))
		renderError(w, http.StatusInternalServerError, "internal_error", "Internal server error.", reqID)

		return
	}

	httpserver.RenderJSON(w, http.StatusOK, map[string]any{
		"acceptance_id": acceptance.ID.String(),
		"revoked_at":    acceptance.RevokedAt,
	})
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func (h *Handler) handleAcceptanceError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
	signingConfigIDStr string,
	reqID string,
) {
	switch {
	case errors.Is(err, sign.ErrConfigNotFound):
		h.logAndRecordRejection(r, signingConfigIDStr, http.StatusNotFound, codeConfigNotFound, reqID)
		renderError(w, http.StatusNotFound, codeConfigNotFound, "Signing config not found.", reqID)
	case errors.Is(err, sign.ErrConfigNotActive):
		h.logAndRecordRejection(r, signingConfigIDStr, http.StatusForbidden, codeConfigNotActive, reqID)
		renderError(w, http.StatusForbidden, codeConfigNotActive, "Signing config is not active.", reqID)
	case errors.Is(err, sign.ErrVersionNotPresentable):
		renderError(w, http.StatusBadRequest, codeVersionNotPresentable, "Document version has no published PDF yet.", reqID)
	case errors.Is(err, sign.ErrVersionMismatch):
		renderError(w, http.StatusBadRequest, codeVersionMismatch, "Document version is not current.", reqID)
	case errors.Is(err, sign.ErrConsentTextMismatch):
		renderError(w, http.StatusBadRequest, codeConsentTextMismatch, "Consent text does not match.", reqID)
	case errors.Is(err, sign.ErrInvalidEmail):
		renderError(w, http.StatusBadRequest, codeInvalidEmail, "Invalid email address.", reqID)
	case errors.Is(err, sign.ErrInvalidIPAddress):
		renderError(w, http.StatusBadRequest, codeInvalidIPAddress, "Invalid IP address.", reqID)
	case errors.Is(err, sign.ErrTestKeyOnLiveConfig):
		h.logAndRecordRejection(r, signingConfigIDStr, http.StatusForbidden, codeTestKeyOnLiveConfig, reqID)
		renderError(w, http.StatusForbidden, codeTestKeyOnLiveConfig, "Test key used against live config.", reqID)
	default:
		h.logger.ErrorCtx(r.Context(), "cannot record acceptance", log.Error(err))
		renderError(w, http.StatusInternalServerError, "internal_error", "Internal server error.", reqID)
	}
}

func (h *Handler) logAndRecordRejection(
	r *http.Request,
	signingConfigIDStr string,
	statusCode int,
	errCode string,
	reqID string,
) {
	h.logger.InfoCtx(
		r.Context(),
		"data plane request rejected",
		log.String("error_code", errCode),
		log.String("path", r.URL.Path),
		log.String("method", r.Method),
		log.Int("status_code", statusCode),
	)

	go func() {
		ctx := context.Background()
		_ = h.pgClient.WithTx(ctx, func(ctx context.Context, tx pg.Tx) error {
			var cfgID *gid.GID

			if signingConfigIDStr != "" {
				id, err := gid.ParseGID(signingConfigIDStr)
				if err == nil {
					cfgID = &id
				}
			}

			var orgID gid.GID

			scope := coredata.NewScope(gid.TenantID{}, gid.Nil)

			if cfgID != nil {
				scope = coredata.NewScopeFromObjectID(*cfgID)
				orgID = *cfgID
			}

			entry := &coredata.DataPlaneRequestLog{
				ID:              gid.New(scope.GetTenantID(), coredata.DataPlaneRequestLogEntityType),
				OrganizationID:  orgID,
				SigningConfigID: cfgID,
				RequestPath:     r.URL.Path,
				Method:          r.Method,
				StatusCode:      statusCode,
				ErrorCode:       errCode,
				RequestID:       reqID,
				OccurredAt:      time.Now(),
			}

			return entry.Insert(ctx, tx, scope)
		})
	}()
}

func (h *Handler) resolveCountryCode(r *http.Request) *string {
	ip := clientip.Extract(r)

	location, err := h.geolocSvc.LookupLocation(r.Context(), ip)
	if err != nil || location.CountryCode == "" {
		return nil
	}

	cc := string(location.CountryCode)

	return &cc
}

func setCORSHeaders(w http.ResponseWriter, r *http.Request, cfg *coredata.DocumentSigningConfig) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return
	}

	if originAllowed(origin, cfg.AllowedOrigins) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
		w.Header().Set("Vary", "Origin")
	}
}

func buildConsentText(cfg *coredata.DocumentSigningConfig, dv *coredata.DocumentVersion) string {
	kindLabel := "Terms of Service"
	switch cfg.Kind {
	case coredata.DocumentSigningConfigKindTermsOfService:
		kindLabel = "Terms of Service"
	case coredata.DocumentSigningConfigKindAcceptableUsePolicy:
		kindLabel = "Acceptable Use Policy"
	case coredata.DocumentSigningConfigKindOther:
		kindLabel = "Agreement"
	}

	return fmt.Sprintf(
		"By checking this box, you agree to our %s (v%d.%d).",
		kindLabel, dv.Major, dv.Minor,
	)
}

func ptrStr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

func isSecretKeyStr(key string) bool {
	return strings.HasPrefix(key, "sk_live_") || strings.HasPrefix(key, "sk_test_")
}

func keyBelongsToOrg(key string, cfg *coredata.DocumentSigningConfig) bool {
	return cfg.PublishableKey == key ||
		strings.HasPrefix(key, "sk_live_") ||
		strings.HasPrefix(key, "sk_test_")
}

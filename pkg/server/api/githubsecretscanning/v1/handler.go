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

package githubsecretscanning_v1

import (
	"context"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.gearno.de/kit/httpserver"
	"go.gearno.de/kit/log"
)

const (
	maxAlertBodyBytes = 5 << 20

	publicKeyIdentifierHeader = "Github-Public-Key-Identifier"
	publicKeySignatureHeader  = "Github-Public-Key-Signature"
)

type (
	TokenRevoker interface {
		SecretScanningTokenType() string
		RevokeLeakedManualAccessToken(ctx context.Context, tokenValue string) (bool, error)
	}

	alert struct {
		Token  string `json:"token"`
		Type   string `json:"type"`
		URL    string `json:"url"`
		Source string `json:"source"`
	}

	handler struct {
		logger       *log.Logger
		tokenRevoker TokenRevoker
		keyProvider  KeyProvider
	}
)

func NewMux(logger *log.Logger, tokenRevoker TokenRevoker, keyProvider KeyProvider) http.Handler {
	h := &handler{
		logger:       logger,
		tokenRevoker: tokenRevoker,
		keyProvider:  keyProvider,
	}

	mux := chi.NewMux()
	mux.Post("/alerts", h.handleAlerts)

	return mux
}

func (h *handler) handleAlerts(w http.ResponseWriter, r *http.Request) {
	defer func() { _ = r.Body.Close() }()

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxAlertBodyBytes))
	if err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			renderError(w, http.StatusRequestEntityTooLarge, "request body is too large")

			return
		}

		renderError(w, http.StatusBadRequest, "invalid request body")

		return
	}

	keyIdentifier := r.Header.Get(publicKeyIdentifierHeader)
	signature := r.Header.Get(publicKeySignatureHeader)
	if keyIdentifier == "" || signature == "" {
		renderError(w, http.StatusUnauthorized, "invalid signature")

		return
	}

	key, err := h.keyProvider.PublicKey(r.Context(), keyIdentifier)
	if err != nil {
		if errors.Is(err, ErrPublicKeyNotFound) {
			renderError(w, http.StatusUnauthorized, "invalid signature")

			return
		}

		h.logger.ErrorCtx(r.Context(), "cannot load GitHub secret scanning public key", log.Error(err))
		renderError(w, http.StatusServiceUnavailable, "signature verification unavailable")

		return
	}

	if !verifySignature(body, signature, key) {
		renderError(w, http.StatusUnauthorized, "invalid signature")

		return
	}

	var alerts []alert
	if err := json.Unmarshal(body, &alerts); err != nil || len(alerts) == 0 {
		renderError(w, http.StatusBadRequest, "invalid alert payload")

		return
	}

	tokenType := h.tokenRevoker.SecretScanningTokenType()
	for _, item := range alerts {
		if item.Type != tokenType || item.Token == "" {
			continue
		}

		if _, err := h.tokenRevoker.RevokeLeakedManualAccessToken(r.Context(), item.Token); err != nil {
			h.logger.ErrorCtx(r.Context(), "cannot revoke leaked Probo API token", log.Error(err))
			renderError(w, http.StatusInternalServerError, "cannot process alert")

			return
		}
	}

	httpserver.RenderJSON(
		w,
		http.StatusOK,
		map[string]bool{
			"accepted": true,
		},
	)
}

func verifySignature(body []byte, encodedSignature string, key *ecdsa.PublicKey) bool {
	signature, err := base64.StdEncoding.DecodeString(encodedSignature)
	if err != nil {
		return false
	}

	digest := sha256.Sum256(body)

	return ecdsa.VerifyASN1(key, digest[:], signature)
}

func renderError(w http.ResponseWriter, status int, message string) {
	httpserver.RenderJSON(
		w,
		status,
		map[string]string{
			"error": message,
		},
	)
}

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
	"encoding/json"
	"net/http"
)

const (
	codeConsentTextMismatch    = "consent_text_mismatch"
	codeVersionNotPresentable  = "version_not_presentable"
	codeVersionMismatch        = "version_mismatch"
	codeInvalidEmail           = "invalid_email"
	codeInvalidIPAddress       = "invalid_ip_address"
	codeBatchValidationFailed  = "batch_validation_failed"
	codeMissingAuthorization   = "missing_authorization"
	codeInvalidKey             = "invalid_key"
	codeIdentityJWTExpired     = "identity_jwt_expired"
	codeIdentityJWTInvalid     = "identity_jwt_invalid"
	codeConfigNotActive        = "config_not_active"
	codeOriginNotAllowed       = "origin_not_allowed"
	codeTestKeyOnLiveConfig    = "test_key_on_live_config"
	codeConfigNotFound         = "config_not_found"
	codeSubjectNotFound        = "subject_not_found"
	codeAcceptanceNotFound     = "acceptance_not_found"
	codeAlreadyRevoked         = "already_revoked"
	codeRateLimited            = "rate_limited"
	codeDocumentNotPresentable = "document_not_presentable"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

func renderError(w http.ResponseWriter, status int, code, message, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	body := errorBody{
		Error: errorDetail{
			Code:      code,
			Message:   message,
			RequestID: requestID,
		},
	}

	_ = json.NewEncoder(w).Encode(body)
}

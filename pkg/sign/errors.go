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

package sign

import "errors"

var (
	ErrConfigNotFound         = errors.New("signing config not found")
	ErrConfigNotActive        = errors.New("signing config is not active")
	ErrConfigAlreadyExists    = errors.New("signing config already exists for this document")
	ErrSecretNotFound         = errors.New("organization signing secret not found")
	ErrSubjectNotFound        = errors.New("clickwrap subject not found")
	ErrAcceptanceNotFound     = errors.New("acceptance not found")
	ErrAlreadyRevoked         = errors.New("acceptance is already revoked")
	ErrDocumentNotPresentable = errors.New("document has no presentable version")
	ErrVersionNotPresentable  = errors.New("document version is not presentable")
	ErrVersionMismatch        = errors.New("document version is not the current presentable version")
	ErrConsentTextMismatch    = errors.New("consent text does not match the version canonical text")
	ErrInvalidEmail           = errors.New("invalid email address")
	ErrInvalidIPAddress       = errors.New("invalid IP address")
	ErrTestKeyOnLiveConfig    = errors.New("test key used against a live signing config")
	ErrOriginNotAllowed       = errors.New("origin not in allowed origins")
	ErrMissingAuthorization   = errors.New("missing authorization header")
	ErrInvalidKey             = errors.New("key not recognized or key hash mismatch")
	ErrIdentityJWTExpired     = errors.New("identity JWT has expired")
	ErrIdentityJWTInvalid     = errors.New("identity JWT signature verification failed")
	ErrBatchValidationFailed  = errors.New("one or more batch items failed validation")
)

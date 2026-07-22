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

package certmanager

import (
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/acme"
)

const defaultCooldown = time.Hour

var (
	ErrACMERateLimited = errors.New("acme rate limited")
	ErrOrderNotReady   = errors.New("acme order not ready")
	ErrOrderInvalid    = errors.New("acme order invalid")
)

type ACMEError struct {
	op          string
	err         error
	problemType string
	detail      string
	rateLimited bool
	retryAfter  time.Duration
}

func (e *ACMEError) Error() string {
	if e == nil {
		return ""
	}

	if e.err != nil {
		return fmt.Sprintf("%s: %v", e.op, e.err)
	}

	return e.op
}

func (e *ACMEError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.err
}

func (e *ACMEError) Is(target error) bool {
	return e != nil && e.rateLimited && target == ErrACMERateLimited
}

// RetryAfter returns how long callers should wait before retrying.
// For rate-limited errors it prefers the ACME Retry-After value and falls back
// to defaultCooldown when the header is absent. Non-rate-limited errors return 0.
func (e *ACMEError) RetryAfter() time.Duration {
	if e == nil || !e.rateLimited {
		return 0
	}

	if e.retryAfter > 0 {
		return e.retryAfter
	}

	return defaultCooldown
}

func (e *ACMEError) ProblemType() string {
	if e == nil {
		return ""
	}

	return e.problemType
}

func (e *ACMEError) Detail() string {
	if e == nil {
		return ""
	}

	return e.detail
}

func newACMEError(op string, err error) *ACMEError {
	if err == nil {
		return nil
	}

	out := &ACMEError{
		op:  op,
		err: err,
	}

	acmeErr, ok := errors.AsType[*acme.Error](err)
	if !ok {
		out.detail = err.Error()
		return out
	}

	out.problemType = acmeErr.ProblemType
	out.detail = acmeErr.Detail

	if retryAfter, ok := acme.RateLimit(acmeErr); ok {
		out.rateLimited = true
		out.retryAfter = retryAfter
	}

	return out
}

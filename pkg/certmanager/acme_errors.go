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
	"math"
	"net/http"
	"strconv"
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
	op            string
	err           error
	problemType   string
	detail        string
	rateLimited   bool
	retryAfter    time.Duration
	retryAfterSet bool
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

func (e *ACMEError) RetryAfter() time.Duration {
	if e == nil || !e.rateLimited {
		return 0
	}

	if e.retryAfterSet {
		if e.retryAfter < 0 {
			return 0
		}

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

	if _, ok := acme.RateLimit(acmeErr); ok {
		out.rateLimited = true

		if retryAfter, ok := parseRetryAfter(acmeErr.Header); ok {
			out.retryAfter = retryAfter
			out.retryAfterSet = true
		}
	}

	return out
}

// parseRetryAfter reports the Retry-After delay and whether the header was
// present and parseable. It mirrors the delta-seconds and HTTP-date forms the
// ACME client understands. An absent or unparseable value returns ok=false so
// callers can apply their own fallback.
func parseRetryAfter(header http.Header) (time.Duration, bool) {
	if header == nil {
		return 0, false
	}

	value := header.Get("Retry-After")
	if value == "" {
		return 0, false
	}

	// The delta-seconds form is an unsigned decimal integer (RFC 9110 §10.2.3).
	// Parse as signed 64-bit to match time.Duration's underlying type and avoid
	// unsigned-to-signed narrowing conversions. Negative or malformed values are
	// treated as unparseable so callers can apply their default cooldown. A value
	// larger than time.Duration can hold is clamped to the maximum duration.
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil && seconds >= 0 {
		maxSeconds := int64(math.MaxInt64 / int64(time.Second))
		if seconds > maxSeconds {
			return time.Duration(math.MaxInt64), true
		}

		return time.Duration(seconds) * time.Second, true
	}

	if date, err := http.ParseTime(value); err == nil {
		return time.Until(date), true
	}

	return 0, false
}

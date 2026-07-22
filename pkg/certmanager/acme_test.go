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

package certmanager

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/acme"
)

func TestNewACMEError_RateLimited(t *testing.T) {
	t.Parallel()

	err := newACMEError(
		"cannot create order",
		&acme.Error{
			ProblemType: "urn:ietf:params:acme:error:rateLimited",
			Detail:      "too many requests",
			Header:      http.Header{"Retry-After": []string{"120"}},
		},
	)

	require.NotNil(t, err)
	assert.ErrorIs(t, err, ErrACMERateLimited)
	assert.Equal(t, 2*time.Minute, err.RetryAfter())
	assert.Equal(t, "urn:ietf:params:acme:error:rateLimited", err.problemType)
	assert.Equal(t, "too many requests", err.detail)
}

func TestNewACMEError_RateLimitedDefaultCooldown(t *testing.T) {
	t.Parallel()

	err := newACMEError(
		"cannot create order",
		&acme.Error{ProblemType: "URN:IETF:PARAMS:ACME:ERROR:RATELIMITED"},
	)

	require.NotNil(t, err)
	assert.ErrorIs(t, err, ErrACMERateLimited)
	assert.Equal(t, defaultCooldown, err.RetryAfter())
}

func TestNewACMEError_NonRateLimited(t *testing.T) {
	t.Parallel()

	cause := errors.New("network timeout")
	err := newACMEError("cannot get order", cause)

	require.NotNil(t, err)
	assert.False(t, errors.Is(err, ErrACMERateLimited))
	assert.ErrorIs(t, err, cause)
	assert.Equal(t, time.Duration(0), err.RetryAfter())
}

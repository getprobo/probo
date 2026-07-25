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

package drivers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubTransport replays a fixed sequence of responses and records how many
// requests it received.
type stubTransport struct {
	responses []*http.Response
	calls     int
}

func (s *stubTransport) RoundTrip(*http.Request) (*http.Response, error) {
	resp := s.responses[min(s.calls, len(s.responses)-1)]
	s.calls++

	return resp, nil
}

func response(status int, body string, header http.Header) *http.Response {
	if header == nil {
		header = http.Header{}
	}

	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     header,
	}
}

func newRequest(t *testing.T, ctx context.Context) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.example.com/users", nil)
	require.NoError(t, err)

	return req
}

func TestRetryRoundTripper(t *testing.T) {
	t.Parallel()

	t.Run("returns a success without retrying", func(t *testing.T) {
		t.Parallel()

		stub := &stubTransport{responses: []*http.Response{response(http.StatusOK, `{"ok":true}`, nil)}}
		rt := &retryRoundTripper{next: stub, maxRetries: 3}

		resp, err := rt.RoundTrip(newRequest(t, t.Context()))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 1, stub.calls)
	})

	t.Run("does not retry a client error", func(t *testing.T) {
		t.Parallel()

		stub := &stubTransport{responses: []*http.Response{response(http.StatusForbidden, "denied", nil)}}
		rt := &retryRoundTripper{next: stub, maxRetries: 3}

		resp, err := rt.RoundTrip(newRequest(t, t.Context()))
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		assert.Equal(t, 1, stub.calls)
	})

	t.Run("retries a 5xx then succeeds", func(t *testing.T) {
		t.Parallel()

		stub := &stubTransport{responses: []*http.Response{
			response(http.StatusBadGateway, "boom", nil),
			response(http.StatusOK, `{"ok":true}`, nil),
		}}
		rt := &retryRoundTripper{next: stub, maxRetries: 3}

		resp, err := rt.RoundTrip(newRequest(t, t.Context()))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 2, stub.calls)
	})

	// The exhausted response must still be readable: callers decode the error
	// body to report the provider's message.
	t.Run("returns a readable body after exhausting retries", func(t *testing.T) {
		t.Parallel()

		stub := &stubTransport{responses: []*http.Response{response(http.StatusInternalServerError, "upstream failed", nil)}}
		rt := &retryRoundTripper{next: stub, maxRetries: 2}

		resp, err := rt.RoundTrip(newRequest(t, t.Context()))
		require.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, "upstream failed", string(body))
	})

	// The final attempt has nothing after it, so the old code's trailing sleep
	// only spent the caller's deadline to return a response it already had.
	t.Run("does not sleep after the final attempt", func(t *testing.T) {
		t.Parallel()

		stub := &stubTransport{responses: []*http.Response{response(http.StatusServiceUnavailable, "down", nil)}}
		rt := &retryRoundTripper{next: stub, maxRetries: 3}

		start := time.Now()
		resp, err := rt.RoundTrip(newRequest(t, t.Context()))
		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
		assert.Equal(t, 3, stub.calls)
		// Two waits (250ms + 500ms), never a third after the last attempt.
		assert.Less(t, elapsed, 900*time.Millisecond)
		assert.GreaterOrEqual(t, elapsed, 750*time.Millisecond)
	})

	// A long Retry-After cannot be honoured inside a per-source budget, so the
	// throttled response is surfaced immediately rather than slept through.
	t.Run("returns immediately when Retry-After exceeds the cap", func(t *testing.T) {
		t.Parallel()

		header := http.Header{}
		header.Set("Retry-After", "60")

		stub := &stubTransport{responses: []*http.Response{response(http.StatusTooManyRequests, "slow down", header)}}
		rt := &retryRoundTripper{next: stub, maxRetries: 3}

		start := time.Now()
		resp, err := rt.RoundTrip(newRequest(t, t.Context()))
		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		assert.Equal(t, 1, stub.calls, "must not spend a second attempt it cannot wait for")
		assert.Less(t, elapsed, 100*time.Millisecond)
	})

	// Retry-After shorter than the cap is honoured over the exponential step.
	t.Run("honours a short Retry-After", func(t *testing.T) {
		t.Parallel()

		header := http.Header{}
		header.Set("Retry-After", "1")

		stub := &stubTransport{responses: []*http.Response{
			response(http.StatusTooManyRequests, "slow down", header),
			response(http.StatusOK, `{"ok":true}`, nil),
		}}
		rt := &retryRoundTripper{next: stub, maxRetries: 3}

		start := time.Now()
		resp, err := rt.RoundTrip(newRequest(t, t.Context()))
		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		// 1s from Retry-After, not the 250ms exponential step.
		assert.GreaterOrEqual(t, elapsed, 1*time.Second)
	})

	// Sleeping past the deadline turns a reportable 429 into an opaque
	// "context deadline exceeded".
	t.Run("returns the response rather than sleeping past the deadline", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()

		stub := &stubTransport{responses: []*http.Response{response(http.StatusTooManyRequests, "slow down", nil)}}
		rt := &retryRoundTripper{next: stub, maxRetries: 3}

		start := time.Now()
		resp, err := rt.RoundTrip(newRequest(t, ctx))
		elapsed := time.Since(start)

		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		assert.Equal(t, 1, stub.calls)
		assert.Less(t, elapsed, 50*time.Millisecond)
	})

	t.Run("aborts when the context is cancelled mid-backoff", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		stub := &stubTransport{responses: []*http.Response{response(http.StatusInternalServerError, "boom", nil)}}
		rt := &retryRoundTripper{next: stub, maxRetries: 3}

		go func() {
			time.Sleep(30 * time.Millisecond)
			cancel()
		}()

		_, err := rt.RoundTrip(newRequest(t, ctx))
		require.ErrorIs(t, err, context.Canceled)
	})
}

func TestRetryAfter(t *testing.T) {
	t.Parallel()

	t.Run("absent header", func(t *testing.T) {
		t.Parallel()

		_, ok := retryAfter(response(http.StatusTooManyRequests, "", nil))
		assert.False(t, ok)
	})

	t.Run("delta seconds", func(t *testing.T) {
		t.Parallel()

		header := http.Header{}
		header.Set("Retry-After", "30")

		got, ok := retryAfter(response(http.StatusTooManyRequests, "", header))
		require.True(t, ok)
		assert.Equal(t, 30*time.Second, got)
	})

	t.Run("http date in the future", func(t *testing.T) {
		t.Parallel()

		header := http.Header{}
		header.Set("Retry-After", time.Now().Add(2*time.Second).UTC().Format(http.TimeFormat))

		got, ok := retryAfter(response(http.StatusTooManyRequests, "", header))
		require.True(t, ok)
		assert.Positive(t, got)
		assert.LessOrEqual(t, got, 2*time.Second)
	})

	t.Run("http date in the past yields no wait", func(t *testing.T) {
		t.Parallel()

		header := http.Header{}
		header.Set("Retry-After", time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat))

		got, ok := retryAfter(response(http.StatusTooManyRequests, "", header))
		require.True(t, ok)
		assert.Zero(t, got)
	})

	t.Run("garbage and negative values are ignored", func(t *testing.T) {
		t.Parallel()

		for _, value := range []string{"soon", "-5", "  "} {
			header := http.Header{}
			header.Set("Retry-After", value)

			_, ok := retryAfter(response(http.StatusTooManyRequests, "", header))
			assert.False(t, ok, value)
		}
	})
}

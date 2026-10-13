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

package connector

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubTransport struct {
	responses []*http.Response
	bodies    []string
	calls     int
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		_ = req.Body.Close()
		s.bodies = append(s.bodies, string(body))
	}

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
		Status:     http.StatusText(status),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     header,
	}
}

func TestRetryRoundTripper(t *testing.T) {
	t.Parallel()

	t.Run("returns a success without retrying", func(t *testing.T) {
		t.Parallel()

		stub := &stubTransport{responses: []*http.Response{response(http.StatusOK, `{"ok":true}`, nil)}}
		rt := &retryRoundTripper{next: stub, maxRetries: tokenFetchMaxRetries}

		resp, err := rt.RoundTrip(newRetryRequest(t, t.Context(), http.MethodGet, ""))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 1, stub.calls)
	})

	t.Run("does not retry a forbidden response", func(t *testing.T) {
		t.Parallel()

		stub := &stubTransport{responses: []*http.Response{response(http.StatusForbidden, "denied", nil)}}
		rt := &retryRoundTripper{next: stub, maxRetries: tokenFetchMaxRetries}

		resp, err := rt.RoundTrip(newRetryRequest(t, t.Context(), http.MethodGet, ""))
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
		assert.Equal(t, 1, stub.calls)
	})

	t.Run("retries a 503 then succeeds and resends the body", func(t *testing.T) {
		t.Parallel()

		stub := &stubTransport{responses: []*http.Response{
			response(http.StatusServiceUnavailable, `{"error":{"code":503,"message":"The service is currently unavailable.","status":"UNAVAILABLE"}}`, nil),
			response(http.StatusOK, `{"access_token":"new"}`, nil),
		}}
		rt := &retryRoundTripper{next: stub, maxRetries: tokenFetchMaxRetries}

		resp, err := rt.RoundTrip(newRetryRequest(t, t.Context(), http.MethodPost, "grant_type=refresh_token"))
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, 2, stub.calls)
		assert.Equal(t, []string{"grant_type=refresh_token", "grant_type=refresh_token"}, stub.bodies)
	})

	t.Run("returns a readable body after exhausting retries", func(t *testing.T) {
		t.Parallel()

		const unavailable = `{"error":{"code":503,"message":"The service is currently unavailable.","status":"UNAVAILABLE"}}`

		stub := &stubTransport{responses: []*http.Response{response(http.StatusServiceUnavailable, unavailable, nil)}}
		rt := &retryRoundTripper{next: stub, maxRetries: tokenFetchMaxRetries}

		resp, err := rt.RoundTrip(newRetryRequest(t, t.Context(), http.MethodGet, ""))
		require.NoError(t, err)
		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
		assert.Equal(t, tokenFetchMaxRetries, stub.calls)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, unavailable, string(body))
	})

	t.Run("does not sleep after the final attempt", func(t *testing.T) {
		t.Parallel()

		stub := &stubTransport{responses: []*http.Response{response(http.StatusServiceUnavailable, "down", nil)}}
		rt := &retryRoundTripper{next: stub, maxRetries: tokenFetchMaxRetries}

		start := time.Now()
		resp, err := rt.RoundTrip(newRetryRequest(t, t.Context(), http.MethodGet, ""))
		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
		assert.Equal(t, tokenFetchMaxRetries, stub.calls)
		// Two waits (250ms + 500ms), never a third after the last attempt.
		assert.Less(t, elapsed, 900*time.Millisecond)
		assert.GreaterOrEqual(t, elapsed, 750*time.Millisecond)
	})

	t.Run("aborts when the context is cancelled mid-backoff", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())

		stub := &stubTransport{responses: []*http.Response{response(http.StatusServiceUnavailable, "down", nil)}}
		rt := &retryRoundTripper{next: stub, maxRetries: tokenFetchMaxRetries}

		go func() {
			time.Sleep(30 * time.Millisecond)
			cancel()
		}()

		_, err := rt.RoundTrip(newRetryRequest(t, ctx, http.MethodGet, ""))
		require.ErrorIs(t, err, context.Canceled)
	})
}

func newRetryRequest(t *testing.T, ctx context.Context, method string, body string) *http.Request {
	t.Helper()

	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, "https://oauth2.googleapis.com/token", reader)
	require.NoError(t, err)

	if body != "" {
		snapshot := []byte(body)
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(snapshot)), nil
		}
	}

	return req
}

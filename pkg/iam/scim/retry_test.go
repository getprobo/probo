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

package scim

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

type scriptedResponse struct {
	status int
	body   string
	header http.Header
}

// stubTransport replays a fixed sequence of responses and records request bodies.
type stubTransport struct {
	responses []scriptedResponse
	bodies    []string
	calls     int
	err       error
}

func (s *stubTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		payload, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}

		s.bodies = append(s.bodies, string(payload))
	}

	call := s.calls
	s.calls++

	if s.err != nil && call == 0 {
		return nil, s.err
	}

	script := s.responses[min(call, len(s.responses)-1)]

	header := script.header.Clone()
	if header == nil {
		header = http.Header{}
	}

	return &http.Response{
		StatusCode: script.status,
		Body:       io.NopCloser(strings.NewReader(script.body)),
		Header:     header,
	}, nil
}

func script(status int, body string, header http.Header) scriptedResponse {
	return scriptedResponse{status: status, body: body, header: header}
}

func newRequest(t *testing.T, ctx context.Context) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://admin.googleapis.com/admin/directory/v1/users", nil)
	require.NoError(t, err)

	return req
}

func testRetry(next http.RoundTripper) *retryRoundTripper {
	rt := &retryRoundTripper{
		next:        next,
		attempts:    3,
		baseBackoff: 15 * time.Millisecond,
		maxWait:     5 * time.Second,
	}
	rt.sleep = func(context.Context, time.Duration) error {
		return nil
	}

	return rt
}

func TestWithServerErrorRetry(t *testing.T) {
	t.Parallel()

	baseTransport := &stubTransport{responses: []scriptedResponse{script(http.StatusOK, "ok", nil)}}
	base := &http.Client{
		Timeout:   3 * time.Second,
		Transport: baseTransport,
	}

	wrapped := withServerErrorRetry(base)

	assert.Equal(t, 3*time.Second, wrapped.Timeout)
	assert.NotSame(t, base, wrapped)
	assert.Same(t, baseTransport, base.Transport)

	rt, ok := wrapped.Transport.(*retryRoundTripper)
	require.True(t, ok)
	assert.Same(t, baseTransport, rt.next)
	assert.Equal(t, serverErrorRetryAttempts, rt.attempts)
	assert.Equal(t, serverErrorRetryBackoff, rt.baseBackoff)
	assert.Equal(t, serverErrorRetryMaxWait, rt.maxWait)
	assert.Nil(t, rt.sleep)
}

func TestRetryRoundTripper(t *testing.T) {
	t.Parallel()

	t.Run(
		"returns a success without retrying",
		func(t *testing.T) {
			t.Parallel()

			stub := &stubTransport{responses: []scriptedResponse{script(http.StatusOK, `{"ok":true}`, nil)}}
			rt := testRetry(stub)

			resp, err := rt.RoundTrip(newRequest(t, t.Context()))
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, 1, stub.calls)
		},
	)

	t.Run(
		"does not retry a client error",
		func(t *testing.T) {
			t.Parallel()

			stub := &stubTransport{responses: []scriptedResponse{script(http.StatusForbidden, "denied", nil)}}
			rt := testRetry(stub)

			resp, err := rt.RoundTrip(newRequest(t, t.Context()))
			require.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, resp.StatusCode)
			assert.Equal(t, 1, stub.calls)
		},
	)

	t.Run(
		"does not retry too many requests",
		func(t *testing.T) {
			t.Parallel()

			stub := &stubTransport{responses: []scriptedResponse{script(http.StatusTooManyRequests, "slow down", nil)}}
			rt := testRetry(stub)

			resp, err := rt.RoundTrip(newRequest(t, t.Context()))
			require.NoError(t, err)
			assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
			assert.Equal(t, 1, stub.calls)
		},
	)

	t.Run(
		"retries a 5xx then succeeds",
		func(t *testing.T) {
			t.Parallel()

			stub := &stubTransport{responses: []scriptedResponse{
				script(http.StatusBadGateway, "boom", nil),
				script(http.StatusOK, `{"ok":true}`, nil),
			}}

			var waits []time.Duration

			rt := testRetry(stub)
			rt.sleep = func(_ context.Context, wait time.Duration) error {
				waits = append(waits, wait)

				return nil
			}

			resp, err := rt.RoundTrip(newRequest(t, t.Context()))
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, 2, stub.calls)
			assert.Equal(t, []time.Duration{15 * time.Millisecond}, waits)
		},
	)

	t.Run(
		"returns a readable body after exhausting retries",
		func(t *testing.T) {
			t.Parallel()

			stub := &stubTransport{responses: []scriptedResponse{script(http.StatusInternalServerError, "upstream failed", nil)}}
			rt := testRetry(stub)
			rt.attempts = 2

			resp, err := rt.RoundTrip(newRequest(t, t.Context()))
			require.NoError(t, err)
			assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, "upstream failed", string(body))
			assert.Equal(t, 2, stub.calls)
		},
	)

	t.Run(
		"does not sleep after the final attempt",
		func(t *testing.T) {
			t.Parallel()

			stub := &stubTransport{responses: []scriptedResponse{script(http.StatusServiceUnavailable, "down", nil)}}

			var waits []time.Duration

			rt := testRetry(stub)
			rt.sleep = func(_ context.Context, wait time.Duration) error {
				waits = append(waits, wait)

				return nil
			}

			resp, err := rt.RoundTrip(newRequest(t, t.Context()))
			require.NoError(t, err)
			assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
			assert.Equal(t, 3, stub.calls)
			assert.Equal(
				t,
				[]time.Duration{15 * time.Millisecond, 30 * time.Millisecond},
				waits,
			)
		},
	)

	t.Run(
		"returns immediately when Retry-After exceeds the cap",
		func(t *testing.T) {
			t.Parallel()

			header := http.Header{}
			header.Set("Retry-After", "60")

			stub := &stubTransport{responses: []scriptedResponse{script(http.StatusServiceUnavailable, "down", header)}}

			var slept bool

			rt := testRetry(stub)
			rt.sleep = func(context.Context, time.Duration) error {
				slept = true

				return nil
			}

			resp, err := rt.RoundTrip(newRequest(t, t.Context()))
			require.NoError(t, err)
			assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
			assert.Equal(t, 1, stub.calls)
			assert.False(t, slept)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, "down", string(body))
		},
	)

	t.Run(
		"honours a short Retry-After",
		func(t *testing.T) {
			t.Parallel()

			header := http.Header{}
			header.Set("Retry-After", "1")

			stub := &stubTransport{responses: []scriptedResponse{
				script(http.StatusBadGateway, "boom", header),
				script(http.StatusOK, `{"ok":true}`, nil),
			}}

			var waits []time.Duration

			rt := testRetry(stub)
			rt.sleep = func(_ context.Context, wait time.Duration) error {
				waits = append(waits, wait)

				return nil
			}

			resp, err := rt.RoundTrip(newRequest(t, t.Context()))
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, 2, stub.calls)
			assert.Equal(t, []time.Duration{time.Second}, waits)
		},
	)

	t.Run(
		"returns the response rather than sleeping past the deadline",
		func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Millisecond)
			defer cancel()

			stub := &stubTransport{responses: []scriptedResponse{script(http.StatusBadGateway, "boom", nil)}}

			var slept bool

			rt := testRetry(stub)
			rt.baseBackoff = time.Second
			rt.sleep = func(context.Context, time.Duration) error {
				slept = true

				return nil
			}

			resp, err := rt.RoundTrip(newRequest(t, ctx))
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
			assert.Equal(t, 1, stub.calls)
			assert.False(t, slept)
		},
	)

	t.Run(
		"aborts when the context is cancelled mid-backoff",
		func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(t.Context())

			stub := &stubTransport{responses: []scriptedResponse{script(http.StatusInternalServerError, "boom", nil)}}
			rt := testRetry(stub)
			rt.baseBackoff = time.Second
			rt.sleep = nil

			go func() {
				time.Sleep(20 * time.Millisecond)
				cancel()
			}()

			_, err := rt.RoundTrip(newRequest(t, ctx))
			require.ErrorIs(t, err, context.Canceled)
		},
	)

	t.Run(
		"does not retry a transport error",
		func(t *testing.T) {
			t.Parallel()

			stub := &stubTransport{
				responses: []scriptedResponse{script(http.StatusOK, "ok", nil)},
				err:       context.DeadlineExceeded,
			}
			rt := testRetry(stub)

			_, err := rt.RoundTrip(newRequest(t, t.Context()))
			require.ErrorIs(t, err, context.DeadlineExceeded)
			assert.Equal(t, 1, stub.calls)
		},
	)

	t.Run(
		"stops when the body cannot be rewound",
		func(t *testing.T) {
			t.Parallel()

			stub := &stubTransport{responses: []scriptedResponse{
				script(http.StatusBadGateway, "boom", nil),
				script(http.StatusOK, "ok", nil),
			}}
			rt := testRetry(stub)

			req := newRequest(t, t.Context())
			req.Body = io.NopCloser(strings.NewReader("payload"))

			resp, err := rt.RoundTrip(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
			assert.Equal(t, 1, stub.calls)
		},
	)
}

func TestRetryRoundTripper_RewindsBody(t *testing.T) {
	t.Parallel()

	stub := &stubTransport{responses: []scriptedResponse{
		script(http.StatusServiceUnavailable, "down", nil),
		script(http.StatusOK, `{}`, nil),
	}}

	payload := `{"query":"users"}`

	req, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		"https://example.com/graphql",
		strings.NewReader(payload),
	)
	require.NoError(t, err)

	client := &http.Client{Transport: testRetry(stub)}

	resp, err := client.Do(req)
	require.NoError(t, err)

	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, stub.bodies, 2)
	assert.Equal(t, payload, stub.bodies[0])
	assert.Equal(t, payload, stub.bodies[1])
}

func TestRetryAfter(t *testing.T) {
	t.Parallel()

	t.Run(
		"absent header",
		func(t *testing.T) {
			t.Parallel()

			_, ok := retryAfter(&http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Header:     http.Header{},
			})
			assert.False(t, ok)
		},
	)

	t.Run(
		"delta seconds",
		func(t *testing.T) {
			t.Parallel()

			header := http.Header{}
			header.Set("Retry-After", "30")

			got, ok := retryAfter(&http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Header:     header,
			})
			require.True(t, ok)
			assert.Equal(t, 30*time.Second, got)
		},
	)

	t.Run(
		"negative delta",
		func(t *testing.T) {
			t.Parallel()

			header := http.Header{}
			header.Set("Retry-After", "-1")

			_, ok := retryAfter(&http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Header:     header,
			})
			assert.False(t, ok)
		},
	)

	t.Run(
		"http date in the future",
		func(t *testing.T) {
			t.Parallel()

			header := http.Header{}
			header.Set("Retry-After", time.Now().Add(2*time.Second).UTC().Format(http.TimeFormat))

			got, ok := retryAfter(&http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Header:     header,
			})
			require.True(t, ok)
			assert.Greater(t, got, time.Duration(0))
			assert.LessOrEqual(t, got, 2*time.Second)
		},
	)

	t.Run(
		"unparseable header",
		func(t *testing.T) {
			t.Parallel()

			header := http.Header{}
			header.Set("Retry-After", "later")

			_, ok := retryAfter(&http.Response{
				StatusCode: http.StatusServiceUnavailable,
				Header:     header,
			})
			assert.False(t, ok)
		},
	)
}

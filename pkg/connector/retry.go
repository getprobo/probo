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
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.gearno.de/kit/httpclient"
)

const (
	// tokenFetchMaxRetries is the number of attempts, including the first.
	tokenFetchMaxRetries = 3

	// tokenRetryBaseBackoff is the first wait; it doubles on each later attempt.
	tokenRetryBaseBackoff = 250 * time.Millisecond

	// tokenMaxRetryWait caps one wait. A Retry-After longer than this cannot
	// be honoured inside a sync deadline, so the response is returned as-is.
	tokenMaxRetryWait = 5 * time.Second
)

// newTokenHTTPClient is the client that posts to a token endpoint.
//
// A 503 from that POST fails the whole sync. The grant is safe to send
// again after a 429 or 5xx: the endpoint issued no token. A 4xx (revoked
// refresh token, forbidden) is the provider's answer and is not retried.
func newTokenHTTPClient(opts ...httpclient.Option) *http.Client {
	return &http.Client{
		Transport: &retryRoundTripper{
			next:       httpclient.DefaultPooledTransport(opts...),
			maxRetries: tokenFetchMaxRetries,
		},
	}
}

// retryRoundTripper retries 429 and 5xx responses with exponential backoff.
// A 4xx is returned on the first attempt. The request body is rewound
// between attempts; a body that cannot be rewound stops the retry and
// returns the response already in hand.
type retryRoundTripper struct {
	next       http.RoundTripper
	maxRetries int
}

func (rt *retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var lastResp *http.Response

	for attempt := range rt.maxRetries {
		attemptReq := req

		if attempt > 0 && req.Body != nil {
			if req.GetBody == nil {
				break
			}

			body, err := req.GetBody()
			if err != nil {
				return nil, fmt.Errorf("cannot rewind request body for retry: %w", err)
			}

			attemptReq = req.Clone(req.Context())
			attemptReq.Body = body
		}

		resp, err := rt.next.RoundTrip(attemptReq)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode < 500 {
			return resp, nil
		}

		// Buffer the body so a retry-exhausted response is still readable.
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(body))
		lastResp = resp

		if attempt == rt.maxRetries-1 {
			break
		}

		wait := tokenRetryBaseBackoff << attempt
		if after, ok := tokenRetryAfter(resp); ok {
			wait = after
		}

		if wait > tokenMaxRetryWait {
			break
		}

		if deadline, ok := req.Context().Deadline(); ok && time.Until(deadline) <= wait {
			break
		}

		timer := time.NewTimer(wait)

		select {
		case <-req.Context().Done():
			timer.Stop()

			return nil, req.Context().Err()
		case <-timer.C:
		}
	}

	return lastResp, nil
}

// tokenRetryAfter reads Retry-After as delta-seconds or an HTTP-date.
// The bool reports whether the header was present and parseable.
func tokenRetryAfter(resp *http.Response) (time.Duration, bool) {
	value := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if value == "" {
		return 0, false
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds < 0 {
			return 0, false
		}

		return time.Duration(seconds) * time.Second, true
	}

	if at, err := http.ParseTime(value); err == nil {
		return max(time.Until(at), 0), true
	}

	return 0, false
}

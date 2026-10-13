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
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	// serverErrorRetryAttempts is the total number of tries, including the
	// first. Directory reads are paginated, so the budget stays small.
	serverErrorRetryAttempts = 3

	// serverErrorRetryBackoff is the first wait; it doubles after each 5xx.
	serverErrorRetryBackoff = 250 * time.Millisecond

	// serverErrorRetryMaxWait caps one wait. A longer Retry-After cannot fit
	// a sync deadline, so the 5xx is returned to the caller instead.
	serverErrorRetryMaxWait = 5 * time.Second
)

// retryRoundTripper retries transient 5xx responses from identity providers.
// Google Workspace occasionally fails a directory read with 5xx, and the
// Admin SDK does not retry, so the status fails the whole sync.
type retryRoundTripper struct {
	next        http.RoundTripper
	attempts    int
	baseBackoff time.Duration
	maxWait     time.Duration
	// sleep waits between attempts. Nil uses a timer. Tests replace it.
	sleep func(ctx context.Context, wait time.Duration) error
}

func withServerErrorRetry(client *http.Client) *http.Client {
	cloned := *client
	cloned.Transport = &retryRoundTripper{
		next:        client.Transport,
		attempts:    serverErrorRetryAttempts,
		baseBackoff: serverErrorRetryBackoff,
		maxWait:     serverErrorRetryMaxWait,
	}

	return &cloned
}

func (rt *retryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	transport := rt.next
	if transport == nil {
		transport = http.DefaultTransport
	}

	attempts := rt.attempts
	if attempts <= 0 {
		attempts = serverErrorRetryAttempts
	}

	var lastResp *http.Response

	for attempt := range attempts {
		// Rewind onto a clone. Resending a consumed body makes the provider answer 400.
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

		resp, err := transport.RoundTrip(attemptReq)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode < http.StatusInternalServerError {
			return resp, nil
		}

		// Keep the body readable when this is the response we give up on.
		payload, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		resp.Body = io.NopCloser(bytes.NewReader(payload))
		lastResp = resp

		if attempt == attempts-1 {
			break
		}

		wait := rt.backoffFor(attempt)
		if after, ok := retryAfter(resp); ok {
			wait = after
		}

		if wait > rt.waitCap() {
			break
		}

		// Sleeping past the deadline hides the provider status behind a
		// context error.
		if deadline, ok := req.Context().Deadline(); ok && time.Until(deadline) <= wait {
			break
		}

		if err := rt.sleepFor(req.Context(), wait); err != nil {
			return nil, err
		}
	}

	return lastResp, nil
}

func (rt *retryRoundTripper) sleepFor(ctx context.Context, wait time.Duration) error {
	if rt.sleep != nil {
		return rt.sleep(ctx, wait)
	}

	timer := time.NewTimer(wait)

	select {
	case <-ctx.Done():
		timer.Stop()

		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (rt *retryRoundTripper) backoffFor(attempt int) time.Duration {
	step := rt.baseBackoff
	if step <= 0 {
		step = serverErrorRetryBackoff
	}

	return step << attempt
}

func (rt *retryRoundTripper) waitCap() time.Duration {
	if rt.maxWait > 0 {
		return rt.maxWait
	}

	return serverErrorRetryMaxWait
}

// retryAfter reads Retry-After as delta-seconds or an HTTP-date. A date
// already in the past yields a zero wait.
func retryAfter(resp *http.Response) (time.Duration, bool) {
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

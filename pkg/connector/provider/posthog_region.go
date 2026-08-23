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

// PostHog is the one provider whose data host is neither static nor stored: an
// OAuth connection authenticates through the region-agnostic oauth.posthog.com
// gateway, which does not serve /api, so the region has to be discovered with
// the token itself. That makes it endpoint knowledge rather than review logic,
// which is why it lives here and every consumer — the connection probe, the
// access-review driver — calls the same function.
package provider

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// PostHogOrganizationPath is the @current organization endpoint, the cheapest
// authenticated call PostHog offers.
const PostHogOrganizationPath = "/api/organizations/@current/"

const (
	// PostHog Cloud regional data hosts. API-key (us/eu) and self-hosted
	// connections always carry an explicit host instead.
	postHogUSBaseURL = "https://us.posthog.com"
	postHogEUBaseURL = "https://eu.posthog.com"
)

// ErrPostHogCredentialRejected reports that every PostHog Cloud region refused
// the token with 401/403 — a definitively dead or revoked credential, as
// opposed to a transient failure (5xx/network) on the token's own region. The
// connection probe uses it to tell a rejected credential apart from an
// inconclusive result, which must not flap the source to disconnected.
var ErrPostHogCredentialRejected = errors.New("posthog rejected the credential on every region")

// PostHogRegionBaseURLs are the hosts ResolvePostHogRegion probes, exposed so a
// test can pin the discovery order.
func PostHogRegionBaseURLs() []string {
	return []string{postHogUSBaseURL, postHogEUBaseURL}
}

// ResolvePostHogRegion probes the PostHog Cloud region hosts with the given
// token-bearing client and returns the first that answers 2xx on the @current
// organization endpoint.
//
// A token is valid on exactly one region; the other rejects it with 401/403.
// The result distinguishes the two failure classes the connection probe needs:
// ErrPostHogCredentialRejected when every region refused the token (dead or
// revoked credential), and a generic error when a region was merely unreachable
// or errored transiently, which must not be read as a rejection.
func ResolvePostHogRegion(ctx context.Context, client *http.Client) (string, error) {
	allRejected := true

	for _, host := range PostHogRegionBaseURLs() {
		endpoint, err := url.JoinPath(host, PostHogOrganizationPath)
		if err != nil {
			allRejected = false

			continue
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			allRejected = false

			continue
		}

		req.Header.Set("Accept", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			// Surface a cancelled/expired context as the real cause rather
			// than masking it behind "no region accepted the connection".
			if ctx.Err() != nil {
				return "", fmt.Errorf("cannot resolve posthog region: %w", ctx.Err())
			}

			allRejected = false

			continue
		}

		status := resp.StatusCode
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		if status >= http.StatusOK && status < http.StatusMultipleChoices {
			return host, nil
		}

		// Only 401/403 is a credential rejection; a 5xx/429 is transient and
		// leaves the verdict inconclusive rather than rejected.
		if status != http.StatusUnauthorized && status != http.StatusForbidden {
			allRejected = false
		}
	}

	if allRejected {
		return "", fmt.Errorf("cannot resolve posthog region: %w", ErrPostHogCredentialRejected)
	}

	return "", fmt.Errorf("cannot resolve posthog region: no region accepted the connection")
}

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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.gearno.de/kit/httpclient"
)

const (
	// vercelUserPath is the path both helpers below join onto a Vercel REST
	// API origin.
	vercelUserPath = "v2/user"

	// vercelDefaultBaseURL is the Vercel REST API root. Only
	// FetchVercelUserID uses it: that helper runs inside the OAuth callback
	// handler, which has no *provider.Registry — and therefore no resolved
	// Endpoints — in scope. FetchVercelUser takes its origin from the
	// caller instead.
	vercelDefaultBaseURL = "https://api.vercel.com"
)

// VercelUser is the projection of Vercel's /v2/user response that Probo
// consumes: the personal-account UID (used as a synthetic TeamID) and
// the human-readable display fields (used by the source-name resolver
// when a connector targets a personal account rather than a team).
type VercelUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
}

// FetchVercelUser calls Vercel's /v2/user under baseURL, the Vercel REST
// API origin (e.g. https://api.vercel.com), with the provided client. The
// client is expected to carry valid Bearer auth (either via an OAuth2
// round-tripper, as used by the source-name worker, or via a per-request
// header set by the caller).
func FetchVercelUser(ctx context.Context, client *http.Client, baseURL string) (VercelUser, error) {
	endpoint, err := url.JoinPath(baseURL, vercelUserPath)
	if err != nil {
		return VercelUser{}, fmt.Errorf("cannot build vercel user URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return VercelUser{}, fmt.Errorf("cannot create vercel user request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return VercelUser{}, fmt.Errorf("cannot execute vercel user request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return VercelUser{}, fmt.Errorf("cannot fetch vercel user: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		User VercelUser `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return VercelUser{}, fmt.Errorf("cannot decode vercel user response: %w", err)
	}

	return body.User, nil
}

// FetchVercelUserID is the OAuth-callback variant that builds its own
// one-shot SSRF-protected client and applies the freshly-minted access
// token as a Bearer header on the request. The OAuth callback handler
// uses the returned UID as a synthetic TeamID when the install targets
// a personal account (no teamId surfaced by the callback).
func FetchVercelUserID(ctx context.Context, accessToken string) (string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	endpoint, err := url.JoinPath(vercelDefaultBaseURL, vercelUserPath)
	if err != nil {
		return "", fmt.Errorf("cannot build vercel user URL: %w", err)
	}

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("cannot create vercel user request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := httpclient.DefaultClient(httpclient.WithSSRFProtection())

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot execute vercel user request: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("cannot fetch vercel user: unexpected status %d", resp.StatusCode)
	}

	var body struct {
		User VercelUser `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("cannot decode vercel user response: %w", err)
	}

	return body.User.ID, nil
}

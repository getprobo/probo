// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.gearno.de/kit/httpclient"
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

// FetchVercelUser calls Vercel's /v2/user with the provided client. The
// client is expected to carry valid Bearer auth (either via an OAuth2
// round-tripper, as used by the source-name worker, or via a per-request
// header set by the caller).
func FetchVercelUser(ctx context.Context, client *http.Client) (VercelUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.vercel.com/v2/user", nil)
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
// a personal account (no team_id surfaced by the callback).
func FetchVercelUserID(ctx context.Context, accessToken string) (string, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, "https://api.vercel.com/v2/user", nil)
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

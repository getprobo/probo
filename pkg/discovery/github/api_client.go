// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	githubAPIBase   = "https://api.github.com"
	maxReposToScan  = 30
	maxPagesPerList = 10
)

type apiClient struct {
	httpClient *http.Client
}

func newAPIClient(httpClient *http.Client) *apiClient {
	return &apiClient{httpClient: httpClient}
}

func (c *apiClient) getJSON(
	ctx context.Context,
	endpoint string,
	dest any,
) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return 0, fmt.Errorf("cannot create github request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("cannot execute github request: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, resp.Body)

		return resp.StatusCode, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return resp.StatusCode, fmt.Errorf("cannot decode github response: %w", err)
	}

	return resp.StatusCode, nil
}

func (c *apiClient) getPaginated(
	ctx context.Context,
	firstURL string,
	dest any,
) (string, error) {
	next := firstURL

	for page := 0; page < maxPagesPerList && next != ""; page++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, next, nil)
		if err != nil {
			return next, fmt.Errorf("cannot create github request: %w", err)
		}

		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return next, fmt.Errorf("cannot execute github request: %w", err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			_ = resp.Body.Close()

			return next, fmt.Errorf("unexpected status %d", resp.StatusCode)
		}

		if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
			_ = resp.Body.Close()

			return next, fmt.Errorf("cannot decode github response: %w", err)
		}

		next = parseLinkNext(resp.Header.Get("Link"))
		_ = resp.Body.Close()
	}

	return next, nil
}

func (c *apiClient) orgEndpoint(org string, parts ...string) (string, error) {
	segments := append([]string{"orgs", url.PathEscape(org)}, parts...)

	return url.JoinPath(githubAPIBase, segments...)
}

func (c *apiClient) repoEndpoint(owner, repo string, parts ...string) (string, error) {
	segments := append(
		[]string{"repos", url.PathEscape(owner), url.PathEscape(repo)},
		parts...,
	)

	return url.JoinPath(githubAPIBase, segments...)
}

func withPerPage(endpoint string, perPage int) (string, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("cannot parse endpoint: %w", err)
	}

	q := parsed.Query()
	q.Set("per_page", fmt.Sprintf("%d", perPage))
	parsed.RawQuery = q.Encode()

	return parsed.String(), nil
}

func parseLinkNext(linkHeader string) string {
	if linkHeader == "" {
		return ""
	}

	for part := range strings.SplitSeq(linkHeader, ",") {
		part = strings.TrimSpace(part)
		if !strings.Contains(part, `rel="next"`) {
			continue
		}

		start := strings.Index(part, "<")
		end := strings.Index(part, ">")

		if start >= 0 && end > start {
			return part[start+1 : end]
		}
	}

	return ""
}

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

package console_v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.probo.inc/probo/pkg/server/api/console/v1/types"
)

// fetchGitHubOrganizations fetches the list of organizations the
// authenticated GitHub user belongs to.
func fetchGitHubOrganizations(ctx context.Context, httpClient *http.Client) ([]*types.ProviderOrganization, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/orgs", nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create github organizations request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch github organizations: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch github organizations: status %d", resp.StatusCode)
	}

	var orgs []struct {
		Login string `json:"login"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, fmt.Errorf("cannot decode github organizations response: %w", err)
	}

	result := make([]*types.ProviderOrganization, len(orgs))
	for i, org := range orgs {
		displayName := org.Name
		if displayName == "" {
			displayName = org.Login
		}
		result[i] = &types.ProviderOrganization{
			Slug:        org.Login,
			DisplayName: displayName,
		}
	}

	return result, nil
}

// fetchSentryOrganizations fetches the list of organizations the
// authenticated Sentry user belongs to.
func fetchSentryOrganizations(ctx context.Context, httpClient *http.Client) ([]*types.ProviderOrganization, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://sentry.io/api/0/organizations/?member=true", nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create sentry organizations request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch sentry organizations: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch sentry organizations: status %d", resp.StatusCode)
	}

	var orgs []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, fmt.Errorf("cannot decode sentry organizations response: %w", err)
	}

	result := make([]*types.ProviderOrganization, len(orgs))
	for i, org := range orgs {
		displayName := org.Name
		if displayName == "" {
			displayName = org.Slug
		}
		result[i] = &types.ProviderOrganization{
			Slug:        org.Slug,
			DisplayName: displayName,
		}
	}

	return result, nil
}

// probeConnection makes a lightweight API call to the given URL to verify
// the OAuth token is still valid. The probe URL is configured per connector
// in the connector registry.
func probeConnection(ctx context.Context, httpClient *http.Client, probeURL string) error {
	if probeURL == "" {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, probeURL, nil)
	if err != nil {
		return fmt.Errorf("cannot create probe request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("probe request failed: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("token rejected: status %d", resp.StatusCode)
	}

	return nil
}

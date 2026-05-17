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
	"strconv"

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

// fetchGitLabOrganizations fetches the list of groups the authenticated
// GitLab user owns. Group IDs are numeric int64 values; we surface them
// as strings so they fit the ProviderOrganization.Slug shape.
func fetchGitLabOrganizations(ctx context.Context, httpClient *http.Client) ([]*types.ProviderOrganization, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://gitlab.com/api/v4/groups?min_access_level=50&per_page=100",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create gitlab organizations request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch gitlab organizations: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch gitlab organizations: status %d", resp.StatusCode)
	}

	var groups []struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		FullPath string `json:"full_path"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, fmt.Errorf("cannot decode gitlab organizations response: %w", err)
	}

	result := make([]*types.ProviderOrganization, len(groups))
	for i, g := range groups {
		displayName := g.Name
		if displayName == "" {
			displayName = g.FullPath
		}
		result[i] = &types.ProviderOrganization{
			Slug:        strconv.FormatInt(g.ID, 10),
			DisplayName: displayName,
		}
	}

	return result, nil
}

// fetchBitbucketOrganizations fetches the list of workspaces the
// authenticated Bitbucket user belongs to.
func fetchBitbucketOrganizations(ctx context.Context, httpClient *http.Client) ([]*types.ProviderOrganization, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.bitbucket.org/2.0/workspaces?role=member&pagelen=100",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create bitbucket organizations request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch bitbucket organizations: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch bitbucket organizations: status %d", resp.StatusCode)
	}

	var body struct {
		Values []struct {
			Slug string `json:"slug"`
			Name string `json:"name"`
		} `json:"values"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("cannot decode bitbucket organizations response: %w", err)
	}

	result := make([]*types.ProviderOrganization, len(body.Values))
	for i, w := range body.Values {
		displayName := w.Name
		if displayName == "" {
			displayName = w.Slug
		}
		result[i] = &types.ProviderOrganization{
			Slug:        w.Slug,
			DisplayName: displayName,
		}
	}

	return result, nil
}

// fetchHerokuOrganizations fetches the list of teams the authenticated
// Heroku user belongs to.
func fetchHerokuOrganizations(ctx context.Context, httpClient *http.Client) ([]*types.ProviderOrganization, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.heroku.com/teams", nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create heroku organizations request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch heroku organizations: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch heroku organizations: status %d", resp.StatusCode)
	}

	var teams []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&teams); err != nil {
		return nil, fmt.Errorf("cannot decode heroku organizations response: %w", err)
	}

	result := make([]*types.ProviderOrganization, len(teams))
	for i, t := range teams {
		displayName := t.Name
		if displayName == "" {
			displayName = t.ID
		}
		result[i] = &types.ProviderOrganization{
			Slug:        t.ID,
			DisplayName: displayName,
		}
	}

	return result, nil
}

// fetchAsanaOrganizations fetches the list of workspaces the
// authenticated Asana user belongs to.
func fetchAsanaOrganizations(ctx context.Context, httpClient *http.Client) ([]*types.ProviderOrganization, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://app.asana.com/api/1.0/workspaces?limit=100",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create asana organizations request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch asana organizations: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch asana organizations: status %d", resp.StatusCode)
	}

	var body struct {
		Data []struct {
			GID  string `json:"gid"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("cannot decode asana organizations response: %w", err)
	}

	result := make([]*types.ProviderOrganization, len(body.Data))
	for i, w := range body.Data {
		displayName := w.Name
		if displayName == "" {
			displayName = w.GID
		}
		result[i] = &types.ProviderOrganization{
			Slug:        w.GID,
			DisplayName: displayName,
		}
	}

	return result, nil
}

// fetchSnykOrganizations fetches the list of Snyk organizations the
// authenticated user belongs to. The Snyk REST API is JSON:API style;
// the org id is the unique identifier surfaced as the slug.
func fetchSnykOrganizations(ctx context.Context, httpClient *http.Client) ([]*types.ProviderOrganization, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.snyk.io/rest/orgs?version=2024-10-15&limit=100",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create snyk organizations request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.api+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch snyk organizations: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch snyk organizations: status %d", resp.StatusCode)
	}

	var body struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				Slug string `json:"slug"`
				Name string `json:"name"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("cannot decode snyk organizations response: %w", err)
	}

	result := make([]*types.ProviderOrganization, len(body.Data))
	for i, org := range body.Data {
		displayName := org.Attributes.Name
		if displayName == "" {
			displayName = org.Attributes.Slug
		}
		if displayName == "" {
			displayName = org.ID
		}
		result[i] = &types.ProviderOrganization{
			Slug:        org.ID,
			DisplayName: displayName,
		}
	}

	return result, nil
}

// fetchNetlifyOrganizations fetches the list of Netlify accounts the
// authenticated user belongs to.
func fetchNetlifyOrganizations(ctx context.Context, httpClient *http.Client) ([]*types.ProviderOrganization, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.netlify.com/api/v1/accounts",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create netlify organizations request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch netlify organizations: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch netlify organizations: status %d", resp.StatusCode)
	}

	var accounts []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, fmt.Errorf("cannot decode netlify organizations response: %w", err)
	}

	result := make([]*types.ProviderOrganization, len(accounts))
	for i, a := range accounts {
		displayName := a.Name
		if displayName == "" {
			displayName = a.Slug
		}
		result[i] = &types.ProviderOrganization{
			Slug:        a.Slug,
			DisplayName: displayName,
		}
	}

	return result, nil
}

// fetchClickUpOrganizations fetches the list of ClickUp teams (workspaces)
// the authenticated user belongs to.
func fetchClickUpOrganizations(ctx context.Context, httpClient *http.Client) ([]*types.ProviderOrganization, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://api.clickup.com/api/v2/team",
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create clickup organizations request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch clickup organizations: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cannot fetch clickup organizations: status %d", resp.StatusCode)
	}

	var body struct {
		Teams []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"teams"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("cannot decode clickup organizations response: %w", err)
	}

	result := make([]*types.ProviderOrganization, len(body.Teams))
	for i, t := range body.Teams {
		displayName := t.Name
		if displayName == "" {
			displayName = t.ID
		}
		result[i] = &types.ProviderOrganization{
			Slug:        t.ID,
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

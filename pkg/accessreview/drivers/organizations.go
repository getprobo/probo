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

package drivers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// Organization represents a tenant/workspace/team/group surfaced by a
// provider's "list orgs the authenticated user can access" endpoint.
// The OAuth picker UI consumes this to let the user choose which one
// scopes the access source.
type Organization struct {
	Slug        string
	DisplayName string
}

// ListGitHubOrganizations fetches the organizations the authenticated
// GitHub user belongs to.
func ListGitHubOrganizations(ctx context.Context, httpClient *http.Client) ([]Organization, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/orgs?per_page=100", nil)
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
		return nil, fmt.Errorf("cannot fetch github organizations: unexpected status %d", resp.StatusCode)
	}

	var orgs []struct {
		Login string `json:"login"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, fmt.Errorf("cannot decode github organizations response: %w", err)
	}

	result := make([]Organization, len(orgs))
	for i, org := range orgs {
		displayName := org.Name
		if displayName == "" {
			displayName = org.Login
		}

		result[i] = Organization{Slug: org.Login, DisplayName: displayName}
	}

	return result, nil
}

// ListSentryOrganizations fetches the organizations the authenticated
// Sentry user belongs to.
func ListSentryOrganizations(ctx context.Context, httpClient *http.Client) ([]Organization, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://sentry.io/api/0/organizations/?member=true",
		nil,
	)
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
		return nil, fmt.Errorf("cannot fetch sentry organizations: unexpected status %d", resp.StatusCode)
	}

	var orgs []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orgs); err != nil {
		return nil, fmt.Errorf("cannot decode sentry organizations response: %w", err)
	}

	result := make([]Organization, len(orgs))
	for i, org := range orgs {
		displayName := org.Name
		if displayName == "" {
			displayName = org.Slug
		}

		result[i] = Organization{Slug: org.Slug, DisplayName: displayName}
	}

	return result, nil
}

// ListGitLabOrganizations fetches the GitLab groups the authenticated
// user owns. Group IDs are int64; we surface them as strings so they fit
// the Organization.Slug shape.
func ListGitLabOrganizations(ctx context.Context, httpClient *http.Client) ([]Organization, error) {
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
		return nil, fmt.Errorf("cannot fetch gitlab organizations: unexpected status %d", resp.StatusCode)
	}

	var groups []struct {
		ID       int64  `json:"id"`
		Name     string `json:"name"`
		FullPath string `json:"full_path"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&groups); err != nil {
		return nil, fmt.Errorf("cannot decode gitlab organizations response: %w", err)
	}

	result := make([]Organization, len(groups))
	for i, g := range groups {
		displayName := g.Name
		if displayName == "" {
			displayName = g.FullPath
		}

		result[i] = Organization{
			Slug:        strconv.FormatInt(g.ID, 10),
			DisplayName: displayName,
		}
	}

	return result, nil
}

// ListBitbucketOrganizations fetches the workspaces the authenticated
// Bitbucket user belongs to. The legacy /2.0/workspaces endpoint was
// sunset by CHANGE-2770 (April 2026); /2.0/user/workspaces is the
// supported cross-workspace replacement (CHANGE-3022). Bitbucket pages
// via an absolute `next` URL on each response; follow until exhausted.
func ListBitbucketOrganizations(ctx context.Context, httpClient *http.Client) ([]Organization, error) {
	pageURL := "https://api.bitbucket.org/2.0/user/workspaces?pagelen=100"
	result := make([]Organization, 0)

	for range maxPaginationPages {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
		if err != nil {
			return nil, fmt.Errorf("cannot create bitbucket organizations request: %w", err)
		}

		req.Header.Set("Accept", "application/json")

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("cannot fetch bitbucket organizations: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("cannot fetch bitbucket organizations: unexpected status %d", resp.StatusCode)
		}

		// We tolerate both shapes (flat and nested under `workspace`) since
		// Atlassian has shipped variants of similar endpoints with both.
		var body struct {
			Values []struct {
				Slug      string `json:"slug"`
				Name      string `json:"name"`
				Workspace struct {
					Slug string `json:"slug"`
					Name string `json:"name"`
				} `json:"workspace"`
			} `json:"values"`
			Next string `json:"next"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("cannot decode bitbucket organizations response: %w", err)
		}

		_ = resp.Body.Close()

		for _, v := range body.Values {
			slug, name := v.Slug, v.Name
			if slug == "" {
				slug = v.Workspace.Slug
				name = v.Workspace.Name
			}

			displayName := name
			if displayName == "" {
				displayName = slug
			}

			result = append(result, Organization{Slug: slug, DisplayName: displayName})
		}

		if body.Next == "" {
			return result, nil
		}

		pageURL = body.Next
	}

	return nil, fmt.Errorf("cannot list all bitbucket organizations: %w", ErrPaginationLimitReached)
}

// ListHerokuOrganizations fetches the teams the authenticated Heroku
// user belongs to.
func ListHerokuOrganizations(ctx context.Context, httpClient *http.Client) ([]Organization, error) {
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
		return nil, fmt.Errorf("cannot fetch heroku organizations: unexpected status %d", resp.StatusCode)
	}

	var teams []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&teams); err != nil {
		return nil, fmt.Errorf("cannot decode heroku organizations response: %w", err)
	}

	result := make([]Organization, len(teams))
	for i, t := range teams {
		displayName := t.Name
		if displayName == "" {
			displayName = t.ID
		}

		result[i] = Organization{Slug: t.ID, DisplayName: displayName}
	}

	return result, nil
}

// ListAsanaOrganizations fetches the workspaces the authenticated Asana
// user belongs to.
func ListAsanaOrganizations(ctx context.Context, httpClient *http.Client) ([]Organization, error) {
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
		return nil, fmt.Errorf("cannot fetch asana organizations: unexpected status %d", resp.StatusCode)
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

	result := make([]Organization, len(body.Data))
	for i, w := range body.Data {
		displayName := w.Name
		if displayName == "" {
			displayName = w.GID
		}

		result[i] = Organization{Slug: w.GID, DisplayName: displayName}
	}

	return result, nil
}

// ListNetlifyOrganizations fetches the Netlify accounts the authenticated
// user belongs to.
func ListNetlifyOrganizations(ctx context.Context, httpClient *http.Client) ([]Organization, error) {
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
		return nil, fmt.Errorf("cannot fetch netlify organizations: unexpected status %d", resp.StatusCode)
	}

	var accounts []struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, fmt.Errorf("cannot decode netlify organizations response: %w", err)
	}

	result := make([]Organization, len(accounts))
	for i, a := range accounts {
		displayName := a.Name
		if displayName == "" {
			displayName = a.Slug
		}

		result[i] = Organization{Slug: a.Slug, DisplayName: displayName}
	}

	return result, nil
}

// ListClickUpOrganizations fetches the ClickUp teams (workspaces) the
// authenticated user belongs to.
func ListClickUpOrganizations(ctx context.Context, httpClient *http.Client) ([]Organization, error) {
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
		return nil, fmt.Errorf("cannot fetch clickup organizations: unexpected status %d", resp.StatusCode)
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

	result := make([]Organization, len(body.Teams))
	for i, t := range body.Teams {
		displayName := t.Name
		if displayName == "" {
			displayName = t.ID
		}

		result[i] = Organization{Slug: t.ID, DisplayName: displayName}
	}

	return result, nil
}

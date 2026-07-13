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
	"fmt"
	"net/http"
	"time"

	"github.com/google/go-github/v69/github"
)

type restClient struct {
	httpClient *http.Client
	gh         *github.Client
}

func newRESTClient(httpClient *http.Client) *restClient {
	return &restClient{
		httpClient: httpClient,
		gh:         github.NewClient(httpClient),
	}
}

func (c *restClient) HTTPClient() *http.Client {
	return c.httpClient
}

func (c *restClient) listOrgRepositories(ctx context.Context, org string) ([]repoListItem, error) {
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	repos := make([]repoListItem, 0, 100)

	for page := 0; page < maxPagesPerList; page++ {
		ghRepos, resp, err := c.gh.Repositories.ListByOrg(ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("cannot list github repos: %w", err)
		}

		for _, ghRepo := range ghRepos {
			repos = append(repos, repoFromGitHub(ghRepo))
		}

		if resp == nil || resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return repos, nil
}

func (c *restClient) getOrganization(ctx context.Context, org string) (*githubOrganization, error) {
	ghOrg, _, err := c.gh.Organizations.Get(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch github organization: %w", err)
	}

	return organizationFromGitHub(ghOrg), nil
}

func (c *restClient) getBranchProtection(
	ctx context.Context,
	owner, repo, branch string,
) (*branchProtection, bool) {
	protection, resp, err := c.gh.Repositories.GetBranchProtection(ctx, owner, repo, branch)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, false
		}

		return nil, false
	}

	return branchProtectionFromGitHub(protection), true
}

func (c *restClient) listWorkflowCount(ctx context.Context, owner, repo string) (int, bool) {
	page, resp, err := c.gh.Actions.ListWorkflows(ctx, owner, repo, &github.ListOptions{PerPage: 1})
	if err != nil {
		return 0, false
	}

	if page == nil {
		return 0, false
	}

	_ = resp

	return page.GetTotalCount(), page.GetTotalCount() > 0
}

func (c *restClient) searchCodePaths(ctx context.Context, query string) ([]string, error) {
	opts := &github.SearchOptions{ListOptions: github.ListOptions{PerPage: 100}}

	var paths []string

	for page := 0; page < maxPagesPerList; page++ {
		result, resp, err := c.gh.Search.Code(ctx, query, opts)
		if err != nil {
			if resp != nil && (resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnprocessableEntity) {
				return paths, fmt.Errorf("github code search unavailable: %w", err)
			}

			return paths, fmt.Errorf("cannot search github code: %w", err)
		}

		for _, item := range result.CodeResults {
			if item.Repository == nil || item.Path == nil {
				continue
			}

			paths = append(paths, fmt.Sprintf("%s/%s", item.Repository.GetName(), item.GetPath()))
		}

		if resp == nil || resp.NextPage == 0 {
			break
		}

		opts.Page = resp.NextPage
	}

	return paths, nil
}

func repoFromGitHub(repo *github.Repository) repoListItem {
	if repo == nil {
		return repoListItem{}
	}

	item := repoListItem{
		Name:            repo.GetName(),
		DefaultBranch:   repo.GetDefaultBranch(),
		Private:         repo.GetPrivate(),
		Archived:        repo.GetArchived(),
		Disabled:        repo.GetDisabled(),
		Fork:            repo.GetFork(),
		Size:            repo.GetSize(),
		Description:     repo.GetDescription(),
		Language:        repo.GetLanguage(),
		Topics:          append([]string(nil), repo.Topics...),
		StargazersCount: repo.GetStargazersCount(),
		ForksCount:      repo.GetForksCount(),
		OpenIssuesCount: repo.GetOpenIssuesCount(),
	}

	if repo.PushedAt != nil {
		item.PushedAt = repo.PushedAt.Format(time.RFC3339)
	}

	return item
}

func organizationFromGitHub(org *github.Organization) *githubOrganization {
	if org == nil {
		return &githubOrganization{}
	}

	return &githubOrganization{
		TwoFactorRequirementEnabled:        org.TwoFactorRequirementEnabled,
		DefaultRepositoryPermission:        org.GetDefaultRepoPermission(),
		MembersCanCreatePublicRepositories: org.MembersCanCreatePublicRepos,
	}
}

func branchProtectionFromGitHub(protection *github.Protection) *branchProtection {
	if protection == nil {
		return nil
	}

	mapped := &branchProtection{}

	if reviews := protection.RequiredPullRequestReviews; reviews != nil {
		mapped.RequiredPullRequestReviews = &struct {
			RequiredApprovingReviewCount int  `json:"required_approving_review_count"`
			DismissStaleReviews          bool `json:"dismiss_stale_reviews"`
		}{
			RequiredApprovingReviewCount: reviews.RequiredApprovingReviewCount,
			DismissStaleReviews:          reviews.DismissStaleReviews,
		}
	}

	if signatures := protection.RequiredSignatures; signatures != nil {
		mapped.RequiredSignatures = &struct {
			Enabled bool `json:"enabled"`
		}{
			Enabled: signatures.GetEnabled(),
		}
	}

	if forcePushes := protection.AllowForcePushes; forcePushes != nil {
		mapped.AllowForcePushes = &struct {
			Enabled bool `json:"enabled"`
		}{
			Enabled: forcePushes.Enabled,
		}
	}

	if checks := protection.RequiredStatusChecks; checks != nil {
		mapped.RequiredStatusChecks = &struct {
			Strict   bool     `json:"strict"`
			Contexts []string `json:"contexts"`
		}{
			Strict:   checks.Strict,
			Contexts: append([]string(nil), checks.GetContexts()...),
		}
	}

	if admins := protection.EnforceAdmins; admins != nil {
		mapped.EnforceAdmins = &struct {
			Enabled bool `json:"enabled"`
		}{
			Enabled: admins.Enabled,
		}
	}

	if restrictions := protection.Restrictions; restrictions != nil {
		mapped.Restrictions = &struct {
			Users []any `json:"users"`
			Teams []any `json:"teams"`
			Apps  []any `json:"apps"`
		}{
			Users: usersToAny(restrictions.Users),
			Teams: teamsToAny(restrictions.Teams),
			Apps:  appsToAny(restrictions.Apps),
		}
	}

	return mapped
}

func usersToAny(users []*github.User) []any {
	if len(users) == 0 {
		return nil
	}

	out := make([]any, 0, len(users))
	for _, user := range users {
		if user == nil {
			continue
		}

		out = append(out, user.GetLogin())
	}

	return out
}

func teamsToAny(teams []*github.Team) []any {
	if len(teams) == 0 {
		return nil
	}

	out := make([]any, 0, len(teams))
	for _, team := range teams {
		if team == nil {
			continue
		}

		out = append(out, team.GetSlug())
	}

	return out
}

func appsToAny(apps []*github.App) []any {
	if len(apps) == 0 {
		return nil
	}

	out := make([]any, 0, len(apps))
	for _, app := range apps {
		if app == nil {
			continue
		}

		out = append(out, app.GetSlug())
	}

	return out
}

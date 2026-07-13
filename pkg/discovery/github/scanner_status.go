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
	"strings"
)

type (
	branchHeadResponse struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	combinedStatusResponse struct {
		Statuses []struct {
			Context   string `json:"context"`
			TargetURL string `json:"target_url"`
		} `json:"statuses"`
	}

	checkRunsResponse struct {
		CheckRuns []struct {
			Name       string `json:"name"`
			DetailsURL string `json:"details_url"`
			App        struct {
				Slug string `json:"slug"`
			} `json:"app"`
		} `json:"check_runs"`
	}

	ciProviderAggregate struct {
		ReposWithCommitStatusCI int
		ReposWithExternalCI     int
		Providers               map[string]int
	}

	ciProviderCatalog struct {
		Name     string
		Patterns []string
	}
)

var ciProviders = []ciProviderCatalog{
	{Name: "github_actions", Patterns: []string{"github-actions"}},
	{Name: "circleci", Patterns: []string{"ci/circleci", "circleci.com"}},
	{Name: "jenkins", Patterns: []string{"jenkins", "continuous-integration/jenkins"}},
	{Name: "drone", Patterns: []string{"drone", "drone.io"}},
	{Name: "travis_ci", Patterns: []string{"travis-ci", "travis-ci.com"}},
	{Name: "buildkite", Patterns: []string{"buildkite.com", "buildkite"}},
	{Name: "azure_pipelines", Patterns: []string{"azure-pipelines", "dev.azure.com"}},
	{Name: "gitlab_ci", Patterns: []string{"gitlab-ci"}},
	{Name: "teamcity", Patterns: []string{"teamcity"}},
}

func detectCIProviders(parts ...string) []string {
	combined := strings.ToLower(strings.Join(parts, " "))
	found := make([]string, 0, 2)

	for _, provider := range ciProviders {
		for _, pattern := range provider.Patterns {
			if strings.Contains(combined, strings.ToLower(pattern)) {
				found = append(found, provider.Name)

				break
			}
		}
	}

	return found
}

func isExternalCIProvider(name string) bool {
	return name != "" && name != "github_actions"
}

func (s *discoveryScanner) scanRepoCIStatuses(
	ctx context.Context,
	repo repoListItem,
	agg *repoScanAggregate,
	ciAgg *ciProviderAggregate,
) {
	sha, ok := s.fetchDefaultBranchSHA(ctx, repo)
	if !ok {
		return
	}

	providers := s.fetchCIProvidersForCommit(ctx, repo, sha)
	if len(providers) == 0 {
		return
	}

	agg.WithCommitStatusCI++
	ciAgg.ReposWithCommitStatusCI++

	for _, provider := range providers {
		ciAgg.Providers[provider]++
	}

	for _, provider := range providers {
		if isExternalCIProvider(provider) {
			agg.WithExternalCI++
			ciAgg.ReposWithExternalCI++

			break
		}
	}
}

func (s *discoveryScanner) fetchDefaultBranchSHA(
	ctx context.Context,
	repo repoListItem,
) (string, bool) {
	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, "branches", repo.DefaultBranch)
	if err != nil {
		return "", false
	}

	var branch branchHeadResponse

	if _, err := s.api.getJSON(ctx, endpoint, &branch); err != nil {
		return "", false
	}

	if branch.Commit.SHA == "" {
		return "", false
	}

	return branch.Commit.SHA, true
}

func (s *discoveryScanner) fetchCIProvidersForCommit(
	ctx context.Context,
	repo repoListItem,
	sha string,
) []string {
	found := map[string]struct{}{}

	for _, provider := range s.fetchCIProvidersFromStatuses(ctx, repo, sha) {
		found[provider] = struct{}{}
	}

	for _, provider := range s.fetchCIProvidersFromCheckRuns(ctx, repo, sha) {
		found[provider] = struct{}{}
	}

	out := make([]string, 0, len(found))
	for provider := range found {
		out = append(out, provider)
	}

	return out
}

func (s *discoveryScanner) fetchCIProvidersFromStatuses(
	ctx context.Context,
	repo repoListItem,
	sha string,
) []string {
	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, "commits", sha, "status")
	if err != nil {
		return nil
	}

	var status combinedStatusResponse

	if _, err := s.api.getJSON(ctx, endpoint, &status); err != nil {
		return nil
	}

	found := map[string]struct{}{}

	for _, item := range status.Statuses {
		for _, provider := range detectCIProviders(item.Context, item.TargetURL) {
			found[provider] = struct{}{}
		}
	}

	return mapKeys(found)
}

func (s *discoveryScanner) fetchCIProvidersFromCheckRuns(
	ctx context.Context,
	repo repoListItem,
	sha string,
) []string {
	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, "commits", sha, "check-runs")
	if err != nil {
		return nil
	}

	endpoint, err = withPerPage(endpoint, 100)
	if err != nil {
		return nil
	}

	var runs checkRunsResponse

	if _, err := s.api.getJSON(ctx, endpoint, &runs); err != nil {
		return nil
	}

	found := map[string]struct{}{}

	for _, run := range runs.CheckRuns {
		for _, provider := range detectCIProviders(run.Name, run.App.Slug, run.DetailsURL) {
			found[provider] = struct{}{}
		}
	}

	return mapKeys(found)
}

func mapKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))

	for key := range m {
		out = append(out, key)
	}

	return out
}

func ciProviderFact(ciAgg *ciProviderAggregate, reposScanned int) Fact {
	providers := make(map[string]int, len(ciAgg.Providers))
	for name, count := range ciAgg.Providers {
		providers[name] = count
	}

	return Fact{
		FactID:  "f-ci-providers",
		FactKey: "repo_ci_providers",
		Scope:   "org",
		Value: map[string]any{
			"providers":                providers,
			"repos_with_commit_status": ciAgg.ReposWithCommitStatusCI,
			"repos_with_external_ci":   ciAgg.ReposWithExternalCI,
			"repos_scanned":            reposScanned,
		},
		APIRef: "GET /repos/{owner}/{repo}/commits/{sha}/status; GET /repos/{owner}/{repo}/commits/{sha}/check-runs",
	}
}

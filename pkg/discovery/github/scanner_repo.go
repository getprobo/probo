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
	"net/url"
)

type (
	repoListItem struct {
		Name          string `json:"name"`
		DefaultBranch string `json:"default_branch"`
		Private       bool   `json:"private"`
		Archived      bool   `json:"archived"`
		Disabled      bool   `json:"disabled"`
	}

	branchProtection struct {
		RequiredPullRequestReviews *struct {
			RequiredApprovingReviewCount int `json:"required_approving_review_count"`
		} `json:"required_pull_request_reviews"`
		RequiredSignatures *struct {
			Enabled bool `json:"enabled"`
		} `json:"required_signatures"`
		AllowForcePushes *struct {
			Enabled bool `json:"enabled"`
		} `json:"allow_force_pushes"`
	}

	workflowsPage struct {
		TotalCount int `json:"total_count"`
	}

	alertPage struct {
		TotalCount int `json:"total_count"`
	}

	repoScanAggregate struct {
		TotalRepos               int
		PublicRepos              int
		ScannedRepos             int
		WithBranchProtection     int
		WithRequiredReviews      int
		WithSignedCommits        int
		WithWorkflows            int
		WithSecurityMD           int
		WithContributingMD       int
		WithDependabotConfig     int
		DependabotCriticalOpen   int
		SecretScanningOpen       int
		CodeScanningCriticalOpen int
	}
)

func (s *discoveryScanner) scanRepos(ctx context.Context, sheet *FactSheet) {
	repos, err := s.listOrgRepos(ctx)
	if err != nil {
		sheet.Limitations = append(sheet.Limitations, "cannot list organization repositories")

		return
	}

	agg := aggregateRepoList(repos)
	sheet.ReposScanned = agg.TotalRepos

	sheet.Facts = append(sheet.Facts, Fact{
		FactID:  "f-public-repos",
		FactKey: "org_public_repos",
		Scope:   "org",
		Value: map[string]int{
			"public": agg.PublicRepos,
			"total":  agg.TotalRepos,
		},
		APIRef: "GET /orgs/{org}/repos",
	})

	if !s.scopes.hasRepoRead() {
		sheet.Limitations = append(
			sheet.Limitations,
			"repo scope not granted; skipping per-repository discovery checks",
		)

		return
	}

	for _, repo := range selectReposToScan(repos) {
		s.scanRepo(ctx, repo, agg)
	}

	sheet.Facts = append(sheet.Facts, repoAggregateFacts(agg)...)
}

func (s *discoveryScanner) listOrgRepos(ctx context.Context) ([]repoListItem, error) {
	endpoint, err := s.api.orgEndpoint(s.org, "repos")
	if err != nil {
		return nil, fmt.Errorf("cannot build github repos URL: %w", err)
	}

	endpoint, err = withPerPage(endpoint, 100)
	if err != nil {
		return nil, fmt.Errorf("cannot build github repos URL: %w", err)
	}

	var repos []repoListItem

	if _, err := s.api.getPaginated(ctx, endpoint, &repos); err != nil {
		return nil, fmt.Errorf("cannot list github repos: %w", err)
	}

	return repos, nil
}

func selectReposToScan(repos []repoListItem) []repoListItem {
	selected := make([]repoListItem, 0, maxReposToScan)

	for _, repo := range repos {
		if repo.Archived || repo.Disabled || repo.DefaultBranch == "" {
			continue
		}

		selected = append(selected, repo)

		if len(selected) >= maxReposToScan {
			break
		}
	}

	return selected
}

func aggregateRepoList(repos []repoListItem) *repoScanAggregate {
	agg := &repoScanAggregate{TotalRepos: len(repos)}

	for _, repo := range repos {
		if !repo.Private {
			agg.PublicRepos++
		}
	}

	return agg
}

func (s *discoveryScanner) scanRepo(ctx context.Context, repo repoListItem, agg *repoScanAggregate) {
	agg.ScannedRepos++

	if s.probeBranchProtection(ctx, repo) {
		agg.WithBranchProtection++
	}

	if s.probeRequiredReviews(ctx, repo) {
		agg.WithRequiredReviews++
	}

	if s.probeSignedCommitsRequired(ctx, repo) {
		agg.WithSignedCommits++
	}

	if s.probeWorkflows(ctx, repo) {
		agg.WithWorkflows++
	}

	if s.probeRepoFile(ctx, repo, "SECURITY.md") {
		agg.WithSecurityMD++
	}

	if s.probeRepoFile(ctx, repo, "CONTRIBUTING.md") {
		agg.WithContributingMD++
	}

	if s.probeRepoFile(ctx, repo, ".github", "dependabot.yml") {
		agg.WithDependabotConfig++
	}

	if s.scopes.hasSecurityEvents() {
		agg.DependabotCriticalOpen += s.countDependabotCritical(ctx, repo)
		agg.SecretScanningOpen += s.countSecretScanningOpen(ctx, repo)
		agg.CodeScanningCriticalOpen += s.countCodeScanningCritical(ctx, repo)
	}
}

func (s *discoveryScanner) probeBranchProtection(ctx context.Context, repo repoListItem) bool {
	endpoint, err := s.api.repoEndpoint(
		s.org,
		repo.Name,
		"branches",
		repo.DefaultBranch,
		"protection",
	)
	if err != nil {
		return false
	}

	var protection branchProtection

	if _, err := s.api.getJSON(ctx, endpoint, &protection); err != nil {
		return false
	}

	return true
}

func (s *discoveryScanner) probeRequiredReviews(ctx context.Context, repo repoListItem) bool {
	endpoint, err := s.api.repoEndpoint(
		s.org,
		repo.Name,
		"branches",
		repo.DefaultBranch,
		"protection",
	)
	if err != nil {
		return false
	}

	var protection branchProtection

	if _, err := s.api.getJSON(ctx, endpoint, &protection); err != nil {
		return false
	}

	return protection.RequiredPullRequestReviews != nil &&
		protection.RequiredPullRequestReviews.RequiredApprovingReviewCount > 0
}

func (s *discoveryScanner) probeSignedCommitsRequired(ctx context.Context, repo repoListItem) bool {
	endpoint, err := s.api.repoEndpoint(
		s.org,
		repo.Name,
		"branches",
		repo.DefaultBranch,
		"protection",
	)
	if err != nil {
		return false
	}

	var protection branchProtection

	if _, err := s.api.getJSON(ctx, endpoint, &protection); err != nil {
		return false
	}

	return protection.RequiredSignatures != nil && protection.RequiredSignatures.Enabled
}

func (s *discoveryScanner) probeWorkflows(ctx context.Context, repo repoListItem) bool {
	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, "actions", "workflows")
	if err != nil {
		return false
	}

	var page workflowsPage

	if _, err := s.api.getJSON(ctx, endpoint, &page); err != nil {
		return false
	}

	return page.TotalCount > 0
}

func (s *discoveryScanner) probeRepoFile(ctx context.Context, repo repoListItem, parts ...string) bool {
	segments := append([]string{"contents"}, parts...)

	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, segments...)
	if err != nil {
		return false
	}

	var payload map[string]any

	if _, err := s.api.getJSON(ctx, endpoint, &payload); err != nil {
		return false
	}

	return true
}

func (s *discoveryScanner) countDependabotCritical(ctx context.Context, repo repoListItem) int {
	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, "dependabot", "alerts")
	if err != nil {
		return 0
	}

	endpoint, err = withPerPage(endpoint, 100)
	if err != nil {
		return 0
	}

	endpoint, err = appendQuery(endpoint, "severity", "critical")
	if err != nil {
		return 0
	}

	endpoint, err = appendQuery(endpoint, "state", "open")
	if err != nil {
		return 0
	}

	var alerts []map[string]any

	if _, err := s.api.getPaginated(ctx, endpoint, &alerts); err != nil {
		return 0
	}

	return len(alerts)
}

func (s *discoveryScanner) countSecretScanningOpen(ctx context.Context, repo repoListItem) int {
	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, "secret-scanning", "alerts")
	if err != nil {
		return 0
	}

	endpoint, err = withPerPage(endpoint, 100)
	if err != nil {
		return 0
	}

	endpoint, err = appendQuery(endpoint, "state", "open")
	if err != nil {
		return 0
	}

	var alerts []map[string]any

	if _, err := s.api.getPaginated(ctx, endpoint, &alerts); err != nil {
		return 0
	}

	return len(alerts)
}

func (s *discoveryScanner) countCodeScanningCritical(ctx context.Context, repo repoListItem) int {
	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, "code-scanning", "alerts")
	if err != nil {
		return 0
	}

	endpoint, err = withPerPage(endpoint, 100)
	if err != nil {
		return 0
	}

	endpoint, err = appendQuery(endpoint, "severity", "critical")
	if err != nil {
		return 0
	}

	endpoint, err = appendQuery(endpoint, "state", "open")
	if err != nil {
		return 0
	}

	var alerts []map[string]any

	if _, err := s.api.getPaginated(ctx, endpoint, &alerts); err != nil {
		return 0
	}

	return len(alerts)
}

func appendQuery(endpoint, key, value string) (string, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("cannot parse endpoint: %w", err)
	}

	q := parsed.Query()
	q.Set(key, value)
	parsed.RawQuery = q.Encode()

	return parsed.String(), nil
}

func repoAggregateFacts(agg *repoScanAggregate) []Fact {
	if agg.ScannedRepos == 0 {
		return nil
	}

	return []Fact{
		coverageFact(
			"f-branch-protection-coverage",
			"repo_branch_protection_coverage",
			agg.WithBranchProtection,
			agg.ScannedRepos,
			"GET /repos/{owner}/{repo}/branches/{branch}/protection",
		),
		coverageFact(
			"f-pr-review-coverage",
			"repo_pr_reviews_required_coverage",
			agg.WithRequiredReviews,
			agg.ScannedRepos,
			"GET /repos/{owner}/{repo}/branches/{branch}/protection",
		),
		coverageFact(
			"f-signed-commits-coverage",
			"repo_signed_commits_required_coverage",
			agg.WithSignedCommits,
			agg.ScannedRepos,
			"GET /repos/{owner}/{repo}/branches/{branch}/protection",
		),
		coverageFact(
			"f-workflow-coverage",
			"repo_workflow_coverage",
			agg.WithWorkflows,
			agg.ScannedRepos,
			"GET /repos/{owner}/{repo}/actions/workflows",
		),
		coverageFact(
			"f-security-md-coverage",
			"repo_security_md_coverage",
			agg.WithSecurityMD,
			agg.ScannedRepos,
			"GET /repos/{owner}/{repo}/contents/SECURITY.md",
		),
		coverageFact(
			"f-contributing-md-coverage",
			"repo_contributing_md_coverage",
			agg.WithContributingMD,
			agg.ScannedRepos,
			"GET /repos/{owner}/{repo}/contents/CONTRIBUTING.md",
		),
		coverageFact(
			"f-dependabot-config-coverage",
			"repo_dependabot_config_coverage",
			agg.WithDependabotConfig,
			agg.ScannedRepos,
			"GET /repos/{owner}/{repo}/contents/.github/dependabot.yml",
		),
		{
			FactID:  "f-dependabot-critical-open",
			FactKey: "repo_dependabot_critical_open",
			Scope:   "org",
			Value: map[string]int{
				"open_critical": agg.DependabotCriticalOpen,
				"repos_scanned": agg.ScannedRepos,
			},
			APIRef: "GET /repos/{owner}/{repo}/dependabot/alerts?severity=critical&state=open",
		},
		{
			FactID:  "f-secret-scanning-open",
			FactKey: "repo_secret_scanning_alerts_open",
			Scope:   "org",
			Value: map[string]int{
				"open":          agg.SecretScanningOpen,
				"repos_scanned": agg.ScannedRepos,
			},
			APIRef: "GET /repos/{owner}/{repo}/secret-scanning/alerts?state=open",
		},
		{
			FactID:  "f-code-scanning-critical-open",
			FactKey: "repo_code_scanning_critical_open",
			Scope:   "org",
			Value: map[string]int{
				"open_critical": agg.CodeScanningCriticalOpen,
				"repos_scanned": agg.ScannedRepos,
			},
			APIRef: "GET /repos/{owner}/{repo}/code-scanning/alerts?severity=critical&state=open",
		},
	}
}

func coverageFact(id, key string, matched, total int, apiRef string) Fact {
	return Fact{
		FactID:  id,
		FactKey: key,
		Scope:   "org",
		Value: map[string]int{
			"matched": matched,
			"total":   total,
		},
		APIRef: apiRef,
	}
}

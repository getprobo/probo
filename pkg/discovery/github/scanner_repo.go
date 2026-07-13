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
	"strings"

	"go.probo.inc/probo/pkg/discovery/vfs"
)

var (
	lockfilePaths = [][]string{
		{"package-lock.json"},
		{"yarn.lock"},
		{"pnpm-lock.yaml"},
		{"go.sum"},
		{"Gemfile.lock"},
		{"poetry.lock"},
		{"Cargo.lock"},
	}

	docPaths = [][]string{
		{"DEVELOPMENT.md"},
		{"docs", "development.md"},
		{"docs", "code-review.md"},
		{"docs", "security", "README.md"},
		{"docs", "incident-response.md"},
		{"SECURITY_GUIDELINES.md"},
	}

	envPaths = [][]string{
		{".env"},
		{".env.production"},
		{".env.local"},
	}
)

type (
	branchProtection struct {
		RequiredPullRequestReviews *struct {
			RequiredApprovingReviewCount int  `json:"required_approving_review_count"`
			DismissStaleReviews          bool `json:"dismiss_stale_reviews"`
		} `json:"required_pull_request_reviews"`
		RequiredSignatures *struct {
			Enabled bool `json:"enabled"`
		} `json:"required_signatures"`
		AllowForcePushes *struct {
			Enabled bool `json:"enabled"`
		} `json:"allow_force_pushes"`
		RequiredStatusChecks *struct {
			Strict   bool     `json:"strict"`
			Contexts []string `json:"contexts"`
		} `json:"required_status_checks"`
		EnforceAdmins *struct {
			Enabled bool `json:"enabled"`
		} `json:"enforce_admins"`
		Restrictions *struct {
			Users []any `json:"users"`
			Teams []any `json:"teams"`
			Apps  []any `json:"apps"`
		} `json:"restrictions"`
	}

	workflowsListResponse struct {
		TotalCount int `json:"total_count"`
		Workflows  []struct {
			Path  string `json:"path"`
			State string `json:"state"`
		} `json:"workflows"`
	}

	contentResponse struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}

	pushProtectionResponse struct {
		Status string `json:"status"`
	}

	codeScanningDefaultSetup struct {
		State string `json:"state"`
	}

	deployKey struct {
		ReadOnly bool `json:"read_only"`
	}

	commitItem struct {
		Commit struct {
			Verification struct {
				Verified bool `json:"verified"`
			} `json:"verification"`
		} `json:"commit"`
	}

	repoScanAggregate struct {
		TotalRepos                int
		PublicRepos               int
		ScannedRepos              int
		ProductionLikely          int
		WithBranchProtection      int
		WithRequiredReviews       int
		WithSignedCommitsRequired int
		WithSignedCommitsPractice int
		ForcePushDisabled         int
		WithRequiredStatusChecks  int
		WithBypassRestrictions    int
		WithWorkflows             int
		WithPRWorkflow            int
		WithPullRequestTargetRisk int
		WithCodeQL                int
		WithCodeQLDefaultSetup    int
		WithDependencyReview      int
		WithSASTInCI              int
		WithDepScanInCI           int
		WithWorkflowSecrets       int
		WithSecurityMD            int
		WithContributingMD        int
		WithDevGuide              int
		WithCodeReviewGuide       int
		WithDependabotConfig      int
		WithRenovateConfig        int
		WithLockfile              int
		WithEnvOnDefaultBranch    int
		WithPushProtection        int
		DeployKeysWrite           int
		DependabotCriticalOpen    int
		SecretScanningOpen        int
		CodeScanningCriticalOpen  int
		WithCommitStatusCI        int
		WithExternalCI            int
		WithSecurityContact       int
		WithIncidentResponseDoc   int
		WithIssueTemplates        int
		WithDeFactoPRReview       int
		PRSampled                 int
		PRReviewed                int
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
		Name:  MeasureOrgPublicRepos,
		Scope: "org",
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

	ciAgg := &ciProviderAggregate{Providers: map[string]int{}}

	eligible := filterEligibleRepos(repos)
	probeSignals := s.collectRepoProbeSignals(ctx, eligible)
	classifications, classLimitations := s.repoClassifier.Classify(ctx, eligible, probeSignals)
	s.repoClassifications = classifications

	sheet.Limitations = append(sheet.Limitations, classLimitations...)

	if len(eligible) > 0 {
		workspace, workspaceLimitations := s.buildWorkspace(ctx, eligible, classifications)
		s.fs = workspace

		sheet.Limitations = append(sheet.Limitations, workspaceLimitations...)
	}

	fileIndex, indexLimitations := s.buildFileIndex(ctx)
	sheet.Limitations = append(sheet.Limitations, indexLimitations...)

	for _, repo := range eligible {
		s.scanRepo(ctx, repo, fileIndex, agg, ciAgg)
	}

	sheet.Facts = append(sheet.Facts, repoAggregateFacts(agg)...)
	sheet.Facts = append(sheet.Facts, ciProviderFact(ciAgg, agg.ScannedRepos))
}

func (s *discoveryScanner) listOrgRepos(ctx context.Context) ([]repoListItem, error) {
	return s.api.listOrgRepositories(ctx, s.org)
}

func filterEligibleRepos(repos []repoListItem) []repoListItem {
	eligible := make([]repoListItem, 0, len(repos))

	for _, repo := range repos {
		if repo.Archived || repo.Disabled || repo.DefaultBranch == "" {
			continue
		}

		eligible = append(eligible, repo)
	}

	return eligible
}

func (s *discoveryScanner) buildFileIndex(ctx context.Context) (*vfs.FileIndex, []string) {
	index, err := vfs.BuildDiscoveryIndex(ctx, s.fs)
	if err == nil {
		return index, nil
	}

	return index, []string{
		"worktree file indexing partially unavailable; falling back to per-repository file reads",
	}
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

func (s *discoveryScanner) scanRepo(
	ctx context.Context,
	repo repoListItem,
	fileIndex *vfs.FileIndex,
	agg *repoScanAggregate,
	ciAgg *ciProviderAggregate,
) {
	agg.ScannedRepos++

	protection, protected := s.fetchBranchProtection(ctx, repo)
	if protected {
		agg.WithBranchProtection++
		s.applyBranchProtectionSignals(protection, agg)
	}

	hasWorkflows := fileIndex.HasRepoPrefix(repo.Name, ".github/workflows") || s.probeWorkflows(ctx, repo)
	if hasWorkflows {
		agg.WithWorkflows++
	}

	signals := s.analyzeRepoWorkflows(ctx, repo, fileIndex)
	mergeWorkflowSignalsIntoAggregate(&signals, agg)

	if class, ok := s.repoClassifications[repo.Name]; ok {
		if class.ProductionLikely {
			agg.ProductionLikely++
		}
	} else if isLikelyProductionRepo(repo, protected, hasWorkflows) {
		agg.ProductionLikely++
	}

	s.scanRepoCIStatuses(ctx, repo, agg, ciAgg)
	s.scanRepoPullRequestPractice(ctx, repo, agg)

	if content, ok := s.readRepoFile(ctx, repo.Name, "SECURITY.md"); ok {
		agg.WithSecurityMD++

		if securityContactInMarkdown(string(content)) {
			agg.WithSecurityContact++
		}
	} else if s.repoHasFile(ctx, fileIndex, repo.Name, "SECURITY.md") {
		agg.WithSecurityMD++
	}

	if s.repoHasFile(ctx, fileIndex, repo.Name, "CONTRIBUTING.md") {
		agg.WithContributingMD++
	}

	for _, path := range docPaths {
		fullPath := strings.Join(path, "/")

		if len(path) > 1 && path[1] == "incident-response.md" {
			if content, ok := s.readRepoFile(ctx, repo.Name, fullPath); ok {
				if incidentResponseInMarkdown(string(content)) {
					agg.WithIncidentResponseDoc++
				}
			} else if s.repoHasFile(ctx, fileIndex, repo.Name, fullPath) {
				agg.WithIncidentResponseDoc++
			}

			continue
		}

		if s.repoHasFile(ctx, fileIndex, repo.Name, fullPath) {
			switch {
			case len(path) > 1 && path[1] == "code-review.md":
				agg.WithCodeReviewGuide++
			default:
				agg.WithDevGuide++
			}
		}
	}

	if fileIndex.HasRepoPrefix(repo.Name, ".github/ISSUE_TEMPLATE") ||
		s.repoHasFile(ctx, fileIndex, repo.Name, ".github/ISSUE_TEMPLATE.md") {
		agg.WithIssueTemplates++
	}

	if s.repoHasFile(ctx, fileIndex, repo.Name, ".github/dependabot.yml") {
		agg.WithDependabotConfig++
	}

	if s.repoHasFile(ctx, fileIndex, repo.Name, "renovate.json") ||
		s.repoHasFile(ctx, fileIndex, repo.Name, ".github/renovate.json") {
		agg.WithRenovateConfig++
	}

	for _, path := range lockfilePaths {
		if s.repoHasFile(ctx, fileIndex, repo.Name, strings.Join(path, "/")) {
			agg.WithLockfile++

			break
		}
	}

	for _, path := range envPaths {
		if s.repoHasFile(ctx, fileIndex, repo.Name, strings.Join(path, "/")) {
			agg.WithEnvOnDefaultBranch++

			break
		}
	}

	if s.probePushProtection(ctx, repo) {
		agg.WithPushProtection++
	}

	if s.probeCodeQLDefaultSetup(ctx, repo) {
		agg.WithCodeQLDefaultSetup++
	}

	agg.DeployKeysWrite += s.countWriteDeployKeys(ctx, repo)

	if verified, total := s.sampleCommitSignatures(ctx, repo); total > 0 && verified > 0 {
		agg.WithSignedCommitsPractice++
	}

	if s.scopes.hasSecurityEvents() {
		agg.DependabotCriticalOpen += s.countDependabotCritical(ctx, repo)
		agg.SecretScanningOpen += s.countSecretScanningOpen(ctx, repo)
		agg.CodeScanningCriticalOpen += s.countCodeScanningCritical(ctx, repo)
	}
}

func (s *discoveryScanner) repoHasFile(
	ctx context.Context,
	fileIndex *vfs.FileIndex,
	repoName string,
	path string,
) bool {
	if fileIndex.HasRepoFile(repoName, path) {
		return true
	}

	return vfs.HasPath(ctx, s.fs, vfs.RepoPath(repoName, path))
}

func (s *discoveryScanner) readRepoFile(
	ctx context.Context,
	repoName string,
	path string,
) ([]byte, bool) {
	content, err := s.fs.Read(ctx, vfs.RepoPath(repoName, path))
	if err != nil {
		return nil, false
	}

	return content, true
}

func (s *discoveryScanner) fetchBranchProtection(
	ctx context.Context,
	repo repoListItem,
) (*branchProtection, bool) {
	return s.api.getBranchProtection(ctx, s.org, repo.Name, repo.DefaultBranch)
}

func (s *discoveryScanner) applyBranchProtectionSignals(
	protection *branchProtection,
	agg *repoScanAggregate,
) {
	if protection.RequiredPullRequestReviews != nil &&
		protection.RequiredPullRequestReviews.RequiredApprovingReviewCount > 0 {
		agg.WithRequiredReviews++
	}

	if protection.RequiredSignatures != nil && protection.RequiredSignatures.Enabled {
		agg.WithSignedCommitsRequired++
	}

	if protection.AllowForcePushes == nil || !protection.AllowForcePushes.Enabled {
		agg.ForcePushDisabled++
	}

	if protection.RequiredStatusChecks != nil && len(protection.RequiredStatusChecks.Contexts) > 0 {
		agg.WithRequiredStatusChecks++
	}

	if protection.Restrictions != nil &&
		(len(protection.Restrictions.Users) > 0 ||
			len(protection.Restrictions.Teams) > 0 ||
			len(protection.Restrictions.Apps) > 0) {
		agg.WithBypassRestrictions++
	}
}

func mergeWorkflowSignalsIntoAggregate(signals *workflowSignals, agg *repoScanAggregate) {
	if signals.RunsOnPullRequest {
		agg.WithPRWorkflow++
	}

	if signals.UsesPullRequestTarget {
		agg.WithPullRequestTargetRisk++
	}

	if signals.UsesCodeQL {
		agg.WithCodeQL++
	}

	if signals.UsesDependencyReview {
		agg.WithDependencyReview++
	}

	if signals.UsesThirdPartySAST {
		agg.WithSASTInCI++
	}

	if signals.UsesDepScanInCI {
		agg.WithDepScanInCI++
	}

	if signals.UsesWorkflowSecrets {
		agg.WithWorkflowSecrets++
	}
}

func (s *discoveryScanner) analyzeRepoWorkflows(
	ctx context.Context,
	repo repoListItem,
	fileIndex *vfs.FileIndex,
) workflowSignals {
	combined := workflowSignals{}
	workflowPaths := workflowPathsFromIndex(fileIndex, repo.Name)

	if len(workflowPaths) == 0 {
		return s.analyzeRepoWorkflowsFromAPI(ctx, repo)
	}

	for i, path := range workflowPaths {
		if i >= 5 {
			break
		}

		content, ok := s.readRepoFile(ctx, repo.Name, path)
		if !ok {
			continue
		}

		mergeWorkflowSignals(&combined, analyzeWorkflowYAML(string(content)))
	}

	return combined
}

func workflowPathsFromIndex(fileIndex *vfs.FileIndex, repoName string) []string {
	var paths []string

	for _, path := range fileIndex.RepoFiles(repoName) {
		if strings.HasPrefix(path, ".github/workflows/") {
			paths = append(paths, path)
		}
	}

	return paths
}

func (s *discoveryScanner) analyzeRepoWorkflowsFromAPI(
	ctx context.Context,
	repo repoListItem,
) workflowSignals {
	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, "actions", "workflows")
	if err != nil {
		return workflowSignals{}
	}

	var page workflowsListResponse

	if _, err := s.api.getJSON(ctx, endpoint, &page); err != nil {
		return workflowSignals{}
	}

	combined := workflowSignals{}

	for i, workflow := range page.Workflows {
		if i >= 5 {
			break
		}

		if workflow.Path == "" {
			continue
		}

		content, ok := s.readRepoFile(ctx, repo.Name, workflow.Path)
		if !ok {
			continue
		}

		mergeWorkflowSignals(&combined, analyzeWorkflowYAML(string(content)))
	}

	return combined
}

func (s *discoveryScanner) probeWorkflows(ctx context.Context, repo repoListItem) bool {
	_, hasWorkflows := s.api.listWorkflowCount(ctx, s.org, repo.Name)

	return hasWorkflows
}

func splitContentPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}

	return strings.Split(path, "/")
}

func (s *discoveryScanner) probePushProtection(ctx context.Context, repo repoListItem) bool {
	if !s.scopes.hasSecurityEvents() {
		return false
	}

	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, "secret-scanning", "push-protection")
	if err != nil {
		return false
	}

	var resp pushProtectionResponse

	if _, err := s.api.getJSON(ctx, endpoint, &resp); err != nil {
		return false
	}

	return stringsEqualFold(resp.Status, "enabled")
}

func (s *discoveryScanner) probeCodeQLDefaultSetup(ctx context.Context, repo repoListItem) bool {
	if !s.scopes.hasSecurityEvents() {
		return false
	}

	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, "code-scanning", "default-setup")
	if err != nil {
		return false
	}

	var setup codeScanningDefaultSetup

	if _, err := s.api.getJSON(ctx, endpoint, &setup); err != nil {
		return false
	}

	return stringsEqualFold(setup.State, "configured")
}

func (s *discoveryScanner) countWriteDeployKeys(ctx context.Context, repo repoListItem) int {
	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, "keys")
	if err != nil {
		return 0
	}

	var keys []deployKey

	if _, err := s.api.getJSON(ctx, endpoint, &keys); err != nil {
		return 0
	}

	count := 0

	for _, key := range keys {
		if !key.ReadOnly {
			count++
		}
	}

	return count
}

func (s *discoveryScanner) sampleCommitSignatures(ctx context.Context, repo repoListItem) (int, int) {
	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, "commits")
	if err != nil {
		return 0, 0
	}

	endpoint, err = withPerPage(endpoint, 20)
	if err != nil {
		return 0, 0
	}

	var commits []commitItem

	if _, err := s.api.getJSON(ctx, endpoint, &commits); err != nil {
		return 0, 0
	}

	verified := 0

	for _, commit := range commits {
		if commit.Commit.Verification.Verified {
			verified++
		}
	}

	return verified, len(commits)
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

func stringsEqualFold(a, b string) bool {
	return stringsFold(a) == stringsFold(b)
}

func stringsFold(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func repoAggregateFacts(agg *repoScanAggregate) []Fact {
	if agg.ScannedRepos == 0 {
		return nil
	}

	facts := []Fact{
		coverageFact(MeasureRepoProductionClassification, agg.ProductionLikely, agg.ScannedRepos, "heuristic"),
		coverageFact(MeasureRepoBranchProtectionCoverage, agg.WithBranchProtection, agg.ScannedRepos, "GET /repos/{owner}/{repo}/branches/{branch}/protection"),
		coverageFact(MeasureRepoPRReviewsRequiredCoverage, agg.WithRequiredReviews, agg.ScannedRepos, "GET /repos/{owner}/{repo}/branches/{branch}/protection"),
		coverageFact(MeasureRepoSignedCommitsRequiredCoverage, agg.WithSignedCommitsRequired, agg.ScannedRepos, "GET /repos/{owner}/{repo}/branches/{branch}/protection"),
		coverageFact(MeasureRepoSignedCommitsPracticeCoverage, agg.WithSignedCommitsPractice, agg.ScannedRepos, "GET /repos/{owner}/{repo}/commits"),
		coverageFact(MeasureRepoForcePushDisabledCoverage, agg.ForcePushDisabled, agg.ScannedRepos, "GET /repos/{owner}/{repo}/branches/{branch}/protection"),
		coverageFact(MeasureRepoRequiredStatusChecksCoverage, agg.WithRequiredStatusChecks, agg.ScannedRepos, "GET /repos/{owner}/{repo}/branches/{branch}/protection"),
		coverageFact(MeasureRepoBypassActorRestrictionsCoverage, agg.WithBypassRestrictions, agg.ScannedRepos, "GET /repos/{owner}/{repo}/branches/{branch}/protection"),
		coverageFact(MeasureRepoWorkflowCoverage, agg.WithWorkflows, agg.ScannedRepos, "GET /repos/{owner}/{repo}/actions/workflows"),
		coverageFact(MeasureRepoPRCICoverage, agg.WithPRWorkflow, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents/.github/workflows"),
		coverageFact(MeasureRepoPullRequestTargetRisk, agg.WithPullRequestTargetRisk, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents/.github/workflows"),
		coverageFact(MeasureRepoCodeQLEnabledCoverage, agg.WithCodeQL, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents/.github/workflows"),
		coverageFact(MeasureRepoCodeQLDefaultSetupCoverage, agg.WithCodeQLDefaultSetup, agg.ScannedRepos, "GET /repos/{owner}/{repo}/code-scanning/default-setup"),
		coverageFact(MeasureRepoDependencyReviewCoverage, agg.WithDependencyReview, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents/.github/workflows"),
		coverageFact(MeasureRepoSASTInCICoverage, agg.WithSASTInCI, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents/.github/workflows"),
		coverageFact(MeasureRepoDepScanInCICoverage, agg.WithDepScanInCI, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents/.github/workflows"),
		coverageFact(MeasureRepoSecurityMDCoverage, agg.WithSecurityMD, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents/SECURITY.md"),
		coverageFact(MeasureRepoContributingMDCoverage, agg.WithContributingMD, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents/CONTRIBUTING.md"),
		coverageFact(MeasureRepoDevelopmentGuideCoverage, agg.WithDevGuide, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents"),
		coverageFact(MeasureRepoCodeReviewGuideCoverage, agg.WithCodeReviewGuide, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents/docs/code-review.md"),
		coverageFact(MeasureRepoDependabotConfigCoverage, agg.WithDependabotConfig, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents/.github/dependabot.yml"),
		coverageFact(MeasureRepoRenovateConfigCoverage, agg.WithRenovateConfig, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents/renovate.json"),
		coverageFact(MeasureRepoLockfileCoverage, agg.WithLockfile, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents"),
		coverageFact(MeasureRepoSecretScanningPushProtection, agg.WithPushProtection, agg.ScannedRepos, "GET /repos/{owner}/{repo}/secret-scanning/push-protection"),
		{
			Name:  MeasureRepoEnvOnDefaultBranch,
			Scope: "org",
			Value: map[string]int{
				"repos_with_env": agg.WithEnvOnDefaultBranch,
				"repos_scanned":  agg.ScannedRepos,
			},
			APIRef: "GET /repos/{owner}/{repo}/contents/.env",
		},
		{
			Name:  MeasureRepoDeployKeysWriteAccess,
			Scope: "org",
			Value: map[string]int{
				"write_keys":    agg.DeployKeysWrite,
				"repos_scanned": agg.ScannedRepos,
			},
			APIRef: "GET /repos/{owner}/{repo}/keys",
		},
		{
			Name:  MeasureRepoDependabotCriticalOpen,
			Scope: "org",
			Value: map[string]int{
				"open_critical": agg.DependabotCriticalOpen,
				"repos_scanned": agg.ScannedRepos,
			},
			APIRef: "GET /repos/{owner}/{repo}/dependabot/alerts?severity=critical&state=open",
		},
		{
			Name:  MeasureRepoSecretScanningAlertsOpen,
			Scope: "org",
			Value: map[string]int{
				"open":          agg.SecretScanningOpen,
				"repos_scanned": agg.ScannedRepos,
			},
			APIRef: "GET /repos/{owner}/{repo}/secret-scanning/alerts?state=open",
		},
		{
			Name:  MeasureRepoCodeScanningCriticalOpen,
			Scope: "org",
			Value: map[string]int{
				"open_critical": agg.CodeScanningCriticalOpen,
				"repos_scanned": agg.ScannedRepos,
			},
			APIRef: "GET /repos/{owner}/{repo}/code-scanning/alerts?severity=critical&state=open",
		},
		coverageFact(MeasureRepoCommitStatusCICoverage, agg.WithCommitStatusCI, agg.ScannedRepos, "GET /repos/{owner}/{repo}/commits/{sha}/status"),
		coverageFact(MeasureRepoExternalCICoverage, agg.WithExternalCI, agg.ScannedRepos, "GET /repos/{owner}/{repo}/commits/{sha}/status"),
		coverageFact(MeasureRepoSecurityContactCoverage, agg.WithSecurityContact, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents/SECURITY.md"),
		coverageFact(MeasureRepoIncidentResponseDocCoverage, agg.WithIncidentResponseDoc, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents/docs/incident-response.md"),
		coverageFact(MeasureRepoIssueTemplatesCoverage, agg.WithIssueTemplates, agg.ScannedRepos, "GET /repos/{owner}/{repo}/contents/.github/ISSUE_TEMPLATE"),
		coverageFact(MeasureRepoDeFactoPRReviewCoverage, agg.WithDeFactoPRReview, agg.ScannedRepos, "GET /repos/{owner}/{repo}/pulls"),
		{
			Name:  MeasureRepoPRApprovalRate,
			Scope: "org",
			Value: map[string]int{
				"reviewed":      agg.PRReviewed,
				"sampled":       agg.PRSampled,
				"repos_scanned": agg.ScannedRepos,
			},
			APIRef: "GET /repos/{owner}/{repo}/pulls; GET /repos/{owner}/{repo}/pulls/{number}/reviews",
		},
	}

	return facts
}

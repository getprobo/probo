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
	"time"
)

var productionRepoNameHints = []string{
	"api",
	"app",
	"backend",
	"frontend",
	"web",
	"service",
	"platform",
	"core",
	"prod",
	"production",
	"monorepo",
}

var productionTopicHints = []string{
	"production",
	"microservice",
	"backend",
	"frontend",
	"platform",
	"infrastructure",
	"service",
}

var productionDescriptionHints = []string{
	"production",
	"customer-facing",
	"customer facing",
	"core service",
	"main application",
	"primary api",
}

var productionLanguages = map[string]struct{}{
	"go":         {},
	"typescript": {},
	"javascript": {},
	"python":     {},
	"java":       {},
	"ruby":       {},
	"rust":       {},
	"csharp":     {},
}

// HeuristicRepoClassifier scores repositories from metadata and probe signals.
type HeuristicRepoClassifier struct{}

func (HeuristicRepoClassifier) Classify(
	_ context.Context,
	repos []repoListItem,
	probes map[string]repoProbeSignals,
) (map[string]RepoClassification, []string) {
	classifications := make(map[string]RepoClassification, len(repos))

	for _, repo := range repos {
		probe := probes[repo.Name]
		classifications[repo.Name] = classifyRepoHeuristic(repo, probe)
	}

	return classifications, nil
}

func classifyRepoHeuristic(repo repoListItem, probe repoProbeSignals) RepoClassification {
	if repo.Name == orgProfileRepo {
		return RepoClassification{
			ProductionLikely: true,
			CloneScore:       1000,
			Confidence:       classificationConfidenceHigh,
			Source:           classificationSourceHeuristic,
			Rationale:        "organization profile repository",
		}
	}

	if repo.Fork {
		return RepoClassification{
			ProductionLikely: false,
			CloneScore:       0,
			Confidence:       classificationConfidenceHigh,
			Source:           classificationSourceHeuristic,
			Rationale:        "forked repository",
		}
	}

	score := 0
	reasons := make([]string, 0, 8)
	name := strings.ToLower(repo.Name)

	if matched := matchProductionNameHint(name); matched != "" {
		score += 5

		reasons = append(reasons, "name hint "+matched)
	}

	switch strings.ToLower(repo.DefaultBranch) {
	case "main", "master", "production":
		score += 2

		reasons = append(reasons, "production-like default branch")
	}

	if repo.Private {
		score += 2

		reasons = append(reasons, "private repository")
	} else {
		score++
	}

	if repoPushedRecently(repo.PushedAt, 90*24*time.Hour) {
		score += 2

		reasons = append(reasons, "recent pushes")
	}

	if repo.Size > 0 {
		score++
	}

	if repo.StargazersCount >= 20 {
		score += 2

		reasons = append(reasons, "community interest")
	} else if repo.StargazersCount >= 5 {
		score++
	}

	if repo.ForksCount >= 3 {
		score++

		reasons = append(reasons, "fork activity")
	}

	if repo.OpenIssuesCount > 0 && repo.OpenIssuesCount < 500 {
		score++

		reasons = append(reasons, "active issue tracker")
	}

	if topic := matchProductionTopic(repo.Topics); topic != "" {
		score += 3

		reasons = append(reasons, "topic "+topic)
	}

	if hint := matchProductionDescription(repo.Description); hint != "" {
		score += 2

		reasons = append(reasons, "description mentions "+hint)
	}

	if language := strings.ToLower(strings.TrimSpace(repo.Language)); language != "" {
		if _, ok := productionLanguages[language]; ok {
			score++

			reasons = append(reasons, "application language "+language)
		}
	}

	if probe.BranchProtected {
		score += 5

		reasons = append(reasons, "branch protection enabled")
	}

	if probe.HasWorkflows {
		score += 4

		reasons = append(reasons, "github actions workflows")
	}

	if probe.WorkflowCount > 2 {
		score++

		reasons = append(reasons, "multiple workflows")
	}

	if isLowPriorityRepoName(name) {
		score -= 6

		reasons = append(reasons, "low-priority name pattern")
	}

	productionLikely := score >= 7 ||
		(probe.BranchProtected && probe.HasWorkflows) ||
		(matchProductionNameHint(name) != "" &&
			(probe.BranchProtected || probe.HasWorkflows || repo.StargazersCount >= 10))

	confidence := classificationConfidenceMedium

	switch {
	case score >= 10 || score <= 0:
		confidence = classificationConfidenceHigh
	case score <= 1:
		confidence = classificationConfidenceLow
	}

	rationale := "heuristic score"
	if len(reasons) > 0 {
		rationale = strings.Join(reasons, "; ")
	}

	return RepoClassification{
		ProductionLikely: productionLikely,
		CloneScore:       score,
		Confidence:       confidence,
		Source:           classificationSourceHeuristic,
		Rationale:        rationale,
	}
}

func matchProductionNameHint(name string) string {
	for _, hint := range productionRepoNameHints {
		if strings.Contains(name, hint) {
			return hint
		}
	}

	return ""
}

func matchProductionTopic(topics []string) string {
	for _, topic := range topics {
		topic = strings.ToLower(strings.TrimSpace(topic))
		for _, hint := range productionTopicHints {
			if topic == hint || strings.Contains(topic, hint) {
				return topic
			}
		}
	}

	return ""
}

func matchProductionDescription(description string) string {
	description = strings.ToLower(description)
	for _, hint := range productionDescriptionHints {
		if strings.Contains(description, hint) {
			return hint
		}
	}

	return ""
}

func isLowPriorityRepoName(name string) bool {
	lowPriorityHints := []string{
		"sandbox",
		"playground",
		"experiment",
		"demo",
		"sample",
		"test-",
		"-test",
		"tmp",
		"scratch",
		"deprecated",
		"archive",
	}

	for _, hint := range lowPriorityHints {
		if strings.Contains(name, hint) {
			return true
		}
	}

	return false
}

func repoPushedRecently(pushedAt string, maxAge time.Duration) bool {
	if pushedAt == "" {
		return false
	}

	parsed, err := time.Parse(time.RFC3339, pushedAt)
	if err != nil {
		return false
	}

	return time.Since(parsed) <= maxAge
}

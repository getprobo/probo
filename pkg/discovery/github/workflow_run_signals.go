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

type checkRunObservation struct {
	Name       string
	AppSlug    string
	DetailsURL string
}

func detectToolRunSignals(observations []checkRunObservation) workflowRunSignals {
	signals := workflowRunSignals{}

	for _, item := range observations {
		combined := strings.ToLower(strings.Join([]string{
			item.Name,
			item.AppSlug,
			item.DetailsURL,
		}, " "))

		if matchesAny(
			combined,
			"codeql",
			"github-code-scanning",
			"code scanning",
		) {
			signals.RanCodeQL = true
		}

		if strings.Contains(combined, "dependency-review") {
			signals.RanDependencyReview = true
		}

		if matchesAny(
			combined,
			"snyk",
			"semgrep",
			"sonarqube",
			"sonarcloud",
			"sonar",
		) {
			signals.RanThirdPartySAST = true
		}

		if matchesAny(
			combined,
			"trivy",
			"osv-scanner",
			"dependency-check",
			"anchore",
			"grype",
		) {
			signals.RanDepScanInCI = true
		}
	}

	return signals
}

func detectPRWorkflowRan(observations []checkRunObservation) bool {
	for _, item := range observations {
		combined := strings.ToLower(strings.Join([]string{
			item.Name,
			item.AppSlug,
			item.DetailsURL,
		}, " "))

		if matchesAny(combined, "github-actions", "github actions") {
			return true
		}
	}

	return false
}

func (s *discoveryScanner) collectWorkflowRunSignals(
	ctx context.Context,
	repo repoListItem,
) workflowRunSignals {
	combined := workflowRunSignals{}

	if sha, ok := s.fetchDefaultBranchSHA(ctx, repo); ok {
		observations := s.fetchCheckRunObservations(ctx, repo, sha)
		mergeWorkflowRunSignals(&combined, detectToolRunSignals(observations))
	}

	if sha, ok := s.fetchRecentMergedPRHeadSHA(ctx, repo); ok {
		observations := s.fetchCheckRunObservations(ctx, repo, sha)
		mergeWorkflowRunSignals(&combined, detectToolRunSignals(observations))

		if detectPRWorkflowRan(observations) {
			combined.RanOnPullRequest = true
		}
	}

	return combined
}

func (s *discoveryScanner) fetchCheckRunObservations(
	ctx context.Context,
	repo repoListItem,
	sha string,
) []checkRunObservation {
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

	observations := make([]checkRunObservation, 0, len(runs.CheckRuns))

	for _, run := range runs.CheckRuns {
		observations = append(observations, checkRunObservation{
			Name:       run.Name,
			AppSlug:    run.App.Slug,
			DetailsURL: run.DetailsURL,
		})
	}

	return observations
}

func mergeWorkflowRiskSignalsIntoAggregate(signals *workflowSignals, agg *repoScanAggregate) {
	if signals.ConfiguredPullRequestTarget {
		agg.WithPullRequestTargetRisk++
	}

	if signals.ConfiguredWorkflowSecrets {
		agg.WithWorkflowSecrets++
	}
}

func mergeWorkflowRunSignalsIntoAggregate(signals *workflowRunSignals, agg *repoScanAggregate) {
	if signals.RanOnPullRequest {
		agg.WithPRWorkflow++
	}

	if signals.RanCodeQL {
		agg.WithCodeQL++
	}

	if signals.RanDependencyReview {
		agg.WithDependencyReview++
	}

	if signals.RanThirdPartySAST {
		agg.WithSASTInCI++
	}

	if signals.RanDepScanInCI {
		agg.WithDepScanInCI++
	}
}

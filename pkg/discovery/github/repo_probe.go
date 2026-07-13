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
	"sort"
)

const maxRepoProbeCount = 100

func (s *discoveryScanner) collectRepoProbeSignals(
	ctx context.Context,
	repos []repoListItem,
) map[string]repoProbeSignals {
	targets := reposForProbe(repos)
	signals := make(map[string]repoProbeSignals, len(targets))

	for _, repo := range targets {
		if err := ctx.Err(); err != nil {
			break
		}

		protected := false
		if _, ok := s.fetchBranchProtection(ctx, repo); ok {
			protected = true
		}

		workflowCount, hasWorkflows := s.probeWorkflowSignals(ctx, repo)

		signals[repo.Name] = repoProbeSignals{
			BranchProtected: protected,
			HasWorkflows:    hasWorkflows,
			WorkflowCount:   workflowCount,
		}
	}

	return signals
}

func reposForProbe(repos []repoListItem) []repoListItem {
	if len(repos) <= maxRepoProbeCount {
		return repos
	}

	ranked := make([]repoListItem, len(repos))
	copy(ranked, repos)

	sort.Slice(ranked, func(i, j int) bool {
		left := classifyRepoHeuristic(ranked[i], repoProbeSignals{}).CloneScore
		right := classifyRepoHeuristic(ranked[j], repoProbeSignals{}).CloneScore
		if left != right {
			return left > right
		}

		return ranked[i].Name < ranked[j].Name
	})

	return ranked[:maxRepoProbeCount]
}

func (s *discoveryScanner) probeWorkflowSignals(
	ctx context.Context,
	repo repoListItem,
) (count int, hasWorkflows bool) {
	endpoint, err := s.api.repoEndpoint(s.org, repo.Name, "actions", "workflows")
	if err != nil {
		return 0, false
	}

	var page workflowsListResponse

	if _, err := s.api.getJSON(ctx, endpoint, &page); err != nil {
		return 0, false
	}

	return page.TotalCount, page.TotalCount > 0
}

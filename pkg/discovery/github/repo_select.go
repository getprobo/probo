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
	"fmt"
	"sort"
)

const (
	maxReposToClone   = 50
	minRepoCloneScore = 2
	orgProfileRepo    = ".github"
	shallowCloneDepth = 1
)

type repoCloneCandidate struct {
	repo  repoListItem
	score int
}

// selectReposForClone ranks repositories and returns the subset worth cloning.
func selectReposForClone(
	repos []repoListItem,
	classifications map[string]RepoClassification,
) ([]repoListItem, string) {
	candidates := make([]repoCloneCandidate, 0, len(repos))

	for _, repo := range repos {
		score := cloneScoreForRepo(repo, classifications)
		if score < minRepoCloneScore && repo.Name != orgProfileRepo {
			continue
		}

		candidates = append(candidates, repoCloneCandidate{repo: repo, score: score})
	}

	if len(candidates) == 0 {
		return nil, "no repositories met git clone relevance heuristics; using API file reads"
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}

		return candidates[i].repo.Name < candidates[j].repo.Name
	})

	if len(candidates) > maxReposToClone {
		candidates = candidates[:maxReposToClone]
	}

	selected := make([]repoListItem, 0, len(candidates))
	for _, candidate := range candidates {
		selected = append(selected, candidate.repo)
	}

	limitation := fmt.Sprintf(
		"selected %d of %d repositories for shallow git clone (depth %d, single branch, no tags/submodules)",
		len(selected),
		len(repos),
		shallowCloneDepth,
	)

	return selected, limitation
}

func cloneScoreForRepo(
	repo repoListItem,
	classifications map[string]RepoClassification,
) int {
	if class, ok := classifications[repo.Name]; ok {
		return class.CloneScore
	}

	return classifyRepoHeuristic(repo, repoProbeSignals{}).CloneScore
}

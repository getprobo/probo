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
	"strings"
	"time"
)

const (
	maxReposToClone    = 50
	minRepoCloneScore  = 2
	orgProfileRepo     = ".github"
	maxRepoCloneSizeKB = 100_000 // 100 MiB per GitHub repository size metric
	shallowCloneDepth  = 1
)

type repoCloneCandidate struct {
	repo  repoListItem
	score int
}

// selectReposForClone ranks repositories and returns the subset worth cloning.
func selectReposForClone(repos []repoListItem) ([]repoListItem, string) {
	candidates := make([]repoCloneCandidate, 0, len(repos))
	skippedOversized := 0

	for _, repo := range repos {
		if repoTooLargeForClone(repo) {
			skippedOversized++

			continue
		}

		score := scoreRepoForClone(repo)
		if score < minRepoCloneScore && repo.Name != orgProfileRepo {
			continue
		}

		candidates = append(candidates, repoCloneCandidate{repo: repo, score: score})
	}

	if len(candidates) == 0 {
		if skippedOversized > 0 {
			return nil, fmt.Sprintf(
				"no repositories selected for shallow git clone; skipped %d oversized repositories (>%d MiB); using API file reads",
				skippedOversized,
				maxRepoCloneSizeKB/1024,
			)
		}

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
	if skippedOversized > 0 {
		limitation += fmt.Sprintf(
			"; skipped %d oversized repositories (>%d MiB)",
			skippedOversized,
			maxRepoCloneSizeKB/1024,
		)
	}

	return selected, limitation
}

func repoTooLargeForClone(repo repoListItem) bool {
	if repo.Size <= 0 {
		return false
	}

	return repo.Size > maxRepoCloneSizeKB
}

func scoreRepoForClone(repo repoListItem) int {
	if repo.Name == orgProfileRepo {
		return 1000
	}

	if repo.Fork {
		return 0
	}

	score := 0
	name := strings.ToLower(repo.Name)

	for _, hint := range productionRepoNameHints {
		if strings.Contains(name, hint) {
			score += 5

			break
		}
	}

	switch strings.ToLower(repo.DefaultBranch) {
	case "main", "master", "production":
		score += 2
	}

	if !repo.Private {
		score += 3
	}

	if repoPushedRecently(repo.PushedAt, 90*24*time.Hour) {
		score += 2
	}

	if isLowPriorityRepoName(name) {
		score -= 6
	}

	return score
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

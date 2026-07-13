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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testRepoClassifications(t *testing.T, repos []repoListItem) map[string]RepoClassification {
	t.Helper()

	classifications, _ := HeuristicRepoClassifier{}.Classify(context.Background(), repos, nil)

	return classifications
}

func TestSelectReposForClone_PrioritizesRelevantRepos(t *testing.T) {
	t.Parallel()

	repos := make([]repoListItem, 0, 60)
	for i := range 60 {
		repos = append(repos, repoListItem{
			Name:          fmt.Sprintf("repo-%02d", i),
			DefaultBranch: "main",
			Private:       true,
		})
	}

	repos = append(repos,
		repoListItem{Name: "api", DefaultBranch: "main", Private: true, PushedAt: time.Now().UTC().Format(time.RFC3339)},
		repoListItem{Name: "sandbox-playground", DefaultBranch: "main", Private: true},
		repoListItem{Name: ".github", DefaultBranch: "main", Private: true},
	)

	selected, limitation := selectReposForClone(repos, testRepoClassifications(t, repos))
	require.NotEmpty(t, limitation)
	require.LessOrEqual(t, len(selected), maxReposToClone)

	names := repoNames(selected)
	assert.Contains(t, names, "api")
	assert.Contains(t, names, ".github")
	assert.NotContains(t, names, "sandbox-playground")
}

func TestCloneScoreForRepo_UsesMetadataSignals(t *testing.T) {
	t.Parallel()

	classifications := testRepoClassifications(t, []repoListItem{
		{
			Name:            "payments-api",
			DefaultBranch:   "main",
			Private:         false,
			PushedAt:        time.Now().UTC().Format(time.RFC3339),
			Description:     "Primary customer-facing API",
			Topics:          []string{"microservice"},
			StargazersCount: 42,
		},
		{Name: "sandbox-playground"},
		{Name: "forked", Fork: true},
	})

	assert.GreaterOrEqual(t, classifications["payments-api"].CloneScore, minRepoCloneScore)
	assert.Less(t, classifications["sandbox-playground"].CloneScore, minRepoCloneScore)
	assert.Equal(t, 0, classifications["forked"].CloneScore)
}

func TestSelectReposForClone_IncludesLargeRepos(t *testing.T) {
	t.Parallel()

	repos := []repoListItem{
		{Name: "api", DefaultBranch: "main", Private: true, Size: 500_000},
		{Name: "monorepo", DefaultBranch: "main", Private: true, Size: 2_000_000},
		{Name: ".github", DefaultBranch: "main", Private: true},
	}

	selected, limitation := selectReposForClone(repos, testRepoClassifications(t, repos))
	names := repoNames(selected)

	assert.Contains(t, names, "api")
	assert.Contains(t, names, "monorepo")
	assert.Contains(t, names, ".github")
	assert.Contains(t, limitation, "shallow git clone")
}

func repoNames(repos []repoListItem) []string {
	names := make([]string, 0, len(repos))
	for _, repo := range repos {
		names = append(names, repo.Name)
	}

	return names
}

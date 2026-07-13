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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	selected, limitation := selectReposForClone(repos)
	require.NotEmpty(t, limitation)
	require.LessOrEqual(t, len(selected), maxReposToClone)

	names := repoNames(selected)
	assert.Contains(t, names, "api")
	assert.Contains(t, names, ".github")
	assert.NotContains(t, names, "sandbox-playground")
}

func TestScoreRepoForClone(t *testing.T) {
	t.Parallel()

	assert.GreaterOrEqual(t, scoreRepoForClone(repoListItem{
		Name:          "payments-api",
		DefaultBranch: "main",
		Private:       false,
		PushedAt:      time.Now().UTC().Format(time.RFC3339),
	}), minRepoCloneScore)
	assert.Less(t, scoreRepoForClone(repoListItem{Name: "sandbox-playground"}), minRepoCloneScore)
	assert.Equal(t, 0, scoreRepoForClone(repoListItem{Name: "forked", Fork: true}))
}

func TestSelectReposForClone_SkipsOversizedRepos(t *testing.T) {
	t.Parallel()

	repos := []repoListItem{
		{Name: "api", DefaultBranch: "main", Private: true, Size: 1024},
		{Name: "monorepo", DefaultBranch: "main", Private: true, Size: maxRepoCloneSizeKB + 1},
		{Name: ".github", DefaultBranch: "main", Private: true, Size: maxRepoCloneSizeKB + 1},
	}

	selected, limitation := selectReposForClone(repos)
	names := repoNames(selected)

	assert.Contains(t, names, "api")
	assert.NotContains(t, names, "monorepo")
	assert.NotContains(t, names, ".github")
	assert.Contains(t, limitation, "skipped 2 oversized")
}

func TestRepoTooLargeForClone(t *testing.T) {
	t.Parallel()

	assert.False(t, repoTooLargeForClone(repoListItem{Size: 0}))
	assert.False(t, repoTooLargeForClone(repoListItem{Size: maxRepoCloneSizeKB}))
	assert.True(t, repoTooLargeForClone(repoListItem{Size: maxRepoCloneSizeKB + 1}))
}

func repoNames(repos []repoListItem) []string {
	names := make([]string, 0, len(repos))
	for _, repo := range repos {
		names = append(names, repo.Name)
	}

	return names
}

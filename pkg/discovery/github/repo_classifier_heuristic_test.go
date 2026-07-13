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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassifyRepoHeuristic_UsesMetadataAndProbes(t *testing.T) {
	t.Parallel()

	protectedDocs := classifyRepoHeuristic(
		repoListItem{Name: "docs", DefaultBranch: "main"},
		repoProbeSignals{BranchProtected: true, HasWorkflows: true},
	)
	assert.True(t, protectedDocs.ProductionLikely)
	assert.GreaterOrEqual(t, protectedDocs.CloneScore, minRepoCloneScore)

	namedAPI := classifyRepoHeuristic(
		repoListItem{
			Name:            "payments-api",
			DefaultBranch:   "main",
			Description:     "Production payments API",
			Topics:          []string{"microservice"},
			StargazersCount: 25,
		},
		repoProbeSignals{},
	)
	assert.True(t, namedAPI.ProductionLikely)
	assert.GreaterOrEqual(t, namedAPI.CloneScore, minRepoCloneScore)

	sandbox := classifyRepoHeuristic(
		repoListItem{Name: "sandbox-playground"},
		repoProbeSignals{},
	)
	assert.False(t, sandbox.ProductionLikely)
	assert.Less(t, sandbox.CloneScore, minRepoCloneScore)
}

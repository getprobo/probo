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

func TestAnalyzeWorkflowYAML(t *testing.T) {
	t.Parallel()

	signals := analyzeWorkflowYAML(`
name: ci
on:
  pull_request:
jobs:
  scan:
    steps:
      - uses: github/codeql-action/analyze@v3
        with:
          secrets: inherit
`)

	assert.True(t, signals.ConfiguredPullRequest)
	assert.True(t, signals.ConfiguredCodeQL)
	assert.True(t, signals.ConfiguredWorkflowSecrets)
	assert.False(t, signals.ConfiguredPullRequestTarget)
}

func TestAnalyzeWorkflowYAML_ReusableWorkflowUses(t *testing.T) {
	t.Parallel()

	signals := analyzeWorkflowYAML(`
name: security
on: [push, pull_request_target]
jobs:
  sast:
    uses: snyk/actions/node@master
`)

	assert.True(t, signals.ConfiguredPullRequestTarget)
	assert.True(t, signals.ConfiguredThirdPartySAST)
}

func TestAnalyzeWorkflowYAML_ListTrigger(t *testing.T) {
	t.Parallel()

	signals := analyzeWorkflowYAML(`
name: ci
on: [workflow_dispatch, pull_request]
jobs:
  scan:
    steps:
      - uses: aquasecurity/trivy-action@master
`)

	assert.True(t, signals.ConfiguredPullRequest)
	assert.True(t, signals.ConfiguredDepScanInCI)
}

func TestDetectToolRunSignals(t *testing.T) {
	t.Parallel()

	signals := detectToolRunSignals([]checkRunObservation{
		{Name: "CodeQL", AppSlug: "github-code-scanning"},
		{Name: "dependency-review", AppSlug: "github-actions"},
	})

	assert.True(t, signals.RanCodeQL)
	assert.True(t, signals.RanDependencyReview)
	assert.False(t, signals.RanOnPullRequest)
}

func TestDetectPRWorkflowRan(t *testing.T) {
	t.Parallel()

	assert.True(t, detectPRWorkflowRan([]checkRunObservation{
		{Name: "build", AppSlug: "github-actions"},
	}))
	assert.False(t, detectPRWorkflowRan([]checkRunObservation{
		{Name: "ci/circleci", AppSlug: "circleci"},
	}))
}

func TestIsLikelyProductionRepo(t *testing.T) {
	t.Parallel()

	assert.True(t, isLikelyProductionRepo(repoListItem{Name: "docs"}, true, true))
	assert.True(t, isLikelyProductionRepo(repoListItem{
		Name:            "payments-api",
		DefaultBranch:   "main",
		StargazersCount: 10,
	}, false, false))
	assert.False(t, isLikelyProductionRepo(repoListItem{Name: "sandbox-playground"}, false, false))
}

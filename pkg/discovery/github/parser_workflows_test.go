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

	signals := detectToolRunSignals([]ciRunObservation{
		{Label: "CodeQL", Source: "github-code-scanning"},
		{Label: "dependency-review", Source: "github-actions"},
	})

	assert.True(t, signals.RanCodeQL)
	assert.True(t, signals.RanDependencyReview)
	assert.False(t, signals.RanOnPullRequest)
}

func TestDetectToolRunSignals_FromCommitStatus(t *testing.T) {
	t.Parallel()

	signals := detectToolRunSignals([]ciRunObservation{
		{Label: "ci/circleci", URL: "https://circleci.com/gh/acme/api/42"},
		{Label: "security/snyk", URL: "https://snyk.io"},
	})

	assert.True(t, signals.RanThirdPartySAST)
	assert.False(t, signals.RanCodeQL)
}

func TestDetectPRWorkflowRan(t *testing.T) {
	t.Parallel()

	assert.True(t, detectPRWorkflowRan([]ciRunObservation{
		{Label: "build", Source: "github-actions"},
	}))
	assert.True(t, detectPRWorkflowRan([]ciRunObservation{
		{Label: "ci/circleci", URL: "https://circleci.com/gh/acme/api/1"},
	}))
	assert.False(t, detectPRWorkflowRan([]ciRunObservation{
		{Label: "unrelated/context"},
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

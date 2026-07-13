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
	"encoding/base64"
	"strings"
)

type workflowSignals struct {
	RunsOnPullRequest     bool
	UsesPullRequestTarget bool
	UsesCodeQL            bool
	UsesDependencyReview  bool
	UsesThirdPartySAST    bool
	UsesDepScanInCI       bool
	UsesWorkflowSecrets   bool
}

func analyzeWorkflowYAML(content string) workflowSignals {
	lower := strings.ToLower(content)

	signals := workflowSignals{
		RunsOnPullRequest: strings.Contains(lower, "pull_request:") ||
			strings.Contains(lower, "pull_request ]"),
		UsesPullRequestTarget: strings.Contains(lower, "pull_request_target"),
		UsesCodeQL: strings.Contains(lower, "github/codeql-action") ||
			strings.Contains(lower, "codeql-analysis"),
		UsesDependencyReview: strings.Contains(lower, "dependency-review-action"),
		UsesThirdPartySAST: strings.Contains(lower, "snyk/actions") ||
			strings.Contains(lower, "semgrep") ||
			strings.Contains(lower, "sonarqube"),
		UsesDepScanInCI: strings.Contains(lower, "trivy-action") ||
			strings.Contains(lower, "osv-scanner") ||
			strings.Contains(lower, "dependency-check") ||
			strings.Contains(lower, "aquasecurity/trivy"),
		UsesWorkflowSecrets: strings.Contains(lower, "${{ secrets.") ||
			strings.Contains(lower, "secrets:"),
	}

	return signals
}

func mergeWorkflowSignals(dst *workflowSignals, src workflowSignals) {
	dst.RunsOnPullRequest = dst.RunsOnPullRequest || src.RunsOnPullRequest
	dst.UsesPullRequestTarget = dst.UsesPullRequestTarget || src.UsesPullRequestTarget
	dst.UsesCodeQL = dst.UsesCodeQL || src.UsesCodeQL
	dst.UsesDependencyReview = dst.UsesDependencyReview || src.UsesDependencyReview
	dst.UsesThirdPartySAST = dst.UsesThirdPartySAST || src.UsesThirdPartySAST
	dst.UsesDepScanInCI = dst.UsesDepScanInCI || src.UsesDepScanInCI
	dst.UsesWorkflowSecrets = dst.UsesWorkflowSecrets || src.UsesWorkflowSecrets
}

func decodeGitHubContent(encoding, content string) (string, bool) {
	if !strings.EqualFold(encoding, "base64") {
		return "", false
	}

	raw, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", false
	}

	return string(raw), true
}

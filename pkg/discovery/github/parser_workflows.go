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
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

type workflowSignals struct {
	ConfiguredPullRequest       bool
	ConfiguredPullRequestTarget bool
	ConfiguredCodeQL            bool
	ConfiguredDependencyReview  bool
	ConfiguredThirdPartySAST    bool
	ConfiguredDepScanInCI       bool
	ConfiguredWorkflowSecrets   bool
}

type workflowRunSignals struct {
	RanOnPullRequest    bool
	RanCodeQL           bool
	RanDependencyReview bool
	RanThirdPartySAST   bool
	RanDepScanInCI      bool
}

type workflowYAML struct {
	On   any                        `yaml:"on"`
	Jobs map[string]workflowYAMLJob `yaml:"jobs"`
}

type workflowYAMLJob struct {
	Uses  string             `yaml:"uses"`
	Steps []workflowYAMLStep `yaml:"steps"`
}

type workflowYAMLStep struct {
	Uses string `yaml:"uses"`
}

func analyzeWorkflowYAML(content string) workflowSignals {
	var doc workflowYAML

	if err := yaml.Unmarshal([]byte(content), &doc); err != nil {
		return workflowSignals{}
	}

	uses := collectWorkflowUses(doc)
	events := parseWorkflowEvents(doc.On)

	signals := workflowSignals{
		ConfiguredPullRequest: hasWorkflowEvent(events, "pull_request"),
		ConfiguredPullRequestTarget: hasWorkflowEvent(
			events,
			"pull_request_target",
		),
		ConfiguredWorkflowSecrets: strings.Contains(
			strings.ToLower(content),
			"${{ secrets.",
		) || strings.Contains(strings.ToLower(content), "secrets:"),
	}

	for _, action := range uses {
		lower := strings.ToLower(action)

		if matchesAny(lower, "github/codeql-action", "codeql-analysis") {
			signals.ConfiguredCodeQL = true
		}

		if strings.Contains(lower, "dependency-review-action") {
			signals.ConfiguredDependencyReview = true
		}

		if matchesAny(lower, "snyk/actions", "semgrep", "sonarqube", "sonarsource/") {
			signals.ConfiguredThirdPartySAST = true
		}

		if matchesAny(
			lower,
			"trivy-action",
			"osv-scanner",
			"dependency-check",
			"aquasecurity/trivy",
			"anchore/scan-action",
		) {
			signals.ConfiguredDepScanInCI = true
		}
	}

	return signals
}

func mergeWorkflowSignals(dst *workflowSignals, src workflowSignals) {
	dst.ConfiguredPullRequest = dst.ConfiguredPullRequest || src.ConfiguredPullRequest
	dst.ConfiguredPullRequestTarget = dst.ConfiguredPullRequestTarget ||
		src.ConfiguredPullRequestTarget
	dst.ConfiguredCodeQL = dst.ConfiguredCodeQL || src.ConfiguredCodeQL
	dst.ConfiguredDependencyReview = dst.ConfiguredDependencyReview ||
		src.ConfiguredDependencyReview
	dst.ConfiguredThirdPartySAST = dst.ConfiguredThirdPartySAST ||
		src.ConfiguredThirdPartySAST
	dst.ConfiguredDepScanInCI = dst.ConfiguredDepScanInCI || src.ConfiguredDepScanInCI
	dst.ConfiguredWorkflowSecrets = dst.ConfiguredWorkflowSecrets ||
		src.ConfiguredWorkflowSecrets
}

func mergeWorkflowRunSignals(dst *workflowRunSignals, src workflowRunSignals) {
	dst.RanOnPullRequest = dst.RanOnPullRequest || src.RanOnPullRequest
	dst.RanCodeQL = dst.RanCodeQL || src.RanCodeQL
	dst.RanDependencyReview = dst.RanDependencyReview || src.RanDependencyReview
	dst.RanThirdPartySAST = dst.RanThirdPartySAST || src.RanThirdPartySAST
	dst.RanDepScanInCI = dst.RanDepScanInCI || src.RanDepScanInCI
}

func parseWorkflowEvents(on any) []string {
	if on == nil {
		return nil
	}

	switch typed := on.(type) {
	case string:
		return []string{normalizeWorkflowEvent(typed)}
	case []any:
		events := make([]string, 0, len(typed))

		for _, item := range typed {
			events = append(events, parseWorkflowEvents(item)...)
		}

		return events
	case map[string]any:
		events := make([]string, 0, len(typed))

		for key := range typed {
			events = append(events, normalizeWorkflowEvent(key))
		}

		return events
	default:
		return nil
	}
}

func normalizeWorkflowEvent(event string) string {
	event = strings.TrimSpace(strings.ToLower(event))

	if before, _, ok := strings.Cut(event, "["); ok {
		event = strings.TrimSpace(before)
	}

	return event
}

func hasWorkflowEvent(events []string, want string) bool {
	return slices.Contains(events, want)
}

func collectWorkflowUses(doc workflowYAML) []string {
	uses := make([]string, 0, 8)

	for _, job := range doc.Jobs {
		if job.Uses != "" {
			uses = append(uses, job.Uses)
		}

		for _, step := range job.Steps {
			if step.Uses != "" {
				uses = append(uses, step.Uses)
			}
		}
	}

	return uses
}

func matchesAny(value string, patterns ...string) bool {
	for _, pattern := range patterns {
		if strings.Contains(value, pattern) {
			return true
		}
	}

	return false
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

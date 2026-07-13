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

type (
	repoListItem struct {
		Name            string   `json:"name"`
		DefaultBranch   string   `json:"default_branch"`
		Private         bool     `json:"private"`
		Archived        bool     `json:"archived"`
		Disabled        bool     `json:"disabled"`
		Fork            bool     `json:"fork"`
		PushedAt        string   `json:"pushed_at"`
		Size            int      `json:"size"`
		Description     string   `json:"description"`
		Language        string   `json:"language"`
		Topics          []string `json:"topics"`
		StargazersCount int      `json:"stargazers_count"`
		ForksCount      int      `json:"forks_count"`
		OpenIssuesCount int      `json:"open_issues_count"`
	}

	repoProbeSignals struct {
		BranchProtected bool
		HasWorkflows    bool
		WorkflowCount   int
	}

	// RepoClassification captures clone priority and production likelihood.
	RepoClassification struct {
		ProductionLikely bool
		CloneScore       int
		Confidence       string
		Source           string
		Rationale        string
	}
)

const (
	classificationConfidenceHigh   = "high"
	classificationConfidenceMedium = "medium"
	classificationConfidenceLow    = "low"

	classificationSourceHeuristic = "heuristic"
	classificationSourceLLM       = "llm"
)

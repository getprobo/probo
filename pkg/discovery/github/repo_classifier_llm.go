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
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

//go:embed prompts/repo_classification.txt.tmpl
var repoClassificationPromptTemplate string

const (
	maxLLMRepoClassifications = 40
	llmRepoClassificationMin  = 2
	llmRepoClassificationMax  = 9
)

type (
	LLMRepoClassifier struct {
		agent *agent.Agent
	}

	repoClassificationCandidate struct {
		Name            string   `json:"name"`
		Description     string   `json:"description,omitempty"`
		Language        string   `json:"language,omitempty"`
		Topics          []string `json:"topics,omitempty"`
		Private         bool     `json:"private"`
		DefaultBranch   string   `json:"default_branch"`
		StargazersCount int      `json:"stargazers_count"`
		ForksCount      int      `json:"forks_count"`
		OpenIssuesCount int      `json:"open_issues_count"`
		PushedAt        string   `json:"pushed_at,omitempty"`
		BranchProtected bool     `json:"branch_protected"`
		HasWorkflows    bool     `json:"has_workflows"`
		HeuristicScore  int      `json:"heuristic_score"`
		HeuristicReason string   `json:"heuristic_reason,omitempty"`
	}

	repoClassificationLLMResult struct {
		Repositories []repoClassificationLLMItem `json:"repositories"`
	}

	repoClassificationLLMItem struct {
		Name             string  `json:"name"`
		ProductionLikely bool    `json:"production_likely"`
		ClonePriority    string  `json:"clone_priority"`
		Confidence       float64 `json:"confidence"`
		Rationale        string  `json:"rationale"`
	}
)

func NewLLMRepoClassifier(
	client *llm.Client,
	model string,
	temperature float64,
	maxTokens int,
	logger *log.Logger,
) *LLMRepoClassifier {
	outputType, err := agent.NewOutputType[repoClassificationLLMResult]("github_repo_classification")
	if err != nil {
		panic(fmt.Sprintf("cannot build repo classification output type: %v", err))
	}

	return &LLMRepoClassifier{
		agent: agent.New(
			"github-discovery-repo-classification",
			client,
			agent.WithModel(model),
			agent.WithTemperature(temperature),
			agent.WithMaxTokens(maxTokens),
			agent.WithOutputType(outputType),
			agent.WithLogger(logger),
		),
	}
}

func (c *LLMRepoClassifier) Classify(
	ctx context.Context,
	repos []repoListItem,
	probes map[string]repoProbeSignals,
) (map[string]RepoClassification, []string) {
	heuristicClassifier := HeuristicRepoClassifier{}
	classifications, _ := heuristicClassifier.Classify(ctx, repos, probes)

	ambiguous := ambiguousReposForLLM(repos, classifications)
	if len(ambiguous) == 0 {
		return classifications, nil
	}

	prompt, err := renderRepoClassificationPrompt(ambiguous, probes, classifications)
	if err != nil {
		return classifications, []string{
			fmt.Sprintf("llm repository classification unavailable: %v", err),
		}
	}

	result, err := agent.RunTyped[repoClassificationLLMResult](
		ctx,
		c.agent,
		[]llm.Message{{
			Role:  llm.RoleUser,
			Parts: []llm.Part{llm.TextPart{Text: prompt}},
		}},
	)
	if err != nil {
		return classifications, []string{
			fmt.Sprintf("llm repository classification failed: %v", err),
		}
	}

	applyLLMRepoClassifications(classifications, result.Output.Repositories)

	return classifications, []string{
		fmt.Sprintf(
			"llm repository classification refined %d ambiguous repositories",
			len(result.Output.Repositories),
		),
	}
}

func ambiguousReposForLLM(
	repos []repoListItem,
	classifications map[string]RepoClassification,
) []repoListItem {
	ambiguous := make([]repoListItem, 0, len(repos))

	for _, repo := range repos {
		class, ok := classifications[repo.Name]
		if !ok {
			continue
		}

		if class.Confidence == classificationConfidenceHigh {
			continue
		}

		if class.CloneScore < llmRepoClassificationMin || class.CloneScore > llmRepoClassificationMax {
			continue
		}

		ambiguous = append(ambiguous, repo)
	}

	if len(ambiguous) > maxLLMRepoClassifications {
		ambiguous = ambiguous[:maxLLMRepoClassifications]
	}

	return ambiguous
}

func applyLLMRepoClassifications(
	classifications map[string]RepoClassification,
	items []repoClassificationLLMItem,
) {
	for _, item := range items {
		if item.Name == "" || item.Confidence < 0.55 {
			continue
		}

		existing := classifications[item.Name]
		score := existing.CloneScore
		score += clonePriorityScoreDelta(item.ClonePriority)

		confidence := classificationConfidenceMedium
		if item.Confidence >= 0.8 {
			confidence = classificationConfidenceHigh
		} else if item.Confidence < 0.65 {
			confidence = classificationConfidenceLow
		}

		classifications[item.Name] = RepoClassification{
			ProductionLikely: item.ProductionLikely,
			CloneScore:       score,
			Confidence:       confidence,
			Source:           classificationSourceLLM,
			Rationale:        item.Rationale,
		}
	}
}

func clonePriorityScoreDelta(priority string) int {
	switch strings.ToLower(strings.TrimSpace(priority)) {
	case "high":
		return 4
	case "medium":
		return 2
	case "low":
		return 0
	case "skip":
		return -8
	default:
		return 0
	}
}

func renderRepoClassificationPrompt(
	repos []repoListItem,
	probes map[string]repoProbeSignals,
	classifications map[string]RepoClassification,
) (string, error) {
	tmpl, err := template.New("repo_classification").Parse(repoClassificationPromptTemplate)
	if err != nil {
		return "", fmt.Errorf("cannot parse repo classification prompt: %w", err)
	}

	candidates := make([]repoClassificationCandidate, 0, len(repos))
	for _, repo := range repos {
		probe := probes[repo.Name]
		heuristic := classifications[repo.Name]

		candidates = append(candidates, repoClassificationCandidate{
			Name:            repo.Name,
			Description:     repo.Description,
			Language:        repo.Language,
			Topics:          repo.Topics,
			Private:         repo.Private,
			DefaultBranch:   repo.DefaultBranch,
			StargazersCount: repo.StargazersCount,
			ForksCount:      repo.ForksCount,
			OpenIssuesCount: repo.OpenIssuesCount,
			PushedAt:        repo.PushedAt,
			BranchProtected: probe.BranchProtected,
			HasWorkflows:    probe.HasWorkflows,
			HeuristicScore:  heuristic.CloneScore,
			HeuristicReason: heuristic.Rationale,
		})
	}

	candidatesJSON, err := json.Marshal(candidates)
	if err != nil {
		return "", fmt.Errorf("cannot marshal repo classification candidates: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{
		"CandidatesJSON": string(candidatesJSON),
	}); err != nil {
		return "", fmt.Errorf("cannot render repo classification prompt: %w", err)
	}

	return buf.String(), nil
}

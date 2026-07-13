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

//go:embed prompts/glob_code_search.txt.tmpl
var globCodeSearchPromptTemplate string

type (
	LLMGlobQueryResolver struct {
		agent *agent.Agent
		cache map[string]string
	}

	globCodeSearchLLMResult struct {
		Queries []globCodeSearchLLMItem `json:"queries"`
	}

	globCodeSearchLLMItem struct {
		Pattern        string `json:"pattern"`
		SearchFragment string `json:"search_fragment"`
	}
)

func NewLLMGlobQueryResolver(
	client *llm.Client,
	model string,
	temperature float64,
	maxTokens int,
	logger *log.Logger,
) *LLMGlobQueryResolver {
	outputType, err := agent.NewOutputType[globCodeSearchLLMResult]("github_glob_code_search")
	if err != nil {
		panic(fmt.Sprintf("cannot build glob code search output type: %v", err))
	}

	return &LLMGlobQueryResolver{
		agent: agent.New(
			"github-discovery-glob-code-search",
			client,
			agent.WithModel(model),
			agent.WithTemperature(temperature),
			agent.WithMaxTokens(maxTokens),
			agent.WithOutputType(outputType),
			agent.WithLogger(logger),
		),
		cache: map[string]string{},
	}
}

func (r *LLMGlobQueryResolver) Warm(
	ctx context.Context,
	_ string,
	patterns []string,
) []string {
	if len(patterns) == 0 {
		return nil
	}

	prompt, err := renderGlobCodeSearchPrompt(patterns)
	if err != nil {
		return []string{fmt.Sprintf("llm glob code search unavailable: %v", err)}
	}

	result, err := agent.RunTyped[globCodeSearchLLMResult](
		ctx,
		r.agent,
		[]llm.Message{{
			Role:  llm.RoleUser,
			Parts: []llm.Part{llm.TextPart{Text: prompt}},
		}},
	)
	if err != nil {
		return []string{fmt.Sprintf("llm glob code search failed: %v", err)}
	}

	applied := 0

	for _, item := range result.Output.Queries {
		fragment := normalizeSearchFragment(item.SearchFragment)
		if item.Pattern == "" || fragment == "" {
			continue
		}

		r.cache[item.Pattern] = fragment
		applied++
	}

	if applied == 0 {
		return []string{"llm glob code search returned no usable queries; using static map"}
	}

	return []string{
		fmt.Sprintf("llm glob code search resolved %d discovery patterns", applied),
	}
}

func (r *LLMGlobQueryResolver) Query(org, pattern string) (string, bool) {
	if fragment, ok := r.cache[pattern]; ok && fragment != "" {
		return formatCodeSearchQuery(org, fragment), true
	}

	if fragment, ok := staticDiscoveryGlobQueries[pattern]; ok {
		return formatCodeSearchQuery(org, fragment), true
	}

	return "", false
}

func renderGlobCodeSearchPrompt(patterns []string) (string, error) {
	tmpl, err := template.New("glob_code_search").Parse(globCodeSearchPromptTemplate)
	if err != nil {
		return "", fmt.Errorf("cannot parse glob code search prompt: %w", err)
	}

	patternsJSON, err := json.Marshal(patterns)
	if err != nil {
		return "", fmt.Errorf("cannot marshal glob patterns: %w", err)
	}

	staticHintsJSON, err := json.Marshal(staticDiscoveryGlobQueries)
	if err != nil {
		return "", fmt.Errorf("cannot marshal static glob hints: %w", err)
	}

	var buf bytes.Buffer

	if err := tmpl.Execute(&buf, map[string]string{
		"PatternsJSON":    string(patternsJSON),
		"StaticHintsJSON": string(staticHintsJSON),
	}); err != nil {
		return "", fmt.Errorf("cannot render glob code search prompt: %w", err)
	}

	return buf.String(), nil
}

func normalizeSearchFragment(fragment string) string {
	fragment = strings.TrimSpace(fragment)
	fragment = strings.TrimPrefix(fragment, "org:")
	fragment = strings.TrimSpace(fragment)

	if fragment == "" || strings.Contains(fragment, "\n") {
		return ""
	}

	return fragment
}

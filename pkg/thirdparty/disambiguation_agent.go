// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package thirdparty

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/llm"
)

//go:embed prompts/disambiguation.txt.tmpl
var disambiguationPrompt string

const (
	// disambiguationConfidenceThreshold is the floor below which we
	// treat the agent's pick as "no confident match" even when it
	// returned a non-nil matched_id. Mirrors the conservative bias
	// described in the prompt.
	disambiguationConfidenceThreshold = 0.6

	// defaultDisambiguationTimeout caps a single disambiguation run
	// when the config supplies none. The agent has no tools and a
	// single turn, so this is mostly a guard against a hung LLM
	// provider, not a real budget.
	defaultDisambiguationTimeout = 45 * time.Second

	// defaultDisambiguationMaxTokens caps the agent's output when the
	// config carries no max-tokens budget. The final output is tiny (a
	// single id plus a one-sentence rationale), but the budget must
	// leave ample headroom for reasoning models (e.g. the GPT-5
	// family): their reasoning tokens count against max_tokens, so too
	// small a budget gets consumed by reasoning and truncates the JSON,
	// surfacing as "unexpected end of JSON input".
	defaultDisambiguationMaxTokens = 4096
)

// DisambiguationAgentConfig configures the third-party disambiguation
// agent. The agent has no DB tools and no web-search tools: the
// candidate list is supplied entirely in the prompt and the agent
// only picks among it.
//
// MaxTokens and Temperature bound and steer the single LLM call, and
// Timeout caps a single run. Zero-valued fields fall back to package
// defaults.
type DisambiguationAgentConfig struct {
	LLMClient   *llm.Client
	Model       string
	MaxTokens   *int
	Temperature *float64
	Timeout     time.Duration
}

// DisambiguationResult is the structured output the disambiguation
// agent returns when picking the best existing org ThirdParty for a
// catalog entry.
type DisambiguationResult struct {
	MatchedID  *string `json:"matched_id"  jsonschema:"GID of the org third party that best matches, or null if none of the candidates is a confident match."`
	Confidence float64 `json:"confidence"  jsonschema:"Confidence level from 0.0 to 1.0. Below 0.6 means 'no confident match' and matched_id MUST be null."`
	Reasoning  string  `json:"reasoning"   jsonschema:"One short sentence describing the rationale."`
}

// BuildDisambiguationAgent wires the agent that picks the best
// existing org ThirdParty for a catalog entry. It deliberately has
// no tools: the candidate list is supplied in the prompt and the
// agent must only choose among it.
func BuildDisambiguationAgent(
	cfg DisambiguationAgentConfig,
	logger *log.Logger,
) *agent.Agent {
	outputType, err := agent.NewOutputType[DisambiguationResult]("third_party_disambiguation")
	if err != nil {
		panic(fmt.Sprintf("thirdparty: cannot build disambiguation output type: %s", err))
	}

	maxTokens := defaultDisambiguationMaxTokens
	if cfg.MaxTokens != nil && *cfg.MaxTokens > 0 {
		maxTokens = *cfg.MaxTokens
	}

	opts := []agent.Option{
		agent.WithInstructions(disambiguationPrompt),
		agent.WithModel(cfg.Model),
		agent.WithOutputType(outputType),
		agent.WithMaxTurns(1),
		agent.WithMaxTokens(maxTokens),
		agent.WithLogger(logger),
	}

	if cfg.Temperature != nil {
		opts = append(opts, agent.WithTemperature(*cfg.Temperature))
	}

	return agent.New("third-party-disambiguation", cfg.LLMClient, opts...)
}

// Disambiguate runs the agent against the given catalog third party
// and candidate list, and returns the matched candidate's ID — or
// nil when the agent picks "none", returns a confidence below the
// threshold, or fails. Errors from the agent itself are returned;
// "no confident match" is not an error.
//
// The matched candidate is identified by string equality against the
// IDs supplied in `candidates`; we never invent IDs from the agent's
// output, so a model that hallucinates an ID is treated as "none".
func Disambiguate(
	ctx context.Context,
	a *agent.Agent,
	logger *log.Logger,
	commonParty coredata.CommonThirdParty,
	commonDomains coredata.CommonThirdPartyDomains,
	candidates []ScoredCandidate,
	timeout time.Duration,
) (*gid.GID, error) {
	if a == nil || len(candidates) == 0 {
		return nil, nil
	}

	if timeout <= 0 {
		timeout = defaultDisambiguationTimeout
	}

	prompt := buildDisambiguationPrompt(commonParty, commonDomains, candidates)

	agentCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := agent.RunTyped[DisambiguationResult](
		agentCtx,
		a,
		[]llm.Message{
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: prompt}},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot run disambiguation agent: %w", err)
	}

	out := result.Output

	if out.MatchedID == nil || *out.MatchedID == "" {
		return nil, nil
	}

	if out.Confidence < disambiguationConfidenceThreshold {
		logger.InfoCtx(
			ctx,
			"disambiguation agent below confidence threshold",
			log.String("matched_id", *out.MatchedID),
			log.Float64("confidence", out.Confidence),
		)

		return nil, nil
	}

	for _, c := range candidates {
		if c.ThirdParty.ID.String() == *out.MatchedID {
			id := c.ThirdParty.ID

			return &id, nil
		}
	}

	logger.WarnCtx(
		ctx,
		"disambiguation agent returned id not in candidate list",
		log.String("matched_id", *out.MatchedID),
	)

	return nil, nil
}

// buildDisambiguationPrompt formats the catalog third party and the
// heuristic-ranked candidate list into the user message for the
// disambiguation agent. The prompt is intentionally compact: the
// agent only needs ids, names, websites, and the heuristic score to
// decide.
func buildDisambiguationPrompt(
	commonParty coredata.CommonThirdParty,
	commonDomains coredata.CommonThirdPartyDomains,
	candidates []ScoredCandidate,
) string {
	var b strings.Builder

	b.WriteString("Catalog third party:\n")
	fmt.Fprintf(&b, "  name: %s\n", commonParty.Name)

	if commonParty.WebsiteURL != nil && *commonParty.WebsiteURL != "" {
		fmt.Fprintf(&b, "  website: %s\n", *commonParty.WebsiteURL)
	}

	if len(commonDomains) > 0 {
		domains := make([]string, len(commonDomains))
		for i, d := range commonDomains {
			domains[i] = d.Domain
		}

		fmt.Fprintf(&b, "  domains: %s\n", strings.Join(domains, ", "))
	}

	b.WriteString("\nCandidate organisation third parties (heuristic-ranked):\n")

	for i, c := range candidates {
		fmt.Fprintf(&b, "- id: %s\n", c.ThirdParty.ID.String())
		fmt.Fprintf(&b, "  name: %s\n", c.ThirdParty.Name)

		if c.ThirdParty.WebsiteURL != nil && *c.ThirdParty.WebsiteURL != "" {
			fmt.Fprintf(&b, "  website: %s\n", *c.ThirdParty.WebsiteURL)
		}

		fmt.Fprintf(&b, "  heuristic_score: %.2f\n", c.Score)

		if i < len(candidates)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

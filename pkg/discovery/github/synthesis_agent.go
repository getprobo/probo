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
	"text/template"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

//go:embed prompts/synthesis.txt.tmpl
var synthesisPromptTemplate string

type (
	Synthesizer interface {
		Synthesize(
			ctx context.Context,
			sheet *FactSheet,
			existing []ExistingMeasure,
		) (*MeasurePlan, error)
	}

	DeterministicSynthesizer struct{}

	LLMSynthesizer struct {
		agent *agent.Agent
	}
)

var _ Synthesizer = DeterministicSynthesizer{}
var _ Synthesizer = (*LLMSynthesizer)(nil)

func (DeterministicSynthesizer) Synthesize(
	_ context.Context,
	sheet *FactSheet,
	existing []ExistingMeasure,
) (*MeasurePlan, error) {
	return MaterializeFromFacts(sheet, existing)
}

func NewLLMSynthesizer(client *llm.Client, model string, temperature float64, maxTokens int, logger *log.Logger) *LLMSynthesizer {
	return &LLMSynthesizer{
		agent: agent.New(
			"github-discovery-synthesis",
			client,
			agent.WithModel(model),
			agent.WithTemperature(temperature),
			agent.WithMaxTokens(maxTokens),
			agent.WithLogger(logger),
		),
	}
}

func (s *LLMSynthesizer) Synthesize(
	ctx context.Context,
	sheet *FactSheet,
	existing []ExistingMeasure,
) (*MeasurePlan, error) {
	prompt, err := renderSynthesisPrompt(sheet, existing)
	if err != nil {
		return nil, err
	}

	result, err := agent.RunTyped[MeasurePlan](
		ctx,
		s.agent,
		[]llm.Message{
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: prompt}},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot synthesize measure plan: %w", err)
	}

	if err := validateMeasurePlan(&result.Output, existing); err != nil {
		return nil, err
	}

	return &result.Output, nil
}

func renderSynthesisPrompt(sheet *FactSheet, existing []ExistingMeasure) (string, error) {
	tmpl, err := template.New("synthesis").Parse(synthesisPromptTemplate)
	if err != nil {
		return "", fmt.Errorf("cannot parse synthesis prompt: %w", err)
	}

	factsJSON, err := json.Marshal(sheet)
	if err != nil {
		return "", fmt.Errorf("cannot marshal fact sheet: %w", err)
	}

	existingJSON, err := json.Marshal(existing)
	if err != nil {
		return "", fmt.Errorf("cannot marshal existing measures: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{
		"FactsJSON":    string(factsJSON),
		"ExistingJSON": string(existingJSON),
		"MaxCreates":   fmt.Sprintf("%d", maxMeasureCreatesPerRun),
	}); err != nil {
		return "", fmt.Errorf("cannot render synthesis prompt: %w", err)
	}

	return buf.String(), nil
}

func validateMeasurePlan(plan *MeasurePlan, existing []ExistingMeasure) error {
	if plan == nil {
		return fmt.Errorf("synthesis returned empty measure plan")
	}

	if len(plan.Creates) > maxMeasureCreatesPerRun {
		return fmt.Errorf("synthesis returned %d creates, limit is %d", len(plan.Creates), maxMeasureCreatesPerRun)
	}

	allowed := map[string]struct{}{}

	for _, measure := range existing {
		allowed[measure.ID.String()] = struct{}{}
	}

	for _, update := range plan.Updates {
		if _, ok := allowed[update.MeasureID.String()]; !ok {
			return fmt.Errorf("synthesis update references unknown measure %s", update.MeasureID)
		}
	}

	return nil
}

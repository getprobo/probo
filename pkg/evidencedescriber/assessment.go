// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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

package evidencedescriber

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/llm"
)

//go:embed prompt.txt
var systemPrompt string

type (
	Config struct {
		Client    *llm.Client
		Model     string
		Temp      float64
		MaxTokens int
		// Thinking is the extended-thinking budget in tokens. 0 disables
		// extended thinking entirely.
		Thinking int
		Logger   *log.Logger
	}

	Describer struct {
		cfg        Config
		outputType *agent.OutputType
	}
)

// New builds a Describer. The structured output type is decorated once
// (enum on confidence) and cached on the Describer so every call reuses
// the same schema.
func New(cfg Config) (*Describer, error) {
	ot, err := assessmentOutputType()
	if err != nil {
		return nil, fmt.Errorf("cannot build evidence assessment output type: %w", err)
	}

	return &Describer{cfg: cfg, outputType: ot}, nil
}

// Describe runs the evidence describer agent against a single uploaded
// file and returns a structured assessment. The worker persists the
// result as JSONB and derives Evidence.Description from Summary.
func (d *Describer) Describe(
	ctx context.Context,
	filename string,
	mimeType string,
	fileBase64 string,
) (*coredata.EvidenceAssessment, error) {
	opts := []agent.Option{
		agent.WithInstructions(systemPrompt),
		agent.WithModel(d.cfg.Model),
		agent.WithTemperature(d.cfg.Temp),
		agent.WithMaxTokens(d.cfg.MaxTokens),
		agent.WithOutputType(d.outputType),
	}
	if d.cfg.Thinking > 0 {
		opts = append(opts, agent.WithThinking(d.cfg.Thinking))
	}
	if d.cfg.Logger != nil {
		opts = append(opts, agent.WithLogger(d.cfg.Logger))
	}

	ag := agent.New("evidence_describer", d.cfg.Client, opts...)

	result, err := ag.Run(ctx, []llm.Message{
		{
			Role: llm.RoleUser,
			Parts: []llm.Part{
				llm.TextPart{Text: fmt.Sprintf("Filename: %s", filename)},
				llm.FilePart{
					Data:     fileBase64,
					MimeType: mimeType,
					Filename: filename,
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("cannot describe evidence: %w", err)
	}

	var a coredata.EvidenceAssessment
	if err := json.Unmarshal([]byte(result.FinalMessage().Text()), &a); err != nil {
		return nil, fmt.Errorf("cannot parse evidence assessment: %w", err)
	}

	return &a, nil
}

// assessmentOutputType builds the EvidenceAssessment structured output
// type and decorates its JSON Schema with an explicit enum on
// `confidence`. jsonschema-go reads struct tags only as free-form
// descriptions, so the enum cannot be encoded in the tag itself;
// see pkg/vetting/assessment.go for the same pattern applied to
// VendorInfo.
func assessmentOutputType() (*agent.OutputType, error) {
	ot, err := agent.NewOutputType[coredata.EvidenceAssessment]("evidence_assessment")
	if err != nil {
		return nil, fmt.Errorf("cannot create evidence assessment output type: %w", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(ot.Schema, &schema); err != nil {
		return nil, fmt.Errorf("cannot unmarshal evidence assessment schema: %w", err)
	}

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("evidence assessment schema has no properties")
	}

	confidence, ok := properties["confidence"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("evidence assessment schema has no confidence property")
	}
	confidence["enum"] = coredata.EvidenceAssessmentConfidenceEnum

	decorated, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal decorated evidence assessment schema: %w", err)
	}
	ot.Schema = decorated

	return ot, nil
}

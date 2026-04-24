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

package evidenceassessor

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

//go:embed prompt.txt
var systemPrompt string

// assessmentConfidenceEnum is the canonical set of allowed values for
// EvidenceAssessment.Confidence. Injected into the generated JSON
// schema via agent.OutputType.DecorateEnum because jsonschema struct
// tags cannot encode enum constraints directly.
var assessmentConfidenceEnum = []string{"HIGH", "MEDIUM", "LOW"}

type (
	// EvidenceAssessment is the structured output produced by the
	// evidence assessor. The worker persists it as JSONB on the
	// evidences table and derives Evidence.Description from Summary.
	EvidenceAssessment struct {
		Summary         string   `json:"summary"          jsonschema:"One plain-text sentence (two at most) summarising what the evidence shows. When readable is false this field must restate the rejection reason."`
		System          string   `json:"system"           jsonschema:"Tool or platform visible on the file; empty string if not clearly identifiable."`
		Setting         string   `json:"setting"          jsonschema:"Specific configuration or state demonstrated; empty string if not clearly identifiable."`
		Scope           string   `json:"scope"            jsonschema:"Who or what the setting applies to; empty string if not stated."`
		CapturedAt      string   `json:"captured_at"      jsonschema:"ISO-8601 date or datetime visible on the file; empty string when no date is shown."`
		Language        string   `json:"language"         jsonschema:"BCP-47 language tag of the visible text; empty when unclear."`
		Frameworks      []string `json:"frameworks"       jsonschema:"Compliance frameworks explicitly named on the file; empty when unclear."`
		Issues          []string `json:"issues"           jsonschema:"Quality problems on the file itself; empty when the file is clean."`
		Confidence      string   `json:"confidence"       jsonschema:"One of HIGH, MEDIUM, LOW."`
		Readable        bool     `json:"readable"         jsonschema:"True if the file is a usable piece of compliance evidence."`
		RejectionReason string   `json:"rejection_reason" jsonschema:"Populated only when readable is false; empty otherwise."`
	}

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

	Assessor struct {
		cfg        Config
		outputType *agent.OutputType
	}
)

// New builds an Assessor. The structured output type is decorated once
// (enum on confidence) and cached so every call reuses the same schema.
func New(cfg Config) (*Assessor, error) {
	ot, err := assessmentOutputType()
	if err != nil {
		return nil, fmt.Errorf("cannot build evidence assessment output type: %w", err)
	}

	return &Assessor{cfg: cfg, outputType: ot}, nil
}

// Assess runs the evidence assessor agent against a single uploaded
// file and returns a structured assessment. The worker persists the
// result as JSONB and derives Evidence.Description from Summary.
func (a *Assessor) Assess(
	ctx context.Context,
	filename string,
	mimeType string,
	fileBase64 string,
) (*EvidenceAssessment, error) {
	opts := []agent.Option{
		agent.WithInstructions(systemPrompt),
		agent.WithModel(a.cfg.Model),
		agent.WithTemperature(a.cfg.Temp),
		agent.WithMaxTokens(a.cfg.MaxTokens),
		agent.WithOutputType(a.outputType),
	}
	if a.cfg.Thinking > 0 {
		opts = append(opts, agent.WithThinking(a.cfg.Thinking))
	}
	if a.cfg.Logger != nil {
		opts = append(opts, agent.WithLogger(a.cfg.Logger))
	}

	ag := agent.New("evidence_assessor", a.cfg.Client, opts...)

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
		return nil, fmt.Errorf("cannot assess evidence: %w", err)
	}

	var out EvidenceAssessment
	if err := json.Unmarshal([]byte(result.FinalMessage().Text()), &out); err != nil {
		return nil, fmt.Errorf("cannot parse evidence assessment: %w", err)
	}

	return &out, nil
}

// assessmentOutputType builds the EvidenceAssessment structured output
// type and decorates its JSON Schema with an explicit enum on the
// confidence field.
func assessmentOutputType() (*agent.OutputType, error) {
	ot, err := agent.NewOutputType[EvidenceAssessment]("evidence_assessment")
	if err != nil {
		return nil, fmt.Errorf("cannot create evidence assessment output type: %w", err)
	}
	if err := ot.DecorateEnum(map[string][]string{
		"confidence": assessmentConfidenceEnum,
	}); err != nil {
		return nil, fmt.Errorf("cannot decorate evidence assessment schema: %w", err)
	}
	return ot, nil
}

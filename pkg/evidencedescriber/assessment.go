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
	"go.probo.inc/probo/pkg/llm"
)

//go:embed prompt.txt
var systemPrompt string

// EvidenceAssessmentConfidenceEnum is the canonical set of allowed
// values for EvidenceAssessment.Confidence. It is injected into the
// generated JSON schema via agent.OutputType.DecorateEnum because
// jsonschema struct tags cannot encode enum constraints directly.
var evidenceAssessmentConfidenceEnum = []string{"HIGH", "MEDIUM", "LOW"}

type (
	// EvidenceAssessment is the structured output produced by the
	// evidence describer. The worker persists it as JSONB on the
	// evidences table and derives Evidence.Description from Summary.
	EvidenceAssessment struct {
		Summary         string   `json:"summary"          jsonschema:"One plain-text sentence (two at most) summarising what the evidence shows; no markdown, no preamble. When readable is false this field must restate the rejection reason so downstream systems displaying only the summary still inform the user."`
		System          string   `json:"system"           jsonschema:"Tool or platform shown (e.g. 'Google Workspace', 'GitHub', 'AWS IAM'); empty string if not identifiable."`
		Setting         string   `json:"setting"          jsonschema:"Specific configuration, feature, or state demonstrated; empty string if not identifiable."`
		Scope           string   `json:"scope"            jsonschema:"Who or what the evidence applies to (e.g. 'organization-wide', 'all users', 'repository acme/foo'); empty string if not stated."`
		CapturedAt      string   `json:"captured_at"      jsonschema:"ISO-8601 date or datetime visible on the file; empty string if no date is shown."`
		Language        string   `json:"language"         jsonschema:"BCP-47 language tag of the visible text (e.g. 'en', 'fr'); empty string if unclear."`
		Frameworks      []string `json:"frameworks"       jsonschema:"Compliance frameworks the evidence is plausibly relevant to (e.g. 'SOC2', 'ISO27001'); empty when unclear."`
		Issues          []string `json:"issues"           jsonschema:"Quality problems observed: redacted fields, crop, stale date, low resolution, etc."`
		Confidence      string   `json:"confidence"       jsonschema:"Overall confidence in the System / Setting / Scope identification; one of HIGH, MEDIUM, LOW."`
		Readable        bool     `json:"readable"         jsonschema:"True if the file is readable compliance evidence; false if unreadable or off-topic."`
		RejectionReason string   `json:"rejection_reason" jsonschema:"Set only when readable is false; explains why the file is not usable evidence."`
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
) (*EvidenceAssessment, error) {
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

	var a EvidenceAssessment
	if err := json.Unmarshal([]byte(result.FinalMessage().Text()), &a); err != nil {
		return nil, fmt.Errorf("cannot parse evidence assessment: %w", err)
	}

	return &a, nil
}

// assessmentOutputType builds the EvidenceAssessment structured output
// type and decorates its JSON Schema with an explicit enum on the
// `confidence` field.
func assessmentOutputType() (*agent.OutputType, error) {
	ot, err := agent.NewOutputType[EvidenceAssessment]("evidence_assessment")
	if err != nil {
		return nil, fmt.Errorf("cannot create evidence assessment output type: %w", err)
	}
	if err := ot.DecorateEnum(map[string][]string{
		"confidence": evidenceAssessmentConfidenceEnum,
	}); err != nil {
		return nil, fmt.Errorf("cannot decorate evidence assessment schema: %w", err)
	}
	return ot, nil
}

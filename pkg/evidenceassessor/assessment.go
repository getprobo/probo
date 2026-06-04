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
	"time"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

// AssessmentTimeout is the hard upper bound on the assessor's single LLM
// turn. The assessor performs one tool-less structured-output turn, so
// this is far tighter than pkg/vetting's multi-turn budget. It exists so a
// stalled provider connection cannot pin a worker concurrency slot forever
// (the kit worker holds the slot until Process returns). Note it bounds
// only the LLM call, not the surrounding file download or commit, so it
// does not by itself guarantee Process finishes before the worker's
// stale-recovery window (EvidenceAssessmentConfig.StaleAfter, default 300s)
// elapses; correctness under a stale re-claim is enforced by the
// claim-ownership guard on the terminal transitions (see
// coredata.Evidence.SetAssessmentCompleted / SetAssessmentFailed), not by
// this timeout.
const AssessmentTimeout = 4 * time.Minute

//go:embed prompt.txt
var systemPrompt string

// confidenceEnum is the canonical set of allowed values for
// EvidenceAssessment.Confidence. Injected into the generated JSON
// schema via agent.OutputType.DecorateEnum because jsonschema struct
// tags cannot encode enum constraints directly.
var confidenceEnum = []string{"HIGH", "MEDIUM", "LOW"}

type (
	// EvidenceAssessment is the structured output produced by the
	// evidence assessor. The worker persists it as JSONB on the
	// evidences table and derives Evidence.Description from Summary.
	//
	// Only Summary is surfaced to users today (through Description); the
	// remaining structured fields (confidence, readable, issues,
	// frameworks, ...) are stored as deliberate groundwork for a later
	// PR that exposes them on the GraphQL/MCP Evidence type. They are
	// intentionally write-only at the API boundary for now.
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
	outputType, err := agent.NewOutputType[EvidenceAssessment]("evidence_assessment")
	if err != nil {
		return nil, fmt.Errorf("cannot create evidence assessment output type: %w", err)
	}

	if err := outputType.DecorateEnum("confidence", confidenceEnum); err != nil {
		return nil, fmt.Errorf("cannot decorate evidence assessment schema: %w", err)
	}

	return &Assessor{cfg: cfg, outputType: outputType}, nil
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
	// Detach from the caller so an unrelated cancellation cannot abort a
	// committed assessment, but impose a hard deadline so a stalled
	// provider response cannot run forever and pin the worker slot.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), AssessmentTimeout)
	defer cancel()

	opts := []agent.Option{
		agent.WithInstructions(systemPrompt),
		agent.WithModel(a.cfg.Model),
		agent.WithMaxTokens(a.cfg.MaxTokens),
		agent.WithOutputType(a.outputType),
	}
	if a.cfg.Thinking > 0 {
		// Extended thinking and a custom temperature are mutually
		// exclusive on Anthropic (the Messages API rejects any temperature
		// other than 1 when thinking is enabled), and OpenAI reasoning
		// models ignore temperature anyway, so only send a temperature
		// when thinking is off. Mirrors pkg/vetting, which sets thinking
		// without a temperature for the same reason.
		opts = append(opts, agent.WithThinking(a.cfg.Thinking))
	} else {
		opts = append(opts, agent.WithTemperature(a.cfg.Temp))
	}

	if a.cfg.Logger != nil {
		opts = append(opts, agent.WithLogger(a.cfg.Logger))
	}

	assessmentAgent := agent.New("evidence_assessor", a.cfg.Client, opts...)

	result, err := assessmentAgent.Run(
		ctx,
		[]llm.Message{
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
		},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot assess evidence: %w", err)
	}

	text := result.FinalMessage().Text()

	var out EvidenceAssessment
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		// Surface the output size (not the content, which may carry
		// sensitive evidence detail) so a MaxTokens-truncated response is
		// distinguishable from other parse failures in logs.
		return nil, fmt.Errorf("cannot parse evidence assessment (%d bytes of output): %w", len(text), err)
	}

	return &out, nil
}

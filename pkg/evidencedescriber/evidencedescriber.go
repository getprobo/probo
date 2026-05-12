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
	"fmt"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

//go:embed prompt.txt
var systemPrompt string

type (
	Config struct {
		Model     string
		Temp      float64
		MaxTokens int
	}

	Describer struct {
		client *llm.Client
		config Config
	}
)

func New(client *llm.Client, cfg Config) *Describer {
	return &Describer{
		client: client,
		config: cfg,
	}
}

func (d *Describer) Describe(ctx context.Context, filename string, mimeType string, fileBase64 string) (*string, error) {
	ag := agent.New(
		"evidence_describer",
		d.client,
		agent.WithInstructions(systemPrompt),
		agent.WithModel(d.config.Model),
		agent.WithTemperature(d.config.Temp),
		agent.WithMaxTokens(d.config.MaxTokens),
	)

	result, err := ag.Run(
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
		return nil, fmt.Errorf("cannot describe evidence: %w", err)
	}

	text := result.FinalMessage().Text()
	return &text, nil
}

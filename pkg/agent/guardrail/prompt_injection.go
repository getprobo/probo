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

package guardrail

import (
	"context"
	_ "embed"
	"strings"

	"go.gearno.de/kit/log"
	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

//go:embed prompt_injection_classifier.txt
var promptInjectionClassifierPrompt string

type PromptInjectionGuardrail struct {
	client *llm.Client
	logger *log.Logger
}

func NewPromptInjectionGuardrail(client *llm.Client, logger *log.Logger) *PromptInjectionGuardrail {
	return &PromptInjectionGuardrail{client: client, logger: logger}
}

func (g *PromptInjectionGuardrail) Name() string {
	return "prompt-injection"
}

func (g *PromptInjectionGuardrail) Check(ctx context.Context, messages []llm.Message) (*agent.GuardrailResult, error) {
	if len(messages) == 0 {
		return &agent.GuardrailResult{Tripwire: false}, nil
	}

	// Only classify the last user message.
	lastMessage := messages[len(messages)-1]
	if lastMessage.Role != llm.RoleUser {
		return &agent.GuardrailResult{Tripwire: false}, nil
	}

	userText := lastMessage.Text()
	if userText == "" {
		return &agent.GuardrailResult{Tripwire: false}, nil
	}

	resp, err := g.client.ChatCompletion(ctx, &llm.ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []llm.Message{
			{
				Role:  llm.RoleSystem,
				Parts: []llm.Part{llm.TextPart{Text: promptInjectionClassifierPrompt}},
			},
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: userText}},
			},
		},
		MaxTokens: new(10),
	})
	if err != nil {
		// If the classifier fails, allow the message through rather than
		// blocking legitimate users. The system prompt hardening and
		// tool-level authorization provide defense in depth.
		g.logger.WarnCtx(ctx, "prompt injection classifier failed, allowing message through",
			log.Error(err),
		)
		return &agent.GuardrailResult{Tripwire: false}, nil
	}

	responseText := strings.TrimSpace(resp.Message.Text())
	if strings.EqualFold(responseText, "UNSAFE") {
		return &agent.GuardrailResult{
			Tripwire: true,
			Message:  "message classified as prompt injection attempt",
		}, nil
	}

	return &agent.GuardrailResult{Tripwire: false}, nil
}

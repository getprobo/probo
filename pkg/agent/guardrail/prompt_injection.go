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

	resp, err := g.client.ChatCompletion(
		ctx,
		&llm.ChatCompletionRequest{
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
			MaxTokens:   new(10),
			Temperature: new(0.0),
		},
	)
	if err != nil {
		// If the classifier fails, allow the message through rather than
		// blocking legitimate users. The system prompt hardening and
		// tool-level authorization provide defense in depth.
		g.logger.WarnCtx(
			ctx,
			"prompt injection classifier failed, allowing message through",
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

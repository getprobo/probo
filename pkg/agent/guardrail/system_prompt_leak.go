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
	"strings"

	"go.probo.inc/probo/pkg/agent"
	"go.probo.inc/probo/pkg/llm"
)

type SystemPromptLeakGuardrail struct {
	fingerprints []string
}

func NewSystemPromptLeakGuardrail(fingerprints []string) *SystemPromptLeakGuardrail {
	lowered := make([]string, 0, len(fingerprints))
	for _, f := range fingerprints {
		if f == "" {
			continue
		}

		lowered = append(lowered, strings.ToLower(f))
	}

	return &SystemPromptLeakGuardrail{fingerprints: lowered}
}

func (g *SystemPromptLeakGuardrail) Name() string {
	return "system-prompt-leak"
}

func (g *SystemPromptLeakGuardrail) Check(_ context.Context, message llm.Message) (*agent.GuardrailResult, error) {
	text := strings.ToLower(message.Text())

	for _, fp := range g.fingerprints {
		if strings.Contains(text, fp) {
			return &agent.GuardrailResult{
				Tripwire: true,
				Message:  "response contains system prompt content",
			}, nil
		}
	}

	return &agent.GuardrailResult{Tripwire: false}, nil
}

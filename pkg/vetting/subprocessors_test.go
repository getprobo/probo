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

package vetting

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/llm"
)

func TestSubprocessorsFromOrchestratorMessages(t *testing.T) {
	t.Parallel()

	toolJSON := `{"subprocessors":[{"name":"Amazon Web Services","country":"US","purpose":"Cloud hosting"}],"total_count":1,"source":"https://example.com/subprocessors","is_complete":true}`

	messages := []llm.Message{
		{
			Role: llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{{
				ID: "call-1",
				Function: llm.FunctionCall{
					Name: extractSubprocessorsToolName,
				},
			}},
		},
		{
			Role:       llm.RoleTool,
			ToolCallID: "call-1",
			Parts:      []llm.Part{llm.TextPart{Text: toolJSON}},
		},
	}

	subs := subprocessorsFromOrchestratorMessages(messages)

	assert.Equal(
		t,
		[]Subprocessor{{
			Name:    "Amazon Web Services",
			Country: "US",
			Purpose: "Cloud hosting",
		}},
		subs,
	)
}

func TestSubprocessorsFromOrchestratorMessages_LatestCallWins(t *testing.T) {
	t.Parallel()

	messages := []llm.Message{
		{
			Role: llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{
				{
					ID: "call-1",
					Function: llm.FunctionCall{
						Name: extractSubprocessorsToolName,
					},
				},
				{
					ID: "call-2",
					Function: llm.FunctionCall{
						Name: extractSubprocessorsToolName,
					},
				},
			},
		},
		{
			Role:       llm.RoleTool,
			ToolCallID: "call-1",
			Parts:      []llm.Part{llm.TextPart{Text: `{"subprocessors":[{"name":"Stripe","country":"US","purpose":"Payments"}]}`}},
		},
		{
			Role:       llm.RoleTool,
			ToolCallID: "call-2",
			Parts:      []llm.Part{llm.TextPart{Text: `{"subprocessors":[{"name":"Stripe","country":"IE","purpose":"Payment processing"}]}`}},
		},
	}

	subs := subprocessorsFromOrchestratorMessages(messages)

	assert.Equal(
		t,
		[]Subprocessor{{
			Name:    "Stripe",
			Country: "IE",
			Purpose: "Payment processing",
		}},
		subs,
	)
}

func TestSubprocessorsFromOrchestratorMessages_IgnoresOtherTools(t *testing.T) {
	t.Parallel()

	messages := []llm.Message{
		{
			Role: llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{{
				ID: "call-1",
				Function: llm.FunctionCall{
					Name: "assess_security",
				},
			}},
		},
		{
			Role:       llm.RoleTool,
			ToolCallID: "call-1",
			Parts:      []llm.Part{llm.TextPart{Text: `{"subprocessors":[{"name":"Ignored"}]}`}},
		},
	}

	assert.Nil(t, subprocessorsFromOrchestratorMessages(messages))
}

func TestMergeSubprocessors(t *testing.T) {
	t.Parallel()

	toolSubs := []Subprocessor{{
		Name:    "AWS",
		Country: "US",
		Purpose: "Hosting",
	}}
	extractedSubs := []Subprocessor{
		{Name: "AWS", Country: "DE", Purpose: "Wrong"},
		{Name: "SendGrid", Country: "US", Purpose: "Email"},
	}

	merged := mergeSubprocessors(toolSubs, extractedSubs)

	assert.Equal(
		t,
		[]Subprocessor{
			{Name: "AWS", Country: "US", Purpose: "Hosting"},
			{Name: "SendGrid", Country: "US", Purpose: "Email"},
		},
		merged,
	)
}

func TestSubprocessorListURLFromOrchestratorMessages(t *testing.T) {
	t.Parallel()

	messages := []llm.Message{
		{
			Role: llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{{
				ID: "call-1",
				Function: llm.FunctionCall{
					Name: extractSubprocessorsToolName,
				},
			}},
		},
		{
			Role:       llm.RoleTool,
			ToolCallID: "call-1",
			Parts:      []llm.Part{llm.TextPart{Text: `{"subprocessors":[],"source":"https://example.com/legal/subprocessors"}`}},
		},
	}

	assert.Equal(
		t,
		"https://example.com/legal/subprocessors",
		subprocessorListURLFromOrchestratorMessages(messages),
	)
}

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

package anthropic

import (
	"encoding/json"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/llm"
)

func TestResolveThinkingBudget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		maxTokens int
		thinking  *llm.ThinkingConfig
		want      int
		wantOK    bool
	}{
		{
			name:      "disabled",
			maxTokens: 4096,
			thinking:  &llm.ThinkingConfig{Enabled: false, BudgetTokens: 4000},
		},
		{
			name:      "leaves valid budget unchanged",
			maxTokens: 49152,
			thinking:  &llm.ThinkingConfig{Enabled: true, BudgetTokens: 40000},
			want:      40000,
			wantOK:    true,
		},
		{
			name:      "caps oversized default budget",
			maxTokens: 4096,
			thinking:  &llm.ThinkingConfig{Enabled: true, BudgetTokens: 40000},
			want:      3072,
			wantOK:    true,
		},
		{
			name:      "caps oversized raised budget",
			maxTokens: 32768,
			thinking:  &llm.ThinkingConfig{Enabled: true, BudgetTokens: 40000},
			want:      31744,
			wantOK:    true,
		},
		{
			name:      "omits thinking when max tokens cannot hold minimum budget",
			maxTokens: 1024,
			thinking:  &llm.ThinkingConfig{Enabled: true, BudgetTokens: 4000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, ok := resolveThinkingBudget(tt.maxTokens, tt.thinking)

			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildParamsNormalizesThinkingBudget(t *testing.T) {
	t.Parallel()

	maxTokens := 32768
	params, err := buildParams(&llm.ChatCompletionRequest{
		Model: "claude-sonnet-test",
		Messages: []llm.Message{
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: "assess this vendor"}},
			},
		},
		MaxTokens: &maxTokens,
		Thinking: &llm.ThinkingConfig{
			Enabled:      true,
			BudgetTokens: 40000,
		},
	})
	require.NoError(t, err)

	budgetTokens := params.Thinking.GetBudgetTokens()
	require.NotNil(t, budgetTokens)
	assert.Equal(t, int64(31744), *budgetTokens)
	assert.Less(t, *budgetTokens, params.MaxTokens)
}

func TestBuildParamsOmitsInvalidThinkingBudget(t *testing.T) {
	t.Parallel()

	maxTokens := 1024
	params, err := buildParams(&llm.ChatCompletionRequest{
		Model: "claude-sonnet-test",
		Messages: []llm.Message{
			{
				Role:  llm.RoleUser,
				Parts: []llm.Part{llm.TextPart{Text: "assess this vendor"}},
			},
		},
		MaxTokens: &maxTokens,
		Thinking: &llm.ThinkingConfig{
			Enabled:      true,
			BudgetTokens: 4000,
		},
	})
	require.NoError(t, err)

	assert.Nil(t, params.Thinking.GetBudgetTokens())
}

func TestBuildMessagesSkipsForeignThinking(t *testing.T) {
	t.Parallel()

	messages := buildMessages(
		[]llm.Message{
			{
				Role: llm.RoleAssistant,
				Parts: []llm.Part{
					llm.ThinkingPart{Signature: `[{"id":"rs_1","encrypted_content":"enc"}]`, Provider: "openai"},
					llm.ThinkingPart{Text: "legacy", Signature: "sig-legacy"},
					llm.ThinkingPart{Text: "current", Signature: "sig-current", Provider: "anthropic"},
					llm.TextPart{Text: "answer"},
				},
			},
			{
				Role: llm.RoleAssistant,
				Parts: []llm.Part{
					llm.ThinkingPart{Signature: `[{"id":"rs_2","encrypted_content":"enc"}]`, Provider: "openai"},
					llm.TextPart{Text: ""},
				},
			},
		},
	)

	require.Len(t, messages, 1)
	require.Len(t, messages[0].Content, 3)
	assert.Equal(t, "sig-legacy", messages[0].Content[0].OfThinking.Signature)
	assert.Equal(t, "sig-current", messages[0].Content[1].OfThinking.Signature)
	assert.Equal(t, "answer", messages[0].Content[2].OfText.Text)
}

func TestMapResponse_TagsThinkingProvider(t *testing.T) {
	t.Parallel()

	var msg anthropic.Message
	require.NoError(
		t,
		json.Unmarshal(
			[]byte(`{
				"id": "msg_1",
				"type": "message",
				"role": "assistant",
				"model": "claude-sonnet-test",
				"content": [
					{"type": "thinking", "thinking": "reasoning", "signature": "sig"},
					{"type": "text", "text": "answer"}
				],
				"stop_reason": "end_turn",
				"usage": {"input_tokens": 1, "output_tokens": 1}
			}`),
			&msg,
		),
	)

	resp := mapResponse(&msg)

	require.NotEmpty(t, resp.Message.Parts)
	assert.Equal(
		t,
		llm.ThinkingPart{Text: "reasoning", Signature: "sig", Provider: "anthropic"},
		resp.Message.Parts[0],
	)
}

func TestMapStreamEvent_TagsThinkingProvider(t *testing.T) {
	t.Parallel()

	var event anthropic.MessageStreamEventUnion
	require.NoError(
		t,
		json.Unmarshal(
			[]byte(`{"type":"content_block_delta","index":0,"delta":{"type":"signature_delta","signature":"sig"}}`),
			&event,
		),
	)

	stream := &anthropicStream{}
	got, ok := stream.mapStreamEvent(&event)

	require.True(t, ok)
	assert.Equal(
		t,
		llm.MessageDelta{ThinkingSignature: "sig", ThinkingProvider: "anthropic"},
		got.Delta,
	)
}

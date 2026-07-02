// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

package anthropic

import (
	"testing"

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

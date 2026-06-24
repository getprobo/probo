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

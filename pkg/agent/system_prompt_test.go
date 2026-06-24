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

package agent

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildSystemPrompt(t *testing.T) {
	t.Run(
		"empty data returns empty string",
		func(t *testing.T) {
			got := buildSystemPrompt(systemPromptData{})
			assert.Equal(t, "", got)
		},
	)

	t.Run(
		"instructions only",
		func(t *testing.T) {
			got := buildSystemPrompt(systemPromptData{
				Instructions: "You are a helpful assistant.",
			})
			assert.Equal(t, "You are a helpful assistant.", got)
		},
	)

	t.Run(
		"handoffs only",
		func(t *testing.T) {
			got := buildSystemPrompt(systemPromptData{
				Handoffs: []systemPromptHandoff{
					{Name: "billing-agent", Description: "Handles billing questions."},
				},
			})

			assert.Contains(t, got, "## Handoffs")
			assert.Contains(t, got, "- billing-agent: Handles billing questions.")
		},
	)

	t.Run(
		"instructions with handoffs",
		func(t *testing.T) {
			got := buildSystemPrompt(systemPromptData{
				Instructions: "You are a triage agent.",
				Handoffs: []systemPromptHandoff{
					{Name: "billing-agent", Description: "Handles billing."},
					{Name: "support-agent", Description: "Handles support."},
				},
			})

			assert.True(t, strings.HasPrefix(got, "You are a triage agent."))
			assert.Contains(t, got, "## Handoffs")
			assert.Contains(t, got, "- billing-agent: Handles billing.")
			assert.Contains(t, got, "- support-agent: Handles support.")
		},
	)

	t.Run(
		"handoff without description",
		func(t *testing.T) {
			got := buildSystemPrompt(systemPromptData{
				Handoffs: []systemPromptHandoff{
					{Name: "silent-agent"},
				},
			})

			assert.Contains(t, got, "- silent-agent\n")
			assert.NotContains(t, got, "- silent-agent:")
		},
	)

	t.Run(
		"multiple handoffs preserve order",
		func(t *testing.T) {
			got := buildSystemPrompt(systemPromptData{
				Handoffs: []systemPromptHandoff{
					{Name: "alpha"},
					{Name: "beta"},
					{Name: "gamma"},
				},
			})

			idxAlpha := strings.Index(got, "- alpha")
			idxBeta := strings.Index(got, "- beta")
			idxGamma := strings.Index(got, "- gamma")

			assert.Greater(t, idxBeta, idxAlpha)
			assert.Greater(t, idxGamma, idxBeta)
		},
	)

	t.Run(
		"no trailing newline after last handoff",
		func(t *testing.T) {
			got := buildSystemPrompt(systemPromptData{
				Handoffs: []systemPromptHandoff{
					{Name: "agent-a", Description: "Does A."},
				},
			})

			assert.False(t, strings.HasSuffix(got, "\n\n"), "should not end with double newline")
		},
	)
}

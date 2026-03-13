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

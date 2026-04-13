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

package llm_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/llm"
)

func TestRegistry_Lookup(t *testing.T) {
	t.Parallel()

	r := llm.DefaultRegistry()

	t.Run(
		"by full openrouter id",
		func(t *testing.T) {
			t.Parallel()

			m, ok := r.Lookup("anthropic/claude-opus-4.6")
			require.True(t, ok)
			assert.Equal(t, "anthropic", m.Provider)
			assert.Greater(t, m.ContextLength, 0)
			assert.Greater(t, m.MaxOutputTokens, 0)
		},
	)

	t.Run(
		"by bare name with dots",
		func(t *testing.T) {
			t.Parallel()

			m, ok := r.Lookup("claude-opus-4.6")
			require.True(t, ok)
			assert.Equal(t, "anthropic", m.Provider)
		},
	)

	t.Run(
		"by bare name with dashes",
		func(t *testing.T) {
			t.Parallel()

			m, ok := r.Lookup("claude-opus-4-6")
			require.True(t, ok)
			assert.Equal(t, "anthropic", m.Provider)
		},
	)

	t.Run(
		"unknown model returns false",
		func(t *testing.T) {
			t.Parallel()

			_, ok := r.Lookup("nonexistent-model-42")
			assert.False(t, ok)
		},
	)

	t.Run(
		"empty string returns false",
		func(t *testing.T) {
			t.Parallel()

			_, ok := r.Lookup("")
			assert.False(t, ok)
		},
	)

	t.Run(
		"provider prefix only returns false",
		func(t *testing.T) {
			t.Parallel()

			_, ok := r.Lookup("anthropic/")
			assert.False(t, ok)
		},
	)
}

func TestRegistry_Capabilities(t *testing.T) {
	t.Parallel()

	r := llm.DefaultRegistry()

	t.Run(
		"anthropic claude supports temperature and top_p but not frequency_penalty",
		func(t *testing.T) {
			t.Parallel()

			m, ok := r.Lookup("anthropic/claude-opus-4.6")
			require.True(t, ok)
			assert.True(t, m.Supports.Temperature)
			assert.True(t, m.Supports.TopP)
			assert.True(t, m.Supports.TopK)
			assert.True(t, m.Supports.Reasoning)
			assert.False(t, m.Supports.FrequencyPenalty)
			assert.False(t, m.Supports.PresencePenalty)
		},
	)

	t.Run(
		"openai gpt-5.4 is a reasoning model without temperature",
		func(t *testing.T) {
			t.Parallel()

			m, ok := r.Lookup("openai/gpt-5.4")
			require.True(t, ok)
			assert.True(t, m.Supports.Reasoning)
			assert.True(t, m.Supports.Seed)
			assert.False(t, m.Supports.Temperature)
			assert.False(t, m.Supports.TopP)
		},
	)

	t.Run(
		"openai o3 is a reasoning model without temperature",
		func(t *testing.T) {
			t.Parallel()

			m, ok := r.Lookup("openai/o3")
			require.True(t, ok)
			assert.True(t, m.Supports.Reasoning)
			assert.False(t, m.Supports.Temperature)
		},
	)
}

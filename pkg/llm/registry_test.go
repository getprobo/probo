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

	r := llm.NewRegistry(map[string]llm.ModelDefinition{
		"acme/test-model-1.5": {
			Name:            "Acme: Test Model 1.5",
			Provider:        "acme",
			ContextLength:   8192,
			MaxOutputTokens: 4096,
			Supports: llm.SupportedParameters{
				Temperature: true,
				Reasoning:   true,
			},
		},
	})

	t.Run(
		"by full id",
		func(t *testing.T) {
			t.Parallel()

			m, ok := r.Lookup("acme/test-model-1.5")
			require.True(t, ok)
			assert.Equal(t, "acme/test-model-1.5", m.ID)
			assert.Equal(t, "acme", m.Provider)
			assert.Equal(t, 8192, m.ContextLength)
			assert.Equal(t, 4096, m.MaxOutputTokens)
		},
	)

	t.Run(
		"by bare name with dots",
		func(t *testing.T) {
			t.Parallel()

			m, ok := r.Lookup("test-model-1.5")
			require.True(t, ok)
			assert.Equal(t, "acme", m.Provider)
		},
	)

	t.Run(
		"by bare name with dashes",
		func(t *testing.T) {
			t.Parallel()

			m, ok := r.Lookup("test-model-1-5")
			require.True(t, ok)
			assert.Equal(t, "acme", m.Provider)
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

			_, ok := r.Lookup("acme/")
			assert.False(t, ok)
		},
	)
}

func TestRegistry_Capabilities(t *testing.T) {
	t.Parallel()

	r := llm.NewRegistry(map[string]llm.ModelDefinition{
		"acme/reasoning-model": {
			Name:     "Acme: Reasoning Model",
			Provider: "acme",
			Supports: llm.SupportedParameters{
				Reasoning: true,
				Seed:      true,
			},
		},
		"acme/chat-model": {
			Name:     "Acme: Chat Model",
			Provider: "acme",
			Supports: llm.SupportedParameters{
				Temperature:      true,
				TopP:             true,
				TopK:             true,
				FrequencyPenalty: true,
			},
		},
	})

	t.Run(
		"reasoning model has reasoning and seed but not temperature",
		func(t *testing.T) {
			t.Parallel()

			m, ok := r.Lookup("acme/reasoning-model")
			require.True(t, ok)
			assert.True(t, m.Supports.Reasoning)
			assert.True(t, m.Supports.Seed)
			assert.False(t, m.Supports.Temperature)
			assert.False(t, m.Supports.TopP)
		},
	)

	t.Run(
		"chat model has temperature and penalties but not reasoning",
		func(t *testing.T) {
			t.Parallel()

			m, ok := r.Lookup("acme/chat-model")
			require.True(t, ok)
			assert.True(t, m.Supports.Temperature)
			assert.True(t, m.Supports.TopP)
			assert.True(t, m.Supports.TopK)
			assert.True(t, m.Supports.FrequencyPenalty)
			assert.False(t, m.Supports.Reasoning)
		},
	)
}

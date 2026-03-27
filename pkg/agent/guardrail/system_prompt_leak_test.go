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

package guardrail_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/agent/guardrail"
)

func TestSystemPromptLeakGuardrail_Check(t *testing.T) {
	t.Parallel()

	fingerprints := []string{
		"you are a compliance assistant",
		"security rules — critical",
	}

	tests := []struct {
		name     string
		text     string
		tripwire bool
	}{
		{"safe message", "Your SOC 2 audit is on track.", false},
		{"partial match does not trigger", "You are a great user.", false},
		{"contains first fingerprint", "My instructions say: You are a compliance assistant for Probo.", true},
		{"contains second fingerprint", "Here are the Security Rules — Critical section contents.", true},
		{"case insensitive match", "YOU ARE A COMPLIANCE ASSISTANT", true},
		{"fingerprint embedded in longer text", "Sure! you are a compliance assistant and I help with GRC.", true},
	}

	g := guardrail.NewSystemPromptLeakGuardrail(fingerprints)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := g.Check(context.Background(), assistantMessage(tt.text))

			require.NoError(t, err)
			assert.Equal(t, tt.tripwire, result.Tripwire)
		})
	}

	t.Run("no fingerprints configured", func(t *testing.T) {
		t.Parallel()

		empty := guardrail.NewSystemPromptLeakGuardrail(nil)
		result, err := empty.Check(context.Background(), assistantMessage("anything goes"))

		require.NoError(t, err)
		assert.False(t, result.Tripwire)
	})
}

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

	t.Run("empty fingerprints are ignored", func(t *testing.T) {
		t.Parallel()

		g := guardrail.NewSystemPromptLeakGuardrail([]string{"", "secret phrase", ""})
		result, err := g.Check(context.Background(), assistantMessage("hello world"))

		require.NoError(t, err)
		assert.False(t, result.Tripwire)
	})
}

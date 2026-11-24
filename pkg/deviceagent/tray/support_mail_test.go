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

//go:build darwin || windows

package tray

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSupportMailto_RecipientAndEscape(t *testing.T) {
	t.Parallel()

	raw := supportMailto("Probo agent diagnostics (dev, host)", "a&b+c\nline")

	assert.True(t, strings.HasPrefix(raw, "mailto:support@probo.com?"))
	assert.Contains(t, raw, "subject=Probo%20agent%20diagnostics%20%28dev%2C%20host%29")
	assert.Contains(t, raw, "a%26b%2Bc")
	assert.NotContains(t, raw, "+")
	assert.Contains(t, raw, "%0A")
	assert.NotContains(t, raw, "\n")
	assert.NotContains(t, raw, "mailto:other@")
}

func TestDraftMailto_OverflowFallsBack(t *testing.T) {
	t.Parallel()

	t.Run(
		"keeps a short report in the draft",
		func(t *testing.T) {
			t.Parallel()

			raw, overflow := draftMailto("subject", "status ok", 1800)

			assert.False(t, overflow)
			assert.Contains(t, raw, "mailto:support@probo.com?")
			assert.Contains(t, raw, "status%20ok")
		},
	)

	t.Run(
		"rejects an over-long body",
		func(t *testing.T) {
			t.Parallel()

			raw, overflow := draftMailto("subject", strings.Repeat("x", 2000), 1800)

			assert.True(t, overflow)
			assert.Empty(t, raw)
		},
	)
}

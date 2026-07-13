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

package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeSearchQueryForGlob_KnownPatterns(t *testing.T) {
	t.Parallel()

	query, ok := codeSearchQueryForGlob("acme", "*/SECURITY.md")
	require.True(t, ok)
	assert.Equal(t, "org:acme filename:SECURITY.md", query)

	query, ok = codeSearchQueryForGlob("acme", "*/.github/workflows/*.yml")
	require.True(t, ok)
	assert.Equal(t, "org:acme path:.github/workflows+extension:yml", query)
}

func TestCodeSearchQueryForGlob_UnknownPattern(t *testing.T) {
	t.Parallel()

	_, ok := codeSearchQueryForGlob("acme", "*/README.md")
	assert.False(t, ok)
}

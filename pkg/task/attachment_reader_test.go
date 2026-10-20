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

package task

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/filevalidation"
	"go.probo.inc/probo/pkg/prosemirror"
)

func TestNormalizeStoredAttachment_RejectsMismatchedType(t *testing.T) {
	t.Parallel()

	_, _, err := normalizeStoredAttachment("diagram.png", "application/pdf", 9)
	require.Error(t, err)

	var rejected *filevalidation.Error
	require.ErrorAs(t, err, &rejected)
}

func TestNormalizeStoredAttachment_KeepsMatchingType(t *testing.T) {
	t.Parallel()

	name, mimeType, err := normalizeStoredAttachment("diagram.png", "image/png", 9)
	require.NoError(t, err)
	assert.Equal(t, "diagram.png", name)
	assert.Equal(t, "image/png", mimeType)
}

func TestTrimFileName_KeepsRuneBoundary(t *testing.T) {
	t.Parallel()

	name := strings.Repeat("é", 600) + ".png"
	trimmed := trimFileName(name)

	assert.LessOrEqual(t, len(trimmed), prosemirror.MaxFileNameBytes)
	assert.True(t, utf8.ValidString(trimmed))
	assert.True(t, strings.HasSuffix(trimmed, ".png"))
}

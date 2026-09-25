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

package factory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.probo.inc/probo/pkg/prosemirror"
)

// ProseMirrorPlainText wraps plaintext as a ProseMirror document (one
// paragraph per line), matching the task/comment SQL backfill.
func ProseMirrorPlainText(text string) string {
	return prosemirror.FromPlainText(text)
}

// AssertProseMirrorPlainText compares JSONB content to a plaintext document.
// PostgreSQL JSONB reorders keys and adds whitespace on read, so this is a
// semantic JSON comparison rather than a string match.
func AssertProseMirrorPlainText(t testing.TB, expectedPlain, actualJSON string) {
	t.Helper()
	assert.JSONEq(t, ProseMirrorPlainText(expectedPlain), actualJSON)
}

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

package probo_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/probo"
	"go.probo.inc/probo/pkg/prosemirror"
)

func TestRewriteThirdPartyRegisterNotesContent_ConvertsRawMarkdown(t *testing.T) {
	t.Parallel()

	raw := `{
  "type": "doc",
  "content": [
    {
      "type": "heading",
      "attrs": { "level": 3 },
      "content": [{ "type": "text", "text": "2.1.6 Risk Assessments" }]
    },
    {
      "type": "paragraph",
      "content": [
        { "type": "text", "text": "Notes: ", "marks": [{ "type": "bold" }] },
        { "type": "text", "text": "# Third Party Assessment: Nango\n\n## Executive Summary\nApprove.\n\n| Category | Score |\n|----------|-------|\n| Security | 72 |\n\n- **Notable Customers**: Acme" }
      ]
    }
  ]
}`

	out, changed, err := probo.RewriteThirdPartyRegisterNotesContent(raw)
	require.NoError(t, err)
	require.True(t, changed)

	doc, err := prosemirror.Parse(out)
	require.NoError(t, err)

	assert.NotContains(t, out, "# Third Party Assessment")
	assert.NotContains(t, out, "## Executive Summary")
	assert.NotContains(t, out, "| Category")
	assert.NotContains(t, out, "- **Notable Customers**")
	assert.True(t, containsHeading(doc, 4, "Third Party Assessment: Nango"))
	assert.True(t, containsHeading(doc, 5, "Executive Summary"))
	assert.True(t, containsNodeType(doc, prosemirror.NodeTable))
	assert.True(t, containsBoldText(doc, "Notable Customers"))
}

func TestRewriteThirdPartyRegisterNotesContent_LeavesPlainNotes(t *testing.T) {
	t.Parallel()

	raw := `{
  "type": "doc",
  "content": [
    {
      "type": "paragraph",
      "content": [
        { "type": "text", "text": "Notes: ", "marks": [{ "type": "bold" }] },
        { "type": "text", "text": "Not specified" }
      ]
    }
  ]
}`

	out, changed, err := probo.RewriteThirdPartyRegisterNotesContent(raw)
	require.NoError(t, err)
	assert.False(t, changed)
	assert.Equal(t, raw, out)
}

func TestRewriteThirdPartyRegisterNotesContent_Idempotent(t *testing.T) {
	t.Parallel()

	raw := `{
  "type": "doc",
  "content": [
    {
      "type": "paragraph",
      "content": [
        { "type": "text", "text": "Notes: ", "marks": [{ "type": "bold" }] },
        { "type": "text", "text": "## Executive Summary\nLooks good." }
      ]
    }
  ]
}`

	once, changed, err := probo.RewriteThirdPartyRegisterNotesContent(raw)
	require.NoError(t, err)
	require.True(t, changed)

	twice, changedAgain, err := probo.RewriteThirdPartyRegisterNotesContent(once)
	require.NoError(t, err)
	assert.False(t, changedAgain)
	assert.Equal(t, once, twice)
}

func TestRewriteThirdPartyRegisterNotesContent_RewritesSiblingWhenOneLooksLikeMarkdown(t *testing.T) {
	t.Parallel()

	raw := `{
  "type": "doc",
  "content": [
    {
      "type": "paragraph",
      "content": [
        { "type": "text", "text": "Notes: ", "marks": [{ "type": "bold" }] },
        { "type": "text", "text": "Not specified" }
      ]
    },
    {
      "type": "paragraph",
      "content": [
        { "type": "text", "text": "Notes: ", "marks": [{ "type": "bold" }] },
        { "type": "text", "text": "## Executive Summary\nApprove." }
      ]
    }
  ]
}`

	out, changed, err := probo.RewriteThirdPartyRegisterNotesContent(raw)
	require.NoError(t, err)
	require.True(t, changed)

	doc, err := prosemirror.Parse(out)
	require.NoError(t, err)
	assert.Contains(t, collectText(doc), "Not specified")
	assert.True(t, containsHeading(doc, 5, "Executive Summary"))
	assert.Contains(t, collectText(doc), "Approve.")
}

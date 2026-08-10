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

package prosemirror_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
	"go.probo.inc/probo/pkg/prosemirror"
)

// assertCanonicalRenderable proves the rendered JSON only contains node and mark
// types the canonical downstream renderers understand. Those renderers error on
// unknown types, so surviving both is what guarantees an unfamiliar block or mark
// can never abort persistence or publishing.
func assertCanonicalRenderable(t *testing.T, content string) {
	t.Helper()

	node, err := prosemirror.Parse(content)
	require.NoError(t, err)

	_, err = prosemirror.RenderMarkdown(node)
	require.NoError(t, err)

	_, err = prosemirror.RenderHTML(node)
	require.NoError(t, err)
}

func TestRender_UnknownBlockDegradesToParagraph(t *testing.T) {
	t.Parallel()

	content, err := automergeprosemirror.Render(
		[]automerge.Span{
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "callout",
					"parents": []any{},
					"attrs":   map[string]any{"variant": "warning"},
				},
			},
			{Type: automerge.SpanTypeText, Text: "Heads up"},
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "paragraph",
					"parents": []any{},
				},
			},
			{Type: automerge.SpanTypeText, Text: "After"},
		},
	)

	require.NoError(t, err)
	assert.JSONEq(
		t,
		`{
			"type": "doc",
			"content": [
				{"type": "paragraph", "content": [{"type": "text", "text": "Heads up"}]},
				{"type": "paragraph", "content": [{"type": "text", "text": "After"}]}
			]
		}`,
		content,
	)
	assertCanonicalRenderable(t, content)
}

func TestRender_UnknownBlockHoistsChildren(t *testing.T) {
	t.Parallel()

	content, err := automergeprosemirror.Render(
		[]automerge.Span{
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "callout",
					"parents": []any{},
				},
			},
			{Type: automerge.SpanTypeText, Text: "Intro"},
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "paragraph",
					"parents": []any{"callout"},
				},
			},
			{Type: automerge.SpanTypeText, Text: "Child"},
		},
	)

	require.NoError(t, err)
	assert.JSONEq(
		t,
		`{
			"type": "doc",
			"content": [
				{"type": "paragraph", "content": [{"type": "text", "text": "Intro"}]},
				{"type": "paragraph", "content": [{"type": "text", "text": "Child"}]}
			]
		}`,
		content,
	)
	assertCanonicalRenderable(t, content)
}

func TestRender_UnknownMarkIsDropped(t *testing.T) {
	t.Parallel()

	content, err := automergeprosemirror.Render(
		[]automerge.Span{
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "paragraph",
					"parents": []any{},
				},
			},
			{
				Type:  automerge.SpanTypeText,
				Text:  "Text",
				Marks: map[string]any{"highlight": true, "strong": true},
			},
		},
	)

	require.NoError(t, err)
	assert.JSONEq(
		t,
		`{
			"type": "doc",
			"content": [
				{
					"type": "paragraph",
					"content": [
						{"type": "text", "text": "Text", "marks": [{"type": "bold"}]}
					]
				}
			]
		}`,
		content,
	)
	assertCanonicalRenderable(t, content)
}

func TestRender_MalformedLinkMarkIsDropped(t *testing.T) {
	t.Parallel()

	content, err := automergeprosemirror.Render(
		[]automerge.Span{
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "paragraph",
					"parents": []any{},
				},
			},
			{
				Type:  automerge.SpanTypeText,
				Text:  "Text",
				Marks: map[string]any{"link": "not json"},
			},
		},
	)

	require.NoError(t, err)
	assert.JSONEq(
		t,
		`{
			"type": "doc",
			"content": [
				{"type": "paragraph", "content": [{"type": "text", "text": "Text"}]}
			]
		}`,
		content,
	)
	assertCanonicalRenderable(t, content)
}

func TestRender_UnknownMarkOnlyLeavesPlainText(t *testing.T) {
	t.Parallel()

	content, err := automergeprosemirror.Render(
		[]automerge.Span{
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "paragraph",
					"parents": []any{},
				},
			},
			{
				Type:  automerge.SpanTypeText,
				Text:  "Text",
				Marks: map[string]any{"highlight": true},
			},
		},
	)

	require.NoError(t, err)
	assert.JSONEq(
		t,
		`{
			"type": "doc",
			"content": [
				{"type": "paragraph", "content": [{"type": "text", "text": "Text"}]}
			]
		}`,
		content,
	)
	assertCanonicalRenderable(t, content)
}

func TestRender_MalformedStructureNeverAborts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		spans []automerge.Span
		text  string
	}{
		{
			name: "root hard break",
			spans: []automerge.Span{
				{
					Type: automerge.SpanTypeBlock,
					Block: map[string]any{
						"type":    "hard-break",
						"parents": []any{},
						"isEmbed": true,
					},
				},
			},
		},
		{
			name: "missing block type",
			spans: []automerge.Span{
				{Type: automerge.SpanTypeBlock, Block: map[string]any{"parents": []any{}}},
				{Type: automerge.SpanTypeText, Text: "Preserved"},
			},
			text: "Preserved",
		},
		{
			name: "invalid parent value",
			spans: []automerge.Span{
				{
					Type: automerge.SpanTypeBlock,
					Block: map[string]any{
						"type":    "paragraph",
						"parents": []any{"paragraph", 42, nil},
					},
				},
				{Type: automerge.SpanTypeText, Text: "Rooted"},
			},
			text: "Rooted",
		},
		{
			name: "paragraph with child block",
			spans: []automerge.Span{
				testBlock("paragraph", nil),
				{Type: automerge.SpanTypeText, Text: "Parent"},
				testBlock("heading", []any{"paragraph"}),
				{Type: automerge.SpanTypeText, Text: "Child"},
			},
			text: "ParentChild",
		},
		{
			name: "heading with child block",
			spans: []automerge.Span{
				testBlock("heading", nil),
				{Type: automerge.SpanTypeText, Text: "Heading"},
				testBlock("paragraph", []any{"heading"}),
				{Type: automerge.SpanTypeText, Text: "Child"},
			},
			text: "HeadingChild",
		},
		{
			name: "code block with child block",
			spans: []automerge.Span{
				testBlock("code-block", nil),
				{Type: automerge.SpanTypeText, Text: "Code"},
				testBlock("paragraph", []any{"code-block"}),
				{Type: automerge.SpanTypeText, Text: "Child"},
			},
			text: "CodeChild",
		},
		{
			name: "table with non-row child",
			spans: []automerge.Span{
				testBlock("table", nil),
				testBlock("paragraph", []any{"table"}),
				{Type: automerge.SpanTypeText, Text: "Hoisted"},
			},
			text: "Hoisted",
		},
		{
			name: "table row with non-cell child",
			spans: []automerge.Span{
				testBlock("table", nil),
				testBlock("table-row", []any{"table"}),
				testBlock("paragraph", []any{"table", "table-row"}),
				{Type: automerge.SpanTypeText, Text: "Hoisted"},
			},
			text: "Hoisted",
		},
		{
			name: "stray table cell at root",
			spans: []automerge.Span{
				testBlock("table-cell", nil),
				{Type: automerge.SpanTypeText, Text: "Stray"},
			},
			text: "Stray",
		},
		{
			name: "stray table header at root",
			spans: []automerge.Span{
				testBlock("table-header", nil),
				{Type: automerge.SpanTypeText, Text: "Header"},
			},
			text: "Header",
		},
		{
			name: "stray table row with cells at root",
			spans: []automerge.Span{
				testBlock("table-row", nil),
				testBlock("table-cell", []any{"table-row"}),
				{Type: automerge.SpanTypeText, Text: "Cell"},
			},
			text: "Cell",
		},
		{
			name: "table cell skipping its row",
			spans: []automerge.Span{
				testBlock("table", nil),
				testBlock("table-cell", []any{"table"}),
				{Type: automerge.SpanTypeText, Text: "Skipped"},
			},
			text: "Skipped",
		},
		{
			name: "impossible parent chain",
			spans: []automerge.Span{
				testBlock("paragraph", nil),
				{Type: automerge.SpanTypeText, Text: "First"},
				testBlock("paragraph", []any{"table", "table-row", "table-cell"}),
				{Type: automerge.SpanTypeText, Text: "Second"},
			},
			text: "FirstSecond",
		},
		{
			name: "unknown span type",
			spans: []automerge.Span{
				{Type: automerge.SpanType("future")},
				testBlock("paragraph", nil),
				{Type: automerge.SpanTypeText, Text: "Known"},
			},
			text: "Known",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			content, err := automergeprosemirror.Render(tt.spans)
			require.NoError(t, err)
			assertCanonicalRenderable(t, content)

			node, err := prosemirror.Parse(content)
			require.NoError(t, err)
			assert.Equal(t, tt.text, nodeText(node))
		})
	}
}

func testBlock(blockType string, parents []any) automerge.Span {
	return automerge.Span{
		Type: automerge.SpanTypeBlock,
		Block: map[string]any{
			"type":    blockType,
			"parents": parents,
		},
	}
}

func nodeText(node prosemirror.Node) string {
	var text string
	if node.Text != nil {
		text = *node.Text
	}

	for _, child := range node.Content {
		text += nodeText(child)
	}

	return text
}

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

func TestRender_RichText(t *testing.T) {
	t.Parallel()

	content, err := automergeprosemirror.Render(
		[]automerge.Span{
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "heading",
					"parents": []any{},
					"attrs":   map[string]any{"level": float64(2)},
				},
			},
			{
				Type:  automerge.SpanTypeText,
				Text:  "Policy",
				Marks: map[string]any{"strong": true},
			},
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "paragraph",
					"parents": []any{},
					"attrs":   map[string]any{},
				},
			},
			{
				Type: automerge.SpanTypeText,
				Text: "Read ",
			},
			{
				Type: automerge.SpanTypeText,
				Text: "more",
				Marks: map[string]any{
					"link": `{"href":"https://example.com","title":"Example"}`,
				},
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
					"type": "heading",
					"attrs": {"level": 2},
					"content": [{
						"type": "text",
						"text": "Policy",
						"marks": [{"type": "bold"}]
					}]
				},
				{
					"type": "paragraph",
					"content": [
						{"type": "text", "text": "Read "},
						{
							"type": "text",
							"text": "more",
							"marks": [{
								"type": "link",
								"attrs": {
									"href": "https://example.com",
									"title": "Example"
								}
							}]
						}
					]
				}
			]
		}`,
		content,
	)

	_, err = prosemirror.Parse(content)
	require.NoError(t, err)
}

func TestRender_NestedList(t *testing.T) {
	t.Parallel()

	content, err := automergeprosemirror.Render(
		[]automerge.Span{
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "unordered-list-item",
					"parents": []any{},
				},
			},
			{Type: automerge.SpanTypeText, Text: "Outer"},
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "unordered-list-item",
					"parents": []any{"unordered-list-item"},
				},
			},
			{Type: automerge.SpanTypeText, Text: "Inner"},
		},
	)

	require.NoError(t, err)
	assert.JSONEq(
		t,
		`{
			"type": "doc",
			"content": [{
				"type": "bulletList",
				"content": [{
					"type": "listItem",
					"content": [
						{
							"type": "paragraph",
							"content": [{"type": "text", "text": "Outer"}]
						},
						{
							"type": "bulletList",
							"content": [{
								"type": "listItem",
								"content": [{
									"type": "paragraph",
									"content": [{"type": "text", "text": "Inner"}]
								}]
							}]
						}
					]
				}]
			}]
		}`,
		content,
	)

	_, err = prosemirror.Parse(content)
	require.NoError(t, err)
}

func TestRender_HardBreakAndHorizontalRule(t *testing.T) {
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
			{Type: automerge.SpanTypeText, Text: "A"},
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "hard-break",
					"parents": []any{"paragraph"},
					"isEmbed": true,
				},
			},
			{Type: automerge.SpanTypeText, Text: "B"},
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "horizontal-rule",
					"parents": []any{},
					"isEmbed": true,
				},
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
						{"type": "text", "text": "A"},
						{"type": "hardBreak"},
						{"type": "text", "text": "B"}
					]
				},
				{"type": "horizontalRule"}
			]
		}`,
		content,
	)

	_, err = prosemirror.Parse(content)
	require.NoError(t, err)
}

func TestRender_HorizontalRuleFollowedByText(t *testing.T) {
	t.Parallel()

	content, err := automergeprosemirror.Render(
		[]automerge.Span{
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "horizontal-rule",
					"parents": []any{},
				},
			},
			{Type: automerge.SpanTypeText, Text: "Preserved"},
		},
	)

	require.NoError(t, err)
	assert.JSONEq(
		t,
		`{
			"type": "doc",
			"content": [
				{"type": "horizontalRule"},
				{
					"type": "paragraph",
					"content": [{"type": "text", "text": "Preserved"}]
				}
			]
		}`,
		content,
	)

	_, err = prosemirror.Parse(content)
	require.NoError(t, err)
}

func TestRender_TableStructure(t *testing.T) {
	t.Parallel()

	content, err := automergeprosemirror.Render(
		[]automerge.Span{
			tableBlock("table", nil, nil),
			tableBlock("table-row", []any{"table"}, nil),
			tableBlock(
				"table-header",
				[]any{"table", "table-row"},
				map[string]any{
					"colspan":  float64(1),
					"rowspan":  float64(1),
					"colwidth": nil,
				},
			),
			{Type: automerge.SpanTypeText, Text: "A"},
			tableBlock(
				"table-cell",
				[]any{"table", "table-row"},
				map[string]any{
					"colspan":  float64(2),
					"rowspan":  float64(1),
					"colwidth": []any{float64(100), float64(120)},
				},
			),
			{Type: automerge.SpanTypeText, Text: "B"},
			tableBlock("table-row", []any{"table"}, nil),
			tableBlock(
				"table-cell",
				[]any{"table", "table-row"},
				map[string]any{
					"colspan":  float64(1),
					"rowspan":  float64(1),
					"colwidth": nil,
				},
			),
			{Type: automerge.SpanTypeText, Text: "C"},
		},
	)

	require.NoError(t, err)
	assert.JSONEq(
		t,
		`{
			"type": "doc",
			"content": [{
				"type": "table",
				"content": [
					{
						"type": "tableRow",
						"content": [
							{
								"type": "tableHeader",
								"attrs": {
									"colspan": 1,
									"rowspan": 1,
									"colwidth": null
								},
								"content": [{
									"type": "paragraph",
									"content": [{"type": "text", "text": "A"}]
								}]
							},
							{
								"type": "tableCell",
								"attrs": {
									"colspan": 2,
									"rowspan": 1,
									"colwidth": [100, 120]
								},
								"content": [{
									"type": "paragraph",
									"content": [{"type": "text", "text": "B"}]
								}]
							}
						]
					},
					{
						"type": "tableRow",
						"content": [{
							"type": "tableCell",
							"attrs": {
								"colspan": 1,
								"rowspan": 1,
								"colwidth": null
							},
							"content": [{
								"type": "paragraph",
								"content": [{"type": "text", "text": "C"}]
							}]
						}]
					}
				]
			}]
		}`,
		content,
	)

	_, err = prosemirror.Parse(content)
	require.NoError(t, err)
}

func TestRender_CodeBlockLanguage(t *testing.T) {
	t.Parallel()

	content, err := automergeprosemirror.Render(
		[]automerge.Span{
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "code-block",
					"parents": []any{},
					"attrs": map[string]any{
						"language": "mermaid",
					},
				},
			},
			{
				Type: automerge.SpanTypeText,
				Text: "graph TD; A-->B",
			},
		},
	)

	require.NoError(t, err)
	assert.JSONEq(
		t,
		`{
			"type": "doc",
			"content": [{
				"type": "codeBlock",
				"attrs": {"language": "mermaid"},
				"content": [{
					"type": "text",
					"text": "graph TD; A-->B"
				}]
			}]
		}`,
		content,
	)

	_, err = prosemirror.Parse(content)
	require.NoError(t, err)
}

func TestRender_BlockquoteWithExplicitParagraph(t *testing.T) {
	t.Parallel()

	content, err := automergeprosemirror.Render(
		[]automerge.Span{
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "blockquote",
					"parents": []any{},
				},
			},
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type":    "paragraph",
					"parents": []any{"blockquote"},
				},
			},
			{Type: automerge.SpanTypeText, Text: "Only child"},
		},
	)

	require.NoError(t, err)
	assert.JSONEq(
		t,
		`{
			"type": "doc",
			"content": [{
				"type": "blockquote",
				"content": [{
					"type": "paragraph",
					"content": [{"type": "text", "text": "Only child"}]
				}]
			}]
		}`,
		content,
	)
}

func tableBlock(
	blockType string,
	parents []any,
	attrs map[string]any,
) automerge.Span {
	if attrs == nil {
		attrs = map[string]any{}
	}

	return automerge.Span{
		Type: automerge.SpanTypeBlock,
		Block: map[string]any{
			"type":    blockType,
			"parents": parents,
			"attrs":   attrs,
		},
	}
}

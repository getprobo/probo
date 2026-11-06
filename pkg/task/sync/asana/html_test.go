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

package asana

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/prosemirror"
	"go.probo.inc/probo/pkg/task/sync/shared"
)

func asanaOrderedList(t *testing.T, start *int, items ...string) prosemirror.Node {
	t.Helper()

	content := make([]prosemirror.Node, 0, len(items))

	for _, item := range items {
		content = append(
			content,
			prosemirror.Node{
				Type: prosemirror.NodeListItem,
				Content: []prosemirror.Node{
					{
						Type: prosemirror.NodeParagraph,
						Content: []prosemirror.Node{
							shared.TextNode(item, nil),
						},
					},
				},
			},
		)
	}

	node := prosemirror.Node{
		Type:    prosemirror.NodeOrderedList,
		Content: content,
	}

	if start != nil {
		raw, err := json.Marshal(map[string]int{"start": *start})
		require.NoError(t, err)

		node.Attrs = raw
	}

	return node
}

func TestRenderAsanaOrderedList_PreservesStart(t *testing.T) {
	t.Parallel()

	t.Run("default start stays a plain list", func(t *testing.T) {
		t.Parallel()

		got, err := renderAsanaBlock(asanaOrderedList(t, nil, "alpha", "beta"), nil, "")
		require.NoError(t, err)
		assert.Equal(t, "<ol><li>alpha</li><li>beta</li></ol>", got)
	})

	t.Run("explicit start one stays a plain list", func(t *testing.T) {
		t.Parallel()

		start := 1
		got, err := renderAsanaBlock(asanaOrderedList(t, &start, "alpha", "beta"), nil, "")
		require.NoError(t, err)
		assert.Equal(t, "<ol><li>alpha</li><li>beta</li></ol>", got)
	})

	t.Run("start five numbers the items", func(t *testing.T) {
		t.Parallel()

		start := 5
		got, err := renderAsanaBlock(asanaOrderedList(t, &start, "alpha", "beta"), nil, "")
		require.NoError(t, err)
		assert.Equal(t, "<ul><li>5. alpha</li><li>6. beta</li></ul>", got)
	})

	t.Run("configured start zero is kept", func(t *testing.T) {
		t.Parallel()

		start := 0
		got, err := renderAsanaBlock(asanaOrderedList(t, &start, "alpha", "beta"), nil, "")
		require.NoError(t, err)
		assert.Equal(t, "<ul><li>0. alpha</li><li>1. beta</li></ul>", got)
	})
}

func TestRenderAsanaText_SanitizesLinkHref(t *testing.T) {
	t.Parallel()

	node := shared.TextNode(
		"click",
		[]prosemirror.Mark{{
			Type:  prosemirror.MarkLink,
			Attrs: json.RawMessage(`{"href":"javascript:alert(1)"}`),
		}},
	)

	got, err := renderAsanaText(node)
	require.NoError(t, err)
	assert.Equal(t, `<a href="#">click</a>`, got)

	node = shared.TextNode(
		"click",
		[]prosemirror.Mark{{
			Type:  prosemirror.MarkLink,
			Attrs: json.RawMessage(`{"href":"https://example.com/x"}`),
		}},
	)

	got, err = renderAsanaText(node)
	require.NoError(t, err)
	assert.Equal(t, `<a href="https://example.com/x">click</a>`, got)
}

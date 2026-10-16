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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/prosemirror"
)

func TestApplyFragmentEdits_ReplacesOneParagraph(t *testing.T) {
	t.Parallel()

	doc := mustParseDoc(t, `{
		"type":"doc",
		"content":[
			{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"Policy"}]},
			{"type":"paragraph","content":[{"type":"text","text":"Old"}]},
			{"type":"paragraph","content":[{"type":"text","text":"Keep"}]}
		]
	}`)

	target, err := prosemirror.ParseFragment(`{"type":"paragraph","content":[{"type":"text","text":"Old"}]}`)
	require.NoError(t, err)
	replacement, err := prosemirror.ParseFragment(`[{"type":"paragraph","content":[{"type":"text","text":"New"}]}]`)
	require.NoError(t, err)

	out, err := prosemirror.ApplyFragmentEdits(
		doc,
		[]prosemirror.FragmentEdit{{Target: target, Replacement: replacement}},
	)
	require.NoError(t, err)
	require.Len(t, out.Content, 3)
	assert.Equal(t, "New", *out.Content[1].Content[0].Text)
	assert.Equal(t, "Keep", *out.Content[2].Content[0].Text)
	assert.Equal(t, prosemirror.NodeHeading, out.Content[0].Type)
}

func TestApplyFragmentEdits_ConflictWhenMissing(t *testing.T) {
	t.Parallel()

	doc := mustParseDoc(t, `{
		"type":"doc",
		"content":[{"type":"paragraph","content":[{"type":"text","text":"Only"}]}]
	}`)

	target, err := prosemirror.ParseFragment(`{"type":"paragraph","content":[{"type":"text","text":"Gone"}]}`)
	require.NoError(t, err)

	_, err = prosemirror.ApplyFragmentEdits(
		doc,
		[]prosemirror.FragmentEdit{{Target: target, Replacement: nil}},
	)
	require.ErrorIs(t, err, prosemirror.ErrFragmentEditNotFound)
}

func TestApplyFragmentEdits_AmbiguousWhenDuplicated(t *testing.T) {
	t.Parallel()

	doc := mustParseDoc(t, `{
		"type":"doc",
		"content":[
			{"type":"paragraph","content":[{"type":"text","text":"Same"}]},
			{"type":"paragraph","content":[{"type":"text","text":"Same"}]}
		]
	}`)

	target, err := prosemirror.ParseFragment(`{"type":"paragraph","content":[{"type":"text","text":"Same"}]}`)
	require.NoError(t, err)

	_, err = prosemirror.ApplyFragmentEdits(
		doc,
		[]prosemirror.FragmentEdit{{Target: target, Replacement: nil}},
	)
	require.ErrorIs(t, err, prosemirror.ErrFragmentEditAmbiguous)
}

func mustParseDoc(t *testing.T, raw string) prosemirror.Node {
	t.Helper()

	node, err := prosemirror.Parse(raw)
	require.NoError(t, err)

	encoded, err := json.Marshal(node)
	require.NoError(t, err)
	require.NoError(t, prosemirror.ValidateDocumentContentJSON(string(encoded)))

	return node
}

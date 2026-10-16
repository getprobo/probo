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

package probo

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/prosemirror"
)

func TestWriteCollaborationSpansThenFragmentEdit(t *testing.T) {
	t.Parallel()

	actor, err := automerge.NewActorID()
	require.NoError(t, err)

	document, err := automerge.New(actor)
	require.NoError(t, err)
	t.Cleanup(func() { _ = document.Close() })

	initial := `{
		"type":"doc",
		"content":[
			{"type":"heading","attrs":{"level":1},"content":[{"type":"text","text":"Policy"}]},
			{"type":"paragraph","content":[{"type":"text","text":"Old body"}]},
			{"type":"paragraph","content":[{"type":"text","text":"Keep me"}]}
		]
	}`

	require.NoError(t, writeCollaborationSpans(document, mustCompactJSON(t, initial), true))
	_, err = document.Commit("seed", time.Now())
	require.NoError(t, err)

	heads, err := document.Heads()
	require.NoError(t, err)
	require.NotEmpty(t, heads)

	content, err := materializeCollaboration(t.Context(), document)
	require.NoError(t, err)

	doc, err := prosemirror.Parse(content)
	require.NoError(t, err)

	target, err := prosemirror.ParseFragment(
		`{"type":"paragraph","content":[{"type":"text","text":"Old body"}]}`,
	)
	require.NoError(t, err)
	replacement, err := prosemirror.ParseFragment(
		`[{"type":"paragraph","content":[{"type":"text","text":"New body"}]}]`,
	)
	require.NoError(t, err)

	updated, err := prosemirror.ApplyFragmentEdits(
		doc,
		[]prosemirror.FragmentEdit{{Target: target, Replacement: replacement}},
	)
	require.NoError(t, err)

	encoded, err := json.Marshal(updated)
	require.NoError(t, err)
	require.NoError(t, writeCollaborationSpans(document, string(encoded), false))
	_, err = document.Commit("edit", time.Now())
	require.NoError(t, err)

	final, err := materializeCollaboration(t.Context(), document)
	require.NoError(t, err)

	finalDoc, err := prosemirror.Parse(final)
	require.NoError(t, err)
	require.Len(t, finalDoc.Content, 3)
	assert.Equal(t, "Policy", *finalDoc.Content[0].Content[0].Text)
	assert.Equal(t, "New body", *finalDoc.Content[1].Content[0].Text)
	assert.Equal(t, "Keep me", *finalDoc.Content[2].Content[0].Text)

	hasHeads, err := document.HasHeads(heads)
	require.NoError(t, err)
	assert.True(t, hasHeads)
}

func TestDecodeAutomergeHeads(t *testing.T) {
	t.Parallel()

	var a, b automerge.Hash
	a[0] = 1
	b[0] = 2

	encoded := append(append([]byte{}, a[:]...), b[:]...)
	heads, err := decodeAutomergeHeads(encoded)
	require.NoError(t, err)
	require.Equal(t, []automerge.Hash{a, b}, heads)

	_, err = decodeAutomergeHeads([]byte{1, 2, 3})
	require.Error(t, err)
}

func TestFragmentEditKeepsTableOutsideTarget(t *testing.T) {
	t.Parallel()

	actor, err := automerge.NewActorID()
	require.NoError(t, err)
	document, err := automerge.New(actor)
	require.NoError(t, err)
	t.Cleanup(func() { _ = document.Close() })

	initial := `{
		"type":"doc",
		"content":[
			{"type":"paragraph","content":[{"type":"text","text":"Intro"}]},
			{"type":"table","content":[
				{"type":"tableRow","content":[
					{"type":"tableCell","content":[{"type":"paragraph","content":[{"type":"text","text":"Cell"}]}]}
				]}
			]}
		]
	}`
	require.NoError(t, writeCollaborationSpans(document, mustCompactJSON(t, initial), true))
	_, err = document.Commit("seed", time.Now())
	require.NoError(t, err)

	content, err := materializeCollaboration(t.Context(), document)
	require.NoError(t, err)
	doc, err := prosemirror.Parse(content)
	require.NoError(t, err)

	target, err := prosemirror.ParseFragment(
		`{"type":"paragraph","content":[{"type":"text","text":"Intro"}]}`,
	)
	require.NoError(t, err)
	replacement, err := prosemirror.ParseFragment(
		`[{"type":"paragraph","content":[{"type":"text","text":"Changed"}]}]`,
	)
	require.NoError(t, err)

	updated, err := prosemirror.ApplyFragmentEdits(
		doc,
		[]prosemirror.FragmentEdit{{Target: target, Replacement: replacement}},
	)
	require.NoError(t, err)

	encoded, err := json.Marshal(updated)
	require.NoError(t, err)
	require.NoError(t, writeCollaborationSpans(document, string(encoded), false))
	_, err = document.Commit("edit", time.Now())
	require.NoError(t, err)

	final, err := materializeCollaboration(t.Context(), document)
	require.NoError(t, err)
	finalDoc, err := prosemirror.Parse(final)
	require.NoError(t, err)
	require.Len(t, finalDoc.Content, 2)
	assert.Equal(t, "Changed", *finalDoc.Content[0].Content[0].Text)
	assert.Equal(t, prosemirror.NodeTable, finalDoc.Content[1].Type)
}

func mustCompactJSON(t *testing.T, raw string) string {
	t.Helper()

	var value any
	require.NoError(t, json.Unmarshal([]byte(raw), &value))
	encoded, err := json.Marshal(value)
	require.NoError(t, err)

	return string(encoded)
}

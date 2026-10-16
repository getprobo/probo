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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
	"go.probo.inc/probo/pkg/gid"
)

func TestGroupCollaborationHistorySessions_SplitsOnActorAndGap(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	actorA := []byte{0x01}
	actorB := []byte{0x02}
	base := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

	hash := func(b byte) automerge.Hash {
		var h automerge.Hash
		h[0] = b
		return h
	}

	changes := []automerge.ChangeInfo{
		{Hash: hash(1), Actor: actorA, Time: base},
		{Hash: hash(2), Actor: actorA, Time: base.Add(time.Minute)},
		{Hash: hash(3), Actor: actorB, Time: base.Add(2 * time.Minute)},
		{Hash: hash(4), Actor: actorA, Time: base.Add(20 * time.Minute)},
	}

	actors := map[string]collaborationActorAttribution{
		string(actorA): {AuthorName: "Ada"},
		string(actorB): {AuthorName: "Bob"},
	}

	sessions := groupCollaborationHistorySessions(tenantID, changes, actors)
	require.Len(t, sessions, 3)

	assert.Equal(t, "Ada", sessions[0].AuthorName)
	assert.Equal(t, 1, sessions[0].ChangeCount)
	assert.Equal(t, []automerge.Hash{hash(4)}, sessions[0].Heads)

	assert.Equal(t, "Bob", sessions[1].AuthorName)
	assert.Equal(t, 1, sessions[1].ChangeCount)
	assert.Equal(t, []automerge.Hash{hash(3)}, sessions[1].Heads)
	assert.Equal(t, []automerge.Hash{hash(2)}, sessions[1].PreviousHeads)

	assert.Equal(t, "Ada", sessions[2].AuthorName)
	assert.Equal(t, 2, sessions[2].ChangeCount)
	assert.Equal(t, []automerge.Hash{hash(2)}, sessions[2].Heads)
	assert.Empty(t, sessions[2].PreviousHeads)
}

func TestGroupCollaborationHistorySessions_UnmappedActor(t *testing.T) {
	t.Parallel()

	tenantID := gid.NewTenantID()
	actor := []byte{0x09}
	base := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)

	var hash automerge.Hash
	hash[0] = 1

	sessions := groupCollaborationHistorySessions(
		tenantID,
		[]automerge.ChangeInfo{{Hash: hash, Actor: actor, Time: base}},
		map[string]collaborationActorAttribution{},
	)
	require.Len(t, sessions, 1)
	assert.Empty(t, sessions[0].AuthorName)
	assert.Nil(t, sessions[0].IdentityID)
}

func TestCollaborationHistoryHighlights_InsertAndDelete(t *testing.T) {
	t.Parallel()

	actor, err := automerge.NewActorID()
	require.NoError(t, err)

	document, err := automerge.New(actor)
	require.NoError(t, err)
	t.Cleanup(func() { _ = document.Close() })

	text, err := document.CreateText("body")
	require.NoError(t, err)

	base := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	require.NoError(t, text.Splice(0, 0, "hello world"))
	_, err = document.Commit("seed", base)
	require.NoError(t, err)

	before, err := document.Heads()
	require.NoError(t, err)

	require.NoError(t, text.Splice(6, 5, "there"))
	_, err = document.Commit("edit", base.Add(time.Minute))
	require.NoError(t, err)

	after, err := document.Heads()
	require.NoError(t, err)

	beforeText, err := text.StringAt(before)
	require.NoError(t, err)
	assert.Equal(t, "hello world", beforeText)

	afterText, err := text.StringAt(after)
	require.NoError(t, err)
	assert.Equal(t, "hello there", afterText)

	highlights, err := collaborationHistoryHighlights(document, before, after)
	require.NoError(t, err)
	require.NotEmpty(t, highlights)

	var sawInsert, sawDelete bool
	for _, highlight := range highlights {
		switch highlight.Kind {
		case DocumentCollaborationHistoryHighlightKindInsert:
			sawInsert = true
			assert.Equal(t, "there", highlight.Text)
		case DocumentCollaborationHistoryHighlightKindDelete:
			sawDelete = true
			assert.Equal(t, "world", highlight.Text)
		}
	}

	assert.True(t, sawInsert, "expected insert highlight")
	assert.True(t, sawDelete, "expected delete highlight")
}

func TestMaterializeAndRestoreViaUpdateSpans(t *testing.T) {
	t.Parallel()

	actor, err := automerge.NewActorID()
	require.NoError(t, err)

	document, err := automerge.New(actor)
	require.NoError(t, err)
	t.Cleanup(func() { _ = document.Close() })

	text, err := document.CreateText("body")
	require.NoError(t, err)

	base := time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	const firstContent = `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"alpha"}]}]}`
	const secondContent = `{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"alpha beta"}]}]}`

	firstSpans, err := automergeprosemirror.ToSpans(firstContent)
	require.NoError(t, err)
	require.NoError(t, text.UpdateSpans(firstSpans, automergeprosemirror.UpdateSpansConfig()))
	first, err := document.Commit("first", base)
	require.NoError(t, err)

	secondSpans, err := automergeprosemirror.ToSpans(secondContent)
	require.NoError(t, err)
	require.NoError(t, text.UpdateSpans(secondSpans, automergeprosemirror.UpdateSpansConfig()))
	_, err = document.Commit("second", base.Add(time.Minute))
	require.NoError(t, err)

	content, err := materializeCollaborationAtHeads(document, []automerge.Hash{first})
	require.NoError(t, err)
	assert.Contains(t, content, "alpha")
	assert.NotContains(t, content, "beta")

	restoreSpans, err := automergeprosemirror.ToSpans(content)
	require.NoError(t, err)
	require.NoError(t, text.UpdateSpans(restoreSpans, automergeprosemirror.UpdateSpansConfig()))
	_, err = document.Commit("restore", base.Add(2*time.Minute))
	require.NoError(t, err)

	current, err := document.Heads()
	require.NoError(t, err)
	restored, err := materializeCollaborationAtHeads(document, current)
	require.NoError(t, err)
	assert.Contains(t, restored, "alpha")
	assert.NotContains(t, restored, "beta")
}

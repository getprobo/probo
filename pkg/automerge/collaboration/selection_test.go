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

package collaboration_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/automerge/collaboration"
)

// TestTextSelectionValue_SurvivesConcurrentInsert is the reason presence carries
// cursors rather than offsets: it round-trips a caret through the presence
// envelope, then edits the text in front of the caret and shows the same cursor
// bytes still resolve to the same character. An integer offset would be stale.
func TestTextSelectionValue_SurvivesConcurrentInsert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	document, err := automerge.New(ctx, actor(1))
	require.NoError(t, err)
	defer func() { _ = document.Close(ctx) }()

	text, err := document.CreateText(ctx, "body")
	require.NoError(t, err)
	require.NoError(t, text.Splice(ctx, 0, 0, "hello world"))
	_, err = document.Commit(ctx, "seed", commitTime())
	require.NoError(t, err)

	// Put the caret on the "w" of "world".
	const caretIndex = 6

	cursor, err := text.Cursor(ctx, caretIndex)
	require.NoError(t, err)

	selection := collaboration.TextSelectionValue{
		Field:  "body",
		Anchor: cursor,
		Head:   cursor,
	}
	assert.True(t, selection.Collapsed())

	message, err := collaboration.NewTextSelectionPresence("", selection)
	require.NoError(t, err)
	assert.Equal(t, collaboration.TextSelectionChannel, message.Channel)

	payload, err := collaboration.EncodePresence(message)
	require.NoError(t, err)

	// The remote side decodes the presence frame back into a selection.
	decodedMessage, err := collaboration.DecodePresence(payload)
	require.NoError(t, err)

	decoded, err := decodedMessage.TextSelection()
	require.NoError(t, err)
	assert.Equal(t, "body", decoded.Field)
	assert.True(t, decoded.Collapsed())

	positionBefore, err := text.CursorPosition(ctx, automerge.Cursor(decoded.Head))
	require.NoError(t, err)
	assert.Equal(t, uint32(caretIndex), positionBefore)

	// Someone types three characters at the very start of the document.
	require.NoError(t, text.Splice(ctx, 0, 0, "XX "))
	_, err = document.Commit(ctx, "insert", commitTime())
	require.NoError(t, err)

	// The very same cursor bytes now resolve three positions later: the caret
	// stayed anchored to "w". A stored integer offset of 6 would now point at
	// the wrong character.
	positionAfter, err := text.CursorPosition(ctx, automerge.Cursor(decoded.Head))
	require.NoError(t, err)
	assert.Equal(t, positionBefore+3, positionAfter)

	// The text is ASCII, so the UTF-16 position equals the byte index.
	value, err := text.String(ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(value), int(positionAfter)+1)
	assert.Equal(t, byte('w'), value[positionAfter])
}

// TestTextSelectionValue_Validation rejects selections missing a field or a
// cursor on both encode and decode.
func TestTextSelectionValue_Validation(t *testing.T) {
	t.Parallel()

	cursor := []byte{1, 2, 3}

	cases := map[string]collaboration.TextSelectionValue{
		"missing field":  {Anchor: cursor, Head: cursor},
		"missing anchor": {Field: "body", Head: cursor},
		"missing head":   {Field: "body", Anchor: cursor},
	}

	for name, selection := range cases {
		t.Run(
			name,
			func(t *testing.T) {
				t.Parallel()

				_, err := selection.Encode()
				assert.Error(t, err)

				_, err = collaboration.NewTextSelectionPresence("", selection)
				assert.Error(t, err)
			},
		)
	}
}

// TestPresenceMessage_TextSelectionRejectsNonUpdate refuses to read a selection
// from a heartbeat or goodbye.
func TestPresenceMessage_TextSelectionRejectsNonUpdate(t *testing.T) {
	t.Parallel()

	message := collaboration.PresenceMessage{Type: collaboration.PresenceHeartbeat}
	_, err := message.TextSelection()
	assert.Error(t, err)
}

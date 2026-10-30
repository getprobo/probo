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

package collaboration

import (
	"context"
	"strings"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/fxamacker/cbor/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	collabpkg "go.probo.inc/probo/pkg/automerge/collaboration"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

func TestProjectAcceptedSpans_DropsInsertionsAndKeepsDeletions(t *testing.T) {
	t.Parallel()

	spans := []automerge.Span{
		{Type: automerge.SpanTypeBlock, Block: map[string]any{"type": "paragraph"}},
		{Type: automerge.SpanTypeText, Text: "Keep "},
		{
			Type: automerge.SpanTypeText,
			Text: "inserted",
			Marks: map[string]any{
				SuggestionInsertMark: `{"id":"s1"}`,
			},
		},
		{
			Type: automerge.SpanTypeText,
			Text: "gone",
			Marks: map[string]any{
				SuggestionDeleteMark: `{"id":"s1"}`,
				"strong":             true,
			},
		},
	}

	projected := ProjectAcceptedSpans(spans)
	require.Len(t, projected, 3)
	assert.Equal(t, "Keep ", projected[1].Text)
	assert.Equal(t, "gone", projected[2].Text)
	assert.NotContains(t, projected[2].Marks, SuggestionDeleteMark)
	assert.Equal(t, true, projected[2].Marks["strong"])
	assert.True(t, HasOpenSuggestions(spans))
	assert.False(t, HasOpenSuggestions(projected))
}

func TestCommentChangeAllowed_AcceptsSuggestionShapedEdits(t *testing.T) {
	t.Parallel()

	before := []automerge.Span{
		{Type: automerge.SpanTypeBlock, Block: map[string]any{"type": "paragraph"}},
		{Type: automerge.SpanTypeText, Text: "Hello"},
	}
	inserted := []automerge.Span{
		before[0],
		before[1],
		{
			Type: automerge.SpanTypeText,
			Text: "!",
			Marks: map[string]any{
				SuggestionInsertMark: `{"id":"s1"}`,
			},
		},
	}
	deleted := []automerge.Span{
		before[0],
		{
			Type: automerge.SpanTypeText,
			Text: "Hello",
			Marks: map[string]any{
				SuggestionDeleteMark: `{"id":"s2"}`,
			},
		},
	}
	directDelete := []automerge.Span{
		before[0],
		{Type: automerge.SpanTypeText, Text: "Hell"},
	}

	open := map[string]bool{"s1": true, "s2": true}

	assert.True(t, CommentChangeAllowed(before, inserted, open))
	assert.True(t, CommentChangeAllowed(before, deleted, open))
	assert.False(t, CommentChangeAllowed(before, directDelete, open))
	assert.False(t, CommentChangeAllowed(inserted, before, open))
}

func TestCommentChangeAllowed_RejectsUntrackedSuggestionID(t *testing.T) {
	t.Parallel()

	before := []automerge.Span{
		{Type: automerge.SpanTypeBlock, Block: map[string]any{"type": "paragraph"}},
		{Type: automerge.SpanTypeText, Text: "Hello"},
	}
	inserted := []automerge.Span{
		before[0],
		before[1],
		{
			Type: automerge.SpanTypeText,
			Text: "!",
			Marks: map[string]any{
				SuggestionInsertMark: `{"id":"phantom"}`,
			},
		},
	}

	assert.False(t, CommentChangeAllowed(before, inserted, map[string]bool{}))
	assert.False(t, CommentChangeAllowed(before, inserted, map[string]bool{"other": true}))
	assert.True(t, CommentChangeAllowed(before, inserted, map[string]bool{"phantom": true}))
}

func TestCommentChangeAllowed_ParagraphSplit(t *testing.T) {
	t.Parallel()

	before := []automerge.Span{
		paragraphBlock(),
		{Type: automerge.SpanTypeText, Text: "HelloWorld"},
	}
	split := []automerge.Span{
		paragraphBlock(),
		{Type: automerge.SpanTypeText, Text: "Hello"},
		splitBlock("s1", suggestionSplitStructural),
		{Type: automerge.SpanTypeText, Text: "World"},
	}

	assert.True(t, CommentChangeAllowed(before, split, map[string]bool{"s1": true}))
	assert.False(t, CommentChangeAllowed(before, split, map[string]bool{}))
	assert.False(t, CommentChangeAllowed(split, before, map[string]bool{"s1": true}))
}

func TestDecideSuggestionSpans_AcceptAndReject(t *testing.T) {
	t.Parallel()

	spans := []automerge.Span{
		{Type: automerge.SpanTypeText, Text: "old", Marks: map[string]any{SuggestionDeleteMark: `{"id":"s1"}`}},
		{Type: automerge.SpanTypeText, Text: "new", Marks: map[string]any{SuggestionInsertMark: `{"id":"s1"}`}},
	}

	accepted, changed := decideSuggestionSpans(spans, "s1", true)
	require.True(t, changed)
	require.Len(t, accepted, 1)
	assert.Equal(t, "new", accepted[0].Text)
	assert.Empty(t, accepted[0].Marks)

	rejected, changed := decideSuggestionSpans(spans, "s1", false)
	require.True(t, changed)
	require.Len(t, rejected, 1)
	assert.Equal(t, "old", rejected[0].Text)
	assert.Empty(t, rejected[0].Marks)
}

func TestDecideSuggestionSpans_Splits(t *testing.T) {
	t.Parallel()

	insertion := []automerge.Span{
		paragraphBlock(),
		{Type: automerge.SpanTypeText, Text: "Hello"},
		splitBlock("s1", suggestionSplitInsertion),
		{Type: automerge.SpanTypeText, Text: "there", Marks: map[string]any{SuggestionInsertMark: `{"id":"s1"}`}},
	}

	rejected, changed := decideSuggestionSpans(insertion, "s1", false)
	require.True(t, changed)
	require.Len(t, rejected, 2)
	assert.Equal(t, automerge.SpanTypeBlock, rejected[0].Type)
	assert.Equal(t, "Hello", rejected[1].Text)

	structural := []automerge.Span{
		paragraphBlock(),
		{Type: automerge.SpanTypeText, Text: "Hello"},
		splitBlock("s1", suggestionSplitStructural),
		{Type: automerge.SpanTypeText, Text: "World"},
	}

	merged, changed := decideSuggestionSpans(structural, "s1", false)
	require.True(t, changed)
	require.Len(t, merged, 3)
	assert.Equal(t, "Hello", merged[1].Text)
	assert.Equal(t, "World", merged[2].Text)

	accepted, changed := decideSuggestionSpans(structural, "s1", true)
	require.True(t, changed)
	require.Len(t, accepted, 4)
	_, kind := suggestionSplit(accepted[2])
	assert.Empty(t, kind)
	assert.Equal(t, "World", accepted[3].Text)

	quote := []automerge.Span{
		paragraphBlock(),
		{Type: automerge.SpanTypeText, Text: "Hello", Marks: map[string]any{SuggestionInsertMark: `{"id":"s1"}`}},
		splitBlock("s1", suggestionSplitInsertion),
		{Type: automerge.SpanTypeText, Text: "there", Marks: map[string]any{SuggestionInsertMark: `{"id":"s1"}`}},
	}
	assert.Equal(t, "Hello\nthere", SuggestionQuote(quote, "s1"))
	assert.True(t, HasOpenSuggestions([]automerge.Span{splitBlock("s1", suggestionSplitStructural)}))

	projected := ProjectAcceptedSpans(structural)
	require.Len(t, projected, 3)
	assert.Equal(t, "Hello", projected[1].Text)
	assert.Equal(t, "World", projected[2].Text)
}

func paragraphBlock() automerge.Span {
	return automerge.Span{
		Type:  automerge.SpanTypeBlock,
		Block: map[string]any{"type": "paragraph"},
	}
}

func splitBlock(id string, kind string) automerge.Span {
	return automerge.Span{
		Type: automerge.SpanTypeBlock,
		Block: map[string]any{
			"type": "paragraph",
			"attrs": map[string]any{
				suggestionSplitAttr:     id,
				suggestionSplitKindAttr: kind,
			},
		},
	}
}

func TestQuoteCursors_ResolvesAFullParagraph(t *testing.T) {
	t.Parallel()

	actorID, err := automerge.NewActorID()
	require.NoError(t, err)

	document, err := automerge.New(actorID)
	require.NoError(t, err)

	defer func() { _ = document.Close() }()

	const body = "Keep this sentence."

	require.NoError(t, writeSpans(
		document,
		`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Keep this sentence."}]}]}`,
		true,
	))
	_, err = document.Commit("Set document content", time.Now())
	require.NoError(t, err)

	start, end, err := QuoteCursors(document, body, 0)
	require.NoError(t, err)
	require.NotEmpty(t, start)
	require.NotEmpty(t, end)

	text, err := document.Text("body")
	require.NoError(t, err)

	startPos, err := text.CursorPosition(start)
	require.NoError(t, err)
	endPos, err := text.CursorPosition(end)
	require.NoError(t, err)
	assert.Equal(t, uint32(0), startPos)
	assert.Equal(t, utf16Len(body), endPos)

	start, end, err = QuoteCursors(document, "this", 0)
	require.NoError(t, err)
	startPos, err = text.CursorPosition(start)
	require.NoError(t, err)
	endPos, err = text.CursorPosition(end)
	require.NoError(t, err)
	assert.Equal(t, utf16Len("Keep "), startPos)
	assert.Equal(t, utf16Len("Keep this"), endPos)
}

func TestQuoteCursors_ResolvesAcrossParagraphs(t *testing.T) {
	t.Parallel()

	actorID, err := automerge.NewActorID()
	require.NoError(t, err)

	document, err := automerge.New(actorID)
	require.NoError(t, err)

	defer func() { _ = document.Close() }()

	require.NoError(t, writeSpans(
		document,
		`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"First"}]},{"type":"paragraph","content":[{"type":"text","text":"Second"}]}]}`,
		true,
	))
	_, err = document.Commit("Set document content", time.Now())
	require.NoError(t, err)

	start, end, err := QuoteCursors(document, "First\nSecond", 0)
	require.NoError(t, err)

	text, err := document.Text("body")
	require.NoError(t, err)

	startPos, err := text.CursorPosition(start)
	require.NoError(t, err)
	endPos, err := text.CursorPosition(end)
	require.NoError(t, err)

	spans, err := BodySpans(document)
	require.NoError(t, err)

	quoted := textBetween(spans, startPos, endPos)
	assert.Equal(t, "FirstSecond", quoted)

	start, end, err = QuoteCursors(document, "st\nSe", 0)
	require.NoError(t, err)
	startPos, err = text.CursorPosition(start)
	require.NoError(t, err)
	endPos, err = text.CursorPosition(end)
	require.NoError(t, err)
	assert.Equal(t, "stSe", textBetween(spans, startPos, endPos))
}

func textBetween(spans []automerge.Span, start uint32, end uint32) string {
	var builder strings.Builder

	var absolute uint32
	for _, span := range spans {
		if span.Type != automerge.SpanTypeText {
			continue
		}

		for _, unit := range utf16.Encode([]rune(span.Text)) {
			if absolute >= start && absolute < end {
				builder.WriteString(string(utf16.Decode([]uint16{unit})))
			}

			absolute++
		}
	}

	return builder.String()
}

func TestMaterialize_DropsInsertionsAndOmitsSuggestionMarks(t *testing.T) {
	t.Parallel()

	actorID, err := automerge.NewActorID()
	require.NoError(t, err)

	document, err := automerge.New(actorID)
	require.NoError(t, err)

	defer func() { _ = document.Close() }()

	require.NoError(t, writeSpans(
		document,
		`{"type":"doc","content":[{"type":"paragraph","content":[{"type":"text","text":"Keep gone"}]}]}`,
		true,
	))

	spans, err := BodySpans(document)
	require.NoError(t, err)

	var block automerge.Span
	for _, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			block = span
			break
		}
	}

	require.Equal(t, automerge.SpanTypeBlock, block.Type)

	require.NoError(t, WriteBodySpans(document, []automerge.Span{
		block,
		{Type: automerge.SpanTypeText, Text: "Keep "},
		{
			Type: automerge.SpanTypeText,
			Text: "gone",
			Marks: map[string]any{
				SuggestionDeleteMark: `{"id":"s1"}`,
			},
		},
		{
			Type: automerge.SpanTypeText,
			Text: "inserted",
			Marks: map[string]any{
				SuggestionInsertMark: `{"id":"s1"}`,
			},
		},
	}))
	_, err = document.Commit("Suggest", time.Now())
	require.NoError(t, err)

	stored, err := BodySpans(document)
	require.NoError(t, err)
	require.True(t, HasOpenSuggestions(stored))

	content, err := materialize(context.Background(), document)
	require.NoError(t, err)
	assert.Contains(t, content, "Keep ")
	assert.Contains(t, content, "gone")
	assert.NotContains(t, content, "inserted")
	assert.False(t, strings.Contains(content, SuggestionInsertMark))
	assert.False(t, strings.Contains(content, SuggestionDeleteMark))
}

func TestReviewEphemeralFrame_UsesBareDocumentID(t *testing.T) {
	t.Parallel()

	documentID := gid.New(gid.NilTenant, coredata.DocumentEntityType)
	frame, err := reviewEphemeralFrame(documentID)
	require.NoError(t, err)

	message, err := collabpkg.DecodeMessage(frame)
	require.NoError(t, err)

	assert.Equal(t, collabpkg.DeriveDocumentID(documentID.String()), message.DocumentID)
	assert.NotContains(t, message.DocumentID, "automerge:")

	var payload map[string]string
	require.NoError(t, cbor.Unmarshal(message.Data, &payload))
	assert.Equal(t, "document-review", payload["kind"])
}

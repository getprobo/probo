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
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
)

const suggestionEditID = "suggestion-1"

// These cases follow replaceSelection in
// packages/ui/src/RichEditor/SuggestionExtension.ts. The editor tests cover
// the same replacements in SuggestionExtension.test.ts ("shows a replacement
// as a deletion plus an insertion", "keeps the exact characters that were
// typed back", "restores a selection after marking it deleted").
func TestApplySuggestionEdit_MatchesEditorReplacement(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		text        string
		replacement string
		want        []markedRun
	}{
		{
			name:        "replacement is a deletion plus an insertion",
			text:        "scope",
			replacement: "range",
			want: []markedRun{
				{text: "scop", mark: SuggestionDeleteMark},
				{text: "rang", mark: SuggestionInsertMark},
				{text: "e"},
			},
		},
		{
			name:        "shared suffix stays unmarked",
			text:        "Out of scope",
			replacement: "scope",
			want: []markedRun{
				{text: "Out of ", mark: SuggestionDeleteMark},
				{text: "scope"},
			},
		},
		{
			name:        "identical text is not a suggestion",
			text:        "hello",
			replacement: "hello",
		},
		{
			name:        "empty replacement deletes the quote",
			text:        "hello",
			replacement: "",
			want: []markedRun{
				{text: "hello", mark: SuggestionDeleteMark},
			},
		},
		{
			name:        "emoji suffix stays unmarked",
			text:        "scope 😀",
			replacement: "range 😀",
			want: []markedRun{
				{text: "scop", mark: SuggestionDeleteMark},
				{text: "rang", mark: SuggestionInsertMark},
				{text: "e 😀"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON(tc.text)+`]}`)
			err := ApplySuggestionEdit(document, suggestionEditID, tc.text, tc.replacement, 0)
			if tc.want == nil {
				require.ErrorIs(t, err, ErrSuggestionUnchanged)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, markedRuns(t, document))
		})
	}
}

func TestApplySuggestionEdit_AcceptKeepsOnlyTheReplacement(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("scope")+`]}`)

	require.NoError(t, ApplySuggestionEdit(document, suggestionEditID, "scope", "range", 0))

	changed, err := ApplySuggestionDecision(document, suggestionEditID, true)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, []markedRun{{text: "range"}}, markedRuns(t, document))

	rejected := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("scope")+`]}`)
	require.NoError(t, ApplySuggestionEdit(rejected, suggestionEditID, "scope", "range", 0))

	changed, err = ApplySuggestionDecision(rejected, suggestionEditID, false)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(t, []markedRun{{text: "scope"}}, markedRuns(t, rejected))
}

func TestQuoteRange_HugeOccurrenceDoesNotPanic(t *testing.T) {
	t.Parallel()

	_, _, err := quoteRange("hello", "h", math.MaxInt)
	require.ErrorIs(t, err, ErrQuoteNotFound)
}

func TestApplySuggestionEdit_AddsAnotherRedlineToTheSameSuggestion(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(
		t,
		`{"type":"doc","content":[`+paragraphJSON("Keep the first sentence. Keep the second sentence.")+`]}`,
	)

	require.NoError(t, ApplySuggestionEdit(document, suggestionEditID, "first", "opening", 0))
	require.NoError(t, ApplySuggestionEdit(document, suggestionEditID, "second", "closing", 0))

	assert.Equal(t, []markedRun{
		{text: "Keep the "},
		{text: "first", mark: SuggestionDeleteMark},
		{text: "opening", mark: SuggestionInsertMark},
		{text: " sentence. Keep the "},
		{text: "second", mark: SuggestionDeleteMark},
		{text: "closing", mark: SuggestionInsertMark},
		{text: " sentence."},
	}, markedRuns(t, document))

	err := ApplySuggestionEdit(document, suggestionEditID, "opening", "initial", 0)
	require.ErrorIs(t, err, ErrSuggestionOverlaps)

	changed, err := ApplySuggestionDecision(document, suggestionEditID, true)
	require.NoError(t, err)
	require.True(t, changed)
	assert.Equal(
		t,
		[]markedRun{{text: "Keep the opening sentence. Keep the closing sentence."}},
		markedRuns(t, document),
	)
}

func TestApplySuggestionEdit_LeavesTheRestOfTheDocumentUnmarked(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(
		t,
		`{"type":"doc","content":[`+
			paragraphJSON("Your rights")+","+
			paragraphJSON("Access the data")+
			`]}`,
	)

	require.NoError(t, ApplySuggestionEdit(document, suggestionEditID, "Access", "Reach", 0))

	assert.Equal(t, []markedRun{
		{text: "Your rights"},
		{text: "Access", mark: SuggestionDeleteMark},
		{text: "Reach", mark: SuggestionInsertMark},
		{text: " the data"},
	}, markedRuns(t, document))
}

func TestApplySuggestionEdit_UsesTheRequestedOccurrence(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, `{"type":"doc","content":[`+paragraphJSON("data and data")+`]}`)

	require.NoError(t, ApplySuggestionEdit(document, suggestionEditID, "data", "info", 1))

	assert.Equal(t, []markedRun{
		{text: "data and "},
		{text: "data", mark: SuggestionDeleteMark},
		{text: "info", mark: SuggestionInsertMark},
	}, markedRuns(t, document))
}

func TestApplySuggestionEdit_MarksAParagraphBreakInsideTheQuote(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(
		t,
		`{"type":"doc","content":[`+paragraphJSON("Hello")+","+paragraphJSON("World")+`]}`,
	)

	require.NoError(t, ApplySuggestionEdit(document, suggestionEditID, "Hello\nWorld", "Hi", 0))

	spans, err := BodySpans(document)
	require.NoError(t, err)

	splits := 0
	for _, span := range spans {
		id, kind := suggestionSplit(span)
		if id == "" {
			continue
		}

		splits++
		assert.Equal(t, suggestionEditID, id)
		assert.Equal(t, suggestionSplitDeletion, kind)
	}

	assert.Equal(t, 1, splits)
	assert.Equal(t, []markedRun{
		{text: "H"},
		{text: "elloWorld", mark: SuggestionDeleteMark},
		{text: "i", mark: SuggestionInsertMark},
	}, markedRuns(t, document))
}

func TestApplySuggestionEdit_DoesNotRewriteUnchangedTableCells(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(
		t,
		`{"type":"doc","content":[{"type":"table","content":[`+
			`{"type":"tableRow","content":[`+
			`{"type":"tableHeader","content":[`+paragraphJSON("Right")+`]},`+
			`{"type":"tableHeader","content":[`+paragraphJSON("Access")+`]}`+
			`]}]}]}`,
	)

	spans, err := BodySpans(document)
	require.NoError(t, err)

	for index := range spans {
		attrs, ok := spans[index].Block["attrs"].(map[string]any)
		if !ok {
			continue
		}

		if _, ok := attrs["colspan"]; ok {
			attrs["colwidth"] = nil
		}
	}

	require.NoError(t, WriteBodySpans(document, spans))
	_, err = document.Commit("browser table", time.Now())
	if err != nil && !errors.Is(err, automerge.ErrNoOperations) {
		require.NoError(t, err)
	}

	before, err := document.Heads()
	require.NoError(t, err)

	require.NoError(t, ApplySuggestionEdit(document, suggestionEditID, "Access", "Entry", 0))
	_, err = document.Commit("suggest", time.Now())
	require.NoError(t, err)

	after, err := document.Heads()
	require.NoError(t, err)

	patches, err := document.Diff(before, after)
	require.NoError(t, err)

	for _, patch := range patches {
		assert.NotEqual(t, automerge.PatchDeleteSeq, patch.Action)
	}

	assert.Equal(t, []markedRun{
		{text: "Right"},
		{text: "Access", mark: SuggestionDeleteMark},
		{text: "Entry", mark: SuggestionInsertMark},
	}, markedRuns(t, document))
}

func suggestionDocument(t *testing.T, content string) *automerge.Document {
	t.Helper()

	actorID, err := automerge.NewActorID()
	require.NoError(t, err)

	document, err := automerge.New(actorID)
	require.NoError(t, err)
	t.Cleanup(func() { _ = document.Close() })

	require.NoError(t, writeSpans(document, content, true))
	_, err = document.Commit("seed", time.Now())
	require.NoError(t, err)

	return document
}

func paragraphJSON(text string) string {
	encoded, err := json.Marshal(text)
	if err != nil {
		return `{"type":"paragraph"}`
	}

	return `{"type":"paragraph","content":[{"type":"text","text":` + string(encoded) + `}]}`
}

type markedRun struct {
	text string
	mark string
}

func markedRuns(t *testing.T, document *automerge.Document) []markedRun {
	t.Helper()

	spans, err := BodySpans(document)
	require.NoError(t, err)

	runs := make([]markedRun, 0)

	for _, span := range spans {
		if span.Type != automerge.SpanTypeText || span.Text == "" {
			continue
		}

		mark := ""
		if markID(span.Marks, SuggestionInsertMark) == suggestionEditID {
			mark = SuggestionInsertMark
		}

		if markID(span.Marks, SuggestionDeleteMark) == suggestionEditID {
			mark = SuggestionDeleteMark
		}

		if len(runs) > 0 && runs[len(runs)-1].mark == mark {
			runs[len(runs)-1].text += span.Text

			continue
		}

		runs = append(runs, markedRun{text: span.Text, mark: mark})
	}

	return runs
}

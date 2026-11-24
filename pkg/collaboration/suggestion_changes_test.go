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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
)

const changesSuggestionID = "suggestion-1"

func plainSpan(text string) automerge.Span {
	return automerge.Span{Type: automerge.SpanTypeText, Text: text}
}

func markedSpan(text string, mark string) automerge.Span {
	return automerge.Span{
		Type:  automerge.SpanTypeText,
		Text:  text,
		Marks: map[string]any{mark: suggestionMarkValue(changesSuggestionID)},
	}
}

func formatSpan(text string, suggestionMark string, action string) automerge.Span {
	return automerge.Span{
		Type:  automerge.SpanTypeText,
		Text:  text,
		Marks: map[string]any{suggestionMark: `{"id":"` + changesSuggestionID + `","action":"` + action + `"}`},
	}
}

func structureBlock(kind string, parents []string, splitKind string) automerge.Span {
	block := map[string]any{"type": kind, "parents": toAnyStrings(parents)}
	if splitKind != "" {
		block["attrs"] = map[string]any{
			suggestionSplitAttr:     changesSuggestionID,
			suggestionSplitKindAttr: splitKind,
		}
	}

	return automerge.Span{Type: automerge.SpanTypeBlock, Block: block}
}

func textChange(deleted string, inserted string) coredata.DocumentSuggestionChange {
	return coredata.DocumentSuggestionChange{
		Kind:     coredata.DocumentSuggestionChangeKindText,
		Deleted:  new(deleted),
		Inserted: new(inserted),
	}
}

func nodeChange(
	node coredata.DocumentSuggestionNodeKind,
	action coredata.DocumentSuggestionNodeAction,
) coredata.DocumentSuggestionChange {
	return coredata.DocumentSuggestionChange{
		Kind:       coredata.DocumentSuggestionChangeKindNode,
		Node:       node,
		NodeAction: action,
	}
}

func TestSuggestionChanges_Text(t *testing.T) {
	t.Parallel()

	t.Run(
		"adjacent delete and insert read as one replacement",
		func(t *testing.T) {
			t.Parallel()

			changes := SuggestionChanges(
				[]automerge.Span{
					paragraphBlock(),
					plainSpan("Keep "),
					markedSpan("old", SuggestionDeleteMark),
					markedSpan("new", SuggestionInsertMark),
				},
				changesSuggestionID,
			)

			require.NotNil(t, changes)
			assert.Equal(t, []coredata.DocumentSuggestionChange{textChange("old", "new")}, changes.Changes)
			assert.Equal(t, 1, changes.ChangeCount)
		},
	)

	t.Run(
		"separated edits stay separate changes",
		func(t *testing.T) {
			t.Parallel()

			changes := SuggestionChanges(
				[]automerge.Span{
					paragraphBlock(),
					markedSpan("a", SuggestionInsertMark),
					plainSpan(" plain "),
					markedSpan("b", SuggestionInsertMark),
				},
				changesSuggestionID,
			)

			require.NotNil(t, changes)
			assert.Equal(
				t,
				[]coredata.DocumentSuggestionChange{textChange("", "a"), textChange("", "b")},
				changes.Changes,
			)
			assert.Equal(t, 2, changes.ChangeCount)
		},
	)

	t.Run(
		"text across a line break joins into one change",
		func(t *testing.T) {
			t.Parallel()

			changes := SuggestionChanges(
				[]automerge.Span{
					paragraphBlock(),
					markedSpan("first", SuggestionInsertMark),
					paragraphBlock(),
					markedSpan("second", SuggestionInsertMark),
				},
				changesSuggestionID,
			)

			require.NotNil(t, changes)
			assert.Equal(t, []coredata.DocumentSuggestionChange{textChange("", "first\nsecond")}, changes.Changes)
		},
	)

	t.Run(
		"long text is cut with an ellipsis",
		func(t *testing.T) {
			t.Parallel()

			changes := SuggestionChanges(
				[]automerge.Span{
					paragraphBlock(),
					markedSpan(strings.Repeat("é", maxChangeTextLength+20), SuggestionInsertMark),
				},
				changesSuggestionID,
			)

			require.NotNil(t, changes)
			require.Len(t, changes.Changes, 1)
			require.NotNil(t, changes.Changes[0].Inserted)
			assert.Equal(t, strings.Repeat("é", maxChangeTextLength)+"…", *changes.Changes[0].Inserted)
		},
	)
}

func TestSuggestionChanges_Format(t *testing.T) {
	t.Parallel()

	t.Run(
		"formatting merges per style and action",
		func(t *testing.T) {
			t.Parallel()

			changes := SuggestionChanges(
				[]automerge.Span{
					paragraphBlock(),
					formatSpan("one", "suggestion-format-strong", formatActionAdd),
					plainSpan(" plain "),
					formatSpan("two", "suggestion-format-strong", formatActionAdd),
					formatSpan("three", "suggestion-format-em", formatActionRemove),
				},
				changesSuggestionID,
			)

			require.NotNil(t, changes)
			assert.Equal(
				t,
				[]coredata.DocumentSuggestionChange{
					{
						Kind:         coredata.DocumentSuggestionChangeKindFormat,
						Mark:         string(FormatSuggestionBold),
						FormatAction: string(FormatSuggestionAdd),
					},
					{
						Kind:         coredata.DocumentSuggestionChangeKindFormat,
						Mark:         string(FormatSuggestionItalic),
						FormatAction: string(FormatSuggestionRemove),
					},
				},
				changes.Changes,
			)
			assert.Equal(t, 2, changes.ChangeCount)
		},
	)

	t.Run(
		"link reads the new and previous address",
		func(t *testing.T) {
			t.Parallel()

			changes := SuggestionChanges(
				[]automerge.Span{
					paragraphBlock(),
					{
						Type: automerge.SpanTypeText,
						Text: "site",
						Marks: map[string]any{
							"link":                   linkJSON("https://new.example"),
							"suggestion-format-link": `{"id":"` + changesSuggestionID + `","action":"add","href":"https://old.example"}`,
						},
					},
				},
				changesSuggestionID,
			)

			require.NotNil(t, changes)
			assert.Equal(
				t,
				[]coredata.DocumentSuggestionChange{
					{
						Kind:         coredata.DocumentSuggestionChangeKindFormat,
						Mark:         string(FormatSuggestionLink),
						FormatAction: string(FormatSuggestionAdd),
						Href:         new("https://new.example"),
						PreviousHref: new("https://old.example"),
					},
				},
				changes.Changes,
			)
		},
	)
}

func TestSuggestionChanges_Block(t *testing.T) {
	t.Parallel()

	changes := SuggestionChanges(
		[]automerge.Span{
			{
				Type: automerge.SpanTypeBlock,
				Block: map[string]any{
					"type": blockTypeHeading,
					"attrs": map[string]any{
						"level":             int64(2),
						suggestionBlockAttr: `{"id":"` + changesSuggestionID + `","type":"paragraph","attrs":{}}`,
					},
				},
			},
			plainSpan("Title"),
		},
		changesSuggestionID,
	)

	require.NotNil(t, changes)
	assert.Equal(
		t,
		[]coredata.DocumentSuggestionChange{
			{
				Kind:   coredata.DocumentSuggestionChangeKindBlock,
				Before: &coredata.DocumentSuggestionBlockLabel{Kind: coredata.DocumentSuggestionBlockKindParagraph},
				After: &coredata.DocumentSuggestionBlockLabel{
					Kind:  coredata.DocumentSuggestionBlockKindHeading,
					Level: new(2),
				},
			},
		},
		changes.Changes,
	)
}

func TestSuggestionChanges_Structure(t *testing.T) {
	t.Parallel()

	table := []string{blockTypeTable}
	row := []string{blockTypeTable, blockTypeTableRow}
	cell := []string{blockTypeTable, blockTypeTableRow, blockTypeTableCell}

	t.Run(
		"deleted row reads once without the text inside it",
		func(t *testing.T) {
			t.Parallel()

			changes := SuggestionChanges(
				[]automerge.Span{
					structureBlock(blockTypeTable, nil, ""),
					structureBlock(blockTypeTableRow, table, suggestionSplitDeletion),
					structureBlock(blockTypeTableCell, row, suggestionSplitDeletion),
					structureBlock(blockTypeParagraph, cell, ""),
					markedSpan("cell", SuggestionDeleteMark),
				},
				changesSuggestionID,
			)

			require.NotNil(t, changes)
			assert.Equal(
				t,
				[]coredata.DocumentSuggestionChange{
					nodeChange(coredata.DocumentSuggestionNodeKindRow, coredata.DocumentSuggestionNodeActionDelete),
				},
				changes.Changes,
			)
			assert.Equal(t, 1, changes.ChangeCount)
		},
	)

	t.Run(
		"structurally split row reads only the text inside it",
		func(t *testing.T) {
			t.Parallel()

			changes := SuggestionChanges(
				[]automerge.Span{
					structureBlock(blockTypeTable, nil, ""),
					structureBlock(blockTypeTableRow, table, suggestionSplitStructural),
					structureBlock(blockTypeTableCell, row, ""),
					structureBlock(blockTypeParagraph, cell, ""),
					markedSpan("cell", SuggestionInsertMark),
				},
				changesSuggestionID,
			)

			require.NotNil(t, changes)
			assert.Equal(t, []coredata.DocumentSuggestionChange{textChange("", "cell")}, changes.Changes)
			assert.Equal(t, 1, changes.ChangeCount)
		},
	)

	t.Run(
		"inserted column cells collapse into one change",
		func(t *testing.T) {
			t.Parallel()

			changes := SuggestionChanges(
				[]automerge.Span{
					structureBlock(blockTypeTable, nil, ""),
					structureBlock(blockTypeTableRow, table, ""),
					structureBlock(blockTypeTableCell, row, suggestionSplitInsertion),
					structureBlock(blockTypeParagraph, cell, ""),
					structureBlock(blockTypeTableRow, table, ""),
					structureBlock(blockTypeTableCell, row, suggestionSplitInsertion),
					structureBlock(blockTypeParagraph, cell, ""),
				},
				changesSuggestionID,
			)

			require.NotNil(t, changes)
			assert.Equal(
				t,
				[]coredata.DocumentSuggestionChange{
					nodeChange(coredata.DocumentSuggestionNodeKindColumn, coredata.DocumentSuggestionNodeActionInsert),
				},
				changes.Changes,
			)
		},
	)
}

func TestSuggestionChanges_Line(t *testing.T) {
	t.Parallel()

	t.Run(
		"blank inserted line without a text change",
		func(t *testing.T) {
			t.Parallel()

			changes := SuggestionChanges(
				[]automerge.Span{
					paragraphBlock(),
					plainSpan("Hello"),
					splitBlock(changesSuggestionID, suggestionSplitInsertion),
				},
				changesSuggestionID,
			)

			require.NotNil(t, changes)
			assert.Equal(
				t,
				[]coredata.DocumentSuggestionChange{
					nodeChange(coredata.DocumentSuggestionNodeKindLine, coredata.DocumentSuggestionNodeActionInsert),
				},
				changes.Changes,
			)
		},
	)

	t.Run(
		"inserted line with text reads as the text",
		func(t *testing.T) {
			t.Parallel()

			changes := SuggestionChanges(
				[]automerge.Span{
					paragraphBlock(),
					plainSpan("Hello"),
					splitBlock(changesSuggestionID, suggestionSplitInsertion),
					markedSpan("World", SuggestionInsertMark),
				},
				changesSuggestionID,
			)

			require.NotNil(t, changes)
			assert.Equal(t, []coredata.DocumentSuggestionChange{textChange("", "World")}, changes.Changes)
		},
	)
}

func TestSuggestionChanges_Cap(t *testing.T) {
	t.Parallel()

	spans := []automerge.Span{paragraphBlock()}
	for index := range 5 {
		spans = append(spans, markedSpan("edit"+string(rune('0'+index)), SuggestionInsertMark), plainSpan(" "))
	}

	changes := SuggestionChanges(spans, changesSuggestionID)

	require.NotNil(t, changes)
	assert.Len(t, changes.Changes, maxSuggestionChanges)
	assert.Equal(t, 5, changes.ChangeCount)
}

func TestSuggestionChanges_UnknownSuggestion(t *testing.T) {
	t.Parallel()

	changes := SuggestionChanges(
		[]automerge.Span{
			paragraphBlock(),
			markedSpan("new", SuggestionInsertMark),
		},
		"other",
	)

	assert.Nil(t, changes)
}

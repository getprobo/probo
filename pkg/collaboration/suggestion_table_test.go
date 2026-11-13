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
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.probo.inc/probo/pkg/automerge"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
)

const tableSuggestionID = "suggestion-table"

func TestApplyTableSuggestion_InsertRow(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		change TableSuggestion
		open   [][]string
	}{
		{
			name:   "after a body row",
			change: TableSuggestion{Operation: TableSuggestionInsertRow, Quote: "Read", Cells: []string{"Write", "No"}},
			open:   [][]string{{"H:Right", "H:Access"}, {"Read", "Yes"}, {"Write", "No"}},
		},
		{
			name:   "below the header row",
			change: TableSuggestion{Operation: TableSuggestionInsertRow, Quote: "Access", Cells: []string{"Write", "No"}},
			open:   [][]string{{"H:Right", "H:Access"}, {"Write", "No"}, {"Read", "Yes"}},
		},
		{
			name:   "above the header row holds plain cells",
			change: TableSuggestion{Operation: TableSuggestionInsertRow, Quote: "Right", Before: true},
			open:   [][]string{{"", ""}, {"H:Right", "H:Access"}, {"Read", "Yes"}},
		},
		{
			name:   "a multi-line cell keeps its paragraphs",
			change: TableSuggestion{Operation: TableSuggestionInsertRow, Quote: "Yes", Cells: []string{"Write\nDelete", ""}},
			open:   [][]string{{"H:Right", "H:Access"}, {"Read", "Yes"}, {"Write\nDelete", ""}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			document := suggestionDocument(t, tableDocumentJSON())
			original := bodySpans(t, document)

			require.NoError(t, ApplyTableSuggestion(document, tableSuggestionID, tc.change))

			assert.Equal(t, tc.open, tableGrid(t, renderOpen(t, document)))
			assert.Equal(t, originalGrid(), tableGrid(t, materializeContent(t, document)))
			assert.True(t, CommentChangeAllowed(original, bodySpans(t, document), map[string]bool{tableSuggestionID: true}))

			accepted := decided(t, document, true)
			assert.Equal(t, tc.open, tableGrid(t, accepted))
			assert.NotContains(t, accepted, suggestionSplitAttr)

			rejected := suggestionDocument(t, tableDocumentJSON())
			require.NoError(t, ApplyTableSuggestion(rejected, tableSuggestionID, tc.change))
			assert.Equal(t, originalGrid(), tableGrid(t, decided(t, rejected, false)))
			assert.True(t, spansEqual(original, bodySpans(t, rejected)))
		})
	}
}

func TestApplyTableSuggestion_DeleteRow(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, tableDocumentJSON())
	original := bodySpans(t, document)
	change := TableSuggestion{Operation: TableSuggestionDeleteRow, Quote: "Yes"}

	require.NoError(t, ApplyTableSuggestion(document, tableSuggestionID, change))

	assert.Equal(t, originalGrid(), tableGrid(t, renderOpen(t, document)))
	assert.Equal(t, originalGrid(), tableGrid(t, materializeContent(t, document)))
	assert.True(t, CommentChangeAllowed(original, bodySpans(t, document), map[string]bool{tableSuggestionID: true}))
	assert.Equal(t, "Read\nYes", SuggestionQuote(bodySpans(t, document), tableSuggestionID))

	accepted := decided(t, document, true)
	assert.Equal(t, [][]string{{"H:Right", "H:Access"}}, tableGrid(t, accepted))
	assert.Contains(t, accepted, "Intro")
	assert.Contains(t, accepted, "Outro")

	rejected := suggestionDocument(t, tableDocumentJSON())
	require.NoError(t, ApplyTableSuggestion(rejected, tableSuggestionID, change))
	assert.Equal(t, originalGrid(), tableGrid(t, decided(t, rejected, false)))
	assert.True(t, spansEqual(original, bodySpans(t, rejected)))
}

func TestApplyTableSuggestion_InsertColumn(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		change TableSuggestion
		open   [][]string
	}{
		{
			name:   "between two columns",
			change: TableSuggestion{Operation: TableSuggestionInsertColumn, Quote: "Right", Cells: []string{"Owner", "Alice"}},
			open:   [][]string{{"H:Right", "H:Owner", "H:Access"}, {"Read", "Alice", "Yes"}},
		},
		{
			name:   "before the first column",
			change: TableSuggestion{Operation: TableSuggestionInsertColumn, Quote: "Read", Before: true},
			open:   [][]string{{"H:", "H:Right", "H:Access"}, {"", "Read", "Yes"}},
		},
		{
			name:   "after the last column",
			change: TableSuggestion{Operation: TableSuggestionInsertColumn, Quote: "Yes", Cells: []string{"Since", "2026"}},
			open:   [][]string{{"H:Right", "H:Access", "H:Since"}, {"Read", "Yes", "2026"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			document := suggestionDocument(t, tableDocumentJSON())
			original := bodySpans(t, document)

			require.NoError(t, ApplyTableSuggestion(document, tableSuggestionID, tc.change))

			assert.Equal(t, tc.open, tableGrid(t, renderOpen(t, document)))
			assert.Equal(t, originalGrid(), tableGrid(t, materializeContent(t, document)))
			assert.True(t, CommentChangeAllowed(original, bodySpans(t, document), map[string]bool{tableSuggestionID: true}))
			assert.Equal(t, tc.open, tableGrid(t, decided(t, document, true)))

			rejected := suggestionDocument(t, tableDocumentJSON())
			require.NoError(t, ApplyTableSuggestion(rejected, tableSuggestionID, tc.change))
			assert.Equal(t, originalGrid(), tableGrid(t, decided(t, rejected, false)))
			assert.True(t, spansEqual(original, bodySpans(t, rejected)))
		})
	}
}

func TestApplyTableSuggestion_DeleteColumn(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, tableDocumentJSON())
	original := bodySpans(t, document)
	change := TableSuggestion{Operation: TableSuggestionDeleteColumn, Quote: "Access"}

	require.NoError(t, ApplyTableSuggestion(document, tableSuggestionID, change))

	assert.Equal(t, originalGrid(), tableGrid(t, materializeContent(t, document)))
	assert.True(t, CommentChangeAllowed(original, bodySpans(t, document), map[string]bool{tableSuggestionID: true}))
	assert.Equal(t, [][]string{{"H:Right"}, {"Read"}}, tableGrid(t, decided(t, document, true)))

	rejected := suggestionDocument(t, tableDocumentJSON())
	require.NoError(t, ApplyTableSuggestion(rejected, tableSuggestionID, change))
	assert.Equal(t, originalGrid(), tableGrid(t, decided(t, rejected, false)))
	assert.True(t, spansEqual(original, bodySpans(t, rejected)))
}

func TestApplyTableSuggestion_InsertTable(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(
		t,
		`{"type":"doc","content":[`+paragraphJSON("Intro")+`,`+paragraphJSON("Outro")+`]}`,
	)
	original := bodySpans(t, document)
	change := TableSuggestion{
		Operation: TableSuggestionInsertTable,
		Quote:     "Intro",
		Rows:      [][]string{{"Right", "Access"}, {"Read", "Yes"}},
		HeaderRow: true,
	}

	require.NoError(t, ApplyTableSuggestion(document, tableSuggestionID, change))

	open := renderOpen(t, document)
	assert.Equal(t, originalGrid(), tableGrid(t, open))
	assert.Less(t, strings.Index(open, "Intro"), strings.Index(open, `"table"`))
	assert.Less(t, strings.Index(open, `"table"`), strings.Index(open, "Outro"))
	assert.NotContains(t, materializeContent(t, document), `"table"`)
	assert.True(t, CommentChangeAllowed(original, bodySpans(t, document), map[string]bool{tableSuggestionID: true}))

	_, _, quote, err := SuggestionCursors(document, tableSuggestionID)
	require.NoError(t, err)
	assert.Equal(t, "Right\nAccess\nRead\nYes", quote)

	assert.Equal(t, originalGrid(), tableGrid(t, decided(t, document, true)))

	rejected := suggestionDocument(
		t,
		`{"type":"doc","content":[`+paragraphJSON("Intro")+`,`+paragraphJSON("Outro")+`]}`,
	)
	require.NoError(t, ApplyTableSuggestion(rejected, tableSuggestionID, change))
	assert.NotContains(t, decided(t, rejected, false), `"table"`)
	assert.True(t, spansEqual(original, bodySpans(t, rejected)))
}

func TestApplyTableSuggestion_DeleteTable(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, tableDocumentJSON())
	original := bodySpans(t, document)
	change := TableSuggestion{Operation: TableSuggestionDeleteTable, Quote: "Read"}

	require.NoError(t, ApplyTableSuggestion(document, tableSuggestionID, change))

	assert.Equal(t, originalGrid(), tableGrid(t, materializeContent(t, document)))
	assert.True(t, CommentChangeAllowed(original, bodySpans(t, document), map[string]bool{tableSuggestionID: true}))

	accepted := decided(t, document, true)
	assert.NotContains(t, accepted, `"table"`)
	assert.Contains(t, accepted, "Intro")
	assert.Contains(t, accepted, "Outro")
	assert.NotContains(t, accepted, "Right")

	rejected := suggestionDocument(t, tableDocumentJSON())
	require.NoError(t, ApplyTableSuggestion(rejected, tableSuggestionID, change))
	assert.Equal(t, originalGrid(), tableGrid(t, decided(t, rejected, false)))
	assert.True(t, spansEqual(original, bodySpans(t, rejected)))
}

func TestApplyTableSuggestion_EmptyRowStillAnchors(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, tableDocumentJSON())

	require.NoError(t, ApplyTableSuggestion(
		document,
		tableSuggestionID,
		TableSuggestion{Operation: TableSuggestionInsertRow, Quote: "Yes"},
	))

	start, end, quote, err := SuggestionCursors(document, tableSuggestionID)
	require.NoError(t, err)
	assert.NotEmpty(t, start)
	assert.NotEmpty(t, end)
	assert.Empty(t, quote)
}

func TestApplyTableSuggestion_Errors(t *testing.T) {
	t.Parallel()

	singleRow := `{"type":"doc","content":[{"type":"table","content":[` +
		tableRowJSON("tableCell", "Only", "Row") + `]}]}`
	singleColumn := `{"type":"doc","content":[{"type":"table","content":[` +
		tableRowJSON("tableCell", "Top") + `,` + tableRowJSON("tableCell", "Bottom") + `]}]}`
	merged := `{"type":"doc","content":[{"type":"table","content":[` +
		`{"type":"tableRow","content":[{"type":"tableCell","attrs":{"colspan":2},"content":[` + paragraphJSON("Wide") + `]}]},` +
		tableRowJSON("tableCell", "Left", "Right") + `]}]}`

	cases := []struct {
		name    string
		content string
		change  TableSuggestion
		want    error
	}{
		{
			name:    "quote outside a table",
			content: tableDocumentJSON(),
			change:  TableSuggestion{Operation: TableSuggestionInsertRow, Quote: "Intro"},
			want:    ErrQuoteNotInTable,
		},
		{
			name:    "quote not found",
			content: tableDocumentJSON(),
			change:  TableSuggestion{Operation: TableSuggestionDeleteRow, Quote: "Missing"},
			want:    ErrQuoteNotFound,
		},
		{
			name:    "row cell count does not match",
			content: tableDocumentJSON(),
			change:  TableSuggestion{Operation: TableSuggestionInsertRow, Quote: "Read", Cells: []string{"Only one"}},
			want:    ErrTableCellCount,
		},
		{
			name:    "column cell count does not match",
			content: tableDocumentJSON(),
			change:  TableSuggestion{Operation: TableSuggestionInsertColumn, Quote: "Read", Cells: []string{"a", "b", "c"}},
			want:    ErrTableCellCount,
		},
		{
			name:    "last row",
			content: singleRow,
			change:  TableSuggestion{Operation: TableSuggestionDeleteRow, Quote: "Only"},
			want:    ErrTableLastRow,
		},
		{
			name:    "last column",
			content: singleColumn,
			change:  TableSuggestion{Operation: TableSuggestionDeleteColumn, Quote: "Top"},
			want:    ErrTableLastColumn,
		},
		{
			name:    "merged cells",
			content: merged,
			change:  TableSuggestion{Operation: TableSuggestionInsertColumn, Quote: "Left"},
			want:    ErrTableHasMergedCells,
		},
		{
			name:    "ragged new table",
			content: tableDocumentJSON(),
			change:  TableSuggestion{Operation: TableSuggestionInsertTable, Quote: "Intro", Rows: [][]string{{"a", "b"}, {"c"}}},
			want:    ErrTableShape,
		},
		{
			name:    "unknown operation",
			content: tableDocumentJSON(),
			change:  TableSuggestion{Operation: "MERGE_CELLS", Quote: "Read"},
			want:    ErrTableOperation,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			document := suggestionDocument(t, tc.content)
			original := bodySpans(t, document)

			require.ErrorIs(t, ApplyTableSuggestion(document, tableSuggestionID, tc.change), tc.want)
			assert.True(t, spansEqual(original, bodySpans(t, document)))
		})
	}
}

func TestApplyTableSuggestion_RefusesToSwallowAnotherSuggestion(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, tableDocumentJSON())

	require.NoError(t, ApplySuggestionEdit(document, "other", "Yes", "No", 0))
	require.ErrorIs(
		t,
		ApplyTableSuggestion(document, tableSuggestionID, TableSuggestion{Operation: TableSuggestionDeleteRow, Quote: "Read"}),
		ErrSuggestionOverlaps,
	)

	require.NoError(t, ApplyTableSuggestion(
		document,
		tableSuggestionID,
		TableSuggestion{Operation: TableSuggestionDeleteRow, Quote: "Right"},
	))
	require.ErrorIs(
		t,
		ApplyTableSuggestion(document, tableSuggestionID, TableSuggestion{Operation: TableSuggestionDeleteColumn, Quote: "Right"}),
		ErrSuggestionOverlaps,
	)
}

func TestApplyTableSuggestion_LastRowCountsPendingDeletions(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(t, tableDocumentJSON())

	require.NoError(t, ApplyTableSuggestion(
		document,
		tableSuggestionID,
		TableSuggestion{Operation: TableSuggestionDeleteRow, Quote: "Read"},
	))
	require.ErrorIs(
		t,
		ApplyTableSuggestion(document, "second", TableSuggestion{Operation: TableSuggestionDeleteRow, Quote: "Right"}),
		ErrTableLastRow,
	)
}

func TestApplySuggestionEdit_QuoteAcrossCellsKeepsTheTable(t *testing.T) {
	t.Parallel()

	// The quoted text writes one newline per block marker, so crossing into a
	// row passes the row and cell markers.
	cases := []struct {
		name  string
		quote string
	}{
		{name: "between two cells", quote: "Right\nAccess"},
		{name: "between two rows", quote: "Access\n\nRead"},
		{name: "into the table", quote: "Intro\n\n\nRight"},
		{name: "out of the table", quote: "Yes\nOutro"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			document := suggestionDocument(t, tableDocumentJSON())

			require.NoError(t, ApplySuggestionEdit(document, suggestionEditID, tc.quote, "", 0))

			for _, span := range bodySpans(t, document) {
				id, _ := suggestionSplit(span)
				assert.Empty(t, id, "a table boundary must not become a deleted line break")
			}

			accepted := decidedFor(t, document, suggestionEditID, true)
			grid := tableGrid(t, accepted)
			require.Len(t, grid, 2)
			assert.Len(t, grid[0], 2)
			assert.Len(t, grid[1], 2)
		})
	}
}

func TestApplySuggestionEdit_JoinsParagraphsInsideOneCell(t *testing.T) {
	t.Parallel()

	document := suggestionDocument(
		t,
		`{"type":"doc","content":[{"type":"table","content":[`+
			`{"type":"tableRow","content":[`+
			`{"type":"tableCell","content":[`+paragraphJSON("First")+`,`+paragraphJSON("Second")+`]},`+
			`{"type":"tableCell","content":[`+paragraphJSON("Other")+`]}`+
			`]}]}]}`,
	)

	require.NoError(t, ApplySuggestionEdit(document, suggestionEditID, "First\nSecond", "First Second", 0))

	assert.Equal(t, [][]string{{"First Second", "Other"}}, tableGrid(t, decidedFor(t, document, suggestionEditID, true)))
}

func TestDecideSuggestionSpans_RemovesTheWholeTableSubtree(t *testing.T) {
	t.Parallel()

	row := automerge.Span{
		Type: automerge.SpanTypeBlock,
		Block: map[string]any{
			"type":    blockTypeTableRow,
			"parents": []any{blockTypeTable},
			"attrs": map[string]any{
				suggestionSplitAttr:     "s1",
				suggestionSplitKindAttr: suggestionSplitInsertion,
			},
		},
	}
	cell := automerge.Span{
		Type: automerge.SpanTypeBlock,
		Block: map[string]any{
			"type":    blockTypeTableCell,
			"parents": []any{blockTypeTable, blockTypeTableRow},
		},
	}
	spans := []automerge.Span{
		{Type: automerge.SpanTypeBlock, Block: map[string]any{"type": blockTypeTable, "parents": []any{}}},
		{Type: automerge.SpanTypeBlock, Block: map[string]any{"type": blockTypeTableRow, "parents": []any{blockTypeTable}}},
		cell,
		{Type: automerge.SpanTypeText, Text: "kept"},
		row,
		cell,
		{Type: automerge.SpanTypeText, Text: "unmarked child"},
		{Type: automerge.SpanTypeBlock, Block: map[string]any{"type": blockTypeParagraph, "parents": []any{}}},
		{Type: automerge.SpanTypeText, Text: "after"},
	}

	rejected, changed := decideSuggestionSpans(spans, "s1", false)
	require.True(t, changed)
	assert.Equal(t, []string{"kept", "after"}, spanTexts(rejected))
	assert.Len(t, rejected, 6)

	projected := ProjectAcceptedSpans(spans)
	assert.Equal(t, []string{"kept", "after"}, spanTexts(projected))
}

func tableDocumentJSON() string {
	return `{"type":"doc","content":[` + paragraphJSON("Intro") + `,{"type":"table","content":[` +
		tableRowJSON("tableHeader", "Right", "Access") + `,` +
		tableRowJSON("tableCell", "Read", "Yes") +
		`]},` + paragraphJSON("Outro") + `]}`
}

func originalGrid() [][]string {
	return [][]string{{"H:Right", "H:Access"}, {"Read", "Yes"}}
}

func tableRowJSON(cellType string, texts ...string) string {
	cells := make([]string, 0, len(texts))
	for _, text := range texts {
		cells = append(cells, `{"type":"`+cellType+`","content":[`+paragraphJSON(text)+`]}`)
	}

	return `{"type":"tableRow","content":[` + strings.Join(cells, ",") + `]}`
}

func bodySpans(t *testing.T, document *automerge.Document) []automerge.Span {
	t.Helper()

	spans, err := BodySpans(document)
	require.NoError(t, err)

	return spans
}

func renderOpen(t *testing.T, document *automerge.Document) string {
	t.Helper()

	content, err := automergeprosemirror.Render(bodySpans(t, document))
	require.NoError(t, err)

	return content
}

func materializeContent(t *testing.T, document *automerge.Document) string {
	t.Helper()

	content, err := materialize(context.Background(), document)
	require.NoError(t, err)

	return content
}

func decided(t *testing.T, document *automerge.Document, accept bool) string {
	t.Helper()

	return decidedFor(t, document, tableSuggestionID, accept)
}

func decidedFor(t *testing.T, document *automerge.Document, suggestionID string, accept bool) string {
	t.Helper()

	changed, err := ApplySuggestionDecision(document, suggestionID, accept)
	require.NoError(t, err)
	require.True(t, changed)
	require.False(t, HasOpenSuggestions(bodySpans(t, document)))

	_, err = document.Commit("decide", time.Now())
	if err != nil {
		require.ErrorIs(t, err, automerge.ErrNoOperations)
	}

	return renderOpen(t, document)
}

type tableTestNode struct {
	Type    string          `json:"type"`
	Text    string          `json:"text"`
	Content []tableTestNode `json:"content"`
}

// tableGrid lists the first table's cells row by row. Header cells carry an
// "H:" prefix and a cell's paragraphs are joined with a newline.
func tableGrid(t *testing.T, content string) [][]string {
	t.Helper()

	var root tableTestNode
	require.NoError(t, json.Unmarshal([]byte(content), &root))

	table := findTableNode(root)
	if table == nil {
		return nil
	}

	grid := make([][]string, 0, len(table.Content))
	for _, row := range table.Content {
		cells := make([]string, 0, len(row.Content))
		for _, cell := range row.Content {
			paragraphs := make([]string, 0, len(cell.Content))
			for _, paragraph := range cell.Content {
				paragraphs = append(paragraphs, nodeText(paragraph))
			}

			text := strings.Join(paragraphs, "\n")
			if cell.Type == "tableHeader" {
				text = "H:" + text
			}

			cells = append(cells, text)
		}

		grid = append(grid, cells)
	}

	return grid
}

func findTableNode(node tableTestNode) *tableTestNode {
	if node.Type == "table" {
		return &node
	}

	for _, child := range node.Content {
		if found := findTableNode(child); found != nil {
			return found
		}
	}

	return nil
}

func nodeText(node tableTestNode) string {
	var text strings.Builder

	text.WriteString(node.Text)

	for _, child := range node.Content {
		text.WriteString(nodeText(child))
	}

	return text.String()
}

func spanTexts(spans []automerge.Span) []string {
	texts := make([]string, 0)
	for _, span := range spans {
		if span.Type == automerge.SpanTypeText {
			texts = append(texts, span.Text)
		}
	}

	return texts
}

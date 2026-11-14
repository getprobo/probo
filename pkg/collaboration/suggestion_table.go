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
	"fmt"
	"strings"

	"go.probo.inc/probo/pkg/automerge"
)

const (
	blockTypeParagraph   = "paragraph"
	blockTypeTable       = "table"
	blockTypeTableRow    = "table-row"
	blockTypeTableCell   = "table-cell"
	blockTypeTableHeader = "table-header"
)

var (
	ErrQuoteNotInTable     = errors.New("quoted text is not inside a table")
	ErrTableHasMergedCells = errors.New("table has merged cells")
	ErrTableCellCount      = errors.New("cell count does not match the table")
	ErrTableShape          = errors.New("table rows must have the same number of cells")
	ErrTableLastRow        = errors.New("cannot delete the last row of a table")
	ErrTableLastColumn     = errors.New("cannot delete the last column of a table")
	ErrTableOperation      = errors.New("unknown table operation")
)

type TableSuggestionOperation string

const (
	TableSuggestionInsertRow    TableSuggestionOperation = "INSERT_ROW"
	TableSuggestionDeleteRow    TableSuggestionOperation = "DELETE_ROW"
	TableSuggestionInsertColumn TableSuggestionOperation = "INSERT_COLUMN"
	TableSuggestionDeleteColumn TableSuggestionOperation = "DELETE_COLUMN"
	TableSuggestionInsertTable  TableSuggestionOperation = "INSERT_TABLE"
	TableSuggestionDeleteTable  TableSuggestionOperation = "DELETE_TABLE"
)

// TableSuggestion is one structural redline. Quote locates the target: a cell
// of the row, column, or table to change, or for INSERT_TABLE any text of the
// top-level block the table goes beside. Cells fills a new row (left to right)
// or a new column (top to bottom). Rows fills a new table.
type TableSuggestion struct {
	Operation  TableSuggestionOperation
	Quote      string
	Occurrence int
	Before     bool
	Cells      []string
	Rows       [][]string
	HeaderRow  bool
}

type tableLayout struct {
	index int
	end   int
	rows  []tableRowLayout
}

type tableRowLayout struct {
	index int
	end   int
	cells []tableCellLayout
}

type tableCellLayout struct {
	index int
	end   int
}

// ApplyTableSuggestion writes a structural redline for suggestionID. New
// blocks carry an insertion split and their text an insert mark; removed
// blocks carry a deletion split and their text a delete mark. Accept and
// reject then treat the table, row, or cell as one subtree.
func ApplyTableSuggestion(
	document *automerge.Document,
	suggestionID string,
	change TableSuggestion,
) error {
	if suggestionID == "" {
		return ErrSuggestionNotInDocument
	}

	spans, err := BodySpans(document)
	if err != nil {
		return err
	}

	next, err := tableSuggestionSpans(spans, suggestionID, change)
	if err != nil {
		return err
	}

	if err := WriteBodySpans(document, next); err != nil {
		return fmt.Errorf("cannot write table suggestion: %w", err)
	}

	return nil
}

func tableSuggestionSpans(
	spans []automerge.Span,
	suggestionID string,
	change TableSuggestion,
) ([]automerge.Span, error) {
	anchor, err := quoteBlock(spans, change.Quote, change.Occurrence)
	if err != nil {
		return nil, err
	}

	if change.Operation == TableSuggestionInsertTable {
		return insertTable(spans, anchor, suggestionID, change)
	}

	tableIndex, ok := enclosingTable(spans, anchor)
	if !ok {
		return nil, ErrQuoteNotInTable
	}

	layout := parseTable(spans, tableIndex)

	rowIndex, cellIndex, ok := layout.locate(anchor)
	if !ok {
		return nil, ErrQuoteNotInTable
	}

	switch change.Operation {
	case TableSuggestionDeleteTable:
		return markSubtreeDeleted(spans, layout.index, layout.end, suggestionID)
	case TableSuggestionInsertRow:
		return insertRow(spans, layout, rowIndex, suggestionID, change)
	case TableSuggestionDeleteRow:
		return deleteRow(spans, layout, rowIndex, suggestionID)
	case TableSuggestionInsertColumn:
		return insertColumn(spans, layout, cellIndex, suggestionID, change)
	case TableSuggestionDeleteColumn:
		return deleteColumn(spans, layout, cellIndex, suggestionID)
	default:
		return nil, ErrTableOperation
	}
}

func insertRow(
	spans []automerge.Span,
	layout tableLayout,
	rowIndex int,
	suggestionID string,
	change TableSuggestion,
) ([]automerge.Span, error) {
	if layout.hasMergedCells(spans) {
		return nil, ErrTableHasMergedCells
	}

	width, ok := layout.width()
	if !ok {
		return nil, ErrTableShape
	}

	if len(change.Cells) > 0 && len(change.Cells) != width {
		return nil, ErrTableCellCount
	}

	at := rowIndex + 1
	if change.Before {
		at = rowIndex
	}

	rowPath := append(blockPath(spans[layout.index]), blockTypeTableRow)
	types := layout.newRowCellTypes(spans, at)

	inserted := []automerge.Span{insertionBlock(blockTypeTableRow, blockPath(spans[layout.index]), nil, suggestionID)}

	for column := range width {
		text := ""
		if len(change.Cells) > 0 {
			text = change.Cells[column]
		}

		inserted = append(inserted, insertedCell(types[column], rowPath, text, suggestionID)...)
	}

	position := layout.end
	if at < len(layout.rows) {
		position = layout.rows[at].index
	}

	return spliceSpans(spans, position, inserted), nil
}

func deleteRow(
	spans []automerge.Span,
	layout tableLayout,
	rowIndex int,
	suggestionID string,
) ([]automerge.Span, error) {
	if layout.hasMergedCells(spans) {
		return nil, ErrTableHasMergedCells
	}

	remaining := 0

	for index, row := range layout.rows {
		if index == rowIndex {
			continue
		}

		if _, kind := suggestionSplit(spans[row.index]); kind != suggestionSplitDeletion {
			remaining++
		}
	}

	if remaining == 0 {
		return nil, ErrTableLastRow
	}

	row := layout.rows[rowIndex]

	return markSubtreeDeleted(spans, row.index, row.end, suggestionID)
}

func insertColumn(
	spans []automerge.Span,
	layout tableLayout,
	cellIndex int,
	suggestionID string,
	change TableSuggestion,
) ([]automerge.Span, error) {
	if layout.hasMergedCells(spans) {
		return nil, ErrTableHasMergedCells
	}

	if _, ok := layout.width(); !ok {
		return nil, ErrTableShape
	}

	if len(change.Cells) > 0 && len(change.Cells) != len(layout.rows) {
		return nil, ErrTableCellCount
	}

	at := cellIndex + 1
	if change.Before {
		at = cellIndex
	}

	types := layout.newColumnCellTypes(spans, at)
	insertions := make(map[int][]automerge.Span, len(layout.rows))

	for rowIndex, row := range layout.rows {
		text := ""
		if len(change.Cells) > 0 {
			text = change.Cells[rowIndex]
		}

		position := row.end
		if at < len(row.cells) {
			position = row.cells[at].index
		}

		insertions[position] = insertedCell(
			types[rowIndex],
			blockPath(spans[row.index]),
			text,
			suggestionID,
		)
	}

	next := make([]automerge.Span, 0, len(spans)+len(insertions)*2)
	for index, span := range spans {
		next = append(next, insertions[index]...)
		next = append(next, span)
	}

	return append(next, insertions[len(spans)]...), nil
}

func deleteColumn(
	spans []automerge.Span,
	layout tableLayout,
	cellIndex int,
	suggestionID string,
) ([]automerge.Span, error) {
	if layout.hasMergedCells(spans) {
		return nil, ErrTableHasMergedCells
	}

	width, ok := layout.width()
	if !ok {
		return nil, ErrTableShape
	}

	remaining := 0

	for column := range width {
		if column == cellIndex {
			continue
		}

		if !layout.columnDeleted(spans, column) {
			remaining++
		}
	}

	if remaining == 0 {
		return nil, ErrTableLastColumn
	}

	next := spans

	for _, row := range layout.rows {
		cell := row.cells[cellIndex]

		marked, err := markSubtreeDeleted(next, cell.index, cell.end, suggestionID)
		if err != nil {
			return nil, err
		}

		next = marked
	}

	return next, nil
}

func insertTable(
	spans []automerge.Span,
	anchor int,
	suggestionID string,
	change TableSuggestion,
) ([]automerge.Span, error) {
	if len(change.Rows) == 0 || len(change.Rows[0]) == 0 {
		return nil, ErrTableShape
	}

	for _, row := range change.Rows {
		if len(row) != len(change.Rows[0]) {
			return nil, ErrTableShape
		}
	}

	root := topLevelBlock(spans, anchor)

	position := subtreeEnd(spans, root)
	if change.Before {
		position = root
	}

	parents := blockParents(spans[root])
	tablePath := append(append([]string(nil), parents...), blockTypeTable)
	rowPath := append(append([]string(nil), tablePath...), blockTypeTableRow)

	inserted := []automerge.Span{insertionBlock(blockTypeTable, parents, nil, suggestionID)}
	for rowIndex, row := range change.Rows {
		inserted = append(inserted, insertionBlock(blockTypeTableRow, tablePath, nil, suggestionID))

		cellType := blockTypeTableCell
		if rowIndex == 0 && change.HeaderRow {
			cellType = blockTypeTableHeader
		}

		for _, text := range row {
			inserted = append(inserted, insertedCell(cellType, rowPath, text, suggestionID)...)
		}
	}

	return spliceSpans(spans, position, inserted), nil
}

// quoteBlock returns the index of the block marker holding the quote's first
// character. The visible text matches suggestionUnits, so a quote and
// occurrence resolve to the same place for text and table redlines.
func quoteBlock(spans []automerge.Span, quote string, occurrence int) (int, error) {
	type segment struct {
		start int
		end   int
		block int
	}

	var visible strings.Builder

	segments := make([]segment, 0, len(spans))
	block := -1
	wrote := false

	for index, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			block = index

			if wrote {
				visible.WriteByte('\n')
			}

			continue
		}

		if span.Type != automerge.SpanTypeText || span.Text == "" {
			continue
		}

		start := visible.Len()
		visible.WriteString(span.Text)
		segments = append(segments, segment{start: start, end: visible.Len(), block: block})
		wrote = true
	}

	index, err := findQuote(visible.String(), quote, occurrence)
	if err != nil {
		return 0, err
	}

	for _, current := range segments {
		if index >= current.start && index < current.end {
			if current.block < 0 {
				return 0, ErrQuoteNotFound
			}

			return current.block, nil
		}
	}

	return 0, ErrQuoteNotFound
}

func enclosingTable(spans []automerge.Span, block int) (int, bool) {
	for index := block; index >= 0; index-- {
		if blockType(spans[index]) == blockTypeTable && block < subtreeEnd(spans, index) {
			return index, true
		}
	}

	return 0, false
}

func topLevelBlock(spans []automerge.Span, block int) int {
	for index := block; index >= 0; index-- {
		if spans[index].Type != automerge.SpanTypeBlock || len(blockParents(spans[index])) != 0 {
			continue
		}

		if block < subtreeEnd(spans, index) {
			return index
		}
	}

	return block
}

func parseTable(spans []automerge.Span, tableIndex int) tableLayout {
	layout := tableLayout{index: tableIndex, end: subtreeEnd(spans, tableIndex)}
	tablePath := blockPath(spans[tableIndex])

	for index := tableIndex + 1; index < layout.end; index++ {
		span := spans[index]
		if blockType(span) != blockTypeTableRow || !equalPath(blockParents(span), tablePath) {
			continue
		}

		row := tableRowLayout{index: index, end: subtreeEnd(spans, index)}
		rowPath := blockPath(span)

		for cell := index + 1; cell < row.end; cell++ {
			if !isTableCell(spans[cell]) || !equalPath(blockParents(spans[cell]), rowPath) {
				continue
			}

			row.cells = append(row.cells, tableCellLayout{index: cell, end: subtreeEnd(spans, cell)})
		}

		layout.rows = append(layout.rows, row)
		index = row.end - 1
	}

	return layout
}

func (t tableLayout) locate(block int) (int, int, bool) {
	for rowIndex, row := range t.rows {
		if block < row.index || block >= row.end {
			continue
		}

		for cellIndex, cell := range row.cells {
			if block >= cell.index && block < cell.end {
				return rowIndex, cellIndex, true
			}
		}

		return rowIndex, 0, false
	}

	return 0, 0, false
}

func (t tableLayout) width() (int, bool) {
	if len(t.rows) == 0 {
		return 0, false
	}

	width := len(t.rows[0].cells)
	for _, row := range t.rows {
		if len(row.cells) != width {
			return 0, false
		}
	}

	return width, width > 0
}

func (t tableLayout) hasMergedCells(spans []automerge.Span) bool {
	for _, row := range t.rows {
		for _, cell := range row.cells {
			attrs := blockAttrs(spans[cell.index])
			if intAttr(attrs, "colspan", 1) != 1 || intAttr(attrs, "rowspan", 1) != 1 {
				return true
			}
		}
	}

	return false
}

func (t tableLayout) columnDeleted(spans []automerge.Span, column int) bool {
	for _, row := range t.rows {
		if _, kind := suggestionSplit(spans[row.cells[column].index]); kind != suggestionSplitDeletion {
			return false
		}
	}

	return true
}

func (t tableLayout) rowIsHeader(spans []automerge.Span, rowIndex int) bool {
	for _, cell := range t.rows[rowIndex].cells {
		if blockType(spans[cell.index]) != blockTypeTableHeader {
			return false
		}
	}

	return true
}

func (t tableLayout) columnIsHeader(spans []automerge.Span, column int) bool {
	for _, row := range t.rows {
		if blockType(spans[row.cells[column].index]) != blockTypeTableHeader {
			return false
		}
	}

	return true
}

// newRowCellTypes follows prosemirror-tables addRow: copy the cell types of
// the row above (or below at the top), except beside a header row at the
// table edge, where the new row holds plain cells.
func (t tableLayout) newRowCellTypes(spans []automerge.Span, at int) []string {
	width, _ := t.width()
	types := make([]string, width)

	reference := 0
	if at > 0 {
		reference = at - 1
	}

	plain := false

	if t.rowIsHeader(spans, reference) {
		if at == 0 || at == len(t.rows) {
			plain = true
		} else {
			reference = at
		}
	}

	for column := range width {
		types[column] = blockTypeTableCell
		if !plain {
			types[column] = blockType(spans[t.rows[reference].cells[column].index])
		}
	}

	return types
}

// newColumnCellTypes follows prosemirror-tables addColumn, the column
// counterpart of newRowCellTypes.
func (t tableLayout) newColumnCellTypes(spans []automerge.Span, at int) []string {
	width, _ := t.width()
	types := make([]string, len(t.rows))

	reference := 0
	if at > 0 {
		reference = at - 1
	}

	plain := false

	if t.columnIsHeader(spans, reference) {
		if at == 0 || at == width {
			plain = true
		} else {
			reference = at
		}
	}

	for rowIndex, row := range t.rows {
		types[rowIndex] = blockTypeTableCell
		if !plain {
			types[rowIndex] = blockType(spans[row.cells[reference].index])
		}
	}

	return types
}

// insertedCell builds a cell whose first line is the cell's own inline text.
// Further lines become paragraphs inside the cell, as the editor stores them.
func insertedCell(cellType string, rowPath []string, text string, suggestionID string) []automerge.Span {
	attrs := map[string]any{"colspan": 1, "rowspan": 1}

	cellPath := append(append([]string(nil), rowPath...), cellType)

	spans := []automerge.Span{insertionBlock(cellType, rowPath, attrs, suggestionID)}

	for index, line := range strings.Split(text, "\n") {
		if index > 0 {
			spans = append(spans, insertionBlock(blockTypeParagraph, cellPath, nil, suggestionID))
		}

		if line != "" {
			spans = append(spans, insertedText(line, suggestionID))
		}
	}

	return spans
}

func insertionBlock(blockType string, parents []string, attrs map[string]any, suggestionID string) automerge.Span {
	nextAttrs := copyMap(attrs)
	nextAttrs[suggestionSplitAttr] = suggestionID
	nextAttrs[suggestionSplitKindAttr] = suggestionSplitInsertion

	return automerge.Span{
		Type: automerge.SpanTypeBlock,
		Block: map[string]any{
			"type":    blockType,
			"parents": toAnyStrings(parents),
			"attrs":   nextAttrs,
			"isEmbed": false,
		},
	}
}

func insertedText(text string, suggestionID string) automerge.Span {
	return automerge.Span{
		Type:  automerge.SpanTypeText,
		Text:  text,
		Marks: map[string]any{SuggestionInsertMark: suggestionMarkValue(suggestionID)},
	}
}

// markSubtreeDeleted redlines spans[from:to] for removal. A span that already
// belongs to a suggestion would be swallowed by this one, so it is refused.
func markSubtreeDeleted(
	spans []automerge.Span,
	from int,
	to int,
	suggestionID string,
) ([]automerge.Span, error) {
	next := append([]automerge.Span(nil), spans...)

	for index := from; index < to; index++ {
		span := next[index]

		switch span.Type {
		case automerge.SpanTypeBlock:
			if id, _ := suggestionSplit(span); id != "" {
				return nil, ErrSuggestionOverlaps
			}

			if _, suggested := readBlockSuggestion(span); suggested {
				return nil, ErrSuggestionOverlaps
			}

			next[index] = withDeletionSplit(span, suggestionID)
		case automerge.SpanTypeText:
			if markID(span.Marks, SuggestionInsertMark) != "" ||
				markID(span.Marks, SuggestionDeleteMark) != "" ||
				len(formatMarkIDs(span.Marks)) > 0 {
				return nil, ErrSuggestionOverlaps
			}

			marks := copyMap(span.Marks)
			marks[SuggestionDeleteMark] = suggestionMarkValue(suggestionID)
			span.Marks = marks
			next[index] = span
		}
	}

	return next, nil
}

func spliceSpans(spans []automerge.Span, position int, inserted []automerge.Span) []automerge.Span {
	next := make([]automerge.Span, 0, len(spans)+len(inserted))
	next = append(next, spans[:position]...)
	next = append(next, inserted...)

	return append(next, spans[position:]...)
}

func suggestionMarkValue(suggestionID string) string {
	encoded, _ := json.Marshal(map[string]string{"id": suggestionID})

	return string(encoded)
}

func isTableStructure(span automerge.Span) bool {
	switch blockType(span) {
	case blockTypeTable, blockTypeTableRow, blockTypeTableCell, blockTypeTableHeader:
		return true
	default:
		return false
	}
}

func isTableCell(span automerge.Span) bool {
	kind := blockType(span)

	return kind == blockTypeTableCell || kind == blockTypeTableHeader
}

// subtreeEnd returns the index just past the block at index and everything
// nested under it: the first later block whose parents leave its path.
func subtreeEnd(spans []automerge.Span, index int) int {
	if spans[index].Type != automerge.SpanTypeBlock {
		return index + 1
	}

	path := blockPath(spans[index])

	for next := index + 1; next < len(spans); next++ {
		if spans[next].Type != automerge.SpanTypeBlock {
			continue
		}

		if !hasPathPrefix(blockParents(spans[next]), path) {
			return next
		}
	}

	return len(spans)
}

func blockType(span automerge.Span) string {
	if span.Type != automerge.SpanTypeBlock || span.Block == nil {
		return ""
	}

	value, _ := span.Block["type"].(string)

	return value
}

func blockParents(span automerge.Span) []string {
	if span.Type != automerge.SpanTypeBlock || span.Block == nil {
		return nil
	}

	switch values := span.Block["parents"].(type) {
	case []string:
		return append([]string(nil), values...)
	case []any:
		parents := make([]string, 0, len(values))
		for _, value := range values {
			if name, ok := value.(string); ok {
				parents = append(parents, name)
			}
		}

		return parents
	default:
		return nil
	}
}

func blockPath(span automerge.Span) []string {
	return append(blockParents(span), blockType(span))
}

func blockAttrs(span automerge.Span) map[string]any {
	if span.Block == nil {
		return nil
	}

	attrs, _ := span.Block["attrs"].(map[string]any)

	return attrs
}

func intAttr(attrs map[string]any, name string, fallback int64) int64 {
	switch value := attrs[name].(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case uint64:
		return int64(value)
	case float64:
		return int64(value)
	default:
		return fallback
	}
}

func hasPathPrefix(values []string, prefix []string) bool {
	if len(values) < len(prefix) {
		return false
	}

	for index := range prefix {
		if values[index] != prefix[index] {
			return false
		}
	}

	return true
}

func equalPath(left []string, right []string) bool {
	return len(left) == len(right) && hasPathPrefix(left, right)
}

func toAnyStrings(values []string) []any {
	result := make([]any, len(values))
	for index, value := range values {
		result[index] = value
	}

	return result
}

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
	"strconv"
	"unicode/utf8"

	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
)

const (
	// The console lists this many changes of a suggestion and counts the rest.
	maxSuggestionChanges = 3
	// Inserted or deleted text past this length is cut, so a pasted page is
	// not stored or drawn on a card.
	maxChangeTextLength = 280
)

type (
	changeTextField struct {
		text      string
		parent    int
		truncated bool
	}

	openTextChange struct {
		index     int
		deleted   changeTextField
		inserted  changeTextField
		lastIndex int
	}

	suggestionChangeReader struct {
		changes      []coredata.DocumentSuggestionChange
		keys         map[string]bool
		openText     *openTextChange
		lineInserted bool
		lineDeleted  bool
	}
)

// SuggestionChanges reads what one suggestion changes as an ordered list of
// distinct changes, the way the console lists an open suggestion. It returns
// nil when the suggestion is not in spans.
func SuggestionChanges(spans []automerge.Span, suggestionID string) *coredata.DocumentSuggestionChanges {
	if suggestionID == "" {
		return nil
	}

	reader := suggestionChangeReader{
		changes: make([]coredata.DocumentSuggestionChange, 0),
		keys:    map[string]bool{},
	}

	var (
		found     bool
		textIndex int
	)

	// Inserted or deleted structure redlines everything inside it too. Those
	// inner marks belong to that one change, so they are skipped.
	coveredUntil := -1
	parent := -1

	for index, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			if !blockIsEmbed(span) {
				parent = index
			}

			if id, kind := suggestionSplit(span); id == suggestionID && isSuggestionSplitKind(kind) {
				found = true

				if index >= coveredUntil {
					switch {
					case isTextBlock(span):
						if kind == suggestionSplitDeletion {
							reader.lineDeleted = true
						} else {
							reader.lineInserted = true
						}
					case kind != suggestionSplitStructural:
						reader.addNode(suggestionNodeKind(span), suggestionNodeAction(kind))

						coveredUntil = subtreeEnd(spans, index)
					}
				}
			}

			if payload, ok := readBlockSuggestion(span); ok && payload.ID == suggestionID {
				found = true

				if index >= coveredUntil {
					reader.addBlock(
						suggestionBlockLabel(payload.Type, payload.Attrs),
						suggestionBlockLabel(blockType(span), blockAttrs(span)),
					)
				}
			}

			continue
		}

		if span.Type != automerge.SpanTypeText || span.Text == "" {
			continue
		}

		textIndex++

		deleted := markID(span.Marks, SuggestionDeleteMark) == suggestionID
		inserted := markID(span.Marks, SuggestionInsertMark) == suggestionID
		formatted := formatMarkMatches(span.Marks, suggestionID)

		if !deleted && !inserted && !formatted {
			continue
		}

		found = true

		if index < coveredUntil {
			continue
		}

		if deleted {
			reader.appendTextRun(true, span.Text, parent, textIndex)
		}

		if inserted {
			reader.appendTextRun(false, span.Text, parent, textIndex)
		}

		for _, spec := range formatStyles {
			payload, ok := readFormat(span.Marks, spec.suggestion)
			if !ok || payload.ID != suggestionID {
				continue
			}

			reader.addFormat(spec, payload, span.Marks)
		}
	}

	if !found {
		return nil
	}

	return reader.result()
}

func (r *suggestionChangeReader) result() *coredata.DocumentSuggestionChanges {
	changes := r.changes

	hasText := false

	for _, change := range changes {
		if change.Kind == coredata.DocumentSuggestionChangeKindText {
			hasText = true
			break
		}
	}

	if !hasText {
		if r.lineInserted {
			changes = append(changes, lineChange(coredata.DocumentSuggestionNodeActionInsert))
		}

		if r.lineDeleted {
			changes = append(changes, lineChange(coredata.DocumentSuggestionNodeActionDelete))
		}
	}

	count := len(changes)
	if count > maxSuggestionChanges {
		changes = changes[:maxSuggestionChanges]
	}

	return &coredata.DocumentSuggestionChanges{
		Changes:     changes,
		ChangeCount: count,
	}
}

func (r *suggestionChangeReader) addKeyed(key string, change coredata.DocumentSuggestionChange) {
	if r.keys[key] {
		return
	}

	r.keys[key] = true
	r.changes = append(r.changes, change)
}

func (r *suggestionChangeReader) addNode(
	node coredata.DocumentSuggestionNodeKind,
	action coredata.DocumentSuggestionNodeAction,
) {
	r.addKeyed(
		"node:"+node.String()+":"+action.String(),
		coredata.DocumentSuggestionChange{
			Kind:       coredata.DocumentSuggestionChangeKindNode,
			Node:       node,
			NodeAction: action,
		},
	)
}

func (r *suggestionChangeReader) addBlock(before coredata.DocumentSuggestionBlockLabel, after coredata.DocumentSuggestionBlockLabel) {
	r.addKeyed(
		"block:"+blockLabelKey(before)+":"+blockLabelKey(after),
		coredata.DocumentSuggestionChange{
			Kind:   coredata.DocumentSuggestionChangeKindBlock,
			Before: &before,
			After:  &after,
		},
	)
}

func (r *suggestionChangeReader) addFormat(spec formatStyle, payload formatMarkPayload, marks map[string]any) {
	action := FormatSuggestionAdd
	if payload.Action == formatActionRemove {
		action = FormatSuggestionRemove
	}

	change := coredata.DocumentSuggestionChange{
		Kind:         coredata.DocumentSuggestionChangeKindFormat,
		Mark:         string(spec.mark),
		FormatAction: string(action),
		PreviousHref: nonEmptyString(payload.Href),
	}

	if spec.link {
		change.Href = nonEmptyString(linkHrefOf(marks))
	}

	r.addKeyed("format:"+string(spec.mark)+":"+string(action), change)
}

// appendTextRun keeps delete and insert runs of the suggestion one change
// while they are consecutive text spans, so a replacement reads as old next
// to new. Any other text in between starts a new change.
func (r *suggestionChangeReader) appendTextRun(deleted bool, text string, parent int, textIndex int) {
	open := r.openText
	if open == nil || textIndex-open.lastIndex > 1 {
		r.changes = append(
			r.changes,
			coredata.DocumentSuggestionChange{
				Kind:     coredata.DocumentSuggestionChangeKindText,
				Deleted:  new(""),
				Inserted: new(""),
			},
		)

		open = &openTextChange{index: len(r.changes) - 1}
		r.openText = open
	}

	if deleted {
		open.deleted.append(text, parent)
		r.changes[open.index].Deleted = new(open.deleted.text)
	} else {
		open.inserted.append(text, parent)
		r.changes[open.index].Inserted = new(open.inserted.text)
	}

	open.lastIndex = textIndex
}

func (f *changeTextField) append(text string, parent int) {
	if f.truncated {
		return
	}

	joined := f.text + text
	if f.text != "" && f.parent != parent {
		joined = f.text + "\n" + text
	}

	f.parent = parent

	if utf8.RuneCountInString(joined) > maxChangeTextLength {
		f.text = string([]rune(joined)[:maxChangeTextLength]) + "…"
		f.truncated = true

		return
	}

	f.text = joined
}

func lineChange(action coredata.DocumentSuggestionNodeAction) coredata.DocumentSuggestionChange {
	return coredata.DocumentSuggestionChange{
		Kind:       coredata.DocumentSuggestionChangeKindNode,
		Node:       coredata.DocumentSuggestionNodeKindLine,
		NodeAction: action,
	}
}

func isSuggestionSplitKind(kind string) bool {
	switch kind {
	case suggestionSplitInsertion, suggestionSplitStructural, suggestionSplitDeletion:
		return true
	default:
		return false
	}
}

func isTextBlock(span automerge.Span) bool {
	switch blockType(span) {
	case blockTypeParagraph, blockTypeHeading, blockTypeCode:
		return true
	default:
		return false
	}
}

func suggestionNodeAction(splitKind string) coredata.DocumentSuggestionNodeAction {
	if splitKind == suggestionSplitDeletion {
		return coredata.DocumentSuggestionNodeActionDelete
	}

	return coredata.DocumentSuggestionNodeActionInsert
}

func suggestionNodeKind(span automerge.Span) coredata.DocumentSuggestionNodeKind {
	switch blockType(span) {
	case blockTypeTable:
		return coredata.DocumentSuggestionNodeKindTable
	case blockTypeTableRow:
		return coredata.DocumentSuggestionNodeKindRow
	case blockTypeTableCell, blockTypeTableHeader:
		return coredata.DocumentSuggestionNodeKindColumn
	case blockTypeImage:
		return coredata.DocumentSuggestionNodeKindImage
	case blockTypeHorizontalRule:
		return coredata.DocumentSuggestionNodeKindDivider
	case blockTypeFile:
		return coredata.DocumentSuggestionNodeKindFile
	default:
		return coredata.DocumentSuggestionNodeKindOther
	}
}

func suggestionBlockLabel(kind string, attrs map[string]any) coredata.DocumentSuggestionBlockLabel {
	switch kind {
	case blockTypeParagraph:
		return coredata.DocumentSuggestionBlockLabel{Kind: coredata.DocumentSuggestionBlockKindParagraph}
	case blockTypeHeading:
		return coredata.DocumentSuggestionBlockLabel{
			Kind:  coredata.DocumentSuggestionBlockKindHeading,
			Level: new(int(intAttr(attrs, "level", 1))),
		}
	case blockTypeBlockquote:
		return coredata.DocumentSuggestionBlockLabel{Kind: coredata.DocumentSuggestionBlockKindQuote}
	case blockTypeCode:
		return coredata.DocumentSuggestionBlockLabel{Kind: coredata.DocumentSuggestionBlockKindCode}
	case blockTypeUnorderedListItem:
		return coredata.DocumentSuggestionBlockLabel{Kind: coredata.DocumentSuggestionBlockKindBulletList}
	case blockTypeOrderedListItem:
		return coredata.DocumentSuggestionBlockLabel{Kind: coredata.DocumentSuggestionBlockKindOrderedList}
	case blockTypeTaskListItem:
		return coredata.DocumentSuggestionBlockLabel{Kind: coredata.DocumentSuggestionBlockKindTaskList}
	default:
		return coredata.DocumentSuggestionBlockLabel{Kind: coredata.DocumentSuggestionBlockKindOther}
	}
}

func blockLabelKey(label coredata.DocumentSuggestionBlockLabel) string {
	if label.Level == nil {
		return label.Kind.String()
	}

	return label.Kind.String() + strconv.Itoa(*label.Level)
}

func nonEmptyString(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

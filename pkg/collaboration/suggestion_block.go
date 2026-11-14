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
	"errors"
	"fmt"
	"strings"
	"unicode/utf16"

	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
	"go.probo.inc/probo/pkg/prosemirror"
)

const (
	blockTypeHardBreak         = "hard-break"
	blockTypeHeading           = "heading"
	blockTypeBlockquote        = "blockquote"
	blockTypeCode              = "code-block"
	blockTypeHorizontalRule    = "horizontal-rule"
	blockTypeImage             = "image"
	blockTypeFile              = "file"
	blockTypeMention           = "mention"
	blockTypeOrderedListItem   = "ordered-list-item"
	blockTypeTaskListItem      = "task-list-item"
	blockTypeUnorderedListItem = "unordered-list-item"

	suggestionBlockAttr = "suggestionBlock"
)

// BlockOperation is one structural change that is not a table edit.
type BlockOperation string

const (
	BlockSet           BlockOperation = "SET_BLOCK"
	BlockInsert        BlockOperation = "INSERT_BLOCK"
	BlockDelete        BlockOperation = "DELETE_BLOCK"
	BlockInsertItem    BlockOperation = "INSERT_LIST_ITEM"
	BlockDeleteItem    BlockOperation = "DELETE_LIST_ITEM"
	BlockNest          BlockOperation = "NEST_LIST_ITEM"
	BlockUnnest        BlockOperation = "UNNEST_LIST_ITEM"
	BlockSetTask       BlockOperation = "SET_TASK"
	BlockInsertDivider BlockOperation = "INSERT_DIVIDER"
	BlockInsertImage   BlockOperation = "INSERT_IMAGE"
	BlockInsertFile    BlockOperation = "INSERT_FILE"
	BlockInsertMention BlockOperation = "INSERT_MENTION"
	BlockDeleteMention BlockOperation = "DELETE_MENTION"
)

// BlockKind is the block a set or insert writes.
type BlockKind string

const (
	BlockKindParagraph   BlockKind = "PARAGRAPH"
	BlockKindHeading     BlockKind = "HEADING"
	BlockKindBlockquote  BlockKind = "BLOCKQUOTE"
	BlockKindBulletList  BlockKind = "BULLET_LIST"
	BlockKindOrderedList BlockKind = "ORDERED_LIST"
	BlockKindTaskList    BlockKind = "TASK_LIST"
	BlockKindCodeBlock   BlockKind = "CODE_BLOCK"
)

// BlockChange is one block edit. SuggestionID empty writes the change directly.
// Before nil means after the quote for an insert, and the quoted block for a
// delete. A set Before on a delete removes the neighboring textless block.
type BlockChange struct {
	Operation    BlockOperation
	Quote        string
	Occurrence   int
	Before       *bool
	Block        BlockKind
	Level        int
	Language     string
	Text         string
	Items        []string
	Checked      *bool
	FileID       string
	Src          string
	Alt          string
	Width        *int
	FileName     string
	MimeType     string
	Size         int64
	ProfileID    string
	Label        string
	SuggestionID string
}

type blockSuggestionPayload struct {
	ID      string         `json:"id"`
	Type    string         `json:"type"`
	Parents []string       `json:"parents,omitempty"`
	Attrs   map[string]any `json:"attrs,omitempty"`
}

var (
	ErrBlockOperation = errors.New("unknown block operation")
	ErrBlockTarget    = errors.New("quoted text is not in a block that supports this change")
	ErrLastListItem   = errors.New("cannot delete the last item of a list")
)

// ApplyBlock writes a block change. A suggestion id stores the previous block
// so accept keeps the new block and reject restores the old one. No id writes
// the block directly.
func ApplyBlock(document *automerge.Document, change BlockChange) error {
	spans, err := BodySpans(document)
	if err != nil {
		return err
	}

	next, err := blockChangeSpans(spans, change)
	if err != nil {
		return err
	}

	if err := WriteBodySpans(document, next); err != nil {
		return fmt.Errorf("cannot write block change: %w", err)
	}

	return nil
}

// PaintBlockSuggestion writes a suggested block change and commits it.
func (s *Service) PaintBlockSuggestion(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	organizationID gid.GID,
	identityID gid.GID,
	profileID *gid.GID,
	change BlockChange,
) (*automerge.Document, automerge.ActorID, error) {
	return s.paintCommit(
		ctx,
		scope,
		documentID,
		organizationID,
		identityID,
		profileID,
		"Suggest",
		func(document *automerge.Document) error {
			return ApplyBlock(document, change)
		},
	)
}

// PaintDirectBlock writes a block change with no suggestion and commits it.
func (s *Service) PaintDirectBlock(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	organizationID gid.GID,
	identityID gid.GID,
	profileID *gid.GID,
	change BlockChange,
) (*automerge.Document, automerge.ActorID, error) {
	change.SuggestionID = ""

	return s.paintCommit(
		ctx,
		scope,
		documentID,
		organizationID,
		identityID,
		profileID,
		"Edit",
		func(document *automerge.Document) error {
			return ApplyBlock(document, change)
		},
	)
}

func blockChangeSpans(spans []automerge.Span, change BlockChange) ([]automerge.Span, error) {
	if change.Occurrence < 0 {
		return nil, ErrQuoteNotFound
	}

	switch change.Operation {
	case BlockSet:
		return setBlock(spans, change)
	case BlockInsert:
		return insertBlock(spans, change)
	case BlockDelete:
		return deleteBlock(spans, change)
	case BlockInsertItem:
		return insertListItem(spans, change)
	case BlockDeleteItem:
		return deleteListItem(spans, change)
	case BlockNest:
		return nestListItem(spans, change)
	case BlockUnnest:
		return unnestListItem(spans, change)
	case BlockSetTask:
		return setTask(spans, change)
	case BlockInsertDivider:
		return insertLeaf(spans, change, blockTypeHorizontalRule, map[string]any{})
	case BlockInsertImage:
		return insertImage(spans, change)
	case BlockInsertFile:
		return insertFile(spans, change)
	case BlockInsertMention:
		return insertMention(spans, change)
	case BlockDeleteMention:
		return deleteMention(spans, change)
	default:
		return nil, ErrBlockOperation
	}
}

func setBlock(spans []automerge.Span, change BlockChange) ([]automerge.Span, error) {
	index, err := quoteBlock(spans, change.Quote, change.Occurrence)
	if err != nil {
		return nil, err
	}

	if !isConvertibleBlock(blockType(spans[index])) {
		return nil, ErrBlockTarget
	}

	if err := ensureBlockClear(spans, index); err != nil {
		return nil, err
	}

	kind, attrs, err := blockKindAttrs(change)
	if err != nil {
		return nil, err
	}

	parents := blockParents(spans[index])
	oldPath := blockPath(spans[index])

	newPath := append(append([]string(nil), parents...), kind)

	next := append([]automerge.Span(nil), spans...)
	next[index] = rewrittenBlock(spans[index], kind, parents, attrs, change.SuggestionID)

	if equalPath(oldPath, newPath) {
		return next, nil
	}

	// Nested blocks name their ancestors by type, so a changed container
	// renames the path of everything under it. Each one records its previous
	// path so rejecting the suggestion puts it back.
	end := subtreeEnd(spans, index)
	for cursor := index + 1; cursor < end; cursor++ {
		current := blockParents(spans[cursor])
		if spans[cursor].Type != automerge.SpanTypeBlock || !hasPathPrefix(current, oldPath) {
			continue
		}

		updated := append(append([]string(nil), newPath...), current[len(oldPath):]...)
		next[cursor] = rewrittenBlock(
			spans[cursor],
			blockType(spans[cursor]),
			updated,
			copyMap(blockAttrs(spans[cursor])),
			change.SuggestionID,
		)
	}

	return next, nil
}

func insertBlock(spans []automerge.Span, change BlockChange) ([]automerge.Span, error) {
	kind, attrs, err := blockKindAttrs(change)
	if err != nil {
		return nil, err
	}

	position, parents, err := insertPosition(spans, change)
	if err != nil {
		return nil, err
	}

	lines := change.Items
	if len(lines) == 0 {
		lines = []string{change.Text}
	}

	inserted := make([]automerge.Span, 0, len(lines)*2)

	for _, line := range lines {
		block := plainBlock(kind, parents, attrs, change.SuggestionID)
		inserted = append(inserted, block)

		if line != "" {
			inserted = append(inserted, plainText(line, change.SuggestionID))
		}
	}

	if change.Quote == "" {
		return inserted, nil
	}

	return spliceSpans(spans, position, inserted), nil
}

func deleteBlock(spans []automerge.Span, change BlockChange) ([]automerge.Span, error) {
	index, err := quoteBlock(spans, change.Quote, change.Occurrence)
	if err != nil {
		return nil, err
	}

	if change.Before != nil {
		index, err = neighborLeaf(spans, index, *change.Before)
		if err != nil {
			return nil, err
		}
	}

	if err := ensureBlockClear(spans, index); err != nil {
		return nil, err
	}

	end := subtreeEnd(spans, index)

	if change.SuggestionID == "" {
		next := append([]automerge.Span(nil), spans[:index]...)

		return append(next, spans[end:]...), nil
	}

	return markSubtreeDeleted(spans, index, end, change.SuggestionID)
}

func insertListItem(spans []automerge.Span, change BlockChange) ([]automerge.Span, error) {
	index, err := listItemBlock(spans, change.Quote, change.Occurrence)
	if err != nil {
		return nil, err
	}

	position := subtreeEnd(spans, index)
	if change.Before != nil && *change.Before {
		position = index
	}

	attrs := map[string]any{}
	if blockType(spans[index]) == blockTypeTaskListItem {
		attrs["checked"] = change.Checked != nil && *change.Checked
	}

	inserted := []automerge.Span{plainBlock(blockType(spans[index]), blockParents(spans[index]), attrs, change.SuggestionID)}
	if change.Text != "" {
		inserted = append(inserted, plainText(change.Text, change.SuggestionID))
	}

	return spliceSpans(spans, position, inserted), nil
}

func deleteListItem(spans []automerge.Span, change BlockChange) ([]automerge.Span, error) {
	index, err := listItemBlock(spans, change.Quote, change.Occurrence)
	if err != nil {
		return nil, err
	}

	if listSiblingCount(spans, index) < 2 {
		return nil, ErrLastListItem
	}

	if err := ensureBlockClear(spans, index); err != nil {
		return nil, err
	}

	end := subtreeEnd(spans, index)
	if change.SuggestionID == "" {
		next := append([]automerge.Span(nil), spans[:index]...)

		return append(next, spans[end:]...), nil
	}

	return markSubtreeDeleted(spans, index, end, change.SuggestionID)
}

func nestListItem(spans []automerge.Span, change BlockChange) ([]automerge.Span, error) {
	index, err := listItemBlock(spans, change.Quote, change.Occurrence)
	if err != nil {
		return nil, err
	}

	previous := previousSibling(spans, index)
	if previous < 0 || blockType(spans[previous]) != blockType(spans[index]) {
		return nil, ErrBlockTarget
	}

	if err := ensureBlockClear(spans, index); err != nil {
		return nil, err
	}

	return reparentBlock(spans, index, blockPath(spans[previous]), change.SuggestionID), nil
}

func unnestListItem(spans []automerge.Span, change BlockChange) ([]automerge.Span, error) {
	index, err := listItemBlock(spans, change.Quote, change.Occurrence)
	if err != nil {
		return nil, err
	}

	parents := blockParents(spans[index])
	if len(parents) == 0 || !isListItem(parents[len(parents)-1]) {
		return nil, ErrBlockTarget
	}

	if err := ensureBlockClear(spans, index); err != nil {
		return nil, err
	}

	return reparentBlock(spans, index, parents[:len(parents)-1], change.SuggestionID), nil
}

func setTask(spans []automerge.Span, change BlockChange) ([]automerge.Span, error) {
	index, err := quoteBlock(spans, change.Quote, change.Occurrence)
	if err != nil {
		return nil, err
	}

	if blockType(spans[index]) != blockTypeTaskListItem || change.Checked == nil {
		return nil, ErrBlockTarget
	}

	if err := ensureBlockClear(spans, index); err != nil {
		return nil, err
	}

	attrs := copyMap(blockAttrs(spans[index]))
	if attrs == nil {
		attrs = map[string]any{}
	}

	delete(attrs, suggestionBlockAttr)
	attrs["checked"] = *change.Checked

	next := append([]automerge.Span(nil), spans...)
	next[index] = rewrittenBlock(spans[index], blockTypeTaskListItem, blockParents(spans[index]), attrs, change.SuggestionID)

	return next, nil
}

func insertImage(spans []automerge.Span, change BlockChange) ([]automerge.Span, error) {
	attrs := map[string]any{}

	if change.FileID != "" {
		attrs["fileId"] = change.FileID
	}

	if change.Src != "" {
		src, err := persistedImageSource(change.Src)
		if err != nil {
			return nil, err
		}

		attrs["src"] = src
	}

	if change.FileID == "" && attrs["src"] == nil {
		return nil, ErrSuggestionHref
	}

	if change.Alt != "" {
		attrs["alt"] = change.Alt
	}

	if change.Width != nil {
		if *change.Width < 1 || *change.Width > 100 {
			return nil, ErrBlockTarget
		}

		attrs["width"] = *change.Width
	}

	return insertLeaf(spans, change, blockTypeImage, attrs)
}

func insertFile(spans []automerge.Span, change BlockChange) ([]automerge.Span, error) {
	if strings.TrimSpace(change.FileID) == "" {
		return nil, ErrBlockTarget
	}

	attrs := map[string]any{
		"fileId":   change.FileID,
		"fileName": change.FileName,
		"mimeType": change.MimeType,
		"size":     change.Size,
	}

	return insertLeaf(spans, change, blockTypeFile, attrs)
}

func insertLeaf(
	spans []automerge.Span,
	change BlockChange,
	kind string,
	attrs map[string]any,
) ([]automerge.Span, error) {
	position, parents, err := insertPosition(spans, change)
	if err != nil {
		return nil, err
	}

	leaf := plainBlock(kind, parents, attrs, change.SuggestionID)

	if change.Quote == "" {
		return []automerge.Span{leaf}, nil
	}

	return spliceSpans(spans, position, []automerge.Span{leaf}), nil
}

func insertMention(spans []automerge.Span, change BlockChange) ([]automerge.Span, error) {
	if strings.TrimSpace(change.ProfileID) == "" || strings.TrimSpace(change.Label) == "" {
		return nil, ErrBlockTarget
	}

	if change.Quote == "" {
		return nil, ErrQuoteNotFound
	}

	visible, units, _ := suggestionUnits(spans)

	start, end, err := quoteRange(visible, change.Quote, change.Occurrence)
	if err != nil {
		return nil, err
	}

	at := end
	if change.Before != nil && *change.Before {
		at = start
	}

	if suggestionRangeMarked(units, at, at+1) || (at < uint32(len(units)) && units[at].marked) {
		return nil, ErrSuggestionOverlaps
	}

	engine := uint32(0)
	if int(at) < len(units) {
		engine = units[at].engine
	} else if len(units) > 0 {
		engine = units[len(units)-1].engine + 1
	}

	if !engineTargetClear(spans, engine) {
		return nil, ErrSuggestionOverlaps
	}

	parents := parentsAtEngine(spans, engine)
	mention := plainBlock(blockTypeMention, parents, map[string]any{
		"id":    change.ProfileID,
		"label": change.Label,
	}, change.SuggestionID)
	mention.Block["isEmbed"] = true

	return insertAtEngine(spans, engine, mention), nil
}

// engineTargetClear reports whether an embed can go at engine without landing
// in another open suggestion: the text on either side carries no suggestion
// mark and the enclosing block carries no split or block suggestion.
func engineTargetClear(spans []automerge.Span, engine uint32) bool {
	cursor := uint32(0)
	block := -1

	for index, span := range spans {
		if cursor > engine {
			break
		}

		if span.Type == automerge.SpanTypeBlock {
			if cursor == engine {
				break
			}

			if !blockIsEmbed(span) {
				block = index
			}

			cursor++

			continue
		}

		if span.Type != automerge.SpanTypeText || span.Text == "" {
			continue
		}

		width := utf16Len(span.Text)
		touches := engine >= cursor && engine <= cursor+width
		marked := markID(span.Marks, SuggestionInsertMark) != "" ||
			markID(span.Marks, SuggestionDeleteMark) != "" ||
			len(formatMarkIDs(span.Marks)) > 0

		if touches && marked {
			return false
		}

		cursor += width
	}

	if block < 0 {
		return true
	}

	if id, _ := suggestionSplit(spans[block]); id != "" {
		return false
	}

	_, suggested := readBlockSuggestion(spans[block])

	return !suggested
}

func deleteMention(spans []automerge.Span, change BlockChange) ([]automerge.Span, error) {
	if change.Quote == "" || change.Occurrence < 0 {
		return nil, ErrQuoteNotFound
	}

	seen := 0

	for index, span := range spans {
		if blockType(span) != blockTypeMention {
			continue
		}

		label, _ := blockAttrs(span)["label"].(string)
		if label != change.Quote {
			continue
		}

		if seen != change.Occurrence {
			seen++

			continue
		}

		if err := ensureBlockClear(spans, index); err != nil {
			return nil, err
		}

		if change.SuggestionID == "" {
			next := append([]automerge.Span(nil), spans[:index]...)

			return append(next, spans[index+1:]...), nil
		}

		return markSubtreeDeleted(spans, index, index+1, change.SuggestionID)
	}

	return nil, ErrQuoteNotFound
}

func insertPosition(spans []automerge.Span, change BlockChange) (int, []string, error) {
	if change.Quote == "" {
		if documentText(spans) != "" {
			return 0, nil, ErrQuoteNotFound
		}

		return 0, nil, nil
	}

	index, err := quoteBlock(spans, change.Quote, change.Occurrence)
	if err != nil {
		return 0, nil, err
	}

	for blockType(spans[index]) != "" && len(blockParents(spans[index])) > 0 && !isListItem(blockType(spans[index])) {
		parent := parentIndex(spans, index)
		if parent < 0 {
			break
		}

		index = parent
	}

	parents := blockParents(spans[index])
	if change.Before != nil && *change.Before {
		return index, parents, nil
	}

	return subtreeEnd(spans, index), parents, nil
}

func neighborLeaf(spans []automerge.Span, index int, before bool) (int, error) {
	neighbor := -1
	if before {
		neighbor = previousSibling(spans, index)
	} else {
		next := subtreeEnd(spans, index)
		if next < len(spans) && spans[next].Type == automerge.SpanTypeBlock &&
			equalPath(blockParents(spans[next]), blockParents(spans[index])) {
			neighbor = next
		}
	}

	if neighbor < 0 || !isLeafBlock(blockType(spans[neighbor])) {
		return 0, ErrBlockTarget
	}

	return neighbor, nil
}

func listItemBlock(spans []automerge.Span, quote string, occurrence int) (int, error) {
	index, err := quoteBlock(spans, quote, occurrence)
	if err != nil {
		return 0, err
	}

	if !isListItem(blockType(spans[index])) {
		return 0, ErrBlockTarget
	}

	return index, nil
}

func listSiblingCount(spans []automerge.Span, index int) int {
	kind := blockType(spans[index])
	parents := blockParents(spans[index])
	count := 0

	for _, span := range spans {
		if span.Type != automerge.SpanTypeBlock || blockType(span) != kind {
			continue
		}

		if !equalPath(blockParents(span), parents) {
			continue
		}

		if _, splitKind := suggestionSplit(span); splitKind == suggestionSplitDeletion {
			continue
		}

		count++
	}

	return count
}

func previousSibling(spans []automerge.Span, index int) int {
	parents := blockParents(spans[index])
	previous := -1

	for cursor := 0; cursor < index; cursor++ {
		span := spans[cursor]
		if span.Type != automerge.SpanTypeBlock || !equalPath(blockParents(span), parents) {
			continue
		}

		previous = cursor
	}

	return previous
}

func parentIndex(spans []automerge.Span, index int) int {
	parents := blockParents(spans[index])
	if len(parents) == 0 {
		return -1
	}

	for cursor := index - 1; cursor >= 0; cursor-- {
		if spans[cursor].Type != automerge.SpanTypeBlock {
			continue
		}

		if equalPath(blockPath(spans[cursor]), parents) {
			return cursor
		}
	}

	return -1
}

func reparentBlock(
	spans []automerge.Span,
	index int,
	parents []string,
	suggestionID string,
) []automerge.Span {
	next := append([]automerge.Span(nil), spans...)
	oldPath := blockPath(spans[index])
	newPath := append(append([]string(nil), parents...), blockType(spans[index]))
	end := subtreeEnd(spans, index)

	for cursor := index; cursor < end; cursor++ {
		if next[cursor].Type != automerge.SpanTypeBlock {
			continue
		}

		current := blockParents(next[cursor])
		updated := parents

		if cursor != index && hasPathPrefix(current, oldPath) {
			updated = append(append([]string(nil), newPath...), current[len(oldPath):]...)
		}

		attrs := copyMap(blockAttrs(next[cursor]))
		if cursor == index {
			next[cursor] = rewrittenBlock(next[cursor], blockType(next[cursor]), updated, attrs, suggestionID)

			continue
		}

		next[cursor] = rewrittenBlock(next[cursor], blockType(next[cursor]), updated, attrs, suggestionID)
	}

	return next
}

func rewrittenBlock(
	span automerge.Span,
	kind string,
	parents []string,
	attrs map[string]any,
	suggestionID string,
) automerge.Span {
	if suggestionID != "" {
		attrs = withBlockSuggestion(span, attrs, suggestionID)
	}

	return plainBlock(kind, parents, attrs, "")
}

func withBlockSuggestion(span automerge.Span, attrs map[string]any, suggestionID string) map[string]any {
	previous := blockAttrs(span)
	if previous == nil {
		previous = map[string]any{}
	}

	previous = copyMap(previous)
	delete(previous, suggestionBlockAttr)
	delete(previous, suggestionSplitAttr)
	delete(previous, suggestionSplitKindAttr)

	payload := blockSuggestionPayload{
		ID:      suggestionID,
		Type:    blockType(span),
		Parents: blockParents(span),
		Attrs:   previous,
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return attrs
	}

	next := copyMap(attrs)
	if next == nil {
		next = map[string]any{}
	}

	next[suggestionBlockAttr] = string(encoded)

	return next
}

func plainBlock(kind string, parents []string, attrs map[string]any, suggestionID string) automerge.Span {
	if attrs == nil {
		attrs = map[string]any{}
	}

	attrs = copyMap(attrs)

	if suggestionID != "" {
		attrs[suggestionSplitAttr] = suggestionID
		attrs[suggestionSplitKindAttr] = suggestionSplitInsertion
	}

	block := map[string]any{
		"type":    kind,
		"parents": toAnyStrings(parents),
		"attrs":   attrs,
		"isEmbed": kind == blockTypeMention || kind == blockTypeHardBreak,
	}

	return automerge.Span{Type: automerge.SpanTypeBlock, Block: block}
}

func plainText(text string, suggestionID string) automerge.Span {
	if suggestionID == "" {
		return automerge.Span{Type: automerge.SpanTypeText, Text: text}
	}

	return insertedText(text, suggestionID)
}

func blockKindAttrs(change BlockChange) (string, map[string]any, error) {
	switch change.Block {
	case BlockKindParagraph:
		return blockTypeParagraph, map[string]any{}, nil
	case BlockKindHeading:
		if change.Level < 1 || change.Level > 6 {
			return "", nil, ErrBlockTarget
		}

		return blockTypeHeading, map[string]any{"level": change.Level}, nil
	case BlockKindBlockquote:
		return blockTypeBlockquote, map[string]any{}, nil
	case BlockKindBulletList:
		return blockTypeUnorderedListItem, map[string]any{}, nil
	case BlockKindOrderedList:
		return blockTypeOrderedListItem, map[string]any{}, nil
	case BlockKindTaskList:
		return blockTypeTaskListItem, map[string]any{"checked": false}, nil
	case BlockKindCodeBlock:
		attrs := map[string]any{}
		if change.Language != "" {
			attrs["language"] = change.Language
		}

		return blockTypeCode, attrs, nil
	default:
		return "", nil, ErrBlockOperation
	}
}

func ensureBlockClear(spans []automerge.Span, index int) error {
	end := subtreeEnd(spans, index)

	for cursor := index; cursor < end; cursor++ {
		span := spans[cursor]

		if id, _ := suggestionSplit(span); id != "" {
			return ErrSuggestionOverlaps
		}

		if _, ok := readBlockSuggestion(span); ok {
			return ErrSuggestionOverlaps
		}

		if span.Type != automerge.SpanTypeText {
			continue
		}

		if markID(span.Marks, SuggestionInsertMark) != "" || markID(span.Marks, SuggestionDeleteMark) != "" {
			return ErrSuggestionOverlaps
		}

		if len(formatMarkIDs(span.Marks)) > 0 {
			return ErrSuggestionOverlaps
		}
	}

	return nil
}

func isConvertibleBlock(kind string) bool {
	switch kind {
	case blockTypeParagraph, blockTypeHeading, blockTypeBlockquote, blockTypeCode,
		blockTypeUnorderedListItem, blockTypeOrderedListItem, blockTypeTaskListItem:
		return true
	default:
		return false
	}
}

func isListItem(kind string) bool {
	switch kind {
	case blockTypeUnorderedListItem, blockTypeOrderedListItem, blockTypeTaskListItem:
		return true
	default:
		return false
	}
}

func isLeafBlock(kind string) bool {
	switch kind {
	case blockTypeHorizontalRule, blockTypeImage, blockTypeFile:
		return true
	default:
		return false
	}
}

func documentText(spans []automerge.Span) string {
	visible, _, _ := suggestionUnits(spans)

	return strings.TrimSpace(visible)
}

func persistedImageSource(src string) (string, error) {
	safe := prosemirror.PersistedImageSrc(strings.TrimSpace(src))
	if safe == "" {
		return "", ErrSuggestionHref
	}

	return safe, nil
}

func parentsAtEngine(spans []automerge.Span, engine uint32) []string {
	cursor := uint32(0)
	parents := []string(nil)

	for _, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			if cursor >= engine {
				return parents
			}

			if !blockIsEmbed(span) {
				parents = blockPath(span)
			}

			cursor++

			continue
		}

		if span.Type == automerge.SpanTypeText {
			width := utf16Len(span.Text)
			if cursor+width >= engine {
				return parents
			}

			cursor += width
		}
	}

	return parents
}

func insertAtEngine(spans []automerge.Span, engine uint32, inserted automerge.Span) []automerge.Span {
	cursor := uint32(0)

	for index, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			if cursor == engine {
				return spliceSpans(spans, index, []automerge.Span{inserted})
			}

			cursor++

			continue
		}

		if span.Type != automerge.SpanTypeText || span.Text == "" {
			continue
		}

		width := utf16Len(span.Text)
		if cursor+width < engine {
			cursor += width

			continue
		}

		offset := int(engine - cursor)

		units := utf16.Encode([]rune(span.Text))
		if offset > len(units) {
			offset = len(units)
		}

		before := string(utf16.Decode(units[:offset]))
		after := string(utf16.Decode(units[offset:]))
		parts := make([]automerge.Span, 0, 3)

		if before != "" {
			parts = append(parts, textSpan(before, span.Marks))
		}

		parts = append(parts, inserted)

		if after != "" {
			parts = append(parts, textSpan(after, span.Marks))
		}

		next := append([]automerge.Span(nil), spans[:index]...)
		next = append(next, parts...)

		return append(next, spans[index+1:]...)
	}

	return append(append([]automerge.Span(nil), spans...), inserted)
}

func readBlockSuggestion(span automerge.Span) (blockSuggestionPayload, bool) {
	raw, _ := blockAttrs(span)[suggestionBlockAttr].(string)
	if raw == "" {
		return blockSuggestionPayload{}, false
	}

	var payload blockSuggestionPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || payload.ID == "" || payload.Type == "" {
		return blockSuggestionPayload{}, false
	}

	return payload, true
}

func clearBlockSuggestion(span automerge.Span) automerge.Span {
	attrs := copyMap(blockAttrs(span))
	if attrs == nil {
		return span
	}

	delete(attrs, suggestionBlockAttr)

	return rewriteAttrs(span, attrs)
}

func restoreBlockSuggestion(span automerge.Span) automerge.Span {
	payload, ok := readBlockSuggestion(span)
	if !ok {
		return span
	}

	attrs := payload.Attrs
	if attrs == nil {
		attrs = map[string]any{}
	}

	return plainBlock(payload.Type, payload.Parents, attrs, "")
}

func rewriteAttrs(span automerge.Span, attrs map[string]any) automerge.Span {
	block := copyMap(span.Block)
	block["attrs"] = attrs
	span.Block = block

	return span
}

func blockIsEmbed(span automerge.Span) bool {
	if span.Block == nil {
		return false
	}

	embed, _ := span.Block["isEmbed"].(bool)

	return embed
}

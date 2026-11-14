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
	"fmt"
	"strings"
	"time"
	"unicode/utf16"

	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// suggestionUnit is one UTF-16 code unit of the text a reviewer can quote.
// A newline between blocks points at that block marker. Engine is the index
// Splice and Mark use, which counts each block as one unit. Marked means the
// unit already belongs to a suggestion, so another redline must not cover it.
// Joinable means deleting the newline may merge the block into the previous
// one; a boundary into, out of, or between table cells is never joinable.
type suggestionUnit struct {
	engine   uint32
	block    bool
	marked   bool
	joinable bool
}

// PaintSuggestion marks quote on the live document and commits that change.
// The document stays open so the caller can persist it after storing the
// suggestion row. Close it on every path.
func (s *Service) PaintSuggestion(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	organizationID gid.GID,
	identityID gid.GID,
	profileID *gid.GID,
	suggestionID string,
	quote string,
	replacement string,
	occurrence int,
) (*automerge.Document, automerge.ActorID, error) {
	return s.paint(
		ctx,
		scope,
		documentID,
		organizationID,
		identityID,
		profileID,
		func(document *automerge.Document) error {
			return ApplySuggestionEdit(document, suggestionID, quote, replacement, occurrence)
		},
	)
}

// PaintFormatSuggestion writes a style redline on the live document and
// commits it. Like PaintSuggestion, the caller persists and closes the
// returned document.
func (s *Service) PaintFormatSuggestion(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	organizationID gid.GID,
	identityID gid.GID,
	profileID *gid.GID,
	suggestionID string,
	change FormatSuggestion,
) (*automerge.Document, automerge.ActorID, error) {
	return s.paint(
		ctx,
		scope,
		documentID,
		organizationID,
		identityID,
		profileID,
		func(document *automerge.Document) error {
			return ApplySuggestionFormat(document, suggestionID, change)
		},
	)
}

// PaintTableSuggestion writes a structural table redline on the live
// document and commits it. Like PaintSuggestion, the caller persists and
// closes the returned document.
func (s *Service) PaintTableSuggestion(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	organizationID gid.GID,
	identityID gid.GID,
	profileID *gid.GID,
	suggestionID string,
	change TableSuggestion,
) (*automerge.Document, automerge.ActorID, error) {
	return s.paint(
		ctx,
		scope,
		documentID,
		organizationID,
		identityID,
		profileID,
		func(document *automerge.Document) error {
			return ApplyTableSuggestion(document, suggestionID, change)
		},
	)
}

func (s *Service) paint(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	organizationID gid.GID,
	identityID gid.GID,
	profileID *gid.GID,
	apply func(document *automerge.Document) error,
) (*automerge.Document, automerge.ActorID, error) {
	actorID, err := s.actorForIdentity(
		ctx,
		scope,
		documentID,
		organizationID,
		identityID,
		profileID,
	)
	if err != nil {
		return nil, automerge.ActorID{}, err
	}

	opened, err := s.openWithActor(ctx, scope, documentID, actorID)
	if err != nil {
		return nil, automerge.ActorID{}, err
	}

	if err := apply(opened.Document); err != nil {
		_ = opened.Document.Close()

		return nil, automerge.ActorID{}, err
	}

	if _, err := opened.Document.Commit("Suggest", time.Now()); err != nil {
		_ = opened.Document.Close()

		return nil, automerge.ActorID{}, fmt.Errorf("cannot commit suggestion: %w", err)
	}

	return opened.Document, actorID, nil
}

// ApplySuggestionEdit marks quote as a redline for suggestionID and inserts
// replacement beside it. Shared prefix and suffix stay unmarked, matching the
// editor's replaceSelection. Text outside the quote is left alone.
func ApplySuggestionEdit(
	document *automerge.Document,
	suggestionID string,
	quote string,
	replacement string,
	occurrence int,
) error {
	if suggestionID == "" || quote == "" || occurrence < 0 {
		return ErrQuoteNotFound
	}

	spans, err := BodySpans(document)
	if err != nil {
		return err
	}

	visible, units, end := suggestionUnits(spans)

	start, quoteEnd, err := quoteRange(visible, quote, occurrence)
	if err != nil {
		return err
	}

	prefix, suffix := sharedUTF16Edges(quote, replacement)
	if prefix+suffix > quoteEnd-start {
		return ErrSuggestionUnchanged
	}

	deleteStart := start + prefix
	deleteEnd := quoteEnd - suffix

	inserted := edgeTrimmed(replacement, prefix, suffix)
	if deleteStart == deleteEnd && inserted == "" {
		return ErrSuggestionUnchanged
	}

	if suggestionRangeMarked(units, deleteStart, deleteEnd) {
		return ErrSuggestionOverlaps
	}

	text, err := document.Text("body")
	if err != nil {
		return fmt.Errorf("cannot get collaboration body: %w", err)
	}

	mark, err := suggestionMark(suggestionID)
	if err != nil {
		return err
	}

	textRanges, blocks := suggestionRanges(units, deleteStart, deleteEnd)
	if len(textRanges) == 0 && len(blocks) == 0 && inserted == "" {
		return ErrSuggestionUnchanged
	}

	for _, textRange := range textRanges {
		if err := text.Mark(
			textRange[0],
			textRange[1],
			SuggestionDeleteMark,
			mark,
			automerge.MarkExpandNone,
		); err != nil {
			return fmt.Errorf("cannot mark suggested deletion: %w", err)
		}
	}

	if inserted != "" {
		insertAt := end
		if int(deleteEnd) < len(units) {
			insertAt = units[deleteEnd].engine
		}

		if err := text.Splice(insertAt, 0, inserted); err != nil {
			return fmt.Errorf("cannot insert suggested text: %w", err)
		}

		if err := text.Mark(
			insertAt,
			insertAt+utf16Len(inserted),
			SuggestionInsertMark,
			mark,
			automerge.MarkExpandNone,
		); err != nil {
			return fmt.Errorf("cannot mark suggested insertion: %w", err)
		}
	}

	if len(blocks) == 0 {
		return nil
	}

	painted, err := BodySpans(document)
	if err != nil {
		return err
	}

	if err := WriteBodySpans(document, withDeletionSplits(painted, blocks, suggestionID)); err != nil {
		return fmt.Errorf("cannot mark suggested paragraph break: %w", err)
	}

	return nil
}

func suggestionMark(suggestionID string) (automerge.Scalar, error) {
	encoded, err := json.Marshal(map[string]string{"id": suggestionID})
	if err != nil {
		return automerge.Scalar{}, fmt.Errorf("cannot encode suggestion mark: %w", err)
	}

	return automerge.StringScalar(string(encoded)), nil
}

func suggestionUnits(spans []automerge.Span) (string, []suggestionUnit, uint32) {
	var visible strings.Builder

	type openCell struct {
		index int
		path  []string
	}

	units := make([]suggestionUnit, 0)
	engine := uint32(0)
	wrote := false
	cells := make([]openCell, 0)
	previousOwner := -1

	for index, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			parents := blockParents(span)
			for len(cells) > 0 && !hasPathPrefix(parents, cells[len(cells)-1].path) {
				cells = cells[:len(cells)-1]
			}

			if isTableCell(span) {
				cells = append(cells, openCell{index: index, path: blockPath(span)})
			}

			owner := -1
			if len(cells) > 0 {
				owner = cells[len(cells)-1].index
			}

			if wrote {
				splitID, _ := suggestionSplit(span)

				visible.WriteByte('\n')

				units = append(units, suggestionUnit{
					engine:   engine,
					block:    true,
					marked:   splitID != "",
					joinable: !isTableStructure(span) && owner == previousOwner,
				})
			}

			previousOwner = owner
			engine++

			continue
		}

		if span.Type != automerge.SpanTypeText || span.Text == "" {
			continue
		}

		marked := markID(span.Marks, SuggestionInsertMark) != "" ||
			markID(span.Marks, SuggestionDeleteMark) != ""

		for range utf16.Encode([]rune(span.Text)) {
			units = append(units, suggestionUnit{engine: engine, block: false, marked: marked})
			engine++
		}

		visible.WriteString(span.Text)

		wrote = true
	}

	return visible.String(), units, engine
}

func quoteRange(visible string, quote string, occurrence int) (uint32, uint32, error) {
	index, err := findQuote(visible, quote, occurrence)
	if err != nil {
		return 0, 0, err
	}

	start := utf16Len(visible[:index])

	return start, start + utf16Len(quote), nil
}

// sharedUTF16Edges matches the editor's sharedEdges: the longest common
// prefix and suffix, measured in UTF-16 code units so an emoji counts as two.
func sharedUTF16Edges(before string, after string) (uint32, uint32) {
	oldChars := []rune(before)
	newChars := []rune(after)
	prefix := 0

	for prefix < len(oldChars) && prefix < len(newChars) && oldChars[prefix] == newChars[prefix] {
		prefix++
	}

	suffix := 0
	for suffix < len(oldChars)-prefix &&
		suffix < len(newChars)-prefix &&
		oldChars[len(oldChars)-1-suffix] == newChars[len(newChars)-1-suffix] {
		suffix++
	}

	prefixWidth := utf16Len(string(oldChars[:prefix]))

	suffixWidth := uint32(0)
	if suffix > 0 {
		suffixWidth = utf16Len(string(oldChars[len(oldChars)-suffix:]))
	}

	return prefixWidth, suffixWidth
}

func edgeTrimmed(value string, prefix uint32, suffix uint32) string {
	units := utf16.Encode([]rune(value))
	if prefix+suffix > uint32(len(units)) {
		return ""
	}

	return string(utf16.Decode(units[prefix : len(units)-int(suffix)]))
}

func suggestionRangeMarked(units []suggestionUnit, from uint32, to uint32) bool {
	if from == to {
		return int(from) < len(units) && units[from].marked
	}

	for index := from; index < to && int(index) < len(units); index++ {
		if units[index].marked {
			return true
		}
	}

	return false
}

func suggestionRanges(units []suggestionUnit, from uint32, to uint32) ([][2]uint32, []uint32) {
	texts := make([][2]uint32, 0)
	blocks := make([]uint32, 0)
	open := false
	start := uint32(0)
	previous := uint32(0)

	for index := from; index < to; index++ {
		unit := units[index]
		if unit.block {
			if open {
				texts = append(texts, [2]uint32{start, previous + 1})
				open = false
			}

			if unit.joinable {
				blocks = append(blocks, unit.engine)
			}

			continue
		}

		if !open || unit.engine != previous+1 {
			if open {
				texts = append(texts, [2]uint32{start, previous + 1})
			}

			start = unit.engine
			open = true
		}

		previous = unit.engine
	}

	if open {
		texts = append(texts, [2]uint32{start, previous + 1})
	}

	return texts, blocks
}

func withDeletionSplits(spans []automerge.Span, blocks []uint32, suggestionID string) []automerge.Span {
	wanted := make(map[uint32]struct{}, len(blocks))
	for _, block := range blocks {
		wanted[block] = struct{}{}
	}

	engine := uint32(0)

	for index, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			if _, ok := wanted[engine]; ok {
				spans[index] = withDeletionSplit(span, suggestionID)
			}

			engine++

			continue
		}

		if span.Type == automerge.SpanTypeText {
			engine += utf16Len(span.Text)
		}
	}

	return spans
}

func withDeletionSplit(span automerge.Span, suggestionID string) automerge.Span {
	if id, _ := suggestionSplit(span); id != "" {
		return span
	}

	block := copyMap(span.Block)
	attrs, _ := block["attrs"].(map[string]any)
	nextAttrs := copyMap(attrs)
	nextAttrs[suggestionSplitAttr] = suggestionID
	nextAttrs[suggestionSplitKindAttr] = suggestionSplitDeletion
	block["attrs"] = nextAttrs
	span.Block = block

	return span
}

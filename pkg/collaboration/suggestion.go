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
	"time"
	"unicode/utf16"

	"go.gearno.de/kit/log"
	"go.gearno.de/kit/pg"
	"go.probo.inc/probo/pkg/automerge"
	collabpkg "go.probo.inc/probo/pkg/automerge/collaboration"
	automergeprosemirror "go.probo.inc/probo/pkg/automerge/prosemirror"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

const (
	SuggestionInsertMark = "suggestion-insert"
	SuggestionDeleteMark = "suggestion-delete"

	suggestionSplitAttr       = "suggestionSplit"
	suggestionSplitKindAttr   = "splitKind"
	suggestionSplitInsertion  = "insertion"
	suggestionSplitStructural = "structural"

	// BlockSpanWidth is the cursor-index width of one block marker. Block
	// operations are not characters in this text index, so the width is zero.
	BlockSpanWidth = 0
)

var (
	ErrOpenSuggestions         = errors.New("document has open suggestions")
	ErrSuggestionNotInDocument = errors.New("suggestion is not in the document")
	ErrQuoteNotFound           = errors.New("quoted text was not found in the document")
)

type suggestionCoverage struct {
	insert string
	delete string
}

// ProjectAcceptedSpans drops suggested insertions and keeps strikethrough
// text without the suggestion mark. The result is the document readers see
// outside the editor.
func ProjectAcceptedSpans(spans []automerge.Span) []automerge.Span {
	projected := make([]automerge.Span, 0, len(spans))

	for _, span := range spans {
		if span.Type == automerge.SpanTypeText && markID(span.Marks, SuggestionInsertMark) != "" {
			continue
		}

		if span.Type == automerge.SpanTypeText {
			span.Marks = withoutMark(span.Marks, SuggestionDeleteMark)
		}

		projected = append(projected, span)
	}

	return omitOpenSplits(projected)
}

// HasOpenSuggestions reports whether the document still carries a redline.
func HasOpenSuggestions(spans []automerge.Span) bool {
	for _, span := range spans {
		if id, _ := suggestionSplit(span); id != "" {
			return true
		}

		if span.Type != automerge.SpanTypeText {
			continue
		}

		if markID(span.Marks, SuggestionInsertMark) != "" || markID(span.Marks, SuggestionDeleteMark) != "" {
			return true
		}
	}

	return false
}

// CommentChangeAllowed reports whether a comment-only peer's sync is a
// suggestion. The accepted document must stay the same, and existing
// redlines may grow but must not shrink.
func CommentChangeAllowed(before []automerge.Span, after []automerge.Span) bool {
	if !spansEqual(normalizeSuggestionSpans(before), normalizeSuggestionSpans(after)) {
		return false
	}

	return coveragePreserved(suggestionCoverageByID(before), suggestionCoverageByID(after))
}

// ApplySuggestionDecision rewrites body spans for an accept or reject.
// Accept keeps insertions and drops strikethrough text. Reject does the opposite.
// changed is false when the suggestion id is not in the document.
func ApplySuggestionDecision(
	document *automerge.Document,
	suggestionID string,
	accept bool,
) (bool, error) {
	spans, err := BodySpans(document)
	if err != nil {
		return false, err
	}

	next, changed := decideSuggestionSpans(spans, suggestionID, accept)
	if !changed {
		return false, nil
	}

	if err := WriteBodySpans(document, next); err != nil {
		return false, err
	}

	if _, err := document.Commit("Resolve suggestion", time.Now()); err != nil && !errors.Is(err, automerge.ErrNoOperations) {
		return false, fmt.Errorf("cannot commit suggestion decision: %w", err)
	}

	return true, nil
}

// BodySpans reads the collaborative body. A document with no body has no spans.
func BodySpans(document *automerge.Document) ([]automerge.Span, error) {
	text, err := document.Text("body")
	if err != nil {
		return nil, fmt.Errorf("cannot get collaboration body: %w", err)
	}

	spans, err := text.Spans()
	if err != nil {
		return nil, fmt.Errorf("cannot read collaboration spans: %w", err)
	}

	return spans, nil
}

// WriteBodySpans replaces the body with spans. Suggestion marks do not expand
// into text typed after them.
func WriteBodySpans(document *automerge.Document, spans []automerge.Span) error {
	text, err := document.Text("body")
	if err != nil {
		return fmt.Errorf("cannot get collaboration body: %w", err)
	}

	inputs, err := spanInputs(spans)
	if err != nil {
		return err
	}

	config := automergeprosemirror.UpdateSpansConfig()
	if config.PerMarkExpands == nil {
		config.PerMarkExpands = map[string]automerge.MarkExpand{}
	}

	config.PerMarkExpands[SuggestionInsertMark] = automerge.MarkExpandNone
	config.PerMarkExpands[SuggestionDeleteMark] = automerge.MarkExpandNone

	if err := text.UpdateSpans(inputs, config); err != nil {
		return fmt.Errorf("cannot write collaboration spans: %w", err)
	}

	return nil
}

// QuoteCursors finds quote in the visible text and returns Automerge cursors
// for the match. occurrence is zero-based.
func QuoteCursors(
	document *automerge.Document,
	quote string,
	occurrence int,
) (start []byte, end []byte, err error) {
	if quote == "" || occurrence < 0 {
		return nil, nil, ErrQuoteNotFound
	}

	spans, err := BodySpans(document)
	if err != nil {
		return nil, nil, err
	}

	visible, offsets := visibleText(spans)
	index := -1
	searchFrom := 0

	for range occurrence + 1 {
		found := strings.Index(visible[searchFrom:], quote)
		if found < 0 {
			return nil, nil, ErrQuoteNotFound
		}

		index = searchFrom + found
		searchFrom = index + len(quote)
	}

	startUTF16 := utf16Len(visible[:index])
	endUTF16 := startUTF16 + utf16Len(quote)
	absoluteStart, ok := absoluteUTF16(offsets, startUTF16)
	if !ok {
		return nil, nil, ErrQuoteNotFound
	}

	absoluteEnd, ok := absoluteUTF16(offsets, endUTF16)
	if !ok {
		return nil, nil, ErrQuoteNotFound
	}

	text, err := document.Text("body")
	if err != nil {
		return nil, nil, fmt.Errorf("cannot get collaboration body: %w", err)
	}

	startCursor, err := textCursor(text, absoluteStart)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create suggestion start cursor: %w", err)
	}

	endCursor, err := textCursor(text, absoluteEnd)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot create suggestion end cursor: %w", err)
	}

	return startCursor, endCursor, nil
}

// SuggestionQuote returns the visible text covered by a suggestion id.
func SuggestionQuote(spans []automerge.Span, suggestionID string) string {
	var quote strings.Builder

	wrote := false
	pendingBreak := false

	for _, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			if wrote {
				pendingBreak = true
			}

			continue
		}

		if span.Type != automerge.SpanTypeText {
			continue
		}

		if markID(span.Marks, SuggestionInsertMark) != suggestionID &&
			markID(span.Marks, SuggestionDeleteMark) != suggestionID {
			continue
		}

		if pendingBreak {
			quote.WriteByte('\n')
			pendingBreak = false
		}

		quote.WriteString(span.Text)
		wrote = true
	}

	return quote.String()
}

// SuggestionCursors returns cursors covering every span of a suggestion.
func SuggestionCursors(
	document *automerge.Document,
	suggestionID string,
) (start []byte, end []byte, quote string, err error) {
	spans, err := BodySpans(document)
	if err != nil {
		return nil, nil, "", err
	}

	var (
		found      bool
		startUTF16 uint32
		endUTF16   uint32
		absolute   uint32
	)

	for _, span := range spans {
		width := spanUTF16Width(span)
		matches := span.Type == automerge.SpanTypeText &&
			(markID(span.Marks, SuggestionInsertMark) == suggestionID ||
				markID(span.Marks, SuggestionDeleteMark) == suggestionID)

		if matches {
			if !found {
				startUTF16 = absolute
				found = true
			}

			endUTF16 = absolute + width
		}

		absolute += width
	}

	if !found {
		return nil, nil, "", ErrSuggestionNotInDocument
	}

	text, err := document.Text("body")
	if err != nil {
		return nil, nil, "", fmt.Errorf("cannot get collaboration body: %w", err)
	}

	startCursor, err := textCursor(text, startUTF16)
	if err != nil {
		return nil, nil, "", fmt.Errorf("cannot create suggestion start cursor: %w", err)
	}

	endCursor, err := textCursor(text, endUTF16)
	if err != nil {
		return nil, nil, "", fmt.Errorf("cannot create suggestion end cursor: %w", err)
	}

	return startCursor, endCursor, SuggestionQuote(spans, suggestionID), nil
}

func (s *Service) EnsureNoOpenSuggestions(
	ctx context.Context,
	tx pg.Tx,
	scope coredata.Scoper,
	documentID gid.GID,
) error {
	doc := &coredata.CollaborationDocument{}
	if err := doc.LoadByDocumentID(ctx, tx, scope, documentID); err != nil {
		if errors.Is(err, coredata.ErrResourceNotFound) {
			return nil
		}

		return fmt.Errorf("cannot load collaboration document: %w", err)
	}

	if !doc.Seeded {
		return nil
	}

	changes, err := s.collectChanges(
		ctx,
		tx,
		scope,
		documentID,
		doc.SnapshotRevision,
		doc.ChangeRevision,
	)
	if err != nil {
		return err
	}

	document, err := loadAutomergeDocument(ctx, doc.Snapshot)
	if err != nil {
		return fmt.Errorf("cannot open collaboration snapshot: %w", err)
	}

	defer func() { _ = document.Close() }()

	if len(changes) > 0 {
		if err := document.ApplyChanges(changes); err != nil {
			return fmt.Errorf("cannot apply collaboration changes: %w", err)
		}
	}

	spans, err := BodySpans(document)
	if err != nil {
		return err
	}

	if HasOpenSuggestions(spans) {
		return ErrOpenSuggestions
	}

	return nil
}

func decideSuggestionSpans(
	spans []automerge.Span,
	suggestionID string,
	accept bool,
) ([]automerge.Span, bool) {
	next := make([]automerge.Span, 0, len(spans))
	changed := false

	for _, span := range spans {
		if span.Type != automerge.SpanTypeText {
			id, kind := suggestionSplit(span)
			if id == suggestionID {
				changed = true
				if !accept && kind == suggestionSplitStructural {
					continue
				}

				if accept {
					next = append(next, clearSuggestionSplit(span))
					continue
				}
			}

			next = append(next, span)
			continue
		}

		insertID := markID(span.Marks, SuggestionInsertMark)
		deleteID := markID(span.Marks, SuggestionDeleteMark)

		switch {
		case accept && insertID == suggestionID && insertID != "":
			span.Marks = withoutMark(span.Marks, SuggestionInsertMark)
			changed = true
		case accept && deleteID == suggestionID && deleteID != "":
			changed = true
			continue
		case !accept && insertID == suggestionID && insertID != "":
			changed = true
			continue
		case !accept && deleteID == suggestionID && deleteID != "":
			span.Marks = withoutMark(span.Marks, SuggestionDeleteMark)
			changed = true
		}

		next = append(next, span)
	}

	if !accept {
		next = dropEmptyInsertionBlocks(next, suggestionID)
	}

	return next, changed
}

func omitOpenSplits(spans []automerge.Span) []automerge.Span {
	next := make([]automerge.Span, 0, len(spans))

	for index, span := range spans {
		id, kind := suggestionSplit(span)
		if id == "" {
			next = append(next, span)
			continue
		}

		if kind == suggestionSplitStructural ||
			(kind == suggestionSplitInsertion && !blockHasFollowingText(spans, index)) {
			continue
		}

		next = append(next, clearSuggestionSplit(span))
	}

	return restoreBlock(spans, next)
}

func dropEmptyInsertionBlocks(spans []automerge.Span, suggestionID string) []automerge.Span {
	next := make([]automerge.Span, 0, len(spans))

	for index, span := range spans {
		id, kind := suggestionSplit(span)
		if id == suggestionID && kind == suggestionSplitInsertion {
			if !blockHasFollowingText(spans, index) {
				continue
			}

			next = append(next, clearSuggestionSplit(span))
			continue
		}

		next = append(next, span)
	}

	return restoreBlock(spans, next)
}

func restoreBlock(original []automerge.Span, next []automerge.Span) []automerge.Span {
	if hasBlock(next) || !hasBlock(original) {
		return next
	}

	for _, span := range original {
		if span.Type != automerge.SpanTypeBlock {
			continue
		}

		return append([]automerge.Span{clearSuggestionSplit(span)}, next...)
	}

	return next
}

func blockHasFollowingText(spans []automerge.Span, index int) bool {
	for _, span := range spans[index+1:] {
		if span.Type == automerge.SpanTypeBlock {
			return false
		}

		if span.Text != "" {
			return true
		}
	}

	return false
}

func hasBlock(spans []automerge.Span) bool {
	for _, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			return true
		}
	}

	return false
}

func suggestionSplit(span automerge.Span) (string, string) {
	if span.Type != automerge.SpanTypeBlock || span.Block == nil {
		return "", ""
	}

	attrs, ok := span.Block["attrs"].(map[string]any)
	if !ok {
		return "", ""
	}

	id, _ := attrs[suggestionSplitAttr].(string)
	kind, _ := attrs[suggestionSplitKindAttr].(string)

	return id, kind
}

func clearSuggestionSplit(span automerge.Span) automerge.Span {
	block := copyMap(span.Block)
	attrs, ok := block["attrs"].(map[string]any)
	if !ok {
		span.Block = block
		return span
	}

	nextAttrs := copyMap(attrs)
	delete(nextAttrs, suggestionSplitAttr)
	delete(nextAttrs, suggestionSplitKindAttr)
	if len(nextAttrs) == 0 {
		delete(block, "attrs")
	} else {
		block["attrs"] = nextAttrs
	}

	span.Block = block

	return span
}

func copyMap(value map[string]any) map[string]any {
	next := make(map[string]any, len(value))
	for key, item := range value {
		next[key] = item
	}

	return next
}

func normalizeSuggestionSpans(spans []automerge.Span) []automerge.Span {
	normalized := make([]automerge.Span, 0, len(spans))

	for _, span := range spans {
		if span.Type == automerge.SpanTypeText && markID(span.Marks, SuggestionInsertMark) != "" {
			continue
		}

		if span.Type == automerge.SpanTypeText {
			span.Marks = withoutMark(span.Marks, SuggestionDeleteMark)
		}

		normalized = append(normalized, span)
	}

	return normalized
}

func suggestionCoverageByID(spans []automerge.Span) map[string]suggestionCoverage {
	coverage := map[string]suggestionCoverage{}

	for _, span := range spans {
		if span.Type != automerge.SpanTypeText {
			continue
		}

		if id := markID(span.Marks, SuggestionInsertMark); id != "" {
			current := coverage[id]
			current.insert += span.Text
			coverage[id] = current
		}

		if id := markID(span.Marks, SuggestionDeleteMark); id != "" {
			current := coverage[id]
			current.delete += span.Text
			coverage[id] = current
		}
	}

	return coverage
}

func coveragePreserved(before map[string]suggestionCoverage, after map[string]suggestionCoverage) bool {
	for id, previous := range before {
		next, ok := after[id]
		if !ok {
			return false
		}

		if !strings.Contains(next.insert, previous.insert) || !strings.Contains(next.delete, previous.delete) {
			return false
		}
	}

	return true
}

func spansEqual(left []automerge.Span, right []automerge.Span) bool {
	if len(left) != len(right) {
		return false
	}

	for index := range left {
		if left[index].Type != right[index].Type || left[index].Text != right[index].Text {
			return false
		}

		if !jsonEqual(left[index].Block, right[index].Block) || !jsonEqual(left[index].Marks, right[index].Marks) {
			return false
		}
	}

	return true
}

func jsonEqual(left any, right any) bool {
	leftRaw, leftErr := json.Marshal(left)
	rightRaw, rightErr := json.Marshal(right)
	if leftErr != nil || rightErr != nil {
		return false
	}

	return string(leftRaw) == string(rightRaw)
}

func markID(marks map[string]any, name string) string {
	if len(marks) == 0 {
		return ""
	}

	value, ok := marks[name]
	if !ok || value == nil {
		return ""
	}

	raw, ok := value.(string)
	if !ok || raw == "" {
		return ""
	}

	var payload struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil || payload.ID == "" {
		return ""
	}

	return payload.ID
}

func withoutMark(marks map[string]any, name string) map[string]any {
	if len(marks) == 0 {
		return nil
	}

	next := make(map[string]any, len(marks))
	for key, value := range marks {
		if key == name {
			continue
		}

		next[key] = value
	}

	if len(next) == 0 {
		return nil
	}

	return next
}

func spanInputs(spans []automerge.Span) ([]automerge.SpanInput, error) {
	inputs := make([]automerge.SpanInput, 0, len(spans))

	for _, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			inputs = append(inputs, automerge.SpanInput{Block: span.Block})
			continue
		}

		marks := make(map[string]automerge.Scalar, len(span.Marks))
		for name, value := range span.Marks {
			scalar, err := scalarFromAny(value)
			if err != nil {
				return nil, fmt.Errorf("cannot encode span mark %q: %w", name, err)
			}

			marks[name] = scalar
		}

		inputs = append(inputs, automerge.SpanInput{Text: span.Text, Marks: marks})
	}

	return inputs, nil
}

func scalarFromAny(value any) (automerge.Scalar, error) {
	switch typed := value.(type) {
	case nil:
		return automerge.NullScalar(), nil
	case string:
		return automerge.StringScalar(typed), nil
	case bool:
		return automerge.BoolScalar(typed), nil
	case float64:
		return automerge.FloatScalar(typed), nil
	default:
		raw, err := json.Marshal(typed)
		if err != nil {
			return automerge.Scalar{}, err
		}

		return automerge.StringScalar(string(raw)), nil
	}
}

type visibleOffset struct {
	visible  uint32
	absolute uint32
}

func visibleText(spans []automerge.Span) (string, []visibleOffset) {
	var visible strings.Builder

	offsets := make([]visibleOffset, 0, len(spans)+1)
	var visibleUTF16 uint32
	var absolute uint32

	offsets = append(offsets, visibleOffset{visible: 0, absolute: 0})

	for _, span := range spans {
		width := spanUTF16Width(span)
		if span.Type == automerge.SpanTypeText && span.Text != "" {
			visible.WriteString(span.Text)
			visibleUTF16 += width
			absolute += width
			offsets = append(offsets, visibleOffset{visible: visibleUTF16, absolute: absolute})
			continue
		}

		absolute += width
	}

	return visible.String(), offsets
}

func absoluteUTF16(offsets []visibleOffset, visible uint32) (uint32, bool) {
	if len(offsets) == 0 {
		return 0, false
	}

	for index := 1; index < len(offsets); index++ {
		previous := offsets[index-1]
		current := offsets[index]
		if visible < previous.visible || visible > current.visible {
			continue
		}

		return previous.absolute + (visible - previous.visible), true
	}

	last := offsets[len(offsets)-1]
	if visible == last.visible {
		return last.absolute, true
	}

	return 0, false
}

func spanUTF16Width(span automerge.Span) uint32 {
	if span.Type == automerge.SpanTypeBlock {
		return BlockSpanWidth
	}

	return utf16Len(span.Text)
}

func utf16Len(value string) uint32 {
	return uint32(len(utf16.Encode([]rune(value))))
}

func textCursor(text *automerge.Text, index uint32) ([]byte, error) {
	cursor, err := text.Cursor(index)
	if err != nil {
		if index > 0 && strings.Contains(err.Error(), "out of bounds") {
			return append([]byte(nil), automerge.EndCursor()...), nil
		}

		return nil, err
	}

	return append([]byte(nil), cursor...), nil
}

// reviewEphemeralFrame tells every open editor to refetch the margin.
// The protocol document id is the bare id. An automerge: URL is a different
// handle, so the tab that already has the redline never hears the notice.
func reviewEphemeralFrame(documentID gid.GID) ([]byte, error) {
	return collabpkg.EncodeReviewEphemeral(
		"probo-gateway",
		"probo-gateway",
		collabpkg.DeriveDocumentID(documentID.String()),
		fmt.Sprintf("%s:%d", documentID.String(), time.Now().UnixNano()),
	)
}

// NotifyReview tells open editors to refetch comment threads and suggestions.
func (s *Service) NotifyReview(
	ctx context.Context,
	documentID gid.GID,
) error {
	frame, err := reviewEphemeralFrame(documentID)
	if err != nil {
		return err
	}

	if err := s.NotifyEphemeral(ctx, documentID, "", frame); err != nil {
		s.logger.WarnCtx(
			ctx,
			"cannot notify document review",
			log.String("document_id", documentID.String()),
			log.Error(err),
		)
		return err
	}

	return nil
}

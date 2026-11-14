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
	"fmt"
	"strings"
	"sync/atomic"

	"go.probo.inc/probo/pkg/automerge"
	"go.probo.inc/probo/pkg/coredata"
	"go.probo.inc/probo/pkg/gid"
)

// An applied edit is painted with a throwaway suggestion id and accepted
// before the commit, so the stored document has the change and no redline.
// Other open suggestions stay where they are. The painters already refuse a
// quote that covers one of them.
var syntheticSuggestionSeq atomic.Uint64

func syntheticSuggestionID() string {
	return fmt.Sprintf("apply-%d", syntheticSuggestionSeq.Add(1))
}

// ApplyDirectText writes quote's replacement with no suggestion marks. A
// newline inside one block becomes a hard break. A newline that replaces a
// paragraph boundary stays a paragraph split, because the suggestion painter
// already treats that boundary as a block.
func ApplyDirectText(
	document *automerge.Document,
	quote string,
	replacement string,
	occurrence int,
) error {
	inserted := ""

	err := applyThenAccept(document, func(id string) error {
		spans, err := BodySpans(document)
		if err != nil {
			return err
		}

		visible, _, _ := suggestionUnits(spans)
		if _, _, err := quoteRange(visible, quote, occurrence); err != nil {
			return err
		}

		prefix, suffix := sharedUTF16Edges(quote, replacement)
		inserted = edgeTrimmed(replacement, prefix, suffix)

		return ApplySuggestionEdit(document, id, quote, replacement, occurrence)
	})
	if err != nil {
		return err
	}

	if !strings.Contains(inserted, "\n") || strings.Contains(quote, "\n") {
		return nil
	}

	return splitHardBreaks(document, inserted)
}

// ApplyDirectFormat sets or clears a style mark and leaves no suggestion mark.
func ApplyDirectFormat(document *automerge.Document, change FormatSuggestion) error {
	return applyThenAccept(document, func(id string) error {
		return ApplySuggestionFormat(document, id, change)
	})
}

// ApplyDirectTable inserts or deletes the row, column, or table. Removed
// blocks are dropped and new cells keep their text, with no suggestion marks.
func ApplyDirectTable(document *automerge.Document, change TableSuggestion) error {
	return applyThenAccept(document, func(id string) error {
		return ApplyTableSuggestion(document, id, change)
	})
}

// PaintDirectText applies a text edit on the live document and commits it.
func (s *Service) PaintDirectText(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	organizationID gid.GID,
	identityID gid.GID,
	profileID *gid.GID,
	quote string,
	replacement string,
	occurrence int,
) (*automerge.Document, automerge.ActorID, error) {
	return s.paintCommit(
		ctx,
		scope,
		documentID,
		organizationID,
		identityID,
		profileID,
		"Edit",
		func(document *automerge.Document) error {
			return ApplyDirectText(document, quote, replacement, occurrence)
		},
	)
}

// PaintDirectFormat applies a style edit on the live document and commits it.
func (s *Service) PaintDirectFormat(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	organizationID gid.GID,
	identityID gid.GID,
	profileID *gid.GID,
	change FormatSuggestion,
) (*automerge.Document, automerge.ActorID, error) {
	return s.paintCommit(
		ctx,
		scope,
		documentID,
		organizationID,
		identityID,
		profileID,
		"Edit",
		func(document *automerge.Document) error {
			return ApplyDirectFormat(document, change)
		},
	)
}

// PaintDirectTable applies a table edit on the live document and commits it.
func (s *Service) PaintDirectTable(
	ctx context.Context,
	scope coredata.Scoper,
	documentID gid.GID,
	organizationID gid.GID,
	identityID gid.GID,
	profileID *gid.GID,
	change TableSuggestion,
) (*automerge.Document, automerge.ActorID, error) {
	return s.paintCommit(
		ctx,
		scope,
		documentID,
		organizationID,
		identityID,
		profileID,
		"Edit",
		func(document *automerge.Document) error {
			return ApplyDirectTable(document, change)
		},
	)
}

func applyThenAccept(document *automerge.Document, paint func(id string) error) error {
	id := syntheticSuggestionID()
	if err := paint(id); err != nil {
		return err
	}

	spans, err := BodySpans(document)
	if err != nil {
		return err
	}

	next, changed := decideSuggestionSpans(spans, id, true)
	if !changed {
		return ErrSuggestionUnchanged
	}

	if err := WriteBodySpans(document, next); err != nil {
		return fmt.Errorf("cannot write applied edit: %w", err)
	}

	return nil
}

// splitHardBreaks turns newlines that were spliced as characters into hard-break
// embeds. The quote stayed inside one block, so the newline is not a new paragraph.
func splitHardBreaks(document *automerge.Document, inserted string) error {
	spans, err := BodySpans(document)
	if err != nil {
		return err
	}

	next, ok := spansWithHardBreaks(spans, inserted)
	if !ok {
		return nil
	}

	if err := WriteBodySpans(document, next); err != nil {
		return fmt.Errorf("cannot write hard breaks: %w", err)
	}

	return nil
}

func spansWithHardBreaks(spans []automerge.Span, inserted string) ([]automerge.Span, bool) {
	parents := []string(nil)

	for index, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			if !blockIsEmbed(span) {
				parents = blockPath(span)
			}

			continue
		}

		if span.Type != automerge.SpanTypeText || !strings.Contains(span.Text, inserted) {
			continue
		}

		at := strings.Index(span.Text, inserted)
		before := span.Text[:at]
		after := span.Text[at+len(inserted):]
		parts := strings.Split(inserted, "\n")
		replacement := make([]automerge.Span, 0, len(parts)*2+2)

		if before != "" {
			replacement = append(replacement, textSpan(before, span.Marks))
		}

		for partIndex, part := range parts {
			if partIndex > 0 {
				replacement = append(replacement, hardBreakSpan(parents))
			}

			if part != "" {
				replacement = append(replacement, textSpan(part, span.Marks))
			}
		}

		if after != "" {
			replacement = append(replacement, textSpan(after, span.Marks))
		}

		next := make([]automerge.Span, 0, len(spans)+len(replacement))
		next = append(next, spans[:index]...)
		next = append(next, replacement...)
		next = append(next, spans[index+1:]...)

		return next, true
	}

	return nil, false
}

func textSpan(text string, marks map[string]any) automerge.Span {
	span := automerge.Span{Type: automerge.SpanTypeText, Text: text}
	if len(marks) > 0 {
		span.Marks = copyMap(marks)
	}

	return span
}

func hardBreakSpan(parents []string) automerge.Span {
	return automerge.Span{
		Type: automerge.SpanTypeBlock,
		Block: map[string]any{
			"type":    blockTypeHardBreak,
			"parents": toAnyStrings(parents),
			"attrs":   map[string]any{},
			"isEmbed": true,
		},
	}
}

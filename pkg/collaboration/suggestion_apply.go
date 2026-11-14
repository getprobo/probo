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
	"errors"
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
// newline inside one block becomes a hard break, except in a code block,
// which keeps it as text. A newline that replaces a paragraph boundary stays
// a paragraph split, because the suggestion painter already treats that
// boundary as a block.
func ApplyDirectText(
	document *automerge.Document,
	quote string,
	replacement string,
	occurrence int,
) error {
	split := false

	err := applyThenAccept(document, func(id string) error {
		if err := ApplySuggestionEdit(document, id, quote, replacement, occurrence); err != nil {
			return err
		}

		if !strings.Contains(replacement, "\n") || strings.Contains(quote, "\n") {
			return nil
		}

		var err error

		split, err = splitHardBreaks(document, id)

		return err
	})

	// An edit that only adds a line break leaves no marked text to accept.
	if split && errors.Is(err, ErrSuggestionUnchanged) {
		return nil
	}

	return err
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

// splitHardBreaks turns the newlines this edit spliced as characters into
// hard-break embeds. Only text carrying the edit's own insertion mark is
// split, so the same text elsewhere in the document is left alone.
func splitHardBreaks(document *automerge.Document, id string) (bool, error) {
	spans, err := BodySpans(document)
	if err != nil {
		return false, err
	}

	next, ok := spansWithHardBreaks(spans, id)
	if !ok {
		return false, nil
	}

	if err := WriteBodySpans(document, next); err != nil {
		return false, fmt.Errorf("cannot write hard breaks: %w", err)
	}

	return true, nil
}

func spansWithHardBreaks(spans []automerge.Span, id string) ([]automerge.Span, bool) {
	var (
		parents []string
		code    bool
		changed bool
	)

	next := make([]automerge.Span, 0, len(spans))

	for _, span := range spans {
		if span.Type == automerge.SpanTypeBlock {
			if !blockIsEmbed(span) {
				parents = blockPath(span)
				code = blockType(span) == blockTypeCode
			}

			next = append(next, span)

			continue
		}

		split := span.Type == automerge.SpanTypeText &&
			!code &&
			strings.Contains(span.Text, "\n") &&
			markID(span.Marks, SuggestionInsertMark) == id
		if !split {
			next = append(next, span)

			continue
		}

		for partIndex, part := range strings.Split(span.Text, "\n") {
			if partIndex > 0 {
				next = append(next, hardBreakSpan(parents))
			}

			if part != "" {
				next = append(next, textSpan(part, span.Marks))
			}
		}

		changed = true
	}

	return next, changed
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

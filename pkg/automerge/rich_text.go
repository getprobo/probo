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

package automerge

import (
	"context"
	"encoding/json"
	"fmt"
)

type (
	SpanType string

	// MarkExpand controls whether inserted text inherits a mark at its edges.
	MarkExpand string

	Span struct {
		Type  SpanType
		Text  string
		Marks map[string]any
		Block map[string]any
	}

	encodedSpan struct {
		Type  SpanType        `json:"type"`
		Value json.RawMessage `json:"value"`
		Marks map[string]any  `json:"marks"`
	}
)

const (
	SpanTypeBlock SpanType = "block"
	SpanTypeText  SpanType = "text"

	MarkExpandBefore MarkExpand = "before"
	MarkExpandAfter  MarkExpand = "after"
	MarkExpandBoth   MarkExpand = "both"
	MarkExpandNone   MarkExpand = "none"
)

// Mark applies a typed annotation to a UTF-16 text range.
func (t *Text) Mark(
	ctx context.Context,
	start uint32,
	end uint32,
	name string,
	value Scalar,
	expand MarkExpand,
) error {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return ErrClosed
	}

	if !validMarkExpand(expand) {
		return fmt.Errorf("unknown Automerge mark expansion %q", expand)
	}

	encoded, err := encodeScalarWire(value)
	if err != nil {
		return fmt.Errorf("cannot encode Automerge mark value: %w", err)
	}

	if err := t.document.backend.MarkText(
		ctx,
		t.handle,
		start,
		end,
		name,
		encoded,
		string(expand),
	); err != nil {
		return fmt.Errorf("cannot mark Automerge text: %w", err)
	}

	return nil
}

// Unmark removes a named annotation from a UTF-16 text range.
func (t *Text) Unmark(
	ctx context.Context,
	start uint32,
	end uint32,
	name string,
	expand MarkExpand,
) error {
	return t.Mark(
		ctx,
		start,
		end,
		name,
		Scalar{Type: ScalarTypeNull},
		expand,
	)
}

// SplitBlock inserts a block marker at a UTF-16 text position.
func (t *Text) SplitBlock(ctx context.Context, index uint32) (*Object, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	handle, err := t.document.backend.SplitBlock(ctx, t.handle, index)
	if err != nil {
		return nil, fmt.Errorf("cannot split Automerge block: %w", err)
	}

	return &Object{
		document: t.document,
		handle:   handle,
		Type:     ObjectTypeMap,
	}, nil
}

// JoinBlock deletes the block marker at a UTF-16 text position.
func (t *Text) JoinBlock(ctx context.Context, index uint32) error {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return ErrClosed
	}

	if err := t.document.backend.JoinBlock(ctx, t.handle, index); err != nil {
		return fmt.Errorf("cannot join Automerge block: %w", err)
	}

	return nil
}

// ReplaceBlock replaces a block marker and returns its new map object.
func (t *Text) ReplaceBlock(ctx context.Context, index uint32) (*Object, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	handle, err := t.document.backend.ReplaceBlock(ctx, t.handle, index)
	if err != nil {
		return nil, fmt.Errorf("cannot replace Automerge block: %w", err)
	}

	return &Object{
		document: t.document,
		handle:   handle,
		Type:     ObjectTypeMap,
	}, nil
}

type (
	// Mark is one active annotation over a UTF-16 range of a text object.
	Mark struct {
		Start uint32
		End   uint32
		Name  string
		Value Scalar
	}

	encodedMark struct {
		Start uint32          `json:"start"`
		End   uint32          `json:"end"`
		Name  string          `json:"name"`
		Value json.RawMessage `json:"value"`
	}
)

// Marks returns the active marks over the text object as UTF-16 ranges.
func (t *Text) Marks(ctx context.Context) ([]Mark, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	data, err := t.document.backend.Marks(ctx, t.handle)
	if err != nil {
		return nil, fmt.Errorf("cannot read Automerge marks: %w", err)
	}

	return decodeMarks(data)
}

// MarksAt returns the active marks over the text object at a historical frontier.
func (t *Text) MarksAt(ctx context.Context, heads []Hash) ([]Mark, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	data, err := t.document.backend.MarksAt(ctx, t.handle, backendHashes(heads))
	if err != nil {
		return nil, fmt.Errorf("cannot read historical Automerge marks: %w", err)
	}

	return decodeMarks(data)
}

func decodeMarks(data []byte) ([]Mark, error) {
	var encoded []encodedMark
	if err := json.Unmarshal(data, &encoded); err != nil {
		return nil, fmt.Errorf("cannot decode Automerge marks: %w", err)
	}

	marks := make([]Mark, len(encoded))
	for i, source := range encoded {
		value, err := decodeScalarWire(source.Value)
		if err != nil {
			return nil, fmt.Errorf("cannot decode Automerge mark %d value: %w", i, err)
		}

		marks[i] = Mark{
			Start: source.Start,
			End:   source.End,
			Name:  source.Name,
			Value: value,
		}
	}

	return marks, nil
}

func (t *Text) Spans(ctx context.Context) ([]Span, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	data, err := t.document.backend.TextSpans(ctx, t.handle)
	if err != nil {
		return nil, fmt.Errorf("cannot read Automerge rich-text spans: %w", err)
	}

	var encoded []encodedSpan
	if err := json.Unmarshal(data, &encoded); err != nil {
		return nil, fmt.Errorf("cannot decode Automerge rich-text spans: %w", err)
	}

	spans := make([]Span, len(encoded))
	for i, source := range encoded {
		spans[i].Type = source.Type
		spans[i].Marks = source.Marks

		switch source.Type {
		case SpanTypeText:
			if err := json.Unmarshal(source.Value, &spans[i].Text); err != nil {
				return nil, fmt.Errorf("cannot decode Automerge text span %d: %w", i, err)
			}
		case SpanTypeBlock:
			if err := json.Unmarshal(source.Value, &spans[i].Block); err != nil {
				return nil, fmt.Errorf("cannot decode Automerge block span %d: %w", i, err)
			}
		default:
			return nil, fmt.Errorf("cannot decode Automerge span %d: unknown type %q", i, source.Type)
		}
	}

	return spans, nil
}

func validMarkExpand(value MarkExpand) bool {
	switch value {
	case MarkExpandBefore,
		MarkExpandAfter,
		MarkExpandBoth,
		MarkExpandNone:
		return true
	default:
		return false
	}
}

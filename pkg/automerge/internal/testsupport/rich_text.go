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

package testsupport

import (
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

	if err := t.document.engine.MarkText(
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

// SpanInput is one span supplied to UpdateSpans. When Block is non-nil the span
// is a block marker carrying those attributes; otherwise it is a text span whose
// Marks map names the annotations active over the span.
type SpanInput struct {
	Text  string
	Marks map[string]Scalar
	Block map[string]any
}

// UpdateSpansConfig controls the mark expansion applied by UpdateSpans.
type UpdateSpansConfig struct {
	DefaultExpand  MarkExpand
	PerMarkExpands map[string]MarkExpand
}

// UpdateSpans reconciles the text so its spans equal the supplied spans,
// computing a minimal text diff and then setting the marks to exactly those
// named on the spans. It mirrors the Rust updateSpans helper.
func (t *Text) UpdateSpans(
	spans []SpanInput,
	config UpdateSpansConfig,
) error {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return ErrClosed
	}

	encodedSpans := make([]map[string]any, 0, len(spans))

	for _, span := range spans {
		if span.Block != nil {
			encodedSpans = append(
				encodedSpans,
				map[string]any{
					"type":  "block",
					"block": span.Block,
				},
			)

			continue
		}

		marks := make(map[string]json.RawMessage, len(span.Marks))

		for name, value := range span.Marks {
			encoded, err := encodeScalarWire(value)
			if err != nil {
				return fmt.Errorf("cannot encode span mark %q: %w", name, err)
			}

			marks[name] = json.RawMessage(encoded)
		}

		encodedSpans = append(
			encodedSpans,
			map[string]any{
				"type":  "text",
				"text":  span.Text,
				"marks": marks,
			},
		)
	}

	spansPayload, err := json.Marshal(encodedSpans)
	if err != nil {
		return fmt.Errorf("cannot encode Automerge spans: %w", err)
	}

	perMark := make(map[string]string, len(config.PerMarkExpands))
	for name, expand := range config.PerMarkExpands {
		if expand == "" {
			continue
		}

		perMark[name] = string(expand)
	}

	configuration := map[string]any{"perMarkExpands": perMark}
	if config.DefaultExpand != "" {
		configuration["defaultExpand"] = string(config.DefaultExpand)
	}

	configPayload, err := json.Marshal(configuration)
	if err != nil {
		return fmt.Errorf("cannot encode Automerge spans config: %w", err)
	}

	if err := t.document.engine.UpdateSpans(t.handle, spansPayload, configPayload); err != nil {
		return fmt.Errorf("cannot update Automerge spans: %w", err)
	}

	return nil
}

// Unmark removes a named annotation from a UTF-16 text range.
func (t *Text) Unmark(
	start uint32,
	end uint32,
	name string,
	expand MarkExpand,
) error {
	return t.Mark(

		start,
		end,
		name,
		Scalar{Type: ScalarTypeNull},
		expand,
	)
}

// SplitBlock inserts a block marker at a UTF-16 text position.
func (t *Text) SplitBlock(index uint32) (*Object, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	handle, err := t.document.engine.SplitBlock(t.handle, index)
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
func (t *Text) JoinBlock(index uint32) error {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return ErrClosed
	}

	if err := t.document.engine.JoinBlock(t.handle, index); err != nil {
		return fmt.Errorf("cannot join Automerge block: %w", err)
	}

	return nil
}

// ReplaceBlock replaces a block marker and returns its new map object.
func (t *Text) ReplaceBlock(index uint32) (*Object, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	handle, err := t.document.engine.ReplaceBlock(t.handle, index)
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
func (t *Text) Marks() ([]Mark, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	data, err := t.document.engine.Marks(t.handle)
	if err != nil {
		return nil, fmt.Errorf("cannot read Automerge marks: %w", err)
	}

	return decodeMarks(data)
}

// MarksAt returns the active marks over the text object at a historical frontier.
func (t *Text) MarksAt(heads []Hash) ([]Mark, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	data, err := t.document.engine.MarksAt(t.handle, engineHashes(heads))
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

func (t *Text) Spans() ([]Span, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	data, err := t.document.engine.TextSpans(t.handle)
	if err != nil {
		return nil, fmt.Errorf("cannot read Automerge rich-text spans: %w", err)
	}

	return decodeSpans(data)
}

// SpansAt returns the rich-text spans as they existed at a historical frontier.
func (t *Text) SpansAt(heads []Hash) ([]Span, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	data, err := t.document.engine.TextSpansAt(t.handle, engineHashes(heads))
	if err != nil {
		return nil, fmt.Errorf("cannot read historical Automerge rich-text spans: %w", err)
	}

	return decodeSpans(data)
}

func decodeSpans(data []byte) ([]Span, error) {
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

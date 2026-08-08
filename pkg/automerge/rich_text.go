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
)

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

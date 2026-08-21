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
	"fmt"
)

type CursorMove string

const (
	CursorMoveBefore CursorMove = "before"
	CursorMoveAfter  CursorMove = "after"
)

// StartCursor returns a cursor that always resolves to the sequence start.
func StartCursor() Cursor {
	return Cursor{1, 1}
}

// EndCursor returns a cursor that always resolves to the sequence end.
func EndCursor() Cursor {
	return Cursor{1, 2}
}

// CursorFor returns a stable cursor with JavaScript-compatible index clamping.
func (t *Text) CursorFor(
	ctx context.Context,
	index int64,
	move CursorMove,
) (Cursor, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	if move != CursorMoveBefore && move != CursorMoveAfter {
		return nil, fmt.Errorf("unknown Automerge cursor movement %q", move)
	}

	if index < 0 {
		return StartCursor(), nil
	}

	value, err := t.document.engine.Text(ctx, t.handle)
	if err != nil {
		return nil, fmt.Errorf("cannot read Automerge text for cursor: %w", err)
	}

	length := int64(utf16StringLength(value))
	if index >= length {
		return EndCursor(), nil
	}

	cursor, err := t.document.engine.TextCursorMoving(
		ctx,
		t.handle,
		uint32(index),
		move == CursorMoveBefore,
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create Automerge text cursor: %w", err)
	}

	return Cursor(cursor), nil
}

// CursorForAt returns a stable cursor for an index resolved against the text as
// it existed at a historical frontier, mirroring get_cursor with heads.
func (t *Text) CursorForAt(
	ctx context.Context,
	index int64,
	move CursorMove,
	heads []Hash,
) (Cursor, error) {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return nil, ErrClosed
	}

	if move != CursorMoveBefore && move != CursorMoveAfter {
		return nil, fmt.Errorf("unknown Automerge cursor movement %q", move)
	}

	if index < 0 {
		return StartCursor(), nil
	}

	value, err := t.document.engine.TextAt(ctx, t.handle, engineHashes(heads))
	if err != nil {
		return nil, fmt.Errorf("cannot read historical Automerge text for cursor: %w", err)
	}

	length := int64(utf16StringLength(value))
	if index >= length {
		return EndCursor(), nil
	}

	cursor, err := t.document.engine.TextCursorMovingAt(
		ctx,
		t.handle,
		uint32(index),
		move == CursorMoveBefore,
		engineHashes(heads),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot create historical Automerge text cursor: %w", err)
	}

	return Cursor(cursor), nil
}

// SpliceCursor resolves cursor and applies a text splice at its current position.
func (t *Text) SpliceCursor(
	ctx context.Context,
	cursor Cursor,
	deleteCount int32,
	value string,
) error {
	t.document.mu.Lock()
	defer t.document.mu.Unlock()

	if t.document.closed {
		return ErrClosed
	}

	position, err := t.document.engine.TextCursorPosition(
		ctx,
		t.handle,
		cursor,
	)
	if err != nil {
		return fmt.Errorf("cannot resolve Automerge text cursor: %w", err)
	}

	if err := t.document.engine.SpliceText(
		ctx,
		t.handle,
		position,
		deleteCount,
		value,
	); err != nil {
		return fmt.Errorf("cannot splice Automerge text at cursor: %w", err)
	}

	return nil
}

func utf16StringLength(value string) uint32 {
	var length uint32

	for _, character := range value {
		if character > 0xffff {
			length += 2
		} else {
			length++
		}
	}

	return length
}

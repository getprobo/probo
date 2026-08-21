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

import * as Automerge from "@automerge/automerge";

// A collaborator's caret or selection carried in automerge-repo presence,
// expressed with stable Automerge text cursors rather than integer offsets. An
// offset is invalidated by any concurrent edit before it, so a remote caret
// drawn from an offset drifts onto the wrong character; a cursor resolves to the
// position of the same character after arbitrary concurrent edits, which is what
// keeps remote carets anchored while other people type.
//
// The shape (field, anchor, head) mirrors the Go TextSelectionValue in
// pkg/automerge/collaboration so the two describe the same thing. The cursor
// values here are the JavaScript Automerge cursor type (a string), which is the
// representation browser peers exchange; cross-runtime cursor interop with Go
// agents is a separate concern because the two runtimes encode cursors
// differently.
export type TextSelection = {
  // The Automerge map key of the text object the selection addresses.
  field: string;
  // The stable cursor for the fixed end of the selection.
  anchor: Automerge.Cursor;
  // The stable cursor for the moving end (the caret). When it equals the anchor
  // the selection is a collapsed caret.
  head: Automerge.Cursor;
};

// A selection resolved back to UTF-16 offsets in the current document, ready for
// ProseMirror decorations.
export type ResolvedSelection = {
  anchor: number;
  head: number;
};

// textSelection builds a stable selection from the current caret offsets in a
// text field. move controls which side of a character a collapsed caret anchors
// to; "before" keeps it in front of the following character, matching a caret
// that stays put as text is inserted after it.
export function textSelection<T>(
  document: Automerge.Doc<T>,
  field: string,
  anchorOffset: number,
  headOffset: number,
  move: Automerge.MoveCursor = "before",
): TextSelection {
  return {
    field,
    anchor: Automerge.getCursor(document, [field], anchorOffset, move),
    head: Automerge.getCursor(document, [field], headOffset, move),
  };
}

// resolveSelection resolves a selection's cursors to offsets in the given
// document, which may have advanced since the selection was created.
export function resolveSelection<T>(
  document: Automerge.Doc<T>,
  selection: TextSelection,
): ResolvedSelection {
  return {
    anchor: Automerge.getCursorPosition(
      document,
      [selection.field],
      selection.anchor,
    ),
    head: Automerge.getCursorPosition(
      document,
      [selection.field],
      selection.head,
    ),
  };
}

// isCollapsed reports whether a selection is a single caret.
export function isCollapsed(selection: TextSelection): boolean {
  return selection.anchor === selection.head;
}

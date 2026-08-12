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
import type { SchemaAdapter } from "@automerge/prosemirror";
// These position-mapping helpers are not re-exported from the package index, so
// they are imported from the pinned build. They convert between ProseMirror
// document positions and Automerge text offsets, which is what lets a caret
// captured in the editor be stored as a stable Automerge cursor and drawn back
// at the right place after concurrent edits.
import {
  amSpliceIdxToPmIdx,
  pmRangeToAmRange,
} from "@automerge/prosemirror/dist/traversal.js";

import {
  createSchemaAdapter,
  type RichEditorAutomergeDocument,
} from "./collaboration";
import { richEditorCollaborationExtensions } from "./RichEditor";
import {
  resolveSelection,
  type TextSelection,
  textSelection,
} from "./repoSelection";

// richEditorPresenceAdapter builds the schema adapter used to map positions
// between ProseMirror and Automerge for presence. It matches the schema the
// collaborative editor uses, so positions line up.
export function richEditorPresenceAdapter(): SchemaAdapter {
  return createSchemaAdapter(richEditorCollaborationExtensions);
}

// The map key of the rich-text field and the path used to read its spans.
const textField = "body";
const textPath: Automerge.Prop[] = [textField];

// A collaborator's caret/selection in ProseMirror position space, ready for the
// presence decorations.
export type PmPresenceSelection = {
  anchorPosition: number;
  headPosition: number;
};

// presenceFromPmSelection converts a ProseMirror caret/selection into a stable
// Automerge-cursor selection to publish over presence. It returns null when a
// position cannot be mapped (for example a selection on a structural node), so
// the caller can skip publishing rather than send a bad selection.
export function presenceFromPmSelection(
  adapter: SchemaAdapter,
  document: Automerge.Doc<RichEditorAutomergeDocument>,
  anchorPosition: number,
  headPosition: number,
): TextSelection | null {
  const spans = Automerge.spans(document, textPath);

  const anchorOffset = pmPositionToAmOffset(adapter, spans, anchorPosition);
  const headOffset = pmPositionToAmOffset(adapter, spans, headPosition);
  if (anchorOffset === null || headOffset === null) {
    return null;
  }

  return textSelection(document, textField, anchorOffset, headOffset);
}

// pmSelectionFromPresence resolves a presence selection's stable cursors against
// the current document and maps them back to ProseMirror positions. It returns
// null when a cursor no longer resolves (for example the surrounding text was
// deleted).
export function pmSelectionFromPresence(
  adapter: SchemaAdapter,
  document: Automerge.Doc<RichEditorAutomergeDocument>,
  selection: TextSelection,
): PmPresenceSelection | null {
  const resolved = resolveSelection(document, selection);
  const spans = Automerge.spans(document, textPath);

  const anchorPosition = amSpliceIdxToPmIdx(adapter, spans, resolved.anchor);
  const headPosition = amSpliceIdxToPmIdx(adapter, spans, resolved.head);
  if (anchorPosition === null || headPosition === null) {
    return null;
  }

  return { anchorPosition, headPosition };
}

function pmPositionToAmOffset(
  adapter: SchemaAdapter,
  spans: Automerge.Span[],
  position: number,
): number | null {
  const range = pmRangeToAmRange(adapter, spans, {
    from: position,
    to: position,
  });

  return range ? range.start : null;
}

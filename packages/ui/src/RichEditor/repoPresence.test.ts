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
import { describe, expect, it } from "vitest";

import {
  createSchemaAdapter,
  type RichEditorAutomergeDocument,
} from "./collaboration";
import { richEditorCollaborationExtensions } from "./RichEditor";
import {
  pmSelectionFromPresence,
  presenceFromPmSelection,
} from "./repoPresence";

function seededDocument(
  text: string,
): Automerge.Doc<RichEditorAutomergeDocument> {
  const document = Automerge.from<RichEditorAutomergeDocument>({ body: "" });
  return Automerge.change(document, (draft) => {
    Automerge.splice(draft, ["body"], 0, 0, text);
  });
}

describe("repo presence mapping", () => {
  it("round-trips a caret between ProseMirror and Automerge cursors", () => {
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    const document = seededDocument("hello world");

    // In a single paragraph, ProseMirror position 1 is the start of the text, so
    // the "w" of "world" (text offset 6) is at position 7.
    const selection = presenceFromPmSelection(adapter, document, 7, 7);
    expect(selection).not.toBeNull();

    const resolved = pmSelectionFromPresence(adapter, document, selection!);
    expect(resolved).toEqual({ anchorPosition: 7, headPosition: 7 });
  });

  it("keeps a remote caret anchored across a concurrent insertion", () => {
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    let document = seededDocument("hello world");

    const selection = presenceFromPmSelection(adapter, document, 7, 7);
    expect(selection).not.toBeNull();

    // Someone types three characters at the start of the text.
    document = Automerge.change(document, (draft) => {
      Automerge.splice(draft, ["body"], 0, 0, "XX ");
    });

    // The same stable selection now maps three positions later; a stored
    // ProseMirror position of 7 would point at the wrong character.
    const resolved = pmSelectionFromPresence(adapter, document, selection!);
    expect(resolved).toEqual({ anchorPosition: 10, headPosition: 10 });
  });

  it("carries a selection range with distinct endpoints", () => {
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    const document = seededDocument("hello world");

    // "hello" spans text offsets 0..5, i.e. ProseMirror positions 1..6.
    const selection = presenceFromPmSelection(adapter, document, 1, 6);
    expect(selection).not.toBeNull();

    const resolved = pmSelectionFromPresence(adapter, document, selection!);
    expect(resolved).toEqual({ anchorPosition: 1, headPosition: 6 });
  });
});

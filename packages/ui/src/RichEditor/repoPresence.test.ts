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
import {
  pmSelectionFromPresence,
  presenceFromPmSelection,
} from "./repoPresence";
import { richEditorCollaborationExtensions } from "./RichEditor";

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

  it("places a caret on the typed character once an edit that beat it syncs", () => {
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    const before = seededDocument("hello");
    const author = Automerge.clone(before, {
      actor: "02000000000000000000000000000000",
    });
    const edited = Automerge.change(author, (draft) => {
      Automerge.splice(draft, ["body"], 3, 0, "X");
    });

    // Caret sits after the inserted "X" in "helXlo" (text offset 4, position 5).
    const selection = presenceFromPmSelection(adapter, edited, 5, 5);
    expect(selection).not.toBeNull();
    expect(pmSelectionFromPresence(adapter, edited, selection!)).toEqual({
      anchorPosition: 5,
      headPosition: 5,
    });

    // Resolved against the document that does not contain "X" yet, the same
    // cursor sits where the caret was before that edit.
    expect(pmSelectionFromPresence(adapter, before, selection!)).toEqual({
      anchorPosition: 4,
      headPosition: 4,
    });

    const merged = Automerge.merge(Automerge.clone(before), edited);
    expect(pmSelectionFromPresence(adapter, merged, selection!)).toEqual({
      anchorPosition: 5,
      headPosition: 5,
    });
  });

  it("keeps a remote caret anchored across a concurrent insertion", () => {
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    const document = seededDocument("hello world");

    const selection = presenceFromPmSelection(adapter, document, 7, 7);
    expect(selection).not.toBeNull();

    const other = Automerge.clone(document, {
      actor: "02000000000000000000000000000000",
    });
    const edited = Automerge.change(other, (draft) => {
      Automerge.splice(draft, ["body"], 0, 0, "XX ");
    });
    const merged = Automerge.merge(document, edited);

    const resolved = pmSelectionFromPresence(adapter, merged, selection!);
    expect(resolved).toEqual({ anchorPosition: 10, headPosition: 10 });
  });

  it("ignores a selection aimed at an unknown field", () => {
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    const document = seededDocument("hello world");

    expect(pmSelectionFromPresence(adapter, document, {
      field: "missing",
      anchor: "cursor",
      head: "cursor",
    })).toBeNull();
  });

  it("ignores a cursor the document cannot resolve", () => {
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    const document = seededDocument("hello world");

    expect(pmSelectionFromPresence(adapter, document, {
      field: "body",
      anchor: "not-a-cursor",
      head: "not-a-cursor",
    })).toBeNull();
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

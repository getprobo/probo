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
import type { DocHandle } from "@automerge/prosemirror";
import { pmDocFromSpans } from "@automerge/prosemirror";
import { EditorState, type Transaction } from "@tiptap/pm/state";
import type { EditorView } from "@tiptap/pm/view";
import { describe, expect, it } from "vitest";

import {
  createAutomergeSyncPlugin,
  isSurrogatePair,
} from "./AutomergeSyncPlugin";
import { createSchemaAdapter, type RichEditorAutomergeDocument } from "./collaboration";
import { createRichEditorAutomergeDocument } from "./CollaborativeRichEditor";
import { richEditorCollaborationExtensions } from "./RichEditor";

type RemoteChange = (payload: {
  doc: Automerge.Doc<RichEditorAutomergeDocument>;
  patches: Automerge.Patch[];
  patchInfo: { before: Automerge.Doc<RichEditorAutomergeDocument> };
}) => void;

describe("Automerge sync plugin", () => {
  it("reconciles the first remote divider instead of using text offsets", () => {
    const local = createRichEditorAutomergeDocument(JSON.stringify({
      type: "doc",
      content: [
        { type: "paragraph", content: [{ type: "text", text: "Before" }] },
        { type: "paragraph", content: [{ type: "text", text: "After" }] },
      ],
    }));
    const remote = createRichEditorAutomergeDocument(JSON.stringify({
      type: "doc",
      content: [
        { type: "paragraph", content: [{ type: "text", text: "Before" }] },
        { type: "horizontalRule" },
        { type: "paragraph", content: [{ type: "text", text: "After" }] },
      ],
    }));
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    let onChange: RemoteChange | undefined;
    const handle = {
      doc: () => local,
      change: () => {},
      on: (_event: string, callback: RemoteChange) => {
        onChange = callback;
      },
      off: () => {},
    } as unknown as DocHandle<RichEditorAutomergeDocument>;
    let state = EditorState.create({
      schema: adapter.schema,
      doc: pmDocFromSpans(adapter, Automerge.spans(local, ["body"])),
      plugins: [createAutomergeSyncPlugin(adapter, handle, ["body"])],
    });
    expect(state.doc.textContent).toBe("BeforeAfter");
    expect(nodeNames(state.doc)).not.toContain("horizontalRule");

    const plugin = state.plugins.find(item => item.spec.view);
    const start = plugin?.spec.view;
    if (!start) throw new Error("expected sync view");
    const view = {
      get state() {
        return state;
      },
      dispatch(transaction: Transaction) {
        state = state.apply(transaction);
      },
    } as EditorView;
    start(view);
    if (!onChange) throw new Error("expected remote change handler");

    onChange({
      doc: remote,
      patches: [],
      patchInfo: { before: local },
    });

    expect(nodeNames(state.doc)).toContain("horizontalRule");
    expect(state.doc.textContent).toBe("BeforeAfter");
  });

  it("reconciles a remote image and keeps a local upload placeholder", () => {
    let local = createRichEditorAutomergeDocument(JSON.stringify({
      type: "doc",
      content: [
        { type: "paragraph", content: [{ type: "text", text: "Before" }] },
        { type: "paragraph", content: [{ type: "text", text: "After" }] },
      ],
    }));
    const remote = createRichEditorAutomergeDocument(JSON.stringify({
      type: "doc",
      content: [
        { type: "paragraph", content: [{ type: "text", text: "Before" }] },
        { type: "image", attrs: { alt: "diagram", fileId: "file-1" } },
        { type: "paragraph", content: [{ type: "text", text: "After" }] },
      ],
    }));
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    let onChange: RemoteChange | undefined;
    const handle = {
      doc: () => local,
      change: (callback: (draft: RichEditorAutomergeDocument) => void) => {
        local = Automerge.change(local, callback);
      },
      on: (_event: string, callback: RemoteChange) => {
        onChange = callback;
      },
      off: () => {},
    } as unknown as DocHandle<RichEditorAutomergeDocument>;
    let state = EditorState.create({
      schema: adapter.schema,
      doc: pmDocFromSpans(adapter, Automerge.spans(local, ["body"])),
      plugins: [createAutomergeSyncPlugin(adapter, handle, ["body"])],
    });
    const plugin = state.plugins.find(item => item.spec.view);
    const start = plugin?.spec.view;
    if (!start) throw new Error("expected sync view");
    const view = {
      get state() {
        return state;
      },
      dispatch(transaction: Transaction) {
        state = state.apply(transaction);
      },
    } as EditorView;
    start(view);

    const upload = state.schema.nodes.contentUpload.create({
      id: "upload-1",
      fileName: "scan.png",
    });
    const firstParagraph = state.doc.firstChild;
    if (!firstParagraph) throw new Error("expected first paragraph");
    state = state.apply(state.tr.insert(firstParagraph.nodeSize, upload));
    expect(nodeNames(state.doc)).toContain("contentUpload");
    expect(nodeNames(state.doc)).not.toContain("image");
    expect(blockTypes(local)).toEqual(["paragraph", "paragraph"]);

    if (!onChange) throw new Error("expected remote change handler");
    onChange({
      doc: remote,
      patches: [],
      patchInfo: { before: local },
    });

    expect(nodeNames(state.doc)).toContain("image");
    expect(nodeNames(state.doc)).toContain("contentUpload");
    expect(state.doc.textContent).toBe("BeforeAfter");
    let fileId = "";
    state.doc.descendants((node) => {
      if (node.type.name === "image") {
        fileId = String(node.attrs.fileId);
        return false;
      }
      return true;
    });
    expect(fileId).toBe("file-1");
  });

  it("keeps syncing after a local Automerge change throws", () => {
    let document = createRichEditorAutomergeDocument(JSON.stringify({
      type: "doc",
      content: [{
        type: "paragraph",
        content: [{ type: "text", text: "Hello" }],
      }],
    }));
    let fail = true;
    const handle = {
      doc: () => document,
      change: (callback: (draft: RichEditorAutomergeDocument) => void) => {
        if (fail) throw new Error("change failed");
        document = Automerge.change(document, callback);
      },
      on: () => {},
      off: () => {},
    } as unknown as DocHandle<RichEditorAutomergeDocument>;
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    const state = EditorState.create({
      schema: adapter.schema,
      doc: pmDocFromSpans(adapter, Automerge.spans(document, ["body"])),
      plugins: [createAutomergeSyncPlugin(adapter, handle, ["body"])],
    });
    const position = textPosition(state, "Hello");

    expect(() => {
      state.apply(state.tr.insertText("!", position + 1));
    }).toThrow(/change failed/);

    fail = false;
    state.apply(state.tr.insertText("!", position + 1));
    expect(document.body).toContain("H!ello");
  });

  it("keeps syncing after a remote dispatch throws", () => {
    let document = createRichEditorAutomergeDocument(JSON.stringify({
      type: "doc",
      content: [{
        type: "paragraph",
        content: [{ type: "text", text: "Hello" }],
      }],
    }));
    let onChange: RemoteChange | undefined;
    const handle = {
      doc: () => document,
      change: (callback: (draft: RichEditorAutomergeDocument) => void) => {
        document = Automerge.change(document, callback);
      },
      on: (_event: string, callback: RemoteChange) => {
        onChange = callback;
      },
      off: () => {},
    } as unknown as DocHandle<RichEditorAutomergeDocument>;
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    const state = EditorState.create({
      schema: adapter.schema,
      doc: pmDocFromSpans(adapter, Automerge.spans(document, ["body"])),
      plugins: [createAutomergeSyncPlugin(adapter, handle, ["body"])],
    });
    const plugin = state.plugins.find(item => item.spec.view);
    const start = plugin?.spec.view;
    if (!start) throw new Error("expected sync view");
    start({
      state,
      dispatch() {
        throw new Error("dispatch failed");
      },
    } as unknown as EditorView);
    if (!onChange) throw new Error("expected remote change handler");

    expect(() => onChange?.({
      doc: document,
      patches: [],
      patchInfo: { before: document },
    })).toThrow(/dispatch failed/);

    const position = textPosition(state, "Hello");
    state.apply(state.tr.insertText("!", position + 1));
    expect(document.body).toContain("H!ello");
  });
});

describe("surrogate pairs", () => {
  it("detects a UTF-16 pair and rejects reversed halves", () => {
    expect(isSurrogatePair("😀")).toBe(true);
    expect(isSurrogatePair("\uDE00\uD83D")).toBe(false);
    expect(isSurrogatePair("ab")).toBe(false);
    expect(isSurrogatePair("a")).toBe(false);
  });
});

function blockTypes(document: Automerge.Doc<RichEditorAutomergeDocument>): string[] {
  return Automerge.spans(document, ["body"])
    .filter(span => span.type === "block")
    .map((span) => {
      const type = span.value.type;
      if (typeof type === "string") return type;
      if (Automerge.isImmutableString(type)) return type.val;
      throw new Error("expected block type");
    });
}

function nodeNames(doc: { descendants: (callback: (node: { type: { name: string } }) => void) => void }): string[] {
  const names: string[] = [];
  doc.descendants((node) => {
    names.push(node.type.name);
  });
  return names;
}

function textPosition(state: EditorState, text: string): number {
  let position: number | undefined;
  state.doc.descendants((node, at) => {
    if (node.isText && node.text === text) {
      position = at;
      return false;
    }
    return true;
  });
  if (position === undefined) throw new Error(`expected text ${text}`);
  return position;
}

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
import { pmDocFromSpans, pmNodeToSpans } from "@automerge/prosemirror";
import { baseKeymap, joinBackward, joinForward } from "@tiptap/pm/commands";
import { EditorState, TextSelection, type Transaction } from "@tiptap/pm/state";
import type { EditorView } from "@tiptap/pm/view";
import { describe, expect, it, vi } from "vitest";

let rejectSpanRewrite = false;

vi.mock("@automerge/automerge", async () => {
  const actual = await vi.importActual<typeof import("@automerge/automerge")>(
    "@automerge/automerge",
  );
  return {
    ...actual,
    updateSpans: (...args: Parameters<typeof actual.updateSpans>) => {
      if (rejectSpanRewrite) throw new Error("full span rewrite");
      return actual.updateSpans(...args);
    },
  };
});

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

  it("deletes a character inside a paragraph", () => {
    let document = createRichEditorAutomergeDocument(JSON.stringify({
      type: "doc",
      content: [{
        type: "paragraph",
        content: [{ type: "text", text: "Hello" }],
      }],
    }));
    const handle = {
      doc: () => document,
      change: (callback: (draft: RichEditorAutomergeDocument) => void) => {
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
    state.apply(state.tr.delete(position, position + 1));
    expect(document.body).toContain("ello");
    expect(document.body).not.toContain("Hello");
  });

  it("toggles bold without rewriting the rest of the paragraph", () => {
    let document = createRichEditorAutomergeDocument(JSON.stringify({
      type: "doc",
      content: [{
        type: "paragraph",
        content: [{ type: "text", text: "Hello" }],
      }],
    }));
    const handle = {
      doc: () => document,
      change: (callback: (draft: RichEditorAutomergeDocument) => void) => {
        document = Automerge.change(document, callback);
      },
      on: () => {},
      off: () => {},
    } as unknown as DocHandle<RichEditorAutomergeDocument>;
    const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
    let state = EditorState.create({
      schema: adapter.schema,
      doc: pmDocFromSpans(adapter, Automerge.spans(document, ["body"])),
      plugins: [createAutomergeSyncPlugin(adapter, handle, ["body"])],
    });
    const position = textPosition(state, "Hello");
    const mark = state.schema.marks.bold?.create();
    if (!mark) throw new Error("expected a bold mark");
    state = state.apply(state.tr.addMark(position, position + 2, mark));
    expect(document.body).toContain("Hello");
    const spans = Automerge.spans(document, ["body"]);
    const text = spans
      .filter(span => span.type === "text")
      .map(span => String(span.value))
      .join("");
    expect(text).toBe("Hello");
    const bold = spans.find(span => span.type === "text" && span.value === "He");
    expect(bold && "marks" in bold ? bold.marks : undefined).toMatchObject({ strong: true });

    state.apply(state.tr.removeMark(position, position + 2, mark));
    const cleared = Automerge.spans(document, ["body"])
      .filter(span => span.type === "text")
      .map(span => String(span.value))
      .join("");
    expect(cleared).toBe("Hello");
  });

  it("splits a paragraph without rewriting the document", () => {
    const editor = openDocument({
      type: "doc",
      content: [{
        type: "paragraph",
        content: [{ type: "text", text: "Hello" }],
      }],
    });
    const start = textPosition(editor.state, "Hello");
    editor.select(start + 3);
    const before = Automerge.clone(editor.document);

    expectNoSpanRewrite(() => {
      editor.apply(pressEnter(editor.state));
    });

    expect(editor.state.doc.childCount).toBe(2);
    expect(editor.state.doc.child(0).textContent).toBe("Hel");
    expect(editor.state.doc.child(1).textContent).toBe("lo");
    expectSameDocument(editor);

    const change = Automerge.getLastLocalChange(editor.document);
    if (!change) throw new Error("expected a split change");
    const [remote] = Automerge.applyChanges(before, [change]);
    expect(pmDocFromSpans(editor.adapter, Automerge.spans(remote, ["body"])).toJSON())
      .toEqual(canonicalDocument(editor));
  });

  it("splits the end of a paragraph and an empty paragraph", () => {
    const editor = openDocument({
      type: "doc",
      content: [{
        type: "paragraph",
        content: [{ type: "text", text: "Hello" }],
      }],
    });
    const start = textPosition(editor.state, "Hello");
    editor.select(start + "Hello".length);
    expectNoSpanRewrite(() => {
      editor.apply(pressEnter(editor.state));
    });
    expect(paragraphs(editor.state.doc)).toEqual(["Hello", ""]);
    expectSameDocument(editor);

    expectNoSpanRewrite(() => {
      editor.apply(pressEnter(editor.state));
    });
    expect(paragraphs(editor.state.doc)).toEqual(["Hello", "", ""]);
    expectSameDocument(editor);
  });

  it("splits a heading into a heading and a paragraph", () => {
    const editor = openDocument({
      type: "doc",
      content: [{
        type: "heading",
        attrs: { level: 2 },
        content: [{ type: "text", text: "Title" }],
      }],
    });
    const start = textPosition(editor.state, "Title");
    editor.select(start + 2);
    expectNoSpanRewrite(() => {
      editor.apply(pressEnter(editor.state));
    });
    expect(editor.state.doc.child(0).type.name).toBe("heading");
    expect(editor.state.doc.child(0).attrs.level).toBe(2);
    expect(editor.state.doc.child(1).type.name).toBe("heading");
    expect(paragraphs(editor.state.doc)).toEqual(["Ti", "tle"]);
    expectSameDocument(editor);

    editor.select(textPosition(editor.state, "tle") + "tle".length);
    expectNoSpanRewrite(() => {
      editor.apply(pressEnter(editor.state));
    });
    expect(editor.state.doc.child(2).type.name).toBe("paragraph");
    expect(paragraphs(editor.state.doc)).toEqual(["Ti", "tle", ""]);
    expectSameDocument(editor);
  });

  it("splits a paragraph inside a list item", () => {
    const editor = openDocument({
      type: "doc",
      content: [{
        type: "bulletList",
        content: [{
          type: "listItem",
          content: [{
            type: "paragraph",
            content: [{ type: "text", text: "Item" }],
          }],
        }],
      }],
    });
    const start = textPosition(editor.state, "Item");
    editor.select(start + 2);
    expectNoSpanRewrite(() => {
      editor.apply(pressEnter(editor.state));
    });
    expect(editor.state.doc.child(0).child(0).childCount).toBe(2);
    expect(paragraphs(editor.state.doc)).toEqual(["It", "em"]);
    expectSameDocument(editor);
  });

  it("keeps bold text when a paragraph splits", () => {
    const editor = openDocument({
      type: "doc",
      content: [{
        type: "paragraph",
        content: [{ type: "text", text: "Hello" }],
      }],
    });
    const start = textPosition(editor.state, "Hello");
    const mark = editor.state.schema.marks.bold?.create();
    if (!mark) throw new Error("expected a bold mark");
    editor.apply(editor.state.tr.addMark(start, start + 2, mark));
    editor.select(start + 2);
    expectNoSpanRewrite(() => {
      editor.apply(pressEnter(editor.state));
    });
    expect(paragraphs(editor.state.doc)).toEqual(["He", "llo"]);
    expectSameDocument(editor);
  });

  it("stores a suggestion split on the new block", () => {
    const editor = openDocument({
      type: "doc",
      content: [{
        type: "paragraph",
        content: [{ type: "text", text: "Hello" }],
      }],
    });
    const start = textPosition(editor.state, "Hello");
    editor.select(start + "Hello".length);
    const enter = pressEnter(editor.state);
    enter.setMeta("suggestion-split", { id: "s1", kind: "insertion" });
    expectNoSpanRewrite(() => {
      editor.apply(enter);
    });
    const blocks = Automerge.spans(editor.document, ["body"])
      .filter(span => span.type === "block");
    expect(blocks.at(-1)?.value).toMatchObject({
      attrs: { suggestionSplit: "s1", splitKind: "insertion" },
    });

    const remap = editor.state.tr.setMeta("suggestion-reassign", {
      from: "s1",
      to: "server",
    });
    editor.apply(remap);
    const remapped = Automerge.spans(editor.document, ["body"])
      .filter(span => span.type === "block");
    expect(remapped.at(-1)?.value).toMatchObject({
      attrs: { suggestionSplit: "server", splitKind: "insertion" },
    });
  });

  it("keeps a suggestion split and a block flag off the full rewrite", () => {
    const editor = openDocument({
      type: "doc",
      content: [{
        type: "paragraph",
        content: [{ type: "text", text: "Hello" }],
      }],
    });
    const start = textPosition(editor.state, "Hello");
    editor.select(start + "Hello".length);
    const enter = pressEnter(editor.state);
    const blockPos = enter.selection.$from.before();
    const block = enter.doc.nodeAt(blockPos);
    if (!block) throw new Error("expected the new paragraph");
    enter.setNodeMarkup(blockPos, undefined, {
      ...block.attrs,
      unknownAttrs: { suggestionSplit: "s1", splitKind: "insertion" },
    });
    enter.setMeta("suggestion-split", { id: "s1", kind: "insertion" });
    expectNoSpanRewrite(() => {
      editor.apply(enter);
    });
    expect(paragraphs(editor.state.doc)).toEqual(["Hello", ""]);

    const paragraph = editor.state.doc.firstChild;
    if (!paragraph) throw new Error("expected a paragraph");
    const flagged = editor.state.tr.setNodeMarkup(0, undefined, {
      ...paragraph.attrs,
      isAmgBlock: true,
    });
    expectNoSpanRewrite(() => {
      editor.apply(flagged);
    });
    expectSameDocument(editor);
  });

  it("joins a paragraph with backspace without rewriting every span", () => {
    const editor = twoParagraphs();
    editor.select(textPosition(editor.state, "World"));
    expectNoSpanRewrite(() => {
      editor.apply(pressJoin(editor.state, joinBackward));
    });
    expectJoinedParagraph(editor);
  });

  it("joins a paragraph with delete without rewriting every span", () => {
    const editor = twoParagraphs();
    const start = textPosition(editor.state, "Hello");
    editor.select(start + "Hello".length);
    expectNoSpanRewrite(() => {
      editor.apply(pressJoin(editor.state, joinForward));
    });
    expectJoinedParagraph(editor);
  });

  it("returns a split paragraph to one block", () => {
    const editor = openDocument({
      type: "doc",
      content: [{
        type: "paragraph",
        content: [{ type: "text", text: "Hello" }],
      }],
    });
    const start = textPosition(editor.state, "Hello");
    editor.select(start + 2);
    expectNoSpanRewrite(() => {
      editor.apply(pressEnter(editor.state));
      editor.apply(pressJoin(editor.state, joinBackward));
    });
    expect(paragraphs(editor.state.doc)).toEqual(["Hello"]);
    expectSameDocument(editor);
  });

  it("removes an empty paragraph without rewriting every span", () => {
    const editor = openDocument({
      type: "doc",
      content: [
        { type: "paragraph", content: [{ type: "text", text: "Hello" }] },
        { type: "paragraph" },
        { type: "paragraph", content: [{ type: "text", text: "World" }] },
      ],
    });
    editor.select(textPosition(editor.state, "World"));
    expectNoSpanRewrite(() => {
      editor.apply(pressJoin(editor.state, joinBackward));
    });
    expect(paragraphs(editor.state.doc)).toEqual(["Hello", "World"]);
    expectSameDocument(editor);
  });

  it("removes a leading empty paragraph without rewriting every span", () => {
    const editor = openDocument({
      type: "doc",
      content: [
        { type: "paragraph" },
        { type: "paragraph", content: [{ type: "text", text: "Hello" }] },
      ],
    });
    editor.select(textPosition(editor.state, "Hello"));
    expectNoSpanRewrite(() => {
      editor.apply(pressJoin(editor.state, joinBackward));
    });
    expect(paragraphs(editor.state.doc)).toEqual(["Hello"]);
    const spans = Automerge.spans(editor.document, ["body"]);
    const text = spans
      .filter(span => span.type === "text")
      .map(span => String(span.value))
      .join("");
    expect(text).toBe("Hello");
    expect(spans.some(span => span.type === "block")).toBe(false);
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

function nodeNames(doc: { descendants: (callback: (node: { type: { name: string } }) => void) => void }): string[] {
  const names: string[] = [];
  doc.descendants((node) => {
    names.push(node.type.name);
  });
  return names;
}

function openDocument(content: object) {
  let document = createRichEditorAutomergeDocument(JSON.stringify(content));
  const handle = {
    doc: () => document,
    change: (callback: (draft: RichEditorAutomergeDocument) => void) => {
      document = Automerge.change(document, callback);
    },
    on: () => {},
    off: () => {},
  } as unknown as DocHandle<RichEditorAutomergeDocument>;
  const adapter = createSchemaAdapter(richEditorCollaborationExtensions);
  let state = EditorState.create({
    schema: adapter.schema,
    doc: pmDocFromSpans(adapter, Automerge.spans(document, ["body"])),
    plugins: [createAutomergeSyncPlugin(adapter, handle, ["body"])],
  });
  return {
    adapter,
    get document() {
      return document;
    },
    get state() {
      return state;
    },
    apply(transaction: Transaction) {
      state = state.apply(transaction);
    },
    select(position: number) {
      state = state.apply(
        state.tr.setSelection(TextSelection.create(state.doc, position)),
      );
    },
  };
}

function twoParagraphs() {
  return openDocument({
    type: "doc",
    content: [
      { type: "paragraph", content: [{ type: "text", text: "Hello" }] },
      { type: "paragraph", content: [{ type: "text", text: "World" }] },
    ],
  });
}

function pressEnter(state: EditorState): Transaction {
  const enter = baseKeymap.Enter;
  if (!enter) throw new Error("expected an enter binding");
  let transaction = state.tr;
  const dispatched = enter(state, (next) => {
    transaction = next;
  });
  if (!dispatched) throw new Error("enter did not split the block");
  return transaction;
}

function pressJoin(
  state: EditorState,
  command: typeof joinBackward,
): Transaction {
  let transaction = state.tr;
  const dispatched = command(state, (next) => {
    transaction = next;
  });
  if (!dispatched) throw new Error("join did not remove the boundary");
  return transaction;
}

function expectNoSpanRewrite(run: () => void) {
  rejectSpanRewrite = true;
  try {
    run();
  } finally {
    rejectSpanRewrite = false;
  }
}

function expectJoinedParagraph(editor: {
  document: Automerge.Doc<RichEditorAutomergeDocument>;
  state: EditorState;
}) {
  expect(paragraphs(editor.state.doc)).toEqual(["HelloWorld"]);
  const spans = Automerge.spans(editor.document, ["body"]);
  const text = spans
    .filter(span => span.type === "text")
    .map(span => String(span.value))
    .join("");
  expect(text).toBe("HelloWorld");
  expect(spans.some(span => span.type === "block")).toBe(false);
}

function expectSameDocument(editor: {
  adapter: ReturnType<typeof createSchemaAdapter>;
  document: Automerge.Doc<RichEditorAutomergeDocument>;
  state: EditorState;
}) {
  expect(pmDocFromSpans(
    editor.adapter,
    Automerge.spans(editor.document, ["body"]),
  ).toJSON()).toEqual(canonicalDocument(editor));
}

function canonicalDocument(editor: {
  adapter: ReturnType<typeof createSchemaAdapter>;
  state: EditorState;
}) {
  return pmDocFromSpans(
    editor.adapter,
    pmNodeToSpans(editor.adapter, editor.state.doc),
  ).toJSON();
}

function paragraphs(doc: {
  descendants: (callback: (node: { isTextblock: boolean; textContent: string }) => void) => void;
}): string[] {
  const text: string[] = [];
  doc.descendants((node) => {
    if (node.isTextblock) text.push(node.textContent);
  });
  return text;
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

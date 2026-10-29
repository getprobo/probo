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

import { getSchema } from "@tiptap/core";
import { baseKeymap, joinBackward } from "@tiptap/pm/commands";
import { keymap } from "@tiptap/pm/keymap";
import { type Node as ProseMirrorNode } from "@tiptap/pm/model";
import { EditorState, TextSelection, type Transaction } from "@tiptap/pm/state";
import type { EditorView } from "@tiptap/pm/view";
import { describe, expect, it } from "vitest";

import { richEditorCollaborationExtensions } from "./RichEditor";
import {
  createSuggestingExtension,
  reassignSuggestionId,
  suggestionDeleteName,
  suggestionInputTransaction,
  suggestionInsertName,
  type SuggestingController,
} from "./SuggestionExtension";

const schema = getSchema(richEditorCollaborationExtensions);
const suggestionId = "suggestion-1";

describe("suggesting input", () => {
  it("lets enter split the paragraph without ending the suggestion", () => {
    let finished = false;
    const controller: SuggestingController = {
      enabled: true,
      suggestionId: suggestionId,
      pending: null,
      request: null,
      onNeedId: () => Promise.resolve(suggestionId),
      onBurstFinished: () => {
        finished = true;
      },
      onOpenSuggestion: () => {},
    };
    const extension = createSuggestingExtension({ current: controller });
    const plugins = extension.config.addProseMirrorPlugins?.call(extension) ?? [];
    const suggesting = plugins[0];
    if (!suggesting?.props.handleKeyDown) {
      throw new Error("expected a suggesting key handler");
    }

    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [schema.text("Hello")]),
    ]);
    let state = EditorState.create({
      schema,
      doc,
      selection: TextSelection.create(doc, 6),
      plugins: [...plugins, keymap(baseKeymap)],
    });
    const view = {
      get state() {
        return state;
      },
      dispatch(transaction: Transaction) {
        state = state.apply(transaction);
      },
    } as EditorView;

    expect(suggesting.props.handleKeyDown.call(suggesting, view, { key: "Enter" } as KeyboardEvent)).toBe(true);
    expect(finished).toBe(false);
    expect(state.doc.childCount).toBe(2);
    expect(state.doc.child(0).textContent).toBe("Hello");
    expect(state.doc.child(1).textContent).toBe("");
    expect(state.doc.child(1).attrs.unknownAttrs).toMatchObject({
      suggestionSplit: suggestionId,
      splitKind: "insertion",
    });

    expect(suggesting.props.handleKeyDown.call(suggesting, view, { key: "Enter" } as KeyboardEvent)).toBe(true);
    expect(finished).toBe(false);
    expect(state.doc.child(2).attrs.unknownAttrs).toMatchObject({
      suggestionSplit: suggestionId,
      splitKind: "insertion",
    });

    const beforeClick = state;
    const click = state.tr.setSelection(TextSelection.create(state.doc, 1));
    suggesting.spec.appendTransaction?.(
      [click],
      beforeClick,
      beforeClick.apply(click),
    );
    expect(finished).toBe(true);
  });

  it("keeps the suggestion when the caret stays on the inserted line", () => {
    let finished = false;
    const controller: SuggestingController = {
      enabled: true,
      suggestionId: suggestionId,
      pending: null,
      request: null,
      onNeedId: () => Promise.resolve(suggestionId),
      onBurstFinished: () => {
        finished = true;
      },
      onOpenSuggestion: () => {},
    };
    const extension = createSuggestingExtension({ current: controller });
    const plugins = extension.config.addProseMirrorPlugins?.call(extension) ?? [];
    const suggesting = plugins[0];
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [schema.text("Hello", [
        schema.marks[suggestionInsertName].create({ id: suggestionId }),
      ])]),
      schema.node("paragraph", {
        unknownAttrs: { suggestionSplit: suggestionId, splitKind: "insertion" },
      }),
    ]);
    const blank = doc.child(0).nodeSize + 1;
    const state = EditorState.create({
      schema,
      doc,
      selection: TextSelection.create(doc, 3),
      plugins,
    });
    const moved = state.tr.setSelection(TextSelection.create(state.doc, blank));
    suggesting.spec.appendTransaction?.([moved], state, state.apply(moved));
    expect(finished).toBe(false);
  });

  it("continues the author's own suggestion after the burst ended", () => {
    let needed = false;
    const controller: SuggestingController = {
      enabled: true,
      suggestionId: null,
      pending: null,
      request: null,
      ownsSuggestion: (id) => id === suggestionId,
      onNeedId: () => {
        needed = true;
        controller.suggestionId = "other";
        return Promise.resolve("other");
      },
      onBurstFinished: () => {},
      onOpenSuggestion: () => {},
    };
    const extension = createSuggestingExtension({ current: controller });
    const plugins = extension.config.addProseMirrorPlugins?.call(extension) ?? [];
    const suggesting = plugins[0];
    if (!suggesting?.props.handleTextInput) {
      throw new Error("expected a suggesting text handler");
    }
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [schema.text("hello", [
        schema.marks[suggestionInsertName].create({ id: suggestionId }),
      ])]),
    ]);
    let state = EditorState.create({ schema, doc, plugins });
    const view = {
      get state() {
        return state;
      },
      dispatch(transaction: Transaction) {
        state = state.apply(transaction);
      },
    } as EditorView;

    expect(suggesting.props.handleTextInput.call(
      suggesting,
      view,
      6,
      6,
      "!",
    )).toBe(true);
    expect(needed).toBe(false);
    expect(controller.suggestionId).toBe(suggestionId);
    expect(markedText(state.doc)).toEqual([
      { text: "hello!", marks: [suggestionInsertName] },
    ]);
  });

  it("starts a new suggestion beside another author's change", () => {
    let needed = false;
    const controller: SuggestingController = {
      enabled: true,
      suggestionId: null,
      pending: null,
      request: null,
      ownsSuggestion: () => false,
      onNeedId: () => {
        needed = true;
        controller.suggestionId = "other";
        return Promise.resolve("other");
      },
      onBurstFinished: () => {},
      onOpenSuggestion: () => {},
    };
    const extension = createSuggestingExtension({ current: controller });
    const plugins = extension.config.addProseMirrorPlugins?.call(extension) ?? [];
    const suggesting = plugins[0];
    if (!suggesting?.props.handleTextInput) {
      throw new Error("expected a suggesting text handler");
    }
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [schema.text("hello", [
        schema.marks[suggestionInsertName].create({ id: suggestionId }),
      ])]),
    ]);
    let state = EditorState.create({ schema, doc, plugins });
    const view = {
      get state() {
        return state;
      },
      dispatch(transaction: Transaction) {
        state = state.apply(transaction);
      },
    } as EditorView;

    expect(suggesting.props.handleTextInput.call(
      suggesting,
      view,
      6,
      6,
      "!",
    )).toBe(true);
    expect(needed).toBe(true);
    expect(controller.suggestionId).toBe("other");
  });

  it("leaves a paragraph boundary to the editor without ending the suggestion", () => {
    let finished = false;
    const controller: SuggestingController = {
      enabled: true,
      suggestionId: suggestionId,
      pending: null,
      request: null,
      onNeedId: () => Promise.resolve(suggestionId),
      onBurstFinished: () => {
        finished = true;
      },
      onOpenSuggestion: () => {},
    };
    const extension = createSuggestingExtension({ current: controller });
    const plugins = extension.config.addProseMirrorPlugins?.call(extension) ?? [];
    const suggesting = plugins[0];
    if (!suggesting?.props.handleKeyDown) {
      throw new Error("expected a suggesting key handler");
    }
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [schema.text("Hello")]),
      schema.node("paragraph", null, [schema.text("World")]),
    ]);
    let state = EditorState.create({
      schema,
      doc,
      selection: TextSelection.create(doc, doc.child(0).nodeSize + 1),
      plugins,
    });
    const view = {
      get state() {
        return state;
      },
      dispatch(transaction: Transaction) {
        state = state.apply(transaction);
      },
    } as EditorView;

    expect(suggesting.props.handleKeyDown.call(suggesting, view, { key: "Backspace" } as KeyboardEvent)).toBe(false);
    const before = state;
    let join = state.tr;
    const joined = joinBackward(state, (next) => {
      join = next;
    });
    if (!joined) throw new Error("expected backspace to join the paragraphs");
    suggesting.spec.appendTransaction?.([join], before, before.apply(join));
    expect(finished).toBe(false);
  });

  it("marks an enter inside existing text as a structural split", () => {
    const controller: SuggestingController = {
      enabled: true,
      suggestionId: suggestionId,
      pending: null,
      request: null,
      onNeedId: () => Promise.resolve(suggestionId),
      onBurstFinished: () => {},
      onOpenSuggestion: () => {},
    };
    const extension = createSuggestingExtension({ current: controller });
    const plugins = extension.config.addProseMirrorPlugins?.call(extension) ?? [];
    const suggesting = plugins[0];
    if (!suggesting?.props.handleKeyDown) {
      throw new Error("expected a suggesting key handler");
    }
    let state = EditorState.create({
      schema,
      doc: schema.node("doc", null, [
        schema.node("paragraph", null, [schema.text("HelloWorld")]),
      ]),
      plugins,
    });
    state = state.apply(state.tr.setSelection(TextSelection.create(state.doc, 6)));
    const view = {
      get state() {
        return state;
      },
      dispatch(transaction: Transaction) {
        state = state.apply(transaction);
      },
    } as EditorView;

    expect(suggesting.props.handleKeyDown.call(suggesting, view, { key: "Enter" } as KeyboardEvent)).toBe(true);
    expect(state.doc.child(0).textContent).toBe("Hello");
    expect(state.doc.child(1).textContent).toBe("World");
    expect(state.doc.child(1).attrs.unknownAttrs).toMatchObject({
      suggestionSplit: suggestionId,
      splitKind: "structural",
    });
    expect(markedText(state.doc)).toEqual([
      { text: "Hello", marks: [] },
      { text: "World", marks: [] },
    ]);
  });

  it("restores a deleted line when the same text is typed", () => {
    const state = editorState(schema.text("hello", [
      schema.marks[suggestionDeleteName].create({ id: suggestionId }),
    ]));

    const next = state.apply(suggestionInputTransaction(state, 1, 1, "hello", suggestionId));

    expect(markedText(next.doc)).toEqual([
      { text: "hello", marks: [] },
    ]);
  });

  it("restores a deletion made before the current suggesting burst", () => {
    const state = editorState(schema.text("hello", [
      schema.marks[suggestionDeleteName].create({ id: "persisted-suggestion" }),
    ]));

    const next = state.apply(
      suggestionInputTransaction(state, 1, 1, "hello", "new-suggestion"),
    );

    expect(markedText(next.doc)).toEqual([
      { text: "hello", marks: [] },
    ]);
  });

  it("inserts a different character in front of the remaining deletion", () => {
    const state = editorState(schema.text("hello", [
      schema.marks[suggestionDeleteName].create({ id: suggestionId }),
    ]));

    const next = state.apply(suggestionInputTransaction(state, 1, 1, "H", suggestionId));

    expect(markedText(next.doc)).toEqual([
      { text: "H", marks: [suggestionInsertName] },
      { text: "hello", marks: [suggestionDeleteName] },
    ]);
  });

  it("does not delete text that is already deleted", () => {
    const state = editorState(schema.text("thread whose", [
      schema.marks[suggestionDeleteName].create({ id: "older" }),
    ]));

    const next = state.apply(suggestionInputTransaction(state, 1, 13, "", "newer"));

    expect(markedText(next.doc)).toEqual([
      { text: "thread whose", marks: [suggestionDeleteName] },
    ]);
    expect(next.doc.textContent).toBe("thread whose");
    expect(deletionIds(next.doc)).toEqual(["older"]);
  });

  it("keeps the exact characters that were typed back", () => {
    const state = editorState(schema.text("Out of scope"));

    const next = state.apply(suggestionInputTransaction(state, 1, 13, "scope", suggestionId));

    expect(markedText(next.doc)).toEqual([
      { text: "Out of ", marks: [suggestionDeleteName] },
      { text: "scope", marks: [] },
    ]);
  });

  it("restores matching text from individual browser input events", () => {
    let state = editorState(schema.text("Out of scope"));

    state = state.apply(suggestionInputTransaction(state, 1, 13, "s", suggestionId));
    for (const character of "cope") {
      const cursor = state.selection.from;
      state = state.apply(
        suggestionInputTransaction(state, cursor, cursor, character, suggestionId),
      );
    }

    expect(markedText(state.doc)).toEqual([
      { text: "Out of ", marks: [suggestionDeleteName] },
      { text: "scope", marks: [] },
    ]);
  });

  it("shows a replacement as a deletion plus an insertion", () => {
    const state = editorState(schema.text("scope"));

    const next = state.apply(suggestionInputTransaction(state, 1, 6, "range", suggestionId));

    expect(markedText(next.doc)).toEqual([
      { text: "scop", marks: [suggestionDeleteName] },
      { text: "rang", marks: [suggestionInsertName] },
      { text: "e", marks: [] },
    ]);
  });

  it("restores a selection after marking it deleted", () => {
    const state = editorState(schema.text("hello"));

    const next = state.apply(suggestionInputTransaction(state, 1, 6, "hello", suggestionId));

    expect(markedText(next.doc)).toEqual([
      { text: "hello", marks: [] },
    ]);
  });

  it("remaps a split id on the block when the server id arrives", () => {
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [schema.text("Hello")]),
      schema.node("paragraph", {
        unknownAttrs: { suggestionSplit: "local:1", splitKind: "insertion" },
      }),
    ]);
    let state = EditorState.create({ schema, doc });
    const editor = {
      get state() {
        return state;
      },
      view: {
        dispatch(transaction: Transaction) {
          state = state.apply(transaction);
        },
      },
    };
    reassignSuggestionId(editor, "local:1", "server");
    expect(state.doc.child(1).attrs.unknownAttrs).toMatchObject({
      suggestionSplit: "server",
      splitKind: "insertion",
    });
  });
});

function editorState(text: ProseMirrorNode): EditorState {
  return EditorState.create({
    schema,
    doc: schema.node("doc", null, [
      schema.node("paragraph", null, [text]),
    ]),
  });
}

function deletionIds(doc: ProseMirrorNode): string[] {
  const ids: string[] = [];
  doc.descendants((node) => {
    if (!node.isText) return;
    for (const mark of node.marks) {
      if (mark.type.name === suggestionDeleteName && typeof mark.attrs.id === "string") {
        ids.push(mark.attrs.id);
      }
    }
  });
  return ids;
}

function markedText(doc: ProseMirrorNode): Array<{ text: string; marks: string[] }> {
  const parts: Array<{ text: string; marks: string[] }> = [];
  doc.descendants((node) => {
    if (!node.isText || !node.text) return;
    parts.push({
      text: node.text,
      marks: node.marks.map(mark => mark.type.name),
    });
  });
  return parts;
}

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
import type { Node as PmNode } from "@tiptap/pm/model";
import { EditorState, NodeSelection, type Selection, TextSelection, type Transaction } from "@tiptap/pm/state";
import type { Editor } from "@tiptap/react";
import { describe, expect, it } from "vitest";

import { richEditorCollaborationExtensions } from "../RichEditor";
import type { SuggestingController } from "../SuggestionExtension";
import { structuralSuggestionSplit } from "../suggestionStructure";

import { blockItemLabelKey, blockItems, insertMermaidFromToolbar, mermaidToolbarTarget } from "./blockItems";

const schema = getSchema(richEditorCollaborationExtensions);

function paragraph(text = "") {
  return schema.node("paragraph", null, text ? [schema.text(text)] : []);
}

function docOf(...blocks: PmNode[]) {
  return schema.node("doc", null, blocks);
}

function stateWith(doc: PmNode, select: (doc: PmNode) => Selection) {
  return EditorState.create({ schema, doc, selection: select(doc) });
}

function editorFor(state: EditorState) {
  const controller: SuggestingController = {
    enabled: true,
    suggestionId: null,
    pending: null,
    request: null,
    onStartSuggesting: () => {
      controller.suggestionId = "local:1";
      return Promise.resolve("local:1");
    },
    onSuggestionDone: () => {},
    onOpenSuggestion: () => {},
  };
  let current = state;
  return {
    storage: { documentSuggesting: { controller: { current: controller } } },
    get state() {
      return current;
    },
    view: {
      dispatch(transaction: Transaction) {
        current = current.apply(transaction);
      },
    },
  };
}

describe("blockItemLabelKey", () => {
  it("uses the toolbar's insert and block keys", () => {
    const keys = Object.fromEntries(blockItems.map(item => [item.id, blockItemLabelKey(item)]));
    expect(keys.heading2).toBe("ui.richEditor.toolbar.blocks.heading2");
    expect(keys.mermaid).toBe("ui.richEditor.toolbar.inserts.mermaid");
    expect(keys.image).toBe("ui.richEditor.toolbar.inserts.image");
  });
});

describe("mermaidToolbarTarget", () => {
  it("turns an empty paragraph into the diagram", () => {
    const state = stateWith(docOf(paragraph()), doc => TextSelection.create(doc, 1));
    expect(mermaidToolbarTarget(state)).toEqual({ convert: true });
  });

  it("inserts after a paragraph with text", () => {
    const state = stateWith(
      docOf(paragraph("Scope"), paragraph("Next")),
      doc => TextSelection.create(doc, 3),
    );
    expect(mermaidToolbarTarget(state)).toEqual({
      convert: false,
      position: state.doc.child(0).nodeSize,
    });
  });

  it("inserts after the whole list around the caret", () => {
    const list = schema.node("bulletList", null, [
      schema.node("listItem", null, [paragraph("One")]),
      schema.node("listItem", null, [paragraph("Two")]),
    ]);
    const state = stateWith(
      docOf(list, paragraph("After")),
      doc => TextSelection.create(doc, doc.child(0).nodeSize - 3),
    );
    expect(state.selection.$from.parent.textContent).toBe("Two");
    expect(mermaidToolbarTarget(state)).toEqual({
      convert: false,
      position: state.doc.child(0).nodeSize,
    });
  });

  it("inserts after an empty line a range selection starts on", () => {
    const state = stateWith(
      docOf(paragraph(), paragraph("Text")),
      doc => TextSelection.create(doc, 1, 5),
    );
    expect(mermaidToolbarTarget(state)).toEqual({
      convert: false,
      position: state.doc.child(0).nodeSize,
    });
  });

  it("inserts after a selected top-level node", () => {
    const state = stateWith(
      docOf(paragraph("Intro"), paragraph("Next")),
      doc => NodeSelection.create(doc, 0),
    );
    expect(mermaidToolbarTarget(state)).toEqual({
      convert: false,
      position: state.doc.child(0).nodeSize,
    });
  });
});

describe("insertMermaidFromToolbar while suggesting", () => {
  it.each([
    ["the caret's block", (doc: PmNode) => TextSelection.create(doc, 3)],
    ["a selected block", (doc: PmNode) => NodeSelection.create(doc, 0)],
  ])("inserts a suggested diagram after %s", (_name, select) => {
    const editor = editorFor(stateWith(docOf(paragraph("Scope"), paragraph("Next")), select));

    insertMermaidFromToolbar(editor as unknown as Editor);

    const { doc } = editor.state;
    expect(doc.childCount).toBe(3);
    expect(doc.child(0).textContent).toBe("Scope");
    const diagram = doc.child(1);
    expect(diagram.type.name).toBe("codeBlock");
    expect(diagram.attrs.language).toBe("mermaid");
    expect(structuralSuggestionSplit(diagram)).toEqual({ id: "local:1", kind: "insertion" });
    expect(doc.child(2).textContent).toBe("Next");
  });
});

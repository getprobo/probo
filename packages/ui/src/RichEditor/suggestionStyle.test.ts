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
import type { Mark, Node as ProseMirrorNode } from "@tiptap/pm/model";
import { EditorState, Plugin, TextSelection, type Transaction } from "@tiptap/pm/state";
import type { EditorView } from "@tiptap/pm/view";
import { describe, expect, it } from "vitest";

import { richEditorCollaborationExtensions } from "./RichEditor";
import {
  applySuggestingLink,
  applySuggestingStyle,
  createSuggestingExtension,
  discardSuggestionId,
  reassignSuggestionId,
  type SuggestingController,
  suggestionInsertName,
} from "./SuggestionExtension";

const schema = getSchema(richEditorCollaborationExtensions);
const suggestionId = "suggestion-1";

function suggestingPlugins(controller: SuggestingController): Plugin[] {
  const extension = createSuggestingExtension({ current: controller });
  const addPlugins = extension.config.addProseMirrorPlugins;
  if (!addPlugins) return [];
  return addPlugins.call(
    extension as unknown as ThisParameterType<typeof addPlugins>,
  );
}

function controller(): SuggestingController & { blocked: boolean } {
  const value = {
    enabled: true,
    suggestionId,
    pending: null,
    request: null,
    blocked: false,
    onNeedId: () => Promise.resolve(suggestionId),
    onBurstFinished: () => {},
    onOpenSuggestion: () => {},
    onBlocked: () => {
      value.blocked = true;
    },
  };
  return value;
}

function paragraph(text: string, marks: Mark[] = []) {
  return schema.node("doc", null, [
    schema.node("paragraph", null, text === "" ? [] : [schema.text(text, marks)]),
  ]);
}

function open(doc: ProseMirrorNode, from: number, to: number, current = controller()) {
  let state = EditorState.create({
    schema,
    doc,
    selection: TextSelection.create(doc, from, to),
    plugins: suggestingPlugins(current),
  });
  const view = {
    get state() {
      return state;
    },
    dispatch(transaction: Transaction) {
      state = state.apply(transaction);
    },
  } as EditorView;
  return {
    current,
    view,
    get state() {
      return state;
    },
    select(start: number, end: number) {
      view.dispatch(state.tr.setSelection(TextSelection.create(state.doc, start, end)));
    },
  };
}

function marksOn(doc: ProseMirrorNode, text: string): Mark[] {
  const found: Mark[] = [];
  doc.descendants((node) => {
    if (node.isText && node.text === text) found.push(...node.marks);
  });
  return found;
}

function markNames(marks: readonly Mark[]): string[] {
  return marks.map(mark => mark.type.name);
}

describe("format suggestions", () => {
  it("suggests bold and clears it when toggled again", () => {
    const editor = open(paragraph("Hello"), 1, 6);
    expect(applySuggestingStyle(editor.view, editor.current, "bold")).toBe(true);

    const marks = marksOn(editor.state.doc, "Hello");
    expect(markNames(marks)).toEqual(expect.arrayContaining(["bold", "suggestionFormatBold"]));
    expect(marks.find(mark => mark.type.name === "suggestionFormatBold")?.attrs).toMatchObject({
      id: suggestionId,
      action: "add",
    });

    editor.select(1, 6);
    expect(applySuggestingStyle(editor.view, editor.current, "bold")).toBe(true);
    expect(markNames(marksOn(editor.state.doc, "Hello"))).not.toEqual(
      expect.arrayContaining(["bold", "suggestionFormatBold"]),
    );
  });

  it("styles text inserted by the same suggestion without a format mark", () => {
    const inserted = schema.marks[suggestionInsertName].create({ id: suggestionId });
    const editor = open(paragraph("i", [inserted]), 1, 2);
    expect(applySuggestingStyle(editor.view, editor.current, "bold")).toBe(true);
    const marks = marksOn(editor.state.doc, "i");
    expect(markNames(marks)).toContain("bold");
    expect(markNames(marks)).toContain(suggestionInsertName);
    expect(markNames(marks)).not.toContain("suggestionFormatBold");
  });

  it("does not let typed text inherit a suggested style", () => {
    const editor = open(paragraph("Hello"), 1, 6);
    applySuggestingStyle(editor.view, editor.current, "bold");
    const plugin = suggestingPlugins(editor.current)[0];
    expect(plugin?.props.handleTextInput?.call(plugin, editor.view, 3, 3, "X", () => editor.state.tr)).toBe(true);
    const typed = marksOn(editor.state.doc, "X");
    expect(markNames(typed)).toContain(suggestionInsertName);
    expect(markNames(typed)).not.toContain("bold");
    expect(markNames(typed)).not.toContain("suggestionFormatBold");
  });

  it("blocks an untracked bold mark while suggesting", () => {
    const editor = open(paragraph("Hello"), 1, 6);
    const plugin = suggestingPlugins(editor.current)[0];
    const untracked = editor.state.tr.addMark(1, 6, schema.marks.bold.create());
    expect(plugin?.spec.filterTransaction?.call(plugin, untracked, editor.state)).toBe(false);
    expect(editor.current.blocked).toBe(true);
  });

  it("suggests a link and removing it", () => {
    const editor = open(paragraph("Hello"), 1, 6);
    expect(applySuggestingLink(editor.view, editor.current, "https://example.com")).toBe(true);
    let marks = marksOn(editor.state.doc, "Hello");
    expect(marks.find(mark => mark.type.name === "link")?.attrs.href).toBe("https://example.com");
    expect(marks.find(mark => mark.type.name === "suggestionFormatLink")?.attrs.action).toBe("add");

    editor.select(1, 6);
    expect(applySuggestingLink(editor.view, editor.current, null)).toBe(true);
    marks = marksOn(editor.state.doc, "Hello");
    expect(markNames(marks)).not.toContain("link");
    expect(marks.find(mark => mark.type.name === "suggestionFormatLink")).toBeUndefined();
  });

  it("suggests code by removing the other styles", () => {
    const editor = open(paragraph("Hello", [schema.marks.bold.create()]), 1, 6);
    expect(applySuggestingStyle(editor.view, editor.current, "code")).toBe(true);
    const marks = marksOn(editor.state.doc, "Hello");
    expect(markNames(marks)).toContain("code");
    expect(markNames(marks)).not.toContain("bold");
    expect(marks.find(mark => mark.type.name === "suggestionFormatCode")?.attrs.action).toBe("add");
    expect(marks.find(mark => mark.type.name === "suggestionFormatBold")?.attrs.action).toBe("remove");
  });

  it("refuses to format another suggestion", () => {
    const inserted = schema.marks[suggestionInsertName].create({ id: "someone-else" });
    const editor = open(paragraph("Hello", [inserted]), 1, 6);
    expect(applySuggestingStyle(editor.view, editor.current, "italic")).toBe(true);
    expect(editor.current.blocked).toBe(true);
    expect(markNames(marksOn(editor.state.doc, "Hello"))).toEqual([suggestionInsertName]);
  });

  it("rewrites and reverts format marks with the suggestion id", () => {
    const editor = open(paragraph("Hello"), 1, 6);
    applySuggestingStyle(editor.view, editor.current, "italic");
    const host = {
      get state() {
        return editor.state;
      },
      view: editor.view,
    };
    reassignSuggestionId(host, suggestionId, "server-id");
    expect(marksOn(editor.state.doc, "Hello").find(mark => mark.type.name === "suggestionFormatItalic")?.attrs.id)
      .toBe("server-id");

    discardSuggestionId(host, "server-id");
    expect(markNames(marksOn(editor.state.doc, "Hello"))).not.toContain("italic");
    expect(markNames(marksOn(editor.state.doc, "Hello"))).not.toContain("suggestionFormatItalic");
  });
});

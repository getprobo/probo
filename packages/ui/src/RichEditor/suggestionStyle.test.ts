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
    onStartSuggesting: () => Promise.resolve(suggestionId),
    onSuggestionDone: () => {},
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

function flushClaims(): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, 0));
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

  it("clears an owned underline suggestion after the burst id is dropped", () => {
    const editor = open(paragraph("Hello"), 1, 6);
    editor.current.ownsSuggestion = id => id === suggestionId;
    expect(applySuggestingStyle(editor.view, editor.current, "underline")).toBe(true);

    editor.current.suggestionId = null;
    let started = false;
    editor.current.onStartSuggesting = () => {
      started = true;
      return Promise.resolve(null);
    };
    editor.select(1, 6);
    editor.current.blocked = false;
    expect(applySuggestingStyle(editor.view, editor.current, "underline")).toBe(true);

    expect(started).toBe(false);
    expect(editor.current.blocked).toBe(false);
    expect(editor.current.suggestionId).toBeNull();
    const marks = marksOn(editor.state.doc, "Hello");
    expect(markNames(marks)).not.toContain("underline");
    expect(markNames(marks)).not.toContain("suggestionFormatUnderline");
  });

  it("refuses to retoggle a style suggestion the controller does not own", () => {
    const foreign = "someone-else";
    const current = controller();
    current.suggestionId = null;
    current.ownsSuggestion = () => false;
    current.onStartSuggesting = () => {
      current.suggestionId = "new-suggestion";
      return Promise.resolve(current.suggestionId);
    };
    const editor = open(paragraph("Hello", [
      schema.marks.underline.create(),
      schema.marks.suggestionFormatUnderline.create({ id: foreign, action: "add" }),
    ]), 1, 6, current);

    expect(applySuggestingStyle(editor.view, editor.current, "underline")).toBe(true);
    expect(editor.current.blocked).toBe(true);
    const marks = marksOn(editor.state.doc, "Hello");
    expect(markNames(marks)).toEqual(expect.arrayContaining(["underline", "suggestionFormatUnderline"]));
    expect(marks.find(mark => mark.type.name === "suggestionFormatUnderline")?.attrs.id).toBe(foreign);
  });

  it("claims an owned underline suggestion before clearing it", async () => {
    const owned = new Set<string>();
    const current = controller();
    current.suggestionId = null;
    current.ownsSuggestion = id => owned.has(id);
    let started = false;
    let claims = 0;
    current.onStartSuggesting = () => {
      started = true;
      return Promise.resolve(null);
    };
    current.claimSuggestion = (id) => {
      claims += 1;
      return Promise.resolve().then(() => {
        owned.add(id);
        return true;
      });
    };
    const editor = open(paragraph("Hello", [
      schema.marks.underline.create(),
      schema.marks.suggestionFormatUnderline.create({ id: suggestionId, action: "add" }),
    ]), 1, 6, current);

    expect(applySuggestingStyle(editor.view, editor.current, "underline")).toBe(true);
    expect(applySuggestingStyle(editor.view, editor.current, "underline")).toBe(true);
    await flushClaims();

    expect(claims).toBe(1);
    expect(started).toBe(false);
    expect(editor.current.blocked).toBe(false);
    expect(editor.current.suggestionId).toBeNull();
    const marks = marksOn(editor.state.doc, "Hello");
    expect(markNames(marks)).not.toContain("underline");
    expect(markNames(marks)).not.toContain("suggestionFormatUnderline");
  });

  it("blocks a claimed suggestion that belongs to someone else", async () => {
    const current = controller();
    current.suggestionId = null;
    current.ownsSuggestion = () => false;
    let started = false;
    current.onStartSuggesting = () => {
      started = true;
      return Promise.resolve(null);
    };
    let claims = 0;
    current.claimSuggestion = () => {
      claims += 1;
      return Promise.resolve(false);
    };
    const editor = open(paragraph("Hello", [
      schema.marks.underline.create(),
      schema.marks.suggestionFormatUnderline.create({ id: "someone-else", action: "add" }),
    ]), 1, 6, current);

    expect(applySuggestingStyle(editor.view, editor.current, "underline")).toBe(true);
    await flushClaims();

    expect(claims).toBe(1);
    expect(started).toBe(false);
    expect(editor.current.blocked).toBe(true);
    expect(editor.current.claimingSuggestionId).toBeNull();
    const marks = marksOn(editor.state.doc, "Hello");
    expect(markNames(marks)).toEqual(expect.arrayContaining(["underline", "suggestionFormatUnderline"]));
  });

  it("drops a claimed toggle when the selected text changed meanwhile", async () => {
    const owned = new Set<string>();
    const current = controller();
    current.suggestionId = null;
    current.ownsSuggestion = id => owned.has(id);
    current.claimSuggestion = (id) => {
      owned.add(id);
      return Promise.resolve(true);
    };
    const editor = open(paragraph("Hello", [
      schema.marks.underline.create(),
      schema.marks.suggestionFormatUnderline.create({ id: suggestionId, action: "add" }),
    ]), 1, 6, current);

    expect(applySuggestingStyle(editor.view, editor.current, "underline")).toBe(true);
    editor.view.dispatch(editor.state.tr.insertText("J", 1, 2));
    editor.select(1, 6);
    await flushClaims();

    expect(editor.current.blocked).toBe(false);
    expect(editor.state.doc.textContent).toBe("Jello");
    expect(markNames(marksOn(editor.state.doc, "Jello"))).toEqual(
      expect.arrayContaining(["underline", "suggestionFormatUnderline"]),
    );
  });

  it("removes bold from a selection that is only partly bold", () => {
    const owned = "owned-bold";
    const current = controller();
    current.suggestionId = null;
    current.ownsSuggestion = id => id === owned;
    current.onStartSuggesting = () => {
      current.suggestionId = "bold-burst";
      return Promise.resolve(current.suggestionId);
    };
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [
        schema.text("Hello", [
          schema.marks.bold.create(),
          schema.marks.suggestionFormatBold.create({ id: owned, action: "add" }),
        ]),
        schema.text("World"),
      ]),
    ]);
    const editor = open(doc, 1, 11, current);

    expect(applySuggestingStyle(editor.view, editor.current, "bold")).toBe(true);

    expect(editor.current.blocked).toBe(false);
    expect(markNames(marksOn(editor.state.doc, "Hello"))).not.toContain("bold");
    const world = markNames(marksOn(editor.state.doc, "World"));
    expect(world).not.toContain("bold");
    expect(world).not.toContain("suggestionFormatBold");
  });

  it("clears a loaded bold suggestion while another burst is open", () => {
    const loaded = "loaded-bold";
    const burst = "underline-burst";
    const current = controller();
    current.suggestionId = burst;
    current.ownsSuggestion = id => id === loaded || id === burst;
    let started = false;
    current.onStartSuggesting = () => {
      started = true;
      return Promise.resolve(null);
    };
    const editor = open(paragraph("Hello", [
      schema.marks.bold.create(),
      schema.marks.suggestionFormatBold.create({ id: loaded, action: "add" }),
    ]), 1, 6, current);

    expect(applySuggestingStyle(editor.view, editor.current, "bold")).toBe(true);

    expect(started).toBe(false);
    expect(editor.current.blocked).toBe(false);
    const marks = marksOn(editor.state.doc, "Hello");
    expect(markNames(marks)).not.toContain("bold");
    expect(markNames(marks)).not.toContain("suggestionFormatBold");
  });

  it("clears two owned bold suggestions in one selection", () => {
    const first = "bold-a";
    const second = "bold-b";
    const current = controller();
    current.suggestionId = null;
    current.ownsSuggestion = id => id === first || id === second;
    const suggestedBold = (id: string) => [
      schema.marks.bold.create(),
      schema.marks.suggestionFormatBold.create({ id, action: "add" }),
    ];
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [
        schema.text("Hello", suggestedBold(first)),
        schema.text("World", suggestedBold(second)),
      ]),
    ]);
    const editor = open(doc, 1, 11, current);

    expect(applySuggestingStyle(editor.view, editor.current, "bold")).toBe(true);

    expect(editor.current.blocked).toBe(false);
    expect(markNames(marksOn(editor.state.doc, "Hello"))).not.toContain("bold");
    expect(markNames(marksOn(editor.state.doc, "World"))).not.toContain("bold");
  });

  it("shortens an owned bold suggestion to the text outside the selection", () => {
    const loaded = "loaded-bold";
    const current = controller();
    current.suggestionId = null;
    current.ownsSuggestion = id => id === loaded;
    const editor = open(paragraph("HelloWorld", [
      schema.marks.bold.create(),
      schema.marks.suggestionFormatBold.create({ id: loaded, action: "add" }),
    ]), 6, 11, current);

    expect(applySuggestingStyle(editor.view, editor.current, "bold")).toBe(true);

    expect(editor.current.blocked).toBe(false);
    const kept = marksOn(editor.state.doc, "Hello");
    expect(markNames(kept)).toContain("bold");
    expect(kept.find(mark => mark.type.name === "suggestionFormatBold")?.attrs.id).toBe(loaded);
    expect(markNames(marksOn(editor.state.doc, "World"))).not.toContain("bold");
  });

  it("underlines text without removing the bold suggestion under it", () => {
    const loaded = "loaded-bold";
    const current = controller();
    current.suggestionId = null;
    current.ownsSuggestion = id => id === loaded;
    current.onStartSuggesting = () => {
      current.suggestionId = "underline-burst";
      return Promise.resolve(current.suggestionId);
    };
    const editor = open(paragraph("Hello", [
      schema.marks.bold.create(),
      schema.marks.suggestionFormatBold.create({ id: loaded, action: "add" }),
    ]), 1, 6, current);

    expect(applySuggestingStyle(editor.view, editor.current, "underline")).toBe(true);

    expect(editor.current.blocked).toBe(false);
    const marks = marksOn(editor.state.doc, "Hello");
    expect(markNames(marks)).toEqual(expect.arrayContaining([
      "bold",
      "suggestionFormatBold",
      "underline",
      "suggestionFormatUnderline",
    ]));
    expect(marks.find(mark => mark.type.name === "suggestionFormatBold")?.attrs.id).toBe(loaded);
    expect(marks.find(mark => mark.type.name === "suggestionFormatUnderline")?.attrs.id)
      .toBe("underline-burst");
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

  it("types a stored bold mark at a caret as part of the insertion", () => {
    const editor = open(paragraph("Hello"), 3, 3);
    expect(applySuggestingStyle(editor.view, editor.current, "bold")).toBe(false);
    editor.view.dispatch(editor.state.tr.addStoredMark(schema.marks.bold.create()));
    expect(editor.current.blocked).toBe(false);
    expect(markNames(editor.state.storedMarks ?? [])).toEqual(["bold"]);

    const plugin = suggestingPlugins(editor.current)[0];
    expect(plugin?.props.handleTextInput?.call(plugin, editor.view, 3, 3, "X", () => editor.state.tr)).toBe(true);
    const typed = markNames(marksOn(editor.state.doc, "X"));
    expect(typed).toContain("bold");
    expect(typed).toContain(suggestionInsertName);
    expect(typed).not.toContain("suggestionFormatBold");
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

  it("blocks a javascript link without adding the mark", () => {
    const editor = open(paragraph("Hello"), 1, 6);
    expect(applySuggestingLink(editor.view, editor.current, "javascript:alert(1)")).toBe(true);
    expect(editor.current.blocked).toBe(true);
    expect(markNames(marksOn(editor.state.doc, "Hello"))).not.toContain("link");
  });

  it("leaves an unsafe link to the direct editor when suggesting is off", () => {
    const current = controller();
    current.enabled = false;
    const editor = open(paragraph("Hello"), 1, 6, current);
    expect(applySuggestingLink(editor.view, editor.current, "javascript:alert(1)")).toBe(false);
    expect(editor.current.blocked).toBe(false);
    expect(markNames(marksOn(editor.state.doc, "Hello"))).not.toContain("link");
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

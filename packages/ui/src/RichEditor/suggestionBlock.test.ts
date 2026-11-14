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
import { EditorState, TextSelection } from "@tiptap/pm/state";
import { describe, expect, it } from "vitest";

import { richEditorCollaborationExtensions } from "./RichEditor";
import {
  discardBlockSuggestions,
  headingCommand,
  listCommand,
  paragraphCommand,
  suggestCommand,
  suggestionBlockAttr,
  suggestTurnInto,
} from "./suggestionBlock";

const schema = getSchema(richEditorCollaborationExtensions);

function paragraphState(text: string) {
  const doc = schema.node("doc", null, [
    schema.node("paragraph", null, [schema.text(text)]),
  ]);
  return EditorState.create({ doc, schema });
}

function blockMark(state: EditorState) {
  const node = state.doc.firstChild;
  const unknown = node?.attrs.unknownAttrs as Record<string, string> | null;
  const raw = unknown?.[suggestionBlockAttr];
  if (!raw) return null;
  return JSON.parse(raw) as { id: string; type: string };
}

describe("suggestCommand", () => {
  it("records the paragraph when a heading is suggested", () => {
    const state = paragraphState("Hello");
    const next = suggestCommand(state, "sug-1", headingCommand(2));
    expect(next).not.toBeNull();
    const applied = state.apply(next!);
    expect(applied.doc.firstChild?.type.name).toBe("heading");
    expect(applied.doc.firstChild?.attrs.level).toBe(2);
    expect(blockMark(applied)).toMatchObject({ id: "sug-1", type: "paragraph" });
  });

  it("records the paragraph when a bullet list is suggested", () => {
    const state = paragraphState("Hello");
    const next = suggestCommand(state, "sug-2", listCommand("bulletList"));
    expect(next).not.toBeNull();
    const applied = state.apply(next!);
    expect(applied.doc.firstChild?.type.name).toBe("bulletList");
    const item = applied.doc.firstChild?.firstChild;
    const unknown = item?.attrs.unknownAttrs as Record<string, string> | null;
    const mark = JSON.parse(unknown?.[suggestionBlockAttr] ?? "null") as { id: string; type: string };
    expect(mark).toMatchObject({ id: "sug-2", type: "paragraph" });
  });

  it("keeps the other items in their list when one becomes text", () => {
    const item = (text: string) => schema.node("listItem", null, [schema.node("paragraph", null, [schema.text(text)])]);
    const list = schema.node("bulletList", null, [item("One"), item("Two")]);
    const state = EditorState.create({ doc: schema.node("doc", null, [list]), schema });
    const next = suggestTurnInto(state, "sug-4", 0, list, paragraphCommand(), "textblock");
    expect(next).not.toBeNull();
    const applied = state.apply(next!);
    const types: string[] = [];
    applied.doc.forEach(node => types.push(node.type.name));
    expect(types).toEqual(["paragraph", "bulletList"]);
    expect(applied.doc.child(1).textContent).toBe("Two");
  });

  it("wraps every paragraph of a quote in the new list", () => {
    const quote = schema.node("blockquote", null, [
      schema.node("paragraph", null, [schema.text("One")]),
      schema.node("paragraph", null, [schema.text("Two")]),
    ]);
    const state = EditorState.create({ doc: schema.node("doc", null, [quote]), schema });
    const next = suggestTurnInto(state, "sug-5", 0, quote, listCommand("bulletList"), "wrapper");
    expect(next).not.toBeNull();
    const applied = state.apply(next!);
    expect(applied.doc.childCount).toBe(1);
    expect(applied.doc.firstChild?.type.name).toBe("bulletList");
    expect(applied.doc.firstChild?.childCount).toBe(2);
  });

  it("restores the paragraph when a one-item list suggestion is discarded", () => {
    const state = paragraphState("Hello");
    const suggested = suggestCommand(state, "sug-6", listCommand("bulletList"));
    expect(suggested).not.toBeNull();
    const withList = state.apply(suggested!);
    const tr = withList.tr;
    discardBlockSuggestions(tr, "sug-6");
    const restored = withList.apply(tr);
    expect(restored.doc.childCount).toBe(1);
    expect(restored.doc.firstChild?.type.name).toBe("paragraph");
    expect(restored.doc.textContent).toBe("Hello");
    expect(blockMark(restored)).toBeNull();
  });

  it("records every paragraph a selection turns into headings", () => {
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [schema.text("One")]),
      schema.node("paragraph", null, [schema.text("Two")]),
    ]);
    const base = EditorState.create({ doc, schema });
    const state = base.apply(base.tr.setSelection(TextSelection.create(base.doc, 1, base.doc.content.size - 1)));
    const next = suggestCommand(state, "sug-3", headingCommand(1));
    expect(next).not.toBeNull();
    const applied = state.apply(next!);
    const marks: Array<{ id: string; type: string }> = [];
    applied.doc.forEach((node) => {
      const unknown = node.attrs.unknownAttrs as Record<string, string> | null;
      const raw = unknown?.[suggestionBlockAttr];
      if (raw) marks.push(JSON.parse(raw) as { id: string; type: string });
    });
    expect(marks).toEqual([
      expect.objectContaining({ id: "sug-3", type: "paragraph" }),
      expect.objectContaining({ id: "sug-3", type: "paragraph" }),
    ]);
  });
});

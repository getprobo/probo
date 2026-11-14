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

import { Schema } from "@tiptap/pm/model";
import { describe, expect, it } from "vitest";

import { collectSuggestionAnchors } from "./suggestionPreview";

const schema = new Schema({
  nodes: {
    doc: { content: "paragraph+" },
    paragraph: {
      content: "text*",
      attrs: { unknownAttrs: { default: null } },
    },
    text: { group: "inline" },
  },
  marks: {
    suggestionInsert: {
      attrs: { id: { default: null } },
    },
    suggestionDelete: {
      attrs: { id: { default: null } },
    },
    suggestionFormatBold: { attrs: { id: { default: null }, action: { default: "add" }, href: { default: null } } },
    suggestionFormatItalic: { attrs: { id: { default: null }, action: { default: "add" }, href: { default: null } } },
    suggestionFormatStrike: { attrs: { id: { default: null }, action: { default: "add" }, href: { default: null } } },
  },
});

const id = "suggestion-1";

describe("suggestion preview", () => {
  it("counts a blank inserted line", () => {
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [
        schema.text("hello", [schema.marks.suggestionInsert.create({ id })]),
      ]),
      schema.node("paragraph", {
        unknownAttrs: { suggestionSplit: id, splitKind: "insertion" },
      }),
      schema.node("paragraph", {
        unknownAttrs: { suggestionSplit: id, splitKind: "insertion" },
      }, [
        schema.text("next", [schema.marks.suggestionInsert.create({ id })]),
      ]),
    ]);

    expect(collectSuggestionAnchors(doc).get(id)?.inserted).toBe("hello\n\nnext");
  });

  it("counts an inserted line that has no text", () => {
    const doc = schema.node("doc", null, [
      schema.node("paragraph", {
        unknownAttrs: { suggestionSplit: id, splitKind: "insertion" },
      }),
    ]);

    expect(collectSuggestionAnchors(doc).get(id)?.inserted).toBe("\n");
  });

  it("keeps the letters between two deletions", () => {
    const deleted = schema.marks.suggestionDelete.create({ id });
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [
        schema.text("s"),
        schema.text("elec", [deleted]),
        schema.text("tio"),
        schema.text("n", [deleted]),
      ]),
    ]);

    expect(collectSuggestionAnchors(doc).get(id)?.parts).toEqual([
      { kind: "deleted", text: "elec" },
      { kind: "kept", text: "tio" },
      { kind: "deleted", text: "n" },
    ]);
  });

  it("counts a deleted line break", () => {
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [schema.text("Hello")]),
      schema.node("paragraph", {
        unknownAttrs: { suggestionSplit: id, splitKind: "deletion" },
      }, [schema.text("World")]),
    ]);

    expect(collectSuggestionAnchors(doc).get(id)?.deleted).toBe("\n");
    expect(collectSuggestionAnchors(doc).get(id)?.parts).toEqual([
      { kind: "deleted", text: "\n" },
    ]);
  });

  it("counts the line between two deleted lines", () => {
    const deleted = schema.marks.suggestionDelete.create({ id });
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [
        schema.text("Hello", [deleted]),
      ]),
      schema.node("paragraph", {
        unknownAttrs: { suggestionSplit: id, splitKind: "deletion" },
      }, [
        schema.text("World", [deleted]),
      ]),
    ]);

    expect(collectSuggestionAnchors(doc).get(id)?.deleted).toBe("Hello\nWorld");
  });

  it("does not count a structural split as an inserted line", () => {
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [schema.text("Hello")]),
      schema.node("paragraph", {
        unknownAttrs: { suggestionSplit: id, splitKind: "structural" },
      }, [schema.text("World")]),
    ]);

    expect(collectSuggestionAnchors(doc).has(id)).toBe(false);
  });

  it("previews a deleted table row from its cells", () => {
    const deleted = tableSchema.marks.suggestionDelete.create({ id });
    const split = { suggestionSplit: id, splitKind: "deletion" };
    const doc = tableSchema.node("doc", null, [
      tableSchema.node("table", null, [
        tableRow(null, ["Right", "Access"]),
        tableSchema.node("tableRow", { unknownAttrs: split }, [
          tableSchema.node("tableCell", { unknownAttrs: split }, [
            tableSchema.node("paragraph", null, [tableSchema.text("Read", [deleted])]),
          ]),
          tableSchema.node("tableCell", { unknownAttrs: split }, [
            tableSchema.node("paragraph", null, [tableSchema.text("Yes", [deleted])]),
          ]),
        ]),
      ]),
    ]);

    const preview = collectSuggestionAnchors(doc).get(id);
    expect(preview?.structure).toEqual({ action: "remove", target: "row", text: "Read Yes" });
    expect(preview?.parts).toEqual([]);
    expect(preview?.from).toBeGreaterThan(0);
  });

  it("anchors an inserted empty table row", () => {
    const split = { suggestionSplit: id, splitKind: "insertion" };
    const doc = tableSchema.node("doc", null, [
      tableSchema.node("table", null, [
        tableRow(null, ["Right", "Access"]),
        tableRow(split, ["", ""], split),
      ]),
    ]);

    const preview = collectSuggestionAnchors(doc).get(id);
    expect(preview?.structure).toEqual({ action: "add", target: "row", text: "" });
    expect(preview?.parts).toEqual([]);
    expect(preview?.from).toBeGreaterThan(0);
  });

  it("describes an added column from its cells", () => {
    const split = { suggestionSplit: id, splitKind: "insertion" };
    const doc = tableSchema.node("doc", null, [
      tableSchema.node("table", null, [
        tableSchema.node("tableRow", null, [
          tableSchema.node("tableCell", null, [
            tableSchema.node("paragraph", null, [tableSchema.text("Keep")]),
          ]),
          tableSchema.node("tableCell", { unknownAttrs: split }, [
            tableSchema.node("paragraph", null, [tableSchema.text("dddd")]),
          ]),
        ]),
        tableSchema.node("tableRow", null, [
          tableSchema.node("tableCell", null, [
            tableSchema.node("paragraph", null, [tableSchema.text("Stay")]),
          ]),
          tableSchema.node("tableCell", { unknownAttrs: split }, [
            tableSchema.node("paragraph", null, [tableSchema.text("d")]),
          ]),
        ]),
      ]),
    ]);

    expect(collectSuggestionAnchors(doc).get(id)?.structure).toEqual({
      action: "add",
      target: "column",
      text: "dddd d",
    });
  });

  it("describes suggested formatting without a text diff", () => {
    const bold = schema.marks.suggestionFormatBold.create({ id, action: "add" });
    const italic = schema.marks.suggestionFormatItalic.create({ id, action: "add" });
    const strike = schema.marks.suggestionFormatStrike.create({ id: "other", action: "remove" });
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [
        schema.text("Keep", [bold, italic]),
        schema.text(" this", [strike]),
      ]),
    ]);

    const preview = collectSuggestionAnchors(doc).get(id);
    expect(preview?.formats).toEqual({ adds: ["bold", "italic"], removes: [] });
    expect(preview?.parts).toEqual([]);
    expect(preview?.structure).toBeNull();
  });
});

const tableSchema = new Schema({
  nodes: {
    doc: { content: "table+" },
    table: { content: "tableRow+", attrs: { unknownAttrs: { default: null } } },
    tableRow: { content: "tableCell+", attrs: { unknownAttrs: { default: null } } },
    tableCell: { content: "paragraph+", attrs: { unknownAttrs: { default: null } } },
    paragraph: {
      content: "text*",
      attrs: { unknownAttrs: { default: null } },
    },
    text: { group: "inline" },
  },
  marks: {
    suggestionInsert: { attrs: { id: { default: null } } },
    suggestionDelete: { attrs: { id: { default: null } } },
  },
});

function tableRow(
  unknownAttrs: Record<string, unknown> | null,
  cells: string[],
  cellAttrs: Record<string, unknown> | null = null,
) {
  return tableSchema.node("tableRow", { unknownAttrs }, cells.map(text =>
    tableSchema.node("tableCell", { unknownAttrs: cellAttrs }, [
      tableSchema.node("paragraph", null, text ? [tableSchema.text(text)] : []),
    ]),
  ));
}

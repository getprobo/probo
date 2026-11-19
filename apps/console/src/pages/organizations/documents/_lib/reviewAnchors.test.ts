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

import { type Node as ProseMirrorNode, Schema } from "@tiptap/pm/model";
import { describe, expect, it } from "vitest";

import { maxChangeTextLength, maxSuggestionChanges, suggestionAnchors } from "./reviewAnchors";

const formatAttrs = {
  id: { default: null },
  action: { default: "add" },
  href: { default: null },
};
const blockAttrs = { unknownAttrs: { default: null } };

const schema = new Schema({
  nodes: {
    doc: { content: "block+" },
    paragraph: { group: "block", content: "text*", attrs: blockAttrs },
    heading: {
      group: "block",
      content: "text*",
      attrs: { level: { default: 1 }, ...blockAttrs },
    },
    table: { group: "block", content: "tableRow+", attrs: blockAttrs },
    tableRow: { content: "tableCell+", attrs: blockAttrs },
    tableCell: { content: "paragraph+", attrs: blockAttrs },
    text: { group: "inline" },
  },
  marks: {
    suggestionInsert: { attrs: { id: { default: null } } },
    suggestionDelete: { attrs: { id: { default: null } } },
    suggestionFormatBold: { attrs: formatAttrs },
    suggestionFormatItalic: { attrs: formatAttrs },
    suggestionFormatLink: { attrs: formatAttrs },
    link: { attrs: { href: { default: "" } } },
  },
});

const id = "suggestion-1";
const insert = schema.marks.suggestionInsert.create({ id });
const remove = schema.marks.suggestionDelete.create({ id });

function paragraph(...text: ProseMirrorNode[]) {
  return schema.node("paragraph", null, text);
}

function doc(...blocks: ProseMirrorNode[]) {
  return schema.node("doc", null, blocks);
}

function split(kind: "insertion" | "deletion") {
  return { unknownAttrs: { suggestionSplit: id, splitKind: kind } };
}

function anchor(node: ProseMirrorNode, key = id) {
  const found = suggestionAnchors(node).get(key);
  if (!found) throw new Error(`expected suggestion ${key}`);
  return found;
}

describe("suggestionAnchors", () => {
  it("reads an adjacent delete and insert as one replacement", () => {
    const result = anchor(doc(paragraph(
      schema.text("Keep "),
      schema.text("old", [remove]),
      schema.text("new", [insert]),
    )));

    expect(result.from).toBe(6);
    expect(result.changes).toEqual([{ kind: "text", deleted: "old", inserted: "new" }]);
    expect(result.changeCount).toBe(1);
  });

  it("keeps separated edits as separate changes", () => {
    const result = anchor(doc(paragraph(
      schema.text("a", [insert]),
      schema.text(" plain "),
      schema.text("b", [insert]),
    )));

    expect(result.changes).toEqual([
      { kind: "text", deleted: "", inserted: "a" },
      { kind: "text", deleted: "", inserted: "b" },
    ]);
  });

  it("joins text across a line break into one change", () => {
    const result = anchor(doc(
      paragraph(schema.text("first", [insert])),
      paragraph(schema.text("second", [insert])),
    ));

    expect(result.changes).toEqual([{ kind: "text", deleted: "", inserted: "first\nsecond" }]);
  });

  it("merges formatting per style and action", () => {
    const bold = schema.marks.suggestionFormatBold.create({ id, action: "add" });
    const italic = schema.marks.suggestionFormatItalic.create({ id, action: "remove" });
    const result = anchor(doc(paragraph(
      schema.text("one", [bold]),
      schema.text(" plain "),
      schema.text("two", [bold]),
      schema.text("three", [italic]),
    )));

    expect(result.changes).toEqual([
      { kind: "format", style: "bold", action: "add", href: null, previousHref: null },
      { kind: "format", style: "italic", action: "remove", href: null, previousHref: null },
    ]);
  });

  it("reads the new and previous link address", () => {
    const result = anchor(doc(paragraph(schema.text("site", [
      schema.marks.link.create({ href: "https://new.example" }),
      schema.marks.suggestionFormatLink.create({ id, action: "add", href: "https://old.example" }),
    ]))));

    expect(result.changes).toEqual([{
      kind: "format",
      style: "link",
      action: "add",
      href: "https://new.example",
      previousHref: "https://old.example",
    }]);
  });

  it("reads a block retype from the previous block", () => {
    const previous = JSON.stringify({ id, type: "paragraph", parents: [], attrs: {} });
    const result = anchor(doc(schema.node("heading", {
      level: 2,
      unknownAttrs: { suggestionBlock: previous },
    }, [schema.text("Title")])));

    expect(result.changes).toEqual([{
      kind: "block",
      before: { kind: "paragraph", level: null },
      after: { kind: "heading", level: 2 },
    }]);
  });

  it("reads a deleted row once, without the text inside it", () => {
    const result = anchor(doc(schema.node("table", null, [
      schema.node("tableRow", split("deletion"), [
        schema.node("tableCell", split("deletion"), [paragraph(schema.text("cell", [remove]))]),
      ]),
    ])));

    expect(result.changes).toEqual([{ kind: "node", node: "row", action: "delete" }]);
    expect(result.changeCount).toBe(1);
  });

  it("collapses inserted column cells into one change", () => {
    const result = anchor(doc(schema.node("table", null, [
      schema.node("tableRow", null, [schema.node("tableCell", split("insertion"), [paragraph()])]),
      schema.node("tableRow", null, [schema.node("tableCell", split("insertion"), [paragraph()])]),
    ])));

    expect(result.changes).toEqual([{ kind: "node", node: "column", action: "insert" }]);
  });

  it("reports a blank inserted line when there is no text change", () => {
    const result = anchor(doc(paragraph(schema.text("Hello")), schema.node("paragraph", split("insertion"))));

    expect(result.changes).toEqual([{ kind: "node", node: "line", action: "insert" }]);
  });

  it("lists the first changes and counts every change", () => {
    const text: ProseMirrorNode[] = [];
    for (let index = 0; index < 5; index += 1) {
      text.push(schema.text(`edit${index}`, [insert]), schema.text(" "));
    }
    const result = anchor(doc(paragraph(...text)));

    expect(result.changes).toHaveLength(maxSuggestionChanges);
    expect(result.changeCount).toBe(5);
  });

  it("cuts long text with an ellipsis", () => {
    const result = anchor(doc(paragraph(schema.text("x".repeat(maxChangeTextLength + 20), [insert]))));
    const change = result.changes[0];

    expect(change?.kind === "text" && change.inserted).toBe(`${"x".repeat(maxChangeTextLength)}…`);
  });
});

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

  it("does not count a structural split as an inserted line", () => {
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [schema.text("Hello")]),
      schema.node("paragraph", {
        unknownAttrs: { suggestionSplit: id, splitKind: "structural" },
      }, [schema.text("World")]),
    ]);

    expect(collectSuggestionAnchors(doc).has(id)).toBe(false);
  });
});

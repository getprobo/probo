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

import { suggestionAnchors } from "./reviewAnchors";

const formatAttrs = {
  id: { default: null },
  action: { default: "add" },
  href: { default: null },
};

const schema = new Schema({
  nodes: {
    doc: { content: "block+" },
    paragraph: {
      group: "block",
      content: "text*",
      attrs: { unknownAttrs: { default: null } },
    },
    table: {
      group: "block",
      content: "paragraph+",
      attrs: { unknownAttrs: { default: null } },
    },
    text: { group: "inline" },
  },
  marks: {
    suggestionInsert: { attrs: { id: { default: null } } },
    suggestionDelete: { attrs: { id: { default: null } } },
    suggestionFormatBold: { attrs: formatAttrs },
  },
});

const id = "suggestion-1";

function paragraph(...text: ReturnType<typeof schema.text>[]) {
  return schema.node("paragraph", null, text);
}

describe("suggestionAnchors", () => {
  it("collects inserted and deleted text from the first position", () => {
    const doc = schema.node("doc", null, [
      paragraph(
        schema.text("Keep "),
        schema.text("old", [schema.marks.suggestionDelete.create({ id })]),
        schema.text("new", [schema.marks.suggestionInsert.create({ id })]),
      ),
    ]);

    expect(suggestionAnchors(doc).get(id)).toEqual({
      id,
      from: 6,
      inserted: "new",
      deleted: "old",
      formatting: false,
      structure: false,
    });
  });

  it("joins text from separate blocks with a line break", () => {
    const insert = schema.marks.suggestionInsert.create({ id });
    const doc = schema.node("doc", null, [
      paragraph(schema.text("first", [insert])),
      paragraph(schema.text("second", [insert])),
    ]);

    expect(suggestionAnchors(doc).get(id)?.inserted).toBe("first\nsecond");
  });

  it("flags formatting and structure changes", () => {
    const doc = schema.node("doc", null, [
      paragraph(schema.text("bold", [schema.marks.suggestionFormatBold.create({ id })])),
      schema.node("table", {
        unknownAttrs: { suggestionSplit: "table-1", splitKind: "structural" },
      }, [paragraph(schema.text("cell"))]),
    ]);
    const anchors = suggestionAnchors(doc);

    expect(anchors.get(id)?.formatting).toBe(true);
    expect(anchors.get(id)?.structure).toBe(false);
    expect(anchors.get("table-1")?.structure).toBe(true);
  });

  it("keeps separate suggestions apart", () => {
    const doc = schema.node("doc", null, [
      paragraph(
        schema.text("a", [schema.marks.suggestionInsert.create({ id: "one" })]),
        schema.text("b", [schema.marks.suggestionInsert.create({ id: "two" })]),
      ),
    ]);
    const anchors = suggestionAnchors(doc);

    expect(anchors.get("one")?.inserted).toBe("a");
    expect(anchors.get("two")?.inserted).toBe("b");
    expect(anchors.get("two")?.from).toBe(2);
  });
});

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

import { decorationsForHighlights } from "./historyDecorations";

const schema = new Schema({
  nodes: {
    doc: { content: "block+" },
    paragraph: { content: "inline*", group: "block" },
    text: { group: "inline" },
  },
});

describe("history decorations", () => {
  it("keeps a highlight in an earlier paragraph", () => {
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [schema.text("ab")]),
      schema.node("paragraph", null, [schema.text("cd")]),
    ]);

    const [highlight] = decorationsForHighlights(doc, [{
      kind: "INSERT",
      from: 1,
      to: 2,
      text: "a",
    }]).find();

    expect(highlight?.from).toBe(1);
    expect(highlight?.to).toBe(2);
  });

  it("places a deletion at the start of a later paragraph", () => {
    const doc = schema.node("doc", null, [
      schema.node("paragraph", null, [schema.text("ab")]),
      schema.node("paragraph", null, [schema.text("c")]),
    ]);

    const [highlight] = decorationsForHighlights(doc, [{
      kind: "DELETE",
      from: 4,
      to: 4,
      text: "c",
    }]).find();

    expect(highlight?.from).toBe(5);
    expect(highlight?.to).toBe(5);
  });
});

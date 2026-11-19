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
import { EditorState, TextSelection } from "@tiptap/pm/state";
import { describe, expect, it } from "vitest";

import { richEditorCollaborationExtensions } from "../RichEditor";

import { currentBlockType } from "./currentBlockType";

const schema = getSchema(richEditorCollaborationExtensions);

function stateAt(block: PmNode, pos = 2) {
  const doc = schema.node("doc", null, [block]);
  const state = EditorState.create({ doc, schema });
  return state.apply(state.tr.setSelection(TextSelection.create(state.doc, pos)));
}

function paragraph(text = "Hello") {
  return schema.node("paragraph", null, [schema.text(text)]);
}

describe("currentBlockType", () => {
  it("names a paragraph", () => {
    expect(currentBlockType(stateAt(paragraph()))).toBe("paragraph");
  });

  it("names a heading by level", () => {
    const heading = schema.node("heading", { level: 3 }, [schema.text("Title")]);
    expect(currentBlockType(stateAt(heading))).toBe("heading3");
  });

  it("returns null for a heading the menu cannot produce", () => {
    const heading = schema.node("heading", { level: 5 }, [schema.text("Title")]);
    expect(currentBlockType(stateAt(heading))).toBeNull();
  });

  it("names the list around a list item paragraph", () => {
    const list = schema.node("orderedList", null, [
      schema.node("listItem", null, [paragraph()]),
    ]);
    expect(currentBlockType(stateAt(list, 4))).toBe("orderedList");
  });

  it("names a blockquote", () => {
    const quote = schema.node("blockquote", null, [paragraph()]);
    expect(currentBlockType(stateAt(quote, 3))).toBe("blockquote");
  });

  it("names a code block but not a Mermaid diagram", () => {
    const code = schema.node("codeBlock", null, [schema.text("x = 1")]);
    const mermaid = schema.node("codeBlock", { language: "mermaid" }, [schema.text("flowchart TD")]);
    expect(currentBlockType(stateAt(code))).toBe("codeBlock");
    expect(currentBlockType(stateAt(mermaid))).toBeNull();
  });
});

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
import { EditorState } from "@tiptap/pm/state";
import { describe, expect, it } from "vitest";

import { richEditorCollaborationExtensions } from "./RichEditor";
import {
  headingCommand,
  listCommand,
  suggestCommand,
  suggestionBlockAttr,
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
    const mark = JSON.parse(unknown?.[suggestionBlockAttr] ?? "null") as { type: string };
    expect(mark.type).toBe("paragraph");
  });
});

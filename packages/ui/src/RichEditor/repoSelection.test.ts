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

import * as Automerge from "@automerge/automerge";
import { describe, expect, it } from "vitest";

import { isCollapsed, resolveSelection, textSelection } from "./repoSelection";

type Doc = { body: string };

function seededDocument(text: string): Automerge.Doc<Doc> {
  const document = Automerge.from<Doc>({ body: "" });
  return Automerge.change(document, (draft) => {
    Automerge.splice(draft, ["body"], 0, 0, text);
  });
}

describe("repo selection", () => {
  it("resolves a caret to its offset", () => {
    const document = seededDocument("hello world");
    const selection = textSelection(document, "body", 6, 6);

    expect(isCollapsed(selection)).toBe(true);
    expect(resolveSelection(document, selection)).toEqual({
      anchor: 6,
      head: 6,
    });
  });

  it("keeps a caret anchored across a concurrent insertion", () => {
    let document = seededDocument("hello world");
    // Caret on the "w" of "world".
    const selection = textSelection(document, "body", 6, 6);
    const before = resolveSelection(document, selection);
    expect(before).toEqual({ anchor: 6, head: 6 });

    // Someone types three characters at the start of the document.
    document = Automerge.change(document, (draft) => {
      Automerge.splice(draft, ["body"], 0, 0, "XX ");
    });

    // The same cursors now resolve three positions later: an integer offset of
    // 6 would point at the wrong character.
    const after = resolveSelection(document, selection);
    expect(after).toEqual({ anchor: before.anchor + 3, head: before.head + 3 });
  });

  it("carries a non-collapsed selection range", () => {
    const document = seededDocument("hello world");
    const selection = textSelection(document, "body", 0, 5);

    expect(isCollapsed(selection)).toBe(false);
    expect(resolveSelection(document, selection)).toEqual({
      anchor: 0,
      head: 5,
    });
  });
});

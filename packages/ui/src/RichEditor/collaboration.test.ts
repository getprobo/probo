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

import {
  createRichEditorAutomergeDocument,
  supportsRichEditorCollaboration,
} from "./RichEditor";

describe("RichEditor collaboration", () => {
  it("imports supported ProseMirror content into Automerge rich text", () => {
    const document = createRichEditorAutomergeDocument(
      JSON.stringify({
        type: "doc",
        content: [
          {
            type: "heading",
            attrs: { level: 2 },
            content: [{ type: "text", text: "Policy" }],
          },
          {
            type: "paragraph",
            content: [
              { type: "text", text: "Hello ", marks: [{ type: "bold" }] },
              { type: "text", text: "world" },
            ],
          },
        ],
      }),
    );

    expect(Automerge.spans(document, ["body"])).not.toHaveLength(0);
    expect(document.body).toContain("Hello world");
  });

  it("rejects schema nodes that cannot round-trip yet", () => {
    expect(
      supportsRichEditorCollaboration(
        JSON.stringify({
          type: "doc",
          content: [{ type: "table", content: [] }],
        }),
      ),
    ).toBe(false);
  });
});

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
import { describe, expect, it } from "vitest";

import { richEditorCollaborationExtensions, richEditorExtensions } from "./RichEditor";

describe("rich editor extensions", () => {
  it("keeps automerge attributes off the base editor schema", () => {
    const schema = getSchema(richEditorExtensions);

    expect(schema.nodes.paragraph.spec.attrs?.isAmgBlock).toBeUndefined();
    expect(schema.nodes.paragraph.spec.attrs?.unknownAttrs).toBeUndefined();
    expect(schema.nodes.table.spec.attrs?.isAmgBlock).toBeUndefined();
    expect(schema.nodes.horizontalRule.spec.attrs?.isAmgBlock).toBeUndefined();

    const doc = schema.nodeFromJSON({
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: [{ type: "text", text: "Hello" }],
        },
      ],
    });

    expect(doc.toJSON()).toEqual({
      type: "doc",
      content: [
        {
          type: "paragraph",
          content: [{ type: "text", text: "Hello" }],
        },
      ],
    });
  });

  it("keeps automerge attributes on the collaboration schema", () => {
    const schema = getSchema(richEditorCollaborationExtensions);

    expect(schema.nodes.paragraph.spec.attrs?.isAmgBlock).toBeDefined();
    expect(schema.nodes.table.spec.attrs?.isAmgBlock).toBeDefined();
  });
});

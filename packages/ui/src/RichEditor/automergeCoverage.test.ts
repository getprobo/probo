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

import { richEditorCollaborationExtensions } from "./RichEditor";

// Nodes that round-trip through collaboration without their own block mapping:
// the document root, the text leaf, and the list containers whose identity is
// carried by their items. contentUpload is a UI-only upload placeholder that is
// stripped before a document is saved.
const nodesWithoutAutomerge = new Set([
  "doc",
  "text",
  "bulletList",
  "orderedList",
  "taskList",
  "contentUpload",
]);

function automergeField(spec: unknown): unknown {
  return (spec as { automerge?: unknown }).automerge;
}

describe("automerge mapping coverage", () => {
  it("maps every collaborative content node onto an extension", () => {
    const schema = getSchema(richEditorCollaborationExtensions);

    const unmapped: string[] = [];
    schema.spec.nodes.forEach((name, spec) => {
      if (automergeField(spec) === undefined) {
        unmapped.push(name);
      }
    });

    expect(new Set(unmapped)).toEqual(nodesWithoutAutomerge);
  });

  it("maps every collaborative mark onto an extension", () => {
    const schema = getSchema(richEditorCollaborationExtensions);

    const unmapped: string[] = [];
    schema.spec.marks.forEach((name, spec) => {
      if (automergeField(spec) === undefined) {
        unmapped.push(name);
      }
    });

    expect(unmapped).toEqual([]);
  });
});

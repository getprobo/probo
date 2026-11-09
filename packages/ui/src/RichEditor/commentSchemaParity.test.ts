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

import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

import { getSchema } from "@tiptap/core";
import { describe, expect, it } from "vitest";

import { richEditorCommentExtensions } from "./RichEditor";

type CommentSchemaLedger = {
  nodes: string[];
  marks: string[];
};

const ledgerPath = join(
  dirname(fileURLToPath(import.meta.url)),
  "../../../../pkg/richtext/testdata/comment-schema.json",
);

describe("comment schema parity", () => {
  it("keeps richEditorCommentExtensions aligned with CommentSchema", () => {
    const ledger = JSON.parse(
      readFileSync(ledgerPath, "utf8"),
    ) as CommentSchemaLedger;
    const schema = getSchema(richEditorCommentExtensions);

    const nodes: string[] = [];
    schema.spec.nodes.forEach((name) => {
      nodes.push(name);
    });

    const marks: string[] = [];
    schema.spec.marks.forEach((name) => {
      marks.push(name);
    });

    expect(nodes.sort()).toEqual([...ledger.nodes].sort());
    expect(marks.sort()).toEqual([...ledger.marks].sort());
  });
});

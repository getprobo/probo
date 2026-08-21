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
import { pmDocFromSpans, pmNodeToSpans } from "@automerge/prosemirror";
import { bench, describe } from "vitest";

import {
  createRichEditorAutomergeDocument,
  createSchemaAdapter,
} from "./collaboration";
import { richEditorCollaborationExtensions } from "./RichEditor";

const adapter = createSchemaAdapter(richEditorCollaborationExtensions);

const paragraphDocumentJSON = {
  type: "doc",
  content: Array.from({ length: 1_000 }, () => ({
    type: "paragraph",
    attrs: { isAmgBlock: true },
    content: [{ type: "text", text: "a".repeat(100) }],
  })),
};
const paragraphDocument = adapter.schema.nodeFromJSON(paragraphDocumentJSON);
const paragraphAutomerge = createRichEditorAutomergeDocument(
  JSON.stringify(paragraphDocumentJSON),
  richEditorCollaborationExtensions,
);
const paragraphSpans = Automerge.spans(paragraphAutomerge, ["body"]);

const tableCell = (text: string) => ({
  type: "tableCell",
  attrs: {
    isAmgBlock: true,
    colspan: 1,
    rowspan: 1,
    colwidth: null,
  },
  content: [{
    type: "paragraph",
    content: [{ type: "text", text }],
  }],
});

const tableDocumentJSON = {
  type: "doc",
  content: [{
    type: "table",
    content: Array.from({ length: 100 }, (_, row) => ({
      type: "tableRow",
      content: Array.from({ length: 10 }, (_, column) =>
        tableCell(`${row}:${column}:${"a".repeat(20)}`),
      ),
    })),
  }],
};
const tableDocument = adapter.schema.nodeFromJSON(tableDocumentJSON);
const tableAutomerge = createRichEditorAutomergeDocument(
  JSON.stringify(tableDocumentJSON),
  richEditorCollaborationExtensions,
);
const tableSpans = Automerge.spans(tableAutomerge, ["body"]);

describe("ProseMirror bridge", () => {
  bench("pmNodeToSpans / 1,000 paragraphs / 100 KiB", () => {
    pmNodeToSpans(adapter, paragraphDocument);
  });

  bench("pmDocFromSpans / 1,000 paragraphs / 100 KiB", () => {
    pmDocFromSpans(adapter, paragraphSpans);
  });

  bench("pmNodeToSpans / table 100x10", () => {
    pmNodeToSpans(adapter, tableDocument);
  });

  bench("pmDocFromSpans / table 100x10", () => {
    pmDocFromSpans(adapter, tableSpans);
  });
});

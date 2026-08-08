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

import type { RichEditorAutomergeDocument } from "./collaboration";
import {
  createRichEditorAutomergeDocument,
  readRichEditorAutomergeDocument,
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

  it("round-trips table rows and cells with stable IDs", () => {
    const document = createRichEditorAutomergeDocument(tableJSON);
    const table = Object.values(document.tables)[0];

    expect(table.rowIDs).toHaveLength(2);
    expect(table.rowIDs[0]).not.toBe(table.rowIDs[1]);
    expect(table.rows[table.rowIDs[0]].cellIDs).toHaveLength(2);
    expect(table.rows[table.rowIDs[1]].cellIDs).toHaveLength(2);

    const content = readRichEditorAutomergeDocument(document) as {
      content: Array<{ type: string; content: unknown[] }>;
    };
    const roundTrippedTable = content.content[0];
    expect(roundTrippedTable.type).toBe("table");
    expect(roundTrippedTable.content).toHaveLength(2);
    expect(
      (roundTrippedTable.content[0] as { content: unknown[] }).content,
    ).toHaveLength(2);
    expect(
      (roundTrippedTable.content[1] as { content: unknown[] }).content,
    ).toHaveLength(2);
  });

  it("merges concurrent edits in different cells independently", () => {
    const base = createRichEditorAutomergeDocument(tableJSON);
    let left = Automerge.load<RichEditorAutomergeDocument>(
      Automerge.save(base),
      { actor: "aa".repeat(16) },
    );
    let right = Automerge.load<RichEditorAutomergeDocument>(
      Automerge.save(base),
      { actor: "bb".repeat(16) },
    );
    const table = Object.values(base.tables)[0];
    const firstRow = table.rows[table.rowIDs[0]];
    const secondRow = table.rows[table.rowIDs[1]];
    const firstCellID = firstRow.cellIDs[0];
    const lastCellID = secondRow.cellIDs[1];

    left = Automerge.change(left, (draft) => {
      Automerge.splice(
        draft,
        cellBodyPath(table.id, firstRow.id, firstCellID),
        1,
        0,
        "-left",
      );
    });
    right = Automerge.change(right, (draft) => {
      Automerge.splice(
        draft,
        cellBodyPath(table.id, secondRow.id, lastCellID),
        1,
        0,
        "-right",
      );
    });

    const merged = Automerge.merge(left, right);
    expect(merged.tables[table.id].rows[firstRow.id].cells[firstCellID].body)
      .toBe("A-left");
    expect(merged.tables[table.id].rows[secondRow.id].cells[lastCellID].body)
      .toBe("D-right");
    expect(merged.tables[table.id].rowIDs).toHaveLength(2);
  });

  it("accepts Tiptap table nodes for collaboration", () => {
    expect(supportsRichEditorCollaboration(tableJSON)).toBe(true);
  });
});

const tableJSON = JSON.stringify({
  type: "doc",
  content: [
    {
      type: "table",
      content: [
        {
          type: "tableRow",
          content: [
            tableCell("tableHeader", "A"),
            tableCell("tableHeader", "B"),
          ],
        },
        {
          type: "tableRow",
          content: [
            tableCell("tableCell", "C"),
            tableCell("tableCell", "D"),
          ],
        },
      ],
    },
  ],
});

function tableCell(type: "tableCell" | "tableHeader", text: string) {
  return {
    type,
    attrs: { colspan: 1, rowspan: 1, colwidth: null, align: null },
    content: [{
      type: "paragraph",
      content: [{ type: "text", text }],
    }],
  };
}

function cellBodyPath(tableID: string, rowID: string, cellID: string) {
  return [
    "tables",
    tableID,
    "rows",
    rowID,
    "cells",
    cellID,
    "body",
  ];
}

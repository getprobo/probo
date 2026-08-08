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
  automergeTableToTiptap,
  createAutomergeTableStore,
  insertAutomergeTableColumn,
  insertAutomergeTableRow,
  spliceAutomergeTableCellText,
} from "./automergeTable";

const tableJSON = {
  type: "table",
  attrs: { collaborationId: "table-1" },
  content: [{
    type: "tableRow",
    attrs: { collaborationId: "row-1" },
    content: [
      {
        type: "tableHeader",
        attrs: {
          collaborationId: "cell-1",
          colspan: 1,
          rowspan: 1,
          colwidth: [120],
        },
        content: [{
          type: "paragraph",
          content: [{ type: "text", text: "Owner" }],
        }],
      },
      {
        type: "tableCell",
        attrs: {
          collaborationId: "cell-2",
          colspan: 2,
          rowspan: 1,
          colwidth: null,
        },
        content: [{
          type: "paragraph",
          content: [{ type: "text", text: "Security" }],
        }],
      },
    ],
  }],
};

describe("Automerge tables", () => {
  it("round-trips TableKit JSON through maps, lists, and text", () => {
    const document = createAutomergeTableStore(tableJSON);

    expect(Automerge.getObjectId(document.tables)).toBeTruthy();
    expect(Automerge.getObjectId(document.tables["table-1"].rows)).toBeTruthy();
    expect(
      Automerge.spans(document, [
        "tables",
        "table-1",
        "rows",
        0,
        "cells",
        0,
        "text",
      ]),
    ).toEqual([{ type: "text", value: "Owner" }]);
    expect(automergeTableToTiptap(document, "table-1")).toEqual(tableJSON);
  });

  it("merges concurrent edits without replacing the table", () => {
    const base = createAutomergeTableStore(tableJSON);
    const left = spliceAutomergeTableCellText(
      Automerge.clone(base),
      "table-1",
      "cell-1",
      5,
      0,
      " team",
    );
    const right = spliceAutomergeTableCellText(
      Automerge.clone(base),
      "table-1",
      "cell-2",
      8,
      0,
      " review",
    );

    const merged = Automerge.merge(left, right);

    expect(merged.tables["table-1"].rows[0].cells[0].text).toBe("Owner team");
    expect(merged.tables["table-1"].rows[0].cells[1].text).toBe(
      "Security review",
    );
    expect(merged.tables["table-1"].id).toBe("table-1");
  });

  it("preserves concurrent row and column insertions with stable IDs", () => {
    const base = createAutomergeTableStore(tableJSON);
    const withRow = insertAutomergeTableRow(
      Automerge.clone(base),
      "table-1",
      1,
      idFactory("new-row", "new-row-cell-1", "new-row-cell-2"),
    );
    const withColumn = insertAutomergeTableColumn(
      Automerge.clone(base),
      "table-1",
      1,
      idFactory("new-column-cell"),
    );

    const merged = Automerge.merge(withRow, withColumn);
    const rows = merged.tables["table-1"].rows;

    expect(rows.map(row => row.id)).toEqual(["row-1", "new-row"]);
    expect(rows[0].cells.map(cell => cell.id)).toContain("new-column-cell");
    expect(rows[1].cells.map(cell => cell.id)).toEqual([
      "new-row-cell-1",
      "new-row-cell-2",
    ]);
  });

  it("rejects cell content that the plain-text slice cannot preserve", () => {
    expect(() => createAutomergeTableStore({
      ...tableJSON,
      content: [{
        ...tableJSON.content[0],
        content: [{
          ...tableJSON.content[0].content[0],
          content: [{
            type: "paragraph",
            content: [{
              type: "text",
              text: "Owner",
              marks: [{ type: "bold" }],
            }],
          }],
        }],
      }],
    })).toThrow("plain text only");
  });
});

function idFactory(...ids: string[]): () => string {
  return () => {
    const id = ids.shift();
    if (!id) throw new Error("test ID factory exhausted");
    return id;
  };
}

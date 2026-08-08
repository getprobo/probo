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

export const collaborationIDAttribute = "collaborationId";

export type AutomergeTableCell = {
  id: string;
  type: "tableCell" | "tableHeader";
  attrs: {
    colspan: number;
    rowspan: number;
    colwidth: number[] | null;
  };
  text: string;
};

export type AutomergeTableRow = {
  id: string;
  cells: AutomergeTableCell[];
};

export type AutomergeTable = {
  id: string;
  rows: AutomergeTableRow[];
};

export type AutomergeTableStore = {
  tables: Record<string, AutomergeTable>;
};

type TiptapNode = {
  type: string;
  attrs?: Record<string, unknown>;
  content?: TiptapNode[];
  marks?: unknown[];
  text?: string;
};

type IDFactory = () => string;

export function createAutomergeTableStore(
  tableNode: object,
  createID: IDFactory = () => crypto.randomUUID(),
): Automerge.Doc<AutomergeTableStore> {
  const table = readTableNode(tableNode, createID);
  return Automerge.from<AutomergeTableStore>({
    tables: { [table.id]: table },
  });
}

export function automergeTableToTiptap(
  document: Automerge.Doc<AutomergeTableStore>,
  tableID: string,
): object {
  const table = document.tables[tableID];
  if (!table) throw new Error(`unknown collaborative table: ${tableID}`);

  return {
    type: "table",
    attrs: { [collaborationIDAttribute]: table.id },
    content: table.rows.map(row => ({
      type: "tableRow",
      attrs: { [collaborationIDAttribute]: row.id },
      content: row.cells.map(cell => ({
        type: cell.type,
        attrs: {
          ...cell.attrs,
          [collaborationIDAttribute]: cell.id,
        },
        content: [{
          type: "paragraph",
          content: cell.text
            ? [{ type: "text", text: cell.text }]
            : undefined,
        }],
      })),
    })),
  };
}

export function spliceAutomergeTableCellText(
  document: Automerge.Doc<AutomergeTableStore>,
  tableID: string,
  cellID: string,
  index: number,
  deleteCount: number,
  value: string,
): Automerge.Doc<AutomergeTableStore> {
  return Automerge.change(document, (draft) => {
    const location = findCell(draft, tableID, cellID);
    Automerge.splice(
      draft,
      [
        "tables",
        tableID,
        "rows",
        location.rowIndex,
        "cells",
        location.cellIndex,
        "text",
      ],
      index,
      deleteCount,
      value,
    );
  });
}

export function insertAutomergeTableRow(
  document: Automerge.Doc<AutomergeTableStore>,
  tableID: string,
  index: number,
  createID: IDFactory = () => crypto.randomUUID(),
): Automerge.Doc<AutomergeTableStore> {
  return Automerge.change(document, (draft) => {
    const table = requireTable(draft, tableID);
    const reference = table.rows[Math.min(index, table.rows.length - 1)];
    if (!reference) throw new Error("cannot add a row to an empty table");

    table.rows.splice(index, 0, {
      id: createID(),
      cells: reference.cells.map(cell => ({
        id: createID(),
        type: cell.type,
        attrs: { ...cell.attrs, colwidth: cell.attrs.colwidth?.slice() ?? null },
        text: "",
      })),
    });
  });
}

export function insertAutomergeTableColumn(
  document: Automerge.Doc<AutomergeTableStore>,
  tableID: string,
  index: number,
  createID: IDFactory = () => crypto.randomUUID(),
): Automerge.Doc<AutomergeTableStore> {
  return Automerge.change(document, (draft) => {
    const table = requireTable(draft, tableID);
    for (const row of table.rows) {
      const reference = row.cells[Math.min(index, row.cells.length - 1)];
      if (!reference) throw new Error("cannot add a column to an empty row");
      row.cells.splice(index, 0, {
        id: createID(),
        type: reference.type,
        attrs: { colspan: 1, rowspan: 1, colwidth: null },
        text: "",
      });
    }
  });
}

function readTableNode(tableNode: object, createID: IDFactory): AutomergeTable {
  const node = asNode(tableNode);
  if (node.type !== "table" || !node.content?.length) {
    throw new Error("expected a non-empty Tiptap table");
  }

  return {
    id: readID(node, createID),
    rows: node.content.map(row => readRowNode(row, createID)),
  };
}

function readRowNode(node: TiptapNode, createID: IDFactory): AutomergeTableRow {
  if (node.type !== "tableRow" || !node.content?.length) {
    throw new Error("expected a non-empty Tiptap table row");
  }
  return {
    id: readID(node, createID),
    cells: node.content.map(cell => readCellNode(cell, createID)),
  };
}

function readCellNode(node: TiptapNode, createID: IDFactory): AutomergeTableCell {
  if (
    (node.type !== "tableCell" && node.type !== "tableHeader")
    || node.content?.length !== 1
    || node.content[0]?.type !== "paragraph"
  ) {
    throw new Error("collaborative cells currently require one paragraph");
  }
  const paragraph = node.content[0];
  const children = paragraph.content ?? [];
  const hasUnsupportedContent = children.some((child) => {
    return child.type !== "text"
      || typeof child.text !== "string"
      || Boolean(child.marks?.length);
  });
  if (hasUnsupportedContent) {
    throw new Error("collaborative cells currently support plain text only");
  }

  return {
    id: readID(node, createID),
    type: node.type,
    attrs: {
      colspan: readPositiveInteger(node.attrs?.colspan, 1),
      rowspan: readPositiveInteger(node.attrs?.rowspan, 1),
      colwidth: readColumnWidths(node.attrs?.colwidth),
    },
    text: children.map(child => child.text).join(""),
  };
}

function asNode(value: object): TiptapNode {
  if (!("type" in value) || typeof value.type !== "string") {
    throw new Error("expected a Tiptap node");
  }
  return value as TiptapNode;
}

function readID(node: TiptapNode, createID: IDFactory): string {
  const id = node.attrs?.[collaborationIDAttribute];
  return typeof id === "string" && id ? id : createID();
}

function readPositiveInteger(value: unknown, fallback: number): number {
  return typeof value === "number" && Number.isInteger(value) && value > 0
    ? value
    : fallback;
}

function readColumnWidths(value: unknown): number[] | null {
  return isColumnWidthArray(value) ? value : null;
}

function isColumnWidthArray(value: unknown): value is number[] {
  return Array.isArray(value)
    && value.every(
      (width: unknown) => typeof width === "number" && width > 0,
    );
}

function requireTable(
  document: AutomergeTableStore,
  tableID: string,
): AutomergeTable {
  const table = document.tables[tableID];
  if (!table) throw new Error(`unknown collaborative table: ${tableID}`);
  return table;
}

function findCell(
  document: AutomergeTableStore,
  tableID: string,
  cellID: string,
): { rowIndex: number; cellIndex: number } {
  const table = requireTable(document, tableID);
  for (const [rowIndex, row] of table.rows.entries()) {
    const cellIndex = row.cells.findIndex(cell => cell.id === cellID);
    if (cellIndex !== -1) return { rowIndex, cellIndex };
  }
  throw new Error(`unknown collaborative table cell: ${cellID}`);
}

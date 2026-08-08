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
import {
  type DocHandle,
  pmDocFromSpans,
  pmNodeToSpans,
  type SchemaAdapter,
} from "@automerge/prosemirror";
import type { Node as ProseMirrorNode } from "@tiptap/pm/model";
import { Plugin, Selection } from "@tiptap/pm/state";

const bodyPath: Automerge.Prop[] = ["body"];
const tableBlockType = "tiptap-table";
export const automergeIDAttribute = "automergeId";

type JSONAttributes = Record<string, unknown>;

export type TableCellState = {
  id: string;
  type: "tableCell" | "tableHeader";
  attrs: JSONAttributes;
  body: string;
};

export type TableRowState = {
  id: string;
  attrs: JSONAttributes;
  cellIDs: string[];
  cells: Record<string, TableCellState>;
};

export type TableState = {
  id: string;
  attrs: JSONAttributes;
  rowIDs: string[];
  rows: Record<string, TableRowState>;
};

export type TableRichEditorAutomergeDocument = {
  body: string;
  tables: Record<string, TableState>;
};

type IDFactory = () => string;

export function createTableRichEditorAutomergeDocument(
  adapter: SchemaAdapter,
  document: ProseMirrorNode,
  createID: IDFactory = () => crypto.randomUUID(),
): Automerge.Doc<TableRichEditorAutomergeDocument> {
  const result = Automerge.from<TableRichEditorAutomergeDocument>({
    body: "",
    tables: {},
  });

  return Automerge.change(result, (draft) => {
    updateTableRichEditorAutomergeDocument(
      adapter,
      draft,
      document,
      createID,
    );
  });
}

export function updateTableRichEditorAutomergeDocument(
  adapter: SchemaAdapter,
  draft: TableRichEditorAutomergeDocument,
  document: ProseMirrorNode,
  createID: IDFactory = () => crypto.randomUUID(),
): void {
  const existingTableIDs = tableIDsFromBody(draft);
  const nextTableIDs = new Set<string>();
  const spans: Automerge.Span[] = [];
  let tableIndex = 0;

  document.forEach((node) => {
    if (node.type.name !== "table") {
      const fragment = document.type.create(null, node);
      spans.push(...pmNodeToSpans(adapter, fragment));
      return;
    }

    const tableID = readID(node) ?? existingTableIDs[tableIndex] ?? createID();
    tableIndex++;
    nextTableIDs.add(tableID);
    reconcileTable(adapter, draft, tableID, node, createID);
    spans.push({
      type: "block",
      value: {
        type: tableBlockType,
        parents: [],
        attrs: { id: tableID },
        isEmbed: true,
      },
    });
  });

  Automerge.updateSpans(draft, bodyPath, spans, adapter.updateSpansConfig());
  for (const tableID of Object.keys(draft.tables)) {
    if (!nextTableIDs.has(tableID)) delete draft.tables[tableID];
  }
}

export function pmDocFromTableRichEditorAutomergeDocument(
  adapter: SchemaAdapter,
  document: Automerge.Doc<TableRichEditorAutomergeDocument>,
): ProseMirrorNode {
  const content: ProseMirrorNode[] = [];
  let regularSpans: Automerge.Span[] = [];

  function flushRegularSpans() {
    if (regularSpans.length === 0) return;
    pmDocFromSpans(adapter, regularSpans).forEach(node => content.push(node));
    regularSpans = [];
  }

  for (const span of Automerge.spans(document, bodyPath)) {
    const tableID = tableIDFromSpan(span);
    if (!tableID) {
      regularSpans.push(span);
      continue;
    }

    flushRegularSpans();
    const table = document.tables[tableID];
    if (table) content.push(tableNodeFromState(adapter, document, table));
  }
  flushRegularSpans();

  return adapter.schema.nodes.doc.create(null, content);
}

export function tableSyncPlugin(
  adapter: SchemaAdapter,
  handle: DocHandle<TableRichEditorAutomergeDocument>,
): Plugin {
  let applyingAutomerge = false;

  return new Plugin({
    view: view => {
      const onChange = () => {
        if (applyingAutomerge) return;
        const nextDocument = pmDocFromTableRichEditorAutomergeDocument(
          adapter,
          handle.doc(),
        );
        if (nextDocument.eq(view.state.doc)) return;

        applyingAutomerge = true;
        const transaction = view.state.tr.replaceWith(
          0,
          view.state.doc.content.size,
          nextDocument.content,
        );
        transaction.setMeta("addToHistory", false);
        view.dispatch(transaction);
        applyingAutomerge = false;
      };

      handle.on("change", onChange);
      return { destroy: () => handle.off("change", onChange) };
    },
    appendTransaction(transactions, _oldState, state) {
      if (applyingAutomerge || !transactions.some(item => item.docChanged)) {
        return undefined;
      }

      applyingAutomerge = true;
      handle.change((draft) => {
        updateTableRichEditorAutomergeDocument(adapter, draft, state.doc);
      });
      applyingAutomerge = false;

      const nextDocument = pmDocFromTableRichEditorAutomergeDocument(
        adapter,
        handle.doc(),
      );
      if (nextDocument.eq(state.doc)) return undefined;

      const transaction = state.tr.replaceWith(
        0,
        state.doc.content.size,
        nextDocument.content,
      );
      try {
        transaction.setSelection(
          Selection.fromJSON(transaction.doc, state.selection.toJSON()),
        );
      } catch (error) {
        if (!(error instanceof RangeError)) throw error;
      }
      transaction.setStoredMarks(state.storedMarks);
      return transaction;
    },
  });
}

function reconcileTable(
  adapter: SchemaAdapter,
  draft: TableRichEditorAutomergeDocument,
  tableID: string,
  node: ProseMirrorNode,
  createID: IDFactory,
): void {
  if (!draft.tables[tableID]) {
    draft.tables[tableID] = {
      id: tableID,
      attrs: {},
      rowIDs: [],
      rows: {},
    };
  }
  const table = draft.tables[tableID];
  table.attrs = attributesWithoutID(node);

  const nextRowIDs: string[] = [];
  node.forEach((rowNode, _offset, rowIndex) => {
    const rowID = readID(rowNode) ?? table.rowIDs[rowIndex] ?? createID();
    nextRowIDs.push(rowID);
    reconcileRow(adapter, draft, table, tableID, rowID, rowNode, createID);
  });
  reconcileIDList(table.rowIDs, nextRowIDs);

  const retainedRows = new Set(nextRowIDs);
  for (const rowID of Object.keys(table.rows)) {
    if (!retainedRows.has(rowID)) delete table.rows[rowID];
  }
}

function reconcileRow(
  adapter: SchemaAdapter,
  draft: TableRichEditorAutomergeDocument,
  table: TableState,
  tableID: string,
  rowID: string,
  node: ProseMirrorNode,
  createID: IDFactory,
): void {
  if (!table.rows[rowID]) {
    table.rows[rowID] = {
      id: rowID,
      attrs: {},
      cellIDs: [],
      cells: {},
    };
  }
  const row = table.rows[rowID];
  row.attrs = attributesWithoutID(node);

  const nextCellIDs: string[] = [];
  node.forEach((cellNode, _offset, cellIndex) => {
    const cellID = readID(cellNode) ?? row.cellIDs[cellIndex] ?? createID();
    nextCellIDs.push(cellID);
    reconcileCell(adapter, draft, tableID, row, rowID, cellID, cellNode);
  });
  reconcileIDList(row.cellIDs, nextCellIDs);

  const retainedCells = new Set(nextCellIDs);
  for (const cellID of Object.keys(row.cells)) {
    if (!retainedCells.has(cellID)) delete row.cells[cellID];
  }
}

function reconcileCell(
  adapter: SchemaAdapter,
  draft: TableRichEditorAutomergeDocument,
  tableID: string,
  row: TableRowState,
  rowID: string,
  cellID: string,
  node: ProseMirrorNode,
): void {
  if (!row.cells[cellID]) {
    row.cells[cellID] = {
      id: cellID,
      type: node.type.name as TableCellState["type"],
      attrs: {},
      body: "",
    };
  }
  const cell = row.cells[cellID];
  cell.type = node.type.name as TableCellState["type"];
  cell.attrs = attributesWithoutID(node);

  const cellDocument = adapter.schema.nodes.doc.create(null, node.content);
  const spans = pmNodeToSpans(adapter, cellDocument);
  Automerge.updateSpans(
    draft,
    ["tables", tableID, "rows", rowID, "cells", cellID, "body"],
    spans,
    adapter.updateSpansConfig(),
  );
}

function tableNodeFromState(
  adapter: SchemaAdapter,
  document: Automerge.Doc<TableRichEditorAutomergeDocument>,
  table: TableState,
): ProseMirrorNode {
  const rows = table.rowIDs.flatMap((rowID) => {
    const row = table.rows[rowID];
    if (!row) return [];
    const cells = row.cellIDs.flatMap((cellID) => {
      const cell = row.cells[cellID];
      if (!cell) return [];
      const content = pmDocFromSpans(
        adapter,
        Automerge.spans(
          document,
          [
            "tables",
            table.id,
            "rows",
            row.id,
            "cells",
            cell.id,
            "body",
          ],
        ),
      ).content;
      return adapter.schema.nodes[cell.type].create(
        { ...cell.attrs, [automergeIDAttribute]: cell.id },
        content,
      );
    });
    return adapter.schema.nodes.tableRow.create(
      { ...row.attrs, [automergeIDAttribute]: row.id },
      cells,
    );
  });

  return adapter.schema.nodes.table.create(
    { ...table.attrs, [automergeIDAttribute]: table.id },
    rows,
  );
}

function reconcileIDList(current: string[], next: string[]): void {
  let prefix = 0;
  while (prefix < current.length && current[prefix] === next[prefix]) prefix++;

  let suffix = 0;
  while (
    suffix < current.length - prefix
    && suffix < next.length - prefix
    && current[current.length - 1 - suffix] === next[next.length - 1 - suffix]
  ) {
    suffix++;
  }

  current.splice(
    prefix,
    current.length - prefix - suffix,
    ...next.slice(prefix, next.length - suffix),
  );
}

function tableIDsFromBody(
  document: TableRichEditorAutomergeDocument,
): string[] {
  return Automerge.spans(document, bodyPath)
    .map(tableIDFromSpan)
    .filter((value): value is string => value !== null);
}

function tableIDFromSpan(span: Automerge.Span): string | null {
  if (span.type !== "block") return null;
  if (immutableStringValue(span.value.type) !== tableBlockType) return null;
  const attributes = span.value.attrs;
  if (!attributes || typeof attributes !== "object" || Array.isArray(attributes)) {
    return null;
  }
  const id = (attributes as Record<string, unknown>).id;
  return typeof id === "string" ? id : null;
}

function immutableStringValue(value: unknown): string | null {
  if (typeof value === "string") return value;
  if (!value || typeof value !== "object") return null;
  const val = (value as { val?: unknown }).val;
  return typeof val === "string" ? val : null;
}

function readID(node: ProseMirrorNode): string | null {
  const value: unknown = node.attrs[automergeIDAttribute];
  return typeof value === "string" && value ? value : null;
}

function attributesWithoutID(node: ProseMirrorNode): JSONAttributes {
  const attributes = { ...node.attrs } as JSONAttributes;
  delete attributes[automergeIDAttribute];
  return attributes;
}

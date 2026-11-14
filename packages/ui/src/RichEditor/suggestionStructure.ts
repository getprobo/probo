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

import type { Node as ProseMirrorNode, NodeType } from "@tiptap/pm/model";
import { type EditorState, TextSelection, type Transaction } from "@tiptap/pm/state";
import { TableMap } from "@tiptap/pm/tables";

import {
  type SuggestionSplit,
  suggestionSplitAttr,
  type SuggestionSplitKind,
  suggestionSplitKindAttr,
} from "./suggestionSplit";

// Marks a transaction whose structural steps are already tracked as a
// suggestion, so the suggesting safety net lets it through.
export const suggestionStructureMeta = "suggestion-structure";

const tableStructureTypes = new Set(["table", "tableRow", "tableCell", "tableHeader"]);

export type TableTarget = {
  // Position of the table's first child, as table menus report it.
  tableStart: number;
};

export type TableRowTarget = TableTarget & { rowIndex: number };
export type TableColumnTarget = TableTarget & { columnIndex: number };

export function isTableStructureNode(node: ProseMirrorNode): boolean {
  return tableStructureTypes.has(node.type.name);
}

export function structuralSuggestionSplit(node: ProseMirrorNode): SuggestionSplit | null {
  const unknown = node.attrs.unknownAttrs as unknown;
  if (!unknown || typeof unknown !== "object" || Array.isArray(unknown)) return null;
  const values = unknown as Record<string, unknown>;
  const id = values[suggestionSplitAttr];
  const kind = values[suggestionSplitKindAttr];
  if (typeof id !== "string" || id === "") return null;
  if (kind !== "insertion" && kind !== "structural" && kind !== "deletion") return null;
  return { id, kind };
}

// structureSignature lists every node that is not text or a textblock, with
// its persisted attributes. Typing, Enter and Backspace only touch textblocks,
// so a change here is a structural edit (rows, columns, tables, images, lists).
export function structureSignature(doc: ProseMirrorNode): string {
  const parts: string[] = [];
  doc.descendants((node) => {
    if (node.isText || node.isTextblock) return node.isTextblock ? false : true;
    const attrs: Record<string, unknown> = {};
    for (const [name, value] of Object.entries(node.attrs)) {
      if (name === "isAmgBlock" || name === "unknownAttrs") continue;
      attrs[name] = value;
    }
    parts.push(`${node.type.name}${JSON.stringify(attrs)}/${node.childCount}`);
    return true;
  });
  return parts.join("|");
}

export function suggestInsertRow(
  state: EditorState,
  id: string,
  target: TableRowTarget,
  before: boolean,
): Transaction | null {
  const table = tableAt(state.doc, target.tableStart);
  if (!table || hasMergedCells(table)) return null;
  const map = TableMap.get(table);
  if (target.rowIndex < 0 || target.rowIndex >= map.height) return null;

  const at = before ? target.rowIndex : target.rowIndex + 1;
  const cells = newRowCellTypes(table, map, at).map(type =>
    insertedCell(type, id),
  );
  if (cells.some(cell => cell == null)) return null;
  const row = table.child(0).type.create(
    insertedAttrs(id),
    cells as ProseMirrorNode[],
  );

  let position = target.tableStart;
  for (let index = 0; index < at; index += 1) {
    position += table.child(index).nodeSize;
  }

  const tr = state.tr.insert(position, row);
  tr.setSelection(TextSelection.near(tr.doc.resolve(position + 1)));
  return tracked(tr);
}

export function suggestDeleteRow(
  state: EditorState,
  id: string,
  target: TableRowTarget,
): Transaction | null {
  const table = tableAt(state.doc, target.tableStart);
  if (!table || hasMergedCells(table)) return null;
  if (target.rowIndex < 0 || target.rowIndex >= table.childCount) return null;

  let remaining = 0;
  table.forEach((row, _offset, index) => {
    if (index !== target.rowIndex && structuralSuggestionSplit(row)?.kind !== "deletion") {
      remaining += 1;
    }
  });
  if (remaining === 0) return null;

  let position = target.tableStart;
  for (let index = 0; index < target.rowIndex; index += 1) {
    position += table.child(index).nodeSize;
  }

  const tr = state.tr;
  if (!markSubtreeDeleted(tr, position, id)) return null;
  return tracked(tr);
}

export function suggestInsertColumn(
  state: EditorState,
  id: string,
  target: TableColumnTarget,
  before: boolean,
): Transaction | null {
  const table = tableAt(state.doc, target.tableStart);
  if (!table || hasMergedCells(table)) return null;
  const map = TableMap.get(table);
  if (target.columnIndex < 0 || target.columnIndex >= map.width) return null;

  const at = before ? target.columnIndex : target.columnIndex + 1;
  const types = newColumnCellTypes(table, map, at);
  const tr = state.tr;
  for (let rowIndex = map.height - 1; rowIndex >= 0; rowIndex -= 1) {
    const cell = insertedCell(types[rowIndex], id);
    if (!cell) return null;
    const position = at < map.width
      ? target.tableStart + map.map[rowIndex * map.width + at]
      : target.tableStart + map.map[rowIndex * map.width + map.width - 1]
        + table.nodeAt(map.map[rowIndex * map.width + map.width - 1])!.nodeSize;
    tr.insert(position, cell);
  }

  const firstCell = at < map.width
    ? target.tableStart + map.map[at]
    : target.tableStart + map.map[map.width - 1] + table.nodeAt(map.map[map.width - 1])!.nodeSize;
  tr.setSelection(TextSelection.near(tr.doc.resolve(firstCell + 1)));
  return tracked(tr);
}

export function suggestDeleteColumn(
  state: EditorState,
  id: string,
  target: TableColumnTarget,
): Transaction | null {
  const table = tableAt(state.doc, target.tableStart);
  if (!table || hasMergedCells(table)) return null;
  const map = TableMap.get(table);
  if (target.columnIndex < 0 || target.columnIndex >= map.width) return null;

  let remaining = 0;
  for (let column = 0; column < map.width; column += 1) {
    if (column !== target.columnIndex && !columnDeleted(table, map, column)) remaining += 1;
  }
  if (remaining === 0) return null;

  const tr = state.tr;
  for (let rowIndex = 0; rowIndex < map.height; rowIndex += 1) {
    const position = target.tableStart + map.map[rowIndex * map.width + target.columnIndex];
    if (!markSubtreeDeleted(tr, position, id)) return null;
  }
  return tracked(tr);
}

export function suggestInsertTable(
  state: EditorState,
  id: string,
  rows: number,
  columns: number,
  headerRow: boolean,
): Transaction | null {
  const { schema } = state;
  const tableType = schema.nodes.table;
  const rowType = schema.nodes.tableRow;
  const cellType = schema.nodes.tableCell;
  const headerType = schema.nodes.tableHeader ?? cellType;
  if (!tableType || !rowType || !cellType || rows < 1 || columns < 1) return null;

  const tableRows: ProseMirrorNode[] = [];
  for (let rowIndex = 0; rowIndex < rows; rowIndex += 1) {
    const cells: ProseMirrorNode[] = [];
    for (let column = 0; column < columns; column += 1) {
      const cell = insertedCell(rowIndex === 0 && headerRow ? headerType : cellType, id);
      if (!cell) return null;
      cells.push(cell);
    }
    tableRows.push(rowType.create(insertedAttrs(id), cells));
  }
  const table = tableType.create(insertedAttrs(id), tableRows);

  const { $from } = state.selection;
  const position = $from.depth === 0 ? $from.pos : $from.after(1);
  const tr = state.tr.insert(position, table);
  tr.setSelection(TextSelection.near(tr.doc.resolve(position + 1)));
  return tracked(tr);
}

export function suggestDeleteTable(
  state: EditorState,
  id: string,
  target: TableTarget,
): Transaction | null {
  if (!tableAt(state.doc, target.tableStart)) return null;
  const tr = state.tr;
  if (!markSubtreeDeleted(tr, target.tableStart - 1, id)) return null;
  return tracked(tr);
}

// structuralSuggestionNodes lists the non-textblock nodes that carry a split
// for id. Text and textblock splits are handled by the text suggestion code.
export function structuralSuggestionNodes(
  doc: ProseMirrorNode,
  id: string,
): Array<{ pos: number; node: ProseMirrorNode; kind: SuggestionSplitKind }> {
  const found: Array<{ pos: number; node: ProseMirrorNode; kind: SuggestionSplitKind }> = [];
  doc.descendants((node, pos) => {
    if (node.isTextblock || node.isText) return !node.isText;
    const split = structuralSuggestionSplit(node);
    if (split?.id === id) found.push({ pos, node, kind: split.kind });
    return true;
  });
  return found;
}

// discardStructuralSuggestion removes nodes the suggestion inserted and
// clears the deletion split from nodes it marked for removal. Text inside
// those nodes is left to the caller.
export function discardStructuralSuggestion(tr: Transaction, id: string) {
  let insertedUntil = -1;
  const nodes = structuralSuggestionNodes(tr.doc, id).filter(({ pos, node, kind }) => {
    if (pos < insertedUntil) return false;
    if (kind === "insertion") insertedUntil = pos + node.nodeSize;
    return true;
  });
  const start = tr.steps.length;
  for (const { pos, node, kind } of nodes.reverse()) {
    const mapped = tr.mapping.slice(start).map(pos);
    if (kind === "insertion") {
      const current = tr.doc.nodeAt(mapped);
      if (current?.type === node.type) tr.delete(mapped, mapped + current.nodeSize);
      continue;
    }
    const current = tr.doc.nodeAt(mapped);
    if (!current) continue;
    tr.setNodeMarkup(mapped, undefined, {
      ...current.attrs,
      unknownAttrs: withoutSplit(current.attrs.unknownAttrs),
    });
  }
}

export function reassignStructuralSuggestion(tr: Transaction, fromId: string, toId: string) {
  for (const { pos, node, kind } of structuralSuggestionNodes(tr.doc, fromId)) {
    tr.setNodeMarkup(pos, undefined, {
      ...node.attrs,
      unknownAttrs: splitAttrs(node.attrs.unknownAttrs, toId, kind),
    });
  }
}

function tracked(tr: Transaction): Transaction {
  tr.setMeta(suggestionStructureMeta, true);
  return tr.scrollIntoView();
}

function tableAt(doc: ProseMirrorNode, tableStart: number): ProseMirrorNode | null {
  if (tableStart < 1 || tableStart > doc.content.size) return null;
  const table = doc.nodeAt(tableStart - 1);
  return table?.type.name === "table" ? table : null;
}

function hasMergedCells(table: ProseMirrorNode): boolean {
  let merged = false;
  table.descendants((node) => {
    if (merged) return false;
    if (node.type.name === "tableCell" || node.type.name === "tableHeader") {
      if ((node.attrs.colspan ?? 1) > 1 || (node.attrs.rowspan ?? 1) > 1) merged = true;
      return false;
    }
    return true;
  });
  return merged;
}

function cellAt(table: ProseMirrorNode, map: TableMap, row: number, column: number) {
  return table.nodeAt(map.map[row * map.width + column])!;
}

function rowIsHeader(table: ProseMirrorNode, map: TableMap, row: number): boolean {
  for (let column = 0; column < map.width; column += 1) {
    if (cellAt(table, map, row, column).type.name !== "tableHeader") return false;
  }
  return true;
}

function columnIsHeader(table: ProseMirrorNode, map: TableMap, column: number): boolean {
  for (let row = 0; row < map.height; row += 1) {
    if (cellAt(table, map, row, column).type.name !== "tableHeader") return false;
  }
  return true;
}

function columnDeleted(table: ProseMirrorNode, map: TableMap, column: number): boolean {
  for (let row = 0; row < map.height; row += 1) {
    if (structuralSuggestionSplit(cellAt(table, map, row, column))?.kind !== "deletion") {
      return false;
    }
  }
  return true;
}

// newRowCellTypes follows prosemirror-tables addRow, like the server's table
// redlines: copy the row above (or below at the top), except beside a header
// row at the table edge, where the new row holds plain cells.
function newRowCellTypes(table: ProseMirrorNode, map: TableMap, at: number): NodeType[] {
  const plainType = table.type.schema.nodes.tableCell;
  let reference = at > 0 ? at - 1 : 0;
  let plain = false;
  if (rowIsHeader(table, map, reference)) {
    if (at === 0 || at === map.height) plain = true;
    else reference = at;
  }
  const types: NodeType[] = [];
  for (let column = 0; column < map.width; column += 1) {
    types.push(plain ? plainType : cellAt(table, map, reference, column).type);
  }
  return types;
}

function newColumnCellTypes(table: ProseMirrorNode, map: TableMap, at: number): NodeType[] {
  const plainType = table.type.schema.nodes.tableCell;
  let reference = at > 0 ? at - 1 : 0;
  let plain = false;
  if (columnIsHeader(table, map, reference)) {
    if (at === 0 || at === map.width) plain = true;
    else reference = at;
  }
  const types: NodeType[] = [];
  for (let row = 0; row < map.height; row += 1) {
    types.push(plain ? plainType : cellAt(table, map, row, reference).type);
  }
  return types;
}

function insertedCell(type: NodeType, id: string): ProseMirrorNode | null {
  return type.createAndFill(insertedAttrs(id));
}

// Inserted structure must be its own Automerge block so the split reaches the
// document even before the block identity plugin flags it.
function insertedAttrs(id: string): Record<string, unknown> {
  return { isAmgBlock: true, unknownAttrs: splitAttrs(null, id, "insertion") };
}

// markSubtreeDeleted redlines the node at pos and everything inside it. Like
// the server, it refuses to swallow a change that belongs to a suggestion.
function markSubtreeDeleted(tr: Transaction, pos: number, id: string): boolean {
  const root = tr.doc.nodeAt(pos);
  if (!root) return false;
  const deleteMark = tr.doc.type.schema.marks.suggestionDelete?.create({ id });
  if (!deleteMark) return false;

  const blocks: Array<{ pos: number; node: ProseMirrorNode }> = [{ pos, node: root }];
  const texts: Array<{ from: number; to: number }> = [];
  let overlaps = structuralSuggestionSplit(root) != null;
  root.descendants((node, offset, parent, index) => {
    if (overlaps) return false;
    const nodePos = pos + 1 + offset;
    if (node.isText) {
      if (node.marks.some(mark =>
        mark.type.name === "suggestionInsert" || mark.type.name === "suggestionDelete",
      )) {
        overlaps = true;
      }
      texts.push({ from: nodePos, to: nodePos + node.nodeSize });
      return false;
    }
    if (structuralSuggestionSplit(node)) {
      overlaps = true;
      return false;
    }
    if (persistsAsBlock(node, parent, index)) blocks.push({ pos: nodePos, node });
    return true;
  });
  if (overlaps) return false;

  for (const text of texts) tr.addMark(text.from, text.to, deleteMark);
  for (const block of blocks) {
    if (!block.node.type.spec.attrs?.unknownAttrs) continue;
    tr.setNodeMarkup(block.pos, undefined, {
      ...block.node.attrs,
      unknownAttrs: splitAttrs(block.node.attrs.unknownAttrs, id, "deletion"),
    });
  }
  return true;
}

// The first paragraph of a cell (or list item) is folded into its parent's
// block marker, so attributes set on it would never reach the document.
function persistsAsBlock(
  node: ProseMirrorNode,
  parent: ProseMirrorNode | null,
  index: number,
): boolean {
  if (node.isInline) return false;
  if (node.attrs.isAmgBlock === true) return true;
  return !(parent && index === 0 && parent.type.contentMatch.defaultType === node.type);
}

function splitAttrs(
  current: unknown,
  id: string,
  kind: SuggestionSplitKind,
): Record<string, unknown> {
  return {
    ...attrRecord(current),
    [suggestionSplitAttr]: id,
    [suggestionSplitKindAttr]: kind,
  };
}

function withoutSplit(current: unknown): Record<string, unknown> | null {
  const next = attrRecord(current);
  delete next[suggestionSplitAttr];
  delete next[suggestionSplitKindAttr];
  return Object.keys(next).length > 0 ? next : null;
}

function attrRecord(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== "object" || Array.isArray(value)) return {};
  return { ...(value as Record<string, unknown>) };
}

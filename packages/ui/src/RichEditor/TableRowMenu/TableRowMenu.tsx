// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import type { Node as PMNode } from "@tiptap/pm/model";
import { TableMap } from "@tiptap/pm/tables";
import { type Editor } from "@tiptap/react";
import { useState } from "react";

import { cellDomElement } from "../_lib/cellDomElement";
import { useTableDropdownMenu } from "../_lib/useTableDropdownMenu";

import { TableRowMenuContent } from "./TableRowMenuContent";
import { TableRowMenuTrigger } from "./TableRowMenuTrigger";

export type HoveredRow = {
  rowIndex: number;
  tableStart: number;
};

export function getRowRect(
  editor: Editor,
  tableStart: number,
  rowIndex: number,
): DOMRect | null {
  try {
    const tableNodePos = tableStart - 1;
    const table = editor.state.doc.nodeAt(tableNodePos);
    if (!table) return null;

    const map = TableMap.get(table);
    if (rowIndex < 0 || rowIndex >= map.height) return null;

    const cellPos = map.positionAt(rowIndex, 0, table) + tableStart;
    const el = cellDomElement(editor, cellPos);
    if (!el) return null;

    const leftRect = el.getBoundingClientRect();
    let right = leftRect.right;

    if (map.width > 1) {
      const lastCellPos
        = map.positionAt(rowIndex, map.width - 1, table) + tableStart;
      const lastEl = cellDomElement(editor, lastCellPos);
      if (lastEl) {
        right = lastEl.getBoundingClientRect().right;
      }
    }

    return new DOMRect(
      leftRect.left,
      leftRect.top,
      right - leftRect.left,
      leftRect.height,
    );
  } catch {
    return null;
  }
}

export function moveRow(
  editor: Editor,
  tableStart: number,
  fromRow: number,
  toRow: number,
) {
  if (fromRow === toRow) return;

  const tableNodePos = tableStart - 1;
  const table = editor.state.doc.nodeAt(tableNodePos);
  if (!table) return;

  const rows: PMNode[] = [];
  table.forEach(row => rows.push(row));
  const [moved] = rows.splice(fromRow, 1);
  rows.splice(toRow, 0, moved);

  const newTable = table.type.create(table.attrs, rows);
  const { tr } = editor.state;
  tr.replaceWith(tableNodePos, tableNodePos + table.nodeSize, newTable);
  editor.view.dispatch(tr);
}

type TableRowMenuProps = {
  editor: Editor;
};

export function TableRowMenu({ editor }: TableRowMenuProps) {
  const {
    menuOpen,
    setMenuOpen,
    setTriggerEl,
    setDropdownEl,
    menuRefs,
    menuStyles,
    getFloatingProps,
  } = useTableDropdownMenu();

  const [hoveredRow, setHoveredRow] = useState<HoveredRow | null>(null);

  return (
    <>
      <TableRowMenuTrigger
        editor={editor}
        hoveredRow={hoveredRow}
        setHoveredRow={setHoveredRow}
        menuOpen={menuOpen}
        setMenuOpen={setMenuOpen}
        setTriggerEl={setTriggerEl}
        menuRefs={menuRefs}
      />
      {menuOpen && hoveredRow && (
        <TableRowMenuContent
          editor={editor}
          hoveredRow={hoveredRow}
          setMenuOpen={setMenuOpen}
          setHoveredRow={setHoveredRow}
          setDropdownEl={setDropdownEl}
          menuRefs={menuRefs}
          menuStyles={menuStyles}
          getFloatingProps={getFloatingProps}
        />
      )}
    </>
  );
}

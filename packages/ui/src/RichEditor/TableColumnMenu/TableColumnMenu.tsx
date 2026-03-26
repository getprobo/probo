// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import type { Node as PMNode } from "@tiptap/pm/model";
import { TableMap } from "@tiptap/pm/tables";
import { type Editor } from "@tiptap/react";
import { useState } from "react";

import { cellDomElement } from "../_lib/cellDomElement";
import { useTableDropdownMenu } from "../_lib/useTableDropdownMenu";

import { TableColumnMenuContent } from "./TableColumnMenuContent";
import { TableColumnMenuTrigger } from "./TableColumnMenuTrigger";

export type HoveredColumn = {
  colIndex: number;
  tableStart: number;
};

export function getColumnRect(
  editor: Editor,
  tableStart: number,
  colIndex: number,
): DOMRect | null {
  try {
    const tableNodePos = tableStart - 1;
    const table = editor.state.doc.nodeAt(tableNodePos);
    if (!table) return null;

    const map = TableMap.get(table);
    if (colIndex < 0 || colIndex >= map.width) return null;

    const cellPos = map.positionAt(0, colIndex, table) + tableStart;
    const el = cellDomElement(editor, cellPos);
    if (!el) return null;

    const topRect = el.getBoundingClientRect();
    let bottom = topRect.bottom;

    if (map.height > 1) {
      const lastCellPos
        = map.positionAt(map.height - 1, colIndex, table) + tableStart;
      const lastEl = cellDomElement(editor, lastCellPos);
      if (lastEl) {
        bottom = lastEl.getBoundingClientRect().bottom;
      }
    }

    return new DOMRect(
      topRect.left,
      topRect.top,
      topRect.width,
      bottom - topRect.top,
    );
  } catch {
    return null;
  }
}

export function moveColumn(
  editor: Editor,
  tableStart: number,
  fromCol: number,
  toCol: number,
) {
  if (fromCol === toCol) return;

  const tableNodePos = tableStart - 1;
  const table = editor.state.doc.nodeAt(tableNodePos);
  if (!table) return;

  const rows: PMNode[] = [];
  table.forEach((row) => {
    const cells: PMNode[] = [];
    row.forEach(cell => cells.push(cell));
    const [moved] = cells.splice(fromCol, 1);
    cells.splice(toCol, 0, moved);
    rows.push(row.type.create(row.attrs, cells));
  });

  const newTable = table.type.create(table.attrs, rows);
  const { tr } = editor.state;
  tr.replaceWith(tableNodePos, tableNodePos + table.nodeSize, newTable);
  editor.view.dispatch(tr);
}

type TableColumnMenuProps = {
  editor: Editor;
};

export function TableColumnMenu({ editor }: TableColumnMenuProps) {
  const {
    menuOpen,
    setMenuOpen,
    setTriggerEl,
    setDropdownEl,
    menuRefs,
    menuStyles,
    getFloatingProps,
  } = useTableDropdownMenu();

  const [hoveredCol, setHoveredCol] = useState<HoveredColumn | null>(null);

  if (!hoveredCol && !menuOpen) return null;

  return (
    <>
      <TableColumnMenuTrigger
        editor={editor}
        hoveredCol={hoveredCol}
        setHoveredCol={setHoveredCol}
        menuOpen={menuOpen}
        setMenuOpen={setMenuOpen}
        setTriggerEl={setTriggerEl}
        menuRefs={menuRefs}
      />
      {menuOpen && hoveredCol && (
        <TableColumnMenuContent
          editor={editor}
          hoveredCol={hoveredCol}
          setMenuOpen={setMenuOpen}
          setHoveredCol={setHoveredCol}
          setDropdownEl={setDropdownEl}
          menuRefs={menuRefs}
          menuStyles={menuStyles}
          getFloatingProps={getFloatingProps}
        />
      )}
    </>
  );
}

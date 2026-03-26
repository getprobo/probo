// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import {
  BroomIcon,
  CopyIcon,
  CrownSimpleIcon,
  PlusIcon,
  TrashIcon,
} from "@phosphor-icons/react";
import { TextSelection } from "@tiptap/pm/state";
import { CellSelection, TableMap } from "@tiptap/pm/tables";
import { type Editor } from "@tiptap/react";

import type { DropdownMenu } from "../_lib/useTableDropdownMenu";
import { MenuButton } from "../MenuButton";

import { type HoveredColumn } from "./TableColumnMenu";
import { tableColumnMenuVariants } from "./variants";

const { menu } = tableColumnMenuVariants();

type TableColumnMenuContentProps = {
  editor: Editor;
  hoveredCol: HoveredColumn;
  setMenuOpen: DropdownMenu["setMenuOpen"];
  setHoveredCol: React.Dispatch<React.SetStateAction<HoveredColumn | null>>;
  setDropdownEl: DropdownMenu["setDropdownEl"];
  menuRefs: DropdownMenu["menuRefs"];
  menuStyles: DropdownMenu["menuStyles"];
  getFloatingProps: DropdownMenu["getFloatingProps"];
};

export function TableColumnMenuContent({
  editor,
  hoveredCol,
  setMenuOpen,
  setHoveredCol,
  setDropdownEl,
  menuRefs,
  menuStyles,
  getFloatingProps,
}: TableColumnMenuContentProps) {
  const currentCol = hoveredCol;
  const isFirstColumn = currentCol.colIndex === 0;

  const isHeaderColumn = (): boolean => {
    if (currentCol.colIndex !== 0) return false;
    try {
      const table = editor.state.doc.nodeAt(currentCol.tableStart - 1);
      if (!table) return false;
      const map = TableMap.get(table);
      for (let row = 0; row < map.height; row++) {
        const cellPos = map.map[row * map.width] + currentCol.tableStart;
        const cellNode = editor.state.doc.nodeAt(cellPos);
        if (!cellNode || cellNode.type.name !== "tableHeader") return false;
      }
      return true;
    } catch {
      return false;
    }
  };

  const handleToggleHeaderColumn = () => {
    if (currentCol.colIndex !== 0) return;
    const { tableStart } = currentCol;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const map = TableMap.get(table);
      const cellPos = map.positionAt(0, 0, table) + tableStart;

      editor
        .chain()
        .focus()
        .command(({ tr }) => {
          tr.setSelection(TextSelection.create(tr.doc, cellPos + 1));
          return true;
        })
        .toggleHeaderColumn()
        .run();
    } catch {
      // table may have changed
    }

    setMenuOpen(false);
  };

  const handleDeleteColumn = () => {
    const { colIndex, tableStart } = currentCol;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const map = TableMap.get(table);
      const cellPos = map.positionAt(0, colIndex, table) + tableStart;

      editor
        .chain()
        .focus()
        .command(({ tr }) => {
          tr.setSelection(TextSelection.create(tr.doc, cellPos + 1));
          return true;
        })
        .deleteColumn()
        .run();
    } catch {
      // table may have changed
    }

    setMenuOpen(false);
    setHoveredCol(null);
  };

  const handleDuplicateColumn = () => {
    const { colIndex, tableStart } = currentCol;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const map = TableMap.get(table);

      editor
        .chain()
        .focus()
        .command(({ tr }) => {
          for (let row = map.height - 1; row >= 0; row--) {
            const cellOffset = map.map[row * map.width + colIndex];
            const cell = table.nodeAt(cellOffset);
            if (!cell) continue;

            const insertPos = cellOffset + tableStart + cell.nodeSize;
            tr.insert(insertPos, cell);
          }
          return true;
        })
        .run();
    } catch {
      // table may have changed
    }

    setMenuOpen(false);
  };

  const handleInsertLeft = () => {
    const { colIndex, tableStart } = currentCol;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const map = TableMap.get(table);
      const cellPos = map.positionAt(0, colIndex, table) + tableStart;

      editor
        .chain()
        .focus()
        .command(({ tr }) => {
          tr.setSelection(TextSelection.create(tr.doc, cellPos + 1));
          return true;
        })
        .addColumnBefore()
        .run();
    } catch {
      // table may have changed
    }

    setMenuOpen(false);
  };

  const handleInsertRight = () => {
    const { colIndex, tableStart } = currentCol;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const map = TableMap.get(table);
      const cellPos = map.positionAt(0, colIndex, table) + tableStart;

      editor
        .chain()
        .focus()
        .command(({ tr }) => {
          tr.setSelection(TextSelection.create(tr.doc, cellPos + 1));
          return true;
        })
        .addColumnAfter()
        .run();
    } catch {
      // table may have changed
    }

    setMenuOpen(false);
  };

  const handleClearContents = () => {
    const { colIndex, tableStart } = currentCol;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const map = TableMap.get(table);
      const firstCellPos = map.map[colIndex] + tableStart;
      const lastCellPos
        = map.map[(map.height - 1) * map.width + colIndex] + tableStart;

      const $anchor = editor.state.doc.resolve(firstCellPos);
      const $head = editor.state.doc.resolve(lastCellPos);

      editor.view.dispatch(
        editor.state.tr.setSelection(new CellSelection($anchor, $head)),
      );
      editor.commands.deleteSelection();
    } catch {
      // table may have changed
    }

    setMenuOpen(false);
  };

  return (
    <div
      ref={(node) => {
        setDropdownEl(node);
        menuRefs.setFloating(node);
      }}
      data-column-menu
      style={menuStyles}
      {...getFloatingProps()}
      onMouseDown={e => e.preventDefault()}
      className={menu()}
    >
      {isFirstColumn && (
        <MenuButton active={isHeaderColumn()} onClick={handleToggleHeaderColumn}>
          <CrownSimpleIcon size={16} weight="bold" />
          Header column
        </MenuButton>
      )}
      <MenuButton onClick={handleInsertLeft}>
        <PlusIcon size={16} weight="bold" />
        Insert column left
      </MenuButton>
      <MenuButton onClick={handleInsertRight}>
        <PlusIcon size={16} weight="bold" />
        Insert column right
      </MenuButton>
      <MenuButton onClick={handleDuplicateColumn}>
        <CopyIcon size={16} weight="bold" />
        Duplicate column
      </MenuButton>
      <MenuButton onClick={handleClearContents}>
        <BroomIcon size={16} weight="bold" />
        Clear contents
      </MenuButton>
      <MenuButton onClick={handleDeleteColumn}>
        <TrashIcon size={16} weight="bold" />
        Delete column
      </MenuButton>
    </div>
  );
}

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

import { type HoveredRow } from "./TableRowMenu";
import { tableRowMenuVariants } from "./variants";

const { menu } = tableRowMenuVariants();

type TableRowMenuContentProps = {
  editor: Editor;
  hoveredRow: HoveredRow;
  setMenuOpen: DropdownMenu["setMenuOpen"];
  setHoveredRow: React.Dispatch<React.SetStateAction<HoveredRow | null>>;
  setDropdownEl: DropdownMenu["setDropdownEl"];
  menuRefs: DropdownMenu["menuRefs"];
  menuStyles: DropdownMenu["menuStyles"];
  getFloatingProps: DropdownMenu["getFloatingProps"];
};

export function TableRowMenuContent({
  editor,
  hoveredRow,
  setMenuOpen,
  setHoveredRow,
  setDropdownEl,
  menuRefs,
  menuStyles,
  getFloatingProps,
}: TableRowMenuContentProps) {
  const currentRow = hoveredRow;
  const isFirstRow = currentRow.rowIndex === 0;

  const isHeaderRow = (): boolean => {
    if (currentRow.rowIndex !== 0) return false;
    try {
      const table = editor.state.doc.nodeAt(currentRow.tableStart - 1);
      if (!table) return false;
      const map = TableMap.get(table);
      for (let col = 0; col < map.width; col++) {
        const cellPos = map.map[col] + currentRow.tableStart;
        const cellNode = editor.state.doc.nodeAt(cellPos);
        if (!cellNode || cellNode.type.name !== "tableHeader") return false;
      }
      return true;
    } catch {
      return false;
    }
  };

  const handleToggleHeaderRow = () => {
    if (currentRow.rowIndex !== 0) return;
    const { tableStart } = currentRow;

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
        .toggleHeaderRow()
        .run();
    } catch {
      // table may have changed
    }

    setMenuOpen(false);
  };

  const handleDeleteRow = () => {
    const { rowIndex, tableStart } = currentRow;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const map = TableMap.get(table);
      const cellPos = map.positionAt(rowIndex, 0, table) + tableStart;

      editor
        .chain()
        .focus()
        .command(({ tr }) => {
          tr.setSelection(TextSelection.create(tr.doc, cellPos + 1));
          return true;
        })
        .deleteRow()
        .run();
    } catch {
      // table may have changed
    }

    setMenuOpen(false);
    setHoveredRow(null);
  };

  const handleDuplicateRow = () => {
    const { rowIndex, tableStart } = currentRow;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const rowNode = table.child(rowIndex);

      let insertPos = tableStart;
      for (let i = 0; i <= rowIndex; i++) {
        insertPos += table.child(i).nodeSize;
      }

      editor
        .chain()
        .focus()
        .command(({ tr }) => {
          tr.insert(insertPos, rowNode);
          return true;
        })
        .run();
    } catch {
      // table may have changed
    }

    setMenuOpen(false);
  };

  const handleInsertAbove = () => {
    const { rowIndex, tableStart } = currentRow;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const map = TableMap.get(table);
      const cellPos = map.positionAt(rowIndex, 0, table) + tableStart;

      editor
        .chain()
        .focus()
        .command(({ tr }) => {
          tr.setSelection(TextSelection.create(tr.doc, cellPos + 1));
          return true;
        })
        .addRowBefore()
        .run();
    } catch {
      // table may have changed
    }

    setMenuOpen(false);
  };

  const handleInsertBelow = () => {
    const { rowIndex, tableStart } = currentRow;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const map = TableMap.get(table);
      const cellPos = map.positionAt(rowIndex, 0, table) + tableStart;

      editor
        .chain()
        .focus()
        .command(({ tr }) => {
          tr.setSelection(TextSelection.create(tr.doc, cellPos + 1));
          return true;
        })
        .addRowAfter()
        .run();
    } catch {
      // table may have changed
    }

    setMenuOpen(false);
  };

  const handleClearContents = () => {
    const { rowIndex, tableStart } = currentRow;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const map = TableMap.get(table);
      const firstCellPos = map.map[rowIndex * map.width] + tableStart;
      const lastCellPos
        = map.map[rowIndex * map.width + (map.width - 1)] + tableStart;

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
      data-row-menu
      style={menuStyles}
      {...getFloatingProps()}
      onMouseDown={e => e.preventDefault()}
      className={menu()}
    >
      {isFirstRow && (
        <MenuButton active={isHeaderRow()} onClick={handleToggleHeaderRow}>
          <CrownSimpleIcon size={16} weight="bold" />
          Header row
        </MenuButton>
      )}
      <MenuButton onClick={handleInsertAbove}>
        <PlusIcon size={16} weight="bold" />
        Insert row above
      </MenuButton>
      <MenuButton onClick={handleInsertBelow}>
        <PlusIcon size={16} weight="bold" />
        Insert row below
      </MenuButton>
      <MenuButton onClick={handleDuplicateRow}>
        <CopyIcon size={16} weight="bold" />
        Duplicate row
      </MenuButton>
      <MenuButton onClick={handleClearContents}>
        <BroomIcon size={16} weight="bold" />
        Clear contents
      </MenuButton>
      <MenuButton onClick={handleDeleteRow}>
        <TrashIcon size={16} weight="bold" />
        Delete row
      </MenuButton>
    </div>
  );
}

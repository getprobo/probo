// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { BroomIcon, IntersectIcon, SplitHorizontalIcon } from "@phosphor-icons/react";
import { TextSelection } from "@tiptap/pm/state";
import { cellAround, CellSelection } from "@tiptap/pm/tables";
import { type Editor } from "@tiptap/react";

import type { DropdownMenu } from "../_lib/useTableDropdownMenu";
import { MenuButton } from "../MenuButton";

import { tableCellMenuVariants } from "./variants";

const { menu } = tableCellMenuVariants();

type TableCellMenuContentProps = {
  editor: Editor;
  setMenuOpen: DropdownMenu["setMenuOpen"];
  setDropdownEl: DropdownMenu["setDropdownEl"];
  menuRefs: DropdownMenu["menuRefs"];
  menuStyles: DropdownMenu["menuStyles"];
  getFloatingProps: DropdownMenu["getFloatingProps"];
};

export function TableCellMenuContent({
  editor,
  setMenuOpen,
  setDropdownEl,
  menuRefs,
  menuStyles,
  getFloatingProps,
}: TableCellMenuContentProps) {
  const handleMergeCells = () => {
    editor.chain().focus().mergeCells().run();
    setMenuOpen(false);
  };

  const handleSplitCell = () => {
    editor.chain().focus().splitCell().run();
    setMenuOpen(false);
  };

  const handleClearContents = () => {
    const { state } = editor.view;

    if (state.selection instanceof CellSelection) {
      editor.commands.deleteSelection();
    } else {
      const { dispatch } = editor.view;
      const { tr, schema } = state;
      const $pos = state.doc.resolve(state.selection.from);
      const cell = cellAround($pos);
      if (cell) {
        const cellNode = state.doc.nodeAt(cell.pos);
        if (cellNode) {
          const start = cell.pos + 1;
          const end = cell.pos + cellNode.nodeSize - 1;
          tr.replaceWith(start, end, schema.nodes.paragraph.create());
          tr.setSelection(TextSelection.create(tr.doc, start + 1));
          dispatch(tr);
        }
      }
    }

    setMenuOpen(false);
  };

  return (
    <div
      ref={(node) => {
        setDropdownEl(node);
        menuRefs.setFloating(node);
      }}
      style={menuStyles}
      {...getFloatingProps()}
      onMouseDown={e => e.preventDefault()}
      className={menu()}
    >
      {editor.can().mergeCells() && (
        <MenuButton onClick={handleMergeCells}>
          <IntersectIcon size={16} weight="bold" />
          Merge cells
        </MenuButton>
      )}
      {editor.can().splitCell() && (
        <MenuButton onClick={handleSplitCell}>
          <SplitHorizontalIcon size={16} weight="bold" />
          Split cells
        </MenuButton>
      )}
      <MenuButton onClick={handleClearContents}>
        <BroomIcon size={16} weight="bold" />
        Clear contents
      </MenuButton>
    </div>
  );
}

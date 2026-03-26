// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { cellAround, CellSelection } from "@tiptap/pm/tables";
import { type Editor, useEditorState } from "@tiptap/react";

import { cellDomElement } from "../_lib/cellDomElement";
import { useTableDropdownMenu } from "../_lib/useTableDropdownMenu";

import { TableCellMenuContent } from "./TableCellMenuContent";
import { TableCellMenuTrigger } from "./TableCellMenuTrigger";

type TableCellMenuProps = {
  editor: Editor;
};

function getActiveCellEl(editor: Editor): HTMLElement | null {
  const { selection } = editor.state;

  if (selection instanceof CellSelection) {
    return cellDomElement(editor, selection.$headCell.pos);
  }

  const $pos = editor.state.doc.resolve(selection.from);
  const cell = cellAround($pos);
  if (!cell) return null;
  return cellDomElement(editor, cell.pos);
}

export function TableCellMenu({ editor }: TableCellMenuProps) {
  const {
    menuOpen,
    setMenuOpen,
    setTriggerEl,
    setDropdownEl,
    menuRefs,
    menuStyles,
    getFloatingProps,
  } = useTableDropdownMenu();

  const activeCellEl = useEditorState({
    editor,
    selector: ({ editor: e }) => {
      if (e.isDestroyed || !e.isEditable) return null;
      return getActiveCellEl(e);
    },
  });

  if (!activeCellEl) return null;

  return (
    <>
      <TableCellMenuTrigger
        editor={editor}
        activeCellEl={activeCellEl}
        menuOpen={menuOpen}
        setMenuOpen={setMenuOpen}
        setTriggerEl={setTriggerEl}
        menuRefs={menuRefs}
      />
      {menuOpen && (
        <TableCellMenuContent
          editor={editor}
          setMenuOpen={setMenuOpen}
          setDropdownEl={setDropdownEl}
          menuRefs={menuRefs}
          menuStyles={menuStyles}
          getFloatingProps={getFloatingProps}
        />
      )}
    </>
  );
}

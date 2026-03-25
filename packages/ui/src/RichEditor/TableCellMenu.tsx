// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { autoUpdate, offset, useFloating } from "@floating-ui/react";
import { BroomIcon, CircleIcon, DotsThreeCircleVerticalIcon, IntersectIcon, SplitHorizontalIcon } from "@phosphor-icons/react";
import { TextSelection } from "@tiptap/pm/state";
import { cellAround, CellSelection, TableMap } from "@tiptap/pm/tables";
import { type Editor, useEditorState } from "@tiptap/react";
import { useLayoutEffect, useRef, useState } from "react";
import { tv } from "tailwind-variants";

import { cellDomElement } from "./_lib/cellDomElement";
import { DRAG_THRESHOLD } from "./_lib/constants";
import { useTableDropdownMenu } from "./_lib/useTableDropdownMenu";
import { MenuButton } from "./MenuButton";

const tableCellMenuVariants = tv({
  slots: {
    trigger: [
      "z-10 flex size-5 items-center justify-center",
      "rounded text-border-info cursor-pointer",
    ],
    menu: ["rounded-lg border border-border-mid bg-level-0 p-1 shadow-md z-20"],
  },
});

const { trigger, menu } = tableCellMenuVariants();

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

  const [handleHovered, setHandleHovered] = useState(false);
  const draggingRef = useRef(false);
  const dragStartPos = useRef({ x: 0, y: 0 });
  const anchorCellPosRef = useRef<number | null>(null);
  const selectionBoundsRef = useRef<{
    bottomRow: number;
    tableStart: number;
  } | null>(null);

  const activeCellEl = useEditorState({
    editor,
    selector: ({ editor: e }) => {
      if (e.isDestroyed || !e.isEditable) return null;
      return getActiveCellEl(e);
    },
  });

  const {
    refs: handleRefs,
    floatingStyles: handleStyles,
    isPositioned,
  } = useFloating({
    strategy: "fixed",
    placement: "right",
    middleware: [offset(-10)],
    whileElementsMounted: (ref, floating, update) =>
      autoUpdate(ref, floating, update, { animationFrame: true }),
  });

  useLayoutEffect(() => {
    if (!activeCellEl) {
      handleRefs.setReference(null);
      return;
    }

    const ed = editor;
    const fallback = activeCellEl;

    handleRefs.setReference({
      getBoundingClientRect() {
        const { selection } = ed.state;
        if (selection instanceof CellSelection) {
          let top = Infinity;
          let left = Infinity;
          let bottom = -Infinity;
          let right = -Infinity;
          selection.forEachCell((_node, pos) => {
            const el = cellDomElement(ed, pos);
            if (!el) return;
            const rect = el.getBoundingClientRect();
            top = Math.min(top, rect.top);
            left = Math.min(left, rect.left);
            bottom = Math.max(bottom, rect.bottom);
            right = Math.max(right, rect.right);
          });
          if (top !== Infinity) {
            return new DOMRect(left, top, right - left, bottom - top);
          }
        }
        return fallback.getBoundingClientRect();
      },
    });
  }, [activeCellEl, editor, handleRefs]);

  if (!activeCellEl) return null;

  const getAnchorCellPos = (): number | null => {
    try {
      const { selection, doc } = editor.state;
      const $pos = doc.resolve(selection.from);
      const cell = cellAround($pos);
      return cell ? cell.pos : null;
    } catch {
      return null;
    }
  };

  const onHandleMouseDown = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();

    if (menuOpen) return;

    const { selection } = editor.state;

    if (selection instanceof CellSelection) {
      try {
        const table = selection.$anchorCell.node(-1);
        const map = TableMap.get(table);
        const tableStart = selection.$anchorCell.start(-1);
        const anchorRect = map.findCell(selection.$anchorCell.pos - tableStart);
        const headRect = map.findCell(selection.$headCell.pos - tableStart);

        const topRow = Math.min(anchorRect.top, headRect.top);
        const bottomRow = Math.max(anchorRect.bottom, headRect.bottom) - 1;
        const leftCol = Math.min(anchorRect.left, headRect.left);

        anchorCellPosRef.current = map.positionAt(topRow, leftCol, table) + tableStart;

        selectionBoundsRef.current = topRow !== bottomRow
          ? { bottomRow, tableStart }
          : null;
      } catch {
        anchorCellPosRef.current = getAnchorCellPos();
        selectionBoundsRef.current = null;
      }
    } else {
      anchorCellPosRef.current = getAnchorCellPos();
      selectionBoundsRef.current = null;
    }

    if (anchorCellPosRef.current == null) return;

    draggingRef.current = false;
    dragStartPos.current = { x: e.clientX, y: e.clientY };

    const view = editor.view;

    const onMouseMove = (ev: MouseEvent) => {
      const dx = ev.clientX - dragStartPos.current.x;
      const dy = ev.clientY - dragStartPos.current.y;
      if (!draggingRef.current && Math.hypot(dx, dy) < DRAG_THRESHOLD) return;

      draggingRef.current = true;

      const coords = view.posAtCoords({ left: ev.clientX, top: ev.clientY });
      if (!coords) return;

      try {
        const $head = view.state.doc.resolve(coords.pos);
        const headCell = cellAround($head);
        if (!headCell) return;

        let headPos = headCell.pos;

        if (selectionBoundsRef.current) {
          const { bottomRow, tableStart } = selectionBoundsRef.current;
          if (headCell.start(-1) === tableStart) {
            const table = headCell.node(-1);
            const map = TableMap.get(table);
            const headRect = map.findCell(headCell.pos - tableStart);
            headPos = map.positionAt(bottomRow, headRect.left, table) + tableStart;
          }
        }

        const sel = CellSelection.create(
          view.state.doc,
          anchorCellPosRef.current!,
          headPos,
        );
        const { tr } = view.state;
        tr.setSelection(sel);
        view.dispatch(tr);
      } catch {
        // position may be outside table
      }
    };

    const onMouseUp = () => {
      document.removeEventListener("mousemove", onMouseMove);
      document.removeEventListener("mouseup", onMouseUp);

      if (!draggingRef.current) {
        setMenuOpen(prev => !prev);
      }

      draggingRef.current = false;
      anchorCellPosRef.current = null;
      selectionBoundsRef.current = null;
    };

    document.addEventListener("mousemove", onMouseMove);
    document.addEventListener("mouseup", onMouseUp);
  };

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
    <>
      <button
        ref={(node) => {
          handleRefs.setFloating(node);
          setTriggerEl(node);
          menuRefs.setReference(node);
        }}
        onMouseDown={onHandleMouseDown}
        onMouseEnter={() => setHandleHovered(true)}
        onMouseLeave={() => setHandleHovered(false)}
        type="button"
        style={{
          ...handleStyles,
          visibility: isPositioned ? "visible" : "hidden",
        }}
        className={trigger()}
      >
        {handleHovered || menuOpen
          ? (
              <div className="rounded-full bg-level-0 w-4.5 h-3.5 my-0.5 flex items-center">
                <DotsThreeCircleVerticalIcon size={18} weight="fill" />
              </div>
            )
          : <CircleIcon size={10} weight="fill" />}

      </button>
      {menuOpen && (
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
      )}
    </>
  );
}

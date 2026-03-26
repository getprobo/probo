// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { autoUpdate, offset, useFloating } from "@floating-ui/react";
import { CircleIcon, DotsThreeCircleVerticalIcon } from "@phosphor-icons/react";
import { cellAround, CellSelection, TableMap } from "@tiptap/pm/tables";
import { type Editor } from "@tiptap/react";
import { useLayoutEffect, useRef, useState } from "react";

import { cellDomElement } from "../_lib/cellDomElement";
import { DRAG_THRESHOLD } from "../_lib/constants";
import type { DropdownMenu } from "../_lib/useTableDropdownMenu";

import { tableCellMenuVariants } from "./variants";

const { trigger } = tableCellMenuVariants();

type TableCellMenuTriggerProps = {
  editor: Editor;
  activeCellEl: HTMLElement;
  menuOpen: boolean;
  setMenuOpen: DropdownMenu["setMenuOpen"];
  setTriggerEl: DropdownMenu["setTriggerEl"];
  menuRefs: DropdownMenu["menuRefs"];
};

export function TableCellMenuTrigger({
  editor,
  activeCellEl,
  menuOpen,
  setMenuOpen,
  setTriggerEl,
  menuRefs,
}: TableCellMenuTriggerProps) {
  const [handleHovered, setHandleHovered] = useState(false);
  const draggingRef = useRef(false);
  const dragStartPos = useRef({ x: 0, y: 0 });
  const anchorCellPosRef = useRef<number | null>(null);
  const selectionBoundsRef = useRef<{
    bottomRow: number;
    tableStart: number;
  } | null>(null);

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
    const ed = editor;
    const fallback = activeCellEl;
    const wrapper = fallback.closest(".tableWrapper");

    handleRefs.setReference({
      getBoundingClientRect() {
        let rect: DOMRect;

        const { selection } = ed.state;
        if (selection instanceof CellSelection) {
          let top = Infinity;
          let left = Infinity;
          let bottom = -Infinity;
          let right = -Infinity;
          selection.forEachCell((_node, pos) => {
            const el = cellDomElement(ed, pos);
            if (!el) return;
            const r = el.getBoundingClientRect();
            top = Math.min(top, r.top);
            left = Math.min(left, r.left);
            bottom = Math.max(bottom, r.bottom);
            right = Math.max(right, r.right);
          });
          rect = top !== Infinity
            ? new DOMRect(left, top, right - left, bottom - top)
            : fallback.getBoundingClientRect();
        } else {
          rect = fallback.getBoundingClientRect();
        }

        if (wrapper) {
          const clip = wrapper.getBoundingClientRect();
          const clampedLeft = Math.max(rect.left, clip.left);
          const clampedTop = Math.max(rect.top, clip.top);
          const clampedRight = Math.min(rect.right, clip.right);
          const clampedBottom = Math.min(rect.bottom, clip.bottom);
          const w = Math.max(0, clampedRight - clampedLeft);
          const h = Math.max(0, clampedBottom - clampedTop);
          return new DOMRect(clampedLeft, clampedTop, w, h);
        }

        return rect;
      },
    });
  }, [activeCellEl, editor, handleRefs]);

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

  return (
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
  );
}

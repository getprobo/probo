// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { autoUpdate, offset, size, useFloating } from "@floating-ui/react";
import { DotsThreeIcon } from "@phosphor-icons/react";
import { cellAround, CellSelection, TableMap } from "@tiptap/pm/tables";
import { type Editor } from "@tiptap/react";
import { useEffect, useLayoutEffect, useRef, useState } from "react";

import { DRAG_THRESHOLD } from "../_lib/constants";
import type { DropdownMenu } from "../_lib/useTableDropdownMenu";

import { getColumnRect, type HoveredColumn, moveColumn } from "./TableColumnMenu";
import { tableColumnMenuVariants } from "./variants";

const { trigger } = tableColumnMenuVariants();

type TableColumnMenuTriggerProps = {
  editor: Editor;
  hoveredCol: HoveredColumn | null;
  setHoveredCol: React.Dispatch<React.SetStateAction<HoveredColumn | null>>;
  menuOpen: boolean;
  setMenuOpen: DropdownMenu["setMenuOpen"];
  setTriggerEl: DropdownMenu["setTriggerEl"];
  menuRefs: DropdownMenu["menuRefs"];
};

export function TableColumnMenuTrigger({
  editor,
  hoveredCol,
  setHoveredCol,
  menuOpen,
  setMenuOpen,
  setTriggerEl,
  menuRefs,
}: TableColumnMenuTriggerProps) {
  const [dragIndicator, setDragIndicator] = useState<{
    left: number;
    top: number;
    height: number;
  } | null>(null);

  const draggingRef = useRef(false);
  const dragStartPos = useRef({ x: 0, y: 0 });
  const rafId = useRef<number | null>(null);
  const hoveredColRef = useRef<HoveredColumn | null>(null);

  useEffect(() => {
    hoveredColRef.current = hoveredCol;
  }, [hoveredCol]);

  useEffect(() => {
    if (editor.isDestroyed || !editor.isEditable) return;

    const editorDom = editor.view.dom;

    const onMouseMove = (e: MouseEvent) => {
      if (draggingRef.current || menuOpen) return;

      if (rafId.current) return;
      rafId.current = requestAnimationFrame(() => {
        rafId.current = null;

        const target = e.target as HTMLElement;

        if (
          target.closest("[data-column-handle]")
          || target.closest("[data-column-menu]")
        ) {
          return;
        }

        const cell = target.closest("td, th");
        if (cell && editorDom.contains(cell)) {
          try {
            const pos = editor.view.posAtDOM(cell, 0);
            const $pos = editor.state.doc.resolve(pos);
            const cellResolved = cellAround($pos);
            if (cellResolved) {
              const table = cellResolved.node(-1);
              const ts = cellResolved.start(-1);
              const map = TableMap.get(table);
              const cellRect = map.findCell(cellResolved.pos - ts);
              const ci = cellRect.left;

              setHoveredCol((prev) => {
                if (prev && prev.colIndex === ci && prev.tableStart === ts) {
                  return prev;
                }
                return { colIndex: ci, tableStart: ts };
              });
              return;
            }
          } catch {
            // fall through to clear
          }
        }

        const current = hoveredColRef.current;
        if (current) {
          const rect = getColumnRect(editor, current.tableStart, current.colIndex);
          if (rect) {
            const zoneTop = rect.top - 40;
            if (
              e.clientX >= rect.left
              && e.clientX <= rect.left + rect.width
              && e.clientY >= zoneTop
              && e.clientY <= rect.top
            ) {
              return;
            }
          }
        }

        if (hoveredColRef.current) {
          setHoveredCol(null);
        }
      });
    };

    document.addEventListener("mousemove", onMouseMove);

    return () => {
      document.removeEventListener("mousemove", onMouseMove);
      if (rafId.current) {
        cancelAnimationFrame(rafId.current);
        rafId.current = null;
      }
    };
  }, [editor, menuOpen, setHoveredCol]);

  const {
    refs: handleRefs,
    floatingStyles: handleStyles,
    isPositioned,
  } = useFloating({
    strategy: "fixed",
    placement: "top",
    middleware: [
      offset(6),
      size({
        apply({ rects, elements }) {
          Object.assign(elements.floating.style, {
            width: `${rects.reference.width}px`,
          });
        },
      }),
    ],
    whileElementsMounted: (ref, floating, update) =>
      autoUpdate(ref, floating, update, { animationFrame: true }),
  });

  useLayoutEffect(() => {
    if (!hoveredCol) {
      handleRefs.setReference(null);
      return;
    }

    const { colIndex, tableStart } = hoveredCol;
    const ed = editor;

    handleRefs.setReference({
      getBoundingClientRect() {
        const r = getColumnRect(ed, tableStart, colIndex);
        if (!r) return new DOMRect(0, 0, 0, 0);
        return new DOMRect(r.left, r.top, r.width, 0);
      },
    });
  }, [hoveredCol, editor, handleRefs]);

  const computeTargetGap = (clientX: number, tableStart: number): number => {
    const table = editor.state.doc.nodeAt(tableStart - 1);
    if (!table) return 0;

    const map = TableMap.get(table);
    let targetGap = 0;

    for (let col = 0; col < map.width; col++) {
      const r = getColumnRect(editor, tableStart, col);
      if (!r) continue;
      const midX = r.left + r.width / 2;
      if (clientX > midX) {
        targetGap = col + 1;
      } else {
        targetGap = col;
        break;
      }
    }

    return targetGap;
  };

  const computeGapX = (
    tableStart: number,
    gap: number,
  ): number | null => {
    const table = editor.state.doc.nodeAt(tableStart - 1);
    if (!table) return null;

    const map = TableMap.get(table);

    if (gap <= 0) {
      const r = getColumnRect(editor, tableStart, 0);
      return r ? r.left : null;
    }

    if (gap >= map.width) {
      const r = getColumnRect(editor, tableStart, map.width - 1);
      return r ? r.left + r.width : null;
    }

    const rLeft = getColumnRect(editor, tableStart, gap - 1);
    const rRight = getColumnRect(editor, tableStart, gap);
    if (rLeft && rRight) {
      return (rLeft.left + rLeft.width + rRight.left) / 2;
    }
    return null;
  };

  const onHandleMouseDown = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();

    if (menuOpen || !hoveredCol) return;

    draggingRef.current = false;
    dragStartPos.current = { x: e.clientX, y: e.clientY };

    const { colIndex: fromCol, tableStart } = hoveredCol;
    const view = editor.view;

    const onMouseMove = (ev: MouseEvent) => {
      const dx = ev.clientX - dragStartPos.current.x;
      const dy = ev.clientY - dragStartPos.current.y;

      if (!draggingRef.current) {
        if (Math.hypot(dx, dy) < DRAG_THRESHOLD) return;
        draggingRef.current = true;

        try {
          const table = editor.state.doc.nodeAt(tableStart - 1);
          if (table) {
            const map = TableMap.get(table);
            const anchorPos
              = map.positionAt(0, fromCol, table) + tableStart;
            const headPos
              = map.positionAt(map.height - 1, fromCol, table) + tableStart;
            const sel = CellSelection.create(
              view.state.doc,
              anchorPos,
              headPos,
            );
            view.dispatch(view.state.tr.setSelection(sel));
          }
        } catch {
          // table may have changed
        }
      }

      try {
        const targetGap = computeTargetGap(ev.clientX, tableStart);

        if (targetGap === fromCol || targetGap === fromCol + 1) {
          setDragIndicator(null);
          return;
        }

        const gapX = computeGapX(tableStart, targetGap);
        if (gapX !== null) {
          const colRect = getColumnRect(editor, tableStart, 0);
          if (colRect) {
            setDragIndicator({
              left: gapX,
              top: colRect.top,
              height: colRect.height,
            });
          }
        }
      } catch {
        // position may be outside table
      }
    };

    const onMouseUp = (ev: MouseEvent) => {
      document.removeEventListener("mousemove", onMouseMove);
      document.removeEventListener("mouseup", onMouseUp);

      if (draggingRef.current) {
        setDragIndicator(null);

        try {
          const targetGap = computeTargetGap(ev.clientX, tableStart);

          if (targetGap !== fromCol && targetGap !== fromCol + 1) {
            const toCol
              = fromCol < targetGap ? targetGap - 1 : targetGap;
            moveColumn(editor, tableStart, fromCol, toCol);
          }
        } catch {
          // table may have changed
        }
      } else {
        try {
          const table = editor.state.doc.nodeAt(tableStart - 1);
          if (table) {
            const map = TableMap.get(table);
            const anchorPos
              = map.positionAt(0, fromCol, table) + tableStart;
            const headPos
              = map.positionAt(map.height - 1, fromCol, table) + tableStart;
            const sel = CellSelection.create(
              view.state.doc,
              anchorPos,
              headPos,
            );
            view.dispatch(view.state.tr.setSelection(sel));
          }
        } catch {
          // table may have changed
        }
        setMenuOpen(prev => !prev);
      }

      draggingRef.current = false;
    };

    document.addEventListener("mousemove", onMouseMove);
    document.addEventListener("mouseup", onMouseUp);
  };

  return (
    <>
      <button
        ref={(node) => {
          handleRefs.setFloating(node);
          setTriggerEl(node);
          menuRefs.setReference(node);
        }}
        data-column-handle
        onMouseDown={onHandleMouseDown}
        type="button"
        style={{
          ...handleStyles,
          visibility:
            isPositioned && (hoveredCol || menuOpen) ? "visible" : "hidden",
        }}
        className={trigger()}
      >
        <DotsThreeIcon size={16} weight="bold" />
      </button>
      {dragIndicator && (
        <div
          style={{
            position: "fixed",
            left: dragIndicator.left - 1,
            top: dragIndicator.top,
            width: 2,
            height: dragIndicator.height,
            backgroundColor: "var(--color-border-info)",
            zIndex: 40,
            pointerEvents: "none",
          }}
        />
      )}
    </>
  );
}

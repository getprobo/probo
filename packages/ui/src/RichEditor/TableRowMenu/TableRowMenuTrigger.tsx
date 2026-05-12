// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { autoUpdate, offset, size, useFloating } from "@floating-ui/react";
import { DotsThreeVerticalIcon } from "@phosphor-icons/react";
import { cellAround, CellSelection, TableMap } from "@tiptap/pm/tables";
import { type Editor } from "@tiptap/react";
import { useEffect, useLayoutEffect, useRef, useState } from "react";

import { DRAG_THRESHOLD } from "../_lib/constants";
import type { DropdownMenu } from "../_lib/useTableDropdownMenu";

import { getRowRect, type HoveredRow, moveRow } from "./TableRowMenu";
import { tableRowMenuVariants } from "./variants";

const { trigger } = tableRowMenuVariants();

type TableRowMenuTriggerProps = {
  editor: Editor;
  hoveredRow: HoveredRow | null;
  setHoveredRow: React.Dispatch<React.SetStateAction<HoveredRow | null>>;
  menuOpen: boolean;
  setMenuOpen: DropdownMenu["setMenuOpen"];
  setTriggerEl: DropdownMenu["setTriggerEl"];
  menuRefs: DropdownMenu["menuRefs"];
};

export function TableRowMenuTrigger({
  editor,
  hoveredRow,
  setHoveredRow,
  menuOpen,
  setMenuOpen,
  setTriggerEl,
  menuRefs,
}: TableRowMenuTriggerProps) {
  const [dragIndicator, setDragIndicator] = useState<{
    left: number;
    top: number;
    width: number;
  } | null>(null);

  const draggingRef = useRef(false);
  const dragStartPos = useRef({ x: 0, y: 0 });
  const rafId = useRef<number | null>(null);
  const hoveredRowRef = useRef<HoveredRow | null>(null);
  const dragCleanupRef = useRef<(() => void) | null>(null);

  useEffect(() => {
    hoveredRowRef.current = hoveredRow;
  }, [hoveredRow]);

  useEffect(() => {
    return () => {
      dragCleanupRef.current?.();
    };
  }, []);

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
          target.closest("[data-row-handle]")
          || target.closest("[data-row-menu]")
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
              const ri = cellRect.top;

              setHoveredRow((prev) => {
                if (prev && prev.rowIndex === ri && prev.tableStart === ts) {
                  return prev;
                }
                return { rowIndex: ri, tableStart: ts };
              });
              return;
            }
          } catch {
            // fall through to clear
          }
        }

        const current = hoveredRowRef.current;
        if (current) {
          const rect = getRowRect(editor, current.tableStart, current.rowIndex);
          if (rect) {
            const zoneLeft = rect.left - 40;
            if (
              e.clientY >= rect.top
              && e.clientY <= rect.top + rect.height
              && e.clientX >= zoneLeft
              && e.clientX <= rect.left
            ) {
              return;
            }
          }
        }

        if (hoveredRowRef.current) {
          setHoveredRow(null);
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
  }, [editor, menuOpen, setHoveredRow]);

  const {
    refs: handleRefs,
    floatingStyles: handleStyles,
    isPositioned,
  } = useFloating({
    strategy: "fixed",
    placement: "left",
    middleware: [
      offset(4),
      size({
        apply({ rects, elements }) {
          Object.assign(elements.floating.style, {
            height: `${rects.reference.height}px`,
          });
        },
      }),
    ],
    whileElementsMounted: (ref, floating, update) =>
      autoUpdate(ref, floating, update, { animationFrame: true }),
  });

  useLayoutEffect(() => {
    if (!hoveredRow) {
      handleRefs.setReference(null);
      return;
    }

    const { rowIndex, tableStart } = hoveredRow;
    const ed = editor;

    handleRefs.setReference({
      getBoundingClientRect() {
        const r = getRowRect(ed, tableStart, rowIndex);
        if (!r) return new DOMRect(0, 0, 0, 0);
        return new DOMRect(r.left, r.top, 0, r.height);
      },
    });
  }, [hoveredRow, editor, handleRefs]);

  const computeTargetGap = (clientY: number, tableStart: number): number => {
    const table = editor.state.doc.nodeAt(tableStart - 1);
    if (!table) return 0;

    const map = TableMap.get(table);
    let targetGap = 0;

    for (let row = 0; row < map.height; row++) {
      const r = getRowRect(editor, tableStart, row);
      if (!r) continue;
      const midY = r.top + r.height / 2;
      if (clientY > midY) {
        targetGap = row + 1;
      } else {
        targetGap = row;
        break;
      }
    }

    return targetGap;
  };

  const computeGapY = (
    tableStart: number,
    gap: number,
  ): number | null => {
    const table = editor.state.doc.nodeAt(tableStart - 1);
    if (!table) return null;

    const map = TableMap.get(table);

    if (gap <= 0) {
      const r = getRowRect(editor, tableStart, 0);
      return r ? r.top : null;
    }

    if (gap >= map.height) {
      const r = getRowRect(editor, tableStart, map.height - 1);
      return r ? r.top + r.height : null;
    }

    const rAbove = getRowRect(editor, tableStart, gap - 1);
    const rBelow = getRowRect(editor, tableStart, gap);
    if (rAbove && rBelow) {
      return (rAbove.top + rAbove.height + rBelow.top) / 2;
    }
    return null;
  };

  const onHandleMouseDown = (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();

    if (menuOpen || !hoveredRow) return;

    draggingRef.current = false;
    dragStartPos.current = { x: e.clientX, y: e.clientY };

    const { rowIndex: fromRow, tableStart } = hoveredRow;
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
              = map.positionAt(fromRow, 0, table) + tableStart;
            const headPos
              = map.positionAt(fromRow, map.width - 1, table) + tableStart;
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
        const targetGap = computeTargetGap(ev.clientY, tableStart);

        if (targetGap === fromRow || targetGap === fromRow + 1) {
          setDragIndicator(null);
          return;
        }

        const gapY = computeGapY(tableStart, targetGap);
        if (gapY !== null) {
          const rowRect = getRowRect(editor, tableStart, 0);
          if (rowRect) {
            setDragIndicator({
              left: rowRect.left,
              top: gapY,
              width: rowRect.width,
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
      dragCleanupRef.current = null;

      if (draggingRef.current) {
        setDragIndicator(null);

        try {
          const targetGap = computeTargetGap(ev.clientY, tableStart);

          if (targetGap !== fromRow && targetGap !== fromRow + 1) {
            const toRow
              = fromRow < targetGap ? targetGap - 1 : targetGap;
            moveRow(editor, tableStart, fromRow, toRow);
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
              = map.positionAt(fromRow, 0, table) + tableStart;
            const headPos
              = map.positionAt(fromRow, map.width - 1, table) + tableStart;
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
    dragCleanupRef.current = () => {
      document.removeEventListener("mousemove", onMouseMove);
      document.removeEventListener("mouseup", onMouseUp);
    };
  };

  return (
    <>
      <button
        ref={(node) => {
          handleRefs.setFloating(node);
          setTriggerEl(node);
          menuRefs.setReference(node);
        }}
        data-row-handle
        onMouseDown={onHandleMouseDown}
        type="button"
        style={{
          ...handleStyles,
          visibility:
            isPositioned && (hoveredRow || menuOpen) ? "visible" : "hidden",
        }}
        className={trigger()}
      >
        <DotsThreeVerticalIcon size={16} weight="bold" />
      </button>
      {dragIndicator && (
        <div
          style={{
            position: "fixed",
            left: dragIndicator.left,
            top: dragIndicator.top - 1,
            width: dragIndicator.width,
            height: 2,
            backgroundColor: "var(--color-border-info)",
            zIndex: 40,
            pointerEvents: "none",
          }}
        />
      )}
    </>
  );
}

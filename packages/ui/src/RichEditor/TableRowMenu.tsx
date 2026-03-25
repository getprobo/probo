// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import { autoUpdate, offset, size, useFloating } from "@floating-ui/react";
import {
  BroomIcon,
  CopyIcon,
  CrownSimpleIcon,
  DotsThreeVerticalIcon,
  PlusIcon,
  TrashIcon,
} from "@phosphor-icons/react";
import type { Node as PMNode } from "@tiptap/pm/model";
import { TextSelection } from "@tiptap/pm/state";
import { cellAround, CellSelection, TableMap } from "@tiptap/pm/tables";
import { type Editor } from "@tiptap/react";
import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { tv } from "tailwind-variants";

import { cellDomElement } from "./_lib/cellDomElement";
import { DRAG_THRESHOLD } from "./_lib/constants";
import { useTableDropdownMenu } from "./_lib/useTableDropdownMenu";
import { MenuButton } from "./MenuButton";

const tableRowMenuVariants = tv({
  slots: {
    trigger: [
      "z-10 flex flex-col items-center justify-center",
      "rounded text-txt-tertiary bg-subtle hover:bg-border-solid cursor-grab",
      "px-0.5 w-3",
    ],
    menu: ["rounded-lg border border-border-mid bg-level-0 p-1 shadow-md z-20"],
  },
});

const { trigger, menu } = tableRowMenuVariants();

type HoveredRow = {
  rowIndex: number;
  tableStart: number;
};

type TableRowMenuProps = {
  editor: Editor;
};

function getRowRect(
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

function moveRow(
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
  const [dragIndicator, setDragIndicator] = useState<{
    left: number;
    top: number;
    width: number;
  } | null>(null);

  const draggingRef = useRef(false);
  const dragStartPos = useRef({ x: 0, y: 0 });
  const rafId = useRef<number | null>(null);
  const hoveredRowRef = useRef<HoveredRow | null>(null);

  useEffect(() => {
    hoveredRowRef.current = hoveredRow;
  }, [hoveredRow]);

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
  }, [editor, menuOpen]);

  const {
    refs: handleRefs,
    floatingStyles: handleStyles,
    isPositioned,
  } = useFloating({
    strategy: "fixed",
    placement: "left",
    middleware: [
      offset(6),
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

  if (!hoveredRow && !menuOpen) return null;

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
  };

  const currentRow = hoveredRow;

  const isFirstRow = currentRow?.rowIndex === 0;

  const isHeaderRow = (): boolean => {
    if (!currentRow || currentRow.rowIndex !== 0) return false;
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
    if (!currentRow || currentRow.rowIndex !== 0) return;
    const { tableStart } = currentRow;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const map = TableMap.get(table);
      const { tr, schema } = editor.state;
      const targetType = isHeaderRow()
        ? schema.nodes.tableCell
        : schema.nodes.tableHeader;

      for (let col = 0; col < map.width; col++) {
        const cellPos = map.map[col] + tableStart;
        const cellNode = editor.state.doc.nodeAt(cellPos);
        if (!cellNode) continue;
        tr.setNodeMarkup(cellPos, targetType, cellNode.attrs);
      }

      editor.view.dispatch(tr);
    } catch {
      // table may have changed
    }

    setMenuOpen(false);
  };

  const handleDeleteRow = () => {
    if (!currentRow) return;
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
    if (!currentRow) return;
    const { rowIndex, tableStart } = currentRow;

    try {
      const tableNodePos = tableStart - 1;
      const table = editor.state.doc.nodeAt(tableNodePos);
      if (!table) return;

      const rows: PMNode[] = [];
      table.forEach((row, _offset, i) => {
        rows.push(row);
        if (i === rowIndex) {
          rows.push(row.copy(row.content));
        }
      });

      const newTable = table.type.create(table.attrs, rows);
      const { tr } = editor.state;
      tr.replaceWith(tableNodePos, tableNodePos + table.nodeSize, newTable);
      editor.view.dispatch(tr);
    } catch {
      // table may have changed
    }

    setMenuOpen(false);
  };

  const handleInsertAbove = () => {
    if (!currentRow) return;
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
    if (!currentRow) return;
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
    if (!currentRow) return;
    const { rowIndex, tableStart } = currentRow;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const map = TableMap.get(table);
      const { tr } = editor.state;
      const { schema } = editor.state;

      for (let col = 0; col < map.width; col++) {
        const cellPos = map.map[rowIndex * map.width + col] + tableStart;
        const cellNode = editor.state.doc.nodeAt(cellPos);
        if (!cellNode) continue;

        const start = cellPos + 1;
        const end = cellPos + cellNode.nodeSize - 1;

        tr.replaceWith(
          tr.mapping.map(start),
          tr.mapping.map(end),
          schema.nodes.paragraph.create(),
        );
      }

      editor.view.dispatch(tr);
    } catch {
      // table may have changed
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
      {menuOpen && (
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
      )}
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

// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import {
  autoUpdate,
  flip,
  offset,
  shift,
  size,
  useDismiss,
  useFloating,
  useFloatingRootContext,
  useInteractions,
} from "@floating-ui/react";
import {
  BroomIcon,
  CopyIcon,
  CrownSimpleIcon,
  DotsThreeIcon,
  PlusIcon,
  TrashIcon,
} from "@phosphor-icons/react";
import type { Node as PMNode } from "@tiptap/pm/model";
import { TextSelection } from "@tiptap/pm/state";
import { cellAround, CellSelection, TableMap } from "@tiptap/pm/tables";
import { type useEditor } from "@tiptap/react";
import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { tv } from "tailwind-variants";

import { MenuButton } from "./MenuButton";

const DRAG_THRESHOLD = 4;

const tableColumnMenuVariants = tv({
  slots: {
    trigger: [
      "z-10 flex items-center justify-center",
      "rounded text-txt-tertiary bg-subtle hover:bg-border-solid cursor-grab",
      "py-0.5 h-3",
    ],
    menu: ["rounded-lg border border-border-mid bg-level-0 p-1 shadow-md z-20"],
  },
});

const { trigger, menu } = tableColumnMenuVariants();

type HoveredColumn = {
  colIndex: number;
  tableStart: number;
};

type TableColumnMenuProps = {
  editor: ReturnType<typeof useEditor>;
};

function cellDomElement(
  editor: NonNullable<ReturnType<typeof useEditor>>,
  cellPos: number,
): HTMLElement | null {
  const dom = editor.view.domAtPos(cellPos + 1);
  let el: Node | null = dom.node;
  if (el.nodeType === Node.TEXT_NODE) el = el.parentElement;
  while (el && !(el instanceof HTMLTableCellElement)) {
    el = (el as HTMLElement).parentElement;
  }
  return el as HTMLElement | null;
}

function getColumnRect(
  editor: NonNullable<ReturnType<typeof useEditor>>,
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

function moveColumn(
  editor: NonNullable<ReturnType<typeof useEditor>>,
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

export function TableColumnMenu({ editor }: TableColumnMenuProps) {
  const [menuOpen, setMenuOpen] = useState(false);
  const [hoveredCol, setHoveredCol] = useState<HoveredColumn | null>(null);
  const [triggerEl, setTriggerEl] = useState<Element | null>(null);
  const [dropdownEl, setDropdownEl] = useState<HTMLElement | null>(null);
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
    if (!editor || editor.isDestroyed || !editor.isEditable) return;

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
  }, [editor, menuOpen]);

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

  const menuRootContext = useFloatingRootContext({
    open: menuOpen,
    onOpenChange: setMenuOpen,
    elements: { reference: triggerEl, floating: dropdownEl },
  });

  const { refs: menuRefs, floatingStyles: menuStyles } = useFloating({
    rootContext: menuRootContext,
    strategy: "fixed",
    placement: "bottom-start",
    middleware: [offset(4), flip(), shift()],
    whileElementsMounted: autoUpdate,
  });

  const dismiss = useDismiss(menuRootContext);
  const { getFloatingProps } = useInteractions([dismiss]);

  useLayoutEffect(() => {
    if (!editor || !hoveredCol) {
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

  if (!editor || (!hoveredCol && !menuOpen)) return null;

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

  const currentCol = hoveredCol;

  const isFirstColumn = currentCol?.colIndex === 0;

  const isHeaderColumn = (): boolean => {
    if (!currentCol || currentCol.colIndex !== 0) return false;
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
    if (!currentCol || currentCol.colIndex !== 0) return;
    const { tableStart } = currentCol;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const map = TableMap.get(table);
      const { tr, schema } = editor.state;
      const targetType = isHeaderColumn()
        ? schema.nodes.tableCell
        : schema.nodes.tableHeader;

      for (let row = 0; row < map.height; row++) {
        const cellPos = map.map[row * map.width] + tableStart;
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

  const handleDeleteColumn = () => {
    if (!currentCol) return;
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
    if (!currentCol) return;
    const { colIndex, tableStart } = currentCol;

    try {
      const tableNodePos = tableStart - 1;
      const table = editor.state.doc.nodeAt(tableNodePos);
      if (!table) return;

      const rows: PMNode[] = [];
      table.forEach((row) => {
        const cells: PMNode[] = [];
        row.forEach((cell, _offset, i) => {
          cells.push(cell);
          if (i === colIndex) {
            cells.push(cell.copy(cell.content));
          }
        });
        rows.push(row.type.create(row.attrs, cells));
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

  const handleInsertLeft = () => {
    if (!currentCol) return;
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
    if (!currentCol) return;
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
    if (!currentCol) return;
    const { colIndex, tableStart } = currentCol;

    try {
      const table = editor.state.doc.nodeAt(tableStart - 1);
      if (!table) return;

      const map = TableMap.get(table);
      const { tr } = editor.state;
      const { schema } = editor.state;

      for (let row = 0; row < map.height; row++) {
        const cellPos = map.map[row * map.width + colIndex] + tableStart;
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
      {menuOpen && (
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
      )}
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

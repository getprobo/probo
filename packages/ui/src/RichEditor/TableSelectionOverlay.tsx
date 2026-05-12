// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import {
  autoUpdate,
  type Middleware,
  size,
  useFloating,
} from "@floating-ui/react";
import { cellAround, CellSelection } from "@tiptap/pm/tables";
import { type Editor, useEditorState } from "@tiptap/react";
import { useLayoutEffect } from "react";
import { createPortal } from "react-dom";

import { cellDomElement } from "./_lib/cellDomElement";

const cover: Middleware = {
  name: "cover",
  fn({ rects }) {
    return { x: rects.reference.x, y: rects.reference.y };
  },
};

type TableSelectionOverlayProps = {
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

export function TableSelectionOverlay({ editor }: TableSelectionOverlayProps) {
  const activeCellEl = useEditorState({
    editor,
    selector: ({ editor: e }) => {
      if (e.isDestroyed || !e.isEditable) return null;
      return getActiveCellEl(e);
    },
  });

  const isCellSelection = useEditorState({
    editor,
    selector: ({ editor: e }) =>
      !e.isDestroyed && e.state.selection instanceof CellSelection,
  });

  const wrapperEl = activeCellEl?.closest(".tableWrapper") as HTMLElement | null;

  const {
    refs,
    floatingStyles,
    isPositioned,
  } = useFloating({
    strategy: "absolute",
    placement: "bottom-start",
    middleware: [
      cover,
      size({
        apply({ rects, elements }) {
          Object.assign(elements.floating.style, {
            width: `${rects.reference.width}px`,
            height: `${rects.reference.height}px`,
          });
        },
      }),
    ],
    whileElementsMounted: (ref, floating, update) =>
      autoUpdate(ref, floating, update, { animationFrame: true }),
  });

  useLayoutEffect(() => {
    const ed = editor;
    const fallback = activeCellEl;

    if (!fallback) {
      refs.setReference(null);
      return;
    }

    refs.setReference({
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
            const r = el.getBoundingClientRect();
            top = Math.min(top, r.top);
            left = Math.min(left, r.left);
            bottom = Math.max(bottom, r.bottom);
            right = Math.max(right, r.right);
          });
          if (top !== Infinity) {
            return new DOMRect(left, top, right - left, bottom - top);
          }
        }
        return fallback.getBoundingClientRect();
      },
    });
  }, [activeCellEl, editor, refs]);

  if (!activeCellEl || !wrapperEl) return null;

  return createPortal(
    <div
      ref={(node) => { refs.setFloating(node); }}
      style={{
        ...floatingStyles,
        visibility: isPositioned ? "visible" : "hidden",
        pointerEvents: "none",
        zIndex: 10,
      }}
    >
      <div
        style={{
          width: "100%",
          height: "100%",
          backgroundColor: isCellSelection
            ? "color-mix(in srgb, var(--color-info) 50%, transparent)"
            : undefined,
          border: "2px solid var(--color-border-info)",
          borderRadius: 2,
        }}
      />
    </div>,
    wrapperEl,
  );
}

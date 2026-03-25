// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
// Use of this source code is governed by the ISC license
// that can be found in the LICENSE file.

import type { Editor } from "@tiptap/react";
import { type Dispatch, type SetStateAction, useEffect, useRef, useState } from "react";

function findClosestRootBlock(element: Element, editorDom: Element): HTMLElement | null {
  let current: Element | null = element;

  while (current?.parentElement && current.parentElement !== editorDom) {
    current = current.parentElement;
  }

  return current?.parentElement === editorDom ? (current as HTMLElement) : null;
}

type UseHoveredBlockResult = {
  hoveredBlock: HTMLElement | null;
  setHoveredBlock: Dispatch<SetStateAction<HTMLElement | null>>;
};

export function useHoveredBlock(
  editor: Editor,
  isDisabled: boolean,
): UseHoveredBlockResult {
  const [hoveredBlock, setHoveredBlock] = useState<HTMLElement | null>(null);
  const rafId = useRef<number | null>(null);

  useEffect(() => {
    if (editor.isDestroyed) return;
    const editorDom = editor.view.dom;

    const onMouseMove = (e: MouseEvent) => {
      if (isDisabled) return;

      if (rafId.current) return;
      rafId.current = requestAnimationFrame(() => {
        rafId.current = null;

        if (!editor.isEditable) {
          setHoveredBlock(null);
          return;
        }

        const elements = editorDom.ownerDocument.elementsFromPoint(e.clientX, e.clientY);
        let block: HTMLElement | null = null;

        for (const el of elements) {
          if (!editorDom.contains(el)) continue;
          block = findClosestRootBlock(el, editorDom);
          if (block) break;
        }

        if (block) {
          setHoveredBlock(block);
        }
      });
    };

    editorDom.addEventListener("mousemove", onMouseMove);

    return () => {
      editorDom.removeEventListener("mousemove", onMouseMove);
      if (rafId.current) {
        cancelAnimationFrame(rafId.current);
        rafId.current = null;
      }
    };
  }, [editor, isDisabled]);

  return { hoveredBlock, setHoveredBlock };
}

// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import type { Editor } from "@tiptap/core";
import { useCallback, useEffect, useRef, useState } from "react";

import { placeReviewCards, type ReviewCardSize } from "./reviewCardLayout";

export type ReviewCardAnchor = {
  key: string;
  // The column element the card renders in, when it differs from the key. A
  // card whose key changes keeps its element, place, and height this way.
  slot?: string;
  pos: number;
};

type ReviewCardLayout = {
  tops: Readonly<Record<string, number>>;
  height: number;
};

const emptyLayout: ReviewCardLayout = { tops: {}, height: 0 };

// An unmeasured card still needs a height so the cards below it can be placed.
const estimatedCardHeight = 96;

export function reviewCardSlot(anchor: ReviewCardAnchor): string {
  return anchor.slot ?? anchor.key;
}

// Tops are keyed by slot. A card without a measured height yet is placed with
// an estimate.
export function layoutReviewCards(
  anchors: ReadonlyArray<ReviewCardAnchor>,
  activeKey: string | null,
  anchorTop: (pos: number) => number | null,
  cardHeight: (slot: string) => number | undefined,
): ReviewCardLayout {
  const sizes: ReviewCardSize[] = [];
  let activeSlot: string | null = null;
  for (const anchor of anchors) {
    const slot = reviewCardSlot(anchor);
    if (anchor.key === activeKey) activeSlot = slot;
    const top = anchorTop(anchor.pos);
    if (top == null) continue;
    sizes.push({ key: slot, anchor: top, height: cardHeight(slot) ?? estimatedCardHeight });
  }
  const tops = placeReviewCards(sizes, activeSlot);
  const height = sizes.reduce((max, size) => Math.max(max, tops[size.key] + size.height), 0);
  return { tops, height };
}

// useReviewCardTops places each card beside its anchor. The column must share
// the scroll container with the text, so tops never change on scroll and only
// a document, card height, or width change measures again. `anchors` must be
// memoized. Each card element carries its slot in `data-review-card`.
export function useReviewCardTops(
  editor: Editor | null,
  anchors: ReadonlyArray<ReviewCardAnchor>,
  activeKey: string | null,
) {
  const [column, setColumn] = useState<HTMLElement | null>(null);
  const [layout, setLayout] = useState<ReviewCardLayout>(emptyLayout);
  const cardNodes = useRef(new Map<string, HTMLElement>());
  const observerRef = useRef<ResizeObserver | null>(null);
  const scheduleRef = useRef<() => void>(() => {});
  const inputs = useRef({ anchors, activeKey });

  useEffect(() => {
    inputs.current = { anchors, activeKey };
    scheduleRef.current();
  }, [activeKey, anchors]);

  useEffect(() => {
    if (!editor || !column) {
      return;
    }
    let frame = 0;
    let disposed = false;

    const measure = () => {
      frame = 0;
      // Before the editor is in the page, every anchor measures as zero.
      if (editor.isDestroyed || !editorDom(editor)?.isConnected) return;
      const { anchors, activeKey } = inputs.current;
      const origin = column.getBoundingClientRect().top;
      const { tops, height } = layoutReviewCards(
        anchors,
        activeKey,
        pos => anchorTop(editor, pos, origin),
        slot => cardNodes.current.get(slot)?.offsetHeight,
      );
      setLayout(current => sameLayout(current, tops, height) ? current : { tops, height });
    };

    const schedule = () => {
      if (frame === 0) frame = window.requestAnimationFrame(measure);
    };

    // The text can reflow without the column changing size: the editor fills
    // in after mount, and the web font swaps in late.
    const observer = new ResizeObserver(schedule);
    observer.observe(column);
    if (column.parentElement) observer.observe(column.parentElement);
    const dom = editorDom(editor);
    if (dom) observer.observe(dom);
    for (const node of cardNodes.current.values()) observer.observe(node);
    observerRef.current = observer;
    scheduleRef.current = schedule;
    schedule();
    void document.fonts.ready.then(() => {
      if (!disposed) schedule();
    });

    return () => {
      disposed = true;
      if (frame !== 0) window.cancelAnimationFrame(frame);
      observer.disconnect();
      observerRef.current = null;
      scheduleRef.current = () => {};
    };
  }, [column, editor]);

  const bindCard = useCallback((node: HTMLElement | null) => {
    const key = node?.dataset.reviewCard;
    if (!node || !key) return;
    const nodes = cardNodes.current;
    nodes.set(key, node);
    observerRef.current?.observe(node);
    scheduleRef.current();
    return () => {
      if (nodes.get(key) === node) nodes.delete(key);
      observerRef.current?.unobserve(node);
    };
  }, []);

  return {
    columnRef: setColumn,
    bindCard,
    tops: layout.tops,
    height: layout.height,
  };
}

function editorDom(editor: Editor): HTMLElement | null {
  try {
    return editor.view.dom;
  } catch {
    return null;
  }
}

function anchorTop(editor: Editor, pos: number, origin: number): number | null {
  const clamped = Math.max(0, Math.min(pos, editor.state.doc.content.size));
  try {
    return editor.view.coordsAtPos(clamped).top - origin;
  } catch {
    return null;
  }
}

function sameLayout(current: ReviewCardLayout, tops: Record<string, number>, height: number) {
  if (current.height !== height) return false;
  const keys = Object.keys(tops);
  if (keys.length !== Object.keys(current.tops).length) return false;
  return keys.every(key => current.tops[key] === tops[key]);
}

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
  pos: number;
};

type ReviewCardLayout = {
  tops: Readonly<Record<string, number>>;
  height: number;
};

const emptyLayout: ReviewCardLayout = { tops: {}, height: 0 };

// An unmeasured card still needs a height so the cards below it can be placed.
const estimatedCardHeight = 96;

// useReviewCardTops places each card beside its anchor. The column must share
// the scroll container with the text, so tops never change on scroll and only
// a document, card height, or width change measures again. `anchors` must be
// memoized. Each card element carries its key in `data-review-card`.
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

    const measure = () => {
      frame = 0;
      if (editor.isDestroyed) return;
      const { anchors, activeKey } = inputs.current;
      const origin = column.getBoundingClientRect().top;
      const sizes: ReviewCardSize[] = [];
      for (const anchor of anchors) {
        const top = anchorTop(editor, anchor.pos, origin);
        if (top == null) continue;
        sizes.push({
          key: anchor.key,
          anchor: top,
          height: cardNodes.current.get(anchor.key)?.offsetHeight ?? estimatedCardHeight,
        });
      }
      const tops = placeReviewCards(sizes, activeKey);
      const height = sizes.reduce((max, size) => Math.max(max, tops[size.key] + size.height), 0);
      setLayout(current => sameLayout(current, tops, height) ? current : { tops, height });
    };

    const schedule = () => {
      if (frame === 0) frame = window.requestAnimationFrame(measure);
    };

    const observer = new ResizeObserver(schedule);
    observer.observe(column);
    if (column.parentElement) observer.observe(column.parentElement);
    for (const node of cardNodes.current.values()) observer.observe(node);
    observerRef.current = observer;
    scheduleRef.current = schedule;
    schedule();

    return () => {
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

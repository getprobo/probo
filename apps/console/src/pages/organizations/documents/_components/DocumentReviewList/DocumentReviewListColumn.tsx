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
import { type ReactNode, useEffect, useState } from "react";

import { type ReviewCardAnchor, useReviewCardTops } from "../../_lib/useReviewCardTops";
import { documentReviewList } from "../../variants";

interface DocumentReviewListColumnProps {
  editor: Editor | null;
  // Must be memoized: each change measures the cards again.
  anchors: ReadonlyArray<ReviewCardAnchor>;
  activeKey: string | null;
  renderCard: (key: string) => ReactNode;
  onPressInside: () => void;
}

// Places each card beside the text it belongs to. The panel decides which
// cards exist and renders them; this column only measures and positions.
export function DocumentReviewListColumn({
  editor,
  anchors,
  activeKey,
  renderCard,
  onPressInside,
}: DocumentReviewListColumnProps) {
  const { columnRef, bindCard, tops, height } = useReviewCardTops(editor, anchors, activeKey);
  const [settled, setSettled] = useState<ReadonlySet<string>>(() => new Set());
  // A frame later, once a card's first placement has been styled without a
  // transition. Sooner, the browser would animate it from top 0.
  useEffect(() => {
    const frame = window.requestAnimationFrame(() => {
      setSettled(new Set(Object.keys(tops)));
    });
    return () => window.cancelAnimationFrame(frame);
  }, [tops]);
  const { column, card } = documentReviewList();

  return (
    <div ref={columnRef} className={column()} style={{ minHeight: height }}>
      {anchors.map(({ key }) => (
        <div
          key={key}
          ref={bindCard}
          data-review-card={key}
          className={card({ placed: tops[key] != null, settled: settled.has(key) })}
          style={{ top: tops[key] ?? 0 }}
          onPointerDownCapture={onPressInside}
        >
          {renderCard(key)}
        </div>
      ))}
    </div>
  );
}

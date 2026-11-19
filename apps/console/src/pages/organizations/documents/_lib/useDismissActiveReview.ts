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

import { useCallback, useEffect, useLayoutEffect, useRef } from "react";

// useDismissActiveReview clears the active review item when a press lands
// outside every review card. The returned handler goes on each card's
// onPointerDownCapture: React events travel through portals, so presses in a
// card's menus and dialogs count as inside. The document listener captures
// the press before any handler can stop it, then decides on the next task,
// once React has marked it. Presses on highlighted text are left alone, since
// their click activates the matching card.
export function useDismissActiveReview(active: boolean, onDismiss: () => void) {
  const insideRef = useRef(false);
  const onDismissRef = useRef(onDismiss);

  useLayoutEffect(() => {
    onDismissRef.current = onDismiss;
  }, [onDismiss]);

  useEffect(() => {
    if (!active) {
      return;
    }
    insideRef.current = false;
    let timeout = 0;

    const onPointerDown = (event: PointerEvent) => {
      const target = event.target;
      window.clearTimeout(timeout);
      timeout = window.setTimeout(() => {
        const inside = insideRef.current;
        insideRef.current = false;
        if (inside || isReviewAnchor(target)) {
          return;
        }
        onDismissRef.current();
      });
    };

    document.addEventListener("pointerdown", onPointerDown, true);
    return () => {
      document.removeEventListener("pointerdown", onPointerDown, true);
      window.clearTimeout(timeout);
    };
  }, [active]);

  return useCallback(() => {
    insideRef.current = true;
  }, []);
}

// Comment highlights carry data-comment-thread, and suggestion marks and
// blocks carry a data-suggestion-* attribute.
function isReviewAnchor(target: EventTarget | null): boolean {
  for (let node = target instanceof Element ? target : null; node; node = node.parentElement) {
    for (const { name } of Array.from(node.attributes)) {
      if (name === "data-comment-thread" || name.startsWith("data-suggestion-")) {
        return true;
      }
    }
  }
  return false;
}

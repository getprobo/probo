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

import { type ReactNode, useCallback, useEffect, useState } from "react";

import { ComboboxJoinedProvider } from "./context";
import { comboboxJoined } from "./variants";

export type ComboboxJoinedProps = {
  // When true, the frame around the field is shown and the joined popup
  // continues that card. The field itself does not change size.
  open: boolean;
  field: ReactNode;
  children: ReactNode;
};

// Field plus the list popup that should stick to it. Wrap both so the popup
// can reuse the frame width without each caller measuring it.
export function ComboboxJoined(props: ComboboxJoinedProps) {
  const { open, field, children } = props;
  const [anchor, setAnchor] = useState<HTMLDivElement | null>(null);
  const [width, setWidth] = useState(0);
  const { frame, field: fieldClass, input } = comboboxJoined();

  const setFrame = useCallback((node: HTMLDivElement | null) => {
    setAnchor(node);
    if (node != null) {
      setWidth(node.offsetWidth);
    }
  }, []);

  useEffect(() => {
    if (anchor == null) {
      return;
    }

    const observer = new ResizeObserver((entries) => {
      const entry = entries[0];
      if (entry == null) {
        return;
      }

      const box = entry.borderBoxSize?.[0];
      const next = box?.inlineSize ?? entry.contentRect.width;
      setWidth(current => current === next ? current : next);
    });
    observer.observe(anchor);

    return () => {
      observer.disconnect();
    };
  }, [anchor]);

  return (
    <ComboboxJoinedProvider value={{ anchor, width }}>
      <div className={fieldClass()}>
        <div
          ref={setFrame}
          className={open ? frame() : "pointer-events-none absolute -inset-2 invisible"}
          aria-hidden
        />
        <div className={input()}>{field}</div>
      </div>
      {children}
    </ComboboxJoinedProvider>
  );
}

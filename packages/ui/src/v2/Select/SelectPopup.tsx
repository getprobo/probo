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

import { Select as BaseSelect } from "@base-ui/react/select";
import type { ComponentProps } from "react";

import { selectPopup } from "./variants";

export type SelectPopupProps
  = & Omit<ComponentProps<typeof BaseSelect.Popup>, "className">
    & {
      className?: string;
      // Positioner placement passthrough.
      side?: ComponentProps<typeof BaseSelect.Positioner>["side"];
      align?: ComponentProps<typeof BaseSelect.Positioner>["align"];
      sideOffset?: ComponentProps<typeof BaseSelect.Positioner>["sideOffset"];
    };

// Portal + positioner + styled popup holding the select items.
export function SelectPopup(props: SelectPopupProps) {
  const {
    className, children,
    side = "bottom", align = "start", sideOffset = 4,
    ...popupProps
  } = props;

  return (
    <BaseSelect.Portal>
      {/* z-3 on the Positioner so the portaled root wins over in-page z-1. */}
      <BaseSelect.Positioner className="z-3" side={side} align={align} sideOffset={sideOffset}>
        <BaseSelect.Popup className={selectPopup({ className })} {...popupProps}>
          {children}
        </BaseSelect.Popup>
      </BaseSelect.Positioner>
    </BaseSelect.Portal>
  );
}

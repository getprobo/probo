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

import { Tooltip as BaseTooltip } from "@base-ui/react/tooltip";
import type { ComponentProps } from "react";

import { tooltipPopup } from "./variants";

export type TooltipPopupProps
  = & Omit<ComponentProps<typeof BaseTooltip.Popup>, "className">
    & {
      className?: string;
      side?: ComponentProps<typeof BaseTooltip.Positioner>["side"];
      align?: ComponentProps<typeof BaseTooltip.Positioner>["align"];
      sideOffset?: ComponentProps<typeof BaseTooltip.Positioner>["sideOffset"];
    };

export function TooltipPopup(props: TooltipPopupProps) {
  const {
    className, children,
    side = "top", align = "center", sideOffset = 6,
    ...popupProps
  } = props;

  return (
    <BaseTooltip.Portal>
      {/* z-3 on the Positioner so the portaled root wins over in-page z-1. */}
      <BaseTooltip.Positioner className="z-3" side={side} align={align} sideOffset={sideOffset}>
        <BaseTooltip.Popup className={tooltipPopup({ className })} {...popupProps}>
          {children}
        </BaseTooltip.Popup>
      </BaseTooltip.Positioner>
    </BaseTooltip.Portal>
  );
}

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

import { Combobox as BaseCombobox } from "@base-ui/react/combobox";
import type { ComponentProps } from "react";

import { useComboboxInputNode } from "./context";
import { comboboxPopup } from "./variants";

export type ComboboxPopupProps
  = & Omit<ComponentProps<typeof BaseCombobox.Popup>, "className">
    & {
      className?: string;
      // Mount inside a modal/drawer stacking context (body portal uses z-3,
      // below drawers at z-5).
      container?: ComponentProps<typeof BaseCombobox.Portal>["container"];
      // Positioner placement passthrough.
      side?: ComponentProps<typeof BaseCombobox.Positioner>["side"];
      align?: ComponentProps<typeof BaseCombobox.Positioner>["align"];
      sideOffset?: ComponentProps<typeof BaseCombobox.Positioner>["sideOffset"];
    };

// Portal + positioner + styled popup holding the filtered items.
export function ComboboxPopup(props: ComboboxPopupProps) {
  const {
    className, children, container,
    side = "bottom", align = "start", sideOffset = 4,
    ...popupProps
  } = props;
  const inputNode = useComboboxInputNode();

  return (
    <BaseCombobox.Portal container={container}>
      {/* z-3 on the Positioner so the portaled root wins over in-page z-1.
          Anchor the input (caret), not the chip group; flip start/end when
          the preferred side runs out of horizontal room. */}
      <BaseCombobox.Positioner
        className="z-3"
        anchor={inputNode ?? undefined}
        side={side}
        align={align}
        sideOffset={sideOffset}
        collisionAvoidance={{ align: "flip", fallbackAxisSide: "none" }}
      >
        <BaseCombobox.Popup className={comboboxPopup({ className })} {...popupProps}>
          {children}
        </BaseCombobox.Popup>
      </BaseCombobox.Positioner>
    </BaseCombobox.Portal>
  );
}

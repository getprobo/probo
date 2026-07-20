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

import { Menu } from "@base-ui/react/menu";
import type { ComponentProps } from "react";

import { DropdownProvider } from "./context";
import { dropdownPopup } from "./variants";

export type DropdownPopupProps
  = & Omit<ComponentProps<typeof Menu.Popup>, "className">
    & {
      className?: string;
      // Menu-level look applied to every item.
      size?: 1 | 2;
      variant?: "solid" | "soft";
      highContrast?: boolean;
      // Positioner placement passthrough.
      side?: ComponentProps<typeof Menu.Positioner>["side"];
      align?: ComponentProps<typeof Menu.Positioner>["align"];
      sideOffset?: ComponentProps<typeof Menu.Positioner>["sideOffset"];
    };

// Portal + positioner + styled popup. Provides the menu-level look to its items.
export function DropdownPopup(props: DropdownPopupProps) {
  const {
    className, children,
    size = 2, variant = "solid", highContrast = false,
    side = "bottom", align = "start", sideOffset = 4,
    ...popupProps
  } = props;

  return (
    <Menu.Portal>
      {/* z-3 on the Positioner so the portaled root wins over in-page z-1. */}
      <Menu.Positioner className="z-3" side={side} align={align} sideOffset={sideOffset}>
        <Menu.Popup className={dropdownPopup({ className })} {...popupProps}>
          <DropdownProvider value={{ size, variant, highContrast }}>
            {children}
          </DropdownProvider>
        </Menu.Popup>
      </Menu.Positioner>
    </Menu.Portal>
  );
}

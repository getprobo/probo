// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
      <Menu.Positioner side={side} align={align} sideOffset={sideOffset}>
        <Menu.Popup className={dropdownPopup({ className })} {...popupProps}>
          <DropdownProvider value={{ size, variant, highContrast }}>
            {children}
          </DropdownProvider>
        </Menu.Popup>
      </Menu.Positioner>
    </Menu.Portal>
  );
}

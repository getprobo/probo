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
import type { ComponentProps, ReactNode } from "react";

import { useDropdownContext } from "./context";
import { dropdownItem, dropdownItemLabel, dropdownItemShortcut } from "./variants";

export type DropdownItemProps
  = & Omit<ComponentProps<typeof Menu.Item>, "color" | "className">
    & {
      className?: string;
      color?: "accent" | "error";
      iconStart?: ReactNode;
      // Trailing keyboard shortcut hint (e.g. "⌘ E").
      shortcut?: ReactNode;
    };

// A single actionable menu item. Inherits size/variant/highContrast from the
// popup; set `color="error"` for destructive actions.
export function DropdownItem(props: DropdownItemProps) {
  const { className, color = "accent", iconStart, shortcut, children, ...rest } = props;
  const { size, variant, highContrast } = useDropdownContext();

  return (
    <Menu.Item
      className={dropdownItem({ size, variant, color, highContrast, className })}
      {...rest}
    >
      {iconStart}
      <span className={dropdownItemLabel()}>{children}</span>
      {shortcut != null && <span className={dropdownItemShortcut()}>{shortcut}</span>}
    </Menu.Item>
  );
}

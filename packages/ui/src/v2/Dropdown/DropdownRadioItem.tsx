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
import { dropdownItem, dropdownItemIndicator, dropdownItemLabel, dropdownItemShortcut } from "./variants";

export type DropdownRadioItemProps
  = & Omit<ComponentProps<typeof Menu.RadioItem>, "className">
    & {
      className?: string;
      shortcut?: ReactNode;
    };

// A menu item with a radio dot indicator. Used inside a DropdownRadioGroup.
export function DropdownRadioItem(props: DropdownRadioItemProps) {
  const { className, shortcut, children, ...rest } = props;
  const { size, variant, highContrast } = useDropdownContext();

  return (
    <Menu.RadioItem
      className={dropdownItem({ size, variant, color: "accent", highContrast, className })}
      {...rest}
    >
      <span className={dropdownItemIndicator()}>
        <Menu.RadioItemIndicator className="size-1.5 rounded-full bg-current" />
      </span>
      <span className={dropdownItemLabel()}>{children}</span>
      {shortcut != null && <span className={dropdownItemShortcut()}>{shortcut}</span>}
    </Menu.RadioItem>
  );
}

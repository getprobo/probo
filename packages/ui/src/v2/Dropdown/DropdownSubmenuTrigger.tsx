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
import { CaretRightIcon } from "@phosphor-icons/react";
import type { ComponentProps, ReactNode } from "react";

import { useDropdownContext } from "./context";
import { dropdownItem, dropdownItemLabel, dropdownSubmenuCaret } from "./variants";

export type DropdownSubmenuTriggerProps
  = & Omit<ComponentProps<typeof Menu.SubmenuTrigger>, "className">
    & {
      className?: string;
      iconStart?: ReactNode;
    };

// The item that opens a submenu; renders a trailing caret.
export function DropdownSubmenuTrigger(props: DropdownSubmenuTriggerProps) {
  const { className, iconStart, children, ...rest } = props;
  const { size, variant, highContrast } = useDropdownContext();

  return (
    <Menu.SubmenuTrigger
      className={dropdownItem({ size, variant, color: "accent", highContrast, className })}
      {...rest}
    >
      {iconStart}
      <span className={dropdownItemLabel()}>{children}</span>
      <span className={dropdownSubmenuCaret()}>
        <CaretRightIcon />
      </span>
    </Menu.SubmenuTrigger>
  );
}

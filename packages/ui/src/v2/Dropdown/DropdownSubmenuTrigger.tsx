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

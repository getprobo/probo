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
import { CheckIcon } from "@phosphor-icons/react";
import type { ComponentProps, ReactNode } from "react";

import { useDropdownContext } from "./context";
import { dropdownItem, dropdownItemIndicator, dropdownItemLabel, dropdownItemShortcut } from "./variants";

export type DropdownCheckboxItemProps
  = & Omit<ComponentProps<typeof Menu.CheckboxItem>, "className">
    & {
      className?: string;
      shortcut?: ReactNode;
      // `end` puts the check in the trailing group, before `shortcut`.
      indicator?: "start" | "end";
    };

// A menu item with a checkmark indicator. Controlled with `checked` /
// `onCheckedChange`, like Base UI's CheckboxItem.
export function DropdownCheckboxItem(props: DropdownCheckboxItemProps) {
  const { className, shortcut, indicator = "end", children, ...rest } = props;
  const { size, variant, highContrast } = useDropdownContext();
  const check = (
    <Menu.CheckboxItemIndicator className={dropdownItemIndicator()}>
      <CheckIcon weight="bold" />
    </Menu.CheckboxItemIndicator>
  );

  return (
    <Menu.CheckboxItem
      className={dropdownItem({ size, variant, color: "accent", highContrast, className })}
      {...rest}
    >
      {indicator === "start" && check}
      <span className={dropdownItemLabel()}>{children}</span>
      {indicator === "end" && check}
      {shortcut != null && <span className={dropdownItemShortcut()}>{shortcut}</span>}
    </Menu.CheckboxItem>
  );
}

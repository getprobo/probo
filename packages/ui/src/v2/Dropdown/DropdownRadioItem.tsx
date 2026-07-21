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

export type DropdownRadioItemProps
  = & Omit<ComponentProps<typeof Menu.RadioItem>, "className">
    & {
      className?: string;
      shortcut?: ReactNode;
    };

// A menu item with a checkmark for the selected value. Used inside a
// DropdownRadioGroup (same indicator treatment as DropdownCheckboxItem).
export function DropdownRadioItem(props: DropdownRadioItemProps) {
  const { className, shortcut, children, ...rest } = props;
  const { size, variant, highContrast } = useDropdownContext();

  return (
    <Menu.RadioItem
      className={dropdownItem({ size, variant, color: "accent", highContrast, className })}
      {...rest}
    >
      <span className={dropdownItemIndicator()}>
        <Menu.RadioItemIndicator>
          <CheckIcon weight="bold" />
        </Menu.RadioItemIndicator>
      </span>
      <span className={dropdownItemLabel()}>{children}</span>
      {shortcut != null && <span className={dropdownItemShortcut()}>{shortcut}</span>}
    </Menu.RadioItem>
  );
}

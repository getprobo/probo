// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import {
  Combobox as AriaKitCombobox,
  ComboboxItem as AriaKitComboboxItem,
  ComboboxPopover,
  ComboboxProvider,
} from "@ariakit/react";
import { isEmpty } from "@probo/helpers";
import {
  type ComponentProps,
  type PropsWithChildren,
  type ReactNode,
} from "react";

import { dropdown, dropdownItem } from "../../Atoms/Dropdown/Dropdown";
import { input } from "../../Atoms/Input/Input";

type Props = {
  children: ReactNode;
  loading?: boolean;
  onSearch: (query: string) => void;
  placeholder?: string;
  autoSelect?: boolean;
  onSelect?: (value: string) => void;
  resetValueOnHide?: boolean;
  value?: string;
} & Omit<ComponentProps<typeof AriaKitCombobox>, "onSelect" | "value">;

export function Combobox({
  children,
  onSearch,
  placeholder,
  autoSelect,
  onSelect,
  resetValueOnHide,
  value,
  ...props
}: Props) {
  const showDropdown = !isEmpty(children);
  return (
    <ComboboxProvider
      value={value}
      setValue={onSearch}
      setSelectedValue={v => onSelect?.(v as string)}
      resetValueOnHide={resetValueOnHide}
    >
      <AriaKitCombobox
        {...props}
        autoSelect={autoSelect}
        placeholder={placeholder}
        className={input()}
      />
      {showDropdown && (
        <ComboboxPopover
          gutter={4}
          sameWidth
          className={dropdown()}
          style={{ maxHeight: "var(--popover-available-height)" }}
        >
          {children}
        </ComboboxPopover>
      )}
    </ComboboxProvider>
  );
}

export function ComboboxItem({
  children,
  ...props
}: PropsWithChildren<ComponentProps<typeof AriaKitComboboxItem>>) {
  return (
    <AriaKitComboboxItem hideOnClick className={dropdownItem()} {...props}>
      {children}
    </AriaKitComboboxItem>
  );
}

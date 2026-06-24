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

import { Select as BaseSelect } from "@base-ui/react/select";
import { CaretDownIcon } from "@phosphor-icons/react";
import type { ComponentProps, ReactNode } from "react";

import { selectTrigger } from "./variants";

export type SelectTriggerProps
  = & Omit<ComponentProps<typeof BaseSelect.Trigger>, "className" | "children">
    & {
      className?: string;
      size?: 1 | 2;
      // Surface treatment (defaults to "surface").
      variant?: "classic" | "surface" | "soft" | "ghost";
      // Shown when no value is selected.
      placeholder?: ReactNode;
      // Renders the selected value; pass a function to map a value to its label.
      children?: BaseSelect.Value.Props["children"];
    };

// The bordered surface button that opens the popup, showing the selected value
// (or placeholder) and a trailing chevron.
export function SelectTrigger(props: SelectTriggerProps) {
  const { className, size, variant, placeholder, children, ...rest } = props;
  const { trigger, value, icon } = selectTrigger({ size, variant });

  return (
    <BaseSelect.Trigger className={trigger({ className })} {...rest}>
      <BaseSelect.Value className={value()} placeholder={placeholder}>
        {children}
      </BaseSelect.Value>
      <BaseSelect.Icon className={icon()}>
        <CaretDownIcon />
      </BaseSelect.Icon>
    </BaseSelect.Trigger>
  );
}

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

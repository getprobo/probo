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

import { Input } from "@base-ui/react/input";
import type { ComponentProps, ReactNode } from "react";

import { textField } from "./variants";

export type TextFieldProps
  = & Omit<ComponentProps<typeof Input>, "className">
    & {
      // Applied to the bordered container (the top-level element).
      className?: string;
      size?: 1 | 2;
      // Surface treatment (defaults to "surface").
      variant?: "classic" | "surface" | "soft";
      // Leading icon slot (e.g. a MagnifyingGlassIcon for search).
      icon?: ReactNode;
    };

// Single-line text input on a bordered surface. The container is the top-level
// node; all native input props spread onto the inner control. Controlled with
// Base UI's `value` / `onValueChange` (or native `value` / `onChange`).
export function TextField(props: TextFieldProps) {
  const { className, size, variant, icon, ...inputProps } = props;
  const { root, icon: iconSlot, input } = textField({ size, variant });

  return (
    <div className={root({ className })}>
      {icon != null && <span className={iconSlot()}>{icon}</span>}
      <Input className={input()} {...inputProps} />
    </div>
  );
}

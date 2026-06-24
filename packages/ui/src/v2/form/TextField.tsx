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

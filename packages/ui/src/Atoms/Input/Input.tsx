// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import { type FC, type InputHTMLAttributes, useState } from "react";
import { tv } from "tailwind-variants";

import type { IconProps } from "../Icons/type";

type Props = {
  invalid?: boolean;
  disabled?: boolean;
  icon?: FC<IconProps>;
  onValueChange?: (value: string) => void;
  variant?: "bordered" | "ghost" | "title";
} & InputHTMLAttributes<HTMLInputElement>;

export const input = tv({
  base: "w-full disabled:bg-transparent",
  variants: {
    invalid: {
      true: "border-border-danger",
    },
    variant: {
      bordered:
                "py-[6px] bg-secondary border border-border-mid rounded-[10px] hover:border-border-strong text-sm px-3 text-txt-primary",
      ghost: "text-base text-txt-secondary outline-none",
      title: "text-2xl font-semibold text-txt-primary outline-none",
    },
  },
  defaultVariants: {
    variant: "bordered",
  },
});

export function Input({
  invalid,
  icon: IconComponent,
  onValueChange,
  ...props
}: Props) {
  const [focus, setFocus] = useState(false);

  if (IconComponent) {
    return (
      <div
        className={input({
          className: "flex items-center gap-2",
        })}
        data-focus={focus}
      >
        <IconComponent size={16} className="text-txt-secondary" />
        <input
          {...props}
          onFocus={(e) => {
            setFocus(true);
            props.onFocus?.(e);
          }}
          onBlur={(e) => {
            setFocus(false);
            props.onBlur?.(e);
          }}
          aria-invalid={invalid}
          className="w-full outline-none"
          onChange={(e) => {
            onValueChange?.(e.currentTarget.value);
            props.onChange?.(e);
          }}
        />
      </div>
    );
  }
  return (
    <input
      {...props}
      aria-invalid={invalid}
      className={input(props)}
      onChange={(e) => {
        onValueChange?.(e.currentTarget.value);
        props.onChange?.(e);
      }}
    />
  );
}

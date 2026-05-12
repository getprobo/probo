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

import type { HTMLAttributes } from "react";
import { tv } from "tailwind-variants";

import { Slot } from "../Slot";

type Props = {
  asChild?: boolean;
  variant?:
    | "success"
    | "warning"
    | "danger"
    | "info"
    | "neutral"
    | "outline"
    | "highlight";
  size?: "sm" | "md";
} & HTMLAttributes<HTMLElement>;

const badge = tv({
  base: "font-medium rounded-lg w-max flex gap-1 items-center group whitespace-nowrap",
  variants: {
    variant: {
      success: "bg-success text-txt-success",
      warning: "bg-warning text-txt-warning",
      danger: "bg-danger text-txt-danger",
      info: "bg-info text-txt-info",
      neutral: "bg-subtle text-txt-secondary",
      outline: "text-txt-tertiary border border-border-low",
      highlight: "bg-highlight text-txt-primary",
    },
    size: {
      sm: "text-xs py-[2px] px-[6px]",
      md: "text-sm py-[6px] px-2",
    },
  },
  defaultVariants: {
    variant: "neutral",
    size: "sm",
  },
});

export function Badge(props: Props) {
  const Component = props.asChild ? Slot : "div";
  const { className, size, variant, ...restProps } = props;
  return (
    <Component {...restProps} className={badge({ className, size, variant })} />
  );
}

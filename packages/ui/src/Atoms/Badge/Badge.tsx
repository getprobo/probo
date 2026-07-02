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

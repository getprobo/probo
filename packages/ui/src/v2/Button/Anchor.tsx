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

import type { ComponentProps, ReactNode } from "react";
import type { VariantProps } from "tailwind-variants/lite";

import { button } from "./variants";

export type AnchorProps
  = Omit<ComponentProps<"a">, "color">
    & VariantProps<typeof button>
    & {
      iconStart?: ReactNode;
      iconEnd?: ReactNode;
    };

// Styled <a> sharing the Button look (Radix "Button"). Use for external links
// or mailto:/tel:; for in-app navigation use Link (router). See
// contrib/claude/ui.md.
export function Anchor(props: AnchorProps) {
  const {
    size, variant, color, highContrast, active, className,
    iconStart, iconEnd, children, ...rest
  } = props;

  return (
    <a
      className={button({ size, variant, color, highContrast, active, className })}
      {...rest}
    >
      {iconStart}
      {children}
      {iconEnd}
    </a>
  );
}

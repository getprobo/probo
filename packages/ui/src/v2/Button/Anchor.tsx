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

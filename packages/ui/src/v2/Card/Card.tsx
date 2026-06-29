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

import type { ComponentProps } from "react";
import type { VariantProps } from "tailwind-variants/lite";

import { CardProvider } from "./context";
import { card } from "./variants";

export type CardProps = ComponentProps<"div"> & VariantProps<typeof card>;

// Container that groups related content (Radix "Card"). Renders a <div>; for a
// clickable card set `interactive` and wrap the content in an anchor/link. Use
// CardInset for content that should bleed to the card's edges.
export function Card(props: CardProps) {
  const { size = 1, padding = size, variant, interactive, className, children, ...rest } = props;

  return (
    <div className={card({ size, padding, variant, interactive, className })} {...rest}>
      <CardProvider value={padding}>{children}</CardProvider>
    </div>
  );
}

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

import { InfoIcon } from "@phosphor-icons/react";
import type { ComponentProps, ReactNode } from "react";
import type { VariantProps } from "tailwind-variants/lite";

import { callout } from "./variants";

export type CalloutProps
  = Omit<ComponentProps<"div">, "color">
    & VariantProps<typeof callout>
    & {
      // Leading icon. Defaults to an info icon; pass `null` to omit it.
      icon?: ReactNode;
    };

// Short contextual message (Radix "Callout") with a leading icon. See
// contrib/claude/ui.md.
export function Callout(props: CalloutProps) {
  const {
    size, variant, color, highContrast, className,
    icon = <InfoIcon weight="fill" />, children, ...rest
  } = props;
  const slots = callout({ size, variant, color, highContrast });

  return (
    <div className={slots.root({ className })} {...rest}>
      {icon !== null && (
        <span aria-hidden className={slots.icon()}>
          {icon}
        </span>
      )}
      <div className={slots.text()}>{children}</div>
    </div>
  );
}

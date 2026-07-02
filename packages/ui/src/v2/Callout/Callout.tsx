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

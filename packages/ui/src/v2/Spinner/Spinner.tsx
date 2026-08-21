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

import type { ComponentProps } from "react";
import type { VariantProps } from "tailwind-variants/lite";

import { spinner } from "./variants";

const LEAF_COUNT = 8;

export type SpinnerProps
  = Omit<ComponentProps<"span">, "children">
    & VariantProps<typeof spinner>;

// Eight-leaf loading indicator (Radix "Spinner"). Pass `aria-label` when the
// spinner is the only indication of a pending state. See contrib/claude/ui.md.
export function Spinner(props: SpinnerProps) {
  const { size, className, ...rest } = props;
  const labelled = rest["aria-label"] != null || rest["aria-labelledby"] != null;
  const { root, spoke, leaf } = spinner({ size });

  return (
    <span
      className={root({ className })}
      role={labelled ? "status" : undefined}
      aria-hidden={labelled ? undefined : true}
      {...rest}
    >
      {Array.from({ length: LEAF_COUNT }, (_, index) => (
        <span key={index} className={spoke()}>
          <span className={leaf()} />
        </span>
      ))}
    </span>
  );
}

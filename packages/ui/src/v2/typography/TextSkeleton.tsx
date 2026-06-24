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

import { textSkeleton } from "./variants";

export type TextSkeletonProps = Omit<ComponentProps<"span">, "children"> & VariantProps<typeof textSkeleton>;

// Loading placeholder paired with Text: a pulse block matching a line of text
// at the given size. Supply a width class (e.g. w-24, or w-full for a full
// line); with none the block collapses to its content.
export function TextSkeleton(props: TextSkeletonProps) {
  const { size, className, ...rest } = props;

  return (
    <span className={textSkeleton({ size, className })} {...rest} aria-hidden>
      {"\u00A0"}
    </span>
  );
}

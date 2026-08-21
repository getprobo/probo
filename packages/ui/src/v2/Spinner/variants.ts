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

import { tv } from "tailwind-variants/lite";

// Eight-leaf loading indicator (Radix "Spinner"). The root rotates; each spoke
// is a radial leaf whose opacity ramps sand-a11 → sand-a4.
export const spinner = tv({
  slots: {
    root: [
      "relative inline-block animate-spin",
      "[&>:nth-child(1)]:rotate-0 [&>:nth-child(1)>span]:bg-sand-a11",
      "[&>:nth-child(2)]:rotate-45 [&>:nth-child(2)>span]:bg-sand-a10",
      "[&>:nth-child(3)]:rotate-90 [&>:nth-child(3)>span]:bg-sand-a9",
      "[&>:nth-child(4)]:rotate-135 [&>:nth-child(4)>span]:bg-sand-a8",
      "[&>:nth-child(5)]:rotate-180 [&>:nth-child(5)>span]:bg-sand-a7",
      "[&>:nth-child(6)]:-rotate-135 [&>:nth-child(6)>span]:bg-sand-a6",
      "[&>:nth-child(7)]:-rotate-90 [&>:nth-child(7)>span]:bg-sand-a5",
      "[&>:nth-child(8)]:-rotate-45 [&>:nth-child(8)>span]:bg-sand-a4",
    ],
    spoke: "absolute inset-0",
    leaf: "absolute top-0 left-1/2 h-[30%] w-0.5 -translate-x-1/2 rounded-1",
  },
  variants: {
    size: {
      1: { root: "size-3" },
      2: { root: "size-4" },
      3: { root: "size-5" },
    },
  },
  defaultVariants: {
    size: 2,
  },
});

export const spinnerSkeleton = tv({
  base: "inline-block shrink-0 animate-pulse rounded-full bg-sand-3 align-middle",
  variants: {
    size: {
      1: "size-3",
      2: "size-4",
      3: "size-5",
    },
  },
  defaultVariants: {
    size: 2,
  },
});

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

// Vertical timeline of ordered steps. A 24px marker circle sits on a sand rail;
// callers own the icon and the content copy.
export const timeline = tv({
  slots: {
    root: "m-0 list-none p-0",
    item: [
      "relative flex items-start gap-3 pb-3 last:pb-0",
      "before:absolute before:top-6 before:bottom-0 before:left-3 before:w-px before:-translate-x-1/2 before:bg-sand-a6",
      "last:before:hidden",
    ],
    marker: [
      "relative z-1 flex size-6 shrink-0 items-center justify-center rounded-full",
      "[&_svg]:size-4",
    ],
    content: "flex min-w-0 flex-1 items-center justify-between gap-4 pt-0.5",
  },
  variants: {
    color: {
      neutral: { marker: "bg-sand-3 text-sand-11" },
      red: { marker: "bg-red-3 text-red-11" },
    },
  },
  defaultVariants: {
    color: "neutral",
  },
});

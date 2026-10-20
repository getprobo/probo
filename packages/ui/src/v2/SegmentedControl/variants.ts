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

// Segmented control (single-select radio-cards over Base UI's ToggleGroup).
// There is no surrounding track: items are standalone bordered pills. Size 2
// uses an equal-width card grid so a lone wrapped item does not stretch.
// Size 1 is a compact icon+label row with a gap. Each item keeps a 1px
// border at all sizes (the pressed state only darkens the border).
export const segmentedControl = tv({
  slots: {
    root: "",
    item: [
      "min-w-0 cursor-pointer select-none border border-sand-a6 bg-sand-1",
      "font-medium text-sand-12 outline-none transition-colors",
      "hover:border-sand-a8",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
      "data-pressed:border-sand-a12",
      "data-disabled:pointer-events-none data-disabled:opacity-50",
    ],
  },
  variants: {
    size: {
      1: {
        root: "flex flex-wrap gap-2",
        item: "inline-flex items-center gap-2 rounded-2 px-3 py-1.5 text-left text-2 [&_svg]:size-4",
      },
      2: {
        root: "grid grid-cols-[repeat(auto-fill,minmax(9rem,1fr))] gap-1",
        item: "rounded-3 px-4 py-3.5 text-center text-2",
      },
    },
  },
  defaultVariants: {
    size: 2,
  },
});

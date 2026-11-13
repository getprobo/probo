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

import { tooltipPopup } from "../Tooltip/variants";

export const chart = tv({
  slots: {
    root: "relative",
    frame: "w-full",
    svg: "block overflow-visible",
    tick: "fill-sand-11 font-sans text-1",
    line: "stroke-sand-6",
    grid: "[&_line]:stroke-sand-4",
    tooltip: [tooltipPopup(), "pointer-events-none absolute z-2"],
    legend: "m-0 flex list-none flex-wrap items-center gap-x-4 gap-y-2 p-0",
    legendItem: "flex items-center gap-1.5 text-2 text-sand-11",
    bar: "pointer-events-none",
    swatch: "size-2 shrink-0 rounded-1",
    skeleton: "h-60 animate-pulse rounded-3 bg-sand-3",
  },
});

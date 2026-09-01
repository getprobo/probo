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

// Semantic table (Radix Themes "Table"). Surface chrome matches List/Card
// `soft`. Size steps follow Themes (cell padding + type); ghost has no frame.
export const table = tv({
  slots: {
    root: "w-full overflow-x-auto overflow-y-hidden",
    table: "w-full border-collapse text-left align-top",
    header: "",
    body: "[&_tr:last-child_td]:border-b-0 [&_tr:last-child_th]:border-b-0",
    row: "text-sand-12",
    cell: "box-border border-b border-sand-a3 align-inherit",
    columnHeader: "font-medium text-sand-11",
    rowHeader: "font-normal",
  },
  variants: {
    size: {
      1: {
        root: "rounded-3",
        table: "text-2",
        cell: "h-9 px-2 py-2",
      },
      2: {
        root: "rounded-4",
        table: "text-2",
        cell: "h-11 px-3 py-3",
      },
      3: {
        root: "rounded-4",
        table: "text-3",
        cell: "h-12 px-4 py-3",
      },
    },
    variant: {
      // Same chrome as List / Card `soft`. No thead wash: sand-a2 on sand-1
      // reconstitutes the page canvas (sand-2) and the header disappears.
      surface: {
        root: "overflow-hidden border border-sand-a3 bg-sand-1",
      },
      ghost: {
        root: "",
      },
    },
    layout: {
      auto: {
        table: "table-auto",
      },
      fixed: {
        table: "table-fixed",
      },
    },
    align: {
      start: {
        row: "align-top",
      },
      center: {
        row: "align-middle",
      },
      end: {
        row: "align-bottom",
      },
      baseline: {
        row: "align-baseline",
      },
    },
    justify: {
      start: {
        cell: "text-start",
      },
      center: {
        cell: "text-center",
      },
      end: {
        cell: "text-end",
      },
    },
    // Look-only. On the row: containing block + click-through so a TableLink
    // ::after covers padding and sibling cells. On a cell: lift trailing
    // controls above that overlay. See contrib/claude/ui.md.
    interactive: {
      true: {
        row: [
          "relative isolate cursor-pointer",
          "[&_td]:pointer-events-none [&_th]:pointer-events-none",
        ],
        // Important: the row's `[&_td]:pointer-events-none` is a descendant
        // selector and otherwise beats a plain `pointer-events-auto` on the
        // cell, so the overlay steals :hover from trailing controls.
        cell: "relative z-1 pointer-events-auto!",
      },
    },
  },
  defaultVariants: {
    size: 2,
    variant: "ghost",
    layout: "auto",
    interactive: false,
  },
});

// Stretched in-row link (TableLink). ::after is positioned against the
// interactive TableRow; do not make this relative or the overlay shrinks
// to the title. Unstyled text — not the underlined Link recipe.
export const tableLink = tv({
  base: [
    "min-w-0 pointer-events-auto",
    "after:absolute after:inset-0 after:content-['']",
    "outline-none focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
  ],
});

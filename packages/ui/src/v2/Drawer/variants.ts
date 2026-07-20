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

// Edge drawer over Base UI. Positioning lives in CSS (`side`); pass the matching
// `swipeDirection` on `Drawer` (right↔right, left↔left, bottom↔down, top↔up).
export const drawer = tv({
  slots: {
    backdrop: [
      "fixed inset-0 z-50 bg-sand-12/40",
      "transition-opacity duration-200",
      "data-starting-style:opacity-0 data-ending-style:opacity-0",
    ],
    viewport: "fixed inset-0 z-50 flex",
    popup: [
      "relative flex flex-col bg-sand-1 shadow-6 outline-none",
      "transition-transform duration-200",
    ],
    content: "flex min-h-0 flex-1 flex-col gap-4 overflow-y-auto p-4",
    header: "flex items-center justify-between gap-3",
    title: "text-4 font-medium text-sand-12",
    description: "text-2 text-sand-11",
    body: "flex min-h-0 flex-1 flex-col gap-1",
    footer: "mt-auto flex flex-col gap-2 border-t border-sand-a3 pt-4",
  },
  variants: {
    side: {
      right: {
        viewport: "justify-end",
        popup: [
          "h-full w-[min(20rem,100%)]",
          "transform-[translateX(var(--drawer-swipe-movement-x,0px))]",
          "data-starting-style:translate-x-full data-ending-style:translate-x-full",
        ],
      },
      left: {
        viewport: "justify-start",
        popup: [
          "h-full w-[min(20rem,100%)]",
          "transform-[translateX(var(--drawer-swipe-movement-x,0px))]",
          "data-starting-style:-translate-x-full data-ending-style:-translate-x-full",
        ],
      },
      bottom: {
        viewport: "items-end justify-center",
        popup: [
          "w-full max-h-[min(90dvh,100%)] rounded-t-5",
          "transform-[translateY(var(--drawer-swipe-movement-y,0px))]",
          "data-starting-style:translate-y-full data-ending-style:translate-y-full",
        ],
      },
      top: {
        viewport: "items-start justify-center",
        popup: [
          "w-full max-h-[min(90dvh,100%)] rounded-b-5",
          "transform-[translateY(var(--drawer-swipe-movement-y,0px))]",
          "data-starting-style:-translate-y-full data-ending-style:-translate-y-full",
        ],
      },
    },
  },
  defaultVariants: {
    side: "right",
  },
});

// Static frame matching the drawer popup, without Base UI positioning, so a
// placeholder can render before the interactive drawer loads.
export const drawerSkeleton = tv({
  base: [
    "flex h-full w-[min(20rem,100%)] flex-col gap-4 bg-sand-1 p-4 shadow-6",
  ],
});

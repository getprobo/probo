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

export const organizationLayout = tv({
  slots: {
    root: "flex h-dvh flex-col bg-sand-2",
    body: "flex min-h-0 flex-1",
    content: "min-w-0 flex-1 overflow-y-auto transition-[padding] duration-300",
    contentInner: "mx-auto w-full max-w-300 px-8 py-12",
  },
  variants: {
    hasDrawer: {
      true: { content: "pr-105" },
    },
  },
});

export const organizationSwitcher = tv({
  slots: {
    popup: "w-72 p-0",
    search: "p-2",
    list: "max-h-96 overflow-y-auto px-1 pb-1",
  },
});

export const navRail = tv({
  slots: {
    // `root` holds the collapsed width in the flex row and the rail is absolute
    // inside it, so expanding overlays the panel instead of shoving the page.
    root: "relative w-14 shrink-0",
    rail: [
      "group absolute inset-y-0 left-0 z-2 flex w-14 flex-col gap-1 overflow-hidden",
      "border-r border-sand-a3 bg-sand-1 p-2",
      "transition-[width,box-shadow,border-color] duration-150 ease-out",
    ],
    item: [
      "flex h-10 w-full cursor-pointer items-center gap-2 rounded-3 text-sand-11 outline-none transition-colors",
      "hover:bg-sand-3 hover:text-sand-12",
      "focus-visible:ring-2 focus-visible:ring-sand-8",
    ],
    // overflow-x is pinned because `overflow-y-auto` alone would compute the
    // other axis to `auto` too, and items are wider than the collapsed rail.
    items: "flex min-h-0 flex-1 flex-col gap-1 overflow-x-hidden overflow-y-auto",
    icon: "flex size-10 shrink-0 items-center justify-center",
    label: [
      "min-w-0 flex-1 truncate pr-2 text-left opacity-0",
      "transition-opacity duration-150 ease-out",
      "group-hover:opacity-100 group-has-focus-visible:opacity-100",
      "group-has-data-popup-open:opacity-100",
    ],
    caret: [
      "mr-2 size-4 shrink-0 text-sand-11 opacity-0",
      "transition-opacity duration-150 ease-out",
      "group-hover:opacity-100 group-has-focus-visible:opacity-100",
      "group-has-data-popup-open:opacity-100",
    ],
  },
  variants: {
    active: {
      true: { item: "bg-gold-4 text-gold-12" },
    },
    // data-popup-open keeps the rail open while a menu portaled out of it has
    // focus. focus-visible rather than focus-within: closing a menu restores
    // focus to its in-rail trigger, which would leave the rail stuck open.
    expandable: {
      true: {
        rail: [
          "hover:w-56 hover:border-transparent hover:shadow-4",
          "has-focus-visible:w-56 has-focus-visible:border-transparent has-focus-visible:shadow-4",
          "has-data-popup-open:w-56 has-data-popup-open:border-transparent has-data-popup-open:shadow-4",
        ],
      },
    },
  },
  defaultVariants: {
    expandable: true,
  },
});

export const navPanel = tv({
  slots: {
    panel: "flex w-56 shrink-0 flex-col gap-1 border-r border-sand-a3 bg-sand-1 px-3 py-4",
    title: "px-1 pb-2",
    list: "flex flex-col gap-0.5",
    item: "w-full justify-start",
    group: "mt-3 flex w-full flex-col gap-1 first:mt-0",
    groupLabel: "px-3",
    groupFallback: "mt-2 mb-2 block h-9 w-full animate-pulse rounded-2 bg-sand-3 first:mt-0 last:mb-0",
  },
});

// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { tv } from "tailwind-variants/lite";

// Select (Radix "Select" over Base UI's Select). A bordered surface trigger on
// the neutral background, with an accent-highlighted popup of items.

export const selectTrigger = tv({
  slots: {
    trigger: [
      "flex w-full items-center justify-between gap-2 rounded-2 border border-sand-a7 bg-sand-1 text-2 text-sand-12",
      "cursor-pointer outline-none transition-colors hover:bg-sand-2",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
      "data-disabled:pointer-events-none data-disabled:opacity-50 data-placeholder:text-sand-a10",
    ],
    value: "min-w-0 flex-1 truncate text-left",
    icon: "flex size-4 shrink-0 items-center justify-center text-sand-a10 [&_svg]:size-4",
  },
  variants: {
    size: {
      1: { trigger: "h-7 px-2" },
      2: { trigger: "h-8 px-3" },
    },
  },
  defaultVariants: {
    size: 2,
  },
});

export const selectPopup = tv({
  base: [
    "max-h-(--available-height) min-w-(--anchor-width) origin-(--transform-origin) overflow-y-auto rounded-3 bg-sand-1 p-1 shadow-3 outline-none",
    "transition-[scale,opacity] duration-150 ease-out data-starting-style:scale-95 data-starting-style:opacity-0",
    "data-ending-style:scale-95 data-ending-style:opacity-0",
  ],
});

export const selectItem = tv({
  slots: {
    item: [
      "flex h-8 cursor-pointer items-center justify-between gap-2 rounded-2 px-3 text-2 text-sand-12 outline-none select-none",
      "data-disabled:pointer-events-none data-disabled:opacity-50 data-highlighted:bg-sand-3",
    ],
    label: "min-w-0 flex-1 truncate",
    indicator: "flex size-4 shrink-0 items-center justify-center text-sand-12 [&_svg]:size-4",
  },
});

export const selectSkeleton = tv({
  base: "inline-block animate-pulse rounded-2 bg-sand-3 align-middle",
  variants: {
    size: {
      1: "h-7 w-40",
      2: "h-8 w-40",
    },
  },
  defaultVariants: {
    size: 2,
  },
});

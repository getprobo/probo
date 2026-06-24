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

// Dropdown menu (Radix "Dropdown Menu" over Base UI's Menu). Accent-colored
// menu surface with grouped, highlightable items.

export const dropdownPopup = tv({
  base: [
    "min-w-40 origin-[var(--transform-origin)] rounded-3 border border-sand-6 bg-sand-1 p-2 shadow-5 outline-none",
    "transition-[transform,opacity] data-[starting-style]:scale-95 data-[starting-style]:opacity-0",
    "data-[ending-style]:scale-95 data-[ending-style]:opacity-0",
  ],
});

export const dropdownItem = tv({
  base: [
    "group flex cursor-pointer items-center gap-2 rounded-2 outline-none select-none",
    "data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
  ],
  variants: {
    size: {
      1: "h-7 px-2 text-1",
      2: "h-8 px-3 text-2",
    },
    // Highlight treatment resolves in the compound variants below.
    variant: {
      solid: "",
      soft: "",
    },
    color: {
      accent: "text-sand-12",
      error: "text-red-11",
    },
    highContrast: {
      true: "",
      false: "",
    },
  },
  compoundVariants: [
    // accent
    { variant: "solid", color: "accent", class: "data-[highlighted]:bg-gold-9 data-[highlighted]:text-white" },
    { variant: "soft", color: "accent", class: "data-[highlighted]:bg-gold-4 data-[highlighted]:text-gold-12" },
    { color: "accent", highContrast: true, class: "text-sand-12" },
    // error
    { variant: "solid", color: "error", class: "data-[highlighted]:bg-red-9 data-[highlighted]:text-white" },
    { variant: "soft", color: "error", class: "data-[highlighted]:bg-red-4 data-[highlighted]:text-red-12" },
  ],
  defaultVariants: {
    size: 2,
    variant: "solid",
    color: "accent",
    highContrast: false,
  },
});

export const dropdownItemLabel = tv({
  base: "min-w-0 flex-1 truncate",
});

// Keyboard shortcut hint: trailing and muted, but inherits the item's color
// once the item is highlighted (so it tracks the contrast text).
export const dropdownItemShortcut = tv({
  base: "shrink-0 text-sand-11 group-data-[highlighted]:text-inherit",
});

export const dropdownItemIndicator = tv({
  base: "flex size-4 shrink-0 items-center justify-center [&_svg]:size-4",
});

export const dropdownSubmenuCaret = tv({
  base: "size-4 shrink-0 [&_svg]:size-4",
});

export const dropdownSeparator = tv({
  base: "-mx-2 my-1 h-px bg-sand-6",
});

export const dropdownGroupLabel = tv({
  base: "px-3 py-1 text-1 font-medium text-sand-11",
});

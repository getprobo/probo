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

// Dropdown menu (Radix "Dropdown Menu" over Base UI's Menu). Accent-colored
// menu surface with grouped, highlightable items.

export const dropdownPopup = tv({
  base: [
    "min-w-40 origin-(--transform-origin) rounded-3 bg-sand-1 p-2 shadow-5 outline-none",
    "transition-[scale,opacity] duration-150 ease-out data-starting-style:scale-95 data-starting-style:opacity-0",
    "data-ending-style:scale-95 data-ending-style:opacity-0",
  ],
});

export const dropdownItem = tv({
  base: [
    "group flex cursor-pointer items-center gap-2 rounded-2 outline-none select-none",
    "data-disabled:pointer-events-none data-disabled:opacity-50",
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
      success: "text-green-11",
    },
    highContrast: {
      true: "",
      false: "",
    },
  },
  compoundVariants: [
    // accent
    { variant: "solid", color: "accent", class: "data-highlighted:bg-gold-9 data-highlighted:text-white" },
    { variant: "soft", color: "accent", class: "data-highlighted:bg-gold-4 data-highlighted:text-gold-12" },
    { color: "accent", highContrast: true, class: "text-sand-12" },
    // error
    { variant: "solid", color: "error", class: "data-highlighted:bg-red-9 data-highlighted:text-white" },
    { variant: "soft", color: "error", class: "data-highlighted:bg-red-4 data-highlighted:text-red-12" },
    // success
    { variant: "solid", color: "success", class: "data-highlighted:bg-green-9 data-highlighted:text-white" },
    { variant: "soft", color: "success", class: "data-highlighted:bg-green-4 data-highlighted:text-green-12" },
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
  base: "shrink-0 text-sand-11 group-data-highlighted:text-inherit",
});

export const dropdownItemIndicator = tv({
  base: "flex size-4 shrink-0 items-center justify-center [&_svg]:size-4",
});

export const dropdownSubmenuCaret = tv({
  base: "size-4 shrink-0 [&_svg]:size-4",
});

export const dropdownSeparator = tv({
  base: "mx-3 my-1 h-px bg-sand-a3",
});

export const dropdownGroupLabel = tv({
  base: "px-3 py-1 text-1 font-medium text-sand-11",
});

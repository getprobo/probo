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

// Combobox (Radix-style, over Base UI). A wrapping input surface that can
// hold chips, with a select-like popup of filtered items.

export const comboboxInputGroup = tv({
  base: [
    "flex min-h-8 w-full flex-wrap items-center gap-1.5 rounded-2 border border-sand-a5 bg-sand-1 px-2 py-1 text-2 text-sand-12",
    "transition-colors",
    "focus-within:ring-2 focus-within:ring-sand-8 focus-within:ring-offset-1 focus-within:ring-offset-sand-1",
    "data-disabled:pointer-events-none data-disabled:opacity-50",
  ],
});

export const comboboxChips = tv({
  base: "flex min-w-0 flex-1 flex-wrap items-center gap-1.5",
});

export const comboboxChip = tv({
  base: [
    "inline-flex w-fit items-center justify-center gap-1 rounded-1 border border-transparent bg-sand-3 px-1.5 py-0.5",
    "font-mono text-1 font-medium whitespace-nowrap text-sand-11 [&_svg]:size-3",
  ],
});

export const comboboxChipRemove = tv({
  base: [
    "flex size-3 shrink-0 items-center justify-center rounded-1 text-sand-11 outline-none",
    "hover:text-sand-12 focus-visible:text-sand-12",
  ],
});

export const comboboxInput = tv({
  base: [
    "min-w-12 flex-1 bg-transparent text-2 text-sand-12 outline-none",
    "placeholder:text-sand-a9",
  ],
});

export const comboboxPopup = tv({
  base: [
    "max-h-[min(16rem,var(--available-height))] w-max max-w-80 origin-(--transform-origin) overflow-y-auto rounded-3 bg-sand-1 p-1 shadow-3 outline-none",
    "transition-[scale,opacity] duration-150 ease-out data-starting-style:scale-95 data-starting-style:opacity-0",
    "data-ending-style:scale-95 data-ending-style:opacity-0",
  ],
});

export const comboboxItem = tv({
  slots: {
    item: [
      "flex min-h-8 cursor-pointer items-start justify-between gap-2 rounded-2 px-3 py-1.5 text-2 text-sand-12 outline-none select-none",
      "data-disabled:pointer-events-none data-disabled:opacity-50 data-highlighted:bg-sand-3",
    ],
    label: "min-w-0 flex-1 whitespace-normal break-words",
    indicator: "flex size-4 shrink-0 items-center justify-center text-sand-12 [&_svg]:size-4",
  },
});

export const comboboxEmpty = tv({
  base: "px-3 py-2 text-2 text-sand-11",
});

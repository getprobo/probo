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

// Cookie-banner picker in the privacy panel. Two-line items (name + origin)
// do not fit the kit DropdownItem's fixed height, so the list item owns these
// slots and renders a Menu.Item directly.
export const cookieBannerSwitcher = tv({
  slots: {
    popup: "w-72 p-0",
    list: "max-h-96 overflow-y-auto px-1 py-1",
    empty: "block px-3 py-2",
    item: [
      "group flex w-full cursor-pointer items-center gap-2 rounded-2 px-3 py-2 outline-none select-none",
      "text-sand-12 data-disabled:pointer-events-none data-disabled:opacity-50",
      "data-highlighted:bg-gold-9 data-highlighted:text-white",
    ],
    itemBody: "flex min-w-0 flex-1 flex-col gap-0.5",
    itemName: "truncate group-data-highlighted:text-white",
    itemOrigin: "truncate group-data-highlighted:text-white",
    itemCheck: "size-4 shrink-0",
    trigger: [
      "flex h-9 w-full min-w-0 items-center gap-2 rounded-2 px-3",
      "cursor-pointer outline-none transition-colors select-none",
      "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
    ],
    value: "min-w-0 flex-1 truncate text-left",
    valueCaret: "size-4 shrink-0",
    // h-5 matches text-2 line-height so the trigger does not jump when the
    // selected name replaces this pulse.
    valueSkeletonName: "block h-5 w-28 animate-pulse rounded-1 bg-sand-3",
  },
  variants: {
    // Ring rather than border: Button's base `border-transparent` and the
    // outline `border-sand-7` both emit under lite tv, and the transparent
    // one wins. A ring does not collide with that.
    outlined: {
      true: { trigger: "ring-1 ring-inset ring-sand-7 hover:bg-sand-3" },
    },
    active: {
      true: { trigger: "bg-gold-4 text-gold-12" },
    },
  },
});

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

// Trust Center top navigation bar. Slots are shared by the live TopBar and its
// skeleton so the loading placeholder is structurally identical.
export const topBar = tv({
  slots: {
    bar: "flex h-14 items-center bg-sand-1 px-8",
    inner: "mx-auto flex w-full max-w-[1024px] items-center gap-10",
    brand: "flex items-center gap-2",
    logo: "shrink-0",
    spacer: "h-px flex-1",
    nav: "flex items-center gap-1",
  },
});

// Rounded pill that opens the authenticated user menu.
export const topBarUserMenuTrigger = tv({
  base: [
    "flex h-8 items-center gap-2 rounded-full py-1 pr-2.5 pl-1",
    "cursor-pointer outline-none transition-colors select-none",
    "hover:bg-sand-3 data-[popup-open]:bg-sand-3",
    "focus-visible:ring-2 focus-visible:ring-sand-8 focus-visible:ring-offset-1 focus-visible:ring-offset-sand-1",
  ],
});

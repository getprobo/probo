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

// Text input (Radix "Text Field" over Base UI's Input). A bordered surface on
// the neutral background with an optional leading icon slot.
export const textField = tv({
  slots: {
    root: [
      "flex items-center gap-2 rounded-2 border border-sand-a5 bg-sand-1 text-2 text-sand-12 transition-colors",
      "focus-within:ring-2 focus-within:ring-sand-8 focus-within:ring-offset-1 focus-within:ring-offset-sand-1",
      "has-[input:disabled]:pointer-events-none has-[input:disabled]:opacity-50",
    ],
    icon: "flex size-4 shrink-0 items-center justify-center text-sand-a9 [&_svg]:size-4",
    input: "min-w-0 flex-1 bg-transparent text-sand-12 outline-none placeholder:text-sand-a9",
  },
  variants: {
    size: {
      1: { root: "h-7 px-2" },
      2: { root: "h-8 px-2" },
    },
  },
  defaultVariants: {
    size: 2,
  },
});

export const textFieldSkeleton = tv({
  base: "inline-block animate-pulse rounded-2 bg-sand-3 align-middle",
  variants: {
    size: {
      1: "h-7 w-60",
      2: "h-8 w-60",
    },
  },
  defaultVariants: {
    size: 2,
  },
});

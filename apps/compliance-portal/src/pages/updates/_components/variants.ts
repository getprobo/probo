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

// Updates list surface: a white card holding divider-separated rows. The card
// dims while a page change is loading. Slots are shared by the list and its
// skeleton so the loading placeholder matches the real layout.
export const updatesList = tv({
  slots: {
    card: "overflow-hidden rounded-5 border border-sand-3 bg-sand-1 transition-opacity duration-150",
    rows: "divide-y divide-sand-a2",
  },
  variants: {
    busy: {
      true: { card: "opacity-60" },
      false: {},
    },
  },
  defaultVariants: {
    busy: false,
  },
});

// Update detail article layout: the toolbar row (back link + subscribe), the
// centered content column, and the article body with its gold metadata row.
// Slots are shared by the detail page and its skeleton.
export const updateArticle = tv({
  slots: {
    toolbar: "flex w-full items-center justify-between gap-4",
    content: "flex w-full flex-col items-center px-8 py-8",
    article: "flex w-full max-w-2xl flex-col gap-4",
    meta: "flex items-center gap-1.5",
    metaIcon: "size-4 text-gold-9",
    body: "block whitespace-pre-wrap",
  },
});

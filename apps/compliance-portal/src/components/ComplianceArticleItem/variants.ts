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

// Compliance article list row (Figma "Compliance Article Item"): a leading
// icon, a title (+ optional eyebrow), and right-aligned meta. Designed to sit
// inside a bordered list container; the last row drops its divider. Slots are
// shared by the row and its skeleton.
export const complianceArticleItem = tv({
  slots: {
    root: "flex w-full items-center gap-4 border-b border-sand-a2 px-8 py-4 last:border-b-0",
    icon: "flex size-6 shrink-0 items-center justify-center text-gold-9 [&_svg]:size-full",
    content: "flex min-w-0 flex-1 flex-col gap-1",
    meta: "shrink-0",
    iconPlaceholder: "size-6 shrink-0 animate-pulse rounded-2 bg-sand-3",
  },
});

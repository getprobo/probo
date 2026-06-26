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

// Landing hero content rendered inside the shared HeaderBand: a centered
// title/description section above an optional bottom slot (the contact row).
// Slots are shared by the live Hero and its skeleton.
export const hero = tv({
  slots: {
    content: "flex w-full flex-col gap-10",
    section: "flex w-full flex-col gap-4",
  },
});

// Organization contact block: a horizontal row of icon + label items, with a
// top divider so it reads as the hero's bottom section. Self-contained so the
// divider only appears when there is contact info to show.
export const organizationContactInfo = tv({
  slots: {
    // Only a top divider gap; the band's py-8 provides the bottom spacing.
    root: "flex w-full items-center gap-6 border-t border-sand-6 pt-4",
    item: "flex items-center gap-2 text-sand-11 [&_svg]:size-5",
    link: "flex items-center gap-2 text-sand-11 hover:underline [&_svg]:size-5",
  },
});

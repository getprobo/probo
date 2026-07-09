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

// Subprocessor card: the logo box placed over the BackdropCard header (backed by
// a blurred, magnified copy of the logo) and the hosting-regions row shown in
// the body. The card frame and backdrop live in BackdropCard.
export const subprocessorListItem = tv({
  slots: {
    logo: "relative z-10 flex size-10 items-center justify-center overflow-hidden rounded-2 bg-sand-1",
    logoImage: "size-full object-cover",
    logoFallbackIcon: "text-sand-9",
    region: "flex items-start gap-1",
    regionIcon: "size-4 shrink-0 text-gold-11",
  },
});

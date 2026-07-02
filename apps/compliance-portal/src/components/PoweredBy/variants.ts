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

// "Powered by Probo" footer (Figma "Powered by"): a borderless, full-width row
// on the body surface, with a dotted texture that fades in from the bottom.
export const poweredBy = tv({
  slots: {
    root: "relative flex w-full items-center justify-center overflow-hidden bg-sand-2 py-6 text-sand-11",
    backdrop: "pointer-events-none absolute inset-0",
    backdropFade: "pointer-events-none absolute inset-0 bg-linear-to-b from-sand-2 to-sand-2/0",
    content: "relative z-10 flex items-center justify-center gap-2",
    logo: "h-6 w-auto text-sand-9",
  },
});

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

// Soft Card frame shared by the commitment / subprocessor cards: a header with a
// decorative backdrop (dotted texture or a blurred, magnified logo) faded toward
// the card surface, with centered media on top, above a left-aligned body.
export const backdropCard = tv({
  slots: {
    header: "relative flex w-full items-center overflow-hidden p-8",
    blurBackdrop: "pointer-events-none absolute inset-0 size-full scale-150 object-cover opacity-10 blur-lg",
    backdrop: "pointer-events-none absolute inset-0",
    backdropFade: "pointer-events-none absolute inset-0 bg-linear-to-b from-sand-1/0 to-sand-1",
    body: "flex flex-col gap-2 px-8 pb-8",
  },
});

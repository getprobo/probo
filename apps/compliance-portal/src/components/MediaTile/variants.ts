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

import type { CSSProperties } from "react";
import { tv } from "tailwind-variants/lite";

// Captioned media tile (Figma "Framework Card" / "Logo Card"): a media container
// with a decorative backdrop above a centered caption, rendered inside a soft
// Card. The "icon" variant (default) uses a padded square with a dotted texture;
// the "logo" variant fills a square container and uses a blurred, magnified copy
// of the logo as the backdrop. Slots are shared by the shell and its skeleton.
export const mediaTile = tv({
  slots: {
    media: "relative flex w-full items-center justify-center overflow-hidden border-b border-sand-a2",
    backdrop: "pointer-events-none absolute inset-0",
    blurBackdrop: "pointer-events-none absolute inset-0 size-full scale-150 object-cover opacity-10 blur-lg",
    backdropFade: "pointer-events-none absolute inset-0 bg-linear-to-b from-sand-1/0 to-sand-1",
    mediaContent: "relative z-10 flex size-16 items-center justify-center [&_img]:size-full [&_img]:object-contain",
    caption: "flex w-full items-center justify-center px-4 py-3",
  },
  variants: {
    variant: {
      icon: { media: "p-4" },
      logo: { media: "aspect-square" },
    },
  },
  defaultVariants: {
    variant: "icon",
  },
});

// Repeating dot texture as a radial-gradient using a Radix sand alpha token
// (theme-aware, no binary asset), faded toward the card surface by the
// backdropFade gradient slot. Maps Figma's neutral-alpha dots to sand-a5. Shared
// by the media/commitment cards, the updates list, and the footer.
export const dotPatternStyle: CSSProperties = {
  backgroundImage: "radial-gradient(var(--sand-a5) 0.75px, transparent 1.25px)",
  backgroundSize: "9px 9px",
};

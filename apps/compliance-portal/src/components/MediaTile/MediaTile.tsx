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

import { Card } from "@probo/ui/src/v2/Card/Card";
import type { ReactNode } from "react";
import type { VariantProps } from "tailwind-variants/lite";

import { dotPatternStyle, mediaTile } from "./variants";

export type MediaTileProps = VariantProps<typeof mediaTile> & {
  // Centered media (logo/icon image) shown over the backdrop.
  media: ReactNode;
  // Caption below the media (pass a Text primitive).
  label: ReactNode;
  // For the "logo" variant: the image URL rendered as a blurred, magnified
  // backdrop behind the media. When omitted, the dotted texture is shown.
  backdropSrc?: string;
};

// Captioned media tile, composing the soft Card surface. The "icon" variant
// shows a dotted backdrop; the "logo" variant shows a blurred, magnified copy of
// the logo (via backdropSrc). Purely presentational, so it doubles as the layout
// for its skeleton.
export function MediaTile({ media, label, variant = "icon", backdropSrc }: MediaTileProps) {
  const slots = mediaTile({ variant });

  return (
    <Card variant="soft" size={3} padding="none">
      <div className={slots.media()}>
        {backdropSrc != null
          ? <img src={backdropSrc} alt="" aria-hidden className={slots.blurBackdrop()} />
          : <div className={slots.backdrop()} style={dotPatternStyle} />}
        <div className={slots.backdropFade()} />
        <div className={slots.mediaContent()}>{media}</div>
      </div>
      <div className={slots.caption()}>{label}</div>
    </Card>
  );
}

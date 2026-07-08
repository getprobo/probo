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

import { dotPatternStyle } from "#/components/MediaTile/variants";

import { backdropCard } from "./variants";

interface BackdropCardProps {
  // Centered content shown above the backdrop (an icon or a logo box). The node
  // owns its own sizing/color; position it above the backdrop with `relative
  // z-10`.
  media: ReactNode;
  // When set, a blurred, magnified copy of this image becomes the backdrop;
  // otherwise a dotted texture is shown.
  backdropSrc?: string;
  // Left-aligned body below the header.
  children: ReactNode;
}

// Soft Card frame with a decorative-backdrop header above a body. Purely
// presentational, so it doubles as the layout for its consumers' skeletons.
export function BackdropCard({ media, backdropSrc, children }: BackdropCardProps) {
  const slots = backdropCard();

  return (
    <Card variant="soft" size={3} padding="none">
      <div className={slots.header()}>
        {backdropSrc != null
          ? <img src={backdropSrc} alt="" aria-hidden className={slots.blurBackdrop()} />
          : <div className={slots.backdrop()} style={dotPatternStyle} />}
        <div className={slots.backdropFade()} />
        {media}
      </div>
      <div className={slots.body()}>{children}</div>
    </Card>
  );
}

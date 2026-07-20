// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import { Card } from "@probo/ui/src/v2/Card/Card";
import type { ReactNode } from "react";

import { dotPatternStyle } from "#/components/MediaTile/variants";

import { backdropCard } from "./variants";

interface BackdropCardProps {
  // Centered content shown above the backdrop (an icon or a logo box). The node
  // owns its own sizing/color; position it above the backdrop with `relative
  // z-1`.
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

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
import type { VariantProps } from "tailwind-variants/lite";

import {
  backdropFrameProps,
  blurBackdropClassName,
  blurBackdropStyle,
  onBackdropPointerLeave,
  onBackdropPointerMove,
} from "#/components/backdropParallax";

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
  const parallax = backdropSrc != null;

  return (
    <Card
      variant="soft"
      size={3}
      padding="none"
      onPointerMove={parallax ? onBackdropPointerMove : undefined}
      onPointerLeave={parallax ? onBackdropPointerLeave : undefined}
    >
      <div className={slots.media()} {...(parallax ? backdropFrameProps("logoTile") : undefined)}>
        {backdropSrc != null
          ? (
              <img
                src={backdropSrc}
                alt=""
                aria-hidden
                className={blurBackdropClassName("logoTile")}
                style={blurBackdropStyle("logoTile")}
              />
            )
          : <div className={slots.backdrop()} style={dotPatternStyle} />}
        <div className={slots.backdropFade()} />
        <div className={slots.mediaContent()}>{media}</div>
      </div>
      <div className={slots.caption()}>{label}</div>
    </Card>
  );
}

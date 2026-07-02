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

import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";
import type { VariantProps } from "tailwind-variants/lite";

import { MediaTile } from "./MediaTile";
import type { mediaTile } from "./variants";

export type MediaTileSkeletonProps = VariantProps<typeof mediaTile>;

// Loading placeholder paired with MediaTile: renders the same layout with a
// pulse block for the media and a TextSkeleton for the caption.
export function MediaTileSkeleton({ variant = "icon" }: MediaTileSkeletonProps) {
  return (
    <MediaTile
      variant={variant}
      media={<div className="size-full animate-pulse rounded-3 bg-sand-3" />}
      label={<TextSkeleton size={2} className="w-20" />}
    />
  );
}

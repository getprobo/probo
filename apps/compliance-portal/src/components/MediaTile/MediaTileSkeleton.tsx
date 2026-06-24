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

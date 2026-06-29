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

import { ComplianceArticleItemSkeleton } from "#/components/ComplianceArticleItem/ComplianceArticleItemSkeleton";
import { homeSection } from "#/components/HomeSection/variants";
import { dotPatternStyle } from "#/components/MediaTile/variants";

export function RecentUpdatesSectionSkeleton() {
  const slots = homeSection();

  return (
    <section className={slots.root()} aria-hidden>
      <div className={slots.header()}>
        <TextSkeleton size={2} className="w-28" />
        <TextSkeleton size={2} className="w-12" />
      </div>
      <div className="relative overflow-hidden rounded-5 border border-sand-3 bg-sand-1">
        <div aria-hidden className="pointer-events-none absolute inset-0" style={dotPatternStyle} />
        <div aria-hidden className="pointer-events-none absolute inset-0 bg-linear-to-r from-sand-1/0 to-sand-1 to-[96px]" />
        <div className="relative">
          {Array.from({ length: 5 }, (_, index) => (
            <ComplianceArticleItemSkeleton key={index} />
          ))}
        </div>
      </div>
    </section>
  );
}

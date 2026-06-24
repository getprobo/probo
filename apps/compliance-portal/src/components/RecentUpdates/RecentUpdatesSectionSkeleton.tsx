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
        <div className="relative divide-y divide-sand-a2">
          {Array.from({ length: 5 }, (_, index) => (
            <ComplianceArticleItemSkeleton key={index} />
          ))}
        </div>
      </div>
    </section>
  );
}

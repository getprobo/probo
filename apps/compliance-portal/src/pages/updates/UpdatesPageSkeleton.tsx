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

import { PaginationSkeleton } from "@probo/ui/src/v2/Pagination/PaginationSkeleton";
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";

import { ComplianceArticleItemSkeleton } from "#/components/ComplianceArticleItem/ComplianceArticleItemSkeleton";
import { HeaderBand } from "#/components/HeaderBand/HeaderBand";

const ROW_PLACEHOLDERS = ["a", "b", "c", "d", "e", "f", "g", "h", "i", "j"];

export function UpdatesPageSkeleton() {
  return (
    <>
      <HeaderBand>
        <div className="flex w-full flex-col gap-2">
          <HeadingSkeleton size={7} className="w-40" />
        </div>
      </HeaderBand>
      <div className="flex w-full flex-col items-center px-8 py-8">
        <div className="flex w-full max-w-5xl flex-col gap-8">
          <div className="overflow-hidden rounded-5 border border-sand-3 bg-sand-1" aria-hidden>
            <div className="divide-y divide-sand-a2">
              {ROW_PLACEHOLDERS.map(placeholder => (
                <ComplianceArticleItemSkeleton key={placeholder} />
              ))}
            </div>
          </div>
          <PaginationSkeleton />
        </div>
      </div>
    </>
  );
}

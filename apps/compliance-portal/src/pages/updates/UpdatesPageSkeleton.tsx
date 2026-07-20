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

import { ButtonSkeleton } from "@probo/ui/src/v2/Button/ButtonSkeleton";
import { PaginationSkeleton } from "@probo/ui/src/v2/Pagination/PaginationSkeleton";
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";

import { ComplianceArticleItemSkeleton } from "#/components/ComplianceArticleItem/ComplianceArticleItemSkeleton";
import { HeaderBand } from "#/components/HeaderBand/HeaderBand";
import { pageHeader } from "#/components/PageHeader/variants";

import { updatesList } from "./_components/variants";

// Placeholder rows for the loading card — enough to read as a list, far fewer
// than UPDATES_PAGE_SIZE so the skeleton does not dominate the viewport.
const UPDATES_SKELETON_COUNT = 5;

export function UpdatesPageSkeleton() {
  const { card, rows } = updatesList();
  const header = pageHeader();

  return (
    <>
      <HeaderBand>
        <div className={header.content()}>
          <div className={header.titleRow()}>
            <HeadingSkeleton size={7} className="w-40" />
            <ButtonSkeleton size={2} className="max-sm:w-full" />
          </div>
        </div>
      </HeaderBand>
      <div className="flex w-full flex-col items-center px-8 py-8 max-md:px-4">
        <div className="flex w-full max-w-5xl flex-col gap-8">
          <div className={card()} aria-hidden>
            <div className={rows()}>
              {Array.from({ length: UPDATES_SKELETON_COUNT }, (_, index) => (
                <ComplianceArticleItemSkeleton key={index} />
              ))}
            </div>
          </div>
          <PaginationSkeleton />
        </div>
      </div>
    </>
  );
}

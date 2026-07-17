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

import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";

import { HeaderBand } from "#/components/HeaderBand/HeaderBand";

import { rightsRequestList } from "./_components/variants";
import { requestsLayout } from "./variants";

const ROW_PLACEHOLDERS = ["a", "b", "c", "d"];

export function RequestsPageSkeleton() {
  const { page, results } = requestsLayout();
  const { card } = rightsRequestList();

  return (
    <>
      <HeaderBand>
        <div className="flex w-full items-center justify-between gap-4">
          <HeadingSkeleton size={7} className="w-64" />
          <div className="h-8 w-32 animate-pulse rounded-2 bg-sand-3" />
        </div>
      </HeaderBand>
      <div className={page()}>
        <div className={results()}>
          <div className={card()}>
            {ROW_PLACEHOLDERS.map(row => (
              <div
                key={row}
                className="h-16 animate-pulse border-b border-sand-a3 bg-sand-2 last:border-b-0"
              />
            ))}
          </div>
        </div>
      </div>
    </>
  );
}

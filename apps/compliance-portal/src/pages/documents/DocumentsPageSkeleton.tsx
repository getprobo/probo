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

import { TabsSkeleton } from "@probo/ui/src/v2/Tabs/TabsSkeleton";
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { HeaderBand } from "#/components/HeaderBand/HeaderBand";

const SECTION_PLACEHOLDERS = ["a", "b"];
const ROW_PLACEHOLDERS = ["x", "y", "z"];

export function DocumentsPageSkeleton() {
  return (
    <>
      <HeaderBand flushBottomSpace>
        <div className="flex w-full flex-col gap-2">
          <HeadingSkeleton size={7} className="w-64" />
          <TabsSkeleton />
        </div>
      </HeaderBand>
      <div className="flex w-full flex-col items-center px-8 py-8">
        <div className="flex w-full max-w-5xl flex-col gap-8">
          {SECTION_PLACEHOLDERS.map(section => (
            <div key={section} className="flex flex-col gap-3">
              <TextSkeleton size={3} className="w-40" />
              <div className="overflow-hidden rounded-4 border border-sand-a4 bg-sand-1">
                {ROW_PLACEHOLDERS.map(row => (
                  <div key={row} className="h-16 animate-pulse border-b border-sand-a3 bg-sand-2 last:border-b-0" />
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </>
  );
}

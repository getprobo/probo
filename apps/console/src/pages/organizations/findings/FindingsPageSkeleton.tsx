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
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { TableSkeleton } from "@probo/ui/src/v2/Table/TableSkeleton";
import { HeadingSkeleton } from "@probo/ui/src/v2/typography/HeadingSkeleton";
import { TextSkeleton } from "@probo/ui/src/v2/typography/TextSkeleton";

import { findingsPage } from "./_components/variants";

export function FindingsPageSkeleton() {
  const {
    root,
    header,
    intro,
    title,
    actions,
    results,
    tableFilters,
    filter,
    list,
    pager,
    skeletonTitle,
    skeletonCount,
    skeletonIntro,
    skeletonAction,
  } = findingsPage();

  return (
    <div className={root()}>
      <div className={header()}>
        <div className={intro()}>
          <div className={title()}>
            <HeadingSkeleton size={6} className={skeletonTitle()} />
            <TextSkeleton size={2} className={skeletonCount()} />
          </div>
          <TextSkeleton size={2} className={skeletonIntro()} />
        </div>
        <div className={actions()}>
          <ButtonSkeleton size={2} className={skeletonAction()} />
          <ButtonSkeleton size={2} className={skeletonAction()} />
          <ButtonSkeleton size={2} className={skeletonAction()} />
        </div>
      </div>
      <div className={results()}>
        <div className={tableFilters()}>
          <SelectSkeleton size={2} className={filter()} />
          <SelectSkeleton size={2} className={filter()} />
          <SelectSkeleton size={2} className={filter()} />
          <SelectSkeleton size={2} className={filter()} />
          <SelectSkeleton size={2} className={filter()} />
        </div>
        <div className={list()}>
          <TableSkeleton size={2} variant="surface" layout="fixed" columns={8} count={6} />
          <div className={pager()}>
            <PaginationSkeleton />
          </div>
        </div>
      </div>
    </div>
  );
}

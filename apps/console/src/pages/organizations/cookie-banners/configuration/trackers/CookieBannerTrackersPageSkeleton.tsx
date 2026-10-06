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

import { TextFieldSkeleton } from "@probo/ui/src/v2/form/TextFieldSkeleton";
import { SelectSkeleton } from "@probo/ui/src/v2/Select/SelectSkeleton";
import { TableSkeleton } from "@probo/ui/src/v2/Table/TableSkeleton";

import { CookieBannerPageHeaderSkeleton } from "../../_components/CookieBannerPageHeaderSkeleton";
import { cookieBannerListSkeleton, cookieBannerPage } from "../../variants";

export function CookieBannerTrackersPageSkeleton() {
  const { root, views, view, tools, search, filters, filter } = cookieBannerListSkeleton();

  return (
    <div className={cookieBannerPage()}>
      <CookieBannerPageHeaderSkeleton titleClassName="w-24" />
      <div className={root()}>
        <div className={views()}>
          <span className={view()} aria-hidden />
          <span className={view()} aria-hidden />
        </div>
        <div className={tools()}>
          <div className={search()}>
            <TextFieldSkeleton />
          </div>
          <div className={filters()}>
            <div className={filter()}>
              <SelectSkeleton />
            </div>
            <div className={filter()}>
              <SelectSkeleton />
            </div>
            <div className={filter()}>
              <SelectSkeleton />
            </div>
          </div>
        </div>
        <TableSkeleton variant="surface" columns={7} count={8} />
      </div>
    </div>
  );
}
